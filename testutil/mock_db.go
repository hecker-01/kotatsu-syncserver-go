package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// Test database environment variables.
const (
	TestDBHostEnv = "TEST_DATABASE_HOST"
	TestDBUserEnv = "TEST_DATABASE_USER"
	TestDBPassEnv = "TEST_DATABASE_PASSWORD"
	TestDBNameEnv = "TEST_DATABASE_NAME"
	TestDBPortEnv = "TEST_DATABASE_PORT"
)

// Default test database configuration.
const (
	DefaultTestDBHost = "localhost"
	DefaultTestDBPort = "3306"
	DefaultTestDBName = "kotatsu_test"
)

var (
	testDB     *sql.DB
	testDBOnce sync.Once
	testDBErr  error
)

// SetupTestDB initializes and returns a test database connection.
// Uses TEST_DATABASE_* environment variables or defaults to test database settings.
// The connection is cached and reused across tests in the same process.
// Skips the test if no database is available.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDBOnce.Do(func() {
		testDB, testDBErr = connectTestDB()
	})

	if testDBErr != nil {
		t.Skipf("test database not available: %v", testDBErr)
	}

	return testDB
}

// GetTestDB returns the test database connection, or nil if not available.
// Does not fail or skip the test if database is unavailable.
func GetTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDBOnce.Do(func() {
		testDB, testDBErr = connectTestDB()
	})

	if testDBErr != nil {
		return nil
	}

	return testDB
}

// connectTestDB establishes a connection to the test database.
func connectTestDB() (*sql.DB, error) {
	host := getEnvOrDefault(TestDBHostEnv, DefaultTestDBHost)
	port := getEnvOrDefault(TestDBPortEnv, DefaultTestDBPort)
	name := getEnvOrDefault(TestDBNameEnv, DefaultTestDBName)
	user := os.Getenv(TestDBUserEnv)
	pass := os.Getenv(TestDBPassEnv)

	// If test credentials aren't set, try falling back to regular DATABASE_* vars
	if user == "" {
		user = os.Getenv("DATABASE_USER")
	}
	if pass == "" {
		pass = os.Getenv("DATABASE_PASSWORD")
	}

	// Check required credentials
	if user == "" || pass == "" {
		return nil, fmt.Errorf("database credentials not configured (set %s and %s)", TestDBUserEnv, TestDBPassEnv)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		user, pass, host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool for tests
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	return db, nil
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// WithTestDB runs a test function with a database connection.
// Skips the test if database is not available.
// Cleans up test data before and after the test function.
func WithTestDB(t *testing.T, fn func(db *sql.DB)) {
	t.Helper()

	db := SetupTestDB(t)

	// Clean before running
	CleanupTestDataWithDB(t, db)

	// Run the test
	fn(db)

	// Clean after running
	CleanupTestDataWithDB(t, db)
}

// RequireTestDB returns the test database connection or fails the test.
// Use this instead of SetupTestDB when the test cannot be skipped.
func RequireTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDBOnce.Do(func() {
		testDB, testDBErr = connectTestDB()
	})

	if testDBErr != nil {
		t.Fatalf("test database required but not available: %v", testDBErr)
	}

	return testDB
}

// TruncateTables clears specified tables using the default test connection.
func TruncateTables(t *testing.T, tables ...string) {
	t.Helper()

	db := GetTestDB(t)
	if db == nil {
		return
	}

	TruncateTablesWithDB(t, db, tables...)
}

// TruncateTablesWithDB clears specified tables using the provided connection.
// Disables foreign key checks temporarily to allow truncation in any order.
func TruncateTablesWithDB(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()

	// Disable foreign key checks for truncation
	_, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		t.Logf("warning: failed to disable foreign key checks: %v", err)
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table))
		if err != nil {
			// Try DELETE if TRUNCATE fails (some tables might have restrictions)
			_, err = db.Exec(fmt.Sprintf("DELETE FROM %s", table))
			if err != nil {
				t.Logf("warning: failed to clear table %s: %v", table, err)
			}
		}
	}

	// Re-enable foreign key checks
	_, err = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		t.Logf("warning: failed to re-enable foreign key checks: %v", err)
	}
}

