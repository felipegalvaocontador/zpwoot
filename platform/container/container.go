package container

import (
	"context"
	"fmt"
	"net/http"

	// External dependencies
	"go.mau.fi/whatsmeow/store/sqlstore"
	_ "github.com/lib/pq" // PostgreSQL driver for sqlstore

	// Core business logic
	"zpwoot/internal/core/session"
	"zpwoot/internal/core/messaging"

	// Application services
	"zpwoot/internal/services"
	"zpwoot/internal/services/shared/validation"

	// Adapters
	"zpwoot/internal/adapters/repository"
	"zpwoot/internal/adapters/http/router"
	"zpwoot/internal/adapters/waclient"

	// Platform
	"zpwoot/platform/config"
	"zpwoot/platform/database"
	"zpwoot/platform/logger"
)

// Container é o container principal de Dependency Injection
// Contém apenas os componentes essenciais que sabemos que funcionam
type Container struct {
	// Platform dependencies
	config   *config.Config
	logger   *logger.Logger
	database *database.Database

	// Core business logic services
	sessionCore   *session.Service
	messagingCore *messaging.Service

	// Application services
	sessionService   *services.SessionService
	messagingService *services.MessageService

	// Adapters
	sessionRepo     session.Repository
	messageRepo     messaging.Repository
	whatsappGateway session.WhatsAppGateway
}

// Config estrutura de configuração para o container
type Config struct {
	AppConfig *config.Config
	Logger    *logger.Logger
	Database  *database.Database
}

// New cria uma nova instância do container
func New(cfg *Config) (*Container, error) {
	container := &Container{
		config:   cfg.AppConfig,
		logger:   cfg.Logger,
		database: cfg.Database,
	}

	// Inicializar componentes
	if err := container.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	cfg.Logger.Info("Dependency injection container initialized successfully")
	return container, nil
}

// initialize inicializa todos os componentes
func (c *Container) initialize() error {
	c.logger.Debug("Initializing container...")

	// 1. Database repositories
	c.sessionRepo = repository.NewSessionRepository(c.database.DB)
	c.messageRepo = repository.NewMessageRepository(c.database.DB, c.logger)

	// 2. WhatsApp sqlstore container
	waContainer, err := c.createWhatsAppContainer()
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp container: %w", err)
	}

	// 3. External gateways
	c.whatsappGateway = waclient.NewGateway(waContainer, c.logger)

	// 4. QR Generator
	qrGenerator := waclient.NewQRGenerator(c.logger)

	// 4. Core services
	c.sessionCore = session.NewService(
		c.sessionRepo,
		c.whatsappGateway,
		qrGenerator,
	)

	c.messagingCore = messaging.NewService(
		c.messageRepo,
		c.logger,
	)

	// 5. Validator
	validator := validation.New()

	// 6. Application services
	c.sessionService = services.NewSessionService(
		c.sessionCore,
		c.sessionRepo,
		c.whatsappGateway,
		qrGenerator,
		c.logger,
		validator,
	)

	c.messagingService = services.NewMessageService(
		c.messagingCore,
		c.sessionCore,
		c.messageRepo,
		c.sessionRepo,
		c.whatsappGateway,
		c.logger,
		validator,
	)

	c.logger.Debug("Container initialized successfully")
	return nil
}

// ===== MÉTODOS PÚBLICOS =====

// GetConfig retorna a configuração da aplicação
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// GetLogger retorna o logger da aplicação
func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

// GetDatabase retorna a instância do banco de dados
func (c *Container) GetDatabase() *database.Database {
	return c.database
}

// GetSessionService retorna o service de sessões
func (c *Container) GetSessionService() *services.SessionService {
	return c.sessionService
}

// GetMessageService retorna o service de mensagens
func (c *Container) GetMessageService() *services.MessageService {
	return c.messagingService
}

// GetSessionCore retorna o core service de sessões
func (c *Container) GetSessionCore() *session.Service {
	return c.sessionCore
}

// GetWhatsAppGateway retorna o gateway do WhatsApp
func (c *Container) GetWhatsAppGateway() session.WhatsAppGateway {
	return c.whatsappGateway
}

// ===== LIFECYCLE METHODS =====

// Start inicia todos os componentes que precisam de inicialização
func (c *Container) Start(ctx context.Context) error {
	c.logger.Info("Starting container components...")

	// Iniciar WhatsApp gateway se necessário
	if starter, ok := c.whatsappGateway.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start WhatsApp gateway: %w", err)
		}
	}

	c.logger.Info("Container components started successfully")
	return nil
}

// Stop para todos os componentes gracefully
func (c *Container) Stop(ctx context.Context) error {
	c.logger.Info("Stopping container components...")

	// Parar WhatsApp gateway se necessário
	if stopper, ok := c.whatsappGateway.(interface{ Stop(context.Context) error }); ok {
		if err := stopper.Stop(ctx); err != nil {
			c.logger.ErrorWithFields("Failed to stop WhatsApp gateway", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Fechar conexão com banco de dados
	if err := c.database.Close(); err != nil {
		c.logger.ErrorWithFields("Failed to close database connection", map[string]interface{}{
			"error": err.Error(),
		})
	}

	c.logger.Info("Container components stopped successfully")
	return nil
}

// Handler retorna um handler HTTP completo com todas as rotas
func (c *Container) Handler() http.Handler {
	return router.SetupRoutes(c.config, c.logger, c.sessionService, c.messagingService)
}
