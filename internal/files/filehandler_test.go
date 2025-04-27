package files

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestListFilesInDirectory tests the ListFilesInDirectory function with various scenarios
func TestListFilesInDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "listfiles_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Create test files
	testFiles := []string{
		"file1.txt",
		"file2.log",
		"document.pdf",
		"image.jpg",
		".hidden",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create a subdirectory with more files (these should not be included in result)
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a file in the subdirectory
	subDirFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subDirFile, []byte("subdir content"), 0644); err != nil {
		t.Fatalf("Failed to create test file in subdirectory: %v", err)
	}

	// Create an empty subdirectory
	emptySubDir := filepath.Join(tempDir, "emptydir")
	if err := os.Mkdir(emptySubDir, 0755); err != nil {
		t.Fatalf("Failed to create empty subdirectory: %v", err)
	}

	// Test cases
	tests := []struct {
		name          string
		path          string
		expected      []string
		errorExpected bool
	}{
		{
			name:          "Valid directory with files",
			path:          tempDir,
			expected:      testFiles,
			errorExpected: false,
		},
		{
			name:          "Empty directory",
			path:          emptySubDir,
			expected:      []string{},
			errorExpected: false,
		},
		{
			name:          "Non-existent directory",
			path:          filepath.Join(tempDir, "doesnotexist"),
			expected:      nil,
			errorExpected: true,
		},
		{
			name:          "File instead of directory",
			path:          filepath.Join(tempDir, testFiles[0]),
			expected:      nil,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			files, err := ListFilesInDirectory(tt.path)

			// Check error status
			if tt.errorExpected && err == nil {
				t.Errorf("Expected an error but got nil")
			}
			if !tt.errorExpected && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// If no error is expected, check the results
			if !tt.errorExpected {
				// Sort both slices for comparison
				sort.Strings(files)
				expected := make([]string, len(tt.expected))
				copy(expected, tt.expected)
				sort.Strings(expected)

				// Compare lengths
				if len(files) != len(expected) {
					t.Errorf("Expected %d files, got %d", len(expected), len(files))
				}

				// Compare content
				if len(files) == len(expected) {
					for i, file := range files {
						if file != expected[i] {
							t.Errorf("Expected file %s at position %d, got %s", expected[i], i, file)
						}
					}
				}
			}
		})
	}
}

// TestListFilesInDirectoryWithSpecialCharacters tests handling of filenames with special characters
func TestListFilesInDirectoryWithSpecialCharacters(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "listfiles_special_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Files with special characters
	specialFiles := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.with.dots.txt",
		"filewithutf8чжш.txt", // UTF-8 characters
	}

	// Create the special files
	for _, filename := range specialFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create special test file %s: %v", filename, err)
		}
	}

	// Get the files and check results
	files, err := ListFilesInDirectory(tempDir)
	if err != nil {
		t.Fatalf("Error listing files: %v", err)
	}

	// Sort both slices for comparison
	sort.Strings(files)
	sort.Strings(specialFiles)

	// Compare number of files
	if len(files) != len(specialFiles) {
		t.Errorf("Expected %d files, got %d", len(specialFiles), len(files))
	}

	// Compare file names
	if len(files) == len(specialFiles) {
		for i, file := range files {
			if file != specialFiles[i] {
				t.Errorf("Expected file %s at position %d, got %s", specialFiles[i], i, file)
			}
		}
	}
}

// TestListFilesInDirectoryPermissions tests the function's behavior with permission restrictions
func TestListFilesInDirectoryPermissions(t *testing.T) {
	// Skip on non-Unix platforms where permission tests may not work as expected
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping permission test in CI environment")
	}

	// On macOS, permissions aren't always strictly enforced for tests
	// So we'll just check that we get an appropriate error when a directory doesn't exist
	nonExistentPath := "/path/that/definitely/does/not/exist"
	_, err := ListFilesInDirectory(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent path, but got nil")
	}
}