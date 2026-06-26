package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/usecase"
)

// TestPasswordValidation exercises the policy rules (min len, max len,
// uppercase required, digit required) through Register, which is the only
// public entry-point that calls validatePassword.
//
// Invalid cases return before bcrypt and are fast.
// The single valid case runs bcrypt once (~100 ms at DefaultCost).
func TestPasswordValidation(t *testing.T) {
	cases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"too short (<8)", "Pass1", true},
		{"too long (>72)", "Password1" + strings.Repeat("x", 64), true},
		{"no uppercase", "password1", true},
		{"no digit", "Password!", true},
		{"valid", "Password1", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := usecase.NewAuthUsecase(newFakeUserStore(), "secret")
			_, _, err := uc.Register(context.Background(), "u@example.com", tc.password)
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if err != nil && !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

// TestLogin_EnumerationPrevention verifies that a missing user and a wrong
// password both surface as ErrUnauthorized — callers cannot distinguish which
// half failed, preventing email enumeration.
func TestLogin_EnumerationPrevention(t *testing.T) {
	uc := usecase.NewAuthUsecase(newFakeUserStore(), "secret")
	ctx := context.Background()

	if _, _, err := uc.Register(ctx, "real@example.com", "Password1"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, _, err := uc.Login(ctx, "nobody@example.com", "Password1")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("unknown email: expected ErrUnauthorized, got %v", err)
	}

	_, _, err = uc.Login(ctx, "real@example.com", "WrongPass1")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("wrong password: expected ErrUnauthorized, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	uc := usecase.NewAuthUsecase(newFakeUserStore(), "secret")
	ctx := context.Background()

	if _, _, err := uc.Register(ctx, "user@example.com", "Password1"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	user, token, err := uc.Login(ctx, "user@example.com", "Password1")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("email: got %q", user.Email)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}
