package shortcode_test

import (
	"strings"
	"testing"

	"github.com/shehansuraweera/linkr/internal/shortcode"
)

func TestGenerate_Charset(t *testing.T) {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	set := make(map[rune]bool, len(allowed))
	for _, c := range allowed {
		set[c] = true
	}

	for i := 0; i < 1000; i++ {
		code, err := shortcode.Generate()
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if len(code) != 7 {
			t.Fatalf("expected length 7, got %d (%q)", len(code), code)
		}
		for _, ch := range code {
			if !set[ch] {
				t.Fatalf("unexpected character %q in code %q", ch, code)
			}
		}
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	seen := make(map[string]bool, 10_000)
	for i := 0; i < 10_000; i++ {
		code, err := shortcode.Generate()
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if seen[code] {
			t.Fatalf("duplicate code generated: %q", code)
		}
		seen[code] = true
	}
}

func TestValidateAlias(t *testing.T) {
	cases := []struct {
		alias   string
		wantErr bool
	}{
		{"my-link", false},
		{"My_Link_123", false},
		{"ab", true},                     // too short
		{strings.Repeat("a", 51), true},  // too long
		{"has space", true},
		{"has!char", true},
		{"api", true},                    // reserved
		{"healthz", true},                // reserved
		{"HEALTHZ", true},                // reserved (case-insensitive)
		{"valid-one", false},
	}

	for _, tc := range cases {
		err := shortcode.ValidateAlias(tc.alias)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateAlias(%q): wantErr=%v, got %v", tc.alias, tc.wantErr, err)
		}
	}
}
