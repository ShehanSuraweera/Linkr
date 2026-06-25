package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/auth"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/repository"
)

type AuthHandler struct {
	users     *repository.UserRepo
	jwtSecret string
}

func NewAuthHandler(users *repository.UserRepo, jwtSecret string) *AuthHandler {
	return &AuthHandler{users: users, jwtSecret: jwtSecret}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	user, err := h.users.Create(c.Request.Context(), req.Email, hash)
	if err != nil {
		respondError(c, err)
		return
	}

	token, err := auth.IssueToken(user.ID, h.jwtSecret)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token, "user_id": user.ID})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.users.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		respondError(c, err)
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.IssueToken(user.ID, h.jwtSecret)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": user.ID})
}
