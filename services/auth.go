// Package services implements business logic for user authentication, authorization,
// and domain operations. Services return typed errors for controllers to translate
// into appropriate HTTP responses.
package services

import (
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

var (
	// ErrInvalidCredentials indicates login failed due to wrong email or password.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrEmailExists indicates registration failed because the email is already registered.
	ErrEmailExists = errors.New("email already exists")
	// ErrInvalidInput indicates required fields are missing or malformed.
	ErrInvalidInput = errors.New("invalid input")
)

// RegisterInput holds the data required to create a new user account.
type RegisterInput struct {
	Email    string
	Password string
}

// LoginInput holds the data required to authenticate a user.
type LoginInput struct {
	Email    string
	Password string
}

// AuthService handles user registration and login operations.
type AuthService struct{}

// NewAuthService creates a new auth service instance.
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register creates a new user account with the provided email and password.
// Passwords are hashed with bcrypt before storage.
// Returns ErrInvalidInput if email/password are empty, ErrEmailExists if email is taken.
func (s *AuthService) Register(input RegisterInput) error {
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" {
		return ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", input.Email, string(hash))
	if err != nil {
		return ErrEmailExists
	}

	return nil
}

// Login authenticates a user and returns a JWT token and user ID.
// Returns ErrInvalidInput if email/password are empty, ErrInvalidCredentials if auth fails.
func (s *AuthService) Login(input LoginInput) (string, int, error) {
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" {
		return "", 0, ErrInvalidInput
	}

	var id int
	var hash string
	err := db.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", input.Email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, ErrInvalidCredentials
		}
		return "", 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)); err != nil {
		return "", 0, ErrInvalidCredentials
	}

	tokenString, err := utils.GenerateJWT(id)
	if err != nil {
		return "", 0, err
	}

	return tokenString, id, nil
}
