// Package main provides the entry point for the Kotatsu Sync Server HTTP API.
// It configures environment variables, initializes database and logging,
// sets up middleware and routes, then starts the HTTP server.
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
	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/middleware"
	"github.com/hecker-01/kotatsu-syncserver-go/routes"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

func main() {
	// Load .env early so all configuration values are available.
	_ = godotenv.Load()

	// Load and validate configuration from environment variables.
	cfg, err := utils.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	logger.Init()

	// Auto-initialize database if DATABASE_ROOT_PASSWORD is set
	// This is useful for Docker Compose where database should be created automatically
	if created, err := db.InitializeDatabase(cfg); err != nil {
		logger.L.Error("failed to initialize database", "error", err)
		os.Exit(1)
	} else if created {
		logger.L.Info("database created successfully")
	}

	db.Init()

	r := chi.NewRouter()
	r.Use(middleware.StructuredLogger)
	r.Use(chimiddleware.Recoverer)

	// Health check at root - returns plain text "Alive"
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Alive"))
	})

	// Register all routes at root level (matching Kotatsu API structure)
	routes.RegisterRoutes(r)

	port := fmt.Sprintf("%d", cfg.Port)
	addr := ":" + strings.TrimPrefix(port, ":")
	logger.L.Info("server starting", "port", port)

	if err := http.ListenAndServe(addr, r); err != nil {
		logger.L.Error("server failed to start", "addr", addr, "error", err)
		os.Exit(1)
	}
}
