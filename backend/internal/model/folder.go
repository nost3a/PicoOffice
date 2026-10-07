package model

import "time"

// Folder doc folder with parent/child hierarchy
type Folder struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	OwnerID   uint       `gorm:"index" json:"owner_id"`
	Name      string     `gorm:"size:255" json:"name"`
	ParentID  *uint      `gorm:"index" json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
