package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
	"github.com/shehansuraweera/linkr/internal/usecase"
)

type AnalyticsHandler struct {
	uc *usecase.LinkUsecase
}

func NewAnalyticsHandler(uc *usecase.LinkUsecase) *AnalyticsHandler {
	return &AnalyticsHandler{uc: uc}
}

// Overview godoc
// @Summary      Analytics overview
// @Description  Returns aggregate click stats across all links owned by the authenticated user.
// @Tags         analytics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  domain.OverviewStats
// @Failure      401  {object}  ErrorResponse
// @Router       /api/analytics/overview [get]
func (h *AnalyticsHandler) Overview(c *gin.Context) {
	userID := middleware.UserIDFrom(c)
	stats, err := h.uc.GetOverview(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
