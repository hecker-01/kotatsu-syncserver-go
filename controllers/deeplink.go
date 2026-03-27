// Package controllers provides HTTP handlers for the Kotatsu Sync Server.
// DeeplinkController handles deep link generation for mobile app integration.
package controllers

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
)

// DeeplinkController handles deep link endpoints for mobile app integration.
type DeeplinkController struct{}

// NewDeeplinkController creates a new deeplink controller.
func NewDeeplinkController() *DeeplinkController {
	return &DeeplinkController{}
}

// resetPasswordTemplate is the HTML template for the reset password deep link page.
const resetPasswordTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Password - Kotatsu</title>
    <style>
        body { font-family: sans-serif; text-align: center; padding: 50px; }
        a { color: #007bff; font-size: 1.2em; }
    </style>
</head>
<body>
    <h1>Reset Your Password</h1>
    <p>Click the button below to open Kotatsu and reset your password:</p>
    <p><a href="{{.DeepLink}}">Open in Kotatsu</a></p>
    <p><small>If the link doesn't work, make sure you have Kotatsu installed.</small></p>
</body>
</html>`

// templateData holds the data for rendering the reset password template.
type templateData struct {
	DeepLink string
}

// ResetPasswordDeeplink handles GET /deeplink/reset-password.
// It generates an HTML page containing a kotatsu:// deep link for password reset.
// Requires a 'token' query parameter; returns 400 if missing.
func (c *DeeplinkController) ResetPasswordDeeplink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	// Get BASE_URL from environment with default fallback
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9292"
	}

	// Build the kotatsu:// deep link with properly escaped parameters
	deepLink := "kotatsu://reset-password?base_url=" + url.QueryEscape(baseURL) + "&token=" + url.QueryEscape(token)

	// Parse and execute template
	tmpl, err := template.New("resetPassword").Parse(resetPasswordTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	data := templateData{DeepLink: deepLink}
	if err := tmpl.Execute(w, data); err != nil {
		// Response already started, log error but can't change status
		return
	}
}
