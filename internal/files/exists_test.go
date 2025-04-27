package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandAbsolutePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "expandpath_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create a nested directory structure for testing
	nestedDir := filepath.Join(tempDir, "nested", "path")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Get the home directory for tilde expansion tests
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "Absolute path",
			path:     tempDir,
			expected: tempDir,
			wantErr:  false,
		},
		{
			name:     "Relative path",
			path:     ".",
			expected: mustAbs(t, "."),
			wantErr:  false,
		},
		{
			name:     "Home directory with tilde",
			path:     "~",
			expected: homeDir,
			wantErr:  false,
		},
		{
			name:     "Path with tilde and subdirectory",
			path:     "~/Documents",
			expected: filepath.Join(homeDir, "Documents"),
			wantErr:  false,
		},
		{
			name:     "Nested directory",
			path:     filepath.Join(tempDir, "nested"),
			expected: filepath.Join(tempDir, "nested"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandAbsolutePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandAbsolutePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Normalize paths for comparison (especially important on Windows)
			expectedPath := filepath.Clean(tt.expected)
			gotPath := filepath.Clean(got)
			
			if gotPath != expectedPath {
				t.Errorf("ExpandAbsolutePath() got = %v, want %v", gotPath, expectedPath)
			}
		})
	}
}

// TestExpandAbsolutePathWithNonExistentPath tests expanding paths that don't exist yet
func TestExpandAbsolutePathWithNonExistentPath(t *testing.T) {
	// Get the home directory for tilde expansion tests
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	// Create a unique non-existent path
	nonExistentPath := filepath.Join(os.TempDir(), "nonexistent_"+randString(8))
	
	tests := []struct {
		name     string
		path     string
		contains string // We use contains instead of exact match for non-existent paths
		wantErr  bool
	}{
		{
			name:     "Non-existent absolute path",
			path:     nonExistentPath,
			contains: "nonexistent_",
			wantErr:  false,
		},
		{
			name:     "Non-existent path with tilde",
			path:     "~/nonexistent_" + randString(8),
			contains: homeDir,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandAbsolutePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandAbsolutePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !strings.Contains(got, tt.contains) {
				t.Errorf("ExpandAbsolutePath() got = %v, which doesn't contain %v", got, tt.contains)
			}
		})
	}
}

// TestFolderExists tests the FolderExists function
func TestFolderExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "folderexists_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create a nested directory structure for testing
	nestedDir := filepath.Join(tempDir, "nested", "path")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a nonexistent path
	nonExistentPath := filepath.Join(tempDir, "nonexistent_"+randString(8))

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing directory",
			path:     tempDir,
			expected: true,
		},
		{
			name:     "Nested directory",
			path:     nestedDir,
			expected: true,
		},
		{
			name:     "File (not a directory)",
			path:     testFile,
			expected: false,
		},
		{
			name:     "Non-existent path",
			path:     nonExistentPath,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FolderExists(tt.path)
			if got != tt.expected {
				t.Errorf("FolderExists() got = %v, want %v for path %v", got, tt.expected, tt.path)
			}
		})
	}
}

// TestFileExists tests the FileExists function
func TestFileExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fileexists_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create a test file
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test file with spaces in the name
	testFileWithSpaces := filepath.Join(tempDir, "test file with spaces.txt")
	if err := os.WriteFile(testFileWithSpaces, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file with spaces: %v", err)
	}

	// Create an empty file
	emptyFile := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Create a nonexistent path
	nonExistentPath := filepath.Join(tempDir, "nonexistent_"+randString(8)+".txt")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing file",
			path:     testFile,
			expected: true,
		},
		{
			name:     "File with spaces in name",
			path:     testFileWithSpaces,
			expected: true,
		},
		{
			name:     "Empty file",
			path:     emptyFile,
			expected: true,
		},
		{
			name:     "Directory (not a file)",
			path:     tempDir,
			expected: false,
		},
		{
			name:     "Non-existent file",
			path:     nonExistentPath,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExists(tt.path)
			if got != tt.expected {
				t.Errorf("FileExists() got = %v, want %v for path %v", got, tt.expected, tt.path)
			}
		})
	}
}

// Helper functions

// mustAbs gets the absolute path and fails the test if it encounters an error
func mustAbs(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Failed to get absolute path for %s: %v", path, err)
	}
	return abs
}

// randString generates a random string of specified length
// This is a simple implementation for testing purposes
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

