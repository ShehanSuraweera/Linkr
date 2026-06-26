package domain

import "time"

// Link mirrors the links table exactly — never carries computed fields.
type Link struct {
	ID          int64      `db:"id"`
	ShortCode   string     `db:"short_code"`
	OriginalURL string     `db:"original_url"`
	UserID      int64      `db:"user_id"`
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
	IsActive    bool       `db:"is_active"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// LinkSummary is a read-model for the list query: Link fields + click total
// aggregated via JOIN. pgx recurses into the embedded struct's db tags.
type LinkSummary struct {
	Link
	TotalClicks int64 `db:"total_clicks"`
}

func (l *Link) IsLive() bool {
	return l.DeletedAt == nil && l.IsActive &&
		(l.ExpiresAt == nil || l.ExpiresAt.After(time.Now()))
}
