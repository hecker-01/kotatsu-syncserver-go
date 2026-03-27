package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// ErrUnauthorized indicates the request lacks valid authentication.
var ErrUnauthorized = errors.New("unauthorized")

// ErrNotFound indicates the requested resource does not exist.
var ErrNotFound = errors.New("not found")

// UserService handles user-related business logic.
type UserService struct{}

// NewUserService creates a new user service instance.
func NewUserService() *UserService {
	return &UserService{}
}

// CurrentUserID extracts the authenticated user ID from request context.
// Returns ErrUnauthorized if no user_id is found in context.
// User ID is int64 to support BIGINT database IDs.
func (s *UserService) CurrentUserID(ctx context.Context) (int64, error) {
	userID, ok := utils.UserIDFromContext(ctx)
	if !ok {
		return 0, ErrUnauthorized
	}
	return userID, nil
}

// GetUserByID retrieves a user by their ID and returns a safe response (no password).
// Returns ErrNotFound if the user does not exist.
func (s *UserService) GetUserByID(id int64) (*models.UserResponse, error) {
	var user models.UserResponse
	err := db.DB.QueryRow(
		"SELECT id, email, nickname FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.Nickname)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
