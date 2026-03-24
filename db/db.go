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

func getEnv(key string) string {
	return os.Getenv(key)
}

// Init establishes the database connection using environment variables
// (DB_HOST, DB_NAME, DB_USER, DB_PASS, DB_PORT) and configures the connection pool.
// Exits with code 1 if connection fails or required variables are missing.
func Init() {
	host := getEnv("DB_HOST")
	name := getEnv("DB_NAME")
	user := getEnv("DB_USER")
	pass := getEnv("DB_PASS")
	port := getEnv("DB_PORT")
	if port == "" {
		port = "3306"
	}

	if host == "" || name == "" || user == "" || pass == "" {
		logger.Error("DB_HOST, DB_NAME, DB_USER, and DB_PASS must be set")
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
