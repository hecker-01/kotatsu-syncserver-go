package controllers

import (
	"errors"
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController() *UserController {
	return &UserController{userService: services.NewUserService()}
}

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
