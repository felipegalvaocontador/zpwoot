package middleware

import (
	"net/http"
	"time"

	"zpwoot/internal/app"
	"zpwoot/platform/logger"
)

// Metrics middleware for Chi router
func Metrics(container *app.Container, logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			statusCode := ww.statusCode

			// Log metrics
			fields := map[string]interface{}{
				"component":    "metrics",
				"method":       r.Method,
				"path":         r.URL.Path,
				"status_code":  statusCode,
				"duration_ms":  duration.Milliseconds(),
				"duration_ns":  duration.Nanoseconds(),
			}

			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				fields["request_id"] = requestID
			}

			if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
				fields["session_id"] = sessionID
			}

			// Log based on status code
			switch {
			case statusCode >= 500:
				logger.ErrorWithFields("HTTP request completed", fields)
			case statusCode >= 400:
				logger.WarnWithFields("HTTP request completed", fields)
			default:
				logger.DebugWithFields("HTTP request completed", fields)
			}

			// Here you could add custom metrics collection
			// For example, if you have a metrics service in the container:
			// container.MetricsService.RecordHTTPRequest(r.Method, r.URL.Path, statusCode, duration)
		})
	}
}
