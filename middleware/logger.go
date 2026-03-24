package middleware

import (
	"net/http"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/logger"
)

// StructuredLogger is middleware that logs HTTP requests with method, path,
// status code, duration, and client IP. Log level adjusts based on status:
// 5xx → error, 4xx → warn, others → info.
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		// Determine log level based on status code
		level := logger.LevelInfo
		if ww.status >= 500 {
			level = logger.LevelError
		} else if ww.status >= 400 {
			level = logger.LevelWarn
		}

		logger.AccessLog(level, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
