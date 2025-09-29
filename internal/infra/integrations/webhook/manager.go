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

// WebhookManager coordinates webhook delivery and event dispatching
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

// NewWebhookManager creates a new webhook manager
func NewWebhookManager(
	logger *logger.Logger,
	webhookRepo ports.WebhookRepository,
	workers int,
) *WebhookManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Create delivery service
	deliveryService := NewWebhookDeliveryService(logger, webhookRepo, workers)

	// Create event dispatcher
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

// Start initializes the webhook manager and starts background workers
func (m *WebhookManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		m.logger.Warn("Webhook manager is already started")
		return nil
	}

	m.logger.Info("Starting webhook manager")

	// Start the delivery service workers
	m.deliveryService.Start(m.ctx)

	m.started = true
	m.logger.Info("Webhook manager started successfully")

	return nil
}

// Stop gracefully shuts down the webhook manager
func (m *WebhookManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		m.logger.Warn("Webhook manager is not started")
		return nil
	}

	m.logger.Info("Stopping webhook manager")

	// Cancel the context to stop all workers
	m.cancel()

	m.started = false
	m.logger.Info("Webhook manager stopped successfully")

	return nil
}

// IsStarted returns whether the webhook manager is currently running
func (m *WebhookManager) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

// DispatchEvent processes and dispatches a whatsmeow event to webhooks
func (m *WebhookManager) DispatchEvent(evt interface{}, sessionID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.started {
		m.logger.Debug("Webhook manager not started, skipping event dispatch")
		return nil
	}

	return m.eventDispatcher.DispatchEvent(m.ctx, evt, sessionID)
}

// GetEventDispatcher returns the event dispatcher for direct access
func (m *WebhookManager) GetEventDispatcher() *EventDispatcher {
	return m.eventDispatcher
}

// GetDeliveryService returns the delivery service for direct access
func (m *WebhookManager) GetDeliveryService() *WebhookDeliveryService {
	return m.deliveryService
}

// GetStats returns statistics about webhook operations
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

// WebhookStats contains statistics about webhook operations
type WebhookStats struct {
	Started       bool   `json:"started"`
	Workers       int    `json:"workers"`
	QueueSize     int    `json:"queue_size"`
	QueueCapacity int    `json:"queue_capacity"`
	MaxRetries    int    `json:"max_retries"`
	RetryDelay    string `json:"retry_delay"`
}

// TestWebhook tests a webhook endpoint with a sample event
func (m *WebhookManager) TestWebhook(webhookID, eventType string, testData map[string]interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.started {
		return ErrWebhookManagerNotStarted
	}

	// Get webhook configuration
	_, err := m.webhookRepo.GetByID(m.ctx, webhookID)
	if err != nil {
		return err
	}

	// Create test event
	testEvent := &webhook.WebhookEvent{
		ID:        "test-" + webhookID,
		SessionID: "test-session",
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      testData,
	}

	// Deliver the test event
	return m.deliveryService.DeliverEvent(m.ctx, testEvent)
}

// Errors
var (
	ErrWebhookManagerNotStarted = fmt.Errorf("webhook manager is not started")
)
