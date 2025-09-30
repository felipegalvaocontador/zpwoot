package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

// APIKeyAuth middleware for Chi router
func APIKeyAuth(cfg *config.Config, logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			
			// Skip authentication for certain paths
			if strings.HasPrefix(path, "/health") || 
			   strings.HasPrefix(path, "/swagger") || 
			   strings.Contains(path, "/chatwoot/webhook") {
				next.ServeHTTP(w, r)
				return
			}

			apiKey := r.Header.Get("Authorization")
			if apiKey == "" {
				apiKey = r.Header.Get("X-API-Key")
			}

			if apiKey == "" {
				logger.WarnWithFields("Missing API key", map[string]interface{}{
					"path":   path,
					"method": r.Method,
					"ip":     getClientIP(r),
				})
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Unauthorized",
					"message": "API key is required. Provide it via Authorization header or X-API-Key header",
					"code":    "MISSING_API_KEY",
				})
				return
			}

			if apiKey != cfg.GlobalAPIKey {
				logger.WarnWithFields("Invalid API key", map[string]interface{}{
					"path":    path,
					"method":  r.Method,
					"ip":      getClientIP(r),
					"api_key": maskAPIKey(apiKey),
				})
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Unauthorized",
					"message": "Invalid API key",
					"code":    "INVALID_API_KEY",
				})
				return
			}

			logger.DebugWithFields("API key authenticated", map[string]interface{}{
				"path":    path,
				"method":  r.Method,
				"ip":      getClientIP(r),
				"api_key": maskAPIKey(apiKey),
			})

			// Add authentication info to context
			ctx := context.WithValue(r.Context(), "api_key", apiKey)
			ctx = context.WithValue(ctx, "authenticated", true)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAPIKeyFromContext extracts API key from Chi context
func GetAPIKeyFromContext(r *http.Request) string {
	if apiKey, ok := r.Context().Value("api_key").(string); ok {
		return apiKey
	}
	return ""
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(r *http.Request) bool {
	if authenticated, ok := r.Context().Value("authenticated").(bool); ok {
		return authenticated
	}
	return false
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// maskAPIKey masks an API key for logging purposes
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}
