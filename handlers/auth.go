// Package handlers contains legacy HTTP handlers that directly implement
// request handling logic. New code should use the controllers/ + services/ pattern.
package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/logger"
)

// jwtSecret retrieves the JWT signing secret from environment.
// Exits if JWT_SECRET is not configured.
func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logger.Error("JWT_SECRET is not set")
		os.Exit(1)
	}
	return []byte(secret)
}

// RegisterRequest holds the JSON payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register is a legacy handler for user registration.
// New code should use controllers.AuthController.Register instead.
func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("bcrypt failed", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", req.Email, string(hash))
	if err != nil {
		logger.Warn("register failed", "email", req.Email, "error", err)
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	logger.Info("user registered", "email", req.Email)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created"))
}

// LoginRequest holds the JSON payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login is a legacy handler for user authentication.
// New code should use controllers.AuthController.Login instead.
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var id int
	var hash string
	err := db.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", req.Email).Scan(&id, &hash)
	if err != nil {
		logger.Warn("login attempt failed", "email", req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		logger.Warn("wrong password", "email", req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret())
	if err != nil {
		logger.Error("token signing failed", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	logger.Info("user logged in", "user_id", id)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
