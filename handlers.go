package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/RAshkettle/getter/internal/files"
)

// home handles HTTP requests to the application's root endpoint.
// It returns a JSON response containing a list of all files in the application's
// configured data directory, along with success status and count information.
//
// The response has the following structure:
//   - status: A string indicating request processing status ("success")
//   - files: An array of strings with the names of files in the data directory
//   - count: An integer representing the total number of files
//
// If the file listing operation fails, a 500 Internal Server Error is returned
// and the error is logged with detailed information.
//
// Parameters:
//   - w: The HTTP response writer for sending the response
//   - r: The HTTP request being processed
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get all files in the data directory
	fileList, err := files.ListFilesInDirectory(app.dataPath)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Create a response structure
	response := map[string]interface{}{
		"status": "success",
		"files":  fileList,
		"count":  len(fileList),
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the JSON response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

// getFileRecords handles requests for all records from a JSON file.
// The filename is extracted from the URL path and the corresponding JSON file
// is loaded from the application's data directory.
//
// URL Pattern: /{filename} - where filename should be a JSON file (without the .json extension)
//
// Parameters:
//   - w: The HTTP response writer for sending the response
//   - r: The HTTP request being processed
func (app *application) getFileRecords(w http.ResponseWriter, r *http.Request) {
	// Extract filename from the URL path
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Missing file name", http.StatusBadRequest)
		return
	}

	
	if !strings.HasSuffix(filename, ".json") {
		filename = filename + ".json"
	}
	filePath := filepath.Join(app.dataPath, filename)
fileContent, err := getRecords(filePath)
	if err != nil{
		app.serverError(w,r,err)
		return 
	}

	// Validate JSON format
	var records interface{}
	if err := json.Unmarshal(fileContent, &records); err != nil {
		app.serverError(w, r, fmt.Errorf("invalid JSON in file %s: %w", filename, err))
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(fileContent)
}

func getRecords(filepath string) ([]byte, error) {

	// Enforce .json extension
	if !strings.HasSuffix(filepath, ".json") {
		filepath = filepath + ".json"
	}



	// Check if the file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, err
	}

	fileContent, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return fileContent, nil 
}

