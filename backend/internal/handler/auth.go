package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/nost3a/PicoOffice/internal/middleware"
	"github.com/nost3a/PicoOffice/internal/model"
	"github.com/nost3a/PicoOffice/internal/service"
)

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// Register: first user becomes admin, others are normal users
func Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Password) < 6 {
		c.JSON(400, gin.H{"error": "username>=3, password>=6"})
		return
	}

	var cnt int64
	db.Model(&model.User{}).Count(&cnt)
	role := "user"
	if cnt == 0 {
		role = "admin"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "hash password failed"})
		return
	}
	nick := req.Nickname
	if nick == "" {
		nick = req.Username
	}
	u := model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     nick,
		Role:         role,
	}
	// first admin gets 10GB; others get gorm default 1GB
	if role == "admin" {
		u.QuotaBytes = 10737418240
	}
	if err := db.Create(&u).Error; err != nil {
		c.JSON(409, gin.H{"error": "username already exists"})
		return
	}
	token, _ := middleware.IssueToken(u.ID, u.Username, u.Role)
	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id": u.ID, "username": u.Username, "role": u.Role, "nickname": u.Nickname,
		},
	})
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad body"})
		return
	}
	var u model.User
	if err := db.Where("username = ?", req.Username).First(&u).Error; err != nil {
		c.JSON(401, gin.H{"error": "wrong username or password"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(401, gin.H{"error": "wrong username or password"})
		return
	}
	// audit on login success; skip failure (user unknown, no owner)
	service.LogAudit(db, u.ID, "login", "/api/auth/login", c.ClientIP())
	token, _ := middleware.IssueToken(u.ID, u.Username, u.Role)
	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id": u.ID, "username": u.Username, "role": u.Role, "nickname": u.Nickname,
		},
	})
}

// Me current user info
func Me(c *gin.Context) {
	uid := c.GetUint("uid")
	var u model.User
	if err := db.First(&u, uid).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	c.JSON(200, gin.H{
		"id": u.ID, "username": u.Username, "role": u.Role, "nickname": u.Nickname,
	})
}
