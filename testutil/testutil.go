// Package testutil provides reusable test utilities and helpers for
// HTTP API testing of the Kotatsu sync server.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestServer wraps httptest.Server with helper methods for API testing.
type TestServer struct {
	*httptest.Server
	Router chi.Router
}

// NewTestServer creates a test server with the given router.
// The server is automatically closed when the test completes.
func NewTestServer(t *testing.T, router chi.Router) *TestServer {
	t.Helper()

	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
	})

	return &TestServer{
		Server: server,
		Router: router,
	}
}

// Request makes an HTTP request to the test server and returns the response.
// The body parameter can be nil, a struct (will be JSON encoded), or an io.Reader.
// Headers is optional and can be nil.
func (ts *TestServer) Request(t *testing.T, method, path string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			bodyReader = v
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = strings.NewReader(v)
		default:
			// Assume it's a struct, encode as JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	url := ts.URL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Set default content type for non-nil bodies
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return resp
}

// Get is a convenience method for GET requests.
func (ts *TestServer) Get(t *testing.T, path string, headers map[string]string) *http.Response {
	t.Helper()
	return ts.Request(t, http.MethodGet, path, nil, headers)
}

// Post is a convenience method for POST requests.
func (ts *TestServer) Post(t *testing.T, path string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()
	return ts.Request(t, http.MethodPost, path, body, headers)
}

// Put is a convenience method for PUT requests.
func (ts *TestServer) Put(t *testing.T, path string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()
	return ts.Request(t, http.MethodPut, path, body, headers)
}

// Delete is a convenience method for DELETE requests.
func (ts *TestServer) Delete(t *testing.T, path string, headers map[string]string) *http.Response {
	t.Helper()
	return ts.Request(t, http.MethodDelete, path, nil, headers)
}

// AuthHeader returns a map with the Authorization header set to the given token.
func AuthHeader(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// JSONBody creates a request body reader from a struct by JSON encoding it.
// Returns a bytes.Reader that can be used as a request body.
func JSONBody(v interface{}) *bytes.Reader {
	data, err := json.Marshal(v)
	if err != nil {
		// Return empty reader on error; actual tests will fail on parsing
		return bytes.NewReader([]byte{})
	}
	return bytes.NewReader(data)
}

// ParseJSON parses the response body into the provided struct.
// Fails the test if parsing fails.
func ParseJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("failed to parse JSON response: %v\nBody: %s", err, string(body))
	}
}

// ReadBody reads and returns the response body as a string.
// The response body is closed after reading.
func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(body)
}

// AssertStatus checks that the response has the expected status code.
// Reports an error (not fatal) if the status code doesn't match.
func AssertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		// Read body for error context
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		t.Errorf("expected status %d, got %d\nBody: %s", expected, resp.StatusCode, string(body))
	}
}

// AssertBodyContains checks that the response body contains the expected string.
// Reports an error (not fatal) if the string is not found.
func AssertBodyContains(t *testing.T, resp *http.Response, expected string) {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), expected) {
		t.Errorf("expected body to contain %q, got: %s", expected, string(body))
	}
}

// AssertJSONField checks that a JSON response has a field with the expected value.
// The value is compared as a string representation.
func AssertJSONField(t *testing.T, resp *http.Response, field string, expected interface{}) {
	t.Helper()

	var result map[string]interface{}
	ParseJSON(t, resp, &result)

	actual, ok := result[field]
	if !ok {
		t.Errorf("expected field %q not found in response", field)
		return
	}

	// Compare values
	if actual != expected {
		t.Errorf("field %q: expected %v, got %v", field, expected, actual)
	}
}

// AssertHeader checks that a response header has the expected value.
func AssertHeader(t *testing.T, resp *http.Response, header, expected string) {
	t.Helper()

	actual := resp.Header.Get(header)
	if actual != expected {
		t.Errorf("header %q: expected %q, got %q", header, expected, actual)
	}
}

// AssertContentType checks that the response has the expected Content-Type header.
func AssertContentType(t *testing.T, resp *http.Response, expected string) {
	t.Helper()

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, expected) {
		t.Errorf("expected Content-Type %q, got %q", expected, contentType)
	}
}
