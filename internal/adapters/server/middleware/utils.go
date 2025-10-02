package middleware

import (
	"net/http"
	"strings"
)


func getClientIP(r *http.Request) string {

	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP",
	}

	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {

			if strings.Contains(ip, ",") {
				ip = strings.TrimSpace(strings.Split(ip, ",")[0])
			}
			return ip
		}
	}


	return r.RemoteAddr
}
