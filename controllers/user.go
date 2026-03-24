package controllers

import (
	"errors"
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// UserController handles user-related endpoints.
type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new user controller with initialized service.
func NewUserController() *UserController {
	return &UserController{userService: services.NewUserService()}
}

// Me handles GET /api/users/me. Returns the authenticated user's ID from context.
// Requires authentication middleware.
func (c *UserController) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := c.userService.CurrentUserID(r.Context())
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Server error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"user_id": userID})
}
