// Package controllers provides HTTP handlers that translate between HTTP requests/responses
// and service-layer business logic. Controllers handle JSON parsing, error mapping to status codes,
// and response formatting.
package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// AuthController handles the combined authentication endpoint.
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller with initialized service.
func NewAuthController() *AuthController {
	return &AuthController{authService: services.NewAuthService()}
}

// Auth handles POST /auth. Combined login/register endpoint.
// If user exists: authenticates with password.
// If user doesn't exist and registration allowed: creates new account.
// Returns 200 with token on success, 400 with plain text "Wrong password" on auth failure.
func (c *AuthController) Auth(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Wrong password", http.StatusBadRequest)
		return
	}

	token, err := c.authService.Authenticate(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			http.Error(w, "Wrong password", http.StatusBadRequest)
		case errors.Is(err, services.ErrWrongPassword):
			http.Error(w, "Wrong password", http.StatusBadRequest)
		default:
			logger.Error("auth failed", "error", err)
			http.Error(w, "Wrong password", http.StatusBadRequest)
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.AuthResponse{Token: token})
}
