package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// setupMediaRoutes configura todas as rotas relacionadas a mídia
func setupMediaRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	mediaHandler := handler.NewMediaHandler(sessionService, appLogger)

	r.Route("/{sessionId}/media", func(r chi.Router) {
		// Media download and management
		r.Post("/download", mediaHandler.DownloadMedia)
		r.Get("/info", mediaHandler.GetMediaInfo)
		r.Get("/list", mediaHandler.ListCachedMedia)

		// Cache management
		r.Post("/clear-cache", mediaHandler.ClearCache)

		// Statistics
		r.Get("/stats", mediaHandler.GetStats)
	})
}
