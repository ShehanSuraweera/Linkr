package domain

import "time"

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

func (l *Link) IsLive() bool {
	return l.DeletedAt == nil && l.IsActive &&
		(l.ExpiresAt == nil || l.ExpiresAt.After(time.Now()))
}
