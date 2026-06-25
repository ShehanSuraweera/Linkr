package handler

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/repository"
)

// RedirectHandler serves GET /{code} — the public hot path.
// Click recording is synchronous here (Phase 3); Phase 4 swaps in async.
type RedirectHandler struct {
	links  *repository.LinkRepo
	clicks *repository.ClickRepo
}

func NewRedirectHandler(links *repository.LinkRepo, clicks *repository.ClickRepo) *RedirectHandler {
	return &RedirectHandler{links: links, clicks: clicks}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	link, err := h.links.GetByCode(c.Request.Context(), code)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		respondError(c, err)
		return
	}

	if !link.IsLive() {
		if link.DeletedAt != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		if !link.IsActive {
			c.JSON(http.StatusGone, gin.H{"error": "link is inactive"})
			return
		}
		c.JSON(http.StatusGone, gin.H{"error": "link has expired"})
		return
	}

	// Record click synchronously (replaced with async in Phase 4).
	event := domain.ClickEvent{
		LinkID:    link.ID,
		At:        time.Now(),
		IPHash:    hashIP(c.ClientIP()),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	}
	_ = h.clicks.FlushBatch(c.Request.Context(), []domain.ClickEvent{event})

	c.Redirect(http.StatusFound, link.OriginalURL)
}

func hashIP(ip string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(ip)))[:16]
}
