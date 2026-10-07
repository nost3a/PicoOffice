package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nost3a/PicoOffice/internal/model"
)

// ListFolders all user folders; client builds tree
func ListFolders(c *gin.Context) {
	uid := c.GetUint("uid")
	var list []model.Folder
	if err := db.Where("owner_id = ?", uid).Order("updated_at DESC").Find(&list).Error; err != nil {
		c.JSON(500, gin.H{"error": "list folders failed"})
		return
	}
	c.JSON(200, gin.H{"items": list})
}

type folderReq struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"`
}

// CreateFolder creates folder under a parent
func CreateFolder(c *gin.Context) {
	uid := c.GetUint("uid")
	var req folderReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(400, gin.H{"error": "name required"})
		return
	}
	// parent must be own folder
	if req.ParentID != nil && *req.ParentID > 0 {
		var cnt int64
		db.Model(&model.Folder{}).Where("id = ? AND owner_id = ?", *req.ParentID, uid).Count(&cnt)
		if cnt == 0 {
			c.JSON(400, gin.H{"error": "parent folder not found"})
			return
		}
	}
	f := model.Folder{OwnerID: uid, Name: req.Name, ParentID: req.ParentID}
	if err := db.Create(&f).Error; err != nil {
		c.JSON(500, gin.H{"error": "create folder failed"})
		return
	}
	c.JSON(200, f)
}

// loadFolder loads user's folder; 404 if missing
func loadFolder(c *gin.Context) (*model.Folder, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad id"})
		return nil, false
	}
	var f model.Folder
	if err := db.First(&f, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "folder not found"})
		return nil, false
	}
	if f.OwnerID != c.GetUint("uid") {
		c.JSON(403, gin.H{"error": "no permission"})
		return nil, false
	}
	return &f, true
}

// UpdateFolder renames/moves folder
func UpdateFolder(c *gin.Context) {
	uid := c.GetUint("uid")
	f, ok := loadFolder(c)
	if !ok {
		return
	}
	var req folderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	if req.Name != "" {
		f.Name = req.Name
	}
	if req.ParentID != nil {
		// cannot nest folder under itself
		if *req.ParentID == f.ID {
			c.JSON(400, gin.H{"error": "cannot move into itself"})
			return
		}
		// new parent must be own folder (parent_id=0 = move to root)
		if *req.ParentID > 0 {
			var cnt int64
			db.Model(&model.Folder{}).Where("id = ? AND owner_id = ?", *req.ParentID, uid).Count(&cnt)
			if cnt == 0 {
				c.JSON(400, gin.H{"error": "parent folder not found"})
				return
			}
			// prevent cycle: new parent cannot be a descendant
			if isDescendant(uid, f.ID, *req.ParentID) {
				c.JSON(400, gin.H{"error": "cannot move into a child folder"})
				return
			}
			f.ParentID = req.ParentID
		} else {
			f.ParentID = nil
		}
	}
	if err := db.Save(f).Error; err != nil {
		c.JSON(500, gin.H{"error": "update folder failed"})
		return
	}
	c.JSON(200, f)
}

// isDescendant walks parent_id to check ancestry
func isDescendant(uid uint, ancestorID uint, maybeChild uint) bool {
	cur := maybeChild
	visited := map[uint]bool{}
	for cur != 0 {
		if visited[cur] {
			return false
		}
		visited[cur] = true
		if cur == ancestorID {
			return true
		}
		var f model.Folder
		if err := db.Select("id", "parent_id").First(&f, cur).Error; err != nil {
			return false
		}
		if f.OwnerID != uid || f.ParentID == nil {
			return false
		}
		cur = *f.ParentID
	}
	return false
}

// DeleteFolder: keep docs, null their folder_id; move subfolders to parent
func DeleteFolder(c *gin.Context) {
	f, ok := loadFolder(c)
	if !ok {
		return
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. unlink docs in this folder
		if err := tx.Model(&model.Doc{}).Where("folder_id = ?", f.ID).
			Update("folder_id", nil).Error; err != nil {
			return err
		}
		// 2. move direct subfolders to parent (or root)
		if err := tx.Model(&model.Folder{}).Where("parent_id = ?", f.ID).
			Update("parent_id", f.ParentID).Error; err != nil {
			return err
		}
		// 3. remove the folder itself
		return tx.Delete(&model.Folder{}, f.ID).Error
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "delete folder failed"})
		return
	}
	c.JSON(200, gin.H{"deleted": f.ID, "owner": f.OwnerID})
}
