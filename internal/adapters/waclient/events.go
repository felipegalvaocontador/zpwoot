package waclient

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"

	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/session"
	"zpwoot/platform/logger"
)

type EventHandler struct {
	gateway     *Gateway
	sessionName string
	logger      *logger.Logger

	qrGenerator *QRGenerator

	webhookHandler  WebhookEventHandler
	chatwootManager ChatwootManager
}

type WebhookEventHandler interface {
	HandleWhatsmeowEvent(evt interface{}, sessionID string) error
}

type ChatwootManager interface {
	IsEnabled(sessionID string) bool
	ProcessWhatsAppMessage(sessionID, messageID, from, content, messageType string, timestamp time.Time, fromMe bool) error
}

func NewEventHandler(gateway *Gateway, sessionName string, logger *logger.Logger) *EventHandler {
	return &EventHandler{
		gateway:     gateway,
		sessionName: sessionName,
		logger:      logger,
		qrGenerator: NewQRGenerator(logger),
	}
}

func (h *EventHandler) SetWebhookHandler(handler WebhookEventHandler) {
	h.webhookHandler = handler
}

func (h *EventHandler) SetChatwootManager(manager ChatwootManager) {
	h.chatwootManager = manager
}

func (h *EventHandler) HandleEvent(evt interface{}, sessionID string) {

	h.deliverToWebhook(evt, sessionID)

	h.handleEventInternal(evt, sessionID)
}

func (h *EventHandler) handleEventInternal(evt interface{}, sessionID string) {
	switch v := evt.(type) {
	case *events.Connected:
		h.handleConnected(v, sessionID)
	case *events.Disconnected:
		h.handleDisconnected(v, sessionID)
	case *events.LoggedOut:
		h.handleLoggedOut(v, sessionID)
	case *events.QR:
		h.handleQREvent(sessionID)
	case *QRCodeEvent:
		h.handleQRCodeEvent(v, sessionID)
	case *events.PairSuccess:
		h.handlePairSuccess(v, sessionID)
	case *events.PairError:
		h.handlePairError(v, sessionID)
	case *events.Message:
		h.handleMessage(v, sessionID)
	case *events.Receipt:
		h.handleReceipt(v, sessionID)
	default:
		h.handleOtherEvents(evt, sessionID)
	}
}

func (h *EventHandler) deliverToWebhook(evt interface{}, sessionID string) {
	if h.webhookHandler == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.ErrorWithFields("Webhook handler panic", map[string]interface{}{
					"session_id": sessionID,
					"error":      r,
				})
			}
		}()

		if err := h.webhookHandler.HandleWhatsmeowEvent(evt, sessionID); err != nil {
			h.logger.ErrorWithFields("Failed to deliver event to webhook", map[string]interface{}{
				"session_id": sessionID,
				"event_type": fmt.Sprintf("%T", evt),
				"error":      err.Error(),
			})
		}
	}()
}

