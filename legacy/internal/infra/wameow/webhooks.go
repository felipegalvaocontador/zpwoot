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

var supportedEventTypes = []string{
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
	for _, eventType := range supportedEventTypes {
		eventTypeMap[eventType] = true
	}
}

func isValidEventType(eventType string) bool {
	return eventTypeMap[eventType]
}

type WhatsmeowWebhookHandler struct {
	logger         *logger.Logger
	webhookManager *webhook.WebhookManager
}

func NewWhatsmeowWebhookHandler(logger *logger.Logger, webhookManager *webhook.WebhookManager) *WhatsmeowWebhookHandler {
	return &WhatsmeowWebhookHandler{
		logger:         logger,
		webhookManager: webhookManager,
	}
}

func (h *WhatsmeowWebhookHandler) HandleWhatsmeowEvent(evt interface{}, sessionID string) error {
	if h.webhookManager == nil {
		h.logger.Debug("Webhook manager not available, skipping event delivery")
		return nil
	}

	eventType := h.getEventType(evt)

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

	eventData, err := h.convertEventToRawData(evt)
	if err != nil {
		h.logger.ErrorWithFields("Failed to convert event to raw data", map[string]interface{}{
			"event_type": eventType,
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to convert event to raw data: %w", err)
	}

	webhookPayload := map[string]interface{}{
		"event":     eventType,
		"sessionID": sessionID,
		"timestamp": time.Now().Unix(),
		"data":      eventData,
	}

	webhookEvent := webhookDomain.NewWebhookEvent(sessionID, eventType, webhookPayload)

	return h.webhookManager.GetDeliveryService().DeliverEvent(context.Background(), webhookEvent)
}

func (h *WhatsmeowWebhookHandler) convertEventToRawData(evt interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(evt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	var eventData map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return eventData, nil
}

func (h *WhatsmeowWebhookHandler) getEventType(evt interface{}) string {
	if evt == nil {
		return "nil"
	}

	eventType := reflect.TypeOf(evt)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	typeName := eventType.Name()
	if typeName == "" {
		fullType := eventType.String()
		typeName = strings.TrimPrefix(fullType, "*events.")
		typeName = strings.TrimPrefix(typeName, "events.")
	}

	return typeName
}
