package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"zpwoot/internal/adapters/server/middleware"
	"zpwoot/internal/services"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

func SetupRoutes(cfg *config.Config, logger *logger.Logger, sessionService *services.SessionService, messageService *services.MessageService, groupService *services.GroupService) http.Handler {
	r := chi.NewRouter()

	setupMiddlewares(r, cfg, logger)

	setupSwaggerRoutes(r)

	setupHealthRoutes(r)

	setupAllRoutes(r, logger, sessionService, messageService, groupService)

	return r
}

func setupAllRoutes(r *chi.Mux, appLogger *logger.Logger, sessionService *services.SessionService, messageService *services.MessageService, groupService *services.GroupService) {
	r.Route("/sessions", func(r chi.Router) {

		setupSessionRoutes(r, sessionService, appLogger)

		setupMessageRoutes(r, messageService, sessionService, appLogger)

		setupGroupRoutes(r, groupService, sessionService, appLogger)

		setupContactRoutes(r, sessionService, appLogger)

		setupWebhookRoutes(r, sessionService, appLogger)

		setupMediaRoutes(r, sessionService, appLogger)

		setupChatwootRoutes(r, messageService, sessionService, appLogger)
	})

	setupGlobalRoutes(r, appLogger)
}

func setupHealthRoutes(r *chi.Mux) {
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"zpwoot","version":"2.0.0"}`))
	})
}

func setupGlobalRoutes(r *chi.Mux, appLogger *logger.Logger) {

	r.Get("/webhook/events", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"events":["message","session","contact","group"]}`))
	})

}

func setupMiddlewares(r *chi.Mux, cfg *config.Config, logger *logger.Logger) {

	r.Use(middleware.ErrorLogger(logger))

	r.Use(middleware.HTTPLogger(logger))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.APIKeyAuth(cfg, logger))
}
