package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/handlers"
	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
	authmw "github.com/hecker-01/kotatsu-syncserver-go/middleware"
)

func ensureRequiredEnv(required []string) {
	missing := make([]string, 0)
	for _, key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) == 0 {
		return
	}

	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "missing required env vars: %s; also failed to load .env: %v\n", strings.Join(missing, ", "), err)
		os.Exit(1)
	}

	stillMissing := make([]string, 0)
	for _, key := range required {
		if os.Getenv(key) == "" {
			stillMissing = append(stillMissing, key)
		}
	}

	if len(stillMissing) > 0 {
		fmt.Fprintf(os.Stderr, "required env vars are missing after loading .env: %s\n", strings.Join(stillMissing, ", "))
		os.Exit(1)
	}
}

func main() {
	// Load .env early so optional values like PORT are available.
	_ = godotenv.Load()

	ensureRequiredEnv([]string{"DB_HOST", "DB_NAME", "DB_USER", "DB_PASS", "JWT_SECRET"})

	logger.Init()

	db.Init()

	r := chi.NewRouter()
	r.Use(middleware.StructuredLogger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Post("/register", handlers.Register)
	r.Post("/login", handlers.Login)

	r.Group(func(r chi.Router) {
		r.Use(authmw.RequireAuth)
		r.Get("/me", handlers.Me)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + strings.TrimPrefix(port, ":")
	logger.L.Info("server starting", "port", port)

	if err := http.ListenAndServe(addr, r); err != nil {
		logger.L.Error("server failed to start", "addr", addr, "error", err)
		os.Exit(1)
	}
}
