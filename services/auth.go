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
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidInput       = errors.New("invalid input")
)

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

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
