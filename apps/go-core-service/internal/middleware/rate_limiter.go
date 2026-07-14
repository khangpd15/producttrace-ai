package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/redis/go-redis/v9"
)

type ipLimiter struct {
	requests int
	resetAt  time.Time
}

type RateLimiter struct {
	redisClient *redis.Client
	mu          sync.Mutex
	inMemory    map[string]*ipLimiter
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		inMemory:    make(map[string]*ipLimiter),
	}
}

func (rl *RateLimiter) Limit(limit int, period time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if rl.redisClient != nil {
			ctx := c.Request.Context()
			key := "rate_limit:trace_search:" + ip
			val, err := rl.redisClient.Incr(ctx, key).Result()
			if err == nil {
				if val == 1 {
					rl.redisClient.Expire(ctx, key, period)
				}
				if val > int64(limit) {
					apperror.HandleError(c, apperror.NewTooManyRequests("Too many requests. Please try again later."))
					c.Abort()
					return
				}
				c.Next()
				return
			}
		}

		// Fallback to thread-safe in-memory rate limiter
		rl.mu.Lock()
		now := time.Now()
		lim, exists := rl.inMemory[ip]
		if !exists || now.After(lim.resetAt) {
			lim = &ipLimiter{
				requests: 0,
				resetAt:  now.Add(period),
			}
			rl.inMemory[ip] = lim
		}
		lim.requests++
		reqs := lim.requests
		rl.mu.Unlock()

		if reqs > limit {
			apperror.HandleError(c, apperror.NewTooManyRequests("Too many requests. Please try again later."))
			c.Abort()
			return
		}
		c.Next()
	}
}
