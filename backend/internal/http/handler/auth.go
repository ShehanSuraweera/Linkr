package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/http/middleware"
	"github.com/shehansuraweera/linkr/internal/usecase"
)

type AuthHandler struct {
	uc *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

// RegisterRequest is the request body for account creation.
type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required"       example:"s3cr3tpassword"`
}

// LoginRequest is the request body for authentication.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"s3cr3tpassword"`
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
	if !bindJSON(c, &req) {
		return
	}
	user, token, err := h.uc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, TokenResponse{Token: token, UserID: user.ID})
}

// MeResponse is returned by the /me endpoint.
type MeResponse struct {
	ID    int64  `json:"id" example:"1"`
	Email string `json:"email" example:"user@example.com"`
}

// Me godoc
// @Summary      Get current user
// @Description  Returns the authenticated user's profile.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  MeResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.UserIDFrom(c)
	user, err := h.uc.Me(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, MeResponse{ID: user.ID, Email: user.Email})
}

// Login godoc
// @Summary      Log in
// @Description  Validates credentials and returns a signed JWT.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Credentials"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse  "invalid credentials"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	user, token, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			c.Header("WWW-Authenticate", `Bearer realm="linkr"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, TokenResponse{Token: token, UserID: user.ID})
}
