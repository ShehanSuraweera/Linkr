package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
	"github.com/shehansuraweera/linkr/internal/usecase"
)

// CacheInvalidator is satisfied by *service.TieredCache and *service.LinkCache.
type CacheInvalidator interface {
	Invalidate(ctx context.Context, code string)
}

type LinkHandler struct {
	uc    *usecase.LinkUsecase
	cache CacheInvalidator // nil when no cache is configured
}

func NewLinkHandler(uc *usecase.LinkUsecase, cache CacheInvalidator) *LinkHandler {
	return &LinkHandler{uc: uc, cache: cache}
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
	TotalClicks int64      `json:"total_clicks"          example:"42"`
}

// ListLinksResponse is the paginated list response.
type ListLinksResponse struct {
	Items      []LinkResponse `json:"items"`
	HasMore    bool           `json:"has_more"              example:"false"`
	NextCursor string         `json:"next_cursor,omitempty" example:"eyJjYSI6Ij..."`
}

// StatsResponse is the per-link click analytics response.
type StatsResponse struct {
	TotalClicks int64                   `json:"total_clicks" example:"42"`
	Daily       []domain.DailyClickStat `json:"daily"`
	Devices     []domain.DeviceStat     `json:"devices"`
	Browsers    []domain.BrowserStat    `json:"browsers"`
	Referers    []domain.RefererStat    `json:"referers"`
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

func toLinkSummaryResponse(s domain.LinkSummary) LinkResponse {
	r := toLinkResponse(s.Link)
	r.TotalClicks = s.TotalClicks
	return r
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
	if !bindJSON(c, &req) {
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be RFC3339"})
			return
		}
		expiresAt = &t
	}

	link, err := h.uc.Create(c.Request.Context(), usecase.CreateLinkInput{
		URL:       req.URL,
		Alias:     req.Alias,
		ExpiresAt: expiresAt,
		UserID:    userID,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toLinkResponse(link))
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

	search := strings.TrimSpace(c.Query("q"))

	links, hasMore, err := h.uc.List(c.Request.Context(), userID, cursorCreatedAt, cursorID, int32(limit), search)
	if err != nil {
		respondError(c, err)
		return
	}

	items := make([]LinkResponse, len(links))
	for i, l := range links {
		items[i] = toLinkSummaryResponse(l)
	}

	resp := ListLinksResponse{Items: items, HasMore: hasMore}
	if hasMore && len(links) > 0 {
		resp.NextCursor = encodeCursor(links[len(links)-1].Link)
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

	stats, err := h.uc.GetStats(c.Request.Context(), code, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, StatsResponse{
		TotalClicks: stats.TotalClicks,
		Daily:       stats.Daily,
		Devices:     stats.Devices,
		Browsers:    stats.Browsers,
		Referers:    stats.Referers,
	})
}

// PatchLinkRequest is the request body for updating a link.
// ExpiresAt uses json.RawMessage to distinguish absent (no change) from null (clear).
type PatchLinkRequest struct {
	IsActive  *bool           `json:"is_active"`
	ExpiresAt json.RawMessage `json:"expires_at" swaggertype:"string" example:"2026-12-31T23:59:59Z"`
}

// PatchLink godoc
// @Summary      Update a link
// @Description  Toggles is_active and/or updates expires_at. Send null for expires_at to clear it. Requires auth and ownership.
// @Tags         links
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        code  path      string          true  "Short code"
// @Param        body  body      PatchLinkRequest  true  "Fields to update"
// @Success      200   {object}  LinkResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /api/links/{code} [patch]
func (h *LinkHandler) Patch(c *gin.Context) {
	userID := middleware.UserIDFrom(c)
	code := c.Param("code")

	var req PatchLinkRequest
	if !bindJSON(c, &req) {
		return
	}

	var (
		setExpiresAt bool
		expiresAt    *time.Time
	)
	if req.ExpiresAt != nil {
		setExpiresAt = true
		if string(req.ExpiresAt) != "null" {
			var t time.Time
			if err := json.Unmarshal(req.ExpiresAt, &t); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be RFC3339 or null"})
				return
			}
			expiresAt = &t
		}
	}

	link, err := h.uc.Update(c.Request.Context(), usecase.UpdateLinkInput{
		Code:         code,
		UserID:       userID,
		IsActive:     req.IsActive,
		SetExpiresAt: setExpiresAt,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	if h.cache != nil {
		h.cache.Invalidate(c.Request.Context(), link.ShortCode)
	}
	c.JSON(http.StatusOK, toLinkResponse(link))
}

// DeleteLink godoc
// @Summary      Delete a link
// @Description  Soft-deletes a link owned by the authenticated user.
// @Tags         links
// @Security     BearerAuth
// @Param        code  path  string  true  "Short code"
// @Success      204  "No content"
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/links/{code} [delete]
func (h *LinkHandler) Delete(c *gin.Context) {
	userID := middleware.UserIDFrom(c)
	code := c.Param("code")

	if err := h.uc.Delete(c.Request.Context(), code, userID); err != nil {
		respondError(c, err)
		return
	}
	if h.cache != nil {
		h.cache.Invalidate(c.Request.Context(), code)
	}
	c.Status(http.StatusNoContent)
}
