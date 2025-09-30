package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"zpwoot/internal/app"
	"zpwoot/internal/infra/http/handlers"
	"zpwoot/internal/infra/http/middleware"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/config"
	"zpwoot/platform/db"
	"zpwoot/platform/logger"
)

func SetupRoutes(cfg *config.Config, database *db.DB, logger *logger.Logger, wameowManager *wameow.Manager, container *app.Container) http.Handler {
	r := chi.NewRouter()

	setupMiddlewares(r, cfg, container, logger)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	healthHandler := handlers.NewHealthHandler(logger, wameowManager)
	r.Get("/health", healthHandler.GetHealth)
	r.Get("/health/wameow", healthHandler.GetWameowHealth)

	setupSessionRoutes(r, logger, wameowManager, container)

	setupGlobalRoutes(r, logger, container)

	return r
}

func logWameowAvailability(appLogger *logger.Logger, wameowManager *wameow.Manager) {
	if wameowManager != nil {
		appLogger.Info("Wameow manager is available for session routes")
	} else {
		appLogger.Warn("Wameow manager is nil - session functionality will be limited")
	}
}

func setupMiddlewares(r *chi.Mux, cfg *config.Config, container *app.Container, logger *logger.Logger) {
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

	r.Use(middleware.RequestID(logger))

	r.Use(middleware.HTTPLogger(logger))

	r.Use(middleware.Metrics(container, logger))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.APIKeyAuth(cfg, logger))
}

func setupSessionRoutes(r *chi.Mux, appLogger *logger.Logger, wameowManager *wameow.Manager, container *app.Container) {
	logWameowAvailability(appLogger, wameowManager)

	r.Route("/sessions", func(r chi.Router) {
		setupSessionManagementRoutes(r, container, appLogger)
		setupMessageRoutes(r, container, wameowManager, appLogger)
		setupGroupRoutes(r, container, appLogger)
		setupContactRoutes(r, container, appLogger)
		setupNewsletterRoutes(r, container, appLogger)
		setupCommunityRoutes(r, container, appLogger)
		setupWebhookRoutes(r, container, appLogger)
		setupMediaRoutes(r, container, appLogger)
		setupChatwootRoutes(r, container, appLogger)
	})
}

func setupSessionManagementRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	sessionHandler := handlers.NewSessionHandler(appLogger, container.GetSessionUseCase(), container.GetSessionRepository())

	r.Post("/create", sessionHandler.CreateSession)
	r.Get("/list", sessionHandler.ListSessions)
	r.Get("/{sessionId}/info", sessionHandler.GetSessionInfo)
	r.Delete("/{sessionId}/delete", sessionHandler.DeleteSession)
	r.Post("/{sessionId}/connect", sessionHandler.ConnectSession)
	r.Post("/{sessionId}/logout", sessionHandler.LogoutSession)
	r.Get("/{sessionId}/qr", sessionHandler.GetQRCode)
	r.Post("/{sessionId}/pair", sessionHandler.PairPhone)
	r.Post("/{sessionId}/proxy/set", sessionHandler.SetProxy)
	r.Get("/{sessionId}/proxy/find", sessionHandler.GetProxy)

	r.Get("/{sessionId}/stats", sessionHandler.GetSessionStats)
	r.Get("/{sessionId}/user-jid", sessionHandler.GetUserJID)
	r.Get("/{sessionId}/device-info", sessionHandler.GetDeviceInfo)
}

func setupMessageRoutes(r chi.Router, container *app.Container, wameowManager *wameow.Manager, appLogger *logger.Logger) {
	messageHandler := handlers.NewMessageHandler(
		container.GetMessageUseCase(),
		wameowManager,
		container.GetSessionRepository(),
		appLogger,
	)

	r.Route("/{sessionId}/messages", func(r chi.Router) {
		r.Post("/send/text", messageHandler.SendText)
		r.Post("/send/media", messageHandler.SendMedia)
		r.Post("/send/image", messageHandler.SendImage)
		r.Post("/send/audio", messageHandler.SendAudio)
		r.Post("/send/video", messageHandler.SendVideo)
		r.Post("/send/document", messageHandler.SendDocument)
		r.Post("/send/sticker", messageHandler.SendSticker)
		r.Post("/send/location", messageHandler.SendLocation)
		r.Post("/send/contact", messageHandler.SendContact)
		r.Post("/send/contact-list", messageHandler.SendContactList)
		r.Post("/send/contact-list-business", messageHandler.SendContactListBusiness)
		r.Post("/send/single-contact", messageHandler.SendSingleContact)
		r.Post("/send/single-contact-business", messageHandler.SendSingleContactBusiness)
		r.Post("/send/profile/business", messageHandler.SendBusinessProfile)
		r.Post("/send/button", messageHandler.SendButtonMessage)
		r.Post("/send/list", messageHandler.SendListMessage)
		r.Post("/send/reaction", messageHandler.SendReaction)
		r.Post("/send/presence", messageHandler.SendPresence)
		r.Post("/send/poll", messageHandler.SendPoll)
		r.Post("/edit", messageHandler.EditMessage)
		r.Post("/mark-read", messageHandler.MarkAsRead)
		r.Post("/revoke", messageHandler.RevokeMessage)
		r.Get("/poll/{messageId}/results", messageHandler.GetPollResults)
	})
}

func setupGroupRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	groupHandler := handlers.NewGroupHandler(appLogger, container.GetGroupUseCase(), container.GetSessionRepository())

	r.Route("/{sessionId}/groups", func(r chi.Router) {
		r.Post("/", groupHandler.CreateGroup)
		r.Get("/", groupHandler.ListGroups)
		r.Get("/info", groupHandler.GetGroupInfo)
		r.Post("/participants", groupHandler.UpdateGroupParticipants)
		r.Put("/name", groupHandler.SetGroupName)
		r.Put("/description", groupHandler.SetGroupDescription)
		r.Put("/photo", groupHandler.SetGroupPhoto)

		r.Get("/invite-link", groupHandler.GetGroupInviteLink)
		r.Post("/join-via-link", groupHandler.JoinGroupViaLink)
		r.Post("/leave", groupHandler.LeaveGroup)
		r.Put("/settings", groupHandler.UpdateGroupSettings)
		r.Get("/request-participants", groupHandler.GetGroupRequestParticipants)
		r.Post("/request-participants", groupHandler.UpdateGroupRequestParticipants)
		r.Put("/join-approval-mode", groupHandler.SetGroupJoinApprovalMode)
		r.Put("/member-add-mode", groupHandler.SetGroupMemberAddMode)
		r.Get("/info-from-link", groupHandler.GetGroupInfoFromLink)
		r.Post("/info-from-invite", groupHandler.GetGroupInfoFromInvite)
		r.Post("/join-with-invite", groupHandler.JoinGroupWithInvite)
	})
}

func setupContactRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	contactHandler := handlers.NewContactHandler(appLogger, container.GetContactUseCase(), container.GetSessionRepository())

	r.Route("/{sessionId}/contacts", func(r chi.Router) {
		r.Post("/check", contactHandler.CheckWhatsApp)
		r.Get("/avatar", contactHandler.GetProfilePicture)
		r.Post("/info", contactHandler.GetUserInfo)
		r.Get("/", contactHandler.ListContacts)
		r.Post("/sync", contactHandler.SyncContacts)
		r.Get("/business", contactHandler.GetBusinessProfile)

		r.Post("/is-on-whatsapp", contactHandler.IsOnWhatsApp)
		r.Get("/all", contactHandler.GetAllContacts)
		r.Get("/profile-picture-info", contactHandler.GetProfilePictureInfo)
		r.Post("/detailed-info", contactHandler.GetDetailedUserInfo)
	})
}

func setupNewsletterRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	newsletterHandler := handlers.NewNewsletterHandler(appLogger, container.GetNewsletterUseCase(), container.GetSessionRepository())

	r.Route("/{sessionId}/newsletters", func(r chi.Router) {
		r.Post("/create", newsletterHandler.CreateNewsletter)
		r.Get("/info", newsletterHandler.GetNewsletterInfo)
		r.Post("/info-from-invite", newsletterHandler.GetNewsletterInfoWithInvite)
		r.Post("/follow", newsletterHandler.FollowNewsletter)
		r.Post("/unfollow", newsletterHandler.UnfollowNewsletter)
		r.Get("/", newsletterHandler.GetSubscribedNewsletters)
	})
}

func setupCommunityRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	communityHandler := handlers.NewCommunityHandler(appLogger, container.GetCommunityUseCase(), container.GetSessionRepository())

	r.Route("/{sessionId}/communities", func(r chi.Router) {
		r.Post("/link-group", communityHandler.LinkGroup)
		r.Post("/unlink-group", communityHandler.UnlinkGroup)
		r.Get("/info", communityHandler.GetCommunityInfo)
		r.Get("/subgroups", communityHandler.GetSubGroups)
	})
}

func setupWebhookRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	webhookHandler := handlers.NewWebhookHandler(appLogger, container.GetWebhookUseCase(), container.GetSessionRepository())

	r.Route("/{sessionId}/webhook", func(r chi.Router) {
		r.Post("/set", webhookHandler.SetConfig)
		r.Get("/find", webhookHandler.FindConfig)
		r.Post("/test", webhookHandler.TestWebhook)
	})
}

func setupMediaRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	mediaHandler := handlers.NewMediaHandler(appLogger, container.GetMediaUseCase(), container.GetSessionRepository())

	r.Route("/{sessionId}/media", func(r chi.Router) {
		r.Post("/download", mediaHandler.DownloadMedia)
		r.Get("/info", mediaHandler.GetMediaInfo)
		r.Get("/list", mediaHandler.ListCachedMedia)
		r.Post("/clear-cache", mediaHandler.ClearCache)
		r.Get("/stats", mediaHandler.GetMediaStats)
	})
}

func setupChatwootRoutes(r chi.Router, container *app.Container, appLogger *logger.Logger) {
	chatwootHandler := handlers.NewChatwootHandler(
		container.GetChatwootUseCase(),
		container.GetSessionRepository(),
		appLogger,
	)

	r.Route("/{sessionId}/chatwoot", func(r chi.Router) {
		r.Post("/set", chatwootHandler.CreateConfig)
		r.Get("/", chatwootHandler.GetConfig)
		r.Put("/", chatwootHandler.UpdateConfig)
		r.Delete("/", chatwootHandler.DeleteConfig)
		r.Post("/test", chatwootHandler.TestConnection)
		r.Get("/stats", chatwootHandler.GetStats)
		r.Post("/auto-create-inbox", chatwootHandler.AutoCreateInbox)
	})
}

func setupGlobalRoutes(r *chi.Mux, appLogger *logger.Logger, container *app.Container) {
	webhookHandler := handlers.NewWebhookHandler(appLogger, container.GetWebhookUseCase(), container.GetSessionRepository())
	r.Get("/webhook/events", webhookHandler.GetSupportedEvents)

	chatwootHandler := handlers.NewChatwootHandler(
		container.GetChatwootUseCase(),
		container.GetSessionRepository(),
		appLogger,
	)
	r.Post("/chatwoot/webhook/{sessionId}", chatwootHandler.ReceiveWebhook)
}
