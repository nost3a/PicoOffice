package handler

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

// avatarAllow allowed avatar ext -> Content-Type
var avatarAllow = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
}
const maxAvatarSize = 5 << 20 // 5MB

// GetProfile returns profile.
// simplification (2FA): if enabled, client must send X-Totp-Code before home; verify here, 401 on fail.
// session: if no valid refresh, issue a new pair and write LoginDevice,
// returns plaintext refresh_token (client fetches refresh here after login).
func GetProfile(c *gin.Context) {
	uid := c.GetUint("uid")
	var u model.User
	if err := db.First(&u, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	// 2FA: if enabled and X-Totp-Code present, verify; not enforced when absent (client-side)
	if u.TotpEnabled {
		code := c.GetHeader("X-Totp-Code")
		if code != "" && !service.VerifyTotp(u.TotpSecret, code) {
			c.JSON(401, gin.H{"error": "invalid totp code", "need_totp": true})
			return
		}
	}

	resp := gin.H{
		"id":           u.ID,
		"username":     u.Username,
		"nickname":     u.Nickname,
		"role":         u.Role,
		"avatar":       u.Avatar,
		"bio":          u.Bio,
		"totp_enabled": u.TotpEnabled,
	}

	// issue a new session if no valid refresh; send refresh to client
	if service.ValidRefreshCount(db, uid) == 0 {
		sess, err := service.NewSession(db, uid, u.Username, u.Role,
			c.Request.UserAgent(), c.ClientIP())
		if err == nil {
			resp["refresh_token"] = sess.Refresh
			resp["access_token"] = sess.Access
			// record a login device
			dev := model.LoginDevice{
				UserID:     uid,
				DeviceName: deviceNameFromUA(c.Request.UserAgent()),
				UserAgent:  truncateStr(c.Request.UserAgent(), 255),
				IP:         c.ClientIP(),
				LastSeen:   time.Now(),
			}
			db.Create(&dev)
		}
	}
	c.JSON(200, resp)
}

type updateProfileReq struct {
	Nickname *string `json:"nickname"`
	Bio      *string `json:"bio"`
}

// UpdateProfile updates nickname/bio (PUT, only provided fields)
func UpdateProfile(c *gin.Context) {
	uid := c.GetUint("uid")
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	updates := map[string]interface{}{}
	if req.Nickname != nil {
		updates["nickname"] = truncateStr(strings.TrimSpace(*req.Nickname), 64)
	}
	if req.Bio != nil {
		updates["bio"] = truncateStr(strings.TrimSpace(*req.Bio), 255)
	}
	if len(updates) > 0 {
		if err := db.Model(&model.User{}).Where("id = ?", uid).Updates(updates).Error; err != nil {
			c.JSON(500, gin.H{"error": "update failed"})
			return
		}
	}
	var u model.User
	db.First(&u, uid)
	c.JSON(200, gin.H{"id": u.ID, "nickname": u.Nickname, "bio": u.Bio, "avatar": u.Avatar})
}

// UploadAvatar uploads avatar (multipart), stored at storage/avatars/{uid}{ext}
func UploadAvatar(c *gin.Context) {
	uid := c.GetUint("uid")
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "multipart field 'file' required"})
		return
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if _, ok := avatarAllow[ext]; !ok {
		c.JSON(400, gin.H{"error": "only png/jpg/gif/webp allowed"})
		return
	}
	if fh.Size > maxAvatarSize {
		c.JSON(413, gin.H{"error": "avatar too large (<5MB)"})
		return
	}

	dir := filepath.Join(storageDir, "avatars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(500, gin.H{"error": "mkdir avatars"})
		return
	}
	filename := itoa(uid) + ext
	dst := filepath.Join(dir, filename)

	// write object store first (local: LocalStorage to storageDir/avatars/<file>;
	// s3: object store). Failure returns error, never swallowed.
	if objStore != nil {
		in, err := fh.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "open avatar"})
			return
		}
		if _, err := objStore.Put(c.Request.Context(), "avatars/"+filename, in, fh.Size); err != nil {
			in.Close()
			c.JSON(500, gin.H{"error": "upload avatar to storage: " + err.Error()})
			return
		}
		in.Close()
	}
	// in any mode, also save a copy under local storageDir/avatars,
	// keeps main.go r.Static("/avatars") working in s3 mode.
	if err := c.SaveUploadedFile(fh, dst); err != nil {
		c.JSON(500, gin.H{"error": "save avatar failed"})
		return
	}
	// browser path; main.go mounts storageDir/avatars at /avatars
	avatarPath := "/avatars/" + filename
	if err := db.Model(&model.User{}).Where("id = ?", uid).
		Update("avatar", avatarPath).Error; err != nil {
		c.JSON(500, gin.H{"error": "update avatar failed"})
		return
	}
	c.JSON(200, gin.H{"avatar": avatarPath})
}

// deviceNameFromUA guesses device name from User-Agent
func deviceNameFromUA(ua string) string {
	u := strings.ToLower(ua)
	switch {
	case strings.Contains(u, "micromessenger"):
		return "微信内置浏览器"
	case strings.Contains(u, "iphone"):
		return "iPhone"
	case strings.Contains(u, "ipad"):
		return "iPad"
	case strings.Contains(u, "android"):
		return "Android 设备"
	case strings.Contains(u, "windows"):
		return "Windows"
	case strings.Contains(u, "mac os"):
		return "Mac"
	case strings.Contains(u, "linux"):
		return "Linux"
	default:
		return "未知设备"
	}
}

func truncateStr(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func itoa(v uint) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
