package model

import "time"

// Tag user-defined tag
type Tag struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	OwnerID   uint      `gorm:"index" json:"owner_id"`
	Name      string    `gorm:"size:64" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// DocTag doc-tag many-to-many link
type DocTag struct {
	DocID uint `gorm:"primaryKey"`
	TagID uint `gorm:"primaryKey"`
}
