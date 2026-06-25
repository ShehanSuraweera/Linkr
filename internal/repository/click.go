package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shehansuraweera/linkr/internal/domain"
)

type ClickRepo struct {
	pool *pgxpool.Pool
}

func NewClickRepo(pool *pgxpool.Pool) *ClickRepo {
	return &ClickRepo{pool: pool}
}

// FlushBatch inserts raw click events and upserts the daily rollup in one transaction.
// Uses pgx CopyFrom for the bulk insert — significantly faster than row-by-row inserts.
func (r *ClickRepo) FlushBatch(ctx context.Context, batch []domain.ClickEvent) error {
	if len(batch) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("click flush begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Bulk-insert raw events via COPY protocol.
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"clicks"},
		[]string{"link_id", "clicked_at", "ip_hash", "user_agent", "referer"},
		pgx.CopyFromSlice(len(batch), func(i int) ([]any, error) {
			return []any{
				batch[i].LinkID,
				batch[i].At,
				nilIfEmpty(batch[i].IPHash),
				nilIfEmpty(batch[i].UserAgent),
				nilIfEmpty(batch[i].Referer),
			}, nil
		}))
	if err != nil {
		return fmt.Errorf("click flush copy: %w", err)
	}

	// Aggregate by (link_id, day) in Go first — one upsert per (link, day) instead of per click.
	type key struct {
		linkID int64
		day    time.Time
	}
	counts := make(map[key]int64, len(batch))
	for _, e := range batch {
		counts[key{e.LinkID, e.At.UTC().Truncate(24 * time.Hour)}]++
	}

	for k, n := range counts {
		_, err = tx.Exec(ctx,
			`INSERT INTO click_daily (link_id, day, count)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (link_id, day)
			 DO UPDATE SET count = click_daily.count + EXCLUDED.count`,
			k.linkID, k.day, n)
		if err != nil {
			return fmt.Errorf("click flush upsert daily: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetStats returns total clicks and per-day breakdown for a link.
func (r *ClickRepo) GetStats(ctx context.Context, linkID int64) (domain.LinkStats, error) {
	// Total from rollup — O(days), not O(clicks).
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(count), 0) FROM click_daily WHERE link_id = $1`, linkID).
		Scan(&total)
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats total: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT to_char(day, 'YYYY-MM-DD') AS day, count
		 FROM click_daily WHERE link_id = $1
		 ORDER BY day ASC`, linkID)
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats daily: %w", err)
	}
	daily, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.DailyClickStat])
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats daily scan: %w", err)
	}

	return domain.LinkStats{TotalClicks: total, Daily: daily}, nil
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
