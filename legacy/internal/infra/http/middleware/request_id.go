package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"zpwoot/platform/logger"
)

type requestContextKey string

const (
	requestIDContextKey requestContextKey = "request_id"
	loggerContextKey    requestContextKey = "logger"
)

func RequestID(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
				w.Header().Set("X-Request-ID", requestID)
			}

			ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)

			requestLogger := logger.WithField("request_id", requestID)
			ctx = context.WithValue(ctx, loggerContextKey, requestLogger)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetLoggerFromContext(r *http.Request) *logger.Logger {
	if logger, ok := r.Context().Value(loggerContextKey).(*logger.Logger); ok {
		return logger
	}
	return logger.New()
}

func GetRequestIDFromContext(r *http.Request) string {
	if requestID, ok := r.Context().Value(requestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
