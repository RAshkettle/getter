package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPort(t *testing.T) {
	// Save original environment to restore later
	originalPort := os.Getenv("GETTER_PORT")
	defer os.Setenv("GETTER_PORT", originalPort)

	// Create a temporary directory for our test .env files
	tempDir, err := os.MkdirTemp("", "envtest")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Store current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tests := []struct {
		name           string
		envValue       string
		envFileContent string
		expected       string
		setupFunc      func(t *testing.T)
	}{
		{
			name:      "Port from environment variable",
			envValue:  ":3000",
			expected:  ":3000",
			setupFunc: func(t *testing.T) {
				os.Setenv("GETTER_PORT", ":3000")
				// Move to temp dir where there's no .env file
				if err := os.Chdir(tempDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}
			},
		},
		{
			name:      "No environment variable, no env file",
			envValue:  "",
			expected:  ":8080", // Default value
			setupFunc: func(t *testing.T) {
				os.Unsetenv("GETTER_PORT")
				// Move to temp dir where there's no .env file
				if err := os.Chdir(tempDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}
			},
		},
		{
			name:           "No environment variable, use env file",
			envValue:       "",
			envFileContent: "GETTER_PORT=:5000",
			expected:       ":5000",
			setupFunc: func(t *testing.T) {
				os.Unsetenv("GETTER_PORT")
				
				// Create a temporary directory with a .env file
				testDir := filepath.Join(tempDir, "withenv")
				if err := os.Mkdir(testDir, 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
				
				// Create a .env file in the test directory
				envPath := filepath.Join(testDir, ".env")
				if err := os.WriteFile(envPath, []byte("GETTER_PORT=:5000"), 0644); err != nil {
					t.Fatalf("Failed to write test .env file: %v", err)
				}
				
				// Change to the test directory
				if err := os.Chdir(testDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}
			},
		},
		{
			name:           "Environment variable takes precedence over env file",
			envValue:       ":4000",
			envFileContent: "GETTER_PORT=:5000",
			expected:       ":4000",
			setupFunc: func(t *testing.T) {
				os.Setenv("GETTER_PORT", ":4000")
				
				// Create a temporary directory with a .env file
				testDir := filepath.Join(tempDir, "envprecedence")
				if err := os.Mkdir(testDir, 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
				
				// Create a .env file in the test directory
				envPath := filepath.Join(testDir, ".env")
				if err := os.WriteFile(envPath, []byte("GETTER_PORT=:5000"), 0644); err != nil {
					t.Fatalf("Failed to write test .env file: %v", err)
				}
				
				// Change to the test directory
				if err := os.Chdir(testDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test case
			tt.setupFunc(t)

			// Call the function
			got := getPort()

			// Check results
			if got != tt.expected {
				t.Errorf("getPort() = %v, want %v", got, tt.expected)
			}
			
			// Return to the original directory for the next test
			if err := os.Chdir(currentDir); err != nil {
				t.Fatalf("Failed to restore original directory: %v", err)
			}
		})
	}
}