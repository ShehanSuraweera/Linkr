package service

import (
	"context"
	"fmt"
	"log/slog"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"

	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/repository"
)

// LinkCache wraps the link repository with an in-process LRU cache.
// On a cache hit the redirect serves with zero DB round-trips.
// singleflight collapses concurrent misses for the same code into one DB read.
//
// Scale note: this cache is per-instance. With multiple API replicas each
// has its own cache — a mutation (delete/pause) only invalidates locally.
// The next step is a shared Redis cache or short TTLs. Documented in DECISIONS.md.
type LinkCache struct {
	cache  *lru.Cache[string, domain.Link]
	links  *repository.LinkRepo
	group  singleflight.Group
	logger *slog.Logger
}

func NewLinkCache(links *repository.LinkRepo, size int, logger *slog.Logger) (*LinkCache, error) {
	c, err := lru.New[string, domain.Link](size)
	if err != nil {
		return nil, fmt.Errorf("create lru cache: %w", err)
	}
	return &LinkCache{cache: c, links: links, logger: logger}, nil
}

// Get returns the link for code, using the cache and singleflight.
func (lc *LinkCache) Get(ctx context.Context, code string) (domain.Link, error) {
	if link, ok := lc.cache.Get(code); ok {
		return link, nil
	}

	// Deduplicate concurrent misses for the same code.
	v, err, _ := lc.group.Do(code, func() (any, error) {
		link, err := lc.links.GetByCode(ctx, code)
		if err != nil {
			return nil, err
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
func (lc *LinkCache) Invalidate(code string) {
	lc.cache.Remove(code)
}
