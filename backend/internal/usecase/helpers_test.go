package usecase_test

import (
	"context"
	"time"

	"github.com/shehansuraweera/linkr/internal/domain"
)

// ── fakeLinkStore ─────────────────────────────────────────────────────────────

type fakeLinkStore struct {
	links     map[string]domain.Link
	nextID    int64
	createErr error // when set, every Create call returns this
}

func newFakeLinkStore() *fakeLinkStore {
	return &fakeLinkStore{links: make(map[string]domain.Link), nextID: 1}
}

func (f *fakeLinkStore) Create(_ context.Context, shortCode, originalURL string, userID int64, expiresAt *time.Time) (domain.Link, error) {
	if f.createErr != nil {
		return domain.Link{}, f.createErr
	}
	if _, exists := f.links[shortCode]; exists {
		return domain.Link{}, domain.ErrConflict
	}
	l := domain.Link{
		ID:          f.nextID,
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	f.nextID++
	f.links[shortCode] = l
	return l, nil
}

func (f *fakeLinkStore) GetByCode(_ context.Context, code string) (domain.Link, error) {
	l, ok := f.links[code]
	if !ok {
		return domain.Link{}, domain.ErrNotFound
	}
	return l, nil
}

func (f *fakeLinkStore) GetByIDAndUser(_ context.Context, id, _ int64) (domain.Link, error) {
	for _, l := range f.links {
		if l.ID == id {
			return l, nil
		}
	}
	return domain.Link{}, domain.ErrNotFound
}

func (f *fakeLinkStore) List(_ context.Context, _ int64, _ time.Time, _ int64, _ int32, _ string) ([]domain.LinkSummary, bool, error) {
	return nil, false, nil
}

func (f *fakeLinkStore) Update(_ context.Context, id, _ int64, isActive *bool, setExpiresAt bool, expiresAt *time.Time) (domain.Link, error) {
	for code, l := range f.links {
		if l.ID == id {
			if isActive != nil {
				l.IsActive = *isActive
			}
			if setExpiresAt {
				l.ExpiresAt = expiresAt
			}
			f.links[code] = l
			return l, nil
		}
	}
	return domain.Link{}, domain.ErrNotFound
}

func (f *fakeLinkStore) SoftDelete(_ context.Context, id, _ int64) error {
	for code, l := range f.links {
		if l.ID == id {
			delete(f.links, code)
			return nil
		}
	}
	return domain.ErrNotFound
}

// ── fakeClickStore ────────────────────────────────────────────────────────────

type fakeClickStore struct{ stats domain.LinkStats }

func (f *fakeClickStore) GetStats(_ context.Context, _ int64) (domain.LinkStats, error) {
	return f.stats, nil
}

func (f *fakeClickStore) GetOverview(_ context.Context, _ int64) (domain.OverviewStats, error) {
	return domain.OverviewStats{}, nil
}

// ── fakeUserStore ─────────────────────────────────────────────────────────────

type fakeUserStore struct {
	byEmail map[string]domain.User
	byID    map[int64]domain.User
	nextID  int64
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		byEmail: make(map[string]domain.User),
		byID:    make(map[int64]domain.User),
		nextID:  1,
	}
}

func (f *fakeUserStore) Create(_ context.Context, email, hash string) (domain.User, error) {
	if _, exists := f.byEmail[email]; exists {
		return domain.User{}, domain.ErrConflict
	}
	u := domain.User{ID: f.nextID, Email: email, PasswordHash: hash}
	f.nextID++
	f.byEmail[email] = u
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeUserStore) GetByEmail(_ context.Context, email string) (domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserStore) GetByID(_ context.Context, id int64) (domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}
