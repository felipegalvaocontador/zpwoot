package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

func setupChatwootRoutes(r chi.Router, messageService *services.MessageService, sessionService *services.SessionService, appLogger *logger.Logger) {
	chatwootHandler := handler.NewChatwootHandler(messageService, sessionService, appLogger)

	r.Route("/{sessionName}/chatwoot", func(r chi.Router) {

		r.Post("/set", chatwootHandler.CreateConfig)
		r.Get("/find", chatwootHandler.FindConfig)

		r.Get("/", chatwootHandler.FindConfig)
		r.Put("/", chatwootHandler.UpdateConfig)
		r.Delete("/", chatwootHandler.DeleteConfig)
		r.Post("/test", chatwootHandler.TestConnection)
		r.Post("/auto-create-inbox", chatwootHandler.AutoCreateInbox)
		r.Get("/stats", chatwootHandler.GetStats)
	})
}
