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

// CreateLinkRequest is the request body for creating a short link.
type CreateLinkRequest struct {
	URL       string `json:"url"        example:"https://example.com/very/long/path"`
	Alias     string `json:"alias"      example:"my-link"`
	ExpiresAt string `json:"expires_at" example:"2026-12-31T23:59:59Z"`
}

// LinkResponse is a single link in API responses.
type LinkResponse struct {
	ID          int64      `json:"id"                    example:"1"`
	ShortCode   string     `json:"short_code"            example:"aB3xY7z"`
	OriginalURL string     `json:"original_url"          example:"https://example.com/very/long/path"`
	CreatedAt   time.Time  `json:"created_at"            example:"2026-06-25T10:00:00Z"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"  example:"2026-12-31T23:59:59Z"`
	IsActive    bool       `json:"is_active"             example:"true"`
}

// ListLinksResponse is the paginated list response.
type ListLinksResponse struct {
	Items      []LinkResponse `json:"items"`
	HasMore    bool           `json:"has_more"              example:"false"`
	NextCursor string         `json:"next_cursor,omitempty" example:"eyJjYSI6Ij..."`
}

// StatsResponse is the per-link click analytics response.
type StatsResponse struct {
	TotalClicks int64                  `json:"total_clicks" example:"42"`
	Daily       []domain.DailyClickStat `json:"daily"`
}

func toLinkResponse(l domain.Link) LinkResponse {
	return LinkResponse{
		ID:          l.ID,
		ShortCode:   l.ShortCode,
		OriginalURL: l.OriginalURL,
		CreatedAt:   l.CreatedAt,
		ExpiresAt:   l.ExpiresAt,
		IsActive:    l.IsActive,
	}
}

// CreateLink godoc
// @Summary      Create a short link
// @Description  Generates a short code (or uses the supplied alias) and persists the link. Requires auth.
// @Tags         links
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateLinkRequest  true  "Link to shorten"
// @Success      201   {object}  LinkResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse  "alias already taken"
// @Router       /api/links [post]
func (h *LinkHandler) Create(c *gin.Context) {
	userID := middleware.UserIDFrom(c)

	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	// Validate alias first — it's free. URL validation (DNS) comes after.
	code := req.Alias
	if code != "" {
		if err := shortcode.ValidateAlias(code); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
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

	if code != "" {
		link, err := h.links.Create(c.Request.Context(), code, req.URL, userID, expiresAt)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusCreated, toLinkResponse(link))
		return
	}

	// Auto-generate code with collision retry.
	for i := 0; i < 3; i++ {
		var err error
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
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate unique code"})
}

// cursor is the opaque keyset pagination token.
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

// ListLinks godoc
// @Summary      List links
// @Description  Returns the authenticated user's links, newest first, with keyset cursor pagination.
// @Tags         links
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query     int     false  "Page size (1-100, default 20)"  example(20)
// @Param        cursor  query     string  false  "Opaque cursor from previous response"
// @Success      200     {object}  ListLinksResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Router       /api/links [get]
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

	items := make([]LinkResponse, len(links))
	for i, l := range links {
		items[i] = toLinkResponse(l)
	}

	resp := ListLinksResponse{Items: items, HasMore: hasMore}
	if hasMore && len(links) > 0 {
		resp.NextCursor = encodeCursor(links[len(links)-1])
	}
	c.JSON(http.StatusOK, resp)
}

// GetStats godoc
// @Summary      Get click stats for a link
// @Description  Returns total clicks and a per-day breakdown. Served from the pre-aggregated rollup table. Requires auth and ownership.
// @Tags         links
// @Produce      json
// @Security     BearerAuth
// @Param        code  path      string  true  "Short code"  example(aB3xY7z)
// @Success      200   {object}  StatsResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      403   {object}  ErrorResponse  "not the owner"
// @Failure      404   {object}  ErrorResponse
// @Router       /api/links/{code}/stats [get]
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
	c.JSON(http.StatusOK, StatsResponse{
		TotalClicks: stats.TotalClicks,
		Daily:       stats.Daily,
	})
}
