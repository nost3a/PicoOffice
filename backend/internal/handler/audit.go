package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

// auditPageSize audit log page size
const auditPageSize = 20

// ListAudit admin paginated audit log
func ListAudit(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// other handlers don't call LogAudit yet; seed demo rows when empty
	service.SeedDemoAudit(db)

	var total int64
	if err := db.Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		c.JSON(500, gin.H{"error": "count audit failed"})
		return
	}

	var items []model.AuditLog
	if err := db.Order("id DESC").
		Offset((page - 1) * auditPageSize).
		Limit(auditPageSize).
		Find(&items).Error; err != nil {
		c.JSON(500, gin.H{"error": "list audit failed"})
		return
	}
	if items == nil {
		items = []model.AuditLog{}
	}
	c.JSON(200, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  auditPageSize,
	})
}
