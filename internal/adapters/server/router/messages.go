package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)


func setupMessageRoutes(r chi.Router, messageService *services.MessageService, sessionService *services.SessionService, appLogger *logger.Logger) {
	messageHandler := handler.NewMessageHandler(
		messageService,
		sessionService,
		appLogger,
	)

	r.Route("/{sessionId}/messages", func(r chi.Router) {



		r.Post("/send/text", messageHandler.SendTextMessage)
		r.Post("/send/media", messageHandler.SendMediaMessage)


		r.Post("/send/image", messageHandler.SendImage)
		r.Post("/send/audio", messageHandler.SendAudio)
		r.Post("/send/video", messageHandler.SendVideo)
		r.Post("/send/document", messageHandler.SendDocument)
		r.Post("/send/sticker", messageHandler.SendSticker)


		r.Post("/send/location", messageHandler.SendLocation)
		r.Post("/send/contact", messageHandler.SendContact)
		r.Post("/send/contact-list", messageHandler.SendContactList)


		r.Post("/send/button", messageHandler.SendButton)
		r.Post("/send/list", messageHandler.SendList)
		r.Post("/send/poll", messageHandler.SendPoll)


		r.Post("/send/reaction", messageHandler.SendReaction)
		r.Post("/send/presence", messageHandler.SendPresence)


		r.Post("/send/profile/business", messageHandler.SendBusinessProfile)


		r.Post("/edit", messageHandler.EditMessage)
		r.Post("/revoke", messageHandler.RevokeMessage)
		r.Post("/mark-read", messageHandler.MarkAsRead)


		r.Get("/poll/{messageId}/results", messageHandler.GetPollResults)
	})
}
