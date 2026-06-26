package handler

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/domain"
)

// ClickEnqueuer is satisfied by *clicks.Pipeline.
type ClickEnqueuer interface {
	Enqueue(e domain.ClickEvent) bool
}

// LinkLookup is satisfied by *service.LinkCache.
type LinkLookup interface {
	Get(ctx context.Context, code string) (domain.Link, error)
}

// RedirectHandler serves GET /{code} — the public hot path.
// Most requests are served from the LRU/Redis cache with zero DB round-trips.
type RedirectHandler struct {
	cache          LinkLookup
	clicks         ClickEnqueuer
	cacheMaxAgeSec int // 0 = no Cache-Control header
}

func NewRedirectHandler(cache LinkLookup, clicks ClickEnqueuer, cacheMaxAgeSec int) *RedirectHandler {
	return &RedirectHandler{cache: cache, clicks: clicks, cacheMaxAgeSec: cacheMaxAgeSec}
}

// Redirect godoc
// @Summary      Redirect to original URL
// @Description  Looks up the short code (LRU cache → DB on miss), records a click asynchronously, and issues an HTTP 302. This endpoint is public — no auth required.
// @Tags         redirect
// @Produce      json
// @Param        code  path  string  true  "Short code"  example(aB3xY7z)
// @Success      302   "Redirects to the original URL"
// @Failure      404   {object}  ErrorResponse  "code not found or link deleted"
// @Failure      410   {object}  ErrorResponse  "link expired or inactive"
// @Router       /{code} [get]
func (h *RedirectHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	link, err := h.cache.Get(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		respondError(c, err)
		return
	}

	// GetByCode already filters deleted_at IS NULL, so only active/expiry need checking.
	if !link.IsLive() {
		if !link.IsActive {
			c.JSON(http.StatusGone, gin.H{"error": "link is inactive"})
			return
		}
		c.JSON(http.StatusGone, gin.H{"error": "link has expired"})
		return
	}

	// Enqueue click asynchronously — never blocks the redirect.
	// If the queue is full the event is dropped gracefully (logged by the pipeline).
	h.clicks.Enqueue(domain.ClickEvent{
		LinkID:    link.ID,
		At:        time.Now(),
		IPHash:    hashIP(c.ClientIP()),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	})

	// Let browsers and CDN proxies cache this redirect to reduce origin load.
	if h.cacheMaxAgeSec > 0 {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", h.cacheMaxAgeSec))
	}
	c.Redirect(http.StatusFound, link.OriginalURL)
}

func hashIP(ip string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(ip)))[:16]
}
