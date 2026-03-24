package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
)

func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		slog.Error("JWT_SECRET is not set")
		os.Exit(1)
	}
	return []byte(secret)
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("bcrypt failed", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", req.Email, string(hash))
	if err != nil {
		slog.Warn("register failed", "email", req.Email, "error", err)
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	slog.Info("user registered", "email", req.Email)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created"))
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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
		slog.Warn("login attempt failed", "email", req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		slog.Warn("wrong password", "email", req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret())
	if err != nil {
		slog.Error("token signing failed", "error", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	slog.Info("user logged in", "user_id", id)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	json.NewEncoder(w).Encode(map[string]interface{}{"user_id": userID})
}
