package wameow

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	webhookDomain "zpwoot/internal/domain/webhook"
	"zpwoot/internal/infra/integrations/webhook"
	"zpwoot/platform/logger"
)

// List of supported event types for webhook delivery
var supportedEventTypes = []string{
	// Messages and Communication
	"Message",
	"UndecryptableMessage",
	"Receipt",
	"MediaRetry",
	"ReadReceipt",

	// Groups and Contacts
	"GroupInfo",
	"JoinedGroup",
	"Picture",
	"BlocklistChange",
	"Blocklist",

	// Connection and Session
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

	// Privacy and Settings
	"PrivacySettings",
	"PushNameSetting",
	"UserAbout",

	// Synchronization and State
	"AppState",
	"AppStateSyncComplete",
	"HistorySync",
	"OfflineSyncCompleted",
	"OfflineSyncPreview",

	// Calls
	"CallOffer",
	"CallAccept",
	"CallTerminate",
	"CallOfferNotice",
	"CallRelayLatency",

	// Presence and Activity
	"Presence",
	"ChatPresence",

	// Identity
	"IdentityChange",

	// Errors
	"CATRefreshError",

	// Newsletter (WhatsApp Channels)
	"NewsletterJoin",
	"NewsletterLeave",
	"NewsletterMuteChange",
	"NewsletterLiveUpdate",

	// Facebook/Meta Bridge
	"FBMessage",

	// Special - receives all events
	"All",
}

// Map for quick validation
var eventTypeMap map[string]bool

func init() {
	eventTypeMap = make(map[string]bool)
	for _, eventType := range supportedEventTypes {
		eventTypeMap[eventType] = true
	}
}

// isValidEventType validates if an event type is supported
func isValidEventType(eventType string) bool {
	return eventTypeMap[eventType]
}

// WhatsmeowWebhookHandler implements the WebhookEventHandler interface
// and delivers raw whatsmeow events to webhook clients
type WhatsmeowWebhookHandler struct {
	logger         *logger.Logger
	webhookManager *webhook.WebhookManager
}

// NewWhatsmeowWebhookHandler creates a new webhook handler for whatsmeow events
func NewWhatsmeowWebhookHandler(logger *logger.Logger, webhookManager *webhook.WebhookManager) *WhatsmeowWebhookHandler {
	return &WhatsmeowWebhookHandler{
		logger:         logger,
		webhookManager: webhookManager,
	}
}

// HandleWhatsmeowEvent implements the WebhookEventHandler interface
// It receives raw whatsmeow events and delivers them to webhook clients
func (h *WhatsmeowWebhookHandler) HandleWhatsmeowEvent(evt interface{}, sessionID string) error {
	if h.webhookManager == nil {
		h.logger.Debug("Webhook manager not available, skipping event delivery")
		return nil
	}

	eventType := h.getEventType(evt)

	// Validate event type
	if !isValidEventType(eventType) {
		h.logger.DebugWithFields("Skipping unsupported event type for webhook", map[string]interface{}{
			"event_type": eventType,
			"session_id": sessionID,
		})
		return nil
	}

	h.logger.DebugWithFields("Processing whatsmeow event for webhook delivery", map[string]interface{}{
		"event_type": eventType,
		"session_id": sessionID,
	})

	// Convert event to raw data (no normalization, just JSON serialization)
	eventData, err := h.convertEventToRawData(evt)
	if err != nil {
		h.logger.ErrorWithFields("Failed to convert event to raw data", map[string]interface{}{
			"event_type": eventType,
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to convert event to raw data: %w", err)
	}

	// Create webhook payload in the specified format
	webhookPayload := map[string]interface{}{
		"event":     eventType,
		"sessionID": sessionID,
		"timestamp": time.Now().Unix(),
		"data":      eventData,
	}

	// Create webhook event with the payload as data
	webhookEvent := webhookDomain.NewWebhookEvent(sessionID, eventType, webhookPayload)

	// Use the delivery service directly to deliver the event
	return h.webhookManager.GetDeliveryService().DeliverEvent(context.Background(), webhookEvent)
}

// convertEventToRawData converts a whatsmeow event to raw data without normalization
func (h *WhatsmeowWebhookHandler) convertEventToRawData(evt interface{}) (map[string]interface{}, error) {
	// Marshal to JSON to handle all complex types and nested structures
	jsonBytes, err := json.Marshal(evt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	// Unmarshal to map to get raw data structure
	var eventData map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return eventData, nil
}

// getEventType extracts the event type name using reflection
func (h *WhatsmeowWebhookHandler) getEventType(evt interface{}) string {
	if evt == nil {
		return "nil"
	}

	eventType := reflect.TypeOf(evt)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	// Extract just the type name without package prefix
	typeName := eventType.Name()
	if typeName == "" {
		// Fallback to full type string and clean it
		fullType := eventType.String()
		typeName = strings.TrimPrefix(fullType, "*events.")
		typeName = strings.TrimPrefix(typeName, "events.")
	}

	return typeName
}
