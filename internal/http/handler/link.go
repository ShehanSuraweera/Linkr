package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
	"github.com/shehansuraweera/linkr/internal/repository"
	"github.com/shehansuraweera/linkr/internal/service"
	"github.com/shehansuraweera/linkr/internal/shortcode"
)

type LinkHandler struct {
	links  *repository.LinkRepo
	clicks *repository.ClickRepo
}

func NewLinkHandler(links *repository.LinkRepo, clicks *repository.ClickRepo) *LinkHandler {
	return &LinkHandler{links: links, clicks: clicks}
}

type createLinkRequest struct {
	URL       string `json:"url" binding:"required"`
	Alias     string `json:"alias"`
	ExpiresAt string `json:"expires_at"` // RFC3339 or empty
}

type linkResponse struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

func toLinkResponse(l domain.Link) linkResponse {
	return linkResponse{
		ID:          l.ID,
		ShortCode:   l.ShortCode,
		OriginalURL: l.OriginalURL,
		CreatedAt:   l.CreatedAt,
		ExpiresAt:   l.ExpiresAt,
		IsActive:    l.IsActive,
	}
}

func (h *LinkHandler) Create(c *gin.Context) {
	userID := middleware.UserIDFrom(c)

	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.ValidateURL(req.URL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be RFC3339"})
			return
		}
		if t.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be in the future"})
			return
		}
		expiresAt = &t
	}

	code := req.Alias
	if code != "" {
		if err := shortcode.ValidateAlias(code); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		var err error
		for i := 0; i < 3; i++ {
			code, err = shortcode.Generate()
			if err != nil {
				respondError(c, err)
				return
			}
			link, createErr := h.links.Create(c.Request.Context(), code, req.URL, userID, expiresAt)
			if createErr == nil {
				c.JSON(http.StatusCreated, toLinkResponse(link))
				return
			}
			if createErr != domain.ErrConflict {
				respondError(c, createErr)
				return
			}
			// collision — retry with a new code
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate unique code"})
		return
	}

	link, err := h.links.Create(c.Request.Context(), code, req.URL, userID, expiresAt)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toLinkResponse(link))
}

// cursor is an opaque base64-encoded JSON token for keyset pagination.
type cursor struct {
	CreatedAt time.Time `json:"ca"`
	ID        int64     `json:"id"`
}

func encodeCursor(l domain.Link) string {
	b, _ := json.Marshal(cursor{CreatedAt: l.CreatedAt, ID: l.ID})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (time.Time, int64, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid cursor")
	}
	var cur cursor
	if err := json.Unmarshal(b, &cur); err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid cursor")
	}
	return cur.CreatedAt, cur.ID, nil
}

func (h *LinkHandler) List(c *gin.Context) {
	userID := middleware.UserIDFrom(c)

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	var (
		cursorCreatedAt time.Time
		cursorID        int64
	)
	if cur := c.Query("cursor"); cur != "" {
		cursorCreatedAt, cursorID, err = decodeCursor(cur)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
	}

	links, hasMore, err := h.links.List(c.Request.Context(), userID, cursorCreatedAt, cursorID, int32(limit))
	if err != nil {
		respondError(c, err)
		return
	}

	items := make([]linkResponse, len(links))
	for i, l := range links {
		items[i] = toLinkResponse(l)
	}

	resp := gin.H{"items": items, "has_more": hasMore}
	if hasMore && len(links) > 0 {
		resp["next_cursor"] = encodeCursor(links[len(links)-1])
	}
	c.JSON(http.StatusOK, resp)
}

func (h *LinkHandler) GetStats(c *gin.Context) {
	userID := middleware.UserIDFrom(c)
	code := c.Param("code")

	link, err := h.links.GetByCode(c.Request.Context(), code)
	if err != nil {
		respondError(c, err)
		return
	}
	if link.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	stats, err := h.clicks.GetStats(c.Request.Context(), link.ID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
