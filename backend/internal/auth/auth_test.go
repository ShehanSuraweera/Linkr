package auth_test

import (
	"testing"

	"github.com/shehansuraweera/linkr/internal/auth"
)

const testSecret = "a-32-character-test-secret-key!!"

func TestIssueAndParseToken_RoundTrip(t *testing.T) {
	tok, err := auth.IssueToken(42, testSecret)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if tok == "" {
		t.Fatal("expected non-empty token string")
	}

	claims, err := auth.ParseToken(tok, testSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID: got %d, want 42", claims.UserID)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	tok, _ := auth.IssueToken(1, testSecret)
	_, err := auth.ParseToken(tok, "completely-different-secret-key!!")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestParseToken_Tampered(t *testing.T) {
	tok, _ := auth.IssueToken(1, testSecret)
	tampered := tok[:len(tok)-4] + "xxxx"
	_, err := auth.ParseToken(tampered, testSecret)
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}

// TestHashAndCheckPassword verifies the bcrypt round-trip and that wrong
// passwords are rejected. Runs ~100 ms at DefaultCost — expected.
func TestHashAndCheckPassword(t *testing.T) {
	const plain = "Password1"

	hash, err := auth.HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == plain {
		t.Error("hash must not equal plain text")
	}

	if err := auth.CheckPassword(hash, plain); err != nil {
		t.Errorf("CheckPassword(correct): %v", err)
	}
	if err := auth.CheckPassword(hash, "WrongPass1"); err == nil {
		t.Error("CheckPassword(wrong): expected error, got nil")
	}
}
