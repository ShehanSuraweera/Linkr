package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b) //nolint:errcheck
			id = hex.EncodeToString(b)
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}
