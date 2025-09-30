package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"zpwoot/internal/domain/webhook"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow/types/events"
)

// EventDispatcher converts whatsmeow events to webhook events and dispatches them
type EventDispatcher struct {
	logger          *logger.Logger
	deliveryService *WebhookDeliveryService
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher(logger *logger.Logger, deliveryService *WebhookDeliveryService) *EventDispatcher {
	return &EventDispatcher{
		logger:          logger,
		deliveryService: deliveryService,
	}
}

// DispatchEvent converts and dispatches a whatsmeow event
func (d *EventDispatcher) DispatchEvent(ctx context.Context, evt interface{}, sessionID string) error {
	eventType := d.getEventType(evt)

	// Skip AppState events to reduce webhook spam - they're too frequent and not critical
	if eventType == "AppState" {
		return nil
	}

	// Skip if event type is not supported
	if !webhook.IsValidEventType(eventType) {
		d.logger.DebugWithFields("Skipping unsupported event type", map[string]interface{}{
			"event_type": eventType,
			"session_id": sessionID,
		})
		return nil
	}

	// Convert event to map for JSON serialization
	eventData, err := d.convertEventToMap(evt)
	if err != nil {
		d.logger.ErrorWithFields("Failed to convert event to map", map[string]interface{}{
			"event_type": eventType,
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to convert event: %w", err)
	}

	// Create webhook event
	webhookEvent := webhook.NewWebhookEvent(sessionID, eventType, eventData)

	d.logger.DebugWithFields("Dispatching webhook event", map[string]interface{}{
		"event_id":   webhookEvent.ID,
		"event_type": eventType,
		"session_id": sessionID,
	})

	// Deliver the event
	return d.deliveryService.DeliverEvent(ctx, webhookEvent)
}

// getEventType extracts the event type name from the event interface
func (d *EventDispatcher) getEventType(evt interface{}) string {
	eventType := reflect.TypeOf(evt)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	// Remove the package prefix (e.g., "events.Message" -> "Message")
	typeName := eventType.Name()
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		typeName = parts[len(parts)-1]
	}

	return typeName
}

// convertEventToMap converts a whatsmeow event to a map for JSON serialization
func (d *EventDispatcher) convertEventToMap(evt interface{}) (map[string]interface{}, error) {
	// First, marshal to JSON to handle all the complex types
	jsonBytes, err := json.Marshal(evt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	// Then unmarshal to map
	var eventMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &eventMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	// Add some metadata based on event type
	eventMap = d.enrichEventData(evt, eventMap)

	return eventMap, nil
}

// enrichEventData adds additional metadata to event data based on event type
func (d *EventDispatcher) enrichEventData(evt interface{}, eventMap map[string]interface{}) map[string]interface{} {
	switch v := evt.(type) {
	case *events.Message:
		d.enrichMessageEvent(v, eventMap)
	case *events.Receipt:
		d.enrichReceiptEvent(v, eventMap)
	case *events.Connected:
		eventMap["connected_at"] = time.Now().Unix()
	case *events.Disconnected:
		eventMap["disconnected_at"] = time.Now().Unix()
	case *events.QR:
		d.enrichQREvent(v, eventMap)
	case *events.GroupInfo:
		d.enrichGroupInfoEvent(v, eventMap)
	case *events.Presence:
		d.enrichPresenceEvent(v, eventMap)
	case *events.ChatPresence:
		d.enrichChatPresenceEvent(v, eventMap)
	case *events.PairSuccess:
		d.enrichPairSuccessEvent(v, eventMap)
	case *events.PairError:
		d.enrichPairErrorEvent(v, eventMap)
	case *events.LoggedOut:
		d.enrichLoggedOutEvent(v, eventMap)
	case *events.KeepAliveTimeout:
		eventMap["timeout_at"] = time.Now().Unix()
	case *events.KeepAliveRestored:
		eventMap["restored_at"] = time.Now().Unix()
	case *events.UndecryptableMessage:
		d.enrichUndecryptableMessageEvent(v, eventMap)
	case *events.Picture:
		d.enrichPictureEvent(v, eventMap)
	case *events.JoinedGroup:
		d.enrichJoinedGroupEvent(v, eventMap)
	}

	return eventMap
}

func (d *EventDispatcher) enrichMessageEvent(v *events.Message, eventMap map[string]interface{}) {
	eventMap["message_id"] = v.Info.ID
	eventMap["from_me"] = v.Info.IsFromMe
	eventMap["chat"] = v.Info.Chat.String()
	eventMap["sender"] = v.Info.Sender.String()
	eventMap["timestamp"] = v.Info.Timestamp.Unix()

	d.setMessageType(v, eventMap)
}

func (d *EventDispatcher) setMessageType(v *events.Message, eventMap map[string]interface{}) {
	switch {
	case v.Message.Conversation != nil:
		eventMap["message_type"] = "text"
		eventMap["text"] = *v.Message.Conversation
	case v.Message.ImageMessage != nil:
		eventMap["message_type"] = "image"
		if v.Message.ImageMessage.Caption != nil {
			eventMap["caption"] = *v.Message.ImageMessage.Caption
		}
	case v.Message.AudioMessage != nil:
		eventMap["message_type"] = "audio"
	case v.Message.VideoMessage != nil:
		eventMap["message_type"] = "video"
		if v.Message.VideoMessage.Caption != nil {
			eventMap["caption"] = *v.Message.VideoMessage.Caption
		}
	case v.Message.DocumentMessage != nil:
		eventMap["message_type"] = "document"
		if v.Message.DocumentMessage.FileName != nil {
			eventMap["filename"] = *v.Message.DocumentMessage.FileName
		}
	case v.Message.StickerMessage != nil:
		eventMap["message_type"] = "sticker"
	case v.Message.LocationMessage != nil:
		eventMap["message_type"] = "location"
	case v.Message.ContactMessage != nil:
		eventMap["message_type"] = "contact"
	default:
		eventMap["message_type"] = "unknown"
	}
}

func (d *EventDispatcher) enrichReceiptEvent(v *events.Receipt, eventMap map[string]interface{}) {
	eventMap["message_ids"] = v.MessageIDs
	eventMap["chat"] = v.Chat.String()
	eventMap["sender"] = v.Sender.String()
	eventMap["timestamp"] = v.Timestamp.Unix()
	eventMap["receipt_type"] = string(v.Type)
}

func (d *EventDispatcher) enrichQREvent(v *events.QR, eventMap map[string]interface{}) {
	eventMap["codes_count"] = len(v.Codes)
	delete(eventMap, "codes") // Remove for security
}

func (d *EventDispatcher) enrichGroupInfoEvent(v *events.GroupInfo, eventMap map[string]interface{}) {
	eventMap["group_jid"] = v.JID.String()
	if v.Name != nil {
		eventMap["group_name"] = *v.Name
	}
	if v.Topic != nil {
		eventMap["group_topic"] = *v.Topic
	}
}

func (d *EventDispatcher) enrichPresenceEvent(v *events.Presence, eventMap map[string]interface{}) {
	eventMap["from"] = v.From.String()
	eventMap["unavailable"] = v.Unavailable
	if !v.LastSeen.IsZero() {
		eventMap["last_seen"] = v.LastSeen.Unix()
	}
}

func (d *EventDispatcher) enrichChatPresenceEvent(v *events.ChatPresence, eventMap map[string]interface{}) {
	eventMap["chat"] = v.Chat.String()
	eventMap["state"] = string(v.State)
	if v.Media != "" {
		eventMap["media"] = string(v.Media)
	}
}

func (d *EventDispatcher) enrichPairSuccessEvent(v *events.PairSuccess, eventMap map[string]interface{}) {
	eventMap["paired_at"] = time.Now().Unix()
	eventMap["device_id"] = v.ID.String()
}

func (d *EventDispatcher) enrichPairErrorEvent(v *events.PairError, eventMap map[string]interface{}) {
	eventMap["error_at"] = time.Now().Unix()
	eventMap["error_message"] = v.Error.Error()
}

func (d *EventDispatcher) enrichLoggedOutEvent(v *events.LoggedOut, eventMap map[string]interface{}) {
	eventMap["logged_out_at"] = time.Now().Unix()
	eventMap["reason"] = fmt.Sprintf("%d", v.Reason)
}

func (d *EventDispatcher) enrichUndecryptableMessageEvent(v *events.UndecryptableMessage, eventMap map[string]interface{}) {
	eventMap["message_id"] = v.Info.ID
	eventMap["chat"] = v.Info.Chat.String()
	eventMap["sender"] = v.Info.Sender.String()
	eventMap["timestamp"] = v.Info.Timestamp.Unix()
	eventMap["is_unavailable"] = v.IsUnavailable
}

func (d *EventDispatcher) enrichPictureEvent(v *events.Picture, eventMap map[string]interface{}) {
	eventMap["jid"] = v.JID.String()
	eventMap["author"] = v.Author.String()
	eventMap["timestamp"] = v.Timestamp.Unix()
	eventMap["remove"] = v.Remove
}

func (d *EventDispatcher) enrichJoinedGroupEvent(v *events.JoinedGroup, eventMap map[string]interface{}) {
	eventMap["group_jid"] = v.JID.String()
	eventMap["reason"] = v.Reason
	eventMap["type"] = v.Type
}

// Event type mapping for whatsmeow events to webhook event names
var eventTypeMapping = map[string]string{
	"*events.Message":                     "Message",
	"*events.UndecryptableMessage":        "UndecryptableMessage",
	"*events.Receipt":                     "Receipt",
	"*events.MediaRetry":                  "MediaRetry",
	"*events.ReadReceipt":                 "ReadReceipt",
	"*events.GroupInfo":                   "GroupInfo",
	"*events.JoinedGroup":                 "JoinedGroup",
	"*events.Picture":                     "Picture",
	"*events.BlocklistChange":             "BlocklistChange",
	"*events.Blocklist":                   "Blocklist",
	"*events.Connected":                   "Connected",
	"*events.Disconnected":                "Disconnected",
	"*events.ConnectFailure":              "ConnectFailure",
	"*events.KeepAliveRestored":           "KeepAliveRestored",
	"*events.KeepAliveTimeout":            "KeepAliveTimeout",
	"*events.LoggedOut":                   "LoggedOut",
	"*events.ClientOutdated":              "ClientOutdated",
	"*events.TemporaryBan":                "TemporaryBan",
	"*events.StreamError":                 "StreamError",
	"*events.StreamReplaced":              "StreamReplaced",
	"*events.PairSuccess":                 "PairSuccess",
	"*events.PairError":                   "PairError",
	"*events.QR":                          "QR",
	"*events.QRScannedWithoutMultidevice": "QRScannedWithoutMultidevice",
	"*events.PrivacySettings":             "PrivacySettings",
	"*events.PushNameSetting":             "PushNameSetting",
	"*events.UserAbout":                   "UserAbout",
	"*events.AppState":                    "AppState",
	"*events.AppStateSyncComplete":        "AppStateSyncComplete",
	"*events.HistorySync":                 "HistorySync",
	"*events.OfflineSyncCompleted":        "OfflineSyncCompleted",
	"*events.OfflineSyncPreview":          "OfflineSyncPreview",
	"*events.CallOffer":                   "CallOffer",
	"*events.CallAccept":                  "CallAccept",
	"*events.CallTerminate":               "CallTerminate",
	"*events.CallOfferNotice":             "CallOfferNotice",
	"*events.CallRelayLatency":            "CallRelayLatency",
	"*events.Presence":                    "Presence",
	"*events.ChatPresence":                "ChatPresence",
	"*events.IdentityChange":              "IdentityChange",
	"*events.CATRefreshError":             "CATRefreshError",
	"*events.NewsletterJoin":              "NewsletterJoin",
	"*events.NewsletterLeave":             "NewsletterLeave",
	"*events.NewsletterMuteChange":        "NewsletterMuteChange",
	"*events.NewsletterLiveUpdate":        "NewsletterLiveUpdate",
	"*events.FBMessage":                   "FBMessage",
}

// GetEventTypeFromInterface returns the webhook event type for a whatsmeow event
func (d *EventDispatcher) GetEventTypeFromInterface(evt interface{}) string {
	fullType := fmt.Sprintf("%T", evt)
	if mappedType, exists := eventTypeMapping[fullType]; exists {
		return mappedType
	}

	// Fallback to extracting from type name
	return d.getEventType(evt)
}