func (h *EventHandler) handleConnected(_ *events.Connected, sessionID string) {
	h.logger.InfoWithFields("WhatsApp connected", map[string]interface{}{
		"module":     "events",
		"session_id": sessionID,
	})


	h.notifySessionConnected(sessionID)

	if err := h.gateway.UpdateSessionStatus(sessionID, "connected"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "connected",
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) handleDisconnected(_ *events.Disconnected, sessionID string) {
	h.logger.WarnWithFields("WhatsApp disconnected", map[string]interface{}{
		"session_id": sessionID,
	})


	h.notifySessionDisconnected(sessionID, "disconnected")

	if err := h.gateway.UpdateSessionStatus(sessionID, "disconnected"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "disconnected",
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) handleLoggedOut(evt *events.LoggedOut, sessionID string) {
	h.logger.WarnWithFields("WhatsApp logged out", map[string]interface{}{
		"session_id": sessionID,
		"reason":     evt.Reason,
	})

	if err := h.gateway.UpdateSessionStatus(sessionID, "logged_out"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "logged_out",
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) handleQREvent(sessionID string) {
	h.logger.InfoWithFields("QR code event received", map[string]interface{}{
		"session_id": sessionID,
	})

	if err := h.gateway.UpdateSessionStatus(sessionID, "qr_code"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "qr_code",
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) handleQRCodeEvent(evt *QRCodeEvent, sessionID string) {
	h.logger.InfoWithFields("QR code event with data received", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": evt.SessionName,
		"qr_length":    len(evt.QRCode),
		"expires_at":   evt.ExpiresAt,
	})

	h.qrGenerator.DisplayQRCodeInTerminal(evt.QRCode, evt.SessionName)

	if err := h.gateway.UpdateSessionStatus(sessionID, "qr_code"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "qr_code",
			"error":      err.Error(),
		})
	}

	if err := h.gateway.UpdateSessionQRCode(sessionID, evt.QRCode, evt.ExpiresAt); err != nil {
		h.logger.ErrorWithFields("Failed to update QR code in database", map[string]interface{}{
			"session_id": sessionID,
			"qr_length":  len(evt.QRCode),
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) handlePairSuccess(evt *events.PairSuccess, sessionID string) {
	deviceJID := evt.ID.String()

	h.logger.InfoWithFields("WhatsApp pairing successful", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": deviceJID,
	})

	if err := h.gateway.UpdateSessionDeviceJID(sessionID, deviceJID); err != nil {
		h.logger.ErrorWithFields("Failed to update session device JID", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": deviceJID,
			"error":      err.Error(),
		})
	}

	if err := h.gateway.UpdateSessionStatus(sessionID, "connected"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status after pairing", map[string]interface{}{
			"session_id": sessionID,
			"status":     "connected",
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) handlePairError(evt *events.PairError, sessionID string) {
	h.logger.ErrorWithFields("WhatsApp pairing failed", map[string]interface{}{
		"session_id": sessionID,
		"error":      evt.Error.Error(),
	})

}

func (h *EventHandler) handleMessage(evt *events.Message, sessionID string) {
	h.logger.InfoWithFields("Message received", map[string]interface{}{
		"module":  "events",
		"type":    evt.Info.Type,
		"from_me": evt.Info.IsFromMe,
	})

	if err := h.saveMessageToDatabase(evt, sessionID); err != nil {
		h.logger.ErrorWithFields("Failed to save message to database", map[string]interface{}{
			"session_id": sessionID,
			"message_id": evt.Info.ID,
			"error":      err.Error(),
		})
	}

	if h.chatwootManager != nil && h.chatwootManager.IsEnabled(sessionID) {
		h.processMessageForChatwoot(evt, sessionID)
	}
}

func (h *EventHandler) handleReceipt(evt *events.Receipt, sessionID string) {
	h.logger.DebugWithFields("Receipt received", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Type,
		"sender":     evt.Sender.String(),
		"timestamp":  evt.Timestamp,
	})

}

func (h *EventHandler) handleOtherEvents(evt interface{}, sessionID string) {
	switch v := evt.(type) {
	case *events.Presence:
		h.handlePresence(v, sessionID)
	case *events.ChatPresence:
		h.handleChatPresence(v, sessionID)
	case *events.HistorySync:
		h.handleHistorySync(v, sessionID)
	case *events.AppState:
		h.handleAppState(v)
	case *events.AppStateSyncComplete:
		h.handleAppStateSyncComplete(v, sessionID)
	case *events.KeepAliveTimeout:
		h.handleKeepAliveTimeout(v, sessionID)
	case *events.KeepAliveRestored:
		h.handleKeepAliveRestored(v, sessionID)
	case *events.Contact:
		h.handleContact(v, sessionID)
	case *events.GroupInfo:
		h.handleGroupInfo(v, sessionID)
	case *events.Picture:
		h.handlePicture(v, sessionID)
	case *events.BusinessName:
		h.handleBusinessName(v, sessionID)
	default:
		h.logger.DebugWithFields("Unhandled event", map[string]interface{}{
			"session_id": sessionID,
			"event_type": reflect.TypeOf(evt).String(),
		})
	}
}

func (h *EventHandler) processMessageForChatwoot(evt *events.Message, sessionID string) {
	messageID := evt.Info.ID
	from := evt.Info.Sender.String()
	timestamp := evt.Info.Timestamp
	fromMe := evt.Info.IsFromMe

	content, messageType := h.extractMessageContentString(evt.Message)

	contactNumber := h.extractContactNumber(from)

	h.logger.DebugWithFields("Processing message for Chatwoot", map[string]interface{}{
		"session_id":     sessionID,
		"message_id":     messageID,
		"contact_number": contactNumber,
		"message_type":   messageType,
		"from_me":        fromMe,
	})

	err := h.chatwootManager.ProcessWhatsAppMessage(sessionID, messageID, contactNumber, content, messageType, timestamp, fromMe)
	if err != nil {
		h.logger.ErrorWithFields("Failed to process message for Chatwoot", map[string]interface{}{
			"session_id": sessionID,
			"message_id": messageID,
			"error":      err.Error(),
		})
	} else {
		h.logger.DebugWithFields("Message processed for Chatwoot", map[string]interface{}{
			"session_id":   sessionID,
			"message_id":   messageID,
			"message_type": messageType,
		})
	}
}

func (h *EventHandler) extractMessageContentString(message *waE2E.Message) (string, string) {
	if message == nil {
		return "", "unknown"
	}

	if message.Conversation != nil {
		return *message.Conversation, "text"
	}

	if message.ExtendedTextMessage != nil && message.ExtendedTextMessage.Text != nil {
		return *message.ExtendedTextMessage.Text, "text"
	}

	if message.ImageMessage != nil {
		caption := ""
		if message.ImageMessage.Caption != nil {
			caption = *message.ImageMessage.Caption
		}
		return caption, "image"
	}

	if message.AudioMessage != nil {
		return "[Audio]", "audio"
	}

	if message.VideoMessage != nil {
		caption := ""
		if message.VideoMessage.Caption != nil {
			caption = *message.VideoMessage.Caption
		}
		return caption, "video"
	}

	if message.DocumentMessage != nil {
		filename := ""
		if message.DocumentMessage.FileName != nil {
			filename = *message.DocumentMessage.FileName
		}
		return fmt.Sprintf("[Document: %s]", filename), "document"
	}

	if message.StickerMessage != nil {
		return "[Sticker]", "sticker"
	}

	if message.LocationMessage != nil {
		return "[Location]", "location"
	}

	if message.ContactMessage != nil {
		name := ""
		if message.ContactMessage.DisplayName != nil {
			name = *message.ContactMessage.DisplayName
		}
		return fmt.Sprintf("[Contact: %s]", name), "contact"
	}

	return "[Unknown message type]", "unknown"
}

func (h *EventHandler) extractContactNumber(jid string) string {

	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}

func (h *EventHandler) handlePresence(evt *events.Presence, sessionID string) {
	h.logger.DebugWithFields("Presence update", map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.From.String(),
		"presence":   evt.Unavailable,
	})
}

func (h *EventHandler) handleChatPresence(evt *events.ChatPresence, sessionID string) {
	h.logger.DebugWithFields("Chat presence update", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.Chat.String(),
		"state":      evt.State,
	})
}

