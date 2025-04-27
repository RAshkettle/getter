package files

import (
	"os"
)

// ListFilesInDirectory returns a slice of filenames (without path) of all files in the specified directory.
// It does not include subdirectories in the returned list, only files directly in the specified directory.
//
// Parameters:
//   - dirPath: The path to the directory whose files should be listed
//
// Returns:
//   - []string: A slice of filenames in the directory
//   - error: An error if reading the directory fails or if the provided path is not a directory
func ListFilesInDirectory(dirPath string) ([]string, error) {
	// Check if the path exists and is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	// Read all entries in the directory
	entries, err := dir.ReadDir(-1) // -1 means read all entries
	if err != nil {
		return nil, err
	}

	// Filter out directories, keep only files
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

