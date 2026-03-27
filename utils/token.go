package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT configuration constants matching Kotatsu API specification.
const (
	// DefaultJWTIssuer is the default value for the JWT issuer claim.
	DefaultJWTIssuer = "http://0.0.0.0:9292/"
	// DefaultJWTAudience is the default value for the JWT audience claim.
	DefaultJWTAudience = "http://0.0.0.0:9292/resource"
	// JWTExpiration is the token lifetime (30 days).
	JWTExpiration = 30 * 24 * time.Hour
)

// ErrInvalidToken indicates that a JWT is malformed, uses an unsupported
// signing method, is missing required claims, or failed validation.
var ErrInvalidToken = errors.New("invalid token")

// JWTSecret returns the JWT signing secret from environment configuration.
// It returns an error when JWT_SECRET is not configured.
func JWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	return []byte(secret), nil
}

// jwtIssuer returns the JWT issuer from environment or the default.
func jwtIssuer() string {
	if iss := os.Getenv("JWT_ISSUER"); iss != "" {
		return iss
	}
	return DefaultJWTIssuer
}

// jwtAudience returns the JWT audience from environment or the default.
func jwtAudience() string {
	if aud := os.Getenv("JWT_AUDIENCE"); aud != "" {
		return aud
	}
	return DefaultJWTAudience
}

// GenerateJWT creates a signed access token for the given user ID.
// The token includes user_id, exp, iss, and aud claims per Kotatsu API spec.
// Token expires after 30 days.
func GenerateJWT(userID int64) (string, error) {
	secret, err := JWTSecret()
	if err != nil {
		return "", err
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(JWTExpiration).Unix(),
		"iss":     jwtIssuer(),
		"aud":     jwtAudience(),
	})

	return token.SignedString(secret)
}

// ParseAndValidateJWT parses and validates a JWT string and returns the user ID.
// Validates signature, issuer, audience, and expiration per Kotatsu API spec.
func ParseAndValidateJWT(tokenString string) (int64, error) {
	secret, err := JWTSecret()
	if err != nil {
		return 0, err
	}

	expectedIssuer := jwtIssuer()
	expectedAudience := jwtAudience()

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	},
		jwt.WithIssuer(expectedIssuer),
		jwt.WithAudience(expectedAudience),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	userIDValue, ok := claims["user_id"]
	if !ok {
		return 0, ErrInvalidToken
	}

	// JSON numbers are decoded as float64
	switch userID := userIDValue.(type) {
	case float64:
		return int64(userID), nil
	case int64:
		return userID, nil
	default:
		return 0, ErrInvalidToken
	}
}
