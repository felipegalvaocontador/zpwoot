// @title zpwoot - WhatsApp Multi-Session API
// @version 1.0
// @description A complete REST API for managing multiple WhatsApp sessions using Go, Fiber, PostgreSQL, and whatsmeow.
// @description
// @description ## Authentication
// @description All API endpoints (except /health/* and /swagger/*) require API key authentication.
// @description Provide your API key in the `Authorization` header.
// @contact.name zpwoot Support
// @contact.url https://github.com/your-org/zpwoot
// @contact.email support@zpwoot.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Enter your API key directly (no Bearer prefix required). Example: a0b1125a0eb3364d98e2c49ec6f7d6ba
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_ "zpwoot/docs/swagger" // Import generated swagger docs
	"zpwoot/internal/app"
	sessionApp "zpwoot/internal/app/session"
	domainChatwoot "zpwoot/internal/domain/chatwoot"
	domainCommunity "zpwoot/internal/domain/community"
	domainContact "zpwoot/internal/domain/contact"
	domainGroup "zpwoot/internal/domain/group"
	domainMedia "zpwoot/internal/domain/media"
	domainNewsletter "zpwoot/internal/domain/newsletter"
	"zpwoot/internal/domain/session"
	domainWebhook "zpwoot/internal/domain/webhook"
	"zpwoot/internal/infra/db"
	"zpwoot/internal/infra/http/routers"

	chatwootIntegration "zpwoot/internal/infra/integrations/chatwoot"
	"zpwoot/internal/infra/integrations/webhook"
	"zpwoot/internal/infra/repository"
	"zpwoot/internal/infra/wameow"
	"zpwoot/internal/ports"
	"zpwoot/platform/config"
	platformDB "zpwoot/platform/db"
	"zpwoot/platform/logger"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

type commandFlags struct {
	migrateUp     bool
	migrateDown   bool
	migrateStatus bool
	seed          bool
	version       bool
}

type managers struct {
	whatsapp        *wameow.Manager
	webhook         *webhook.WebhookManager
	chatwoot        *chatwootIntegration.IntegrationManager
	chatwootManager *chatwootIntegration.Manager
}

func main() {
	flags := parseFlags()

	if flags.version {
		showVersion()
		return
	}

	cfg := config.Load()
	appLogger := initializeLogger(cfg)

	database := initializeDatabase(cfg, appLogger)
	defer closeDatabase(database, appLogger)

	migrator := db.NewMigrator(database.GetDB().DB, appLogger)
	if handled := handleDatabaseOperations(flags, migrator, database, appLogger); handled {
		return
	}

	repositories := repository.NewRepositories(database.GetDB(), appLogger)
	managerInstances := initializeManagers(database, repositories, appLogger)
	container := createContainer(repositories, managerInstances, database, appLogger)

	appLogger.Info("Starting server with Chi router")
	chiHandler := setupHTTPServer(cfg, container, database, managerInstances.whatsapp, appLogger)

	startBackgroundServices(container, appLogger)

	setupGracefulShutdown(appLogger)

	startServer(chiHandler, cfg, appLogger)
}

func parseFlags() commandFlags {
	flags := commandFlags{}
	flag.BoolVar(&flags.migrateUp, "migrate-up", false, "Run database migrations up")
	flag.BoolVar(&flags.migrateDown, "migrate-down", false, "Rollback last migration")
	flag.BoolVar(&flags.migrateStatus, "migrate-status", false, "Show migration status")
	flag.BoolVar(&flags.seed, "seed", false, "Seed database with sample data")
	flag.BoolVar(&flags.version, "version", false, "Show version information")

	flag.Parse()
	return flags
}

func initializeLogger(cfg *config.Config) *logger.Logger {
	loggerConfig := &logger.LogConfig{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Output: cfg.LogOutput,
		Caller: cfg.IsDevelopment(),
	}

	if cfg.IsProduction() {
		loggerConfig = logger.ProductionConfig()
		loggerConfig.Level = cfg.LogLevel
	}

	return logger.NewWithConfig(loggerConfig)
}

