package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type WebhookEventProcessor interface {
	ProcessWebhookEvent(ctx context.Context, event *webhook.WebhookEvent) error
}

type WebhookDeliveryService struct {
	webhookRepo   ports.WebhookRepository
	logger        *logger.Logger
	httpClient    *http.Client
	deliveryQueue chan *DeliveryTask
	processors    []WebhookEventProcessor
	maxRetries    int
	retryDelay    time.Duration
	workers       int
}

type DeliveryTask struct {
	WebhookConfig *webhook.WebhookConfig
	Event         *webhook.WebhookEvent
	Attempt       int
	MaxAttempts   int
}

type WebhookPayload struct {
	Data      map[string]interface{} `json:"data"`
	Event     string                 `json:"event"`
	SessionID string                 `json:"sessionId"`
	Timestamp int64                  `json:"timestamp"`
}

type DeliveryResult struct {
	ResponseBody string        `json:"response_body"`
	Error        string        `json:"error,omitempty"`
	StatusCode   int           `json:"status_code"`
	Latency      time.Duration `json:"latency"`
	Attempt      int           `json:"attempt"`
	Success      bool          `json:"success"`
}

func NewWebhookDeliveryService(
	logger *logger.Logger,
	webhookRepo ports.WebhookRepository,
	workers int,
) *WebhookDeliveryService {
	if workers <= 0 {
		workers = 5
	}

	return &WebhookDeliveryService{
		logger:      logger,
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries:    3,
		retryDelay:    2 * time.Second,
		deliveryQueue: make(chan *DeliveryTask, 1000),
		workers:       workers,
	}
}

func (s *WebhookDeliveryService) AddProcessor(processor WebhookEventProcessor) {
	s.processors = append(s.processors, processor)
}

func (s *WebhookDeliveryService) Start(ctx context.Context) {
	s.logger.InfoWithFields("Starting webhook delivery service", map[string]interface{}{
		"workers": s.workers,
	})

	for i := 0; i < s.workers; i++ {
		go s.worker(ctx, i)
	}
}

func (s *WebhookDeliveryService) worker(ctx context.Context, workerID int) {
	s.logger.InfoWithFields("Starting webhook worker", map[string]interface{}{
		"worker_id": workerID,
	})

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoWithFields("Stopping webhook worker", map[string]interface{}{
				"worker_id": workerID,
			})
			return
		case task := <-s.deliveryQueue:
			s.processDeliveryTask(ctx, task, workerID)
		}
	}
}

func (s *WebhookDeliveryService) DeliverEvent(ctx context.Context, event *webhook.WebhookEvent) error {
	logLevel := "DEBUG"
	if event.Type == "Message" || event.Type == "Connected" || event.Type == "Disconnected" {
		logLevel = "INFO"
	}

	if logLevel == "INFO" {
		s.logger.InfoWithFields("Delivering webhook event", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"session_id": event.SessionID,
		})
	} else {
		s.logger.DebugWithFields("Delivering webhook event", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"session_id": event.SessionID,
		})
	}

	for _, processor := range s.processors {
		if err := processor.ProcessWebhookEvent(ctx, event); err != nil {
			s.logger.ErrorWithFields("Processor failed to handle event", map[string]interface{}{
				"event_id":   event.ID,
				"event_type": event.Type,
				"session_id": event.SessionID,
				"error":      err.Error(),
			})
		}
	}

	webhooks, err := s.getWebhooksForEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to get webhooks for event: %w", err)
	}

	if len(webhooks) == 0 {
		s.logger.DebugWithFields("No webhooks configured for event", map[string]interface{}{
			"event_type": event.Type,
			"session_id": event.SessionID,
		})
		return nil
	}

	for _, webhookConfig := range webhooks {
		task := &DeliveryTask{
			WebhookConfig: webhookConfig,
			Event:         event,
			Attempt:       1,
			MaxAttempts:   s.maxRetries,
		}

		select {
		case s.deliveryQueue <- task:
			s.logger.DebugWithFields("Queued webhook delivery task", map[string]interface{}{
				"webhook_id": webhookConfig.ID.String(),
				"event_id":   event.ID,
			})
		default:
			s.logger.WarnWithFields("Webhook delivery queue is full, dropping task", map[string]interface{}{
				"webhook_id": webhookConfig.ID.String(),
				"event_id":   event.ID,
			})
		}
	}

	return nil
}

