package services

import (
	"context"
	"errors"

	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

var ErrUnauthorized = errors.New("unauthorized")

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) CurrentUserID(ctx context.Context) (int, error) {
	userID, ok := utils.UserIDFromContext(ctx)
	if !ok {
		return 0, ErrUnauthorized
	}
	return userID, nil
}
