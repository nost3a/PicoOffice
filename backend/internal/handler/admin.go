package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
)

// ListUsers admin-only, all users with quota/used
func ListUsers(c *gin.Context) {
	var list []model.User
	db.Order("id ASC").Find(&list)
	c.JSON(200, gin.H{"items": list})
}

// SetUserQuota updates a user's quota
func SetUserQuota(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var body struct {
		QuotaBytes int64 `json:"quota_bytes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	if body.QuotaBytes < 0 {
		c.JSON(400, gin.H{"error": "quota must >= 0"})
		return
	}
	res := db.Model(&model.User{}).Where("id = ?", id).Update("quota_bytes", body.QuotaBytes)
	if res.Error != nil {
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
