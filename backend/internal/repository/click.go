package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/ua"
)

type ClickRepo struct {
	pool *pgxpool.Pool
}

func NewClickRepo(pool *pgxpool.Pool) *ClickRepo {
	return &ClickRepo{pool: pool}
}

// FlushBatch writes a batch of click events inside a single transaction:
//  1. Bulk-inserts raw events (link_id, clicked_at, ip_hash only — personal
//     signals user_agent and referer are never persisted).
//  2. Upserts pre-aggregated daily counts into click_daily.
//  3. Upserts device / browser / referrer breakdowns derived by parsing the
//     user_agent and referer strings in Go before they touch the database.
func (r *ClickRepo) FlushBatch(ctx context.Context, batch []domain.ClickEvent) error {
	if len(batch) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("click flush begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. Bulk-insert raw events — personal signals omitted intentionally.
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"clicks"},
		[]string{"link_id", "clicked_at", "ip_hash"},
		pgx.CopyFromSlice(len(batch), func(i int) ([]any, error) {
			return []any{
				batch[i].LinkID,
				batch[i].At,
				nilIfEmpty(batch[i].IPHash),
			}, nil
		}))
	if err != nil {
		return fmt.Errorf("click flush copy: %w", err)
	}

	// 2. Aggregate by (link_id, day) for daily rollup.
	type dayKey struct {
		linkID int64
		day    time.Time
	}
	type breakdownKey struct {
		linkID int64
		day    time.Time
		label  string
	}

	dailyCounts := make(map[dayKey]int64, len(batch))
	deviceCounts := make(map[breakdownKey]int64, len(batch))
	browserCounts := make(map[breakdownKey]int64, len(batch))
	refererCounts := make(map[breakdownKey]int64, len(batch))

	for _, e := range batch {
		day := e.At.UTC().Truncate(24 * time.Hour)
		dailyCounts[dayKey{e.LinkID, day}]++
		deviceCounts[breakdownKey{e.LinkID, day, ua.ParseDevice(e.UserAgent)}]++
		browserCounts[breakdownKey{e.LinkID, day, ua.ParseBrowser(e.UserAgent)}]++
		refererCounts[breakdownKey{e.LinkID, day, ua.ParseRefererDomain(e.Referer)}]++
	}

	for k, n := range dailyCounts {
		if _, err = tx.Exec(ctx,
			`INSERT INTO click_daily (link_id, day, count) VALUES ($1, $2, $3)
			 ON CONFLICT (link_id, day) DO UPDATE SET count = click_daily.count + EXCLUDED.count`,
			k.linkID, k.day, n); err != nil {
			return fmt.Errorf("click flush daily: %w", err)
		}
	}

	for k, n := range deviceCounts {
		if _, err = tx.Exec(ctx,
			`INSERT INTO clicks_by_device (link_id, day, device, count) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (link_id, day, device) DO UPDATE SET count = clicks_by_device.count + EXCLUDED.count`,
			k.linkID, k.day, k.label, n); err != nil {
			return fmt.Errorf("click flush device: %w", err)
		}
	}

	for k, n := range browserCounts {
		if _, err = tx.Exec(ctx,
			`INSERT INTO clicks_by_browser (link_id, day, browser, count) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (link_id, day, browser) DO UPDATE SET count = clicks_by_browser.count + EXCLUDED.count`,
			k.linkID, k.day, k.label, n); err != nil {
			return fmt.Errorf("click flush browser: %w", err)
		}
	}

	for k, n := range refererCounts {
		if _, err = tx.Exec(ctx,
			`INSERT INTO clicks_by_referer (link_id, day, domain, count) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (link_id, day, domain) DO UPDATE SET count = clicks_by_referer.count + EXCLUDED.count`,
			k.linkID, k.day, k.label, n); err != nil {
			return fmt.Errorf("click flush referer: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetStats returns total clicks, daily breakdown, and device/browser/referrer
// distributions for a link. All four queries are pipelined in a single
// network round-trip using pgx.Batch.
func (r *ClickRepo) GetStats(ctx context.Context, linkID int64) (domain.LinkStats, error) {
	b := &pgx.Batch{}
	b.Queue(
		`SELECT to_char(day, 'YYYY-MM-DD') AS day, count
		 FROM click_daily WHERE link_id = $1 ORDER BY day ASC`, linkID)
	b.Queue(
		`SELECT device, SUM(count) AS count
		 FROM clicks_by_device WHERE link_id = $1 GROUP BY device ORDER BY count DESC`, linkID)
	b.Queue(
		`SELECT browser, SUM(count) AS count
		 FROM clicks_by_browser WHERE link_id = $1 GROUP BY browser ORDER BY count DESC`, linkID)
	b.Queue(
		`SELECT domain, SUM(count) AS count
		 FROM clicks_by_referer WHERE link_id = $1 GROUP BY domain ORDER BY count DESC LIMIT 10`, linkID)

	br := r.pool.SendBatch(ctx, b)
	defer br.Close()

	dailyRows, err := br.Query()
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats daily query: %w", err)
	}
	daily, err := pgx.CollectRows(dailyRows, pgx.RowToStructByName[domain.DailyClickStat])
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats daily scan: %w", err)
	}

	deviceRows, err := br.Query()
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats device query: %w", err)
	}
	devices, err := pgx.CollectRows(deviceRows, pgx.RowToStructByName[domain.DeviceStat])
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats device scan: %w", err)
	}

	browserRows, err := br.Query()
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats browser query: %w", err)
	}
	browsers, err := pgx.CollectRows(browserRows, pgx.RowToStructByName[domain.BrowserStat])
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats browser scan: %w", err)
	}

	refererRows, err := br.Query()
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats referer query: %w", err)
	}
	referers, err := pgx.CollectRows(refererRows, pgx.RowToStructByName[domain.RefererStat])
	if err != nil {
		return domain.LinkStats{}, fmt.Errorf("stats referer scan: %w", err)
	}

	var total int64
	for _, d := range daily {
		total += d.Count
	}

	return domain.LinkStats{
		TotalClicks: total,
		Daily:       daily,
		Devices:     devices,
		Browsers:    browsers,
		Referers:    referers,
	}, nil
}

