// Package db provides database initialization and setup functionality.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// InitializeDatabase automatically creates the database and user if DATABASE_ROOT_PASSWORD is set.
// This is useful for Docker Compose deployments where the database should be auto-initialized.
// Returns true if database was created, false if it already existed or root password not set.
func InitializeDatabase(cfg *utils.Config) (bool, error) {
	// Skip if no root password provided
	if cfg.DatabaseRootPassword == "" {
		logger.L.Debug("skipping database initialization - no root password provided")
		return false, nil
	}

	logger.L.Info("checking if database needs initialization", 
		"database", cfg.DatabaseName,
		"host", cfg.DatabaseHost,
		"port", cfg.DatabasePort)

	// Connect to MySQL as root (without specifying database)
	dsn := fmt.Sprintf("root:%s@tcp(%s:%d)/",
		cfg.DatabaseRootPassword,
		cfg.DatabaseHost,
		cfg.DatabasePort,
	)

	logger.L.Debug("attempting to connect to MySQL as root")
	rootDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, fmt.Errorf("failed to connect as root: %w", err)
	}
	defer rootDB.Close()

	// Test connection
	logger.L.Debug("pinging MySQL server")
	if err := rootDB.Ping(); err != nil {
		return false, fmt.Errorf("failed to ping MySQL as root (check host/port/password): %w", err)
	}
	logger.L.Debug("successfully connected to MySQL as root")

	// Check if database exists
	var dbName string
	err = rootDB.QueryRow("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", cfg.DatabaseName).Scan(&dbName)
	if err == nil {
		// Database exists
		logger.L.Info("database already exists", "database", cfg.DatabaseName)
		return false, nil
	}
	if err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to check database existence: %w", err)
	}

	// Database doesn't exist, create it and the user
	logger.L.Info("database not found, creating database and user", "database", cfg.DatabaseName)

	// Create database
	if _, err := rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", cfg.DatabaseName)); err != nil {
		return false, fmt.Errorf("failed to create database: %w", err)
	}
	logger.L.Info("database created", "database", cfg.DatabaseName)

	// Create user if not exists (MySQL 5.7+ syntax)
	createUserSQL := fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'",
		cfg.DatabaseUser,
		cfg.DatabasePassword,
	)
	if _, err := rootDB.Exec(createUserSQL); err != nil {
		return false, fmt.Errorf("failed to create user: %w", err)
	}
	logger.L.Info("database user created", "user", cfg.DatabaseUser)

	// Grant privileges
	grantSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'",
		cfg.DatabaseName,
		cfg.DatabaseUser,
	)
	if _, err := rootDB.Exec(grantSQL); err != nil {
		return false, fmt.Errorf("failed to grant privileges: %w", err)
	}

	if _, err := rootDB.Exec("FLUSH PRIVILEGES"); err != nil {
		return false, fmt.Errorf("failed to flush privileges: %w", err)
	}
	logger.L.Info("privileges granted", "user", cfg.DatabaseUser)

	// Now read setup.sql and execute table creation (skip DB/user creation)
	setupSQL, err := os.ReadFile("setup.sql")
	if err != nil {
		return false, fmt.Errorf("failed to read setup.sql: %w", err)
	}

	// Split into individual statements
	statements := splitSQLStatements(string(setupSQL))

	// Execute each statement, skipping CREATE DATABASE, CREATE USER, GRANT, FLUSH
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		// Skip statements we already handled
		stmtUpper := strings.ToUpper(stmt)
		if strings.HasPrefix(stmtUpper, "CREATE DATABASE") ||
			strings.HasPrefix(stmtUpper, "CREATE USER") ||
			strings.HasPrefix(stmtUpper, "GRANT ") ||
			strings.HasPrefix(stmtUpper, "FLUSH PRIVILEGES") {
			logger.L.Debug("skipping already-handled statement", "index", i+1)
			continue
		}

		logger.L.Debug("executing SQL statement", "index", i+1, "statement", truncate(stmt, 100))
		
		if _, err := rootDB.Exec(stmt); err != nil {
			return false, fmt.Errorf("failed to execute statement %d: %w\nStatement: %s", i+1, err, stmt)
		}
	}

	logger.L.Info("database initialized successfully", "database", cfg.DatabaseName)
	return true, nil
}

// splitSQLStatements splits a SQL script into individual statements.
// Handles multi-line statements and ignores comments.
func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		
		current.WriteString(line)
		current.WriteString("\n")
		
		// Check if line ends with semicolon
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}
	
	// Add any remaining statement
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}
	
	return statements
}

// truncate truncates a string to maxLen characters for logging.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
