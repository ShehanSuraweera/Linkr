package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
	"github.com/shehansuraweera/linkr/internal/usecase"
)

type AnalyticsHandler struct {
	uc *usecase.LinkUsecase
}

func NewAnalyticsHandler(uc *usecase.LinkUsecase) *AnalyticsHandler {
	return &AnalyticsHandler{uc: uc}
}

// OverviewResponse is the HTTP response shape for the analytics overview endpoint.
// Keeping it separate from domain.OverviewStats decouples the wire format from
// the internal model so either can evolve independently.
type OverviewResponse struct {
	TotalLinks  int64                    `json:"total_links"`
	ActiveLinks int64                    `json:"active_links"`
	TotalClicks int64                    `json:"total_clicks"`
	Daily       []domain.DailyClickStat  `json:"daily"`
	Devices     []domain.DeviceStat      `json:"devices"`
	Browsers    []domain.BrowserStat     `json:"browsers"`
	Referers    []domain.RefererStat     `json:"referers"`
	TopLinks    []domain.LinkClickStat   `json:"top_links"`
}

func toOverviewResponse(s domain.OverviewStats) OverviewResponse {
	return OverviewResponse{
		TotalLinks:  s.TotalLinks,
		ActiveLinks: s.ActiveLinks,
		TotalClicks: s.TotalClicks,
		Daily:       s.Daily,
		Devices:     s.Devices,
		Browsers:    s.Browsers,
		Referers:    s.Referers,
		TopLinks:    s.TopLinks,
	}
}

// Overview godoc
// @Summary      Analytics overview
// @Description  Returns aggregate click stats across all links owned by the authenticated user.
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  OverviewResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/analytics/overview [get]
func (h *AnalyticsHandler) Overview(c *gin.Context) {
	userID := middleware.UserIDFrom(c)
	stats, err := h.uc.GetOverview(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toOverviewResponse(stats))
}