// ResetAutoIncrement resets the auto-increment counter for a table.
func ResetAutoIncrement(t *testing.T, db *sql.DB, table string) {
	t.Helper()

	_, err := db.Exec(fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", table))
	if err != nil {
		t.Logf("warning: failed to reset auto increment for %s: %v", table, err)
	}
}

// TableExists checks if a table exists in the test database.
func TableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()

	var exists int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() AND table_name = ?
	`, table).Scan(&exists)

	if err != nil {
		t.Logf("warning: failed to check table existence: %v", err)
		return false
	}

	return exists > 0
}

// CreateTestTables creates the required tables if they don't exist.
// This is useful for initializing a fresh test database.
func CreateTestTables(t *testing.T, db *sql.DB) {
	t.Helper()

	// Users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(320) NOT NULL,
			password_hash VARCHAR(128) NOT NULL,
			nickname VARCHAR(100),
			favourites_sync_timestamp BIGINT,
			history_sync_timestamp BIGINT,
			password_reset_token_hash CHAR(64),
			password_reset_token_expires_at BIGINT,
			CONSTRAINT uq_users_email UNIQUE (email),
			UNIQUE INDEX uq_users_password_reset_token_hash (password_reset_token_hash)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Manga table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga (
			id BIGINT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			alt_title VARCHAR(255),
			url VARCHAR(255) NOT NULL,
			public_url VARCHAR(255) NOT NULL,
			rating FLOAT NOT NULL,
			content_rating ENUM('SAFE', 'SUGGESTIVE', 'ADULT'),
			cover_url VARCHAR(255) NOT NULL,
			large_cover_url VARCHAR(255),
			state ENUM('ONGOING', 'FINISHED', 'ABANDONED', 'PAUSED', 'UPCOMING', 'RESTRICTED'),
			author VARCHAR(64),
			source VARCHAR(32) NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create manga table: %v", err)
	}

	// Tags table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id BIGINT PRIMARY KEY,
			title VARCHAR(64) NOT NULL,
			` + "`key`" + ` VARCHAR(120) NOT NULL,
			source VARCHAR(32) NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tags table: %v", err)
	}

	// Manga-Tags junction table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_tags (
			manga_id BIGINT NOT NULL,
			tag_id BIGINT NOT NULL,
			PRIMARY KEY (manga_id, tag_id),
			INDEX idx_manga_tags_tag_id (tag_id),
			CONSTRAINT fk_manga_tags_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id),
			CONSTRAINT fk_manga_tags_manga_id FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create manga_tags table: %v", err)
	}

	// Categories table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id BIGINT NOT NULL,
			created_at BIGINT NOT NULL,
			sort_key INT NOT NULL,
			title VARCHAR(120) NOT NULL,
			` + "`order`" + ` VARCHAR(16) NOT NULL,
			user_id BIGINT NOT NULL,
			track TINYINT(1) NOT NULL,
			show_in_lib TINYINT(1) NOT NULL,
			deleted_at BIGINT,
			PRIMARY KEY (id, user_id),
			CONSTRAINT fk_categories_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create categories table: %v", err)
	}

	// Favourites table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS favourites (
			manga_id BIGINT NOT NULL,
			category_id BIGINT NOT NULL,
			sort_key INT NOT NULL,
			pinned TINYINT(1) NOT NULL,
			created_at BIGINT NOT NULL,
			deleted_at BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			PRIMARY KEY (manga_id, category_id, user_id),
			INDEX idx_favourites_user_id (user_id),
			CONSTRAINT fk_favourites_manga_id FOREIGN KEY (manga_id) REFERENCES manga(id),
			CONSTRAINT fk_favourites_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT fk_favourites_category FOREIGN KEY (category_id, user_id) REFERENCES categories(id, user_id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create favourites table: %v", err)
	}

	// History table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			manga_id BIGINT NOT NULL,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL,
			chapter_id BIGINT NOT NULL,
			page SMALLINT NOT NULL,
			scroll DOUBLE NOT NULL,
			percent DOUBLE NOT NULL,
			chapters INT NOT NULL,
			deleted_at BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			PRIMARY KEY (user_id, manga_id),
			INDEX idx_manga_id (manga_id),
			CONSTRAINT fk_history_manga_id FOREIGN KEY (manga_id) REFERENCES manga(id),
			CONSTRAINT fk_history_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("failed to create history table: %v", err)
	}
}

// DropTestTables drops all test tables. Use with caution.
func DropTestTables(t *testing.T, db *sql.DB) {
	t.Helper()

	// Disable foreign key checks
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")

	tables := []string{
		"history",
		"favourites",
		"categories",
		"manga_tags",
		"tags",
		"manga",
		"users",
	}

	for _, table := range tables {
		db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	}

	// Re-enable foreign key checks
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
}
