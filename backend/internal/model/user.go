package model

import "time"

// User system user; first registered is admin
type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string    `gorm:"size:128" json:"-"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	Role         string    `gorm:"size:16;default:user" json:"role"`
	QuotaBytes   int64     `gorm:"default:1073741824" json:"quota_bytes"` // storage quota, default 1GB
	UsedBytes    int64     `json:"used_bytes"`                            // used bytes
	Avatar       string    `gorm:"size:512" json:"avatar"`
	Bio          string    `gorm:"size:255" json:"bio"`
	TotpSecret   string    `gorm:"size:128" json:"-"`
	TotpEnabled  bool      `gorm:"default:false" json:"totp_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
