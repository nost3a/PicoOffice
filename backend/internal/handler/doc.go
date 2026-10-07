package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

// docKind source ext; used when writing temp file for export
var kindExt = map[string]string{
	"doc":   "md",
	"sheet": "csv",
	"slide": "html",
}

func ListDocs(c *gin.Context) {
	uid := c.GetUint("uid")
	q := db.Model(&model.Doc{}).Where("owner_id = ?", uid)
	if kind := c.Query("type"); kind != "" {
		q = q.Where("doc_kind = ?", kind)
	}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		q = q.Where("title LIKE ?", "%"+kw+"%")
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	var total int64
	q.Count(&total)
	var list []model.Doc
	q.Order("updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&list)
	c.JSON(200, gin.H{"total": total, "page": page, "size": size, "items": list})
}

type createDocReq struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

func CreateDoc(c *gin.Context) {
	uid := c.GetUint("uid")
	var req createDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	ext, ok := kindExt[req.Type]
	if !ok {
		c.JSON(400, gin.H{"error": "type must be doc|sheet|slide"})
		return
	}
	if req.Title == "" {
		req.Title = "未命名"
	}
	d := model.Doc{
		OwnerID: uid,
		Title:   req.Title,
		DocKind: req.Type,
		FileExt: ext,
	}
	if err := db.Create(&d).Error; err != nil {
		c.JSON(500, gin.H{"error": "create doc failed"})
		return
	}
	c.JSON(200, d)
}

// loadDoc loads doc by :id; writes response and returns false on miss
func loadDoc(c *gin.Context) (*model.Doc, bool) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(400, gin.H{"error": "bad id"})
		return nil, false
	}
	var d model.Doc
	if err := db.First(&d, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "doc not found"})
		return nil, false
	}
	return &d, true
}

func GetDoc(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	c.JSON(200, d)
}

type updateDocReq struct {
	Title *string `json:"title"`
}

func UpdateDoc(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	var req updateDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	if req.Title != nil {
		d.Title = *req.Title
	}
	db.Save(d)
	c.JSON(200, d)
}

func DeleteDoc(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	// Doc has gorm.DeletedAt; db.Delete is soft, rows are kept
	db.Delete(d)
	// after soft-delete, body leaves quota; recompute
	service.RecalcUsedBytes(db, d.OwnerID)
	// soft-delete removes from search index so trashed docs are not found
	if err := service.RemoveDocIndex(indexPath(), docIndexID(d.ID)); err != nil {
		slog.Warn("remove doc index failed", "doc", d.ID, "err", err)
	}
	c.JSON(200, gin.H{"deleted": d.ID})
}

func GetDocContent(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	c.JSON(200, gin.H{"content": d.Content, "updated_at": d.UpdatedAt})
}

type contentReq struct {
	Content     string `json:"content"`
	BaseVersion int    `json:"base_version"` // optional; when >0, do optimistic-lock check
}

func UpdateDocContent(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	var req contentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	// optimistic lock: reject overwrite when client base_version mismatches
	if req.BaseVersion > 0 && req.BaseVersion != d.DocVersion {
		c.JSON(409, gin.H{"error": "version conflict", "current_version": d.DocVersion})
		return
	}
	// sanitize before write; docx/sheet/slide html all pass through
	sanitized := service.SanitizeHTML(req.Content)
	d.Content = sanitized
	d.Size = int64(len(sanitized))
	d.DocVersion++
	db.Save(d)
	// save a content snapshot; editor tracks who edited
	snapshotVersion(d, c.GetUint("uid"), c.GetString("username"))
	service.RecalcUsedBytes(db, d.OwnerID)
	// index after body save; failure warns only, never blocks save
	if err := service.UpsertDocIndex(indexPath(), d); err != nil {
		slog.Warn("upsert doc index failed", "doc", d.ID, "err", err)
	}
	c.JSON(200, gin.H{"content": d.Content, "updated_at": d.UpdatedAt, "doc_version": d.DocVersion})
}

