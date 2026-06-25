package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shehansuraweera/linkr/internal/domain"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash string) (domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`INSERT INTO users (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, email, password_hash, created_at`,
		email, passwordHash)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrConflict
		}
		return domain.User{}, fmt.Errorf("create user scan: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user scan: %w", err)
	}
	return user, nil
}
