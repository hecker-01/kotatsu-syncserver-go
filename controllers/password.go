// Package controllers - password.go handles password-related HTTP endpoints.
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

// PasswordController handles password-related endpoints.
type PasswordController struct {
	passwordResetService *services.PasswordResetService
	mailService          services.MailService
	config               *utils.Config
}

// NewPasswordController creates a new password controller with initialized services.
func NewPasswordController() *PasswordController {
	cfg, _ := utils.LoadConfig()
	return &PasswordController{
		passwordResetService: services.NewPasswordResetService(),
		mailService:          services.NewMailService(cfg),
		config:               cfg,
	}
}

// ResetPassword handles POST /api/auth/reset-password.
// Validates reset token and updates user password.
// Returns 200 on success, 400 for invalid token or password length issues.
func (c *PasswordController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate password length first (return specific message)
	if len(req.Password) < 2 || len(req.Password) > 24 {
		utils.WriteError(w, http.StatusBadRequest, "Password should be from 2 to 24 characters long")
		return
	}

	err := c.passwordResetService.ResetPassword(req.ResetToken, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPasswordLength):
			utils.WriteError(w, http.StatusBadRequest, "Password should be from 2 to 24 characters long")
		case errors.Is(err, services.ErrInvalidToken):
			utils.WriteError(w, http.StatusBadRequest, "Invalid or expired token")
		default:
			logger.Error("password reset failed", "error", err)
			utils.WriteError(w, http.StatusInternalServerError, "Server error")
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Password reset successfully")
}

// ForgotPassword handles POST /api/auth/forgot-password.
// Initiates the password reset flow by sending a reset email.
// Always returns 200 to prevent email enumeration attacks.
func (c *PasswordController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Even on parse error, return 200 to prevent enumeration
		utils.WriteJSON(w, http.StatusOK, "A password reset email was sent")
		return
	}

	// Call service - it always returns nil to prevent enumeration
	_ = c.passwordResetService.RequestPasswordReset(req.Email, c.mailService, c.config.BaseURL)

	// Always return 200 with the same message
	utils.WriteJSON(w, http.StatusOK, "A password reset email was sent")
}
