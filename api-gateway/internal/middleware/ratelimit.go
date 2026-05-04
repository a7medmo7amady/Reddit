package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	r       rate.Limit
	burst   int
}

func newRateLimiter(r rate.Limit, burst int) *rateLimiter {
	rl := &rateLimiter{
		entries: make(map[string]*entry),
		r:       r,
		burst:   burst,
	}
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rl.mu.Lock()
			for key, e := range rl.entries {
				if time.Since(e.lastSeen) > 10*time.Minute {
					delete(rl.entries, key)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	e, ok := rl.entries[key]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(rl.r, rl.burst)}
		rl.entries[key] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

var userLimiter = newRateLimiter(30, 60)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-User-Id")
		if key == "" {
			key = c.ClientIP()
		}
		if !userLimiter.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
