package utils

import "context"

type contextKey string

// UserIDContextKey is the context key for storing authenticated user IDs.
const UserIDContextKey contextKey = "user_id"

// WithUserID returns a new context with the given user ID attached.
func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDContextKey, userID)
}

// UserIDFromContext extracts the user ID from context.
// Returns (userID, true) if present, (0, false) otherwise.
func UserIDFromContext(ctx context.Context) (int, bool) {
	v := ctx.Value(UserIDContextKey)
	if v == nil {
		return 0, false
	}

	switch userID := v.(type) {
	case int:
		return userID, true
	case float64:
		return int(userID), true
	default:
		return 0, false
	}
}
