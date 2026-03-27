package controllers_test

import (
	"net/http"
	"testing"

	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/testutil"
)

func TestMeEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testutil.RequireTestDB(t)
	defer testutil.CleanupTestData(t)

	router := setupTestRouter(t)
	ts := testutil.NewTestServer(t, router)
	defer ts.Close()

	// Create a test user
	db := testutil.GetTestDB(t)
	userID := testutil.CreateTestUserWithCredentials(t, db, "me@example.com", "password123", "TestUser")

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
		checkUser      bool
	}{
		{
			name:           "authenticated request",
			token:          testutil.GenerateTestToken(t, userID),
			expectedStatus: http.StatusOK,
			checkUser:      true,
		},
		{
			name:           "no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "invalid token",
			token:          testutil.InvalidToken(),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			token:          testutil.ExpiredToken(t, userID),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong signature",
			token:          testutil.WrongSignatureToken(t, userID),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := make(map[string]string)
			if tt.token != "" {
				headers["Authorization"] = "Bearer " + tt.token
			}

			resp := ts.Get(t, "/me", headers)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)

			if tt.expectedBody != "" {
				testutil.AssertBodyContains(t, resp, tt.expectedBody)
			}

			if tt.checkUser {
				var user models.UserResponse
				testutil.ParseJSON(t, resp, &user)
				if user.Email != "me@example.com" {
					t.Errorf("Expected email 'me@example.com', got '%s'", user.Email)
				}
				if user.ID != userID {
					t.Errorf("Expected user ID %d, got %d", userID, user.ID)
				}
			}
		})
	}
}
