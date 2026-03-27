package services

import (
	"database/sql"
	"errors"

	"github.com/hecker-01/kotatsu-syncserver-go/db"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
)

// MangaService handles manga-related business logic.
type MangaService struct{}

// NewMangaService creates a new manga service instance.
func NewMangaService() *MangaService {
	return &MangaService{}
}

// ListManga retrieves a paginated list of manga.
// Returns manga ordered by ID with the specified offset and limit.
func (s *MangaService) ListManga(offset, limit int) ([]models.Manga, error) {
	rows, err := db.DB.Query(
		`SELECT id, title, alt_title, url, public_url, rating, content_rating, 
		 cover_url, large_cover_url, state, author, source 
		 FROM manga ORDER BY id LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mangaList []models.Manga
	for rows.Next() {
		var m models.Manga
		err := rows.Scan(
			&m.ID, &m.Title, &m.AltTitle, &m.URL, &m.PublicURL,
			&m.Rating, &m.ContentRating, &m.CoverURL, &m.LargeCoverURL,
			&m.State, &m.Author, &m.Source,
		)
		if err != nil {
			return nil, err
		}
		mangaList = append(mangaList, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return empty array instead of nil for consistent JSON serialization
	if mangaList == nil {
		mangaList = []models.Manga{}
	}

	return mangaList, nil
}

// GetManga retrieves a single manga by ID.
// Returns nil if the manga is not found.
func (s *MangaService) GetManga(id int64) (*models.Manga, error) {
	var m models.Manga
	err := db.DB.QueryRow(
		`SELECT id, title, alt_title, url, public_url, rating, content_rating, 
		 cover_url, large_cover_url, state, author, source 
		 FROM manga WHERE id = ?`,
		id,
	).Scan(
		&m.ID, &m.Title, &m.AltTitle, &m.URL, &m.PublicURL,
		&m.Rating, &m.ContentRating, &m.CoverURL, &m.LargeCoverURL,
		&m.State, &m.Author, &m.Source,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &m, nil
}
