package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/auth"
)

const UserIDKey = "user_id"

func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.Header("WWW-Authenticate", `Bearer realm="linkr"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ParseToken(token, secret)
		if err != nil {
			c.Header("WWW-Authenticate", `Bearer realm="linkr", error="invalid_token"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

func UserIDFrom(c *gin.Context) int64 {
	id, _ := c.Get(UserIDKey)
	v, _ := id.(int64)
	return v
}
