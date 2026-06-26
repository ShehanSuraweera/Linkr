package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/usecase"
)

func TestCreate_AutoCode(t *testing.T) {
	uc := usecase.NewLinkUsecase(newFakeLinkStore(), &fakeClickStore{})
	link, err := uc.Create(context.Background(), usecase.CreateLinkInput{
		URL: "https://example.com", UserID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.ShortCode == "" {
		t.Error("expected non-empty short code")
	}
	if link.OriginalURL != "https://example.com" {
		t.Errorf("URL: got %q", link.OriginalURL)
	}
}

func TestCreate_Alias(t *testing.T) {
	uc := usecase.NewLinkUsecase(newFakeLinkStore(), &fakeClickStore{})
	link, err := uc.Create(context.Background(), usecase.CreateLinkInput{
		URL: "https://example.com", Alias: "my-link", UserID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.ShortCode != "my-link" {
		t.Errorf("short code: got %q, want %q", link.ShortCode, "my-link")
	}
}

func TestCreate_ValidationErrors(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	cases := []struct {
		name  string
		input usecase.CreateLinkInput
	}{
		{"empty url", usecase.CreateLinkInput{URL: "", UserID: 1}},
		{"ftp scheme", usecase.CreateLinkInput{URL: "ftp://example.com", UserID: 1}},
		{"localhost", usecase.CreateLinkInput{URL: "http://localhost/admin", UserID: 1}},
		{"private ip", usecase.CreateLinkInput{URL: "http://192.168.1.1", UserID: 1}},
		{"alias too short", usecase.CreateLinkInput{URL: "https://example.com", Alias: "ab", UserID: 1}},
		{"reserved alias", usecase.CreateLinkInput{URL: "https://example.com", Alias: "api", UserID: 1}},
		{"past expiry", usecase.CreateLinkInput{URL: "https://example.com", ExpiresAt: &past, UserID: 1}},
	}

	uc := usecase.NewLinkUsecase(newFakeLinkStore(), &fakeClickStore{})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Create(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestCreate_AliasConflict(t *testing.T) {
	uc := usecase.NewLinkUsecase(newFakeLinkStore(), &fakeClickStore{})
	ctx := context.Background()
	in := usecase.CreateLinkInput{URL: "https://example.com", Alias: "taken", UserID: 1}

	if _, err := uc.Create(ctx, in); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := uc.Create(ctx, in)
	if !errors.Is(err, domain.ErrAliasTaken) {
		t.Errorf("expected ErrAliasTaken, got %v", err)
	}
}

func TestCreate_AutoCodeExhaustRetries(t *testing.T) {
	// Store that always conflicts — simulates the (vanishingly unlikely) case
	// where all three auto-generated codes are already in use.
	store := &fakeLinkStore{
		links:     make(map[string]domain.Link),
		createErr: domain.ErrConflict,
	}
	uc := usecase.NewLinkUsecase(store, &fakeClickStore{})

	_, err := uc.Create(context.Background(), usecase.CreateLinkInput{
		URL: "https://example.com", UserID: 1,
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// Must not be surfaced as ErrAliasTaken — that sentinel is only for explicit aliases.
	if errors.Is(err, domain.ErrAliasTaken) {
		t.Error("auto-gen exhaustion must not surface as ErrAliasTaken")
	}
}

func TestGetStats_OwnershipEnforced(t *testing.T) {
	store := newFakeLinkStore()
	cs := &fakeClickStore{stats: domain.LinkStats{TotalClicks: 7}}
	uc := usecase.NewLinkUsecase(store, cs)
	ctx := context.Background()

	if _, err := store.Create(ctx, "priv", "https://example.com", 1, nil); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Owner can read stats.
	got, err := uc.GetStats(ctx, "priv", 1)
	if err != nil {
		t.Fatalf("owner access: %v", err)
	}
	if got.TotalClicks != 7 {
		t.Errorf("TotalClicks: got %d, want 7", got.TotalClicks)
	}

	// A different user is forbidden.
	_, err = uc.GetStats(ctx, "priv", 99)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("wrong user: expected ErrForbidden, got %v", err)
	}
}
