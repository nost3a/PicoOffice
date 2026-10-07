package handler

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/service"
)

// indexPath: <storage>/index.bleve; shared by search and write hooks.
func indexPath() string { return filepath.Join(storageDir, "index.bleve") }

// docIndexID bleve primary key; service uses "doc_<id>".
func docIndexID(id uint) string { return "doc_" + itoa(id) }

// Search global. ?q=kw; ?rebuild=1 rebuilds index first.
// index dir is <storage>/index.bleve; results filtered by current user owner_id.
func Search(c *gin.Context) {
	uid := c.GetUint("uid")
	q := strings.TrimSpace(c.Query("q"))
	idxDir := indexPath()

	// first run / after data change: ?rebuild=1 rebuilds index
	if c.Query("rebuild") == "1" {
		n, err := service.RebuildIndex(idxDir, db)
		if err != nil {
			c.JSON(500, gin.H{"error": "rebuild index failed", "detail": err.Error()})
			return
		}
		if q == "" {
			c.JSON(200, gin.H{"rebuilt": n, "items": []any{}})
			return
		}
	}

	if q == "" {
		c.JSON(200, gin.H{"items": []any{}})
		return
	}
	items, err := service.SearchIndex(idxDir, q, uid)
	if err != nil {
		c.JSON(500, gin.H{"error": "search failed", "detail": err.Error()})
		return
	}
	c.JSON(200, gin.H{"query": q, "items": items})
}
