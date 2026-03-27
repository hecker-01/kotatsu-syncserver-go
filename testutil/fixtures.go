package testutil

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// Test user credentials - use these for consistent test data.
const (
	TestEmail    = "test@example.com"
	TestPassword = "testpass123"
	TestNickname = "TestUser"
)

// Additional test users for multi-user scenarios.
const (
	TestEmail2    = "test2@example.com"
	TestPassword2 = "testpass456"
	TestNickname2 = "TestUser2"
)

// TestUser represents a user created for testing.
type TestUser struct {
	ID           int64
	Email        string
	Password     string
	PasswordHash string
	Nickname     *string
}

// CreateTestUser creates a user in the database for testing.
// Returns the user ID. Uses the default test credentials.
func CreateTestUser(t *testing.T) int64 {
	t.Helper()

	db := GetTestDB(t)
	if db == nil {
		t.Skip("test database not available")
	}

	return CreateTestUserWithCredentials(t, db, TestEmail, TestPassword, TestNickname)
}

// CreateTestUserWithDB creates a test user using the provided database connection.
func CreateTestUserWithDB(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	return CreateTestUserWithCredentials(t, db, TestEmail, TestPassword, TestNickname)
}

// CreateTestUserWithCredentials creates a user with custom credentials.
func CreateTestUserWithCredentials(t *testing.T, db *sql.DB, email, password, nickname string) int64 {
	t.Helper()

	// Hash the password
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Insert the user
	result, err := db.Exec(
		"INSERT INTO users (email, password_hash, nickname) VALUES (?, ?, ?)",
		email, passwordHash, nickname,
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	return userID
}

// CreateTestUserFull creates a user and returns all details.
func CreateTestUserFull(t *testing.T, db *sql.DB, email, password, nickname string) *TestUser {
	t.Helper()

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	result, err := db.Exec(
		"INSERT INTO users (email, password_hash, nickname) VALUES (?, ?, ?)",
		email, passwordHash, nickname,
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get user ID: %v", err)
	}

	nick := nickname
	return &TestUser{
		ID:           userID,
		Email:        email,
		Password:     password,
		PasswordHash: passwordHash,
		Nickname:     &nick,
	}
}

// CreateTestManga creates manga entries for testing.
// Returns the IDs of the created manga entries.
func CreateTestManga(t *testing.T, count int) []int64 {
	t.Helper()

	db := GetTestDB(t)
	if db == nil {
		t.Skip("test database not available")
	}

	return CreateTestMangaWithDB(t, db, count)
}

// CreateTestMangaWithDB creates manga entries using the provided database connection.
func CreateTestMangaWithDB(t *testing.T, db *sql.DB, count int) []int64 {
	t.Helper()

	ids := make([]int64, count)
	now := time.Now().UnixMilli()

	for i := 0; i < count; i++ {
		// Generate unique manga ID based on timestamp
		mangaID := now + int64(i)

		_, err := db.Exec(`
			INSERT INTO manga (id, title, url, public_url, rating, cover_url, source)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			mangaID,
			fmt.Sprintf("Test Manga %d", i+1),
			fmt.Sprintf("https://source.com/manga/%d", mangaID),
			fmt.Sprintf("https://public.com/manga/%d", mangaID),
			4.5,
			fmt.Sprintf("https://covers.com/%d.jpg", mangaID),
			"test_source",
		)
		if err != nil {
			t.Fatalf("failed to create test manga %d: %v", i+1, err)
		}

		ids[i] = mangaID
	}

	return ids
}

// CreateTestMangaFull creates a manga entry with all fields populated.
func CreateTestMangaFull(t *testing.T, db *sql.DB, id int64, title string) int64 {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO manga (id, title, alt_title, url, public_url, rating, content_rating, cover_url, large_cover_url, state, author, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		title,
		"Alt "+title,
		fmt.Sprintf("https://source.com/manga/%d", id),
		fmt.Sprintf("https://public.com/manga/%d", id),
		4.5,
		"SAFE",
		fmt.Sprintf("https://covers.com/%d.jpg", id),
		fmt.Sprintf("https://covers.com/%d_large.jpg", id),
		"ONGOING",
		"Test Author",
		"test_source",
	)
	if err != nil {
		t.Fatalf("failed to create full test manga: %v", err)
	}

	return id
}

// CreateTestCategory creates a category for the given user.
func CreateTestCategory(t *testing.T, db *sql.DB, userID int64, title string) int64 {
	t.Helper()

	now := time.Now().UnixMilli()
	categoryID := now

	_, err := db.Exec(`
		INSERT INTO categories (id, created_at, sort_key, title, ` + "`order`" + `, user_id, track, show_in_lib)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		categoryID, now, 0, title, "NAME", userID, true, true,
	)
	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	return categoryID
}

// CreateTestFavourite adds a manga to a user's favourites.
func CreateTestFavourite(t *testing.T, db *sql.DB, userID, mangaID, categoryID int64) {
	t.Helper()

	now := time.Now().UnixMilli()

	_, err := db.Exec(`
		INSERT INTO favourites (manga_id, category_id, sort_key, pinned, created_at, deleted_at, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		mangaID, categoryID, 0, false, now, 0, userID,
	)
	if err != nil {
		t.Fatalf("failed to create test favourite: %v", err)
	}
}

// CreateTestHistory creates a history entry for a user.
func CreateTestHistory(t *testing.T, db *sql.DB, userID, mangaID int64) {
	t.Helper()

	now := time.Now().UnixMilli()

	_, err := db.Exec(`
		INSERT INTO history (manga_id, created_at, updated_at, chapter_id, page, scroll, percent, chapters, deleted_at, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		mangaID, now, now, 12345, 1, 0.0, 0.1, 10, 0, userID,
	)
	if err != nil {
		t.Fatalf("failed to create test history: %v", err)
	}
}

// CreateTestTag creates a tag entry.
func CreateTestTag(t *testing.T, db *sql.DB, id int64, title, key string) int64 {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO tags (id, title, ` + "`key`" + `, source)
		VALUES (?, ?, ?, ?)`,
		id, title, key, "test_source",
	)
	if err != nil {
		t.Fatalf("failed to create test tag: %v", err)
	}

	return id
}

// LinkMangaTag associates a tag with a manga.
func LinkMangaTag(t *testing.T, db *sql.DB, mangaID, tagID int64) {
	t.Helper()

	_, err := db.Exec(
		"INSERT INTO manga_tags (manga_id, tag_id) VALUES (?, ?)",
		mangaID, tagID,
	)
	if err != nil {
		t.Fatalf("failed to link manga tag: %v", err)
	}
}

// CleanupTestData removes all test data from the database.
// Call this in test cleanup or teardown.
func CleanupTestData(t *testing.T) {
	t.Helper()

	db := GetTestDB(t)
	if db == nil {
		return
	}

	CleanupTestDataWithDB(t, db)
}

// CleanupTestDataWithDB removes all test data using the provided connection.
func CleanupTestDataWithDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Order matters due to foreign key constraints
	tables := []string{
		"history",
		"favourites",
		"categories",
		"manga_tags",
		"tags",
		"manga",
		"users",
	}

	TruncateTablesWithDB(t, db, tables...)
}

// SeedTestData creates a complete set of test data for integration tests.
// Returns the created user ID and manga IDs.
func SeedTestData(t *testing.T, db *sql.DB) (userID int64, mangaIDs []int64) {
	t.Helper()

	// Create user
	userID = CreateTestUserWithCredentials(t, db, TestEmail, TestPassword, TestNickname)

	// Create manga
	mangaIDs = CreateTestMangaWithDB(t, db, 3)

	// Create category
	categoryID := CreateTestCategory(t, db, userID, "Test Category")

	// Add manga to favourites
	CreateTestFavourite(t, db, userID, mangaIDs[0], categoryID)

	// Add history
	CreateTestHistory(t, db, userID, mangaIDs[0])

	return userID, mangaIDs
}
