package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
)

// MyQuota returns user's quota usage
func MyQuota(c *gin.Context) {
	uid := c.GetUint("uid")
	var u model.User
	if err := db.First(&u, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	c.JSON(200, gin.H{
		"quota_bytes": u.QuotaBytes,
		"used_bytes":  u.UsedBytes,
	})
}
