package model

import "time"

// RefreshToken refresh record; hash only
type RefreshToken struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	TokenHash string    `gorm:"size:128" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	IP        string    `gorm:"size:64" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}
