package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/http/handler"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
)

type Handlers struct {
	Auth     *handler.AuthHandler
	Link     *handler.LinkHandler
	Redirect *handler.RedirectHandler
}

func NewRouter(h Handlers, jwtSecret string, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))

	// Health endpoints (no auth).
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/readyz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Public redirect hot path.
	r.GET("/:code", h.Redirect.Redirect)

	// Auth endpoints (no JWT required).
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
	}

	// Protected API endpoints.
	api := r.Group("/api", middleware.JWT(jwtSecret))
	{
		api.POST("/links", h.Link.Create)
		api.GET("/links", h.Link.List)
		api.GET("/links/:code/stats", h.Link.GetStats)
	}

	return r
}
