package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"

	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/repository"
)

// LinkCache wraps the link repository with an in-process expirable LRU cache.
// Used when no Redis is configured (single-instance deployments).
//
// TTL bounds stale data so deleted/updated links expire from cache within ttl,
// rather than living until evicted by size pressure (the old behaviour).
type LinkCache struct {
	cache  *expirable.LRU[string, domain.Link]
	links  *repository.LinkRepo
	group  singleflight.Group
	logger *slog.Logger
}

func NewLinkCache(links *repository.LinkRepo, size int, ttl time.Duration, logger *slog.Logger) (*LinkCache, error) {
	if size <= 0 {
		return nil, fmt.Errorf("cache size must be > 0, got %d", size)
	}
	cache := expirable.NewLRU[string, domain.Link](size, nil, ttl)
	return &LinkCache{cache: cache, links: links, logger: logger}, nil
}

// Get returns the link for code, using the cache and singleflight.
func (lc *LinkCache) Get(ctx context.Context, code string) (domain.Link, error) {
	if link, ok := lc.cache.Get(code); ok {
		return link, nil
	}

	// Detach from the request context so a cancelled caller does not abort all
	// coalesced requests and defeat singleflight's stampede guard.
	v, err, _ := lc.group.Do(code, func() (any, error) {
		dCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		link, dbErr := lc.links.GetByCode(dCtx, code)
		if dbErr != nil {
			return nil, dbErr
		}
		lc.cache.Add(code, link)
		return link, nil
	})
	if err != nil {
		return domain.Link{}, err
	}
	return v.(domain.Link), nil
}

// Invalidate removes a code from the cache (call on link update/delete/pause).
func (lc *LinkCache) Invalidate(_ context.Context, code string) {
	lc.cache.Remove(code)
}
