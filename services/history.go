// Package services implements business logic. Services return typed errors
// for controllers to translate into appropriate HTTP responses.
package services

import (
	"database/sql"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
)

// HistoryService handles history-related business logic.
type HistoryService struct{}

// NewHistoryService creates a new history service instance.
func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

// GetHistory retrieves all history records for a user along with the current sync timestamp.
func (s *HistoryService) GetHistory(userID int64) (*models.HistoryPackage, error) {
	// Get user's current history_sync_timestamp
	var syncTimestamp sql.NullInt64
	err := db.DB.QueryRow(
		"SELECT history_sync_timestamp FROM users WHERE id = ?",
		userID,
	).Scan(&syncTimestamp)
	if err != nil {
		return nil, err
	}

	// Query all history for this user
	rows, err := db.DB.Query(`
		SELECT manga_id, created_at, updated_at, chapter_id, page, scroll, percent, chapters, deleted_at
		FROM history
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]models.History, 0)
	for rows.Next() {
		var h models.History
		var deletedAt int64
		err := rows.Scan(
			&h.MangaID, &h.CreatedAt, &h.UpdatedAt, &h.ChapterID,
			&h.Page, &h.Scroll, &h.Percent, &h.Chapters, &deletedAt,
		)
		if err != nil {
			return nil, err
		}
		// Convert deletedAt: 0 means not deleted (nil), non-zero is the timestamp
		if deletedAt != 0 {
			h.DeletedAt = &deletedAt
		}
		history = append(history, h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	pkg := &models.HistoryPackage{
		History: history,
	}
	if syncTimestamp.Valid {
		pkg.Timestamp = &syncTimestamp.Int64
	}

	return pkg, nil
}

// SyncHistory synchronizes history records from the client.
// Returns the merged package with server-newer records, hasChanges bool, and error.
// hasChanges=false means client is up-to-date (return 204).
func (s *HistoryService) SyncHistory(userID int64, clientPkg *models.HistoryPackage) (*models.HistoryPackage, bool, error) {
	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	// Ensure all manga IDs exist before processing history records
	// to satisfy foreign key constraint
	if err := s.ensureMangaExist(tx, clientPkg.History); err != nil {
		return nil, false, err
	}

	// Get user's current history_sync_timestamp
	var serverSyncTimestamp sql.NullInt64
	err = tx.QueryRow(
		"SELECT history_sync_timestamp FROM users WHERE id = ?",
		userID,
	).Scan(&serverSyncTimestamp)
	if err != nil {
		return nil, false, err
	}

	// Get client timestamp (default to 0 if not provided)
	clientTimestamp := int64(0)
	if clientPkg.Timestamp != nil {
		clientTimestamp = *clientPkg.Timestamp
	}

	// Build a map of server history indexed by manga_id
	serverHistory := make(map[int64]models.History)
	rows, err := tx.Query(`
		SELECT manga_id, created_at, updated_at, chapter_id, page, scroll, percent, chapters, deleted_at
		FROM history
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var h models.History
		var deletedAt int64
		err := rows.Scan(
			&h.MangaID, &h.CreatedAt, &h.UpdatedAt, &h.ChapterID,
			&h.Page, &h.Scroll, &h.Percent, &h.Chapters, &deletedAt,
		)
		if err != nil {
			return nil, false, err
		}
		if deletedAt != 0 {
			h.DeletedAt = &deletedAt
		}
		serverHistory[h.MangaID] = h
	}
	if err = rows.Err(); err != nil {
		return nil, false, err
	}

	// Track which server records are newer than client (to return in response)
	newerOnServer := make([]models.History, 0)

	// Process each client history record
	for _, clientRecord := range clientPkg.History {
		serverRecord, exists := serverHistory[clientRecord.MangaID]

		if !exists {
			// New record from client - insert it (use ON DUPLICATE KEY UPDATE for idempotency)
			deletedAt := int64(0)
			if clientRecord.DeletedAt != nil {
				deletedAt = *clientRecord.DeletedAt
			}
			_, err = tx.Exec(`
				INSERT INTO history (user_id, manga_id, created_at, updated_at, chapter_id, page, scroll, percent, chapters, deleted_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					created_at = VALUES(created_at),
					updated_at = VALUES(updated_at),
					chapter_id = VALUES(chapter_id),
					page = VALUES(page),
					scroll = VALUES(scroll),
					percent = VALUES(percent),
					chapters = VALUES(chapters),
					deleted_at = VALUES(deleted_at)
			`, userID, clientRecord.MangaID, clientRecord.CreatedAt, clientRecord.UpdatedAt,
				clientRecord.ChapterID, clientRecord.Page, clientRecord.Scroll, clientRecord.Percent,
				clientRecord.Chapters, deletedAt)
			if err != nil {
				return nil, false, err
			}
		} else if clientRecord.UpdatedAt > serverRecord.UpdatedAt {
			// Client record is newer - update server
			deletedAt := int64(0)
			if clientRecord.DeletedAt != nil {
				deletedAt = *clientRecord.DeletedAt
			}
			_, err = tx.Exec(`
				UPDATE history
				SET created_at = ?, updated_at = ?, chapter_id = ?, page = ?, scroll = ?, percent = ?, chapters = ?, deleted_at = ?
				WHERE user_id = ? AND manga_id = ?
			`, clientRecord.CreatedAt, clientRecord.UpdatedAt, clientRecord.ChapterID,
				clientRecord.Page, clientRecord.Scroll, clientRecord.Percent, clientRecord.Chapters,
				deletedAt, userID, clientRecord.MangaID)
			if err != nil {
				return nil, false, err
			}
		} else if serverRecord.UpdatedAt > clientRecord.UpdatedAt {
			// Server record is newer - add to response
			newerOnServer = append(newerOnServer, serverRecord)
		}
		// If timestamps are equal, no action needed

		// Remove processed record from map to track unprocessed server records
		delete(serverHistory, clientRecord.MangaID)
	}

	// Check for server records that client doesn't have yet
	// These are records updated after client's last sync timestamp
	for _, serverRecord := range serverHistory {
		if serverRecord.UpdatedAt > clientTimestamp {
			newerOnServer = append(newerOnServer, serverRecord)
		}
	}

	// Update the user's history_sync_timestamp
	newTimestamp := time.Now().UnixMilli()
	_, err = tx.Exec(
		"UPDATE users SET history_sync_timestamp = ? WHERE id = ?",
		newTimestamp, userID,
	)
	if err != nil {
		return nil, false, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, false, err
	}

	// Determine if there are changes to return
	hasChanges := len(newerOnServer) > 0

	if !hasChanges {
		return nil, false, nil
	}

	return &models.HistoryPackage{
		History:   newerOnServer,
		Timestamp: &newTimestamp,
	}, true, nil
}

// ensureMangaExist ensures all manga IDs in history records exist in the manga table
// by creating placeholder records if necessary (to satisfy foreign key constraint).
func (s *HistoryService) ensureMangaExist(tx *sql.Tx, history []models.History) error {
	if len(history) == 0 {
		return nil
	}

	// Collect unique manga IDs from history
	mangaIDsToCheck := make(map[int64]bool)
	for _, h := range history {
		if !mangaIDsToCheck[h.MangaID] {
			mangaIDsToCheck[h.MangaID] = true
		}
	}

	// Check which manga IDs are missing and create placeholders if needed
	for mangaID := range mangaIDsToCheck {
		var exists int
		err := tx.QueryRow("SELECT 1 FROM manga WHERE id = ?", mangaID).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				// Manga doesn't exist, create a placeholder record
				// Use INSERT IGNORE to handle race conditions if multiple requests try to insert the same ID
				_, insertErr := tx.Exec(`
					INSERT IGNORE INTO manga (id, title, url, public_url, rating, cover_url, source)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`, mangaID, "Unknown Manga", "http://unknown", "http://unknown", 0.0, "http://unknown", "unknown")
				if insertErr != nil {
					return insertErr
				}
			} else {
				return err
			}
		}
	}

	return nil
}
