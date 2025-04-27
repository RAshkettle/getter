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

// getFileRecordByID handles requests for a single record by ID from a JSON file.
// It retrieves the record that matches the specified ID from the JSON file.
// The file is expected to contain a single JSON object with a property containing an array of records.
//
// URL Pattern: /{filename}/{id} - where:
//   - filename should be a JSON file (without the .json extension)
//   - id is the unique identifier for the record to retrieve
//
// Parameters:
//   - w: The HTTP response writer for sending the response
//   - r: The HTTP request being processed
func (app *application) getFileRecordByID(w http.ResponseWriter, r *http.Request) {
	// Extract filename and ID from the URL path
	filename := r.PathValue("filename")
	id := r.PathValue("id")
	
	// Validate inputs
	if filename == "" {
		http.Error(w, "Missing file name", http.StatusBadRequest)
		return
	}
	
	if id == "" {
		http.Error(w, "Missing record ID", http.StatusBadRequest)
		return
	}
	
	// Add .json extension if needed
	if !strings.HasSuffix(filename, ".json") {
		filename = filename + ".json"
	}
	
	// Construct full file path
	filePath := filepath.Join(app.dataPath, filename)
	
	// Get file content
	fileContent, err := getRecords(filePath)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("error reading file %s: %w", filename, err))
		return
	}
	
	// Parse the JSON file - it contains a single object with a property that holds an array of records
	var fileData map[string][]map[string]interface{}
	if err := json.Unmarshal(fileContent, &fileData); err != nil {
		app.serverError(w, r, fmt.Errorf("invalid JSON in file %s: %w", filename, err))
		return
	}
	
	// Find the array of records (we don't know the key name in advance)
	var records []map[string]interface{}
	var found bool
	
	// Check each key in the object to find an array of records
	for _, value := range fileData {
		if len(value) > 0 {
			// We found an array with at least one record
			records = value
			found = true
			break
		}
	}
	
	if !found {
		// No arrays with records found
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}
	
	// Search for the record with matching ID
	var matchedRecord map[string]interface{}
	for _, record := range records {
		// Convert IDs to strings for reliable comparison
		recordID := fmt.Sprintf("%v", record["id"])
		if recordID == id {
			matchedRecord = record
			break
		}
	}
	
	// If no matching record was found, return an empty object
	if matchedRecord == nil {
		matchedRecord = make(map[string]interface{})
	}
	
	// Set content type header
	w.Header().Set("Content-Type", "application/json")
	
	// Send JSON response
	if err := json.NewEncoder(w).Encode(matchedRecord); err != nil {
		app.serverError(w, r, fmt.Errorf("error encoding response: %w", err))
	}
}

// getRecords loads and returns the contents of a JSON file at the specified path.
// It ensures the file has a .json extension, checks for file existence,
// and reads the file contents into memory.
//
// Parameters:
//   - filepath: The path to the JSON file to read, with or without ".json" extension
//
// Returns:
//   - []byte: The raw file contents if successful
//   - error: An error if the file doesn't exist, can't be read, or another error occurs
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

