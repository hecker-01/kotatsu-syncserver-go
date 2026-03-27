// Package services implements business logic for user authentication, authorization,
// and domain operations. Services return typed errors for controllers to translate
// into appropriate HTTP responses.
package services

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

var (
	// ErrWrongPassword indicates authentication failed due to wrong password.
	ErrWrongPassword = errors.New("wrong password")
	// ErrInvalidInput indicates required fields are missing or malformed.
	ErrInvalidInput = errors.New("invalid input")
)

// AuthService handles user authentication (combined login/register).
type AuthService struct{}

// NewAuthService creates a new auth service instance.
func NewAuthService() *AuthService {
	return &AuthService{}
}

// allowNewRegister returns true if ALLOW_NEW_REGISTER is not set or set to "true".
func allowNewRegister() bool {
	val := os.Getenv("ALLOW_NEW_REGISTER")
	if val == "" {
		return true // default to true
	}
	return strings.ToLower(val) == "true"
}

// Authenticate handles combined login/register flow.
// If user exists: verifies password with Argon2 and returns JWT.
// If user doesn't exist and ALLOW_NEW_REGISTER=true: creates user and returns JWT.
// If user doesn't exist and ALLOW_NEW_REGISTER=false: returns ErrWrongPassword.
// Password must be 2-24 characters.
// Returns ErrInvalidInput for validation failures, ErrWrongPassword for auth failures.
func (s *AuthService) Authenticate(email, password string) (string, error) {
	// Validate input
	email = strings.TrimSpace(email)
	if email == "" {
		return "", ErrInvalidInput
	}

	// Validate password length (2-24 characters)
	if len(password) < 2 || len(password) > 24 {
		return "", ErrInvalidInput
	}

	// Try to find existing user
	var id int64
	var hash string
	err := db.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", email).Scan(&id, &hash)

	if err == nil {
		// User exists - verify password
		if !utils.VerifyPassword(password, hash) {
			return "", ErrWrongPassword
		}
		return utils.GenerateJWT(id)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Database error
		return "", err
	}

	// User doesn't exist - check if registration is allowed
	if !allowNewRegister() {
		return "", ErrWrongPassword
	}

	// Create new user with Argon2 hashed password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}

	result, err := db.DB.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", email, hashedPassword)
	if err != nil {
		// Could be a race condition where user was created between check and insert
		return "", ErrWrongPassword
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	return utils.GenerateJWT(newID)
}
