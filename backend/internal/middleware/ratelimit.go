package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// login/register rate limit: 10/min per IP, else 429
type ipBucket struct {
	count int
	start time.Time
}

var (
	rlMu  sync.Mutex
	rlMap = map[string]*ipBucket{}
)

// RateLimitAuth on /api/auth/login and /api/auth/register
func RateLimitAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rlMu.Lock()
		// lazy cleanup: drop expired buckets along the way
		for k, v := range rlMap {
			if now.Sub(v.start) > time.Minute {
				delete(rlMap, k)
			}
		}
		b, ok := rlMap[ip]
		if !ok {
			b = &ipBucket{start: now}
			rlMap[ip] = b
		}
		if now.Sub(b.start) > time.Minute {
			b.count = 0
			b.start = now
		}
		b.count++
		over := b.count > 10
		rlMu.Unlock()

		if over {
			c.AbortWithStatusJSON(429, gin.H{"error": "too many attempts"})
			return
		}
		c.Next()
	}
}
