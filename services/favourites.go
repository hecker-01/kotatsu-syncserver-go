// Package services implements business logic for favourites sync operations.
package services

import (
	"database/sql"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
)

// FavouritesService handles favourites-related business logic.
type FavouritesService struct{}

// NewFavouritesService creates a new favourites service instance.
func NewFavouritesService() *FavouritesService {
	return &FavouritesService{}
}

// GetFavourites retrieves all categories and favourites for a user.
// Returns the current server data along with the user's sync timestamp.
func (s *FavouritesService) GetFavourites(userID int64) (*models.FavouritesPackage, error) {
	// Get user's current sync timestamp
	var syncTimestamp sql.NullInt64
	err := db.DB.QueryRow(
		"SELECT favourites_sync_timestamp FROM users WHERE id = ?",
		userID,
	).Scan(&syncTimestamp)
	if err != nil {
		return nil, err
	}

	// Fetch all categories for the user
	categories, err := s.fetchCategories(userID)
	if err != nil {
		return nil, err
	}

	// Fetch all favourites for the user
	favourites, err := s.fetchFavourites(userID)
	if err != nil {
		return nil, err
	}

	pkg := &models.FavouritesPackage{
		Categories: categories,
		Favourites: favourites,
	}

	if syncTimestamp.Valid {
		pkg.Timestamp = &syncTimestamp.Int64
	}

	return pkg, nil
}

// SyncFavourites synchronizes the client's favourites with the server.
// Returns the merged package, a boolean indicating if changes were made (true = return 200),
// and any error that occurred.
// If hasChanges is false, the client should receive 204 No Content.
// If hasChanges is true, the client should receive 200 with the merged data.
func (s *FavouritesService) SyncFavourites(userID int64, pkg *models.FavouritesPackage) (*models.FavouritesPackage, bool, error) {
	// Get user's current sync timestamp
	var serverTimestamp sql.NullInt64
	err := db.DB.QueryRow(
		"SELECT favourites_sync_timestamp FROM users WHERE id = ?",
		userID,
	).Scan(&serverTimestamp)
	if err != nil {
		return nil, false, err
	}

	clientTimestamp := int64(0)
	if pkg.Timestamp != nil {
		clientTimestamp = *pkg.Timestamp
	}

	serverTs := int64(0)
	if serverTimestamp.Valid {
		serverTs = serverTimestamp.Int64
	}

	// Begin transaction for atomic operations
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	// Client timestamp >= server timestamp: accept client data, update server, return 204
	if clientTimestamp >= serverTs {
		// Upsert categories from client
		if err := s.upsertCategories(tx, userID, pkg.Categories); err != nil {
			return nil, false, err
		}

		// Upsert favourites from client
		if err := s.upsertFavourites(tx, userID, pkg.Favourites); err != nil {
			return nil, false, err
		}

		// Update user's sync timestamp to current time
		newTimestamp := maxTimestamp(pkg.Categories, pkg.Favourites)
		if newTimestamp == 0 {
			newTimestamp = clientTimestamp
		}

		_, err = tx.Exec(
			"UPDATE users SET favourites_sync_timestamp = ? WHERE id = ?",
			newTimestamp, userID,
		)
		if err != nil {
			return nil, false, err
		}

		if err := tx.Commit(); err != nil {
			return nil, false, err
		}

		// No changes to return (client is up-to-date)
		return nil, false, nil
	}

	// Client timestamp < server timestamp: merge and return server state
	// First, apply client changes
	if err := s.upsertCategories(tx, userID, pkg.Categories); err != nil {
		return nil, false, err
	}
	if err := s.upsertFavourites(tx, userID, pkg.Favourites); err != nil {
		return nil, false, err
	}

	// Fetch merged data from server
	categories, err := s.fetchCategoriesTx(tx, userID)
	if err != nil {
		return nil, false, err
	}

	favourites, err := s.fetchFavouritesTx(tx, userID)
	if err != nil {
		return nil, false, err
	}

	// Calculate new timestamp as max of all items
	newTimestamp := maxTimestamp(categories, favourites)
	if newTimestamp == 0 {
		newTimestamp = serverTs
	}

	// Update user's sync timestamp
	_, err = tx.Exec(
		"UPDATE users SET favourites_sync_timestamp = ? WHERE id = ?",
		newTimestamp, userID,
	)
	if err != nil {
		return nil, false, err
	}

	if err := tx.Commit(); err != nil {
		return nil, false, err
	}

	result := &models.FavouritesPackage{
		Categories: categories,
		Favourites: favourites,
		Timestamp:  &newTimestamp,
	}

	return result, true, nil
}

