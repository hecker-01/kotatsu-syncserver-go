package utils

import "context"

type contextKey string

const UserIDContextKey contextKey = "user_id"

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDContextKey, userID)
}

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
