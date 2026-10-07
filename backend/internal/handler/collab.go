package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/service"
)

// GetCollabSteps GET /api/docs/:id/steps?version=n
// returns steps after version; client replays on doc open.
func GetCollabSteps(c *gin.Context) {
	docID, _ := strconv.Atoi(c.Param("id"))
	after, _ := strconv.Atoi(c.DefaultQuery("version", "0"))
	if docID <= 0 {
		c.JSON(400, gin.H{"error": "bad doc id"})
		return
	}
	latest, steps := service.StepsAfter(db, uint(docID), after)
	c.JSON(200, gin.H{
		"version": latest,
		"steps":   steps,
	})
}
