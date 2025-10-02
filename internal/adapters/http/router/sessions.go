package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/http/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// setupSessionRoutes configura todas as rotas relacionadas a sessões
func setupSessionRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	sessionHandler := handler.NewSessionHandler(sessionService, appLogger)

	// Session management
	r.Post("/create", sessionHandler.CreateSession)
	r.Get("/list", sessionHandler.ListSessions)
	r.Get("/{sessionId}/info", sessionHandler.GetSessionInfo)
	r.Delete("/{sessionId}/delete", sessionHandler.DeleteSession)

	// Session connection
	r.Post("/{sessionId}/connect", sessionHandler.ConnectSession)
	r.Post("/{sessionId}/logout", sessionHandler.LogoutSession)
	r.Get("/{sessionId}/qr", sessionHandler.GetQRCode)
	r.Post("/{sessionId}/pair", sessionHandler.PairPhone)

	// Proxy configuration
	r.Post("/{sessionId}/proxy/set", sessionHandler.SetProxy)
	r.Get("/{sessionId}/proxy/find", sessionHandler.GetProxy)

	// Session statistics
	r.Get("/{sessionId}/stats", sessionHandler.GetSessionStats)
}
