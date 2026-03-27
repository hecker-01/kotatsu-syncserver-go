package controllers_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/testutil"
)

func TestGetFavourites(t *testing.T) {
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
	userID := testutil.CreateTestUserWithCredentials(t, db, "fav@example.com", "password123", "TestUser")
	token := testutil.GenerateTestToken(t, userID)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "authenticated empty favourites",
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

			resp := ts.Get(t, "/resource/favourites", headers)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)

			if tt.expectedStatus == http.StatusOK {
				var pkg models.FavouritesPackage
				testutil.ParseJSON(t, resp, &pkg)
				if pkg.Categories == nil {
					t.Error("Expected categories array, got nil")
				}
				if pkg.Favourites == nil {
					t.Error("Expected favourites array, got nil")
				}
			}
		})
	}
}

func TestSyncFavourites(t *testing.T) {
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
	userID := testutil.CreateTestUserWithCredentials(t, db, "sync@example.com", "password123", "TestUser")
	token := testutil.GenerateTestToken(t, userID)

	now := time.Now().UnixMilli()

	tests := []struct {
		name           string
		token          string
		pkg            models.FavouritesPackage
		expectedStatus int
	}{
		{
			name:  "empty sync returns 204",
			token: token,
			pkg: models.FavouritesPackage{
				Categories: []models.Category{},
				Favourites: []models.Favourite{},
				Timestamp:  &now,
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:  "unauthorized without token",
			token: "",
			pkg: models.FavouritesPackage{
				Categories: []models.Category{},
				Favourites: []models.Favourite{},
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

			resp := ts.Post(t, "/resource/favourites", tt.pkg, headers)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)
		})
	}
}
