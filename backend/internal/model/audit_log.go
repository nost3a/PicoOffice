package model

import "time"

// AuditLog records key operations
type AuditLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:64" json:"action"`
	Target    string    `gorm:"size:255" json:"target"`
	IP        string    `gorm:"size:64" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}
