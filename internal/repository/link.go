package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shehansuraweera/linkr/internal/domain"
)

type LinkRepo struct {
	pool *pgxpool.Pool
}

func NewLinkRepo(pool *pgxpool.Pool) *LinkRepo {
	return &LinkRepo{pool: pool}
}

func (r *LinkRepo) Create(ctx context.Context, shortCode, originalURL string, userID int64, expiresAt *time.Time) (domain.Link, error) {
	rows, err := r.pool.Query(ctx,
		`INSERT INTO links (short_code, original_url, user_id, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, short_code, original_url, user_id, created_at, expires_at, is_active, deleted_at`,
		shortCode, originalURL, userID, expiresAt)
	if err != nil {
		return domain.Link{}, fmt.Errorf("create link: %w", err)
	}
	link, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Link])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Link{}, domain.ErrConflict
		}
		return domain.Link{}, fmt.Errorf("create link scan: %w", err)
	}
	return link, nil
}

func (r *LinkRepo) GetByCode(ctx context.Context, code string) (domain.Link, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, short_code, original_url, user_id, created_at, expires_at, is_active, deleted_at
		 FROM links WHERE short_code = $1 AND deleted_at IS NULL`, code)
	if err != nil {
		return domain.Link{}, fmt.Errorf("get link: %w", err)
	}
	link, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Link])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Link{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Link{}, fmt.Errorf("get link scan: %w", err)
	}
	return link, nil
}

// List returns links for userID using keyset pagination over (created_at DESC, id DESC).
// Pass zero values for cursorCreatedAt/cursorID to get the first page.
func (r *LinkRepo) List(ctx context.Context, userID int64, cursorCreatedAt time.Time, cursorID int64, limit int32) ([]domain.Link, bool, error) {
	if cursorCreatedAt.IsZero() {
		cursorCreatedAt = time.Now().Add(time.Second) // slightly future so first row is included
		cursorID = 1<<62 - 1                          // max int64-ish
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, short_code, original_url, user_id, created_at, expires_at, is_active, deleted_at
		 FROM links
		 WHERE user_id = $1
		   AND deleted_at IS NULL
		   AND (created_at, id) < ($2, $3)
		 ORDER BY created_at DESC, id DESC
		 LIMIT $4`,
		userID, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("list links: %w", err)
	}
	links, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Link])
	if err != nil {
		return nil, false, fmt.Errorf("list links scan: %w", err)
	}

	hasMore := len(links) > int(limit)
	if hasMore {
		links = links[:limit]
	}
	return links, hasMore, nil
}

func (r *LinkRepo) SoftDelete(ctx context.Context, id, userID int64) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE links SET deleted_at = now() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LinkRepo) GetByIDAndUser(ctx context.Context, id, userID int64) (domain.Link, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, short_code, original_url, user_id, created_at, expires_at, is_active, deleted_at
		 FROM links WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return domain.Link{}, fmt.Errorf("get link by id: %w", err)
	}
	link, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Link])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Link{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Link{}, fmt.Errorf("get link by id scan: %w", err)
	}
	return link, nil
}
