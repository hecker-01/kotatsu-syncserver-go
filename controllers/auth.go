package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/services"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{authService: services.NewAuthService()}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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
