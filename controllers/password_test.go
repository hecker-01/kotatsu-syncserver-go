package controllers_test

import (
	"net/http"
	"testing"

	"github.com/hecker-01/kotatsu-syncserver-go/models"
	"github.com/hecker-01/kotatsu-syncserver-go/testutil"
)

func TestForgotPassword(t *testing.T) {
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
	testutil.CreateTestUserWithCredentials(t, db, "reset@example.com", "password123", "TestUser")

	tests := []struct {
		name           string
		email          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "existing user",
			email:          "reset@example.com",
			expectedStatus: http.StatusOK,
			expectedBody:   "A password reset email was sent",
		},
		{
			name:           "non-existent user (still returns 200)",
			email:          "notfound@example.com",
			expectedStatus: http.StatusOK,
			expectedBody:   "A password reset email was sent",
		},
		{
			name:           "empty email (still returns 200)",
			email:          "",
			expectedStatus: http.StatusOK,
			expectedBody:   "A password reset email was sent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.ForgotPasswordRequest{
				Email: tt.email,
			}
			resp := ts.Post(t, "/forgot-password", reqBody, nil)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)
			testutil.AssertBodyContains(t, resp, tt.expectedBody)
		})
	}
}

func TestResetPassword(t *testing.T) {
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
		token          string
		password       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid token",
			token:          "invalidtoken123",
			password:       "newpass123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid or expired token",
		},
		{
			name:           "password too short",
			token:          "sometoken",
			password:       "a",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Password should be from 2 to 24 characters long",
		},
		{
			name:           "password too long",
			token:          "sometoken",
			password:       "12345678901234567890123456",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Password should be from 2 to 24 characters long",
		},
		{
			name:           "missing token",
			token:          "",
			password:       "newpass123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.ResetPasswordRequest{
				ResetToken: tt.token,
				Password:   tt.password,
			}
			resp := ts.Post(t, "/reset-password", reqBody, nil)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)
			testutil.AssertBodyContains(t, resp, tt.expectedBody)
		})
	}
}

func TestDeeplinkResetPassword(t *testing.T) {
	router := setupTestRouter(t)
	ts := testutil.NewTestServer(t, router)
	defer ts.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid token",
			path:           "/deeplink/reset-password?token=abc123",
			expectedStatus: http.StatusOK,
			expectedBody:   "Reset Your Password",
		},
		{
			name:           "missing token",
			path:           "/deeplink/reset-password",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing token",
		},
		{
			name:           "token in HTML",
			path:           "/deeplink/reset-password?token=mytoken123",
			expectedStatus: http.StatusOK,
			expectedBody:   "Open in Kotatsu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.Get(t, tt.path, nil)
			defer resp.Body.Close()

			testutil.AssertStatus(t, resp, tt.expectedStatus)
			testutil.AssertBodyContains(t, resp, tt.expectedBody)

			if tt.expectedStatus == http.StatusOK {
				testutil.AssertContentType(t, resp, "text/html")
			}
		})
	}
}
