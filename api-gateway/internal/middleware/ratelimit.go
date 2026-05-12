package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	rl "api-gateway/pkg/ratelimit"
	"api-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	redisLimiter *rl.Limiter
	rlOnce       sync.Once
)

func SetRedisLimiter(l *rl.Limiter) {
	rlOnce.Do(func() { redisLimiter = l })
}

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type memLimiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	r       rate.Limit
	burst   int
}

func newMemLimiter(r rate.Limit, burst int) *memLimiter {
	ml := &memLimiter{entries: make(map[string]*entry), r: r, burst: burst}
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			ml.mu.Lock()
			for key, e := range ml.entries {
				if time.Since(e.lastSeen) > 10*time.Minute {
					delete(ml.entries, key)
				}
			}
			ml.mu.Unlock()
		}
	}()
	return ml
}

func (ml *memLimiter) allow(key string) bool {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	e, ok := ml.entries[key]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(ml.r, ml.burst)}
		ml.entries[key] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

var fallback = newMemLimiter(30, 60)


func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-User-Id")
		if key == "" {
			key = c.ClientIP()
		}

		allowed := false

		if redisLimiter != nil {
			var err error
			allowed, err = redisLimiter.Allow(context.Background(), key)
			if err != nil {
				logger.Warnf("redis rate limiter error, using fallback: %v\n", err)
				allowed = fallback.allow(key)
			}
		} else {
			allowed = fallback.allow(key)
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
