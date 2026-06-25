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

// RegisterRequest is the request body for register and login.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"s3cr3tpassword"`
}

// TokenResponse is returned on successful register or login.
type TokenResponse struct {
	Token  string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID int64  `json:"user_id" example:"1"`
}

// ErrorResponse is the standard error shape for all 4xx/5xx responses.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a user account and returns a signed JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Credentials"
// @Success      201   {object}  TokenResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse  "email already registered"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
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

	c.JSON(http.StatusCreated, TokenResponse{Token: token, UserID: user.ID})
}

// Login godoc
// @Summary      Log in
// @Description  Validates credentials and returns a signed JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Credentials"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse  "invalid credentials"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req RegisterRequest
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

	c.JSON(http.StatusOK, TokenResponse{Token: token, UserID: user.ID})
}
