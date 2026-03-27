package models

// AuthRequest for POST /auth
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ForgotPasswordRequest for POST /forgot-password
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest for POST /reset-password
type ResetPasswordRequest struct {
	ResetToken string `json:"resetToken"`
	Password   string `json:"password"`
}

// FavouritesPackage for /resource/favourites
type FavouritesPackage struct {
	Categories []Category  `json:"categories"`
	Favourites []Favourite `json:"favourites"`
	Timestamp  *int64      `json:"timestamp,omitempty"`
}

// HistoryPackage for /resource/history
type HistoryPackage struct {
	History   []History `json:"history"`
	Timestamp *int64    `json:"timestamp,omitempty"`
}
