package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/testutil"
	"github.com/hecker-01/kotatsu-syncserver-go/utils"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		window         time.Duration
		requests       int
		expectedStatus []int
	}{
		{
			name:           "under limit",
			limit:          3,
			window:         1 * time.Second,
			requests:       2,
			expectedStatus: []int{http.StatusOK, http.StatusOK},
		},
		{
			name:           "at limit",
			limit:          3,
			window:         1 * time.Second,
			requests:       4,
			expectedStatus: []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusTooManyRequests},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.limit, tt.window)

			handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

				if w.Code != tt.expectedStatus[i] {
					t.Errorf("Request %d: expected status %d, got %d", i+1, tt.expectedStatus[i], w.Code)
				}

				if w.Code == http.StatusTooManyRequests {
					if w.Header().Get("Retry-After") == "" {
						t.Error("Expected Retry-After header in 429 response")
					}
				}
			}
		})
	}
}

func TestRequireAuth(t *testing.T) {
	testutil.SetupTestJWTEnv(t)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + testutil.GenerateTestToken(t, 123),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "invalid format - no Bearer",
			authHeader:     testutil.GenerateTestToken(t, 123),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer " + testutil.InvalidToken(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + testutil.ExpiredToken(t, 123),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "wrong signature",
			authHeader:     "Bearer " + testutil.WrongSignatureToken(t, 123),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := utils.UserIDFromContext(r.Context())
				if !ok || userID == 0 {
					t.Error("Expected user ID in context, got 0")
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				// Middleware returns JSON error responses
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, body)
				}
			}
		})
	}
}
