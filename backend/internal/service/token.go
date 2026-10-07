package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/middleware"
	"github.com/nost3a/PicoOffice/internal/model"
)

// accessTTL short access TTL; refreshTTL long refresh TTL
const (
	accessTTL  = 2 * time.Hour
	refreshTTL = 30 * 24 * time.Hour
	// ResetTokenTTL reset token TTL; reuses RefreshToken table
	ResetTokenTTL = time.Hour
	// resetTokenMaxAge separates reset vs refresh token by TTL window:
	// refresh is 30d; reset is 1h;
	// if a refresh row has <2h TTL left, treat it as reset token.
	resetTokenMaxAge = 2 * time.Hour
)

// ErrInvalidRefresh refresh invalid/expired/revoked
var ErrInvalidRefresh = errors.New("invalid or expired refresh token")

// randomToken returns hex of n random bytes
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken stores sha256 only, never plaintext
func HashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

// Session: access JWT + plaintext refresh + refresh row
type Session struct {
	Access      string
	Refresh     string
	RefreshRow  *model.RefreshToken
}

// NewSession issues a token pair and a refresh row
func NewSession(db *gorm.DB, uid uint, username, role, ua, ip string) (*Session, error) {
	access, err := middleware.IssueTokenTTL(uid, username, role, accessTTL)
	if err != nil {
		return nil, err
	}
	plain, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	row := model.RefreshToken{
		UserID:    uid,
		TokenHash: HashToken(plain),
		ExpiresAt: time.Now().Add(refreshTTL),
		UserAgent: truncate(ua, 255),
		IP:        truncate(ip, 64),
	}
	if err := db.Create(&row).Error; err != nil {
		return nil, err
	}
	return &Session{Access: access, Refresh: plain, RefreshRow: &row}, nil
}

// ValidRefreshCount counts user's unexpired non-revoked refresh,
// used to decide whether to issue a new session in profile
func ValidRefreshCount(db *gorm.DB, uid uint) int64 {
	var n int64
	db.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked = ? AND expires_at > ?", uid, false, time.Now()).
		Count(&n)
	return n
}

// RotateRefresh swaps plaintext refresh for new access: revoke old, issue pair
func RotateRefresh(db *gorm.DB, plain string) (*Session, error) {
	var row model.RefreshToken
	if err := db.Where("token_hash = ?", HashToken(plain)).First(&row).Error; err != nil {
		return nil, ErrInvalidRefresh
	}
	if row.Revoked || time.Now().After(row.ExpiresAt) {
		return nil, ErrInvalidRefresh
	}
	var u model.User
	if err := db.First(&u, row.UserID).Error; err != nil {
		return nil, err
	}
	// revoke old, issue new pair
	db.Model(&row).Update("revoked", true)
	return NewSession(db, u.ID, u.Username, u.Role, row.UserAgent, row.IP)
}

// RevokeRefresh revokes one refresh (logout)
func RevokeRefresh(db *gorm.DB, plain string) {
	if plain == "" {
		return
	}
	db.Model(&model.RefreshToken{}).
		Where("token_hash = ?", HashToken(plain)).
		Update("revoked", true)
}

// CreateResetToken issues reset token, reuses RefreshToken table (short TTL row)
func CreateResetToken(db *gorm.DB, uid uint) (string, error) {
	plain, err := randomToken(32)
	if err != nil {
		return "", err
	}
	row := model.RefreshToken{
		UserID:    uid,
		TokenHash: HashToken(plain),
		ExpiresAt: time.Now().Add(ResetTokenTTL),
	}
	if err := db.Create(&row).Error; err != nil {
		return "", err
	}
	return plain, nil
}

// ConsumeResetToken verifies reset token, returns uid on hit; error on miss/expiry
func ConsumeResetToken(db *gorm.DB, plain string) (uint, error) {
	var row model.RefreshToken
	if err := db.Where("token_hash = ?", HashToken(plain)).First(&row).Error; err != nil {
		return 0, ErrInvalidRefresh
	}
	if row.Revoked || time.Now().After(row.ExpiresAt) {
		return 0, ErrInvalidRefresh
	}
	// separates reset (1h) vs refresh (30d) by remaining TTL
	if time.Until(row.ExpiresAt) > resetTokenMaxAge {
		return 0, ErrInvalidRefresh
	}
	return row.UserID, nil
}

// RevokeAllRefresh revokes all refresh for a user (after password reset)
func RevokeAllRefresh(db *gorm.DB, uid uint) {
	db.Model(&model.RefreshToken{}).Where("user_id = ?", uid).Update("revoked", true)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
