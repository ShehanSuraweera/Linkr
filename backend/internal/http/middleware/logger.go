package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"request_id", c.GetString(RequestIDKey),
			"ip", c.ClientIP(),
		}

		if status >= 500 && len(c.Errors) > 0 {
			logger.Error("request", append(attrs, "err", c.Errors.Last().Err)...)
		} else {
			logger.Info("request", attrs...)
		}
	}
}
