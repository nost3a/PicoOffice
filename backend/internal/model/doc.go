package model

import (
	"time"

	"gorm.io/gorm"
)

// Doc one online doc; body stored in DB content field by default
type Doc struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	OwnerID     uint           `gorm:"index" json:"owner_id"`
	Title       string         `gorm:"size:255" json:"title"`
	DocKind     string         `gorm:"size:16" json:"doc_kind"` // doc|sheet|slide
	FileExt     string         `gorm:"size:16" json:"file_ext"`
	Size        int64          `json:"size"`
	Content     string         `gorm:"type:longtext" json:"-"`
	ContentPath string         `gorm:"size:255" json:"-"`
	Visibility  string         `gorm:"default:private" json:"visibility"` // private|shared
	ShareToken  string         `gorm:"index" json:"-"`
	SharePerm   string         `json:"share_perm"` // view|edit
	DocVersion  int            `gorm:"default:0" json:"doc_version"` // optimistic-lock version
	FolderID    *uint          `gorm:"index" json:"folder_id"`
	Starred     bool           `gorm:"index" json:"starred"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // soft delete
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
