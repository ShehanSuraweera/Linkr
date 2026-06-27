package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shehansuraweera/linkr/internal/config"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/repository"
)

// These tests require a real Postgres instance.
// Set TEST_DATABASE_URL to run them, e.g.:
//   TEST_DATABASE_URL=postgres://linkr:linkr@localhost:5432/linkr?sslmode=disable go test ./internal/repository/...

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration tests")
	}
	cfg := &config.Config{
		DatabaseURL: dsn,
		DBMaxConns:  5,
		DBMinConns:  1,
	}
	pool, err := repository.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestUserRepo_CreateAndGet(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepo(pool)
	ctx := context.Background()

	email := "test+" + time.Now().Format("20060102150405.000") + "@example.com"
	user, err := repo.Create(ctx, email, "hashed")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("got ID %d, want %d", got.ID, user.ID)
	}

	// Duplicate email returns ErrConflict.
	_, err = repo.Create(ctx, email, "hashed")
	if err != domain.ErrConflict {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestLinkRepo_CreateListDelete(t *testing.T) {
	pool := testPool(t)
	userRepo := repository.NewUserRepo(pool)
	linkRepo := repository.NewLinkRepo(pool)
	ctx := context.Background()

	email := "linktest+" + time.Now().Format("20060102150405.000") + "@example.com"
	user, err := userRepo.Create(ctx, email, "hashed")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	code := "tst" + time.Now().Format("0405")
	link, err := linkRepo.Create(ctx, code, "https://example.com", user.ID, nil)
	if err != nil {
		t.Fatalf("Create link: %v", err)
	}
	if link.ShortCode != code {
		t.Errorf("got code %q, want %q", link.ShortCode, code)
	}

	got, err := linkRepo.GetByCode(ctx, code)
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if got.ID != link.ID {
		t.Errorf("got ID %d, want %d", got.ID, link.ID)
	}

	links, hasMore, err := linkRepo.List(ctx, user.ID, time.Time{}, 0, 10, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(links) == 0 {
		t.Fatal("expected at least one link")
	}
	if hasMore {
		t.Error("expected hasMore=false for single link")
	}

	if err := linkRepo.SoftDelete(ctx, code, user.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	// Soft-deleted link should not be found.
	_, err = linkRepo.GetByCode(ctx, code)
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestClickRepo_FlushAndStats(t *testing.T) {
	pool := testPool(t)
	userRepo := repository.NewUserRepo(pool)
	linkRepo := repository.NewLinkRepo(pool)
	clickRepo := repository.NewClickRepo(pool)
	ctx := context.Background()

	email := "clicktest+" + time.Now().Format("20060102150405.000") + "@example.com"
	user, _ := userRepo.Create(ctx, email, "hashed")
	code := "clk" + time.Now().Format("0405")
	link, _ := linkRepo.Create(ctx, code, "https://example.com", user.ID, nil)

	now := time.Now().UTC()
	batch := []domain.ClickEvent{
		{LinkID: link.ID, At: now},
		{LinkID: link.ID, At: now},
		{LinkID: link.ID, At: now.Add(-25 * time.Hour)}, // previous day
	}
	if err := clickRepo.FlushBatch(ctx, batch); err != nil {
		t.Fatalf("FlushBatch: %v", err)
	}

	stats, err := clickRepo.GetStats(ctx, link.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalClicks != 3 {
		t.Errorf("TotalClicks: got %d, want 3", stats.TotalClicks)
	}
	if len(stats.Daily) != 2 {
		t.Errorf("Daily buckets: got %d, want 2", len(stats.Daily))
	}
}
