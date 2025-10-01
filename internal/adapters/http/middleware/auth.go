package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"zpwoot/internal/adapters/http/shared"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

type contextKey string

const (
	apiKeyContextKey        contextKey = "api_key"
	authenticatedContextKey contextKey = "authenticated"
)

// APIKeyAuth middleware para autenticação via API key
func APIKeyAuth(cfg *config.Config, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Pular autenticação para rotas públicas
			if isPublicRoute(path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extrair API key dos headers
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

			// Validar API key
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

			// Log autenticação bem-sucedida
			log.DebugWithFields("API key authenticated", map[string]interface{}{
				"path":    path,
				"method":  r.Method,
				"ip":      getClientIP(r),
				"api_key": maskAPIKey(apiKey),
			})

			// Adicionar informações ao contexto
			ctx := context.WithValue(r.Context(), apiKeyContextKey, apiKey)
			ctx = context.WithValue(ctx, authenticatedContextKey, true)

			// Continuar para próximo handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isPublicRoute verifica se a rota é pública (não requer autenticação)
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

// extractAPIKey extrai API key dos headers da requisição
func extractAPIKey(r *http.Request) string {
	// Tentar Authorization header primeiro
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Suportar formato "Bearer <token>" e token direto
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}

	// Tentar X-API-Key header
	return r.Header.Get("X-API-Key")
}

// isValidAPIKey valida se API key é válida
func isValidAPIKey(apiKey string, cfg *config.Config) bool {
	// Validar contra API key global
	if cfg.Security.APIKey != "" && apiKey == cfg.Security.APIKey {
		return true
	}

	// TODO: Implementar validação de múltiplas API keys se necessário
	// Por enquanto, apenas validar contra a global

	return false
}

// writeUnauthorizedResponse escreve resposta de não autorizado
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

// maskAPIKey mascara API key para logs (mostra apenas primeiros e últimos caracteres)
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}

	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

// getClientIP extrai IP do cliente considerando proxies
func getClientIP(r *http.Request) string {
	// Verificar headers de proxy em ordem de prioridade
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP", // Cloudflare
	}

	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {
			// X-Forwarded-For pode conter múltiplos IPs separados por vírgula
			if strings.Contains(ip, ",") {
				ip = strings.TrimSpace(strings.Split(ip, ",")[0])
			}
			return ip
		}
	}

	// Fallback para RemoteAddr
	return r.RemoteAddr
}