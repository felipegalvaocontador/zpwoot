package middleware

import (
	"fmt"
	"net/http"
	"time"

	"zpwoot/platform/logger"

	"github.com/go-chi/chi/v5"
)

func HTTPLogger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			latency := time.Since(start)
			statusCode := ww.statusCode

			fields := map[string]interface{}{
				"component":      "http",
				"method":         r.Method,
				"path":           r.URL.Path,
				"route":          getRoutePattern(r),
				"status_code":    statusCode,
				"latency_ms":     latency.Milliseconds(),
				"latency_human":  latency.String(),
				"ip":             getLoggerClientIP(r),
				"user_agent":     r.Header.Get("User-Agent"),
				"content_length": ww.bytesWritten,
				"protocol":       r.Proto,
			}

			if queryString := r.URL.RawQuery; queryString != "" {
				fields["query"] = queryString
			}

			if contentType := r.Header.Get("Content-Type"); contentType != "" {
				fields["content_type"] = contentType
			}

			if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
				fields["session_id"] = sessionID
			}

			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				fields["request_id"] = requestID
			}

			message := fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)

			switch {
			case statusCode >= 500:
				logger.ErrorWithFields(message, fields)
			case statusCode >= 400:
				logger.WarnWithFields(message, fields)
			case statusCode >= 300:
				logger.InfoWithFields(message, fields)
			default:
				logger.DebugWithFields(message, fields)
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func getRoutePattern(r *http.Request) string {
	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		return routeCtx.RoutePattern()
	}
	return r.URL.Path
}

func getLoggerClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return r.RemoteAddr
}