// snapshotVersion saves current content; prune oldest beyond 20
func snapshotVersion(d *model.Doc, editorID uint, editorName string) {
	v := model.DocVersion{
		DocID:      d.ID,
		Content:    d.Content,
		Size:       d.Size,
		EditorID:   editorID,
		EditorName: editorName,
	}
	if err := db.Create(&v).Error; err != nil {
		// snapshot failure must not block body save
		return
	}
	var cnt int64
	db.Model(&model.DocVersion{}).Where("doc_id = ?", d.ID).Count(&cnt)
	for cnt > 20 {
		var old model.DocVersion
		if err := db.Where("doc_id = ?", d.ID).Order("created_at ASC").First(&old).Error; err != nil {
			break
		}
		db.Delete(&old)
		cnt--
	}
}

// canViewDoc only owner or admin can view versions/rollback
func canViewDoc(c *gin.Context, d *model.Doc) bool {
	uid := c.GetUint("uid")
	if d.OwnerID == uid || c.GetString("role") == "admin" {
		return true
	}
	c.JSON(403, gin.H{"error": "forbidden"})
	return false
}

// ListDocVersions version list without content to save bandwidth
func ListDocVersions(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	if !canViewDoc(c, d) {
		return
	}
	var list []model.DocVersion
	db.Where("doc_id = ?", d.ID).Order("created_at DESC").Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, v := range list {
		items = append(items, gin.H{
			"id":          v.ID,
			"size":        v.Size,
			"editor_name": v.EditorName,
			"created_at":  v.CreatedAt,
		})
	}
	c.JSON(200, gin.H{"items": items})
}

// GetDocVersion returns full content of one version
func GetDocVersion(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	if !canViewDoc(c, d) {
		return
	}
	vid, _ := strconv.Atoi(c.Param("vid"))
	var v model.DocVersion
	if err := db.First(&v, vid).Error; err != nil {
		c.JSON(404, gin.H{"error": "version not found"})
		return
	}
	if v.DocID != d.ID {
		c.JSON(404, gin.H{"error": "version not found"})
		return
	}
	c.JSON(200, gin.H{"content": v.Content, "size": v.Size, "editor_name": v.EditorName, "created_at": v.CreatedAt})
}

type rollbackReq struct {
	VersionID uint `json:"version_id"`
}

// RollbackDoc writes version back to doc and saves a new snapshot
func RollbackDoc(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	if !canViewDoc(c, d) {
		return
	}
	var req rollbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	var v model.DocVersion
	if err := db.First(&v, req.VersionID).Error; err != nil {
		c.JSON(404, gin.H{"error": "version not found"})
		return
	}
	if v.DocID != d.ID {
		c.JSON(404, gin.H{"error": "version not found"})
		return
	}
	d.Content = v.Content
	d.Size = v.Size
	db.Save(d)
	// rollback itself saves a snapshot so it can be undone
	snapshotVersion(d, c.GetUint("uid"), c.GetString("username"))
	c.JSON(200, gin.H{"content": d.Content, "updated_at": d.UpdatedAt})
}

