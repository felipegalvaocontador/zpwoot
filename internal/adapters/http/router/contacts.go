package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/http/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// setupContactRoutes configura todas as rotas relacionadas a contatos
func setupContactRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	// TODO: Integrar ContactService quando disponível
	contactHandler := handler.NewContactHandler(nil, sessionService, appLogger)

	r.Route("/{sessionId}/contacts", func(r chi.Router) {
		// Contact verification
		r.Post("/check", contactHandler.CheckWhatsApp)
		r.Post("/is-on-whatsapp", contactHandler.IsOnWhatsApp)

		// Profile and info
		r.Get("/avatar", contactHandler.GetProfilePicture)
		r.Post("/info", contactHandler.GetUserInfo)
		r.Get("/profile-picture-info", contactHandler.GetProfilePictureInfo)
		r.Post("/detailed-info", contactHandler.GetDetailedUserInfo)

		// Contact listing
		r.Get("/", contactHandler.ListContacts)
		r.Get("/all", contactHandler.GetAllContacts)

		// Contact sync
		r.Post("/sync", contactHandler.SyncContacts)

		// Business profiles
		r.Get("/business", contactHandler.GetBusinessProfile)
	})
}
