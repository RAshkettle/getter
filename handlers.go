package main

import (
	"encoding/json"
	"net/http"

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
func(app *application) home(w http.ResponseWriter, r *http.Request){
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