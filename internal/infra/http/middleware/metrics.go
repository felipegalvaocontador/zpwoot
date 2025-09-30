package middleware

import (
	"net/http"
	"time"

	"zpwoot/internal/app"
	"zpwoot/platform/logger"
)

func Metrics(container *app.Container, logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			statusCode := ww.statusCode

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

			switch {
			case statusCode >= 500:
				logger.ErrorWithFields("HTTP request completed", fields)
			case statusCode >= 400:
				logger.WarnWithFields("HTTP request completed", fields)
			default:
				logger.DebugWithFields("HTTP request completed", fields)
			}

		})
	}
}
