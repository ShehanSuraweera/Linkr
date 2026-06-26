package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "linkr_http_requests_total",
		Help: "Total HTTP requests, labelled by method, route pattern, and status code.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "linkr_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)

// Metrics returns Gin middleware that records per-route Prometheus metrics.
// Register a click-queue depth gauge separately via RegisterClickQueueGauge
// so the pipeline does not need to import this package.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		path := c.FullPath() // route pattern (e.g. "/:code"), not raw URL
		if path == "" {
			path = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).
			Observe(time.Since(start).Seconds())
	}
}

// RegisterClickQueueGauge creates a Prometheus gauge that reports the current
// click pipeline buffer depth on every scrape. Call once at startup.
func RegisterClickQueueGauge(depthFn func() int) {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "linkr_click_queue_depth",
		Help: "Number of click events currently buffered in the async pipeline.",
	}, func() float64 { return float64(depthFn()) })
}
