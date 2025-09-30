package ports

import (
	"context"

	"zpwoot/internal/domain/webhook"
)

// WebhookService defines the interface for webhook operations
type WebhookService interface {
	SendWebhook(url string, payload interface{}) error
	RegisterWebhook(sessionID, url, secret string, events []string) error
	UnregisterWebhook(sessionID, url string) error
	GetWebhooks(sessionID string) ([]*WebhookRegistration, error)
}

// WebhookRepository defines the interface for webhook data operations
type WebhookRepository interface {
	Create(ctx context.Context, webhook *webhook.WebhookConfig) error
	GetByID(ctx context.Context, id string) (*webhook.WebhookConfig, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*webhook.WebhookConfig, error)
	GetGlobalWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error)
	List(ctx context.Context, req *webhook.ListWebhooksRequest) ([]*webhook.WebhookConfig, int, error)
	Update(ctx context.Context, webhook *webhook.WebhookConfig) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, enabled bool) error
	GetEnabledWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error)
	GetWebhooksByEvent(ctx context.Context, eventType string) ([]*webhook.WebhookConfig, error)
	CountByStatus(ctx context.Context, enabled bool) (int, error)
	GetWebhookStats(ctx context.Context, webhookID string) (*WebhookStats, error)
	UpdateWebhookStats(ctx context.Context, webhookID string, stats *WebhookStats) error
}

// WebhookDeliveryRepository defines the interface for webhook delivery operations
type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery *WebhookDelivery) error
	GetByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*WebhookDelivery, error)
	GetByEventID(ctx context.Context, eventID string) ([]*WebhookDelivery, error)
	GetFailedDeliveries(ctx context.Context, limit int) ([]*WebhookDelivery, error)
	UpdateDeliveryStatus(ctx context.Context, deliveryID string, success bool, statusCode int, responseBody, error string) error
	DeleteOldDeliveries(ctx context.Context, olderThan int64) error
	GetDeliveryStats(ctx context.Context, webhookID string, from, to int64) (*DeliveryStats, error)
}

// WebhookRegistration represents a webhook registration
type WebhookRegistration struct {
	ID        string   `json:"id"`
	SessionID string   `json:"session_id"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
}

// WebhookStats represents statistics for webhook operations
type WebhookStats struct {
	WebhookID       string `json:"webhook_id" db:"webhook_id"`
	TotalDeliveries int64  `json:"total_deliveries" db:"total_deliveries"`
	SuccessCount    int64  `json:"success_count" db:"success_count"`
	FailureCount    int64  `json:"failure_count" db:"failure_count"`
	LastDelivery    int64  `json:"last_delivery" db:"last_delivery"`
	LastSuccess     int64  `json:"last_success" db:"last_success"`
	LastFailure     int64  `json:"last_failure" db:"last_failure"`
	AverageLatency  int64  `json:"average_latency" db:"average_latency"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string `json:"id" db:"id"`
	WebhookID    string `json:"webhook_id" db:"webhook_id"`
	EventID      string `json:"event_id" db:"event_id"`
	URL          string `json:"url" db:"url"`
	Payload      string `json:"payload" db:"payload"`
	ResponseBody string `json:"response_body" db:"response_body"`
	Error        string `json:"error,omitempty" db:"error"`
	StatusCode   int    `json:"status_code" db:"status_code"`
	Latency      int64  `json:"latency" db:"latency"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
	Success      bool   `json:"success" db:"success"`
}

// DeliveryStats represents delivery statistics for webhooks
type DeliveryStats struct {
	WebhookID       string  `json:"webhook_id"`
	TotalDeliveries int64   `json:"total_deliveries"`
	SuccessCount    int64   `json:"success_count"`
	FailureCount    int64   `json:"failure_count"`
	SuccessRate     float64 `json:"success_rate"`
	AverageLatency  float64 `json:"average_latency"`
	From            int64   `json:"from"`
	To              int64   `json:"to"`
}
