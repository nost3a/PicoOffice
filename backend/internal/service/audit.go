package service

import (
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// LogAudit writes audit row; called by login/authz/export etc.
// failure logs only, never blocks main flow; audit must not break business.
func LogAudit(db *gorm.DB, uid uint, action, target, ip string) {
	if db == nil {
		return
	}
	db.Create(&model.AuditLog{
		UserID: uid,
		Action: action,
		Target: target,
		IP:     ip,
	})
}

// SeedDemoAudit inserts two demo rows when table is empty.
// called only when table is empty; no repeated writes.
func SeedDemoAudit(db *gorm.DB) {
	if db == nil {
		return
	}
	var n int64
	db.Model(&model.AuditLog{}).Count(&n)
	if n > 0 {
		return
	}
	rows := []model.AuditLog{
		{UserID: 1, Action: "login", Target: "/api/auth/login", IP: "127.0.0.1"},
		{UserID: 1, Action: "export", Target: "doc:1", IP: "127.0.0.1"},
	}
	for i := range rows {
		db.Create(&rows[i])
	}
}
