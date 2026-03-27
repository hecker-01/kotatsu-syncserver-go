// Package db provides database initialization and a global connection pool
// for MySQL database access throughout the application.
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hecker-01/kotatsu-syncserver-go/logger"
)

// DB is the global database connection pool used by services.
var DB *sql.DB

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Init establishes the database connection using environment variables
// (DATABASE_HOST, DATABASE_NAME, DATABASE_USER, DATABASE_PASSWORD, DATABASE_PORT)
// and configures the connection pool.
// Exits with code 1 if connection fails or required variables are missing.
func Init() {
	host := getEnvOrDefault("DATABASE_HOST", "localhost")
	name := getEnvOrDefault("DATABASE_NAME", "kotatsu_db")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASSWORD")
	port := getEnvOrDefault("DATABASE_PORT", "3306")

	if user == "" || pass == "" {
		logger.Error("DATABASE_USER and DATABASE_PASSWORD must be set")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err = DB.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	logger.Info("database connected")
}
