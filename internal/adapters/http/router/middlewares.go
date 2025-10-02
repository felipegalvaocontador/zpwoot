package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"zpwoot/internal/adapters/http/middleware"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

// setupMiddlewares configura todos os middlewares globais da aplicação
func setupMiddlewares(r *chi.Mux, cfg *config.Config, logger *logger.Logger) {
	// Panic recovery middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.ErrorWithFields("Panic recovered", map[string]interface{}{
						"error":  err,
						"path":   r.URL.Path,
						"method": r.Method,
					})
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	// HTTP logging middleware
	r.Use(middleware.HTTPLogger(logger))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API Key authentication middleware
	r.Use(middleware.APIKeyAuth(cfg, logger))
}
