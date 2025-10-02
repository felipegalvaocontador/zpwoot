package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"zpwoot/platform/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func HTTPLogger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			fields := map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"query":       r.URL.RawQuery,
				"status_code": ww.statusCode,
				"duration_ms": duration.Milliseconds(),
				"size_bytes":  ww.size,
				"ip":          getClientIP(r),
				"user_agent":  r.Header.Get("User-Agent"),
			}

			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				fields["request_id"] = requestID
			}

			if referer := r.Header.Get("Referer"); referer != "" {
				fields["referer"] = referer
			}

			message := "HTTP request processed"
			switch {
			case ww.statusCode >= 500:
				logger.ErrorWithFields(message, fields)
			case ww.statusCode >= 400:
				logger.WarnWithFields(message, fields)
			case ww.statusCode >= 300:
				logger.InfoWithFields(message, fields)
			default:

				if r.URL.Path == "/health" {
					logger.DebugWithFields(message, fields)
				} else {
					logger.InfoWithFields(message, fields)
				}
			}
		})
	}
}

func ErrorLogger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.ErrorWithFields("HTTP handler panic", map[string]interface{}{
						"error":  err,
						"method": r.Method,
						"path":   r.URL.Path,
						"ip":     getClientIP(r),
						"stack":  string(debug.Stack()),
					})

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func PerformanceLogger(logger *logger.Logger, slowThreshold time.Duration) func(http.Handler) http.Handler {
	if slowThreshold == 0 {
		slowThreshold = 1 * time.Second
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			if duration > slowThreshold {
				logger.WarnWithFields("Slow HTTP request", map[string]interface{}{
					"method":       r.Method,
					"path":         r.URL.Path,
					"duration_ms":  duration.Milliseconds(),
					"threshold_ms": slowThreshold.Milliseconds(),
					"status_code":  ww.statusCode,
					"ip":           getClientIP(r),
				})
			}
		})
	}
}
