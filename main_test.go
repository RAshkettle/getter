package main

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestServerError tests that the serverError function properly logs error details
// and returns a 500 response with the correct content
func TestServerError(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a logger that writes to our buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelError, // Ensure we capture error level logs
	}))

	// Create our application instance with the test logger
	app := &application{
		logger: logger,
	}

	// Create test cases
	tests := []struct {
		name           string
		method         string
		url            string
		err            error
		expectedStatus int
		expectedBody   string
		logChecks      []string // Strings that should appear in the logs
	}{
		{
			name:           "Basic server error",
			method:         http.MethodGet,
			url:            "/test-path",
			err:            errors.New("test server error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
			logChecks: []string{
				"test server error", // Error message
				http.MethodGet,      // HTTP method
				"/test-path",        // Request URI
				"trace",             // Stack trace marker
			},
		},
		{
			name:           "POST request error",
			method:         http.MethodPost,
			url:            "/api/submit",
			err:            errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
			logChecks: []string{
				"database connection failed", // Error message
				http.MethodPost,              // HTTP method
				"/api/submit",                // Request URI
				"trace",                      // Stack trace marker
			},
		},
		{
			name:           "Error with query parameters",
			method:         http.MethodGet,
			url:            "/products?id=123&category=electronics",
			err:            errors.New("product not found"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError),
			logChecks: []string{
				"product not found",                     // Error message
				http.MethodGet,                          // HTTP method
				"/products?id=123&category=electronics", // Request URI with query params
				"trace",                                 // Stack trace marker
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the log buffer for each test
			logBuffer.Reset()

			// Create a test HTTP request
			r := httptest.NewRequest(tt.method, tt.url, nil)

			// Create a test response recorder
			w := httptest.NewRecorder()

			// Call the function we're testing
			app.serverError(w, r, tt.err)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q",
					tt.expectedBody, w.Body.String())
			}

			// Check log output
			logOutput := logBuffer.String()
			for _, check := range tt.logChecks {
				if !strings.Contains(logOutput, check) {
					t.Errorf("Expected log to contain %q, log output: %q", check, logOutput)
				}
			}
		})
	}
}

// TestGetDataPath tests the getDataPath function with various input scenarios
func TestGetDataPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "getdatapath_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Create a nested directory for testing
	nestedDir := filepath.Join(tempDir, "nested", "path")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Create a test file (not a directory)
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Non-existent path
	nonExistentPath := filepath.Join(tempDir, "does-not-exist")

	// Test cases
	tests := []struct {
		name          string
		path          string
		expected      string
		errorExpected bool
		errorContains string
	}{
		{
			name:          "Valid directory path",
			path:          tempDir,
			expected:      tempDir,
			errorExpected: false,
		},
		{
			name:          "Valid nested directory path",
			path:          nestedDir,
			expected:      nestedDir,
			errorExpected: false,
		},
		{
			name:          "File path instead of directory",
			path:          testFile,
			expected:      "",
			errorExpected: true,
			errorContains: "not a directory",
		},
		{
			name:          "Non-existent path",
			path:          nonExistentPath,
			expected:      "",
			errorExpected: true,
			errorContains: "not exist",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			result, err := getDataPath(tt.path)

			// Check error expectation
			if tt.errorExpected {
				if err == nil {
					t.Errorf("Expected an error but got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q but got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
			}

			// Check result value
			if result != tt.expected {
				t.Errorf("Expected path %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestGetDataPathWithHomeExpansion tests how getDataPath handles home directory expansion
func TestGetDataPathWithHomeExpansion(t *testing.T) {
	// This test depends on the environment and may need to be skipped in some cases
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Skipping test because user home directory cannot be determined")
	}

	// Create a test directory in the home directory (safely)
	testDir := filepath.Join(homeDir, "getdatapath_test_"+randString(8))
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory in home: %v", err)
	}
	defer os.RemoveAll(testDir) // Clean up after test

	// Convert the absolute path to a tilde-prefixed path for testing
	tildeTestPath := "~" + testDir[len(homeDir):]

	// Test expansion
	result, err := getDataPath(tildeTestPath)
	if err != nil {
		t.Errorf("Expected success with tilde path but got error: %v", err)
	}

	// The result should be the absolute path, not the tilde version
	if result != testDir {
		t.Errorf("Expected %q but got %q", testDir, result)
	}
}

// Helper function to generate random strings for test paths
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
