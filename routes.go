package main

import (
	"net/http"

	"github.com/justinas/alice"
)

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