// ExportDoc runs soffice --headless to pdf/docx/xlsx; 501 on failure
func ExportDoc(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	to := c.DefaultQuery("type", "pdf")
	allowed := map[string]bool{"pdf": true, "docx": true, "xlsx": true}
	isSlidePptx := d.DocKind == "slide" && to == "pptx"
	if !isSlidePptx && !allowed[to] {
		c.JSON(400, gin.H{"error": "type must be pdf|docx|xlsx"})
		return
	}

	tmpDir, err := os.MkdirTemp("", "pico-export-*")
	if err != nil {
		c.JSON(500, gin.H{"error": "make tmp dir"})
		return
	}
	defer os.RemoveAll(tmpDir)

	// slide export pptx: content(html) -> impress filter -> pptx.
	// note: html->odp->pptx does not work in real soffice; html gets
	// open as writer doc; no odp export filter; use impress filter directly
	// MS PowerPoint 2007 XML filter reliably produces valid pptx.
	if isSlidePptx {
		srcHTML := filepath.Join(tmpDir, fmt.Sprintf("doc-%d.html", d.ID))
		if err := os.WriteFile(srcHTML, []byte(d.Content), 0o644); err != nil {
			c.JSON(500, gin.H{"error": "write src html"})
			return
		}
		pptxPath, err := service.ConvertFile(tmpDir, srcHTML, "pptx:Impress MS PowerPoint 2007 XML", "pptx")
		if err != nil {
			c.JSON(500, gin.H{"error": "html->pptx failed", "detail": err.Error()})
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(pptxPath)))
		c.File(pptxPath)
		return
	}

	// pdf: wrap content in A4 Flat ODT before soffice export
	srcPath := filepath.Join(tmpDir, fmt.Sprintf("doc-%d.%s", d.ID, d.FileExt))
	convertTo := to
	if to == "pdf" {
		srcPath = filepath.Join(tmpDir, fmt.Sprintf("doc-%d.fodt", d.ID))
		if err := os.WriteFile(srcPath, []byte(buildFODT(d.Content)), 0o644); err != nil {
			c.JSON(500, gin.H{"error": "write src file"})
			return
		}
		convertTo = "pdf:writer_pdf_Export"
	} else {
		if err := os.WriteFile(srcPath, []byte(d.Content), 0o644); err != nil {
			c.JSON(500, gin.H{"error": "write src file"})
			return
		}
	}

	cmd := exec.Command("soffice", "--headless", "--convert-to", convertTo, "--outdir", tmpDir, srcPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "soffice not available", "detail": string(out)})
		return
	}
	outPath := filepath.Join(tmpDir, fmt.Sprintf("doc-%d.%s", d.ID, to))
	if _, err := os.Stat(outPath); err != nil {
		c.JSON(501, gin.H{"error": "converted file missing"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(outPath)))
	c.File(outPath)
}

// buildFODT wraps body into A4 Flat ODT; soffice PDF = 595x842
func buildFODT(content string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<office:document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" office:version="1.2" office:mimetype="application/vnd.oasis.opendocument.text">`)
	// A4 = 21cm x 29.7cm, 2cm margins
	b.WriteString(`<office:automatic-styles><style:page-layout style:name="pm1"><style:page-layout-properties fo:page-width="21cm" fo:page-height="29.7cm" fo:margin-top="2cm" fo:margin-bottom="2cm" fo:margin-left="2cm" fo:margin-right="2cm"/></style:page-layout></office:automatic-styles>`)
	b.WriteString(`<office:master-styles><style:master-page style:name="Standard" style:page-layout-name="pm1"/></office:master-styles>`)
	b.WriteString(`<office:body><office:text>`)
	// split by line: # / ## / ### as headings, rest as paragraphs
	for _, line := range strings.Split(content, "\n") {
		trim := strings.TrimSpace(strings.TrimRight(line, "\r"))
		if trim == "" {
			continue
		}
		level, body := 0, trim
		switch {
		case strings.HasPrefix(trim, "### "):
			level, body = 3, trim[4:]
		case strings.HasPrefix(trim, "## "):
			level, body = 2, trim[3:]
		case strings.HasPrefix(trim, "# "):
			level, body = 1, trim[2:]
		}
		esc := xmlEscape(body)
		if level > 0 {
			fmt.Fprintf(&b, `<text:h text:outline-level="%d">%s</text:h>`, level, esc)
		} else {
			fmt.Fprintf(&b, `<text:p>%s</text:p>`, esc)
		}
	}
	b.WriteString(`</office:text></office:body></office:document>`)
	return b.String()
}

// xmlEscape escapes XML special chars in body
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
