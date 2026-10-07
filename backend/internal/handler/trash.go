package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

// parseDocID parses uint from :id
func parseDocID(c *gin.Context) (uint, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return 0, false
	}
	return uint(id), true
}

// ListTrash lists user's soft-deleted docs
func ListTrash(c *gin.Context) {
	uid := c.GetUint("uid")
	var list []model.Doc
	// soft-deleted rows are filtered by GORM; need Unscoped + deleted_at not null
	db.Unscoped().
		Where("owner_id = ? AND deleted_at IS NOT NULL", uid).
		Order("deleted_at DESC").
		Find(&list)
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// RestoreTrash clears deleted_at
func RestoreTrash(c *gin.Context) {
	uid := c.GetUint("uid")
	id, ok := parseDocID(c)
	if !ok {
		return
	}
	var d model.Doc
	if err := db.Unscoped().First(&d, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doc not found"})
		return
	}
	if d.OwnerID != uid && c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if d.DeletedAt.Valid {
		if err := db.Unscoped().Model(&d).Update("deleted_at", nil).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "restore failed"})
			return
		}
	}
	// restore adds body back to quota
	service.RecalcUsedBytes(db, d.OwnerID)
	// restore re-indexes; docs outside trash should be searchable
	if err := service.UpsertDocIndex(indexPath(), &d); err != nil {
		slog.Warn("reindex restored doc failed", "doc", d.ID, "err", err)
	}
	c.JSON(http.StatusOK, gin.H{"id": d.ID, "restored": true})
}

// PurgeTrash hard-deletes (incl. snapshots) and restores quota
func PurgeTrash(c *gin.Context) {
	uid := c.GetUint("uid")
	id, ok := parseDocID(c)
	if !ok {
		return
	}
	var d model.Doc
	if err := db.Unscoped().First(&d, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doc not found"})
		return
	}
	if d.OwnerID != uid && c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	// delete version snapshots first
	db.Unscoped().Where("doc_id = ?", d.ID).Delete(&model.DocVersion{})
	if err := db.Unscoped().Delete(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "purge failed"})
		return
	}
	service.RecalcUsedBytes(db, d.OwnerID)
	// hard delete also removes search index entry
	if err := service.RemoveDocIndex(indexPath(), docIndexID(d.ID)); err != nil {
		slog.Warn("remove purged doc index failed", "doc", d.ID, "err", err)
	}
	c.JSON(http.StatusOK, gin.H{"purged": d.ID})
}
