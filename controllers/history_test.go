package controllers_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/testutil"
)

func TestGetHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testutil.RequireTestDB(t)
	defer testutil.CleanupTestData(t)

	router := setupTestRouter(t)
	ts := testutil.NewTestServer(t, router)
	defer ts.Close()

	// Create test user and get token
	db := testutil.GetTestDB(t)
	userID := testutil.CreateTestUserWithCredentials(t, db, "hist@example.com", "password123", "TestUser")
	token := testutil.GenerateTestToken(t, userID)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "authenticated empty history",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized without token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			token:          testutil.InvalidToken(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := make(map[string]string)
			if tt.token != "" {
				headers["Authorization"] = "Bearer " + tt.token
			}

			resp := ts.Get(t, "/resource/history", headers)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)

			if tt.expectedStatus == http.StatusOK {
				var pkg models.HistoryPackage
				testutil.ParseJSON(t, resp, &pkg)
				if pkg.History == nil {
					t.Error("Expected history array, got nil")
				}
			}
		})
	}
}

func TestSyncHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testutil.RequireTestDB(t)
	defer testutil.CleanupTestData(t)

	router := setupTestRouter(t)
	ts := testutil.NewTestServer(t, router)
	defer ts.Close()

	// Create test user
	db := testutil.GetTestDB(t)
	userID := testutil.CreateTestUserWithCredentials(t, db, "histsync@example.com", "password123", "TestUser")
	token := testutil.GenerateTestToken(t, userID)

	now := time.Now().UnixMilli()

	tests := []struct {
		name           string
		token          string
		pkg            models.HistoryPackage
		expectedStatus int
	}{
		{
			name:  "empty sync returns 204",
			token: token,
			pkg: models.HistoryPackage{
				History:   []models.History{},
				Timestamp: &now,
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:  "unauthorized without token",
			token: "",
			pkg: models.HistoryPackage{
				History: []models.History{},
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := make(map[string]string)
			if tt.token != "" {
				headers["Authorization"] = "Bearer " + tt.token
			}

			resp := ts.Post(t, "/resource/history", tt.pkg, headers)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)
		})
	}
}
