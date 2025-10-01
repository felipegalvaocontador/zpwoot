package webhook

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type WebhookManager struct {
	logger          *logger.Logger
	webhookRepo     ports.WebhookRepository
	deliveryService *WebhookDeliveryService
	eventDispatcher *EventDispatcher
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	started         bool
}

func NewWebhookManager(
	logger *logger.Logger,
	webhookRepo ports.WebhookRepository,
	workers int,
) *WebhookManager {
	ctx, cancel := context.WithCancel(context.Background())

	deliveryService := NewWebhookDeliveryService(logger, webhookRepo, workers)

	eventDispatcher := NewEventDispatcher(logger, deliveryService)

	return &WebhookManager{
		logger:          logger,
		webhookRepo:     webhookRepo,
		deliveryService: deliveryService,
		eventDispatcher: eventDispatcher,
		ctx:             ctx,
		cancel:          cancel,
		started:         false,
	}
}

func (m *WebhookManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		m.logger.Warn("Webhook manager is already started")
		return nil
	}

	m.logger.Info("Starting webhook manager")

	m.deliveryService.Start(m.ctx)

	m.started = true
	m.logger.Info("Webhook manager started successfully")

	return nil
}

func (m *WebhookManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		m.logger.Warn("Webhook manager is not started")
		return nil
	}

	m.logger.Info("Stopping webhook manager")

	m.cancel()

	m.started = false
	m.logger.Info("Webhook manager stopped successfully")

	return nil
}

func (m *WebhookManager) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

func (m *WebhookManager) DispatchEvent(evt interface{}, sessionID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.started {
		m.logger.Debug("Webhook manager not started, skipping event dispatch")
		return nil
	}

	return m.eventDispatcher.DispatchEvent(m.ctx, evt, sessionID)
}

func (m *WebhookManager) GetEventDispatcher() *EventDispatcher {
	return m.eventDispatcher
}

func (m *WebhookManager) GetDeliveryService() *WebhookDeliveryService {
	return m.deliveryService
}

func (m *WebhookManager) GetStats() *WebhookStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &WebhookStats{
		Started:       m.started,
		Workers:       m.deliveryService.workers,
		QueueSize:     len(m.deliveryService.deliveryQueue),
		QueueCapacity: cap(m.deliveryService.deliveryQueue),
		MaxRetries:    m.deliveryService.maxRetries,
		RetryDelay:    m.deliveryService.retryDelay.String(),
	}
}

type WebhookStats struct {
	RetryDelay    string `json:"retry_delay"`
	Workers       int    `json:"workers"`
	QueueSize     int    `json:"queue_size"`
	QueueCapacity int    `json:"queue_capacity"`
	MaxRetries    int    `json:"max_retries"`
	Started       bool   `json:"started"`
}

func (m *WebhookManager) TestWebhook(webhookID, eventType string, testData map[string]interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.started {
		return ErrWebhookManagerNotStarted
	}

	_, err := m.webhookRepo.GetByID(m.ctx, webhookID)
	if err != nil {
		return err
	}

	testEvent := &webhook.WebhookEvent{
		ID:        "test-" + webhookID,
		SessionID: "test-session",
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      testData,
	}

	return m.deliveryService.DeliverEvent(m.ctx, testEvent)
}

var (
	ErrWebhookManagerNotStarted = fmt.Errorf("webhook manager is not started")
)
