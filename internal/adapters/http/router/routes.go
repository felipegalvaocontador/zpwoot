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
		setupContactRoutes(r, sessionService, appLogger)
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

func setupContactRoutes(r chi.Router, sessionService *services.SessionService, appLogger *logger.Logger) {
	// TODO: Criar ContactService quando necessário
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