func (h *EventHandler) handleHistorySync(evt *events.HistorySync, sessionID string) {
	h.logger.InfoWithFields("History sync", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Data.SyncType,
		"progress":   evt.Data.Progress,
	})
}

func (h *EventHandler) handleAppState(_ *events.AppState) {
	h.logger.DebugWithFields("App state update", map[string]interface{}{
		"name": "app_state",
	})
}

func (h *EventHandler) handleAppStateSyncComplete(evt *events.AppStateSyncComplete, sessionID string) {
	h.logger.InfoWithFields("App state sync complete", map[string]interface{}{
		"session_id": sessionID,
		"name":       evt.Name,
	})
}

func (h *EventHandler) handleKeepAliveTimeout(_ *events.KeepAliveTimeout, sessionID string) {
	h.logger.WarnWithFields("Keep alive timeout", map[string]interface{}{
		"session_id": sessionID,
	})
}

func (h *EventHandler) handleKeepAliveRestored(_ *events.KeepAliveRestored, sessionID string) {
	h.logger.InfoWithFields("Keep alive restored", map[string]interface{}{
		"session_id": sessionID,
	})
}

func (h *EventHandler) handleContact(evt *events.Contact, sessionID string) {
	h.logger.DebugWithFields("Contact update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleGroupInfo(evt *events.GroupInfo, sessionID string) {
	h.logger.DebugWithFields("Group info update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handlePicture(evt *events.Picture, sessionID string) {
	h.logger.DebugWithFields("Picture update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleBusinessName(evt *events.BusinessName, sessionID string) {
	h.logger.DebugWithFields("Business name update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) saveMessageToDatabase(evt *events.Message, sessionID string) error {

	message, err := h.convertWhatsmeowMessage(evt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	if err := h.gateway.SaveReceivedMessage(message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	h.logger.DebugWithFields("Message saved to database", map[string]interface{}{
		"session_id":    sessionID,
		"message_id":    evt.Info.ID,
		"zp_message_id": message.ZpMessageID,
	})

	return nil
}

func (h *EventHandler) convertWhatsmeowMessage(evt *events.Message, sessionID string) (*messaging.Message, error) {

	contentMap := h.extractMessageContent(evt.Message)

	contentStr := fmt.Sprintf("%v", contentMap)

	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	message := &messaging.Message{
		ID:          uuid.New(),
		SessionID:   sessionUUID,
		ZpMessageID: evt.Info.ID,
		ZpSender:    evt.Info.Sender.String(),
		ZpChat:      evt.Info.Chat.String(),
		ZpTimestamp: evt.Info.Timestamp,
		ZpFromMe:    evt.Info.IsFromMe,
		ZpType:      string(evt.Info.Type),
		Content:     contentStr,
		SyncStatus:  "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return message, nil
}

func (h *EventHandler) extractMessageContent(message *waE2E.Message) map[string]interface{} {
	content := make(map[string]interface{})

	switch {
	case message.GetConversation() != "":
		content["text"] = message.GetConversation()
		content["type"] = "text"
	case message.GetExtendedTextMessage() != nil:
		content["text"] = message.GetExtendedTextMessage().GetText()
		content["type"] = "extended_text"
	case message.GetImageMessage() != nil:
		img := message.GetImageMessage()
		content["type"] = "image"
		content["caption"] = img.GetCaption()
		content["mimetype"] = img.GetMimetype()
		content["url"] = img.GetURL()
	case message.GetVideoMessage() != nil:
		vid := message.GetVideoMessage()
		content["type"] = "video"
		content["caption"] = vid.GetCaption()
		content["mimetype"] = vid.GetMimetype()
		content["url"] = vid.GetURL()
	case message.GetAudioMessage() != nil:
		aud := message.GetAudioMessage()
		content["type"] = "audio"
		content["mimetype"] = aud.GetMimetype()
		content["url"] = aud.GetURL()
	case message.GetDocumentMessage() != nil:
		doc := message.GetDocumentMessage()
		content["type"] = "document"
		content["filename"] = doc.GetFileName()
		content["mimetype"] = doc.GetMimetype()
		content["url"] = doc.GetURL()
	default:
		content["type"] = "unknown"
		content["raw"] = fmt.Sprintf("%+v", message)
	}

	return content
}


func (h *EventHandler) notifySessionConnected(sessionID string) {
	handlers := h.gateway.getEventHandlers("global")
	for _, handler := range handlers {
		go func(sessionHandler session.EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					h.logger.ErrorWithFields("Session event handler panic", map[string]interface{}{
						"session_id": sessionID,
						"event":      "connected",
						"error":      r,
					})
				}
			}()
			sessionHandler.OnSessionConnected(h.sessionName, nil)
		}(handler)
	}
}

func (h *EventHandler) notifySessionDisconnected(sessionID, reason string) {
	handlers := h.gateway.getEventHandlers("global")
	for _, handler := range handlers {
		go func(sessionHandler session.EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					h.logger.ErrorWithFields("Session event handler panic", map[string]interface{}{
						"session_id": sessionID,
						"event":      "disconnected",
						"error":      r,
					})
				}
			}()
			sessionHandler.OnSessionDisconnected(h.sessionName, reason)
		}(handler)
	}
}
