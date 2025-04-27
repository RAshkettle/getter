// Package main provides the entry point for the getter application.
// It processes command-line arguments and initializes the application.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/RAshkettle/getter/internal/files"
)

// application represents the main application instance with its configuration.
// It holds essential components like the logger and data path.
type application struct {
	logger   *slog.Logger
	dataPath string
	port     string
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
	dataPath := getDataPath()

	// Load the port from environment variables or .env file
	port := getPort()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		logger:   logger,
		dataPath: dataPath,
		port:     port,
	}
	
	app.logger.Info("Initialized application", "dataPath", app.dataPath, "port", app.port)
}



// getDataPath processes and validates the data path from command-line arguments.
// It expands the path, ensures it exists, and exits with an error message if invalid.
//
// Returns:
//   - string: The validated and expanded absolute data path
func getDataPath() string {
	dataPath := os.Args[1]

	dataPath, err := files.ExpandAbsolutePath(dataPath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Check if the folder exists
	if !files.FolderExists(dataPath) {
		fmt.Printf("Error: The path '%s' does not exist or is not a directory\n", dataPath)
		os.Exit(1)
	}
	return dataPath
}