// GetOverview returns aggregate analytics across all links owned by userID.
// Six queries are pipelined in a single round-trip using pgx.Batch.
func (r *ClickRepo) GetOverview(ctx context.Context, userID int64) (domain.OverviewStats, error) {
	b := &pgx.Batch{}

	b.Queue(
		`SELECT COUNT(*)::bigint, COUNT(*) FILTER (WHERE is_active)::bigint
		 FROM links WHERE user_id = $1 AND deleted_at IS NULL`, userID)

	b.Queue(
		`SELECT to_char(cd.day, 'YYYY-MM-DD') AS day, SUM(cd.count)::bigint AS count
		 FROM click_daily cd
		 JOIN links l ON l.id = cd.link_id
		 WHERE l.user_id = $1 AND l.deleted_at IS NULL
		 GROUP BY cd.day ORDER BY cd.day ASC`, userID)

	b.Queue(
		`SELECT cbd.device, SUM(cbd.count)::bigint AS count
		 FROM clicks_by_device cbd
		 JOIN links l ON l.id = cbd.link_id
		 WHERE l.user_id = $1 AND l.deleted_at IS NULL
		 GROUP BY cbd.device ORDER BY count DESC`, userID)

	b.Queue(
		`SELECT cbb.browser, SUM(cbb.count)::bigint AS count
		 FROM clicks_by_browser cbb
		 JOIN links l ON l.id = cbb.link_id
		 WHERE l.user_id = $1 AND l.deleted_at IS NULL
		 GROUP BY cbb.browser ORDER BY count DESC`, userID)

	b.Queue(
		`SELECT cbr.domain, SUM(cbr.count)::bigint AS count
		 FROM clicks_by_referer cbr
		 JOIN links l ON l.id = cbr.link_id
		 WHERE l.user_id = $1 AND l.deleted_at IS NULL
		 GROUP BY cbr.domain ORDER BY count DESC LIMIT 10`, userID)

	b.Queue(
		`SELECT l.short_code, COALESCE(SUM(cd.count), 0)::bigint AS total_clicks
		 FROM links l
		 LEFT JOIN click_daily cd ON cd.link_id = l.id
		 WHERE l.user_id = $1 AND l.deleted_at IS NULL
		 GROUP BY l.short_code ORDER BY total_clicks DESC LIMIT 10`, userID)

	br := r.pool.SendBatch(ctx, b)
	defer br.Close()

	var totalLinks, activeLinks int64
	if err := br.QueryRow().Scan(&totalLinks, &activeLinks); err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview link counts: %w", err)
	}

	dailyRows, err := br.Query()
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview daily query: %w", err)
	}
	daily, err := pgx.CollectRows(dailyRows, pgx.RowToStructByName[domain.DailyClickStat])
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview daily scan: %w", err)
	}

	deviceRows, err := br.Query()
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview device query: %w", err)
	}
	devices, err := pgx.CollectRows(deviceRows, pgx.RowToStructByName[domain.DeviceStat])
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview device scan: %w", err)
	}

	browserRows, err := br.Query()
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview browser query: %w", err)
	}
	browsers, err := pgx.CollectRows(browserRows, pgx.RowToStructByName[domain.BrowserStat])
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview browser scan: %w", err)
	}

	refererRows, err := br.Query()
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview referer query: %w", err)
	}
	referers, err := pgx.CollectRows(refererRows, pgx.RowToStructByName[domain.RefererStat])
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview referer scan: %w", err)
	}

	topRows, err := br.Query()
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview top links query: %w", err)
	}
	topLinks, err := pgx.CollectRows(topRows, pgx.RowToStructByName[domain.LinkClickStat])
	if err != nil {
		return domain.OverviewStats{}, fmt.Errorf("overview top links scan: %w", err)
	}

	var total int64
	for _, d := range daily {
		total += d.Count
	}

	return domain.OverviewStats{
		TotalLinks:  totalLinks,
		ActiveLinks: activeLinks,
		TotalClicks: total,
		Daily:       daily,
		Devices:     devices,
		Browsers:    browsers,
		Referers:    referers,
		TopLinks:    topLinks,
	}, nil
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
