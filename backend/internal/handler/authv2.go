package handler

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken swaps refresh for new access (no JWT).
// revoke old refresh immediately, return new access + refresh (rotation).
func RefreshToken(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		c.JSON(400, gin.H{"error": "refresh_token required"})
		return
	}
	sess, err := service.RotateRefresh(db, req.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid or expired refresh token"})
		return
	}
	c.JSON(200, gin.H{
		"token":         sess.Access, // field names match login response (client convention)
		"access_token":  sess.Access,
		"refresh_token": sess.Refresh,
		"token_type":    "Bearer",
	})
}

// Logout revokes current refresh.
// note: access JWT has no server-side blocklist; expires after 2h (pragmatic simplification).
func Logout(c *gin.Context) {
	var req refreshReq
	// body may be empty (JWT only); be tolerant
	_ = c.ShouldBindJSON(&req)
	service.RevokeRefresh(db, req.RefreshToken)
	c.JSON(200, gin.H{"ok": true})
}

type forgotReq struct {
	Username string `json:"username"`
	Email    string `json:"email"` // model has no email; fall back to username match
}

// ForgotPassword starts password recovery (no JWT).
// mail module under parallel dev: reset link logs server-side, always {sent:true}, no token leak, no enumeration.
func ForgotPassword(c *gin.Context) {
	var req forgotReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	name := strings.TrimSpace(req.Username)
	if name == "" {
		name = strings.TrimSpace(req.Email)
	}
	var u model.User
	err := db.Where("username = ?", name).First(&u).Error
	// always return sent:true to avoid account enumeration
	if err == nil {
		token, tErr := service.CreateResetToken(db, u.ID)
		if tErr == nil {
			// TODO: send email once mail module is wired; for now log only
			slog.Warn("password reset link (TODO send via email)",
				"user", u.Username,
				"link", "/reset?token="+token)
		}
	}
	c.JSON(200, gin.H{"sent": true})
}

type resetReq struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword resets password via link (no JWT). Revokes all refresh after.
func ResetPassword(c *gin.Context) {
	var req resetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	if len(req.NewPassword) < 6 {
		c.JSON(400, gin.H{"error": "password too short (>=6)"})
		return
	}
	uid, err := service.ConsumeResetToken(db, req.Token)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid or expired reset token"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "hash password failed"})
		return
	}
	if err := db.Model(&model.User{}).Where("id = ?", uid).
		Update("password_hash", string(hash)).Error; err != nil {
		c.JSON(500, gin.H{"error": "update password failed"})
		return
	}
	// force all devices offline
	service.RevokeAllRefresh(db, uid)
	c.JSON(200, gin.H{"ok": true})
}
