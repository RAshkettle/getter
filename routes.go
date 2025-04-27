package main

import (
	"net/http"

	"github.com/justinas/alice"
)

// routes configures and returns the application's HTTP request router.
// It sets up all request routes and applies the standard middleware chain
// which includes panic recovery, request logging, and common headers.
//
// Routes defined:
//   - GET / : Home page that lists all available data files
//   - GET /{filename} : Returns all records from the specified JSON file
//   - GET /{filename}/{id} : Returns a single record by ID from the specified JSON file
//
// Returns:
//   - http.Handler: The configured router with all middleware applied
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	// Static routes
	mux.HandleFunc("GET /", app.home)

	// Dynamic routes for JSON files
	mux.HandleFunc("GET /{filename}", app.getFileRecords)
	mux.HandleFunc("GET /{filename}/{id}", app.getFileRecordByID)

	return standard.Then(mux)
}
