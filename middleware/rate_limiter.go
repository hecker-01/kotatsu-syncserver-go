package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

// Rate limit configuration constants matching Kotatsu API groups.
const (
	// GlobalAPILimit is the standard rate limit for general API endpoints.
	GlobalAPILimit = 100
	// GlobalAPIWindow is the time window for the global API rate limit.
	GlobalAPIWindow = 5 * time.Minute

	// AuthLimit is the rate limit for authentication endpoints.
	AuthLimit = 5
	// AuthWindow is the time window for the auth rate limit.
	AuthWindow = 5 * time.Minute

	// ForgotPasswordLimit is the rate limit for forgot password endpoint.
	ForgotPasswordLimit = 3
	// ForgotPasswordWindow is the time window for forgot password rate limit.
	ForgotPasswordWindow = 15 * time.Minute

	// ResetPasswordLimit is the rate limit for reset password endpoint.
	ResetPasswordLimit = 3
	// ResetPasswordWindow is the time window for reset password rate limit.
	ResetPasswordWindow = 15 * time.Minute
)

// rateLimitEntry tracks request count and reset time for an IP address.
type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

// GlobalAPILimiter returns middleware for standard API rate limiting (100 req/5min).
func GlobalAPILimiter() func(http.Handler) http.Handler {
	return NewRateLimiter(GlobalAPILimit, GlobalAPIWindow)
}

// AuthLimiter returns middleware for authentication rate limiting (5 req/5min).
func AuthLimiter() func(http.Handler) http.Handler {
	return NewRateLimiter(AuthLimit, AuthWindow)
}

// ForgotPasswordLimiter returns middleware for forgot password rate limiting (3 req/15min).
func ForgotPasswordLimiter() func(http.Handler) http.Handler {
	return NewRateLimiter(ForgotPasswordLimit, ForgotPasswordWindow)
}

// ResetPasswordLimiter returns middleware for reset password rate limiting (3 req/15min).
func ResetPasswordLimiter() func(http.Handler) http.Handler {
	return NewRateLimiter(ResetPasswordLimit, ResetPasswordWindow)
}

// NewRateLimiter creates middleware that limits requests per IP address.
// Returns 429 with Retry-After header when limit is exceeded.
// maxRequests sets the maximum number of requests allowed within the time window.
func NewRateLimiter(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		entries = make(map[string]rateLimitEntry)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			now := time.Now()

			mu.Lock()
			entry, ok := entries[ip]
			if !ok || now.After(entry.resetAt) {
				entry = rateLimitEntry{count: 0, resetAt: now.Add(window)}
			}

			if entry.count >= maxRequests {
				retryAfter := int(time.Until(entry.resetAt).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				mu.Unlock()

				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				utils.WriteError(w, http.StatusTooManyRequests, "Too many requests from this IP, please try again later.")
				return
			}

			entry.count++
			entries[ip] = entry
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the client IP address from the request.
// Checks X-Forwarded-For header first, falls back to RemoteAddr.
func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
