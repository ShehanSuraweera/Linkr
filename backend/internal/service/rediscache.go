package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"golang.org/x/sync/singleflight"

	"github.com/shehansuraweera/linkr/internal/domain"
)

// linkFetcher is the minimal DB interface RedisCache needs.
type linkFetcher interface {
	GetByCode(ctx context.Context, code string) (domain.Link, error)
}

// RedisCache is a distributed link cache backed by Redis.
//
// Circuit breaker: after 5 consecutive Redis errors the breaker opens and
// all requests skip Redis entirely for 30 s, falling straight through to
// singleflight → PostgreSQL. No 500 ms timeouts during a Redis outage.
// After 30 s the breaker half-opens and tries one request; success closes it.
//
// On a Redis error that does not yet trip the breaker the cache fails open —
// it logs a warning and reads from PostgreSQL — so the service degrades to
// slower responses rather than an outage.
type RedisCache struct {
	client  *redis.Client
	links   linkFetcher
	ttl     time.Duration
	group   singleflight.Group
	breaker *gobreaker.CircuitBreaker
	logger  *slog.Logger
}

// NewRedisCache accepts an already-connected *redis.Client so the caller can
// share it between the cache and the rate limiter. URL parsing and the initial
// ping live in main.go via ConnectRedis.
func NewRedisCache(links linkFetcher, client *redis.Client, ttl time.Duration, logger *slog.Logger) *RedisCache {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "redis-cache",
		MaxRequests: 1,            // 1 probe request in half-open state
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Warn("redis circuit breaker", "from", from.String(), "to", to.String())
		},
	})
	return &RedisCache{client: client, links: links, ttl: ttl, breaker: cb, logger: logger}
}

const keyPrefix = "link:"

// Get returns the link for code.
//
//	Circuit closed, Redis hit  → return immediately (zero DB round-trips).
//	Circuit closed, Redis miss → singleflight → PostgreSQL → store in Redis → return.
//	Circuit closed, Redis slow → 500 ms timeout → counts as failure → may trip breaker.
//	Circuit open               → skip Redis entirely → singleflight → PostgreSQL → return.
func (rc *RedisCache) Get(ctx context.Context, code string) (domain.Link, error) {
	key := keyPrefix + code

	// Attempt Redis via circuit breaker.
	var cached []byte
	_, cbErr := rc.breaker.Execute(func() (any, error) {
		b, err := rc.client.Get(ctx, key).Bytes()
		if errors.Is(err, redis.Nil) {
			return nil, nil // miss is not a failure
		}
		if err != nil {
			return nil, err // connection/timeout error — counts toward breaker
		}
		cached = b
		return nil, nil
	})

	if cbErr != nil {
		// Only log real errors; ErrOpenState / ErrTooManyRequests are silent.
		if !errors.Is(cbErr, gobreaker.ErrOpenState) &&
			!errors.Is(cbErr, gobreaker.ErrTooManyRequests) {
			rc.logger.Warn("redis get error, falling through to db", "code", code, "err", cbErr)
		}
	}

	if cached != nil {
		var link domain.Link
		if err := json.Unmarshal(cached, &link); err == nil {
			return link, nil
		}
	}

	// Cache miss or Redis unavailable — hit PostgreSQL via singleflight.
	// Use a detached context so a cancelled caller (client disconnect) does not
	// abort all coalesced requests for the same code and defeat the stampede guard.
	v, err, _ := rc.group.Do(code, func() (any, error) {
		dCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		link, dbErr := rc.links.GetByCode(dCtx, code)
		if dbErr != nil {
			return nil, dbErr
		}
		// Best-effort write back to Redis; failure is non-fatal.
		if b, mErr := json.Marshal(link); mErr == nil {
			_, _ = rc.breaker.Execute(func() (any, error) {
				return nil, rc.client.Set(dCtx, key, b, rc.ttl).Err()
			})
		}
		return link, nil
	})
	if err != nil {
		return domain.Link{}, err
	}
	return v.(domain.Link), nil
}

// Invalidate removes a code from Redis — call on link update, delete, or pause.
func (rc *RedisCache) Invalidate(ctx context.Context, code string) {
	if err := rc.client.Del(ctx, keyPrefix+code).Err(); err != nil {
		rc.logger.Warn("redis del error", "code", code, "err", err)
	}
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}
