package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// ListTags user's tags
func ListTags(c *gin.Context) {
	uid := c.GetUint("uid")
	var list []model.Tag
	if err := db.Where("owner_id = ?", uid).Order("id DESC").Find(&list).Error; err != nil {
		c.JSON(500, gin.H{"error": "list tags failed"})
		return
	}
	c.JSON(200, gin.H{"items": list})
}

type createTagReq struct {
	Name string `json:"name"`
}

// CreateTag creates tag
func CreateTag(c *gin.Context) {
	uid := c.GetUint("uid")
	var req createTagReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(400, gin.H{"error": "name required"})
		return
	}
	t := model.Tag{OwnerID: uid, Name: req.Name}
	if err := db.Create(&t).Error; err != nil {
		c.JSON(500, gin.H{"error": "create tag failed"})
		return
	}
	c.JSON(200, t)
}

// DeleteTag removes tag and DocTag links
func DeleteTag(c *gin.Context) {
	uid := c.GetUint("uid")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad id"})
		return
	}
	var t model.Tag
	if err := db.First(&t, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "tag not found"})
		return
	}
	if t.OwnerID != uid {
		c.JSON(403, gin.H{"error": "no permission"})
		return
	}
	db.Transaction(func(tx *gorm.DB) error {
		tx.Where("tag_id = ?", t.ID).Delete(&model.DocTag{})
		return tx.Delete(&model.Tag{}, t.ID).Error
	})
	c.JSON(200, gin.H{"deleted": t.ID})
}

type setDocTagsReq struct {
	TagIDs []uint `json:"tag_ids"`
}

// SetDocTags replaces doc tags
func SetDocTags(c *gin.Context) {
	uid := c.GetUint("uid")
	docID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad doc id"})
		return
	}
	var d model.Doc
	if err := db.First(&d, docID).Error; err != nil {
		c.JSON(404, gin.H{"error": "doc not found"})
		return
	}
	if d.OwnerID != uid {
		c.JSON(403, gin.H{"error": "no permission"})
		return
	}
	var req setDocTagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	// only own tags
	validIDs := make([]uint, 0, len(req.TagIDs))
	for _, tid := range req.TagIDs {
		var cnt int64
		db.Model(&model.Tag{}).Where("id = ? AND owner_id = ?", tid, uid).Count(&cnt)
		if cnt > 0 {
			validIDs = append(validIDs, tid)
		}
	}
	db.Transaction(func(tx *gorm.DB) error {
		tx.Where("doc_id = ?", d.ID).Delete(&model.DocTag{})
		for _, tid := range validIDs {
			if err := tx.Create(&model.DocTag{DocID: d.ID, TagID: tid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	c.JSON(200, gin.H{"doc_id": d.ID, "tag_ids": validIDs})
}
