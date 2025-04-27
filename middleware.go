package main

import (
	"fmt"
	"net/http"
)

// commonHeaders is a middleware that sets common security headers for all HTTP responses.
// It applies several best-practice security headers to reduce common web vulnerabilities:
//   - Content-Security-Policy: Restricts which resources can be loaded
//   - Referrer-Policy: Controls how much referrer information is included with requests
//   - X-Content-Type-Options: Prevents MIME type sniffing attacks
//   - X-Frame-Options: Prevents clickjacking by disallowing your content in frames
//   - X-XSS-Protection: Explicitly disables outdated XSS protections in favor of CSP
//
// Parameters:
//   - next: The next handler in the middleware chain to be called after this middleware
//
// Returns:
//   - http.Handler: A handler that adds security headers and then calls the next handler
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content Security Policy
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")

		// Set Referrer Policy
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		// Set X-Content-Type-Options to prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Set X-Frame-Options to prevent clickjacking
		w.Header().Set("X-Frame-Options", "deny")
		// Disable X-XSS-Protection
		w.Header().Set("X-XSS-Protection", "0")
		// Set Server header
		w.Header().Set("Server", "Go")

		next.ServeHTTP(w, r)
	})
}

// logRequest is a middleware that logs details of each HTTP request.
// It captures and logs key information about incoming requests including:
//   - Client IP address
//   - HTTP protocol version
//   - HTTP method (GET, POST, etc.)
//   - Request URI
//
// This middleware is useful for monitoring and debugging traffic patterns,
// as well as for security auditing and access logging.
//
// Parameters:
//   - next: The next handler in the middleware chain to be called after this middleware
//
// Returns:
//   - http.Handler: A handler that logs request information and then calls the next handler
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		// Log the request details.
		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)
		next.ServeHTTP(w, r)
	})
}

// recoverPanic is a middleware that recovers from any panics that occur during request handling.
// It prevents a panic in one request from crashing the entire application by:
//   - Catching any panic that occurs during request processing
//   - Setting the Connection header to "close" to prevent keep-alive connections
//   - Logging detailed error information via the serverError helper
//   - Returning a 500 Internal Server Error response to the client
//
// This middleware should typically be added first in the handler chain to ensure
// it can recover from panics in any subsequent middleware or handlers.
//
// Parameters:
//   - next: The next handler in the middleware chain to be called after this middleware
//
// Returns:
//   - http.Handler: A handler that provides panic recovery before calling the next handler
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
