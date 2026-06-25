package shortcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

const (
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength = 7
	maxRetries = 3
)

// reserved codes that would shadow API routes or well-known paths
var reserved = map[string]bool{
	"api": true, "health": true, "healthz": true, "readyz": true,
	"metrics": true, "favicon.ico": true, "robots.txt": true,
}

var aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)

// Generate returns a cryptographically random base62 code of length codeLength.
func Generate() (string, error) {
	b := make([]byte, codeLength)
	alphabetLen := big.NewInt(int64(len(alphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", fmt.Errorf("shortcode: random read: %w", err)
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}

// ValidateAlias checks a user-supplied custom alias.
func ValidateAlias(alias string) error {
	alias = strings.TrimSpace(alias)
	if !aliasPattern.MatchString(alias) {
		return fmt.Errorf("alias must be 3–50 chars, letters/digits/underscore/hyphen only")
	}
	if reserved[strings.ToLower(alias)] {
		return fmt.Errorf("alias %q is reserved", alias)
	}
	return nil
}
