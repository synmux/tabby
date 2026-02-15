// Package server provides the HTTP server, routing, and middleware for Tabby.
package server

import (
	"log/slog"
	"net/http"
	"time"
)

// responseCapture wraps http.ResponseWriter to capture the status code.
type responseCapture struct {
	http.ResponseWriter
	statusCode int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs each HTTP request with method, path, status, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		capture := &responseCapture{ResponseWriter: writer, statusCode: http.StatusOK}

		next.ServeHTTP(capture, request)

		slog.Info("request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", capture.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}

// RecoveryMiddleware catches panics in handlers and returns a 500 response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered",
					"error", recovered,
					"path", request.URL.Path,
				)
				http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(writer, request)
	})
}
