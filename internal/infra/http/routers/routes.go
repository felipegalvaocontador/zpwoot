package routers

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"zpwoot/internal/app"
	"zpwoot/internal/infra/http/handlers"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/db"
	"zpwoot/platform/logger"
)

func SetupRoutes(app *fiber.App, database *db.DB, logger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health check endpoints
	healthHandler := handlers.NewHealthHandler(logger, WameowManager)
	app.Get("/health", healthHandler.GetHealth)
	app.Get("/health/wameow", healthHandler.GetWameowHealth)

	setupSessionRoutes(app, logger, WameowManager, container)

	setupSessionSpecificRoutes(app, database, logger, WameowManager, container)

	setupGlobalRoutes(app, database, logger, WameowManager, container)
}

func setupSessionRoutes(app *fiber.App, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	logWameowAvailability(appLogger, WameowManager)

	sessions := app.Group("/sessions")

	// Setup all route groups
	setupSessionManagementRoutes(sessions, container, appLogger)
	setupMessageRoutes(sessions, container, WameowManager, appLogger)
	setupGroupRoutes(sessions, container, appLogger)
	setupNewsletterRoutes(sessions, container, appLogger)
	setupCommunityRoutes(sessions, container, appLogger)
	setupContactRoutes(sessions, container, appLogger)
	setupWebhookRoutes(sessions, container, appLogger)
	setupChatwootRoutes(sessions, container, appLogger)
}

// logWameowAvailability logs Wameow manager availability
func logWameowAvailability(appLogger *logger.Logger, WameowManager *wameow.Manager) {
	if WameowManager != nil {
		appLogger.Info("Wameow manager is available for session routes")
	} else {
		appLogger.Warn("Wameow manager is nil - session functionality will be limited")
	}
}

// setupSessionManagementRoutes sets up session management routes
func setupSessionManagementRoutes(sessions fiber.Router, container *app.Container, appLogger *logger.Logger) {
	sessionHandler := handlers.NewSessionHandler(appLogger, container.GetSessionUseCase(), container.GetSessionRepository())

	sessions.Post("/create", sessionHandler.CreateSession)
	sessions.Get("/list", sessionHandler.ListSessions)
	sessions.Get("/:sessionId/info", sessionHandler.GetSessionInfo)
	sessions.Delete("/:sessionId/delete", sessionHandler.DeleteSession)
	sessions.Post("/:sessionId/connect", sessionHandler.ConnectSession)
	sessions.Post("/:sessionId/logout", sessionHandler.LogoutSession)
	sessions.Get("/:sessionId/qr", sessionHandler.GetQRCode)
	sessions.Post("/:sessionId/pair", sessionHandler.PairPhone)
	sessions.Post("/:sessionId/proxy/set", sessionHandler.SetProxy)
	sessions.Get("/:sessionId/proxy/find", sessionHandler.GetProxy)
}

// setupMessageRoutes sets up message-related routes
func setupMessageRoutes(sessions fiber.Router, container *app.Container, WameowManager *wameow.Manager, appLogger *logger.Logger) {
	messageHandler := handlers.NewMessageHandler(container.GetMessageUseCase(), WameowManager, container.GetSessionRepository(), appLogger)

	// Basic message sending
	sessions.Post("/:sessionId/messages/send/text", messageHandler.SendText)
	sessions.Post("/:sessionId/messages/send/media", messageHandler.SendMedia)
	sessions.Post("/:sessionId/messages/send/image", messageHandler.SendImage)
	sessions.Post("/:sessionId/messages/send/audio", messageHandler.SendAudio)
	sessions.Post("/:sessionId/messages/send/video", messageHandler.SendVideo)
	sessions.Post("/:sessionId/messages/send/document", messageHandler.SendDocument)
	sessions.Post("/:sessionId/messages/send/sticker", messageHandler.SendSticker)
	sessions.Post("/:sessionId/messages/send/button", messageHandler.SendButtonMessage)
	sessions.Post("/:sessionId/messages/send/contact", messageHandler.SendContact)
	sessions.Post("/:sessionId/messages/send/list", messageHandler.SendListMessage)
	sessions.Post("/:sessionId/messages/send/location", messageHandler.SendLocation)
	sessions.Post("/:sessionId/messages/send/poll", messageHandler.SendPoll)
	sessions.Post("/:sessionId/messages/send/reaction", messageHandler.SendReaction)
	sessions.Post("/:sessionId/messages/send/presence", messageHandler.SendPresence)

	// Message operations
	sessions.Post("/:sessionId/messages/edit", messageHandler.EditMessage)
	sessions.Post("/:sessionId/messages/mark-read", messageHandler.MarkAsRead)
	sessions.Post("/:sessionId/messages/revoke", messageHandler.RevokeMessage)
	sessions.Get("/:sessionId/messages/poll/:messageId/results", messageHandler.GetPollResults)
}

