package handler

import (
	"context"
	"log/slog"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/service"
)

// Healthz liveness probe, always 200
func Healthz(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// Readyz readiness: db/soffice/disk all ok -> 200, else 503 with reason.
func Readyz(c *gin.Context) {
	res := gin.H{}
	allOK := true

	// 1. db reachable
	dbOK := false
	if sqlDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := sqlDB.PingContext(ctx); err != nil {
			slog.Warn("readyz db ping failed", "err", err)
		} else {
			dbOK = true
		}
		cancel()
	}
	if dbOK {
		res["db"] = "ok"
	} else {
		res["db"] = "fail"
		allOK = false
	}

	// 2. soffice executable (export / online preview depends on it)
	if _, err := exec.LookPath("soffice"); err == nil {
		res["soffice"] = "ok"
	} else {
		slog.Warn("readyz soffice missing", "err", err)
		res["soffice"] = "fail"
		allOK = false
	}

	// 3. free disk space (storage volume; <100MB = unhealthy)
	diskPath := storageDir
	if diskPath == "" {
		diskPath = "."
	}
	if healthy, err := service.DiskHealthy(diskPath); err != nil {
		slog.Warn("readyz disk check failed", "err", err)
		res["disk"] = "fail"
		allOK = false
	} else if !healthy {
		res["disk"] = "fail"
		allOK = false
	} else {
		res["disk"] = "ok"
	}

	if !allOK {
		c.JSON(503, res)
		return
	}
	c.JSON(200, res)
}
