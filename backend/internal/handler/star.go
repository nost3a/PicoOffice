package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
)

// starSet sets star state; owner only
func starSet(c *gin.Context, starred bool) {
	d, ok := loadDoc(c)
	if !ok {
		return
	}
	if d.OwnerID != c.GetUint("uid") {
		c.JSON(403, gin.H{"error": "no permission"})
		return
	}
	d.Starred = starred
	if err := db.Model(&model.Doc{}).Where("id = ?", d.ID).
		Update("starred", starred).Error; err != nil {
		c.JSON(500, gin.H{"error": "update starred failed"})
		return
	}
	c.JSON(200, gin.H{"id": d.ID, "starred": starred})
}

// StarDoc stars a doc
func StarDoc(c *gin.Context) { starSet(c, true) }

// UnstarDoc un-stars a doc
func UnstarDoc(c *gin.Context) { starSet(c, false) }
