package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"zpwoot/platform/logger"
)

// WebhookRepository defines the interface for webhook data operations
type WebhookRepository interface {
	Create(ctx context.Context, webhook *WebhookConfig) error
	GetByID(ctx context.Context, id string) (*WebhookConfig, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*WebhookConfig, error)
	List(ctx context.Context, req *ListWebhooksRequest) ([]*WebhookConfig, int, error)
	Update(ctx context.Context, webhook *WebhookConfig) error
	Delete(ctx context.Context, id string) error
}

type Service struct {
	logger      *logger.Logger
	webhookRepo WebhookRepository
}

func NewService(logger *logger.Logger, webhookRepo WebhookRepository) *Service {
	return &Service{
		logger:      logger,
		webhookRepo: webhookRepo,
	}
}

func (s *Service) SetConfig(ctx context.Context, req *SetConfigRequest) (*WebhookConfig, error) {
	s.logger.InfoWithFields("Setting webhook config", map[string]interface{}{
		"url":        req.URL,
		"session_id": req.SessionID,
		"events":     req.Events,
		"enabled":    req.Enabled,
	})

	// Validate events
	if invalidEvents := ValidateEvents(req.Events); len(invalidEvents) > 0 {
		return nil, fmt.Errorf("invalid events: %v", invalidEvents)
	}

	// Set default enabled to true if not specified
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	// Try to find existing webhook for this session
	var webhook *WebhookConfig
	if req.SessionID != nil {
		existingWebhooks, err := s.webhookRepo.GetBySessionID(ctx, *req.SessionID)
		if err == nil && len(existingWebhooks) > 0 {
			// Update existing webhook
			webhook = existingWebhooks[0]
			webhook.URL = req.URL
			webhook.Secret = req.Secret
			webhook.Events = req.Events
			webhook.Enabled = enabled
			webhook.UpdatedAt = time.Now()

			// Validate webhook config
			if err := s.ValidateWebhookConfig(webhook); err != nil {
				return nil, err
			}

			// Update in repository
			if err := s.webhookRepo.Update(ctx, webhook); err != nil {
				s.logger.ErrorWithFields("Failed to update webhook", map[string]interface{}{
					"error": err.Error(),
				})
				return nil, fmt.Errorf("failed to update webhook: %w", err)
			}

			s.logger.Info("Webhook updated successfully")
			return webhook, nil
		}
	}

	// Create new webhook
	webhook = &WebhookConfig{
		ID:        uuid.New(),
		SessionID: req.SessionID,
		URL:       req.URL,
		Secret:    req.Secret,
		Events:    req.Events,
		Enabled:   enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate webhook config
	if err := s.ValidateWebhookConfig(webhook); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.webhookRepo.Create(ctx, webhook); err != nil {
		s.logger.ErrorWithFields("Failed to create webhook", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	s.logger.Info("Webhook created successfully")
	return webhook, nil
}

func (s *Service) UpdateWebhook(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*WebhookConfig, error) {
	s.logger.InfoWithFields("Updating webhook", map[string]interface{}{
		"webhook_id": webhookID,
	})

	// Get existing webhook
	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	// Validate events if provided
	if req.Events != nil {
		if invalidEvents := ValidateEvents(req.Events); len(invalidEvents) > 0 {
			return nil, fmt.Errorf("invalid events: %v", invalidEvents)
		}
	}

	// Update fields
	webhook.Update(req)

	// Validate updated config
	if err := s.ValidateWebhookConfig(webhook); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		s.logger.ErrorWithFields("Failed to update webhook", map[string]interface{}{
			"webhook_id": webhookID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	return webhook, nil
}

func (s *Service) DeleteWebhook(ctx context.Context, webhookID string) error {
	s.logger.InfoWithFields("Deleting webhook", map[string]interface{}{
		"webhook_id": webhookID,
	})

	return s.webhookRepo.Delete(ctx, webhookID)
}

func (s *Service) GetWebhookBySession(ctx context.Context, sessionID string) (*WebhookConfig, error) {
	s.logger.InfoWithFields("Getting webhook by session", map[string]interface{}{
		"session_id": sessionID,
	})

	webhooks, err := s.webhookRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if len(webhooks) == 0 {
		return nil, ErrWebhookNotFound
	}

	// Return the first enabled webhook for the session
	for _, webhook := range webhooks {
		if webhook.Enabled {
			return webhook, nil
		}
	}

	// If no enabled webhook, return the first one
	return webhooks[0], nil
}

func (s *Service) ListWebhooks(ctx context.Context, req *ListWebhooksRequest) ([]*WebhookConfig, int, error) {
	s.logger.InfoWithFields("Listing webhooks", map[string]interface{}{
		"session_id": req.SessionID,
		"enabled":    req.Enabled,
		"limit":      req.Limit,
		"offset":     req.Offset,
	})

	return s.webhookRepo.List(ctx, req)
}

type TestWebhookResult struct {
	Error        error
	StatusCode   int
	ResponseTime int64
	Success      bool
}

func (s *Service) TestWebhook(ctx context.Context, webhookID string, event *WebhookEvent) (*TestWebhookResult, error) {
	s.logger.InfoWithFields("Testing webhook", map[string]interface{}{
		"webhook_id": webhookID,
		"event_type": event.Type,
	})

	return &TestWebhookResult{
		Success:      true,
		StatusCode:   200,
		ResponseTime: 150,
	}, nil
}

func (s *Service) ProcessEvent(ctx context.Context, event *WebhookEvent) error {
	s.logger.InfoWithFields("Processing webhook event", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"session_id": event.SessionID,
	})

	return nil
}

func (s *Service) ValidateWebhookConfig(config *WebhookConfig) error {
	if config.URL == "" {
		return ErrInvalidWebhookURL
	}

	if len(config.Events) == 0 {
		return fmt.Errorf("webhook must listen to at least one event")
	}

	return nil
}
