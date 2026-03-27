package testutil

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Default test JWT configuration values.
const (
	TestJWTSecret   = "test-jwt-secret-for-testing-only"
	TestJWTIssuer   = "http://0.0.0.0:9292/"
	TestJWTAudience = "http://0.0.0.0:9292/resource"
)

// SetupTestJWTEnv configures environment variables for JWT testing.
// Call this in TestMain or at the start of tests that need JWT functionality.
// Returns a cleanup function to restore original values.
func SetupTestJWTEnv(t *testing.T) func() {
	t.Helper()

	// Save original values
	origSecret := os.Getenv("JWT_SECRET")
	origIssuer := os.Getenv("JWT_ISSUER")
	origAudience := os.Getenv("JWT_AUDIENCE")

	// Set test values
	os.Setenv("JWT_SECRET", TestJWTSecret)
	os.Setenv("JWT_ISSUER", TestJWTIssuer)
	os.Setenv("JWT_AUDIENCE", TestJWTAudience)

	// Return cleanup function
	return func() {
		if origSecret != "" {
			os.Setenv("JWT_SECRET", origSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
		if origIssuer != "" {
			os.Setenv("JWT_ISSUER", origIssuer)
		} else {
			os.Unsetenv("JWT_ISSUER")
		}
		if origAudience != "" {
			os.Setenv("JWT_AUDIENCE", origAudience)
		} else {
			os.Unsetenv("JWT_AUDIENCE")
		}
	}
}

// GenerateTestToken creates a valid JWT for testing with the specified user ID.
// The token uses the test JWT secret and has a 30-day expiration.
func GenerateTestToken(t *testing.T, userID int64) string {
	t.Helper()

	// Ensure test JWT environment is set
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", TestJWTSecret)
	}

	secret := os.Getenv("JWT_SECRET")
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = TestJWTIssuer
	}
	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = TestJWTAudience
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(30 * 24 * time.Hour).Unix(),
		"iss":     issuer,
		"aud":     audience,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	return tokenString
}

// GenerateTestTokenWithExpiry creates a JWT with a custom expiration time.
func GenerateTestTokenWithExpiry(t *testing.T, userID int64, expiry time.Time) string {
	t.Helper()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = TestJWTSecret
	}
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = TestJWTIssuer
	}
	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = TestJWTAudience
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiry.Unix(),
		"iss":     issuer,
		"aud":     audience,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	return tokenString
}

// InvalidToken returns a malformed JWT string for testing authentication failures.
// This token cannot be parsed or validated.
func InvalidToken() string {
	return "invalid.token.string"
}

// ExpiredToken returns an expired JWT for testing token expiration handling.
// The token was valid but expired 1 hour ago.
func ExpiredToken(t *testing.T, userID int64) string {
	t.Helper()

	expiredTime := time.Now().Add(-1 * time.Hour)
	return GenerateTestTokenWithExpiry(t, userID, expiredTime)
}

// WrongSignatureToken returns a JWT signed with a different secret.
// This tests that the server correctly rejects tokens with invalid signatures.
func WrongSignatureToken(t *testing.T, userID int64) string {
	t.Helper()

	wrongSecret := "wrong-secret-key"
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = TestJWTIssuer
	}
	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = TestJWTAudience
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(30 * 24 * time.Hour).Unix(),
		"iss":     issuer,
		"aud":     audience,
	})

	tokenString, err := token.SignedString([]byte(wrongSecret))
	if err != nil {
		t.Fatalf("failed to generate wrong signature token: %v", err)
	}

	return tokenString
}

// MissingClaimToken returns a JWT without the user_id claim.
func MissingClaimToken(t *testing.T) string {
	t.Helper()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = TestJWTSecret
	}
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = TestJWTIssuer
	}
	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = TestJWTAudience
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// No user_id claim
		"exp": now.Add(30 * 24 * time.Hour).Unix(),
		"iss": issuer,
		"aud": audience,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to generate missing claim token: %v", err)
	}

	return tokenString
}

// WrongIssuerToken returns a JWT with an incorrect issuer claim.
func WrongIssuerToken(t *testing.T, userID int64) string {
	t.Helper()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = TestJWTSecret
	}
	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = TestJWTAudience
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(30 * 24 * time.Hour).Unix(),
		"iss":     "http://wrong-issuer.com/",
		"aud":     audience,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to generate wrong issuer token: %v", err)
	}

	return tokenString
}

// WrongAudienceToken returns a JWT with an incorrect audience claim.
func WrongAudienceToken(t *testing.T, userID int64) string {
	t.Helper()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = TestJWTSecret
	}
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = TestJWTIssuer
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(30 * 24 * time.Hour).Unix(),
		"iss":     issuer,
		"aud":     "http://wrong-audience.com/",
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to generate wrong audience token: %v", err)
	}

	return tokenString
}
