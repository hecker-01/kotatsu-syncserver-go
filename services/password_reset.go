// Package services - password_reset.go implements password reset functionality.
// Handles password reset token generation, validation, and password updates.
package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

var (
	// ErrInvalidToken indicates the reset token is invalid or expired.
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrPasswordLength indicates the password does not meet length requirements.
	ErrPasswordLength = errors.New("password should be from 2 to 24 characters long")
)

// PasswordResetService handles password reset operations.
type PasswordResetService struct{}

// NewPasswordResetService creates a new password reset service instance.
func NewPasswordResetService() *PasswordResetService {
	return &PasswordResetService{}
}

// ResetPassword validates the reset token and updates the user's password.
// The token is hashed with SHA256 and matched against password_reset_token_hash.
// Returns ErrPasswordLength if password is not 2-24 characters.
// Returns ErrInvalidToken if the token is invalid or expired.
func (s *PasswordResetService) ResetPassword(resetToken string, newPassword string) error {
	// Validate password length
	if len(newPassword) < 2 || len(newPassword) > 24 {
		return ErrPasswordLength
	}

	// Hash the token with SHA256
	tokenHash := sha256.Sum256([]byte(resetToken))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	// Look up user by token hash and check expiration
	var userID int64
	var expiresAt int64
	err := db.DB.QueryRow(
		"SELECT id, password_reset_token_expires_at FROM users WHERE password_reset_token_hash = ?",
		tokenHashHex,
	).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}
		return err
	}

	// Check if token has expired (expiresAt is in seconds)
	if expiresAt <= time.Now().Unix() {
		return ErrInvalidToken
	}

	// Hash new password with Argon2
	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and clear reset token fields
	_, err = db.DB.Exec(
		"UPDATE users SET password_hash = ?, password_reset_token_hash = NULL, password_reset_token_expires_at = NULL WHERE id = ?",
		passwordHash,
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}

// GenerateResetToken generates a cryptographically secure reset token and its hash.
// Returns the raw base64-encoded token (to send to user) and its SHA256 hash (to store in DB).
func GenerateResetToken() (token string, hash string, err error) {
	// Generate 32 random bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64 URL-safe (for use in URLs)
	token = base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token with SHA256 for storage
	tokenHash := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(tokenHash[:])

	return token, hash, nil
}

// RequestPasswordReset handles the forgot-password flow.
// Generates a reset token if the user exists and has no active reset token.
// Sends an email with the reset link.
// Always returns nil to prevent email enumeration (caller should always return 200).
func (s *PasswordResetService) RequestPasswordReset(email string, mailService MailService, baseURL string) error {
	// Check if user exists
	var userID int64
	var expiresAt sql.NullInt64
	err := db.DB.QueryRow(
		"SELECT id, password_reset_token_expires_at FROM users WHERE email = ?",
		email,
	).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User not found - silently return to prevent email enumeration
			logger.Debug("password reset requested for non-existent email", "email", email)
			return nil
		}
		// Database error - log but don't expose
		logger.Error("database error during password reset lookup", "error", err)
		return nil
	}

	// Check if there's an active reset token (not expired)
	now := time.Now().Unix()
	if expiresAt.Valid && expiresAt.Int64 > now {
		// Active token exists - silently return to prevent enumeration
		logger.Debug("password reset requested but active token exists", "user_id", userID)
		return nil
	}

	// Generate new reset token
	token, tokenHash, err := GenerateResetToken()
	if err != nil {
		logger.Error("failed to generate reset token", "error", err)
		return nil
	}

	// Set expiration to 1 hour from now (in seconds)
	expiresAtUnix := now + 3600

	// Store token hash and expiration in database
	_, err = db.DB.Exec(
		"UPDATE users SET password_reset_token_hash = ?, password_reset_token_expires_at = ? WHERE id = ?",
		tokenHash,
		expiresAtUnix,
		userID,
	)
	if err != nil {
		logger.Error("failed to store reset token", "error", err)
		return nil
	}

	// Build reset link
	resetLink := fmt.Sprintf("%s/deeplink/reset-password?token=%s", baseURL, token)

	// Send email
	subject := "Password Reset Request"
	textBody := fmt.Sprintf("Click the link to reset your password: %s", resetLink)
	htmlBody := fmt.Sprintf("<p>Click the link to reset your password:</p><p><a href=\"%s\">%s</a></p>", resetLink, resetLink)

	if err := mailService.Send(email, subject, textBody, htmlBody); err != nil {
		logger.Error("failed to send password reset email", "error", err, "email", email)
		// Don't return error to prevent enumeration
		return nil
	}

	logger.Info("password reset email sent", "user_id", userID)
	return nil
}
