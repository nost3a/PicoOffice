package model

import "time"

// CollabStep collaborative cursor/op, grouped by doc+version
type CollabStep struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	DocID     uint      `gorm:"index:idx_collab_doc_ver" json:"doc_id"`
	Version   int       `gorm:"index:idx_collab_doc_ver" json:"version"`
	StepData  string    `gorm:"type:longtext" json:"step_data"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