func initializeDatabase(cfg *config.Config, appLogger *logger.Logger) *platformDB.DB {
	database, err := platformDB.NewWithMigrations(cfg.DatabaseURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database and run migrations: " + err.Error())
	}
	return database
}

func closeDatabase(database *platformDB.DB, appLogger *logger.Logger) {
	if err := database.Close(); err != nil {
		appLogger.Error("Failed to close database connection: " + err.Error())
	}
}

func handleDatabaseOperations(
	flags commandFlags,
	migrator *db.Migrator,
	database *platformDB.DB,
	appLogger *logger.Logger,
) bool {
	if flags.migrateUp {
		if err := migrator.RunMigrations(); err != nil {
			appLogger.Fatal("Failed to run migrations: " + err.Error())
		}
		appLogger.Info("Migrations completed successfully")
		return true
	}

	if flags.migrateDown {
		if err := migrator.Rollback(); err != nil {
			appLogger.Fatal("Failed to rollback migration: " + err.Error())
		}
		appLogger.Info("Migration rollback completed successfully")
		return true
	}

	if flags.migrateStatus {
		migrations, err := migrator.GetMigrationStatus()
		if err != nil {
			appLogger.Fatal("Failed to get migration status: " + err.Error())
		}
		showMigrationStatus(migrations)
		return true
	}

	if flags.seed {
		if err := seedDatabase(database, appLogger); err != nil {
			appLogger.Fatal("Failed to seed database: " + err.Error())
		}
		appLogger.Info("Database seeding completed successfully")
		return true
	}

	return false
}

func initializeManagers(
	database *platformDB.DB,
	repositories *repository.Repositories,
	appLogger *logger.Logger,
) managers {
	whatsappManager := createWhatsAppManager(database, repositories.GetSessionRepository(), appLogger)
	webhookManager := createWebhookManager(repositories.GetWebhookRepository(), appLogger)
	chatwootIntegrationManager, chatwootManager := createChatwootIntegration(repositories, appLogger)

	configureWebhookIntegration(whatsappManager, webhookManager, appLogger)
	configureChatwootIntegration(whatsappManager, chatwootIntegrationManager, appLogger)

	return managers{
		whatsapp:        whatsappManager,
		webhook:         webhookManager,
		chatwoot:        chatwootIntegrationManager,
		chatwootManager: chatwootManager,
	}
}

func createWhatsAppManager(database *platformDB.DB, sessionRepo ports.SessionRepository, appLogger *logger.Logger) *wameow.Manager {
	factory, err := wameow.NewFactory(appLogger, sessionRepo)
	if err != nil {
		appLogger.Fatal("Failed to create wameow factory: " + err.Error())
	}

	manager, err := factory.CreateManager(database.GetDB().DB)
	if err != nil {
		appLogger.Fatal("Failed to create WhatsApp manager: " + err.Error())
	}

	appLogger.Info("WhatsApp manager initialized")
	return manager
}

func createWebhookManager(webhookRepo ports.WebhookRepository, appLogger *logger.Logger) *webhook.WebhookManager {
	const defaultWebhookWorkers = 5
	webhookManager := webhook.NewWebhookManager(appLogger, webhookRepo, defaultWebhookWorkers)

	if err := webhookManager.Start(); err != nil {
		appLogger.Fatal("Failed to start webhook manager: " + err.Error())
	}

	appLogger.Info("Webhook manager initialized and started")
	return webhookManager
}

