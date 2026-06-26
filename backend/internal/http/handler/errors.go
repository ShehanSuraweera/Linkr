package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shehansuraweera/linkr/internal/domain"
)

// bindJSON decodes the request body into req and writes a sanitised 400 on
// failure. Returns true when binding succeeded.
func bindJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": bindErrMsg(err)})
		return false
	}
	return true
}

// bindErrMsg converts a validator error into a message that is safe to expose:
// field names only, no Go struct paths or internal type names.
func bindErrMsg(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msgs := make([]string, len(ve))
		for i, fe := range ve {
			msgs[i] = fmt.Sprintf("%s: failed %s validation", strings.ToLower(fe.Field()), fe.Tag())
		}
		return strings.Join(msgs, "; ")
	}
	return "invalid request body"
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, domain.ErrAliasTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "alias already taken"})
	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "already exists"})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, domain.ErrUnauthorized):
		c.Header("WWW-Authenticate", `Bearer realm="linkr"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrExpired):
		c.JSON(http.StatusGone, gin.H{"error": "link has expired"})
	case errors.Is(err, domain.ErrInactive):
		c.JSON(http.StatusGone, gin.H{"error": "link is inactive"})
	default:
		_ = c.Error(err) // picked up by the Logger middleware
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
