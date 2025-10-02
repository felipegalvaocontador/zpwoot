package container

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow/store/sqlstore"

	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/session"

	"zpwoot/internal/services"
	"zpwoot/internal/services/shared/validation"

	"zpwoot/internal/adapters/repository"
	"zpwoot/internal/adapters/server"
	"zpwoot/internal/adapters/waclient"

	"zpwoot/platform/config"
	"zpwoot/platform/database"
	"zpwoot/platform/logger"
)

type Container struct {
	config   *config.Config
	logger   *logger.Logger
	database *database.Database

	sessionCore   *session.Service
	messagingCore *messaging.Service

	sessionService   *services.SessionService
	messagingService *services.MessageService
	groupService     *services.GroupService

	sessionRepo     session.Repository
	messageRepo     messaging.Repository
	whatsappGateway session.WhatsAppGateway
}

type Config struct {
	AppConfig *config.Config
	Logger    *logger.Logger
	Database  *database.Database
}

func New(cfg *Config) (*Container, error) {
	container := &Container{
		config:   cfg.AppConfig,
		logger:   cfg.Logger,
		database: cfg.Database,
	}

	if err := container.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	return container, nil
}

func (c *Container) createWhatsAppContainer() (*sqlstore.Container, error) {
	c.logger.Debug("Creating WhatsApp sqlstore container...")

	ctx := context.Background()

	waLogger := waclient.NewWhatsmeowLogger(c.logger)

	container, err := sqlstore.New(ctx, "postgres", c.config.Database.URL, waLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlstore container: %w", err)
	}

	err = container.Upgrade(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade sqlstore: %w", err)
	}

	return container, nil
}

func (c *Container) initialize() error {
	c.logger.Debug("Initializing container...")

	c.sessionRepo = repository.NewSessionRepository(c.database.DB)
	c.messageRepo = repository.NewMessageRepository(c.database.DB, c.logger)

	waContainer, err := c.createWhatsAppContainer()
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp container: %w", err)
	}

	c.whatsappGateway = waclient.NewGateway(waContainer, c.logger)

	if gateway, ok := c.whatsappGateway.(*waclient.Gateway); ok {
		gateway.SetDatabase(c.database.DB)
	}

	qrGenerator := waclient.NewQRGenerator(c.logger)

	c.sessionCore = session.NewService(
		c.sessionRepo,
		c.whatsappGateway,
		qrGenerator,
	)

	c.messagingCore = messaging.NewService(
		c.messageRepo,
		c.logger,
	)

	validator := validation.New()

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

	c.groupService = services.NewGroupService(
		nil,
		nil,
		nil,
		c.logger,
		validator,
	)

	sessionServiceAdapter := &sessionServiceAdapter{service: c.sessionService}
	if gateway, ok := c.whatsappGateway.(*waclient.Gateway); ok {
		gateway.SetSessionService(sessionServiceAdapter)
	}

	c.logger.Debug("Container initialized successfully")
	return nil
}

func (c *Container) Start(ctx context.Context) error {
	return nil
}

func (c *Container) GetConfig() *config.Config {
	return c.config
}

func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

func (c *Container) GetDatabase() *database.Database {
	return c.database
}

func (c *Container) GetSessionService() *services.SessionService {
	return c.sessionService
}

func (c *Container) GetMessageService() *services.MessageService {
	return c.messagingService
}

func (c *Container) GetSessionCore() *session.Service {
	return c.sessionCore
}

func (c *Container) GetWhatsAppGateway() session.WhatsAppGateway {
	return c.whatsappGateway
}

func (c *Container) Stop(ctx context.Context) error {

	if stopper, ok := c.whatsappGateway.(interface{ Stop(context.Context) error }); ok {
		stopper.Stop(ctx)
	}

	c.database.Close()

	return nil
}

func (c *Container) Server() *server.Server {
	return server.New(&server.Config{
		Config:         c.config,
		Logger:         c.logger,
		SessionService: c.sessionService,
		MessageService: c.messagingService,
		GroupService:   c.groupService,
	})
}

func (c *Container) Handler() http.Handler {
	return c.Server().Handler()
}

type sessionServiceAdapter struct {
	service *services.SessionService
}

func (a *sessionServiceAdapter) UpdateDeviceJID(ctx context.Context, id uuid.UUID, deviceJID string) error {

	return nil
}

func (a *sessionServiceAdapter) UpdateQRCode(ctx context.Context, id uuid.UUID, qrCode string, expiresAt time.Time) error {

	return nil
}

func (a *sessionServiceAdapter) ClearQRCode(ctx context.Context, id uuid.UUID) error {

	return nil
}

func (a *sessionServiceAdapter) GetSession(ctx context.Context, sessionID string) (*waclient.SessionInfoResponse, error) {
	response, err := a.service.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return &waclient.SessionInfoResponse{
		Session: &waclient.SessionDTO{
			ID:        response.Session.ID,
			Name:      response.Session.Name,
			DeviceJID: &response.Session.DeviceJID,
		},
	}, nil
}
