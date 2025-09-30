package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"zpwoot/platform/logger"
)

// RequestID middleware for Chi router
func RequestID(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
				w.Header().Set("X-Request-ID", requestID)
			}

			// Add request ID to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			
			// Create logger with request ID
			requestLogger := logger.WithField("request_id", requestID)
			ctx = context.WithValue(ctx, "logger", requestLogger)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLoggerFromContext extracts logger from Chi context
func GetLoggerFromContext(r *http.Request) *logger.Logger {
	if logger, ok := r.Context().Value("logger").(*logger.Logger); ok {
		return logger
	}
	return logger.New()
}

// GetRequestIDFromContext extracts request ID from Chi context
func GetRequestIDFromContext(r *http.Request) string {
	if requestID, ok := r.Context().Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
