package usecase

import (
	"context"
	"errors"

	"github.com/shehansuraweera/linkr/internal/auth"
	"github.com/shehansuraweera/linkr/internal/domain"
)

// UserStore is the persistence interface required by AuthUsecase.
type UserStore interface {
	Create(ctx context.Context, email, passwordHash string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

type AuthUsecase struct {
	users     UserStore
	jwtSecret string
}

func NewAuthUsecase(users UserStore, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{users: users, jwtSecret: jwtSecret}
}

// validatePassword enforces the password policy:
// at least 8 characters, one uppercase letter, one digit.
const maxPasswordLen = 72 // bcrypt silently truncates beyond this

func validatePassword(password string) error {
	if len(password) < 8 {
		return newInputErr("password must be at least 8 characters")
	}
	if len(password) > maxPasswordLen {
		return newInputErr("password must be at most 72 characters")
	}
	hasUpper, hasDigit := false, false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasUpper {
		return newInputErr("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return newInputErr("password must contain at least one number")
	}
	return nil
}

// Register validates the password policy, hashes the password, persists the
// user, and issues a signed JWT.
func (uc *AuthUsecase) Register(ctx context.Context, email, password string) (domain.User, string, error) {
	if err := validatePassword(password); err != nil {
		return domain.User{}, "", err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return domain.User{}, "", err
	}
	user, err := uc.users.Create(ctx, email, hash)
	if err != nil {
		return domain.User{}, "", err
	}
	token, err := auth.IssueToken(user.ID, uc.jwtSecret)
	if err != nil {
		return domain.User{}, "", err
	}
	return user, token, nil
}

// Login verifies credentials and issues a signed JWT.
// Both a missing user and a wrong password return ErrUnauthorized so callers
// cannot distinguish which half failed (prevents email enumeration).
func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	user, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, "", domain.ErrUnauthorized
		}
		return domain.User{}, "", err
	}
	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return domain.User{}, "", domain.ErrUnauthorized
	}
	token, err := auth.IssueToken(user.ID, uc.jwtSecret)
	if err != nil {
		return domain.User{}, "", err
	}
	return user, token, nil
}
