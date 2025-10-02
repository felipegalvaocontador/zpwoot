package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/services"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(cfg *config.Config, logger *logger.Logger, sessionService *services.SessionService, messageService *services.MessageService, groupService *services.GroupService) http.Handler {
	r := chi.NewRouter()

	// Configurar middlewares globais
	setupMiddlewares(r, cfg, logger)

	// Swagger documentation
	setupSwaggerRoutes(r)

	// Health check
	setupHealthRoutes(r)

	// Setup all route groups
	setupAllRoutes(r, logger, sessionService, messageService, groupService)

	return r
}

// setupAllRoutes configura todos os grupos de rotas
func setupAllRoutes(r *chi.Mux, appLogger *logger.Logger, sessionService *services.SessionService, messageService *services.MessageService, groupService *services.GroupService) {
	r.Route("/sessions", func(r chi.Router) {
		// Session management routes
		setupSessionRoutes(r, sessionService, appLogger)

		// Message routes
		setupMessageRoutes(r, messageService, sessionService, appLogger)

		// Group routes
		setupGroupRoutes(r, groupService, sessionService, appLogger)

		// Contact routes
		setupContactRoutes(r, sessionService, appLogger)

		// Webhook routes
		setupWebhookRoutes(r, appLogger)

		// Media routes
		setupMediaRoutes(r, appLogger)

		// Chatwoot routes
		setupChatwootRoutes(r, appLogger)
	})

	// Global routes
	setupGlobalRoutes(r, appLogger)
}

// setupHealthRoutes configura rotas de health check
func setupHealthRoutes(r *chi.Mux) {
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"zpwoot","version":"2.0.0"}`))
	})
}

// setupGlobalRoutes configura rotas globais
func setupGlobalRoutes(r *chi.Mux, appLogger *logger.Logger) {
	// Global webhook events endpoint
	r.Get("/webhook/events", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"events":["message","session","contact","group"]}`))
	})

	// Global chatwoot webhook endpoint
	r.Post("/chatwoot/webhook/{sessionId}", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Chatwoot webhook received"}`))
	})
}