// fetchCategories retrieves all categories for a user.
func (s *FavouritesService) fetchCategories(userID int64) ([]models.Category, error) {
	rows, err := db.DB.Query(
		"SELECT id, created_at, sort_key, title, `order`, track, show_in_lib, deleted_at FROM categories WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCategories(rows)
}

// fetchCategoriesTx retrieves all categories for a user within a transaction.
func (s *FavouritesService) fetchCategoriesTx(tx *sql.Tx, userID int64) ([]models.Category, error) {
	rows, err := tx.Query(
		"SELECT id, created_at, sort_key, title, `order`, track, show_in_lib, deleted_at FROM categories WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCategories(rows)
}

// scanCategories scans rows into Category structs.
func scanCategories(rows *sql.Rows) ([]models.Category, error) {
	var categories []models.Category
	for rows.Next() {
		var c models.Category
		var deletedAt sql.NullInt64
		if err := rows.Scan(&c.ID, &c.CreatedAt, &c.SortKey, &c.Title, &c.Order, &c.Track, &c.ShowInLib, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Int64
		}
		categories = append(categories, c)
	}
	if categories == nil {
		categories = []models.Category{}
	}
	return categories, rows.Err()
}

// fetchFavourites retrieves all favourites for a user.
func (s *FavouritesService) fetchFavourites(userID int64) ([]models.Favourite, error) {
	rows, err := db.DB.Query(
		"SELECT manga_id, category_id, sort_key, pinned, created_at, deleted_at FROM favourites WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFavourites(rows)
}

// fetchFavouritesTx retrieves all favourites for a user within a transaction.
func (s *FavouritesService) fetchFavouritesTx(tx *sql.Tx, userID int64) ([]models.Favourite, error) {
	rows, err := tx.Query(
		"SELECT manga_id, category_id, sort_key, pinned, created_at, deleted_at FROM favourites WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFavourites(rows)
}

// scanFavourites scans rows into Favourite structs.
func scanFavourites(rows *sql.Rows) ([]models.Favourite, error) {
	var favourites []models.Favourite
	for rows.Next() {
		var f models.Favourite
		var deletedAt int64
		if err := rows.Scan(&f.MangaID, &f.CategoryID, &f.SortKey, &f.Pinned, &f.CreatedAt, &deletedAt); err != nil {
			return nil, err
		}
		// deleted_at in favourites is NOT NULL in schema; 0 = deleted, non-zero = active
		if deletedAt != 0 {
			f.DeletedAt = &deletedAt
		}
		favourites = append(favourites, f)
	}
	if favourites == nil {
		favourites = []models.Favourite{}
	}
	return favourites, rows.Err()
}

// upsertCategories inserts or updates categories within a transaction.
func (s *FavouritesService) upsertCategories(tx *sql.Tx, userID int64, categories []models.Category) error {
	if len(categories) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO categories (id, created_at, sort_key, title, ` + "`order`" + `, user_id, track, show_in_lib, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			sort_key = VALUES(sort_key),
			title = VALUES(title),
			` + "`order`" + ` = VALUES(` + "`order`" + `),
			track = VALUES(track),
			show_in_lib = VALUES(show_in_lib),
			deleted_at = VALUES(deleted_at)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range categories {
		var deletedAt sql.NullInt64
		if c.DeletedAt != nil {
			deletedAt = sql.NullInt64{Int64: *c.DeletedAt, Valid: true}
		}
		_, err := stmt.Exec(c.ID, c.CreatedAt, c.SortKey, c.Title, c.Order, userID, c.Track, c.ShowInLib, deletedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

// upsertFavourites inserts or updates favourites within a transaction.
func (s *FavouritesService) upsertFavourites(tx *sql.Tx, userID int64, favourites []models.Favourite) error {
	if len(favourites) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO favourites (manga_id, category_id, sort_key, pinned, created_at, deleted_at, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			sort_key = VALUES(sort_key),
			pinned = VALUES(pinned),
			created_at = VALUES(created_at),
			deleted_at = VALUES(deleted_at)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range favourites {
		// deleted_at is NOT NULL in schema; use 0 for active items
		deletedAt := int64(0)
		if f.DeletedAt != nil {
			deletedAt = *f.DeletedAt
		}
		_, err := stmt.Exec(f.MangaID, f.CategoryID, f.SortKey, f.Pinned, f.CreatedAt, deletedAt, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

// maxTimestamp finds the maximum timestamp across categories and favourites.
func maxTimestamp(categories []models.Category, favourites []models.Favourite) int64 {
	var max int64

	for _, c := range categories {
		if c.CreatedAt > max {
			max = c.CreatedAt
		}
		if c.DeletedAt != nil && *c.DeletedAt > max {
			max = *c.DeletedAt
		}
	}

	for _, f := range favourites {
		if f.CreatedAt > max {
			max = f.CreatedAt
		}
		if f.DeletedAt != nil && *f.DeletedAt > max {
			max = *f.DeletedAt
		}
	}

	return max
}
