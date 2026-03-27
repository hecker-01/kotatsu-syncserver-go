package controllers_test

import (
	"net/http"
	"testing"

	"github.com/hecker-01/kotatsu-syncserver-go/testutil"
)

func TestListManga(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testutil.RequireTestDB(t)
	defer testutil.CleanupTestData(t)

	router := setupTestRouter(t)
	ts := testutil.NewTestServer(t, router)
	defer ts.Close()

	// Create some test manga
	testutil.CreateTestManga(t, 5)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid pagination",
			path:           "/manga?offset=0&limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing offset",
			path:           "/manga?limit=10",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `Parameter "offset" is missing or invalid`,
		},
		{
			name:           "missing limit",
			path:           "/manga?offset=0",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `Parameter "limit" is missing or invalid`,
		},
		{
			name:           "invalid offset",
			path:           "/manga?offset=abc&limit=10",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `Parameter "offset" is missing or invalid`,
		},
		{
			name:           "invalid limit",
			path:           "/manga?offset=0&limit=xyz",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `Parameter "limit" is missing or invalid`,
		},
		{
			name:           "negative offset",
			path:           "/manga?offset=-1&limit=10",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.Get(t, tt.path, nil)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)

			if tt.expectedBody != "" {
				testutil.AssertBodyContains(t, resp, tt.expectedBody)
			}

			if tt.expectedStatus == http.StatusOK {
				testutil.AssertContentType(t, resp, "application/json")
			}
		})
	}
}

func TestGetManga(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testutil.RequireTestDB(t)
	defer testutil.CleanupTestData(t)

	router := setupTestRouter(t)
	ts := testutil.NewTestServer(t, router)
	defer ts.Close()

	// Create a test manga
	mangaIDs := testutil.CreateTestManga(t, 1)
	mangaID := mangaIDs[0]

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "existing manga",
			path:           "/manga/" + string(rune(mangaID)),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent manga",
			path:           "/manga/999999999",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		},
		{
			name:           "invalid ID",
			path:           "/manga/abc",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.Get(t, tt.path, nil)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)

			if tt.expectedBody != "" {
				testutil.AssertBodyContains(t, resp, tt.expectedBody)
			}
		})
	}
}
