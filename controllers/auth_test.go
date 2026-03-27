package controllers_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/routes"
	"github.com/hecker-01/kotatsu-syncserver-go/testutil"
)

func setupTestRouter(t *testing.T) chi.Router {
	t.Helper()
	testutil.SetupTestJWTEnv(t)
	r := chi.NewRouter()
	routes.RegisterRoutes(r)
	return r
}

func TestAuthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testutil.RequireTestDB(t)
	defer testutil.CleanupTestData(t)

	router := setupTestRouter(t)
	ts := testutil.NewTestServer(t, router)
	defer ts.Close()

	tests := []struct {
		name           string
		email          string
		password       string
		allowRegister  string
		setupUser      bool
		expectedStatus int
		expectedBody   string
		checkToken     bool
	}{
		{
			name:           "successful login",
			email:          "existing@example.com",
			password:       "password123",
			setupUser:      true,
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name:           "successful register",
			email:          "newuser@example.com",
			password:       "password123",
			allowRegister:  "true",
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name:           "registration disabled",
			email:          "newuser2@example.com",
			password:       "password123",
			allowRegister:  "false",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Wrong password",
		},
		{
			name:           "wrong password",
			email:          "existing@example.com",
			password:       "wrongpassword",
			setupUser:      true,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Wrong password",
		},
		{
			name:           "missing email",
			email:          "",
			password:       "password123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Wrong password",
		},
		{
			name:           "missing password",
			email:          "test@example.com",
			password:       "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Wrong password",
		},
		{
			name:           "password too short",
			email:          "test@example.com",
			password:       "a",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Wrong password",
		},
		{
			name:           "password too long",
			email:          "test@example.com",
			password:       "12345678901234567890123456", // 26 chars
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Wrong password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.allowRegister != "" {
				os.Setenv("ALLOW_NEW_REGISTER", tt.allowRegister)
				defer os.Unsetenv("ALLOW_NEW_REGISTER")
			}

			if tt.setupUser {
				db := testutil.GetTestDB(t)
				testutil.CreateTestUserWithCredentials(t, db, tt.email, tt.password, "TestUser")
			}

			// Execute
			reqBody := models.AuthRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			resp := ts.Post(t, "/auth", reqBody, nil)
			defer resp.Body.Close()

			// Assert
			testutil.AssertStatus(t, resp, tt.expectedStatus)

			if tt.expectedBody != "" {
				testutil.AssertBodyContains(t, resp, tt.expectedBody)
			}

			if tt.checkToken {
				var authResp models.AuthResponse
				testutil.ParseJSON(t, resp, &authResp)
				if authResp.Token == "" {
					t.Error("Expected token in response, got empty string")
				}
			}
		})
	}
}
