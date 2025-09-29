package webhook

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WebhookConfig struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SessionID *string   `json:"session_id,omitempty" db:"session_id"` // null for global webhooks
	URL       string    `json:"url" db:"url"`
	Secret    string    `json:"secret,omitempty" db:"secret"`
	Events    []string  `json:"events" db:"events"`
	Enabled   bool      `json:"enabled" db:"enabled"` // User-controlled enable/disable
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

var (
	ErrWebhookNotFound       = errors.New("webhook not found")
	ErrWebhookAlreadyExists  = errors.New("webhook already exists")
	ErrInvalidWebhookURL     = errors.New("invalid webhook URL")
	ErrWebhookDeliveryFailed = errors.New("webhook delivery failed")
)

type SetConfigRequest struct {
	SessionID *string  `json:"session_id,omitempty" validate:"omitempty,uuid"`
	URL       string   `json:"url" validate:"required,url"`
	Secret    string   `json:"secret,omitempty"`
	Events    []string `json:"events" validate:"required,min=1"`
	Enabled   *bool    `json:"enabled,omitempty"`
}

type UpdateWebhookRequest struct {
	URL     *string  `json:"url,omitempty" validate:"omitempty,url"`
	Secret  *string  `json:"secret,omitempty"`
	Events  []string `json:"events,omitempty" validate:"omitempty,min=1"`
	Enabled *bool    `json:"enabled,omitempty"`
}

type ListWebhooksRequest struct {
	SessionID *string `json:"session_id,omitempty" query:"session_id"`
	Enabled   *bool   `json:"enabled,omitempty" query:"enabled"`
	Limit     int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset    int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0"`
}

type WebhookEvent struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

var SupportedEventTypes = []string{
	"Message",
	"UndecryptableMessage",
	"Receipt",
	"MediaRetry",
	"ReadReceipt",

	"GroupInfo",
	"JoinedGroup",
	"Picture",
	"BlocklistChange",
	"Blocklist",

	"Connected",
	"Disconnected",
	"ConnectFailure",
	"KeepAliveRestored",
	"KeepAliveTimeout",
	"LoggedOut",
	"ClientOutdated",
	"TemporaryBan",
	"StreamError",
	"StreamReplaced",
	"PairSuccess",
	"PairError",
	"QR",
	"QRScannedWithoutMultidevice",

	"PrivacySettings",
	"PushNameSetting",
	"UserAbout",

	"AppState",
	"AppStateSyncComplete",
	"HistorySync",
	"OfflineSyncCompleted",
	"OfflineSyncPreview",

	"CallOffer",
	"CallAccept",
	"CallTerminate",
	"CallOfferNotice",
	"CallRelayLatency",

	"Presence",
	"ChatPresence",

	"IdentityChange",

	"CATRefreshError",

	"NewsletterJoin",
	"NewsletterLeave",
	"NewsletterMuteChange",
	"NewsletterLiveUpdate",

	"FBMessage",

	"All",
}

var eventTypeMap map[string]bool

func init() {
	eventTypeMap = make(map[string]bool)
	for _, eventType := range SupportedEventTypes {
		eventTypeMap[eventType] = true
	}
}

func IsValidEventType(eventType string) bool {
	return eventTypeMap[eventType]
}

func ValidateEvents(events []string) []string {
	var invalidEvents []string
	for _, event := range events {
		if !IsValidEventType(event) {
			invalidEvents = append(invalidEvents, event)
		}
	}
	return invalidEvents
}

func NewWebhookConfig(sessionID *string, url, secret string, events []string) *WebhookConfig {
	return &WebhookConfig{
		ID:        uuid.New(),
		SessionID: sessionID,
		URL:       url,
		Secret:    secret,
		Events:    events,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (w *WebhookConfig) IsGlobal() bool {
	return w.SessionID == nil
}

func (w *WebhookConfig) HasEvent(eventType string) bool {
	for _, event := range w.Events {
		if event == "All" || event == eventType {
			return true
		}
	}
	return false
}

func (w *WebhookConfig) Update(req *UpdateWebhookRequest) {
	if req.URL != nil {
		w.URL = *req.URL
	}
	if req.Secret != nil {
		w.Secret = *req.Secret
	}
	if req.Events != nil {
		w.Events = req.Events
	}
	if req.Enabled != nil {
		w.Enabled = *req.Enabled
	}
	w.UpdatedAt = time.Now()
}

func NewWebhookEvent(sessionID, eventType string, data map[string]interface{}) *WebhookEvent {
	return &WebhookEvent{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}
