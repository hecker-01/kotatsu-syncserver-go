package services

import (
	"context"
	"errors"

	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// ErrUnauthorized indicates the request lacks valid authentication.
var ErrUnauthorized = errors.New("unauthorized")

// UserService handles user-related business logic.
type UserService struct{}

// NewUserService creates a new user service instance.
func NewUserService() *UserService {
	return &UserService{}
}

// CurrentUserID extracts the authenticated user ID from request context.
// Returns ErrUnauthorized if no user_id is found in context.
func (s *UserService) CurrentUserID(ctx context.Context) (int, error) {
	userID, ok := utils.UserIDFromContext(ctx)
	if !ok {
		return 0, ErrUnauthorized
	}
	return userID, nil
}
