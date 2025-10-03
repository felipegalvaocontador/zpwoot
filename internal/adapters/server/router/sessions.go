package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

func setupSessionRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	sessionHandler := handler.NewSessionHandler(sessionService, appLogger)

	// Session management routes
	r.Post("/create", sessionHandler.CreateSession)
	r.Get("/list", sessionHandler.ListSessions)

	// Session-specific routes using session name (e.g., /sessions/my-session/info)
	r.Get("/{sessionName}/info", sessionHandler.GetSessionInfo)
	r.Delete("/{sessionName}/delete", sessionHandler.DeleteSession)

	// Connection management
	r.Post("/{sessionName}/connect", sessionHandler.ConnectSession)
	r.Post("/{sessionName}/logout", sessionHandler.LogoutSession)
	r.Get("/{sessionName}/qr", sessionHandler.GetQRCode)
	r.Post("/{sessionName}/pair", sessionHandler.PairPhone)

	// Proxy configuration
	r.Post("/{sessionName}/proxy/set", sessionHandler.SetProxy)
	r.Get("/{sessionName}/proxy/find", sessionHandler.GetProxy)

	// Statistics
	r.Get("/{sessionName}/stats", sessionHandler.GetSessionStats)
}
