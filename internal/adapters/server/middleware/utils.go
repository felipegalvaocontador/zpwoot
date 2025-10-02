package middleware

import (
	"net/http"
	"strings"
)

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
