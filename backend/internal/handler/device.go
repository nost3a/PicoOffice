package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
)

// ListDevices lists user's non-revoked login devices
func ListDevices(c *gin.Context) {
	uid := c.GetUint("uid")
	var list []model.LoginDevice
	db.Where("user_id = ? AND revoked = ?", uid, false).
		Order("last_seen DESC").Find(&list)
	c.JSON(200, gin.H{"items": list})
}

// RevokeDevice revokes one device.
// simplification: no FK between LoginDevice and RefreshToken (model is fixed),
// revokes refresh of same session matched by user_agent + ip.
func RevokeDevice(c *gin.Context) {
	uid := c.GetUint("uid")
	id, _ := strconv.Atoi(c.Param("id"))
	var dev model.LoginDevice
	if err := db.Where("id = ? AND user_id = ?", id, uid).First(&dev).Error; err != nil {
		c.JSON(404, gin.H{"error": "device not found"})
		return
	}
	db.Model(&dev).Update("revoked", true)
	// revoke unused refresh for same UA+IP
	db.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked = ? AND user_agent = ? AND ip = ?",
			uid, false, dev.UserAgent, dev.IP).
		Update("revoked", true)
	c.JSON(200, gin.H{"ok": true})
}
