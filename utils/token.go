package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// GenerateJWT creates a signed access token for the given user ID.
// The token expires after 24 hours.
func GenerateJWT(userID int) (string, error) {
	secret, err := JWTSecret()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(secret)
}

// ParseAndValidateJWT parses and validates a JWT string and returns the user ID
// claim when the token is valid.
func ParseAndValidateJWT(tokenString string) (int, error) {
	secret, err := JWTSecret()
	if err != nil {
		return 0, err
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return secret, nil
	})
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

	switch userID := userIDValue.(type) {
	case float64:
		return int(userID), nil
	case int:
		return userID, nil
	default:
		return 0, ErrInvalidToken
	}
}
