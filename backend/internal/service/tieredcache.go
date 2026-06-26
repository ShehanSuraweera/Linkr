package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/shehansuraweera/linkr/internal/domain"
)

// TieredCache is a two-level cache:
//
//	L1 — per-instance expirable LRU (in-process, no network, sub-microsecond)
//	L2 — RedisCache               (distributed, shared across all replicas)
//
// L1 uses a short TTL so stale data across replicas is bounded.
// Redis remains the source of cross-replica consistency — L1 is a local shield only.
//
// When Redis is healthy:  L1 absorbs repeated hits; L2 protects PostgreSQL.
// When Redis crashes:     L1 still serves hot links from memory for up to l1TTL,
//
//	so PostgreSQL only absorbs cold or recently-expired entries.
type TieredCache struct {
	l1     *expirable.LRU[string, domain.Link]
	l2     *RedisCache
	logger *slog.Logger
}

func NewTieredCache(l1Size int, l1TTL time.Duration, l2 *RedisCache, logger *slog.Logger) (*TieredCache, error) {
	cache := expirable.NewLRU[string, domain.Link](l1Size, nil, l1TTL)
	if cache == nil {
		return nil, fmt.Errorf("create l1 lru: invalid size")
	}
	return &TieredCache{l1: cache, l2: l2, logger: logger}, nil
}

// Get checks L1 first, then L2 (Redis → PostgreSQL).
// L2 results are backfilled into L1 so the next request for the same
// code is served from memory — even if Redis is currently down.
// L1 entries expire after l1TTL, bounding stale data across replicas.
func (tc *TieredCache) Get(ctx context.Context, code string) (domain.Link, error) {
	if link, ok := tc.l1.Get(code); ok {
		return link, nil
	}

	link, err := tc.l2.Get(ctx, code)
	if err != nil {
		return domain.Link{}, err
	}

	tc.l1.Add(code, link)
	return link, nil
}

// Invalidate removes the code from both tiers.
// Redis is invalidated immediately (consistent across all replicas).
// L1 expires naturally after l1TTL on other replicas that don't handle this call.
func (tc *TieredCache) Invalidate(ctx context.Context, code string) {
	tc.l1.Remove(code)
	tc.l2.Invalidate(ctx, code)
}
