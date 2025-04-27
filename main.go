// Package main provides the entry point for the getter application.
// It processes command-line arguments and initializes the application.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/RAshkettle/getter/internal/files"
)

// application represents the main application instance with its configuration.
// It holds essential components like the logger and data path.
type application struct {
	logger   *slog.Logger
	dataPath string
}

// main is the entry point of the application.
// It validates command-line arguments, initializes the application,
// and starts the main execution flow.
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: getter <folder>  Example:  getter '~/tempData'")
		os.Exit(1)
	}

	// Determine if the datapath is a valid directory
	dataPath, err := getDataPath(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Load the port from environment variables or .env file
	port := getPort()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		logger:   logger,
		dataPath: dataPath,
	}
	srv := &http.Server{
		Addr:         port,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	app.logger.Info("Initialized application", "dataPath", app.dataPath, "port", srv.Addr)
	serverErr := srv.ListenAndServe()
	logger.Error(serverErr.Error())
	os.Exit(1)
}

// serverError handles internal server errors by logging detailed error information
// and returning a generic 500 Internal Server Error response to the client.
// This function logs the original error, HTTP method, URI, and a stack trace to aid debugging,
// while preventing sensitive error details from being exposed to clients.
//
// Parameters:
//   - w: The HTTP response writer to send the error response
//   - r: The HTTP request that resulted in the error
//   - err: The error that occurred during request processing
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// getDataPath processes and validates a data path string.
// It converts the provided path to an absolute path with expanded home directory symbols,
// and verifies that the path exists and is a directory.
//
// Parameters:
//   - dataPath: The path string to process, can include tilde (~) for home directory
//
// Returns:
//   - string: The validated and expanded absolute data path
//   - error: An error if the path expansion fails or the path does not exist or is not a directory
func getDataPath(dataPath string) (string, error) {

	dataPath, err := files.ExpandAbsolutePath(dataPath)
	if err != nil {
		return "", err
	}

	// Check if the folder exists
	if !files.FolderExists(dataPath) {
		return "", errors.New("The path does not exist or is not a directory")
	}
	return dataPath, nil
}