func createChatwootIntegration(repositories *repository.Repositories, appLogger *logger.Logger) (*chatwootIntegration.IntegrationManager, *chatwootIntegration.Manager) {
	chatwootRepo := repositories.GetChatwootRepository()
	chatwootMessageRepo := repositories.GetChatwootMessageRepository()

	chatwootManager := chatwootIntegration.NewManager(appLogger, chatwootRepo)
	messageMapper := chatwootIntegration.NewMessageMapper(appLogger, chatwootMessageRepo)
	contactSync := chatwootIntegration.NewContactSync(appLogger, nil)
	conversationMgr := chatwootIntegration.NewConversationManager(appLogger, nil)
	formatter := chatwootIntegration.NewMessageFormatter(appLogger)

	integrationManager := chatwootIntegration.NewIntegrationManager(
		appLogger,
		chatwootManager,
		messageMapper,
		contactSync,
		conversationMgr,
		formatter,
	)

	appLogger.Info("Chatwoot integration initialized successfully")
	return integrationManager, chatwootManager
}

func createContainer(repositories *repository.Repositories, managers managers, database *platformDB.DB, appLogger *logger.Logger) *app.Container {
	adapters := createAdapters(repositories, managers, appLogger)

	services := createDomainServices(repositories, managers, appLogger, adapters)

	config := createContainerConfig(repositories, managers, database, appLogger, adapters, services)

	return app.NewContainer(config)
}

func createAdapters(repositories *repository.Repositories, managers managers, appLogger *logger.Logger) *containerAdapters {
	var chatwootMessageMapper ports.ChatwootMessageMapper
	if repositories.GetChatwootMessageRepository() != nil {
		chatwootMessageMapper = chatwootIntegration.NewMessageMapper(appLogger, repositories.GetChatwootMessageRepository())
	}

	return &containerAdapters{
		chatwootMessageMapper: chatwootMessageMapper,
		jidValidator:          wameow.NewJIDValidatorAdapter(),
		newsletterManager:     wameow.NewNewsletterAdapter(managers.whatsapp, *appLogger),
		communityManager:      wameow.NewCommunityAdapter(managers.whatsapp, *appLogger),
		qrGenerator:           wameow.NewQRCodeGenerator(appLogger),
	}
}

type containerAdapters struct {
	chatwootMessageMapper ports.ChatwootMessageMapper
	jidValidator          ports.JIDValidator
	newsletterManager     ports.NewsletterManager
	communityManager      ports.CommunityManager
	qrGenerator           *wameow.QRCodeGenerator
}

func createDomainServices(repositories *repository.Repositories, managers managers, appLogger *logger.Logger, adapters *containerAdapters) *containerServices {
	sessionService := session.NewService(
		repositories.GetSessionRepository(),
		managers.whatsapp,
		adapters.qrGenerator,
		appLogger,
	)

	webhookService := domainWebhook.NewService(
		appLogger,
		repositories.GetWebhookRepository(),
	)

	chatwootService := domainChatwoot.NewService(
		appLogger,
		repositories.GetChatwootRepository(),
		managers.whatsapp,
	)

	if adapters.chatwootMessageMapper != nil {
		chatwootService.SetMessageMapper(adapters.chatwootMessageMapper)
	}

	return &containerServices{
		sessionService:    sessionService,
		webhookService:    webhookService,
		chatwootService:   chatwootService,
		groupService:      domainGroup.NewService(nil, managers.whatsapp, adapters.jidValidator),
		contactService:    domainContact.NewService(managers.whatsapp, appLogger),
		mediaService:      domainMedia.NewService(nil, nil, appLogger, "/tmp/media_cache"),
		newsletterService: domainNewsletter.NewService(nil),
		communityService:  domainCommunity.NewService(),
	}
}

type containerServices struct {
	sessionService    *session.Service
	webhookService    *domainWebhook.Service
	chatwootService   *domainChatwoot.Service
	groupService      *domainGroup.Service
	contactService    domainContact.Service
	mediaService      domainMedia.Service
	newsletterService *domainNewsletter.Service
	communityService  domainCommunity.Service
}

