package utils

import "context"

type contextKey string

// UserIDContextKey is the context key for storing authenticated user IDs.
const UserIDContextKey contextKey = "user_id"

// WithUserID returns a new context with the given user ID attached.
// Uses int64 to support BIGINT user IDs from the database.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserIDContextKey, userID)
}

// UserIDFromContext extracts the user ID from context.
// Returns (userID, true) if present, (0, false) otherwise.
// Uses int64 to support BIGINT user IDs from the database.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(UserIDContextKey)
	if v == nil {
		return 0, false
	}

	switch userID := v.(type) {
	case int64:
		return userID, true
	case int:
		return int64(userID), true
	case float64:
		return int64(userID), true
	default:
		return 0, false
	}
}
