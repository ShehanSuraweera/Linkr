package domain_test

import (
	"testing"
	"time"

	"github.com/shehansuraweera/linkr/internal/domain"
)

func TestIsLive(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	deletedAt := now

	cases := []struct {
		name string
		link domain.Link
		want bool
	}{
		{"active, no expiry", domain.Link{IsActive: true}, true},
		{"inactive", domain.Link{IsActive: false}, false},
		{"soft-deleted", domain.Link{IsActive: true, DeletedAt: &deletedAt}, false},
		{"not yet expired", domain.Link{IsActive: true, ExpiresAt: &future}, true},
		{"expired", domain.Link{IsActive: true, ExpiresAt: &past}, false},
		{"inactive and deleted", domain.Link{IsActive: false, DeletedAt: &deletedAt}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.link.IsLive(); got != tc.want {
				t.Errorf("IsLive() = %v, want %v", got, tc.want)
			}
		})
	}
}
