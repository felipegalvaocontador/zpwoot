package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

func setupContactRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {

	contactHandler := handler.NewContactHandler(nil, sessionService, appLogger)

	r.Route("/{sessionId}/contacts", func(r chi.Router) {

		r.Post("/check", contactHandler.CheckWhatsApp)
		r.Post("/is-on-whatsapp", contactHandler.IsOnWhatsApp)

		r.Get("/avatar", contactHandler.GetProfilePicture)
		r.Post("/info", contactHandler.GetUserInfo)
		r.Get("/profile-picture-info", contactHandler.GetProfilePictureInfo)
		r.Post("/detailed-info", contactHandler.GetDetailedUserInfo)

		r.Get("/", contactHandler.ListContacts)
		r.Get("/all", contactHandler.GetAllContacts)

		r.Post("/sync", contactHandler.SyncContacts)

		r.Get("/business", contactHandler.GetBusinessProfile)
	})
}
