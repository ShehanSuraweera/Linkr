package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	redis_rate "github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RateLimit returns per-IP rate limiting middleware.
//
// If client is non-nil it uses a Redis sliding-window counter — limits are
// enforced globally across all replicas (true distributed rate limiting).
//
// If client is nil, or if Redis is unavailable at runtime, it falls back to
// a per-instance in-process token bucket. The fallback is automatic and silent
// so a Redis outage never takes down the rate limiter.
//
// The cleanup goroutine exits when ctx is cancelled (i.e. on server shutdown).
func RateLimit(ctx context.Context, rps float64, burst int, client *redis.Client) gin.HandlerFunc {
	// Redis-backed limiter (shared across replicas).
	var redisLimiter *redis_rate.Limiter
	if client != nil {
		redisLimiter = redis_rate.NewLimiter(client)
	}

	// Per-instance fallback (also the sole limiter when client == nil).
	type visitor struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				for ip, v := range visitors {
					if time.Since(v.lastSeen) > 3*time.Minute {
						delete(visitors, ip)
					}
				}
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	localAllow := func(ip string) bool {
		mu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		lim := v.limiter
		mu.Unlock()
		return lim.Allow()
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if redisLimiter != nil {
			limit := redis_rate.Limit{Rate: int(rps), Burst: burst, Period: time.Second}
			res, err := redisLimiter.Allow(c.Request.Context(), "rl:"+ip, limit)
			if err == nil {
				// Redis responded — enforce the global decision.
				if res.Allowed == 0 {
					c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
					c.Abort()
					return
				}
				c.Next()
				return
			}
			// Redis error — fall through to per-instance limiter silently.
		}

		if !localAllow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}
