package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCommonHeaders tests that the commonHeaders middleware correctly sets
// all the security headers on HTTP responses
func TestCommonHeaders(t *testing.T) {
	// Create a simple next handler that just returns 200 OK
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the next handler with our commonHeaders middleware
	headersHandler := commonHeaders(nextHandler)

	// Create a test HTTP request
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Call the handler with our request
	headersHandler.ServeHTTP(w, r)

	// Check status code is 200 OK (the next handler's response)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Expected headers and their values
	expectedHeaders := map[string]string{
		"Content-Security-Policy": "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		"Referrer-Policy":         "origin-when-cross-origin",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "deny",
		"X-XSS-Protection":        "0",
		"Server":                  "Go",
	}

	// Check that all expected headers are set with correct values
	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s to be %q, got %q", header, expectedValue, actualValue)
		}
	}
}

// TestCommonHeadersMiddlewareChain tests that the commonHeaders middleware
// works correctly when combined with other middleware
func TestCommonHeadersMiddlewareChain(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a logger that writes to our buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create our application instance with the test logger
	app := &application{
		logger: logger,
	}

	// Final handler that sets a custom header and returns 200 OK
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	})

	// Test different middleware chain configurations
	tests := []struct {
		name           string
		setupHandler   func() http.Handler
		expectedStatus int
		checkExtraFunc func(t *testing.T, w *httptest.ResponseRecorder, buf *bytes.Buffer)
	}{
		{
			name: "commonHeaders before custom handler",
			setupHandler: func() http.Handler {
				return commonHeaders(finalHandler)
			},
			expectedStatus: http.StatusOK,
			checkExtraFunc: func(t *testing.T, w *httptest.ResponseRecorder, buf *bytes.Buffer) {
				// Check that our custom header still exists
				if w.Header().Get("X-Custom-Header") != "custom-value" {
					t.Errorf("Expected X-Custom-Header to be 'custom-value', got %q",
						w.Header().Get("X-Custom-Header"))
				}
			},
		},
		{
			name: "commonHeaders with logRequest",
			setupHandler: func() http.Handler {
				return commonHeaders(app.logRequest(finalHandler))
			},
			expectedStatus: http.StatusOK,
			checkExtraFunc: func(t *testing.T, w *httptest.ResponseRecorder, buf *bytes.Buffer) {
				// Check that request was logged
				logOutput := buf.String()
				if !strings.Contains(logOutput, "received request") {
					t.Errorf("Expected log to contain 'received request', log output: %s", logOutput)
				}
			},
		},
		{
			name: "Full middleware chain",
			setupHandler: func() http.Handler {
				return app.recoverPanic(commonHeaders(app.logRequest(finalHandler)))
			},
			expectedStatus: http.StatusOK,
			checkExtraFunc: func(t *testing.T, w *httptest.ResponseRecorder, buf *bytes.Buffer) {
				// Check that request was logged
				logOutput := buf.String()
				if !strings.Contains(logOutput, "received request") {
					t.Errorf("Expected log to contain 'received request', log output: %s", logOutput)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the log buffer for each test
			logBuffer.Reset()

			// Get the handler with the specified middleware setup
			handler := tt.setupHandler()

			// Create a test HTTP request
			r := httptest.NewRequest(http.MethodGet, "/test-headers-chain", nil)

			// Create a test response recorder
			w := httptest.NewRecorder()

			// Call the handler with our request
			handler.ServeHTTP(w, r)

			// Check status code matches expected
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check that all security headers are set
			securityHeaders := []string{
				"Content-Security-Policy",
				"Referrer-Policy",
				"X-Content-Type-Options",
				"X-Frame-Options",
				"X-XSS-Protection",
				"Server",
			}

			for _, header := range securityHeaders {
				if w.Header().Get(header) == "" {
					t.Errorf("Expected security header %q to be set, but it was not", header)
				}
			}

			// Run additional checks specific to this test case
			if tt.checkExtraFunc != nil {
				tt.checkExtraFunc(t, w, &logBuffer)
			}
		})
	}
}

// TestCommonHeadersWithCustomResponse tests that the commonHeaders middleware
// properly handles responses with different status codes and content types
func TestCommonHeadersWithCustomResponse(t *testing.T) {
	// Test cases for different response types
	tests := []struct {
		name           string
		handlerFunc    func(w http.ResponseWriter, r *http.Request)
		expectedStatus int
		expectedBody   string
		contentType    string
	}{
		{
			name: "JSON response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success"}`))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
			contentType:    "application/json",
		},
		{
			name: "HTML response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<html><body>Hello</body></html>"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Hello</body></html>",
			contentType:    "text/html; charset=utf-8",
		},
		{
			name: "404 Not Found response",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Resource not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Resource not found",
			contentType:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler with the test-specific behavior
			testHandler := http.HandlerFunc(tt.handlerFunc)

			// Wrap with commonHeaders middleware
			handler := commonHeaders(testHandler)

			// Create a test HTTP request
			r := httptest.NewRequest(http.MethodGet, "/test-custom-response", nil)

			// Create a test response recorder
			w := httptest.NewRecorder()

			// Call the handler with our request
			handler.ServeHTTP(w, r)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			if w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, w.Body.String())
			}

			// Check content type header is preserved
			if tt.contentType != "" && w.Header().Get("Content-Type") != tt.contentType {
				t.Errorf("Expected Content-Type %q, got %q",
					tt.contentType, w.Header().Get("Content-Type"))
			}

			// Check that all security headers are still set
			securityHeaders := []string{
				"Content-Security-Policy",
				"Referrer-Policy",
				"X-Content-Type-Options",
				"X-Frame-Options",
				"X-XSS-Protection",
				"Server",
			}

			for _, header := range securityHeaders {
				if w.Header().Get(header) == "" {
					t.Errorf("Expected security header %q to be set, but it was not", header)
				}
			}
		})
	}
}
