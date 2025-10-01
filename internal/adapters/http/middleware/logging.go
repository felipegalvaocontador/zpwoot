package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"zpwoot/platform/logger"
)

// responseWriter wrapper para capturar status code e tamanho da resposta
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captura status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captura tamanho da resposta
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// HTTPLogger middleware para logging de requisições HTTP
func HTTPLogger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrapper para capturar informações da resposta
			ww := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // padrão
			}

			// Executar próximo handler
			next.ServeHTTP(ww, r)

			// Calcular duração
			duration := time.Since(start)

			// Preparar campos para log
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

			// Adicionar request ID se presente
			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				fields["request_id"] = requestID
			}

			// Adicionar referer se presente
			if referer := r.Header.Get("Referer"); referer != "" {
				fields["referer"] = referer
			}

			// Determinar nível de log baseado no status code
			message := "HTTP request processed"
			switch {
			case ww.statusCode >= 500:
				logger.ErrorWithFields(message, fields)
			case ww.statusCode >= 400:
				logger.WarnWithFields(message, fields)
			case ww.statusCode >= 300:
				logger.InfoWithFields(message, fields)
			default:
				// Para requisições de health check, usar debug
				if r.URL.Path == "/health" {
					logger.DebugWithFields(message, fields)
				} else {
					logger.InfoWithFields(message, fields)
				}
			}
		})
	}
}

// ErrorLogger middleware para capturar e logar panics
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

					// Retornar erro 500
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// PerformanceLogger middleware para logging de performance
func PerformanceLogger(logger *logger.Logger, slowThreshold time.Duration) func(http.Handler) http.Handler {
	if slowThreshold == 0 {
		slowThreshold = 1 * time.Second // padrão: 1 segundo
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

			// Log apenas se for lento
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