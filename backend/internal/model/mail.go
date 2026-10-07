package model

import "time"

// MailAccount external mailbox config
type MailAccount struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Name      string    `gorm:"size:128" json:"name"`
	Email     string    `gorm:"size:128" json:"email"`
	ImapHost  string    `gorm:"size:128" json:"imap_host"`
	ImapPort  int       `json:"imap_port"`
	SmtpHost  string    `gorm:"size:128" json:"smtp_host"`
	SmtpPort  int       `json:"smtp_port"`
	Username  string    `gorm:"size:128" json:"username"`
	Password  string    `gorm:"size:255" json:"-"`
	UseTLS    bool      `json:"use_tls"`
	CreatedAt time.Time `json:"created_at"`
}

// MailMessage fetched message
type MailMessage struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	AccountID  uint      `gorm:"index" json:"account_id"`
	UID        string    `gorm:"size:128" json:"uid"`
	From       string    `gorm:"size:255" json:"from"`
	To         string    `gorm:"size:255" json:"to"`
	Subject    string    `gorm:"size:255" json:"subject"`
	Body       string    `gorm:"type:longtext" json:"body"`
	IsRead     bool      `json:"is_read"`
	Folder     string    `gorm:"size:64" json:"folder"`
	ReceivedAt time.Time `json:"received_at"`
}
