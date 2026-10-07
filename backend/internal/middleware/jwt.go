package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte(getEnv("PICO_JWT_SECRET", "picooffice-dev-secret"))

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// IssueToken signs a 3-day token after login
func IssueToken(userID uint, username string, role string) (string, error) {
	return IssueTokenTTL(userID, username, role, 72*time.Hour)
}

// IssueTokenTTL signs access token with custom TTL, used by refresh flow
func IssueTokenTTL(userID uint, username string, role string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"uid":      userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(ttl).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(secret)
}

// ParseToken verifies signature; ws uses it for query token
func ParseToken(s string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(s, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// JWT middleware; Bearer header first, query token= fallback (for ws)
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokStr := ""
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokStr = strings.TrimPrefix(auth, "Bearer ")
		}
		if tokStr == "" {
			tokStr = c.Query("token")
		}
		if tokStr == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}
		claims, err := ParseToken(tokStr)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		c.Set("uid", uint(claims["uid"].(float64)))
		c.Set("username", claims["username"].(string))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

// AdminRequired middleware on /api/admin/*
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(403, gin.H{"error": "admin only"})
			return
		}
		c.Next()
	}
}
