package model

import "time"

// CalendarEvent calendar event
type CalendarEvent struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	Title       string    `gorm:"size:255" json:"title"`
	Description string    `gorm:"type:longtext" json:"description"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	RemindMin   int       `json:"remind_min"`
	Notified    bool      `json:"notified"`
}
