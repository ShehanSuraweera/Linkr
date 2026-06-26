package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shehansuraweera/linkr/internal/domain"
	"github.com/shehansuraweera/linkr/internal/service"
	"github.com/shehansuraweera/linkr/internal/shortcode"
)

// LinkStore is the persistence interface required by LinkUsecase.
// Defined here (at the consumer) so the usecase layer owns the contract
// and the repository is a plug-in detail.
type LinkStore interface {
	Create(ctx context.Context, shortCode, originalURL string, userID int64, expiresAt *time.Time) (domain.Link, error)
	GetByCode(ctx context.Context, code string) (domain.Link, error)
	List(ctx context.Context, userID int64, cursorCreatedAt time.Time, cursorID int64, limit int32, search string) ([]domain.LinkSummary, bool, error)
}

// ClickStore is the read-only analytics interface required by LinkUsecase.
type ClickStore interface {
	GetStats(ctx context.Context, linkID int64) (domain.LinkStats, error)
}

// CreateLinkInput carries the inputs for link creation.
// ExpiresAt is already parsed from RFC3339 by the HTTP layer — only the
// "must be in the future" rule belongs to the usecase.
type CreateLinkInput struct {
	URL       string
	Alias     string     // empty → auto-generate
	ExpiresAt *time.Time // nil → no expiry
	UserID    int64
}

type LinkUsecase struct {
	links  LinkStore
	clicks ClickStore
}

func NewLinkUsecase(links LinkStore, clicks ClickStore) *LinkUsecase {
	return &LinkUsecase{links: links, clicks: clicks}
}

// Create validates inputs, generates a short code when no alias is supplied,
// and persists the link.
func (uc *LinkUsecase) Create(ctx context.Context, in CreateLinkInput) (domain.Link, error) {
	in.URL = strings.TrimSpace(in.URL)
	if in.URL == "" {
		return domain.Link{}, newInputErr("url is required")
	}

	code := strings.TrimSpace(in.Alias)
	if code != "" {
		if err := shortcode.ValidateAlias(code); err != nil {
			return domain.Link{}, newInputErr(err.Error())
		}
	}

	if err := service.ValidateURL(in.URL); err != nil {
		return domain.Link{}, newInputErr(err.Error())
	}

	if in.ExpiresAt != nil && in.ExpiresAt.Before(time.Now()) {
		return domain.Link{}, newInputErr("expires_at must be in the future")
	}

	if code != "" {
		link, err := uc.links.Create(ctx, code, in.URL, in.UserID, in.ExpiresAt)
		if err != nil {
			if errors.Is(err, domain.ErrConflict) {
				return domain.Link{}, domain.ErrAliasTaken
			}
			return domain.Link{}, err
		}
		return link, nil
	}

	// Auto-generate with collision retry.
	for i := 0; i < 3; i++ {
		var err error
		code, err = shortcode.Generate()
		if err != nil {
			return domain.Link{}, err
		}
		link, createErr := uc.links.Create(ctx, code, in.URL, in.UserID, in.ExpiresAt)
		if createErr == nil {
			return link, nil
		}
		if !errors.Is(createErr, domain.ErrConflict) {
			return domain.Link{}, createErr
		}
	}
	return domain.Link{}, fmt.Errorf("could not generate unique code")
}

// List delegates pagination and search to the repository.
func (uc *LinkUsecase) List(ctx context.Context, userID int64, cursorCreatedAt time.Time, cursorID int64, limit int32, search string) ([]domain.LinkSummary, bool, error) {
	return uc.links.List(ctx, userID, cursorCreatedAt, cursorID, limit, search)
}

// GetStats returns click analytics for a link, enforcing that the caller owns it.
func (uc *LinkUsecase) GetStats(ctx context.Context, code string, userID int64) (domain.LinkStats, error) {
	link, err := uc.links.GetByCode(ctx, code)
	if err != nil {
		return domain.LinkStats{}, err
	}
	if link.UserID != userID {
		return domain.LinkStats{}, domain.ErrForbidden
	}
	return uc.clicks.GetStats(ctx, link.ID)
}

// inputErr is a validation error whose message goes directly to the client.
// errors.Is(e, domain.ErrInvalidInput) == true via the Unwrap chain,
// so respondError maps it to 400 and uses err.Error() as the body.
type inputErr string

func newInputErr(msg string) inputErr  { return inputErr(msg) }
func (e inputErr) Error() string       { return string(e) }
func (e inputErr) Unwrap() error       { return domain.ErrInvalidInput }
