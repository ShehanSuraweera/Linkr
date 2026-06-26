package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/shehansuraweera/linkr/internal/http/handler"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
)

type Handlers struct {
	Auth      *handler.AuthHandler
	Link      *handler.LinkHandler
	Redirect  *handler.RedirectHandler
	Analytics *handler.AnalyticsHandler
}

// RouterConfig holds the values NewRouter needs from the app config.
// RateLimit is a pre-built middleware so the router never imports Redis.
type RouterConfig struct {
	JWTSecret string
	RateLimit gin.HandlerFunc
}

func NewRouter(h Handlers, cfg RouterConfig, logger *slog.Logger, ready func() error) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Metrics())
	r.Use(cfg.RateLimit)

	// Health endpoints (no auth, no rate limit — Kubernetes probes these).
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/readyz", func(c *gin.Context) {
		if err := ready(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db unavailable"})
			return
		}
		c.Status(http.StatusOK)
	})

	// Prometheus metrics scrape endpoint.
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger UI — available at /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public redirect hot path (must come after fixed routes to avoid shadowing /swagger/*).
	r.GET("/:code", h.Redirect.Redirect)

	// Auth endpoints (no JWT required).
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
	}

	// Auth endpoints that require a valid JWT.
	authProtected := r.Group("/api/auth", middleware.JWT(cfg.JWTSecret))
	{
		authProtected.GET("/me", h.Auth.Me)
	}

	// Protected API endpoints.
	api := r.Group("/api", middleware.JWT(cfg.JWTSecret))
	{
		api.POST("/links", h.Link.Create)
		api.GET("/links", h.Link.List)
		api.GET("/links/:code/stats", h.Link.GetStats)
		api.PATCH("/links/:id", h.Link.Patch)
		api.DELETE("/links/:id", h.Link.Delete)
		api.GET("/analytics/overview", h.Analytics.Overview)
	}

	return r
}
