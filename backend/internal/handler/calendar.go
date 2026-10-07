package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nost3a/PicoOffice/internal/model"
)

// ListEvents events in range; from/to in RFC3339
func ListEvents(c *gin.Context) {
	uid := c.GetUint("uid")
	q := db.Model(&model.CalendarEvent{}).Where("user_id = ?", uid)
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where("end_at >= ?", t)
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where("start_at <= ?", t)
		}
	}
	var list []model.CalendarEvent
	if err := q.Order("start_at ASC").Find(&list).Error; err != nil {
		c.JSON(500, gin.H{"error": "list events failed"})
		return
	}
	c.JSON(200, gin.H{"items": list})
}

type eventReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartAt     string `json:"start_at"`
	EndAt       string `json:"end_at"`
	RemindMin   int    `json:"remind_min"`
}

// CreateEvent creates event
func CreateEvent(c *gin.Context) {
	uid := c.GetUint("uid")
	var req eventReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
		c.JSON(400, gin.H{"error": "title required"})
		return
	}
	start, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		c.JSON(400, gin.H{"error": "start_at must be RFC3339"})
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		c.JSON(400, gin.H{"error": "end_at must be RFC3339"})
		return
	}
	ev := model.CalendarEvent{
		UserID:      uid,
		Title:       req.Title,
		Description: req.Description,
		StartAt:     start,
		EndAt:       end,
		RemindMin:   req.RemindMin,
	}
	if err := db.Create(&ev).Error; err != nil {
		c.JSON(500, gin.H{"error": "create event failed"})
		return
	}
	c.JSON(200, ev)
}

// loadEvent loads current user's event
func loadEvent(c *gin.Context) (*model.CalendarEvent, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "bad id"})
		return nil, false
	}
	var ev model.CalendarEvent
	if err := db.First(&ev, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "event not found"})
		return nil, false
	}
	if ev.UserID != c.GetUint("uid") {
		c.JSON(403, gin.H{"error": "no permission"})
		return nil, false
	}
	return &ev, true
}

// UpdateEvent updates event
func UpdateEvent(c *gin.Context) {
	ev, ok := loadEvent(c)
	if !ok {
		return
	}
	var req eventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	if req.Title != "" {
		ev.Title = req.Title
	}
	if req.Description != "" {
		ev.Description = req.Description
	}
	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			c.JSON(400, gin.H{"error": "start_at must be RFC3339"})
			return
		}
		ev.StartAt = t
	}
	if req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			c.JSON(400, gin.H{"error": "end_at must be RFC3339"})
			return
		}
		ev.EndAt = t
	}
	if req.RemindMin > 0 {
		ev.RemindMin = req.RemindMin
	}
	if err := db.Save(ev).Error; err != nil {
		c.JSON(500, gin.H{"error": "update event failed"})
		return
	}
	c.JSON(200, ev)
}

// DeleteEvent deletes event
func DeleteEvent(c *gin.Context) {
	ev, ok := loadEvent(c)
	if !ok {
		return
	}
	if err := db.Delete(&model.CalendarEvent{}, ev.ID).Error; err != nil {
		c.JSON(500, gin.H{"error": "delete event failed"})
		return
	}
	c.JSON(200, gin.H{"deleted": ev.ID})
}
