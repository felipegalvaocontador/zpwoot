package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

func setupMediaRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	mediaHandler := handler.NewMediaHandler(sessionService, appLogger)

	r.Route("/{sessionId}/media", func(r chi.Router) {

		r.Post("/download", mediaHandler.DownloadMedia)
		r.Get("/info", mediaHandler.GetMediaInfo)
		r.Get("/list", mediaHandler.ListCachedMedia)

		r.Post("/clear-cache", mediaHandler.ClearCache)

		r.Get("/stats", mediaHandler.GetStats)
	})
}