// setupGroupRoutes sets up group management routes
func setupGroupRoutes(sessions fiber.Router, container *app.Container, appLogger *logger.Logger) {
	groupHandler := handlers.NewGroupHandler(appLogger, container.GetGroupUseCase(), container.GetSessionRepository())

	// Basic group operations
	sessions.Post("/:sessionId/groups/create", groupHandler.CreateGroup)
	sessions.Get("/:sessionId/groups", groupHandler.ListGroups)
	sessions.Get("/:sessionId/groups/info", groupHandler.GetGroupInfo)
	sessions.Post("/:sessionId/groups/participants", groupHandler.UpdateGroupParticipants)
	sessions.Put("/:sessionId/groups/name", groupHandler.SetGroupName)
	sessions.Put("/:sessionId/groups/description", groupHandler.SetGroupDescription)
	sessions.Put("/:sessionId/groups/photo", groupHandler.SetGroupPhoto)
	sessions.Get("/:sessionId/groups/invite-link", groupHandler.GetGroupInviteLink)
	sessions.Post("/:sessionId/groups/join", groupHandler.JoinGroup)
	sessions.Post("/:sessionId/groups/leave", groupHandler.LeaveGroup)
	sessions.Put("/:sessionId/groups/settings", groupHandler.UpdateGroupSettings)

	// Group request management
	sessions.Get("/:sessionId/groups/requests", groupHandler.GetGroupRequestParticipants)
	sessions.Post("/:sessionId/groups/requests", groupHandler.UpdateGroupRequestParticipants)
	sessions.Put("/:sessionId/groups/join-approval", groupHandler.SetGroupJoinApprovalMode)
	sessions.Put("/:sessionId/groups/member-add-mode", groupHandler.SetGroupMemberAddMode)

	// Advanced group operations
	sessions.Get("/:sessionId/groups/info-from-link", groupHandler.GetGroupInfoFromLink)
	sessions.Post("/:sessionId/groups/info-from-invite", groupHandler.GetGroupInfoFromInvite)
	sessions.Post("/:sessionId/groups/join-with-invite", groupHandler.JoinGroupWithInvite)
}

// setupNewsletterRoutes sets up newsletter management routes
func setupNewsletterRoutes(sessions fiber.Router, container *app.Container, appLogger *logger.Logger) {
	newsletterHandler := handlers.NewNewsletterHandler(appLogger, container.GetNewsletterUseCase(), container.GetSessionRepository())

	sessions.Post("/:sessionId/newsletters/create", newsletterHandler.CreateNewsletter)
	sessions.Get("/:sessionId/newsletters/info", newsletterHandler.GetNewsletterInfo)
	sessions.Post("/:sessionId/newsletters/info-from-invite", newsletterHandler.GetNewsletterInfoWithInvite)
	sessions.Post("/:sessionId/newsletters/follow", newsletterHandler.FollowNewsletter)
	sessions.Post("/:sessionId/newsletters/unfollow", newsletterHandler.UnfollowNewsletter)
	sessions.Get("/:sessionId/newsletters/messages", newsletterHandler.GetNewsletterMessages)
	sessions.Get("/:sessionId/newsletters/updates", newsletterHandler.GetNewsletterMessageUpdates)
	sessions.Post("/:sessionId/newsletters/mark-viewed", newsletterHandler.NewsletterMarkViewed)
	sessions.Post("/:sessionId/newsletters/send-reaction", newsletterHandler.NewsletterSendReaction)
	sessions.Post("/:sessionId/newsletters/subscribe-live", newsletterHandler.NewsletterSubscribeLiveUpdates)
	sessions.Post("/:sessionId/newsletters/toggle-mute", newsletterHandler.NewsletterToggleMute)
	sessions.Post("/:sessionId/newsletters/accept-tos", newsletterHandler.AcceptTOSNotice)
	sessions.Post("/:sessionId/newsletters/upload", newsletterHandler.UploadNewsletter)
	sessions.Post("/:sessionId/newsletters/upload-reader", newsletterHandler.UploadNewsletterReader)
	sessions.Get("/:sessionId/newsletters", newsletterHandler.GetSubscribedNewsletters)
}

