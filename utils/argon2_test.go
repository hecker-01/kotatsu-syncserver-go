package utils

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify PHC format
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Errorf("Hash format incorrect, got: %s", hash)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("Expected 6 parts in hash, got %d", len(parts))
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Correct password should verify
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword returned false for correct password")
	}

	// Wrong password should not verify
	if VerifyPassword("wrongpassword", hash) {
		t.Error("VerifyPassword returned true for incorrect password")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	// Invalid formats should return false, not panic
	invalidHashes := []string{
		"",
		"invalid",
		"$argon2id$v=19$m=65536,t=3,p=4$",
		"$argon2i$v=19$m=65536,t=3,p=4$salt$hash", // wrong variant
	}

	for _, hash := range invalidHashes {
		if VerifyPassword("password", hash) {
			t.Errorf("VerifyPassword should return false for invalid hash: %s", hash)
		}
	}
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	password := "testpassword123"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to random salt")
	}
}
