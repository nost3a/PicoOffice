package service

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	_ "github.com/blevesearch/bleve/v2/analysis/lang/cjk" // register cjk bigram analyzer
	"github.com/blevesearch/bleve/v2/mapping"
	search "github.com/blevesearch/bleve/v2/search"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// SearchResult one hit
type SearchResult struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // doc | file
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

var (
	idx     bleve.Index
	idxPath string
	idxMu   sync.Mutex
)

func newIndexMapping() *mapping.IndexMappingImpl {
	m := bleve.NewIndexMapping()
	m.DefaultAnalyzer = "cjk" // CJK bigram tokenizer
	return m
}

// openIndex opens existing index or creates; reuses one handle globally
func openIndex(path string) (bleve.Index, error) {
	idxMu.Lock()
	defer idxMu.Unlock()
	if idx != nil && idxPath == path {
		return idx, nil
	}
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		if opened, err := bleve.Open(path); err == nil {
			idx = opened
			idxPath = path
			return idx, nil
		}
	}
	opened, err := bleve.New(path, newIndexMapping())
	if err != nil {
		return nil, err
	}
	idx = opened
	idxPath = path
	return idx, nil
}

// RebuildIndex rebuilds from scratch: wipe dir, then index all non-deleted docs/files
func RebuildIndex(path string, db *gorm.DB) (int, error) {
	idxMu.Lock()
	if idx != nil {
		idx.Close()
		idx = nil
	}
	idxMu.Unlock()

	if err := os.RemoveAll(path); err != nil {
		return 0, err
	}
	i, err := bleve.New(path, newIndexMapping())
	if err != nil {
		return 0, err
	}
	idxMu.Lock()
	idx = i
	idxPath = path
	idxMu.Unlock()

	batch := i.NewBatch()
	count := 0

	var docs []model.Doc
	if err := db.Where("deleted_at IS NULL").Find(&docs).Error; err != nil {
		return 0, err
	}
	for _, d := range docs {
		docID := "doc_" + strconv.FormatUint(uint64(d.ID), 10)
		batch.Index(docID, map[string]interface{}{
			"id":       docID,
			"type":     "doc",
			"title":    d.Title,
			"body":     d.Content,
			"owner_id": strconv.FormatUint(uint64(d.OwnerID), 10),
		})
		count++
	}

	var files []model.File
	if err := db.Find(&files).Error; err != nil {
		return 0, err
	}
	for _, f := range files {
		fid := "file_" + strconv.FormatUint(uint64(f.ID), 10)
		batch.Index(fid, map[string]interface{}{
			"id":       fid,
			"type":     "file",
			"title":    f.OriginalName,
			"body":     "",
			"owner_id": strconv.FormatUint(uint64(f.UploaderID), 10),
		})
		count++
	}

	if err := i.Batch(batch); err != nil {
		return count, err
	}
	return count, nil
}

// UpsertDocIndex indexes one doc (called by save hook)
func UpsertDocIndex(path string, d *model.Doc) error {
	i, err := openIndex(path)
	if err != nil {
		return err
	}
	docID := "doc_" + strconv.FormatUint(uint64(d.ID), 10)
	return i.Index(docID, map[string]interface{}{
		"id":       docID,
		"type":     "doc",
		"title":    d.Title,
		"body":     d.Content,
		"owner_id": strconv.FormatUint(uint64(d.OwnerID), 10),
	})
}

// RemoveDocIndex deletes one index entry
func RemoveDocIndex(path, docID string) error {
	i, err := openIndex(path)
	if err != nil {
		return err
	}
	return i.Delete(docID)
}

// SearchIndex searches keyword, returns only owner's results
func SearchIndex(path string, q string, ownerID uint) ([]SearchResult, error) {
	i, err := openIndex(path)
	if err != nil {
		return nil, err
	}
	mq := bleve.NewMatchQuery(q)
	req := bleve.NewSearchRequest(mq)
	req.Fields = []string{"id", "type", "title", "body", "owner_id"}
	req.Highlight = bleve.NewHighlight()
	req.Size = 50

	res, err := i.Search(req)
	if err != nil {
		return nil, err
	}
	ownerStr := strconv.FormatUint(uint64(ownerID), 10)
	out := make([]SearchResult, 0, len(res.Hits))
	for _, hit := range res.Hits {
		owner, _ := hit.Fields["owner_id"].(string)
		if owner != ownerStr {
			continue
		}
		id, _ := hit.Fields["id"].(string)
		typ, _ := hit.Fields["type"].(string)
		title, _ := hit.Fields["title"].(string)
		out = append(out, SearchResult{
			ID:      id,
			Type:    typ,
			Title:   title,
			Snippet: pickSnippet(hit),
		})
	}
	return out, nil
}

// pickSnippet prefers highlight; else first 120 chars
func pickSnippet(hit *search.DocumentMatch) string {
	if len(hit.Fragments) > 0 {
		if frags, ok := hit.Fragments["body"]; ok && len(frags) > 0 {
			return strings.Join(frags, " ... ")
		}
		for _, frags := range hit.Fragments {
			if len(frags) > 0 {
				return strings.Join(frags, " ... ")
			}
		}
	}
	if b, ok := hit.Fields["body"].(string); ok && b != "" {
		b = strings.ReplaceAll(b, "\n", " ")
		if len(b) > 120 {
			return b[:120] + "..."
		}
		return b
	}
	return ""
}
