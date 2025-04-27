package main

import (
	"os"

	"github.com/joho/godotenv"
)

// getPort retrieves the port from environment variables with fallback to .env file.
// If GETTER_PORT is set in the environment, it will use that value.
// Otherwise, it will try to load the port from the .env file.
// If neither source provides a port, it defaults to ":8080".
//
// Returns:
//   - string: The port to use for the application
func getPort() string {
	// First check if the port is set in the environment variables
	port := os.Getenv("GETTER_PORT")
	if port != "" {
		return port
	}

	// If not found in environment variables, try to load from .env file
	err := godotenv.Load()
	if err == nil {
		// Check again after loading .env file
		port = os.Getenv("GETTER_PORT")
		if port != "" {
			return port
		}
	}

	// Default port if not found in either place
	return ":8080"
}