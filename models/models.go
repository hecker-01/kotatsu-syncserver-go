package models

// User represents a user account
type User struct {
	ID                          int64   `json:"id"`
	Email                       string  `json:"email"`
	PasswordHash                string  `json:"-"` // Never expose
	Nickname                    *string `json:"nickname,omitempty"`
	FavouritesSyncTimestamp     *int64  `json:"-"` // Internal
	HistorySyncTimestamp        *int64  `json:"-"` // Internal
	PasswordResetTokenHash      *string `json:"-"` // Internal
	PasswordResetTokenExpiresAt *int64  `json:"-"` // Internal
}

// UserResponse is the JSON response for /me endpoint
type UserResponse struct {
	ID       int64   `json:"id"`
	Email    string  `json:"email"`
	Nickname *string `json:"nickname,omitempty"`
}

// Manga represents manga metadata
type Manga struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	AltTitle      *string `json:"altTitle,omitempty"`
	URL           string  `json:"url"`
	PublicURL     string  `json:"publicUrl"`
	Rating        float32 `json:"rating"`
	ContentRating *string `json:"contentRating,omitempty"` // SAFE, SUGGESTIVE, ADULT
	CoverURL      string  `json:"coverUrl"`
	LargeCoverURL *string `json:"largeCoverUrl,omitempty"`
	State         *string `json:"state,omitempty"` // ONGOING, FINISHED, etc.
	Author        *string `json:"author,omitempty"`
	Source        string  `json:"source"`
}

// Tag represents a manga tag/genre
type Tag struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Key    string `json:"key"`
	Source string `json:"source"`
}

// Category represents a user-created collection
type Category struct {
	ID        int64  `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	SortKey   int    `json:"sortKey"`
	Title     string `json:"title"`
	Order     string `json:"order"`
	Track     bool   `json:"track"`
	ShowInLib bool   `json:"showInLib"`
	DeletedAt *int64 `json:"deletedAt,omitempty"`
}

// Favourite represents a manga in a user's category
type Favourite struct {
	MangaID    int64  `json:"mangaId"`
	CategoryID int64  `json:"categoryId"`
	SortKey    int    `json:"sortKey"`
	Pinned     bool   `json:"pinned"`
	CreatedAt  int64  `json:"createdAt"`
	DeletedAt  *int64 `json:"deletedAt,omitempty"`
}

// History represents reading progress
type History struct {
	MangaID   int64   `json:"mangaId"`
	CreatedAt int64   `json:"createdAt"`
	UpdatedAt int64   `json:"updatedAt"`
	ChapterID int64   `json:"chapterId"`
	Page      int16   `json:"page"`
	Scroll    float64 `json:"scroll"`
	Percent   float64 `json:"percent"`
	Chapters  int     `json:"chapters"`
	DeletedAt *int64  `json:"deletedAt,omitempty"`
}
