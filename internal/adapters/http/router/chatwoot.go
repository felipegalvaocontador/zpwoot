package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/http/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// setupChatwootRoutes configura todas as rotas relacionadas ao Chatwoot
func setupChatwootRoutes(r chi.Router, messageService *services.MessageService, sessionService *services.SessionService, appLogger *logger.Logger) {
	chatwootHandler := handler.NewChatwootHandler(messageService, sessionService, appLogger)

	r.Route("/{sessionId}/chatwoot", func(r chi.Router) {
		// Configuration management
		r.Post("/set", chatwootHandler.CreateConfig)
		r.Get("/", chatwootHandler.GetConfig)
		r.Put("/", chatwootHandler.UpdateConfig)
		r.Delete("/", chatwootHandler.DeleteConfig)

		// Connection and testing
		r.Post("/test", chatwootHandler.TestConnection)
		r.Post("/auto-create-inbox", chatwootHandler.AutoCreateInbox)

		// Statistics
		r.Get("/stats", chatwootHandler.GetStats)
	})
}
