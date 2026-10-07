package handler

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

// upload ext -> doc_kind; legacy doc/xls/ppt also accepted
var extToKind = map[string]string{
	".docx": "doc", ".doc": "doc",
	".xlsx": "sheet", ".xls": "sheet",
	".pptx": "slide", ".ppt": "slide",
}

// ImportDoc POST /api/docs/import, multipart field file,
// accept docx/xlsx/pptx (legacy doc/xls/ppt); soffice converts to html body.
func ImportDoc(c *gin.Context) {
	uid := c.GetUint("uid")

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing multipart field 'file'"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	kind, ok := extToKind[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported ext " + ext + ", want docx/doc/xlsx/xls/pptx/ppt"})
		return
	}

	tmpDir, err := os.MkdirTemp("", "pico-import-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "make tmp dir"})
		return
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "uploaded"+ext)
	if err := c.SaveUploadedFile(fh, srcPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save upload failed"})
		return
	}

	htmlPath, err := service.ImportToHTML(tmpDir, srcPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "convert to html failed", "detail": err.Error()})
		return
	}
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read converted html"})
		return
	}
	// sanitize imported html, same as save path
	html := service.SanitizeHTML(string(htmlBytes))

	title := strings.TrimSuffix(fh.Filename, ext)
	if title == "" {
		title = "未命名导入"
	}

	d := model.Doc{
		OwnerID:    uid,
		Title:      title,
		DocKind:    kind,
		FileExt:    strings.TrimPrefix(ext, "."),
		Content:    html,
		Size:       int64(len(html)),
		Visibility: "private",
	}
	if err := db.Create(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create doc failed"})
		return
	}
	service.RecalcUsedBytes(db, uid)
	// index on import success so global search can find it
	if err := service.UpsertDocIndex(indexPath(), &d); err != nil {
		slog.Warn("upsert imported doc index failed", "doc", d.ID, "err", err)
	}
	c.JSON(http.StatusOK, d)
}
