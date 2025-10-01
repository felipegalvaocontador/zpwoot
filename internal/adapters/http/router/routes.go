package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"zpwoot/internal/adapters/http/handler"
	"zpwoot/internal/adapters/http/middleware"
	"zpwoot/internal/services"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

func SetupRoutes(cfg *config.Config, logger *logger.Logger, sessionService *services.SessionService, messageService *services.MessageService) http.Handler {
	r := chi.NewRouter()

	setupMiddlewares(r, cfg, logger)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Health check routes
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"zpwoot","version":"2.0.0"}`))
	})

	setupSessionRoutes(r, logger, sessionService, messageService)
	setupGlobalRoutes(r, logger)

	return r
}

func setupMiddlewares(r *chi.Mux, cfg *config.Config, logger *logger.Logger) {
	// Panic recovery middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.ErrorWithFields("Panic recovered", map[string]interface{}{
						"error":  err,
						"path":   r.URL.Path,
						"method": r.Method,
					})
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	// HTTP logging middleware
	r.Use(middleware.HTTPLogger(logger))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API Key authentication middleware
	r.Use(middleware.APIKeyAuth(cfg, logger))
}

func setupSessionRoutes(r *chi.Mux, appLogger *logger.Logger, sessionService *services.SessionService, messageService *services.MessageService) {
	r.Route("/sessions", func(r chi.Router) {
		setupSessionManagementRoutes(r, sessionService, appLogger)
		setupMessageRoutes(r, messageService, sessionService, appLogger)
		setupGroupRoutes(r, appLogger)
		setupContactRoutes(r, appLogger)
		setupWebhookRoutes(r, appLogger)
		setupMediaRoutes(r, appLogger)
		setupChatwootRoutes(r, appLogger)
	})
}

func setupSessionManagementRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	sessionHandler := handler.NewSessionHandler(sessionService, appLogger)

	r.Post("/create", sessionHandler.CreateSession)
	r.Get("/list", sessionHandler.ListSessions)
	r.Get("/{sessionId}/info", sessionHandler.GetSessionInfo)
	r.Delete("/{sessionId}/delete", sessionHandler.DeleteSession)
	r.Post("/{sessionId}/connect", sessionHandler.ConnectSession)
	// r.Post("/{sessionId}/logout", sessionHandler.LogoutSession) // TODO: Implementar
	r.Get("/{sessionId}/qr", sessionHandler.GetQRCode)
	// r.Post("/{sessionId}/pair", sessionHandler.PairPhone) // TODO: Implementar
	// r.Post("/{sessionId}/proxy/set", sessionHandler.SetProxy) // TODO: Implementar
	// r.Get("/{sessionId}/proxy/find", sessionHandler.GetProxy) // TODO: Implementar

	r.Get("/{sessionId}/stats", sessionHandler.GetSessionStats)
	// r.Get("/{sessionId}/user-jid", sessionHandler.GetUserJID) // TODO: Implementar
	// r.Get("/{sessionId}/device-info", sessionHandler.GetDeviceInfo) // TODO: Implementar
}

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
		r.Post("/send/image", messageHandler.SendMediaMessage)
		r.Post("/send/audio", messageHandler.SendMediaMessage)
		r.Post("/send/video", messageHandler.SendMediaMessage)
		r.Post("/send/document", messageHandler.SendMediaMessage)
		r.Post("/send/sticker", messageHandler.SendMediaMessage)
		r.Post("/send/location", messageHandler.SendTextMessage) // TODO: Implementar location
		r.Post("/send/contact", messageHandler.SendTextMessage)  // TODO: Implementar contact
		r.Post("/send/contact-list", messageHandler.SendTextMessage) // TODO: Implementar contact list
		r.Post("/send/button", messageHandler.SendTextMessage)   // TODO: Implementar button
		r.Post("/send/list", messageHandler.SendTextMessage)     // TODO: Implementar list
		r.Post("/send/reaction", messageHandler.SendTextMessage) // TODO: Implementar reaction
		r.Post("/send/presence", messageHandler.SendTextMessage) // TODO: Implementar presence
		r.Post("/send/poll", messageHandler.SendTextMessage)     // TODO: Implementar poll
		r.Post("/edit", messageHandler.SendTextMessage)          // TODO: Implementar edit
		r.Post("/mark-read", messageHandler.MarkAsRead)
		r.Post("/revoke", messageHandler.SendTextMessage)        // TODO: Implementar revoke
		r.Get("/poll/{messageId}/results", messageHandler.SendTextMessage) // TODO: Implementar poll results
	})
}

func setupGroupRoutes(r chi.Router, appLogger *logger.Logger) {
	// TODO: Implementar GroupHandler
	r.Route("/{sessionId}/groups", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Group routes - Ready for implementation"}`))
		})
	})
}

func setupContactRoutes(r chi.Router, appLogger *logger.Logger) {
	// TODO: Implementar ContactHandler
	r.Route("/{sessionId}/contacts", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Contact routes - Ready for implementation"}`))
		})
	})
}

func setupWebhookRoutes(r chi.Router, appLogger *logger.Logger) {
	// TODO: Implementar WebhookHandler
	r.Route("/{sessionId}/webhook", func(r chi.Router) {
		r.Post("/set", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Webhook routes - Ready for implementation"}`))
		})
	})
}

func setupMediaRoutes(r chi.Router, appLogger *logger.Logger) {
	// TODO: Implementar MediaHandler
	r.Route("/{sessionId}/media", func(r chi.Router) {
		r.Post("/download", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Media routes - Ready for implementation"}`))
		})
	})
}

func setupChatwootRoutes(r chi.Router, appLogger *logger.Logger) {
	// TODO: Implementar ChatwootHandler
	r.Route("/{sessionId}/chatwoot", func(r chi.Router) {
		r.Post("/set", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Chatwoot routes - Ready for implementation"}`))
		})
	})
}

func setupGlobalRoutes(r *chi.Mux, appLogger *logger.Logger) {
	// Global webhook events endpoint
	r.Get("/webhook/events", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"events":["message","session","contact","group"]}`))
	})

	// Global chatwoot webhook endpoint
	r.Post("/chatwoot/webhook/{sessionId}", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Chatwoot webhook received"}`))
	})
}
