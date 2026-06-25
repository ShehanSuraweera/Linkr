package service_test

import (
	"testing"

	"github.com/shehansuraweera/linkr/internal/service"
)

func TestValidateURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://example.com/path?q=1", false},
		{"valid http", "http://example.com", false},
		{"empty", "", true},
		{"ftp scheme", "ftp://example.com", true},
		{"no scheme", "example.com", true},
		{"javascript", "javascript:alert(1)", true},
		{"localhost", "http://localhost/admin", true},
		{"127.0.0.1", "http://127.0.0.1:8080", true},
		{"private 10.x", "http://10.0.0.1/secret", true},
		{"private 192.168.x", "http://192.168.1.1", true},
		{"link-local 169.254", "http://169.254.169.254/latest/meta-data/", true},
		{"dot local", "http://myhost.local/path", true},
		{"no host", "https:///path", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidateURL(tc.url)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateURL(%q) wantErr=%v, got err=%v", tc.url, tc.wantErr, err)
			}
		})
	}
}
