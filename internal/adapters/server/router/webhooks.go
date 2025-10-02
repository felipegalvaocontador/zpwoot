package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

func setupWebhookRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	webhookHandler := handler.NewWebhookHandler(sessionService, appLogger)

	r.Route("/{sessionId}/webhook", func(r chi.Router) {

		r.Post("/set", webhookHandler.SetConfig)
		r.Get("/find", webhookHandler.FindConfig)

		r.Post("/test", webhookHandler.TestWebhook)
	})
}
