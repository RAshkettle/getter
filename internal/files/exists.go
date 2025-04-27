// Package files provides utility functions for file system operations
// such as checking file/folder existence and path manipulation.
package files

import (
	"os"
	"path/filepath"
)

// FolderExists checks if a folder exists and is a directory.
// It returns true if the path exists and is a directory,
// and false if the path doesn't exist, is a file, or there was an error accessing it.
//
// Parameters:
//   - path: The file system path to check
//
// Returns:
//   - bool: True if the path exists and is a directory, otherwise false
func FolderExists(path string) bool {

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists checks if a file exists and is not a directory.
// It returns true if the path exists and is a regular file,
// and false if the path doesn't exist, is a directory, or there was an error accessing it.
//
// Parameters:
//   - path: The file system path to check
//
// Returns:
//   - bool: True if the path exists and is a file, otherwise false
func FileExists(path string) bool {

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ExpandAbsolutePath takes a path (which may contain ~ for home directory)
// and returns the absolute path with any ~ expanded to the user's home directory.
// This is useful for handling user-provided paths that may use the tilde shorthand.
//
// Parameters:
//   - path: The path string to expand, may contain a leading tilde (~)
//
// Returns:
//   - string: The expanded absolute path
//   - error: An error if home directory expansion or absolute path conversion fails
func ExpandAbsolutePath(path string) (string, error) {
	// Expand the path if it contains a tilde
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[1:])
	}
	
	// Make the path absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	
	return absPath, nil
}