func (s *WebhookDeliveryService) getWebhooksForEvent(ctx context.Context, event *webhook.WebhookEvent) ([]*webhook.WebhookConfig, error) {
	var webhooks []*webhook.WebhookConfig

	if event.SessionID != "" {
		sessionWebhooks, err := s.webhookRepo.GetBySessionID(ctx, event.SessionID)
		if err != nil {
			s.logger.ErrorWithFields("Failed to get session webhooks", map[string]interface{}{
				"session_id": event.SessionID,
				"error":      err.Error(),
			})
		} else {
			for _, wh := range sessionWebhooks {
				if wh.Enabled && wh.HasEvent(event.Type) {
					webhooks = append(webhooks, wh)
				}
			}
		}
	}

	if len(webhooks) == 0 {
		globalWebhooks, err := s.webhookRepo.GetGlobalWebhooks(ctx)
		if err != nil {
			s.logger.ErrorWithFields("Failed to get global webhooks", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			for _, wh := range globalWebhooks {
				if wh.Enabled && wh.HasEvent(event.Type) {
					webhooks = append(webhooks, wh)
				}
			}
		}
	}

	return webhooks, nil
}

func (s *WebhookDeliveryService) processDeliveryTask(ctx context.Context, task *DeliveryTask, workerID int) {
	s.logger.DebugWithFields("Processing webhook delivery task", map[string]interface{}{
		"worker_id":  workerID,
		"webhook_id": task.WebhookConfig.ID.String(),
		"event_id":   task.Event.ID,
		"attempt":    task.Attempt,
	})

	result := s.deliverWebhook(ctx, task.WebhookConfig, task.Event)

	if !result.Success && task.Attempt < task.MaxAttempts {
		task.Attempt++

		delay := time.Duration(task.Attempt) * s.retryDelay

		s.logger.InfoWithFields("Retrying webhook delivery", map[string]interface{}{
			"webhook_id": task.WebhookConfig.ID.String(),
			"event_id":   task.Event.ID,
			"attempt":    task.Attempt,
			"delay":      delay.String(),
		})

		time.AfterFunc(delay, func() {
			select {
			case s.deliveryQueue <- task:
			default:
				s.logger.WarnWithFields("Failed to requeue webhook delivery task", map[string]interface{}{
					"webhook_id": task.WebhookConfig.ID.String(),
					"event_id":   task.Event.ID,
				})
			}
		})
	} else {
		if result.Success {
			s.logger.InfoWithFields("Webhook delivered successfully", map[string]interface{}{
				"webhook_id":  task.WebhookConfig.ID.String(),
				"event_id":    task.Event.ID,
				"status_code": result.StatusCode,
				"latency":     result.Latency.String(),
				"attempt":     task.Attempt,
			})
		} else {
			s.logger.ErrorWithFields("Webhook delivery failed permanently", map[string]interface{}{
				"webhook_id":  task.WebhookConfig.ID.String(),
				"event_id":    task.Event.ID,
				"error":       result.Error,
				"status_code": result.StatusCode,
				"attempts":    task.Attempt,
			})
		}
	}
}

func (s *WebhookDeliveryService) deliverWebhook(ctx context.Context, webhookConfig *webhook.WebhookConfig, event *webhook.WebhookEvent) *DeliveryResult {
	startTime := time.Now()

	payload := &WebhookPayload{
		Event:     event.Type,
		SessionID: event.SessionID,
		Timestamp: event.Timestamp.Unix(),
		Data:      event.Data,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &DeliveryResult{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal payload: %v", err),
			Latency: time.Since(startTime),
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookConfig.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return &DeliveryResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
			Latency: time.Since(startTime),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "zpwoot-webhook/1.0")
	req.Header.Set("X-Webhook-Event", event.Type)
	req.Header.Set("X-Webhook-Session", event.SessionID)
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", event.Timestamp.Unix()))

	if webhookConfig.Secret != "" {
		signature := s.generateSignature(payloadBytes, webhookConfig.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &DeliveryResult{
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
			Latency: time.Since(startTime),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		responseBody = []byte("failed to read response body")
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return &DeliveryResult{
		Success:      success,
		StatusCode:   resp.StatusCode,
		ResponseBody: string(responseBody),
		Latency:      time.Since(startTime),
		Error:        "",
	}
}

func (s *WebhookDeliveryService) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}
