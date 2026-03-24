package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func getEnv(key string) string {
	return os.Getenv(key)
}

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
		slog.Error("DB_HOST, DB_NAME, DB_USER, and DB_PASS must be set")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err = DB.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	slog.Info("database connected")
}
