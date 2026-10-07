package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

// ListMailAccounts user's mail accounts (password field json:"-")
func ListMailAccounts(c *gin.Context) {
	uid := c.GetUint("uid")
	var list []model.MailAccount
	db.Where("user_id = ?", uid).Order("id DESC").Find(&list)
	c.JSON(200, gin.H{"items": list})
}

type mailAccountReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	ImapHost string `json:"imap_host"`
	ImapPort int    `json:"imap_port"`
	SmtpHost string `json:"smtp_host"`
	SmtpPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseTLS   bool   `json:"use_tls"`
}

// CreateMailAccount binds a mailbox
func CreateMailAccount(c *gin.Context) {
	uid := c.GetUint("uid")
	var req mailAccountReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(400, gin.H{"error": "email required"})
		return
	}
	if req.ImapPort == 0 {
		req.ImapPort = 993
	}
	if req.SmtpPort == 0 {
		req.SmtpPort = 465
	}
	if req.Username == "" {
		req.Username = req.Email
	}
	acc := model.MailAccount{
		UserID:   uid,
		Name:     req.Name,
		Email:    req.Email,
		ImapHost: req.ImapHost,
		ImapPort: req.ImapPort,
		SmtpHost: req.SmtpHost,
		SmtpPort: req.SmtpPort,
		Username: req.Username,
		Password: service.ObfuscatePassword(req.Password), // light obfuscation at rest; production must encrypt
		UseTLS:   req.UseTLS,
	}
	if err := db.Create(&acc).Error; err != nil {
		c.JSON(500, gin.H{"error": "create account failed"})
		return
	}
	c.JSON(200, acc)
}

// loadAccount loads current user's mail account
func loadAccount(c *gin.Context) (*model.MailAccount, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad id"})
		return nil, false
	}
	var acc model.MailAccount
	if err := db.First(&acc, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "account not found"})
		return nil, false
	}
	if acc.UserID != c.GetUint("uid") {
		c.JSON(403, gin.H{"error": "no permission"})
		return nil, false
	}
	return &acc, true
}

// DeleteMailAccount removes mailbox and its fetched messages
func DeleteMailAccount(c *gin.Context) {
	acc, ok := loadAccount(c)
	if !ok {
		return
	}
	db.Transaction(func(tx *gorm.DB) error {
		tx.Where("account_id = ?", acc.ID).Delete(&model.MailMessage{})
		return tx.Delete(&model.MailAccount{}, acc.ID).Error
	})
	c.JSON(200, gin.H{"deleted": acc.ID})
}

// SyncMail triggers one IMAP fetch
func SyncMail(c *gin.Context) {
	acc, ok := loadAccount(c)
	if !ok {
		return
	}
	added, err := service.SyncInbox(db, acc)
	if err != nil {
		// surface connection/auth errors for client debugging
		c.JSON(502, gin.H{"error": "sync failed", "detail": err.Error()})
		return
	}
	c.JSON(200, gin.H{"account_id": acc.ID, "added": added})
}

// ListMails paginated mail list
func ListMails(c *gin.Context) {
	uid := c.GetUint("uid")
	q := db.Model(&model.MailMessage{}).
		Joins("JOIN mail_accounts ON mail_accounts.id = mail_messages.account_id").
		Where("mail_accounts.user_id = ?", uid)
	if aid := c.Query("account_id"); aid != "" {
		q = q.Where("mail_messages.account_id = ?", aid)
	}
	if folder := c.DefaultQuery("folder", "INBOX"); folder != "" {
		q = q.Where("mail_messages.folder = ?", folder)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	var total int64
	q.Count(&total)
	var list []model.MailMessage
	q.Order("received_at DESC").Offset((page - 1) * size).Limit(size).Find(&list)
	c.JSON(200, gin.H{"total": total, "page": page, "size": size, "items": list})
}

type sendMailReq struct {
	AccountID uint   `json:"account_id"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

// SendMail sends via bound account
func SendMail(c *gin.Context) {
	uid := c.GetUint("uid")
	var req sendMailReq
	if err := c.ShouldBindJSON(&req); err != nil || req.To == "" {
		c.JSON(400, gin.H{"error": "to required"})
		return
	}
	var acc model.MailAccount
	if err := db.First(&acc, req.AccountID).Error; err != nil {
		c.JSON(404, gin.H{"error": "account not found"})
		return
	}
	if acc.UserID != uid {
		c.JSON(403, gin.H{"error": "no permission"})
		return
	}
	if err := service.SendMail(&acc, req.To, req.Subject, req.Body); err != nil {
		c.JSON(502, gin.H{"error": "send failed", "detail": err.Error()})
		return
	}
	c.JSON(200, gin.H{"sent": true, "to": req.To})
}
