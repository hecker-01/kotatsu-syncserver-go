// Package middleware provides HTTP middleware for authentication, logging,
// and rate limiting used by the API server.
package middleware

import (
	"net/http"
	"strings"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// RequireAuth is middleware that validates JWT bearer tokens from the
// Authorization header and injects the authenticated user_id into request context.
// Returns 401 if the token is missing, malformed, or invalid.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := utils.ParseAndValidateJWT(tokenString)
		if err != nil {
			logger.Warn("invalid token", "ip", r.RemoteAddr)
			utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx := utils.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