func createContainerConfig(repositories *repository.Repositories, managers managers, database *platformDB.DB, appLogger *logger.Logger, adapters *containerAdapters, services *containerServices) *app.ContainerConfig {
	return &app.ContainerConfig{
		SessionRepo:         repositories.GetSessionRepository(),
		WebhookRepo:         repositories.GetWebhookRepository(),
		ChatwootRepo:        repositories.GetChatwootRepository(),
		ChatwootMessageRepo: repositories.GetChatwootMessageRepository(),

		WameowManager:         managers.whatsapp,
		ChatwootIntegration:   nil, // IntegrationManager doesn't implement this interface
		ChatwootManager:       managers.chatwootManager,
		ChatwootMessageMapper: adapters.chatwootMessageMapper,
		JIDValidator:          adapters.jidValidator,
		NewsletterManager:     adapters.newsletterManager,
		CommunityManager:      adapters.communityManager,

		SessionService:    services.sessionService,
		WebhookService:    services.webhookService,
		ChatwootService:   services.chatwootService,
		GroupService:      services.groupService,
		ContactService:    services.contactService,
		MediaService:      services.mediaService,
		NewsletterService: services.newsletterService,
		CommunityService:  services.communityService,

		Logger: appLogger,
		DB:     database.GetDB().DB,

		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
	}
}

func startBackgroundServices(container *app.Container, appLogger *logger.Logger) {
	go connectOnStartup(container, appLogger)
}

