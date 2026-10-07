package model

import "time"

// DocVersion snapshot after each save, for history/rollback
type DocVersion struct {
	ID         uint      `gorm:"primaryKey"`
	DocID      uint      `gorm:"index:idx_doc_version"`
	Content    string    `gorm:"type:longtext"`
	Size       int64
	EditorID   uint
	EditorName string
	CreatedAt  time.Time
}
