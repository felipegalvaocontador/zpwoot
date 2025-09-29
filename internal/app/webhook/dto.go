package webhook

import (
	"time"

	"zpwoot/internal/domain/webhook"
)

type SetConfigRequest struct {
	SessionID *string  `json:"sessionId,omitempty" validate:"omitempty,uuid" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	URL       string   `json:"url" validate:"required,url" example:"https://myapp.com/webhook/whatsapp"`
	Secret    string   `json:"secret,omitempty" example:"my-webhook-secret-key-123"`
	Events    []string `json:"events" validate:"required,min=1" example:"message,status,connection"`
	Enabled   *bool    `json:"enabled,omitempty" example:"true"` // Whether webhook is enabled (default: true)
} // @name SetConfigRequest

type SetConfigResponse struct {
	ID        string    `json:"id" example:"webhook-456def"`
	SessionID *string   `json:"sessionId,omitempty" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	URL       string    `json:"url" example:"https://myapp.com/webhook/whatsapp"`
	Events    []string  `json:"events" example:"message,status,connection"`
	Enabled   bool      `json:"enabled" example:"true"` // Whether webhook is enabled
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} // @name SetConfigResponse

type UpdateWebhookRequest struct {
	URL     *string  `json:"url,omitempty" validate:"omitempty,url" example:"https://myapp.com/webhook/whatsapp/v2"`
	Secret  *string  `json:"secret,omitempty" example:"updated-webhook-secret-456"`
	Events  []string `json:"events,omitempty" validate:"omitempty,min=1" example:"message,status,connection,qr"`
	Enabled *bool    `json:"enabled,omitempty" example:"false"` // Whether webhook is enabled
} // @name UpdateWebhookRequest

type ListWebhooksRequest struct {
	SessionID *string `json:"sessionId,omitempty" query:"sessionId" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	Enabled   *bool   `json:"enabled,omitempty" query:"enabled" example:"true"` // Filter by enabled status
	Limit     int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100" example:"20"`
	Offset    int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0" example:"0"`
} // @name ListWebhooksRequest

type ListWebhooksResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"`
	Total    int               `json:"total" example:"5"`
	Limit    int               `json:"limit" example:"20"`
	Offset   int               `json:"offset" example:"0"`
} // @name ListWebhooksResponse

type WebhookResponse struct {
	ID        string    `json:"id" example:"webhook-123"`
	SessionID *string   `json:"sessionId,omitempty" example:"session-123"`
	URL       string    `json:"url" example:"https://example.com/webhook"`
	Events    []string  `json:"events" example:"message,status"`
	Enabled   bool      `json:"enabled" example:"true"` // Whether webhook is enabled
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
} // @name WebhookResponse

type WebhookEventResponse struct {
	ID        string                 `json:"id" example:"event-123"`
	SessionID string                 `json:"sessionId" example:"session-123"`
	Type      string                 `json:"type" example:"message"`
	Timestamp time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Data      map[string]interface{} `json:"data"`
} // @name WebhookEventResponse

type TestWebhookRequest struct {
	EventType string                 `json:"eventType" validate:"required" example:"message"`
	TestData  map[string]interface{} `json:"testData,omitempty"`
} // @name TestWebhookRequest

type TestWebhookResponse struct {
	Success      bool   `json:"success" example:"true"`
	StatusCode   int    `json:"statusCode" example:"200"`
	ResponseTime int64  `json:"responseTimeMs" example:"150"`
	Error        string `json:"error,omitempty"`
}

type WebhookEventsResponse struct {
	Events []WebhookEventInfo `json:"events"`
}

type WebhookEventInfo struct {
	Type        string `json:"type" example:"message"`
	Description string `json:"description" example:"Triggered when a message is received or sent"`
	DataSchema  string `json:"data_schema,omitempty" example:"MessageEventData"`
}

func (r *SetConfigRequest) ToSetConfigRequest() *webhook.SetConfigRequest {
	return &webhook.SetConfigRequest{
		SessionID: r.SessionID,
		URL:       r.URL,
		Secret:    r.Secret,
		Events:    r.Events,
		Enabled:   r.Enabled,
	}
}

func (r *UpdateWebhookRequest) ToUpdateWebhookRequest() *webhook.UpdateWebhookRequest {
	return &webhook.UpdateWebhookRequest{
		URL:     r.URL,
		Secret:  r.Secret,
		Events:  r.Events,
		Enabled: r.Enabled,
	}
}

func (r *ListWebhooksRequest) ToListWebhooksRequest() *webhook.ListWebhooksRequest {
	return &webhook.ListWebhooksRequest{
		SessionID: r.SessionID,
		Enabled:   r.Enabled,
		Limit:     r.Limit,
		Offset:    r.Offset,
	}
}

func FromWebhook(w *webhook.WebhookConfig) *WebhookResponse {
	return &WebhookResponse{
		ID:        w.ID.String(),
		SessionID: w.SessionID,
		URL:       w.URL,
		Events:    w.Events,
		Enabled:   w.Enabled,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func FromWebhookEvent(we *webhook.WebhookEvent) *WebhookEventResponse {
	return &WebhookEventResponse{
		ID:        we.ID,
		SessionID: we.SessionID,
		Type:      we.Type,
		Timestamp: we.Timestamp,
		Data:      we.Data,
	}
}

func GetSupportedEvents() *WebhookEventsResponse {
	return &WebhookEventsResponse{
		Events: []WebhookEventInfo{
			{
				Type:        "message",
				Description: "Triggered when a message is received or sent",
				DataSchema:  "MessageEventData",
			},
			{
				Type:        "status",
				Description: "Triggered when message status changes (sent, delivered, read)",
				DataSchema:  "StatusEventData",
			},
			{
				Type:        "connection",
				Description: "Triggered when connection status changes",
				DataSchema:  "ConnectionEventData",
			},
			{
				Type:        "qr",
				Description: "Triggered when QR code is generated",
				DataSchema:  "QREventData",
			},
			{
				Type:        "pair",
				Description: "Triggered when phone pairing is successful",
				DataSchema:  "PairEventData",
			},
		},
	}
}