func showVersion() {
	fmt.Printf("zpwoot - WhatsApp Multi-Session API\n")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func showMigrationStatus(migrations []*db.Migration) {
	fmt.Printf("Migration Status:\n")
	fmt.Printf("================\n\n")

	if len(migrations) == 0 {
		fmt.Printf("No migrations found.\n")
		return
	}

	for _, migration := range migrations {
		status := "PENDING"
		appliedAt := "Not applied"

		if migration.AppliedAt != nil {
			status = "APPLIED"
			appliedAt = migration.AppliedAt.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("Version: %03d | Status: %-7s | Name: %s | Applied: %s\n",
			migration.Version, status, migration.Name, appliedAt)
	}
	fmt.Printf("\n")
}

func seedDatabase(database *platformDB.DB, logger *logger.Logger) error {
	logger.Info("Starting database seeding...")

	sampleSessions := []map[string]interface{}{
		{
			"id":         "sample-session-1",
			"name":       "Sample WhatsApp Session",
			"device_jid": "5511999999999@s.whatsapp.net",
			"status":     "created",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	sampleWebhooks := []map[string]interface{}{
		{
			"id":         "sample-webhook-1",
			"session_id": "sample-session-1",
			"url":        "https://example.com/webhook",
			"events":     []string{"message", "status"},
			"enabled":    true,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	for _, session := range sampleSessions {
		query := `
			INSERT INTO "zpSessions" ("id", "name", "deviceJid", "status", "createdAt", "updatedAt")
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT ("id") DO NOTHING
		`
		_, err := database.GetDB().Exec(query,
			session["id"], session["name"], session["device_jid"],
			session["status"], session["created_at"], session["updated_at"])
		if err != nil {
			return fmt.Errorf("failed to insert sample session: %w", err)
		}
	}

	for _, webhook := range sampleWebhooks {
		query := `
			INSERT INTO "zpWebhooks" ("id", "sessionId", "url", "events", "enabled", "createdAt", "updatedAt")
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT ("id") DO NOTHING
		`
		_, err := database.GetDB().Exec(query,
			webhook["id"], webhook["session_id"], webhook["url"],
			webhook["events"], webhook["enabled"], webhook["created_at"], webhook["updated_at"])
		if err != nil {
			return fmt.Errorf("failed to insert sample webhook: %w", err)
		}
	}

	logger.InfoWithFields("Database seeding completed", map[string]interface{}{
		"sessions_created": len(sampleSessions),
		"webhooks_created": len(sampleWebhooks),
	})

	return nil
}

func configureWebhookIntegration(wameowManager *wameow.Manager, webhookManager *webhook.WebhookManager, appLogger *logger.Logger) {
	webhookHandler := wameow.NewWhatsmeowWebhookHandler(appLogger, webhookManager)
	wameowManager.SetWebhookHandler(webhookHandler)
	appLogger.Info("Webhook integration configured successfully")
}

func configureChatwootIntegration(whatsappManager *wameow.Manager, integrationManager *chatwootIntegration.IntegrationManager, appLogger *logger.Logger) {
	whatsappManager.SetChatwootManager(integrationManager)
	appLogger.Info("Chatwoot integration configured successfully")
}

func connectOnStartup(container *app.Container, logger *logger.Logger) {
	const (
		startupDelay     = 3 * time.Second
		operationTimeout = 60 * time.Second
		sessionLimit     = 100
		reconnectDelay   = 1 * time.Second
	)

	time.Sleep(startupDelay)

	sessionUC := container.GetSessionUseCase()
	sessionRepo := container.GetSessionRepository()

	if sessionUC == nil || sessionRepo == nil {
		logger.Error("Required components not available, skipping auto-connect")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	sessions := getExistingSessions(ctx, sessionRepo, sessionLimit, logger)
	if len(sessions) == 0 {
		logger.Info("No existing sessions found, skipping auto-connect")
		return
	}

	logger.InfoWithFields("Starting auto-reconnect", map[string]interface{}{
		"total_sessions": len(sessions),
	})

	stats := reconnectSessions(ctx, sessions, sessionUC, logger, reconnectDelay)

	logger.InfoWithFields("Auto-reconnect completed", map[string]interface{}{
		"connected": stats.connected,
		"skipped":   stats.skipped,
		"failed":    stats.failed,
	})
}

type reconnectStats struct {
	connected int
	skipped   int
	failed    int
}

func getExistingSessions(ctx context.Context, sessionRepo ports.SessionRepository, limit int, logger *logger.Logger) []*session.Session {
	sessions, _, err := sessionRepo.List(ctx, &session.ListSessionsRequest{
		Limit:  limit,
		Offset: 0,
	})
	if err != nil {
		logger.ErrorWithFields("Failed to get sessions for auto-connect", map[string]interface{}{
			"error": err.Error(),
		})
		return nil
	}
	return sessions
}

func reconnectSessions(ctx context.Context, sessions []*session.Session, sessionUC sessionApp.UseCase, logger *logger.Logger, delay time.Duration) reconnectStats {
	stats := reconnectStats{}

	for _, sess := range sessions {
		sessionID := sess.ID.String()

		if sess.DeviceJid == "" {
			stats.skipped++
			continue
		}

		if _, err := sessionUC.ConnectSession(ctx, sessionID); err != nil {
			logger.ErrorWithFields("Failed to auto-connect session", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			stats.failed++
			continue
		}

		stats.connected++
		time.Sleep(delay)
	}

	return stats
}

func setupHTTPServer(cfg *config.Config, container *app.Container, database *platformDB.DB, whatsappManager *wameow.Manager, appLogger *logger.Logger) http.Handler {
	return routers.SetupRoutes(cfg, database, appLogger, whatsappManager, container)
}

func setupGracefulShutdown(appLogger *logger.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		appLogger.Info("Shutting down Chi server...")
	}()
}

func startServer(handler http.Handler, cfg *config.Config, appLogger *logger.Logger) {
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	appLogger.InfoWithFields("Starting zpwoot server with Chi", map[string]interface{}{
		"port":        cfg.Port,
		"server_host": cfg.ServerHost,
		"environment": cfg.NodeEnv,
		"log_level":   cfg.LogLevel,
		"router":      "chi",
	})

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.Fatal("Chi server failed to start: " + err.Error())
	}
}
