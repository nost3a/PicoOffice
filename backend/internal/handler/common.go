package handler

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/storage"
	"github.com/nost3a/PicoOffice/internal/ws"
)

// global dependency, injected at main startup
var (
	db         *gorm.DB
	sqlDB      *sql.DB
	storageDir string
	wsHub      *ws.Hub
	objStore   storage.Storage
)

func SetDB(d *gorm.DB)                { db = d }
func SetSQLDB(s *sql.DB)              { sqlDB = s }
func SetStorage(s string)             { storageDir = s }
func SetWSHub(h *ws.Hub)              { wsHub = h }
func SetObjectStore(s storage.Storage) { objStore = s }

// notImpl unified 501 placeholder used across handlers
func notImpl(c *gin.Context) {
	c.JSON(501, gin.H{"error": "not implemented yet"})
}
