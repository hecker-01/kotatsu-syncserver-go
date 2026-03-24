package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

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
