package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
)

type shareReq struct {
	Perm string `json:"perm"`
}

// canManageDoc only owner or admin can manage share/delete
func canManageDoc(c *gin.Context, d *model.Doc) bool {
	uid := c.GetUint("uid")
	if d.OwnerID == uid || c.GetString("role") == "admin" {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	return false
}

// CreateShare POST /api/docs/:id/share {perm:view|edit}
func CreateShare(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	if !canManageDoc(c, d) {
		return
	}
	var req shareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad body"})
		return
	}
	if req.Perm != "view" && req.Perm != "edit" {
		req.Perm = "view"
	}

	// 16 random bytes -> 32 hex chars
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gen token failed"})
		return
	}
	token := hex.EncodeToString(buf)

	d.Visibility = "shared"
	d.ShareToken = token
	d.SharePerm = req.Perm
	if err := db.Save(d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save share failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"share_token": token,
		"url":         "/s/" + token,
		"perm":        req.Perm,
	})
}

// DeleteShare revokes share
func DeleteShare(c *gin.Context) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	if !canManageDoc(c, d) {
		return
	}
	d.Visibility = "private"
	d.ShareToken = ""
	d.SharePerm = ""
	db.Save(d)
	c.JSON(http.StatusOK, gin.H{"deleted": d.ID})
}

// ListShared lists docs shared by current user
func ListShared(c *gin.Context) {
	uid := c.GetUint("uid")
	var list []model.Doc
	db.Where("owner_id = ? AND visibility = ?", uid, "shared").
		Order("updated_at DESC").Find(&list)
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// PublicShare read-only open by token without JWT
func PublicShare(c *gin.Context) {
	token := c.Param("token")
	var d model.Doc
	if err := db.Where("share_token = ? AND visibility = ?", token, "shared").
		First(&d).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
		return
	}
	// loadDoc: db.First excludes soft-deleted; double-guard here
	if d.DeletedAt.Valid {
		c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"title":      d.Title,
		"doc_kind":   d.DocKind,
		"perm":       d.SharePerm,
		"content":    d.Content,
		"updated_at": d.UpdatedAt,
	})
}
