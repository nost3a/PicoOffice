package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/handler"
	"github.com/nost3a/PicoOffice/internal/middleware"
	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
	"github.com/nost3a/PicoOffice/internal/storage"
	"github.com/nost3a/PicoOffice/internal/ws"
)

//go:embed all:web
var webFS embed.FS

func main() {
	dbPath := getEnv("PICO_DB", "./picooffice.db")
	storageDir := getEnv("PICO_STORAGE", "../storage")
	addr := ":" + getEnv("PICO_PORT", "8080")

	os.MkdirAll(storageDir, 0o755)

	// set busy_timeout in DSN so every new gorm connection picks it up
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("open db failed", "err", err)
		os.Exit(1)
	}
	// run once more explicitly; journal_mode is persisted in db file
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout=5000")
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("get sql db failed", "err", err)
		os.Exit(1)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(
		&model.User{}, &model.Doc{}, &model.File{}, &model.DocVersion{},
		&model.Folder{}, &model.Tag{}, &model.DocTag{},
		&model.RefreshToken{}, &model.LoginDevice{}, &model.AuditLog{},
		&model.MailAccount{}, &model.MailMessage{},
		&model.CalendarEvent{}, &model.CollabStep{},
	); err != nil {
		slog.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	handler.SetDB(db)
	handler.SetSQLDB(sqlDB)
	handler.SetStorage(storageDir)
	hub := ws.NewHub()
	go hub.Run()
	handler.SetWSHub(hub)

	// object storage init: default local; PICO_STORAGE_KIND=s3 reads PICO_S3_* .
	// failure logs and falls back to local; storage must never crash the process.
	stoKind := getEnv("PICO_STORAGE_KIND", "local")
	sto, err := storage.Init(stoKind, map[string]string{
		"root":        storageDir,
		"endpoint":    getEnv("PICO_S3_ENDPOINT", ""),
		"access_key":   getEnv("PICO_S3_ACCESS_KEY", ""),
		"secret_key":  getEnv("PICO_S3_SECRET_KEY", ""),
		"bucket":      getEnv("PICO_S3_BUCKET", ""),
		"region":      getEnv("PICO_S3_REGION", ""),
		"ssl":         getEnv("PICO_S3_SSL", "false"),
	})
	if err != nil {
		slog.Error("storage init failed, fallback to local", "kind", stoKind, "err", err)
		sto, _ = storage.Init("local", map[string]string{"root": storageDir})
	}
	slog.Info("storage ready", "kind", sto.Kind(), "root", storageDir)
	// wired into business layer: upload/download/avatar all use this abstraction
	handler.SetObjectStore(sto)

	// search index: if dir missing, rebuild async in background without blocking HTTP startup.
	idxDir := filepath.Join(storageDir, "index.bleve")
	if _, statErr := os.Stat(idxDir); statErr != nil {
		go func() {
			time.Sleep(800 * time.Millisecond) // wait for server to start serving traffic
			n, rebuildErr := service.RebuildIndex(idxDir, db)
			if rebuildErr != nil {
				slog.Error("initial rebuild index failed", "err", rebuildErr)
				return
			}
			slog.Info("initial search index built", "entries", n)
		}()
	}

	// trash sweep: run at startup then hourly, purge docs past retention.
	go func() {
		clean := func() {
			n, cleanErr := service.CleanExpiredTrash(db)
			if cleanErr != nil {
				slog.Warn("clean expired trash failed", "err", cleanErr)
				return
			}
			if n > 0 {
				slog.Info("cleaned expired trash", "rows", n)
			}
		}
		clean()
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for range t.C {
			clean()
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS())

	// liveness/readiness probes, no auth
	r.GET("/healthz", handler.Healthz)
	r.GET("/readyz", handler.Readyz)

	// health check, no auth
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// public share access (no JWT)
	r.GET("/api/share/:token", handler.PublicShare)

	// register/login no auth but rate-limited; me requires JWT
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", middleware.RateLimitAuth(), handler.Register)
		auth.POST("/login", middleware.RateLimitAuth(), handler.Login)
		auth.GET("/me", middleware.JWT(), handler.Me)
		// refresh / forgot-password, no auth
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/forgot-password", handler.ForgotPassword)
		auth.POST("/reset-password", handler.ResetPassword)
	}

	// ws mounted separately; token via query
	r.GET("/api/ws/docs/:id", handler.WSDocs)

	// below all require JWT
	api := r.Group("/api")
	api.Use(middleware.JWT())
	{
		// logout requires JWT
		api.POST("/auth/logout", handler.Logout)

		docs := api.Group("/docs")
		{
			docs.GET("", handler.ListDocs)
			docs.POST("", handler.CreateDoc)
			docs.POST("/import", handler.ImportDoc) // static segment takes priority over :id
			docs.GET("/:id", handler.GetDoc)
			docs.PUT("/:id", handler.UpdateDoc)
			docs.DELETE("/:id", handler.DeleteDoc)
			docs.GET("/:id/content", handler.GetDocContent)
			docs.PUT("/:id/content", handler.UpdateDocContent)
			docs.POST("/:id/export", handler.ExportDoc)
			// version history: list / single / rollback
			docs.GET("/:id/versions", handler.ListDocVersions)
			docs.GET("/:id/versions/:vid", handler.GetDocVersion)
			docs.POST("/:id/rollback", handler.RollbackDoc)
			// share
			docs.POST("/:id/share", handler.CreateShare)
			docs.DELETE("/:id/share", handler.DeleteShare)
			// star
			docs.POST("/:id/star", handler.StarDoc)
			docs.DELETE("/:id/star", handler.UnstarDoc)
			// tags
			docs.PUT("/:id/tags", handler.SetDocTags)
			// collab steps
			docs.GET("/:id/steps", handler.GetCollabSteps)
		}

		// shared by me / shared with me
		api.GET("/shared", handler.ListShared)

		// trash
		trash := api.Group("/trash")
		{
			trash.GET("", handler.ListTrash)
			trash.POST("/:id/restore", handler.RestoreTrash)
			trash.DELETE("/:id", handler.PurgeTrash)
		}

		// folders
		folders := api.Group("/folders")
		{
			folders.GET("", handler.ListFolders)
			folders.POST("", handler.CreateFolder)
			folders.PUT("/:id", handler.UpdateFolder)
			folders.DELETE("/:id", handler.DeleteFolder)
		}

		// tags
		tags := api.Group("/tags")
		{
			tags.GET("", handler.ListTags)
			tags.POST("", handler.CreateTag)
			tags.DELETE("/:id", handler.DeleteTag)
		}

		me := api.Group("/me")
		{
			me.GET("/quota", handler.MyQuota)
			// profile
			me.GET("/profile", handler.GetProfile)
			me.PUT("/profile", handler.UpdateProfile)
			me.POST("/avatar", handler.UploadAvatar)
			// TOTP two-factor
			me.POST("/totp/setup", handler.TotpSetup)
			me.POST("/totp/enable", handler.TotpEnable)
			me.POST("/totp/disable", handler.TotpDisable)
			// login devices
			me.GET("/devices", handler.ListDevices)
			me.DELETE("/devices/:id", handler.RevokeDevice)
		}

		// mailbox
		mail := api.Group("/mail")
		{
			mail.GET("/accounts", handler.ListMailAccounts)
			mail.POST("/accounts", handler.CreateMailAccount)
			mail.DELETE("/accounts/:id", handler.DeleteMailAccount)
			mail.POST("/accounts/:id/sync", handler.SyncMail)
			mail.GET("/messages", handler.ListMails)
			mail.POST("/send", handler.SendMail)
		}

		// calendar
		cal := api.Group("/calendar")
		{
			cal.GET("/events", handler.ListEvents)
			cal.POST("/events", handler.CreateEvent)
			cal.PUT("/events/:id", handler.UpdateEvent)
			cal.DELETE("/events/:id", handler.DeleteEvent)
		}

		// global search
		api.GET("/search", handler.Search)

		fsg := api.Group("/fs")
		{
			fsg.GET("/files", handler.ListFiles)
			fsg.DELETE("/files/:id", handler.DeleteFile)
			fsg.POST("/upload", handler.UploadFile)
			fsg.POST("/upload/init", handler.UploadInit)
			fsg.POST("/upload/chunk", handler.UploadChunk)
			fsg.POST("/upload/complete", handler.UploadComplete)
			fsg.GET("/upload/status", handler.UploadStatus)
			fsg.GET("/download/:id", handler.DownloadFile)
		}
		admin := api.Group("/admin")
		admin.Use(middleware.AdminRequired())
		{
			admin.GET("/users", handler.ListUsers)
			admin.PUT("/users/:id/quota", handler.SetUserQuota)
			admin.GET("/audit", handler.ListAudit)
		}
	}

	// avatar static dir: register before NoRoute so files are served, not swallowed by SPA fallback.
	avatarDir := filepath.Join(storageDir, "avatars")
	os.MkdirAll(avatarDir, 0o755)
	r.Static("/avatars", avatarDir)

	// frontend static; non-/api paths served from embed
	dist, err := fs.Sub(webFS, "web")
	if err != nil {
		slog.Error("open web embed failed", "err", err)
		os.Exit(1)
	}
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		// serve real static file if hit, else index.html (SPA deep path, no 301)
		reqPath := strings.TrimPrefix(c.Request.URL.Path, "/")
		if reqPath != "" && reqPath != "index.html" {
			if f, err := dist.Open(reqPath); err == nil {
				if st, e := f.Stat(); e == nil && !st.IsDir() {
					f.Close()
					c.FileFromFS(c.Request.URL.Path, http.FS(dist))
					return
				}
				f.Close()
			}
		}
		if idx, err := fs.ReadFile(dist, "index.html"); err == nil {
			c.Data(200, "text/html; charset=utf-8", idx)
			return
		}
		c.Status(404)
	})

	slog.Info("picooffice listen", "addr", addr, "db", dbPath, "storage", storageDir)
	if err := r.Run(addr); err != nil {
		slog.Error("server run failed", "err", err)
		os.Exit(1)
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
