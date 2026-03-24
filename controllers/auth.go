// Package controllers provides HTTP handlers that translate between HTTP requests/responses
// and service-layer business logic. Controllers handle JSON parsing, error mapping to status codes,
// and response formatting.
package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// AuthController handles authentication endpoints (register, login).
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller with initialized service.
func NewAuthController() *AuthController {
	return &AuthController{authService: services.NewAuthService()}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles POST /api/auth/register. Creates a new user account with email and password.
// Returns 201 on success, 400 for invalid input, 409 if email already exists.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	err := c.authService.Register(services.RegisterInput{Email: req.Email, Password: req.Password})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			utils.WriteError(w, http.StatusBadRequest, "Email and password are required")
		case errors.Is(err, services.ErrEmailExists):
			utils.WriteError(w, http.StatusConflict, "Email already exists")
		default:
			logger.Error("register failed", "error", err)
			utils.WriteError(w, http.StatusInternalServerError, "Server error")
		}
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": "User created"})
}

// Login handles POST /api/auth/login. Authenticates user and returns a JWT token.
// Returns 200 with token on success, 400 for invalid input, 401 for invalid credentials.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, userID, err := c.authService.Login(services.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			utils.WriteError(w, http.StatusBadRequest, "Email and password are required")
		case errors.Is(err, services.ErrInvalidCredentials):
			utils.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		default:
			logger.Error("login failed", "error", err)
			utils.WriteError(w, http.StatusInternalServerError, "Server error")
		}
		return
	}

	logger.Info("user logged in", "user_id", userID)
	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
