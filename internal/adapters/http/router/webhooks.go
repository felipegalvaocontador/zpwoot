package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/http/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// setupWebhookRoutes configura todas as rotas relacionadas a webhooks
func setupWebhookRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	webhookHandler := handler.NewWebhookHandler(sessionService, appLogger)

	r.Route("/{sessionId}/webhook", func(r chi.Router) {
		// Webhook configuration
		r.Post("/set", webhookHandler.SetConfig)
		r.Get("/find", webhookHandler.FindConfig)

		// Testing
		r.Post("/test", webhookHandler.TestWebhook)
	})
}