// setupCommunityRoutes sets up community management routes
func setupCommunityRoutes(sessions fiber.Router, container *app.Container, appLogger *logger.Logger) {
	communityHandler := handlers.NewCommunityHandler(appLogger, container.GetCommunityUseCase(), container.GetSessionRepository())

	sessions.Post("/:sessionId/communities/link-group", communityHandler.LinkGroup)
	sessions.Post("/:sessionId/communities/unlink-group", communityHandler.UnlinkGroup)
	sessions.Get("/:sessionId/communities/info", communityHandler.GetCommunityInfo)
	sessions.Get("/:sessionId/communities/subgroups", communityHandler.GetSubGroups)
}

// setupContactRoutes sets up contact management routes
func setupContactRoutes(sessions fiber.Router, container *app.Container, appLogger *logger.Logger) {
	contactHandler := handlers.NewContactHandler(appLogger, container.GetContactUseCase(), container.GetSessionRepository())

	sessions.Post("/:sessionId/contacts/check", contactHandler.CheckWhatsApp)
	sessions.Get("/:sessionId/contacts/avatar", contactHandler.GetProfilePicture)
	sessions.Post("/:sessionId/contacts/info", contactHandler.GetUserInfo)
	sessions.Get("/:sessionId/contacts", contactHandler.ListContacts)
	sessions.Post("/:sessionId/contacts/sync", contactHandler.SyncContacts)
	sessions.Get("/:sessionId/contacts/business", contactHandler.GetBusinessProfile)
}

// setupWebhookRoutes sets up webhook management routes
func setupWebhookRoutes(sessions fiber.Router, container *app.Container, appLogger *logger.Logger) {
	webhookHandler := handlers.NewWebhookHandler(container.WebhookUseCase, appLogger)

	sessions.Post("/:sessionId/webhook/set", webhookHandler.SetConfig)
	sessions.Get("/:sessionId/webhook/find", webhookHandler.FindConfig)
	sessions.Post("/:sessionId/webhook/test", webhookHandler.TestWebhook)
}

// setupChatwootRoutes sets up Chatwoot integration routes
func setupChatwootRoutes(sessions fiber.Router, container *app.Container, appLogger *logger.Logger) {
	chatwootHandler := handlers.NewChatwootHandler(container.GetChatwootUseCase(), appLogger)

	sessions.Post("/:sessionId/chatwoot/set", chatwootHandler.SetConfig)
	sessions.Get("/:sessionId/chatwoot/find", chatwootHandler.FindConfig)
	sessions.Post("/:sessionId/chatwoot/contacts/sync", chatwootHandler.SyncContacts)
	sessions.Post("/:sessionId/chatwoot/conversations/sync", chatwootHandler.SyncConversations)
}

func setupSessionSpecificRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Session-specific advanced routes that require additional processing
	// Currently no additional session-specific routes needed
	// All core functionality is handled in setupSessionRoutes
}

func setupGlobalRoutes(app *fiber.App, database *db.DB, appLogger *logger.Logger, WameowManager *wameow.Manager, container *app.Container) {
	// Global webhook info routes
	webhookHandler := handlers.NewWebhookHandler(container.WebhookUseCase, appLogger)
	app.Get("/webhook/events", webhookHandler.GetSupportedEvents) // GET /webhook/events

	// Chatwoot webhook (without authentication - like Evolution API)
	chatwootHandler := handlers.NewChatwootHandler(container.GetChatwootUseCase(), appLogger)
	app.Post("/sessions/:sessionId/chatwoot/webhook", chatwootHandler.ReceiveWebhook) // POST /sessions/:sessionId/chatwoot/webhook
	app.Post("/chatwoot/webhook/:sessionId", chatwootHandler.ReceiveWebhook)          // POST /chatwoot/webhook/:sessionId (alternative route)
}
