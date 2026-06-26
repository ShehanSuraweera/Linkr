package ua_test

import (
	"testing"

	"github.com/shehansuraweera/linkr/internal/ua"
)

func TestParseDevice(t *testing.T) {
	cases := []struct {
		name string
		ua   string
		want string
	}{
		// Desktop
		{"windows chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120", "desktop"},
		{"mac safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605 Safari/605", "desktop"},
		{"empty", "", "desktop"},
		// Tablet — tablet check must come before mobile so these are not misclassified
		{"ipad", "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605 Safari/604", "tablet"},
		{"android tablet (no mobile token)", "Mozilla/5.0 (Linux; Android 12; SM-T870) AppleWebKit/537 Safari/537", "tablet"},
		// Mobile
		{"iphone", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605 Mobile/15E148", "mobile"},
		{"android phone (mobi)", "Mozilla/5.0 (Linux; Android 12; Pixel 6) Mobile Safari/537", "mobile"},
		{"ipod", "Mozilla/5.0 (iPod touch; CPU iOS 15_0 like Mac OS X) AppleWebKit/605", "mobile"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ua.ParseDevice(tc.ua); got != tc.want {
				t.Errorf("ParseDevice(%q) = %q, want %q", tc.ua, got, tc.want)
			}
		})
	}
}

func TestParseBrowser(t *testing.T) {
	// Order matters in the implementation: Edge and Opera carry "Chrome" and
	// "Safari" tokens in their UA strings, so they must be matched first.
	cases := []struct {
		name string
		ua   string
		want string
	}{
		{"edge (contains Chrome token too)", "Mozilla/5.0 Chrome/120 Safari/537 Edg/120.0.0", "edge"},
		{"opera (contains Chrome token too)", "Mozilla/5.0 Chrome/119 Safari/537 OPR/105.0", "opera"},
		{"firefox", "Mozilla/5.0 (Windows NT 10.0; rv:121.0) Gecko/20100101 Firefox/121.0", "firefox"},
		{"chrome", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537 Chrome/120 Safari/537", "chrome"},
		{"safari (no Chrome token)", "Mozilla/5.0 (Macintosh) AppleWebKit/605 Version/17.0 Safari/605", "safari"},
		{"curl / other", "curl/7.88.1", "other"},
		{"empty", "", "other"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ua.ParseBrowser(tc.ua); got != tc.want {
				t.Errorf("ParseBrowser(%q) = %q, want %q", tc.ua, got, tc.want)
			}
		})
	}
}

func TestParseRefererDomain(t *testing.T) {
	cases := []struct {
		name    string
		referer string
		want    string
	}{
		{"empty → direct", "", "direct"},
		{"google with www and path", "https://www.google.com/search?q=linkr", "google.com"},
		{"github", "https://github.com/shehansuraweera", "github.com"},
		{"port stripped", "http://example.com:8080/path?q=1#anchor", "example.com"},
		{"subdomain preserved", "https://sub.domain.io/page", "sub.domain.io"},
		{"no scheme", "not-a-url", "not-a-url"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ua.ParseRefererDomain(tc.referer); got != tc.want {
				t.Errorf("ParseRefererDomain(%q) = %q, want %q", tc.referer, got, tc.want)
			}
		})
	}
}
