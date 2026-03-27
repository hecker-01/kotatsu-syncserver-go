package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters following OWASP recommendations.
const (
	argon2Memory      = 64 * 1024 // 64 MB
	argon2Iterations  = 3
	argon2Parallelism = 4
	argon2SaltLength  = 16
	argon2KeyLength   = 32
)

var (
	ErrInvalidHash         = errors.New("invalid argon2 hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// HashPassword hashes a password using Argon2id and returns a PHC-formatted string.
// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		argon2KeyLength,
	)

	// Encode to PHC string format
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Iterations,
		argon2Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// VerifyPassword verifies a password against an Argon2id hash in PHC format.
// Returns true if the password matches, false otherwise.
func VerifyPassword(password, encodedHash string) bool {
	// Parse the PHC-formatted hash string
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false
	}

	// Recompute the hash with the same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		uint32(len(hash)),
	)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, computedHash) == 1
}

// argon2Params holds the parameters extracted from a PHC hash string.
type argon2Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

// decodeHash parses a PHC-formatted Argon2id hash string.
func decodeHash(encodedHash string) (*argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	// parts[0] is empty (string starts with $)
	// parts[1] is "argon2id"
	// parts[2] is "v=19"
	// parts[3] is "m=65536,t=3,p=4"
	// parts[4] is base64 salt
	// parts[5] is base64 hash

	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	var params argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.iterations, &params.parallelism); err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	return &params, salt, hash, nil
}
