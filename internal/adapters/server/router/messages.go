package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// setupMessageRoutes configura todas as rotas relacionadas a mensagens
func setupMessageRoutes(r chi.Router, messageService *services.MessageService, sessionService *services.SessionService, appLogger *logger.Logger) {
	messageHandler := handler.NewMessageHandler(
		messageService,
		sessionService,
		appLogger,
	)

	r.Route("/{sessionId}/messages", func(r chi.Router) {
		// CRUD operations
		r.Post("/", messageHandler.CreateMessage)
		r.Get("/", messageHandler.ListMessages)
		r.Get("/{messageId}", messageHandler.GetMessage)
		r.Delete("/{messageId}", messageHandler.DeleteMessage)

		// Sync operations
		r.Put("/{messageId}/sync", messageHandler.UpdateSyncStatus)
		r.Get("/pending-sync", messageHandler.GetPendingSyncMessages)

		// Statistics
		r.Get("/stats", messageHandler.GetMessageStats)

		// Messaging operations (seguindo padrão legacy)
		r.Post("/send/text", messageHandler.SendTextMessage)
		r.Post("/send/media", messageHandler.SendMediaMessage)

		// Specific media endpoints
		r.Post("/send/image", messageHandler.SendImage)
		r.Post("/send/audio", messageHandler.SendAudio)
		r.Post("/send/video", messageHandler.SendVideo)
		r.Post("/send/document", messageHandler.SendDocument)
		r.Post("/send/sticker", messageHandler.SendSticker)

		// Location and contact endpoints
		r.Post("/send/location", messageHandler.SendLocation)
		r.Post("/send/contact", messageHandler.SendContact)
		r.Post("/send/contact-list", messageHandler.SendContactList)

		// Interactive message endpoints
		r.Post("/send/button", messageHandler.SendButton)
		r.Post("/send/list", messageHandler.SendList)
		r.Post("/send/poll", messageHandler.SendPoll)

		// Action endpoints
		r.Post("/send/reaction", messageHandler.SendReaction)
		r.Post("/send/presence", messageHandler.SendPresence)

		// Business profile endpoint
		r.Post("/send/profile/business", messageHandler.SendBusinessProfile)

		// Message management endpoints
		r.Post("/edit", messageHandler.EditMessage)
		r.Post("/revoke", messageHandler.RevokeMessage)
		r.Post("/mark-read", messageHandler.MarkAsRead)

		// Poll results endpoint
		r.Get("/poll/{messageId}/results", messageHandler.GetPollResults)
	})
}
