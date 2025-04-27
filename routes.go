package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	mux.HandleFunc("GET /",app.home)
	return standard.Then(mux)
}
