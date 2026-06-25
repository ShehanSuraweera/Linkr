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

// ClickEnqueuer is satisfied by *clicks.Pipeline.
type ClickEnqueuer interface {
	Enqueue(e domain.ClickEvent) bool
}

// RedirectHandler serves GET /{code} — the public hot path.
type RedirectHandler struct {
	links  *repository.LinkRepo
	clicks ClickEnqueuer
}

func NewRedirectHandler(links *repository.LinkRepo, clicks ClickEnqueuer) *RedirectHandler {
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

	// Enqueue click asynchronously — never blocks the redirect.
	// If the queue is full the event is dropped gracefully (logged by the pipeline).
	h.clicks.Enqueue(domain.ClickEvent{
		LinkID:    link.ID,
		At:        time.Now(),
		IPHash:    hashIP(c.ClientIP()),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
	})

	c.Redirect(http.StatusFound, link.OriginalURL)
}

func hashIP(ip string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(ip)))[:16]
}
