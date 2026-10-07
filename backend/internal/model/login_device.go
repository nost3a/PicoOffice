package model

import "time"

// LoginDevice logged-in device for session management/revoke
type LoginDevice struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	DeviceName string   `gorm:"size:128" json:"device_name"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	IP        string    `gorm:"size:64" json:"ip"`
	LastSeen  time.Time `json:"last_seen"`
	Revoked   bool      `json:"revoked"`
}
