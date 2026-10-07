package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

// TotpSetup issues secret + URL (not persisted until enable verifies)
func TotpSetup(c *gin.Context) {
	uid := c.GetUint("uid")
	var u model.User
	if err := db.First(&u, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	key, err := service.GenerateTotp(u.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "generate totp failed"})
		return
	}
	c.JSON(200, gin.H{
		"secret":      key.Secret(),
		"otpauth_url": key.URL(),
	})
}

type totpEnableReq struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

// TotpEnable verifies code then enables 2FA
func TotpEnable(c *gin.Context) {
	var req totpEnableReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Secret == "" || req.Code == "" {
		c.JSON(400, gin.H{"error": "secret and code required"})
		return
	}
	if !service.VerifyTotp(req.Secret, req.Code) {
		c.JSON(400, gin.H{"error": "wrong totp code"})
		return
	}
	uid := c.GetUint("uid")
	if err := db.Model(&model.User{}).Where("id = ?", uid).
		Updates(map[string]interface{}{"totp_secret": req.Secret, "totp_enabled": true}).
		Error; err != nil {
		c.JSON(500, gin.H{"error": "enable totp failed"})
		return
	}
	c.JSON(200, gin.H{"totp_enabled": true})
}

type totpDisableReq struct {
	Code string `json:"code"`
}

// TotpDisable verifies code then disables 2FA
func TotpDisable(c *gin.Context) {
	var req totpDisableReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(400, gin.H{"error": "code required"})
		return
	}
	uid := c.GetUint("uid")
	var u model.User
	if err := db.First(&u, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	if !u.TotpEnabled {
		c.JSON(400, gin.H{"error": "totp not enabled"})
		return
	}
	if !service.VerifyTotp(u.TotpSecret, req.Code) {
		c.JSON(400, gin.H{"error": "wrong totp code"})
		return
	}
	if err := db.Model(&model.User{}).Where("id = ?", uid).
		Updates(map[string]interface{}{"totp_secret": "", "totp_enabled": false}).
		Error; err != nil {
		c.JSON(500, gin.H{"error": "disable totp failed"})
		return
	}
	c.JSON(200, gin.H{"totp_enabled": false})
}
