package model

import "time"

// File user upload record
type File struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UploaderID   uint      `gorm:"index" json:"uploader_id"`
	OriginalName string    `gorm:"size:255" json:"original_name"`
	StoredPath   string    `gorm:"size:512" json:"-"`
	Size         int64     `json:"size"`
	Mime         string    `gorm:"size:128" json:"mime"`
	CreatedAt    time.Time `json:"created_at"`
}
