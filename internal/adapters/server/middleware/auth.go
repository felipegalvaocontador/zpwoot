package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

type contextKey string

const (
	apiKeyContextKey        contextKey = "api_key"
	authenticatedContextKey contextKey = "authenticated"
)

func APIKeyAuth(cfg *config.Config, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if isPublicRoute(path) {
				next.ServeHTTP(w, r)
				return
			}

			apiKey := extractAPIKey(r)
			if apiKey == "" {
				log.WarnWithFields("Missing API key", map[string]interface{}{
					"path":   path,
					"method": r.Method,
					"ip":     getClientIP(r),
				})

				writeUnauthorizedResponse(w, "API key is required. Provide it via Authorization header or X-API-Key header", "MISSING_API_KEY")
				return
			}

			if !isValidAPIKey(apiKey, cfg) {
				log.WarnWithFields("Invalid API key", map[string]interface{}{
					"path":    path,
					"method":  r.Method,
					"ip":      getClientIP(r),
					"api_key": maskAPIKey(apiKey),
				})

				writeUnauthorizedResponse(w, "Invalid API key", "INVALID_API_KEY")
				return
			}

			log.DebugWithFields("API key authenticated", map[string]interface{}{
				"path":    path,
				"method":  r.Method,
				"ip":      getClientIP(r),
				"api_key": maskAPIKey(apiKey),
			})

			ctx := context.WithValue(r.Context(), apiKeyContextKey, apiKey)
			ctx = context.WithValue(ctx, authenticatedContextKey, true)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/health",
		"/swagger",
		"/chatwoot/webhook",
	}

	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}

	return false
}

func extractAPIKey(r *http.Request) string {

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {

		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}

	return r.Header.Get("X-API-Key")
}

func isValidAPIKey(apiKey string, cfg *config.Config) bool {

	if cfg.Security.APIKey != "" && apiKey == cfg.Security.APIKey {
		return true
	}

	return false
}

func writeUnauthorizedResponse(w http.ResponseWriter, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := shared.ErrorResponse{
		Success: false,
		Error:   "Unauthorized",
		Code:    code,
		Details: message,
	}

	json.NewEncoder(w).Encode(response)
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}

	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}
