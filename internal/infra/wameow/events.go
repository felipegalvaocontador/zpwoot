package wameow

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow/types/events"
)

// Message type constants
const (
	MessageTypeText     = "text"
	MessageTypeImage    = "image"
	MessageTypeAudio    = "audio"
	MessageTypeVideo    = "video"
	MessageTypeDocument = "document"
	MessageTypeSticker  = "sticker"
	MessageTypeLocation = "location"
	MessageTypeContact  = "contact"
)

// WebhookEventHandler defines interface for handling webhook events
type WebhookEventHandler interface {
	HandleWhatsmeowEvent(evt interface{}, sessionID string) error
}

type EventHandler struct {
	manager         *Manager
	sessionMgr      SessionUpdater
	qrGen           *QRCodeGenerator
	logger          *logger.Logger
	webhookHandler  WebhookEventHandler
	chatwootManager ChatwootManager // Interface for Chatwoot integration
}

// ChatwootManager interface for Chatwoot integration
type ChatwootManager interface {
	IsEnabled(sessionID string) bool
	ProcessWhatsAppMessage(sessionID, messageID, from, content, messageType string, timestamp time.Time, fromMe bool) error
}

func NewEventHandler(manager *Manager, sessionMgr SessionUpdater, qrGen *QRCodeGenerator, logger *logger.Logger) *EventHandler {
	return &EventHandler{
		manager:    manager,
		sessionMgr: sessionMgr,
		qrGen:      qrGen,
		logger:     logger,
	}
}

// SetChatwootManager sets the Chatwoot manager for integration
func (h *EventHandler) SetChatwootManager(chatwootManager ChatwootManager) {
	h.chatwootManager = chatwootManager
}

func (h *EventHandler) HandleEvent(evt interface{}, sessionID string) {
	// First, deliver to webhook if configured
	h.deliverToWebhook(evt, sessionID)

	// Then handle the event internally
	switch v := evt.(type) {
	case *events.Connected:
		h.handleConnected(v, sessionID)
	case *events.Disconnected:
		h.handleDisconnected(v, sessionID)
	case *events.LoggedOut:
		h.handleLoggedOut(v, sessionID)
	case *events.QR:
		// QR events are handled by client QR channel to avoid duplication
		// All QR code processing is done through client.handleQREvent()
		h.logger.DebugWithFields("QR event received but skipped (handled by client channel)", map[string]interface{}{
			"session_id": sessionID,
		})
	case *events.PairSuccess:
		h.handlePairSuccess(v, sessionID)
	case *events.PairError:
		h.handlePairError(v, sessionID)
	case *events.Message:
		h.handleMessage(v, sessionID)
	case *events.Receipt:
		h.handleReceipt(v, sessionID)
	case *events.Presence:
		h.handlePresence(v, sessionID)
	case *events.ChatPresence:
		h.handleChatPresence(v, sessionID)
	case *events.HistorySync:
		h.handleHistorySync(v, sessionID)
	case *events.AppState:
		h.handleAppState(v, sessionID)
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
	case *events.PushName:
		h.handlePushName(v, sessionID)
	case *events.Archive:
		h.handleArchive(v, sessionID)
	case *events.Pin:
		h.handlePin(v, sessionID)
	case *events.Mute:
		h.handleMute(v, sessionID)
	case *events.Star:
		h.handleStar(v, sessionID)
	case *events.DeleteForMe:
		h.handleDeleteForMe(v, sessionID)
	case *events.MarkChatAsRead:
		h.handleMarkChatAsRead(v, sessionID)
	case *events.UndecryptableMessage:
		h.handleUndecryptableMessage(v, sessionID)
	case *events.OfflineSyncPreview:
		h.handleOfflineSyncPreview(v, sessionID)
	case *events.OfflineSyncCompleted:
		h.handleOfflineSyncCompleted(v, sessionID)
	default:
		h.logger.DebugWithFields("Unhandled event", map[string]interface{}{
			"session_id": sessionID,
			"event_type": getEventType(evt),
		})
	}
}

func (h *EventHandler) handleConnected(evt *events.Connected, sessionID string) {
	h.logger.InfoWithFields("Wameow connected", map[string]interface{}{
		"session_id":   sessionID,
		"event_type":   "Connected",
		"connected_at": time.Now().Unix(),
	})

	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)
}

func (h *EventHandler) handleDisconnected(evt *events.Disconnected, sessionID string) {
	h.logger.InfoWithFields("Wameow disconnected", map[string]interface{}{
		"session_id":      sessionID,
		"event_type":      "Disconnected",
		"disconnected_at": time.Now().Unix(),
	})

	_ = evt

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

func (h *EventHandler) handleLoggedOut(evt *events.LoggedOut, sessionID string) {
	h.logger.InfoWithFields("Wameow logged out", map[string]interface{}{
		"session_id": sessionID,
		"reason":     evt.Reason,
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

func (h *EventHandler) handleQR(evt *events.QR, sessionID string) {
	if len(evt.Codes) == 0 {
		h.logger.WarnWithFields("QR event received with no codes", map[string]interface{}{
			"session_id": sessionID,
		})
		return
	}

	h.logger.InfoWithFields("QR code received", map[string]interface{}{
		"session_id":  sessionID,
		"codes_count": len(evt.Codes),
	})

	// Salva o QR code no banco e exibe no terminal
	qrCode := evt.Codes[0]
	if qrCode != "" {
		h.updateSessionQRCode(sessionID, qrCode) // Save raw QR code to database

		// Exibe o QR code no terminal
		h.qrGen.DisplayQRCodeInTerminal(qrCode, sessionID)
	}
}

func (h *EventHandler) handlePairSuccess(evt *events.PairSuccess, sessionID string) {
	h.logger.InfoWithFields("Pairing successful", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": evt.ID.String(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, true)

	h.updateSessionDeviceJID(sessionID, evt.ID.String())

	h.clearSessionQRCode(sessionID)
}

func (h *EventHandler) handlePairError(evt *events.PairError, sessionID string) {
	h.logger.ErrorWithFields("Pairing failed", map[string]interface{}{
		"session_id": sessionID,
		"error":      evt.Error.Error(),
	})

	h.sessionMgr.UpdateConnectionStatus(sessionID, false)
}

func (h *EventHandler) handleMessage(evt *events.Message, sessionID string) {
	messageInfo := map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.Info.Sender.String(),
		"message_id": evt.Info.ID,
		"timestamp":  evt.Info.Timestamp,
	}

	// Removed detailed message analysis for cleaner logs

	if evt.Message.ContactMessage != nil {
		contactMsg := evt.Message.ContactMessage
		messageInfo["message_type"] = MessageTypeContact

		if contactMsg.DisplayName != nil {
			messageInfo["contact_display_name"] = *contactMsg.DisplayName
		}

		if contactMsg.Vcard != nil {
			messageInfo["contact_vcard"] = *contactMsg.Vcard
			messageInfo["vcard_length"] = len(*contactMsg.Vcard)
		}

		h.logger.InfoWithFields("📞 CONTACT MESSAGE RECEIVED", messageInfo)

		if contactMsg.Vcard != nil {
			h.logger.InfoWithFields("📋 FULL VCARD CONTENT", map[string]interface{}{
				"session_id": sessionID,
				"from":       evt.Info.Sender.String(),
				"vcard":      *contactMsg.Vcard,
			})
		}
	} else if evt.Message.ContactsArrayMessage != nil {
		contactsMsg := evt.Message.ContactsArrayMessage
		messageInfo["message_type"] = "contacts_array"

		if contactsMsg.DisplayName != nil {
			messageInfo["contacts_display_name"] = *contactsMsg.DisplayName
		}

		if contactsMsg.Contacts != nil {
			messageInfo["contacts_count"] = len(contactsMsg.Contacts)
		}

		h.logger.InfoWithFields("📞📞📞 CONTACTS ARRAY MESSAGE RECEIVED", messageInfo)

		if contactsMsg.Contacts != nil {
			for i, contact := range contactsMsg.Contacts {
				contactInfo := map[string]interface{}{
					"session_id":    sessionID,
					"from":          evt.Info.Sender.String(),
					"contact_index": i,
				}

				if contact.DisplayName != nil {
					contactInfo["contact_display_name"] = *contact.DisplayName
				}

				if contact.Vcard != nil {
					contactInfo["contact_vcard"] = *contact.Vcard
					contactInfo["vcard_length"] = len(*contact.Vcard)
				}

				h.logger.InfoWithFields(fmt.Sprintf("📋 CONTACT %d VCARD CONTENT", i+1), contactInfo)
			}
		}
	} else {
		messageType := MessageTypeText
		if evt.Message.ImageMessage != nil {
			messageType = MessageTypeImage
		} else if evt.Message.AudioMessage != nil {
			messageType = MessageTypeAudio
		} else if evt.Message.VideoMessage != nil {
			messageType = MessageTypeVideo
		} else if evt.Message.DocumentMessage != nil {
			messageType = MessageTypeDocument
		} else if evt.Message.StickerMessage != nil {
			messageType = MessageTypeSticker
		} else if evt.Message.LocationMessage != nil {
			messageType = MessageTypeLocation
		}

		messageInfo["message_type"] = messageType

		if evt.Message.GetConversation() != "" {
			messageInfo["text_content"] = evt.Message.GetConversation()
		}

		h.logger.InfoWithFields("Message received", messageInfo)
	}

	h.updateSessionLastSeen(sessionID)

	// Process message for Chatwoot integration if enabled
	h.processChatwootIntegration(evt, sessionID)
}

// processChatwootIntegration processes the message for Chatwoot integration
func (h *EventHandler) processChatwootIntegration(evt *events.Message, sessionID string) {
	// Check if Chatwoot manager is available and enabled
	if h.chatwootManager == nil || !h.chatwootManager.IsEnabled(sessionID) {
		return
	}

	// Extract message information
	messageID := evt.Info.ID
	from := evt.Info.Sender.String()
	chat := evt.Info.Chat.String() // This is the actual recipient/chat
	timestamp := evt.Info.Timestamp
	fromMe := evt.Info.IsFromMe

	// For from_me=true messages, we need to use the Chat (recipient) instead of Sender
	// The Sender is always our session number, but Chat is who we sent the message to
	contactNumber := from
	if fromMe {
		contactNumber = chat // Use the recipient (who we sent the message to)
	}

	// For messages sent by us (from_me=true), we need to distinguish:
	// 1. Messages sent from Chatwoot → WhatsApp (should be skipped to avoid loop)
	// 2. Messages sent from phone/other devices → WhatsApp (should be processed for history sync)
	//
	// We can identify Chatwoot-originated messages by checking if they're already mapped
	// in our message mapping system (they would have been mapped when sent from Chatwoot)
	if fromMe {
		// Check if this message was already processed/mapped (came from Chatwoot)
		// We'll let the Chatwoot integration manager handle this check
		h.logger.DebugWithFields("Processing from_me message (could be from phone or Chatwoot)", map[string]interface{}{
			"session_id":     sessionID,
			"message_id":     messageID,
			"from":           from,
			"chat":           chat,
			"contact_number": contactNumber,
			"from_me":        fromMe,
		})
	}

	// Determine message type and content
	messageType := MessageTypeText
	content := ""

	if evt.Message.ContactMessage != nil {
		messageType = MessageTypeContact
		if evt.Message.ContactMessage.DisplayName != nil {
			content = "Contact: " + *evt.Message.ContactMessage.DisplayName
		} else {
			content = "Contact shared"
		}
	} else if evt.Message.ContactsArrayMessage != nil {
		messageType = "contacts"
		content = fmt.Sprintf("Contacts shared (%d contacts)", len(evt.Message.ContactsArrayMessage.Contacts))
	} else if evt.Message.ImageMessage != nil {
		messageType = "image"
		if evt.Message.ImageMessage.Caption != nil {
			content = *evt.Message.ImageMessage.Caption
		} else {
			content = "Image"
		}
	} else if evt.Message.AudioMessage != nil {
		messageType = "audio"
		content = "Audio message"
	} else if evt.Message.VideoMessage != nil {
		messageType = "video"
		if evt.Message.VideoMessage.Caption != nil {
			content = *evt.Message.VideoMessage.Caption
		} else {
			content = "Video"
		}
	} else if evt.Message.DocumentMessage != nil {
		messageType = "document"
		if evt.Message.DocumentMessage.Title != nil {
			content = "Document: " + *evt.Message.DocumentMessage.Title
		} else {
			content = "Document"
		}
	} else if evt.Message.StickerMessage != nil {
		messageType = "sticker"
		content = "Sticker"
	} else if evt.Message.LocationMessage != nil {
		messageType = "location"
		content = "Location shared"
	} else if evt.Message.GetConversation() != "" {
		messageType = "text"
		content = evt.Message.GetConversation()
	}

	// Process the message with Chatwoot
	// Use contactNumber which is the correct contact (sender for incoming, recipient for outgoing)
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

func (h *EventHandler) handleReceipt(evt *events.Receipt, sessionID string) {
	h.logger.InfoWithFields("Receipt received", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Type,
		"sender":     evt.Sender.String(),
		"timestamp":  evt.Timestamp,
	})
}

func (h *EventHandler) handlePresence(evt *events.Presence, sessionID string) {
	h.logger.InfoWithFields("Presence update", map[string]interface{}{
		"session_id":  sessionID,
		"from":        evt.From.String(),
		"unavailable": evt.Unavailable,
		"last_seen":   evt.LastSeen,
	})
}

func (h *EventHandler) handleChatPresence(evt *events.ChatPresence, sessionID string) {
	h.logger.InfoWithFields("Chat presence update", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.Chat.String(),
		"state":      evt.State,
	})
}

func (h *EventHandler) handleHistorySync(evt *events.HistorySync, sessionID string) {
	h.logger.InfoWithFields("History sync", map[string]interface{}{
		"session_id": sessionID,
		"data_size":  len(evt.Data.String()), // Just log the data size for now
	})
}

func (h *EventHandler) handleAppState(evt *events.AppState, sessionID string) {
	// Skip AppState events to reduce spam - they're not critical for most use cases
	// Only log at trace level to avoid log pollution
	_ = evt // Avoid unused parameter warning
}

func (h *EventHandler) handleAppStateSyncComplete(evt *events.AppStateSyncComplete, sessionID string) {
	h.logger.DebugWithFields("App state sync complete", map[string]interface{}{
		"session_id": sessionID,
		"name":       evt.Name,
	})
}

func (h *EventHandler) handleKeepAliveTimeout(evt *events.KeepAliveTimeout, sessionID string) {
	h.logger.DebugWithFields("Keep alive timeout", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

func (h *EventHandler) handleKeepAliveRestored(evt *events.KeepAliveRestored, sessionID string) {
	h.logger.DebugWithFields("Keep alive restored", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
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

func (h *EventHandler) handlePushName(evt *events.PushName, sessionID string) {
	h.logger.DebugWithFields("Push name update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleArchive(evt *events.Archive, sessionID string) {
	h.logger.DebugWithFields("Archive update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handlePin(evt *events.Pin, sessionID string) {
	h.logger.DebugWithFields("Pin update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleMute(evt *events.Mute, sessionID string) {
	h.logger.DebugWithFields("Mute update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

func (h *EventHandler) handleStar(evt *events.Star, sessionID string) {
	h.logger.DebugWithFields("Star update", map[string]interface{}{
		"session_id": sessionID,
	})
	_ = evt // Avoid unused parameter warning
}

func (h *EventHandler) handleDeleteForMe(evt *events.DeleteForMe, sessionID string) {
	h.logger.DebugWithFields("Delete for me", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.ChatJID.String(),
	})
}

func (h *EventHandler) handleMarkChatAsRead(evt *events.MarkChatAsRead, sessionID string) {
	h.logger.DebugWithFields("Mark chat as read", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.JID.String(),
	})
}

func (h *EventHandler) handleUndecryptableMessage(evt *events.UndecryptableMessage, sessionID string) {
	h.logger.DebugWithFields("Undecryptable message", map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.Info.Sender.String(),
	})
}

func (h *EventHandler) handleOfflineSyncPreview(evt *events.OfflineSyncPreview, sessionID string) {
	h.logger.DebugWithFields("Offline sync preview", map[string]interface{}{
		"session_id": sessionID,
		"messages":   evt.Messages,
	})
}

func (h *EventHandler) handleOfflineSyncCompleted(evt *events.OfflineSyncCompleted, sessionID string) {
	h.logger.DebugWithFields("Offline sync completed", map[string]interface{}{
		"session_id": sessionID,
		"count":      evt.Count,
	})
}

func (h *EventHandler) updateSessionQRCode(sessionID, qrCode string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.GetSessionRepo().GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for QR update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	sess.QRCode = qrCode
	expiresAt := time.Now().Add(2 * time.Minute) // QR code expires in 2 minutes
	sess.QRCodeExpiresAt = &expiresAt
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.GetSessionRepo().Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) updateSessionDeviceJID(sessionID, deviceJID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.GetSessionRepo().GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for device JID update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	sess.DeviceJid = deviceJID
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.GetSessionRepo().Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session device JID", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) updateSessionLastSeen(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.GetSessionRepo().GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for last seen update", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	now := time.Now()
	sess.LastSeen = &now
	sess.UpdatedAt = now

	if err := h.sessionMgr.GetSessionRepo().Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to update session last seen", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

func (h *EventHandler) clearSessionQRCode(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := h.sessionMgr.GetSessionRepo().GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get session for QR code clearing", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	sess.QRCode = ""
	sess.QRCodeExpiresAt = nil
	sess.UpdatedAt = time.Now()

	if err := h.sessionMgr.GetSessionRepo().Update(ctx, sess); err != nil {
		h.logger.ErrorWithFields("Failed to clear session QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	} else {
		h.logger.InfoWithFields("QR code cleared after successful connection", map[string]interface{}{
			"session_id": sessionID,
		})
	}
}

// getEventType extracts the event type name using reflection
func getEventType(evt interface{}) string {
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
		// Fallback to full type string
		typeName = strings.TrimPrefix(eventType.String(), "*events.")
	}

	return typeName
}

// SetWebhookHandler sets the webhook handler in the EventHandler
func (h *EventHandler) SetWebhookHandler(webhookHandler WebhookEventHandler) {
	h.webhookHandler = webhookHandler
}

// deliverToWebhook delivers an event to the webhook handler if configured
func (h *EventHandler) deliverToWebhook(evt interface{}, sessionID string) {
	if h.webhookHandler == nil {
		return
	}

	if err := h.webhookHandler.HandleWhatsmeowEvent(evt, sessionID); err != nil {
		h.logger.ErrorWithFields("Failed to deliver event to webhook", map[string]interface{}{
			"session_id": sessionID,
			"event_type": getEventType(evt),
			"error":      err.Error(),
		})
	}
}

// HandleQRCode processes QR codes from client channel (not automatic events)
// This is the single source of truth for all QR code processing
func (h *EventHandler) HandleQRCode(sessionID string, qrCode string) {
	h.logger.InfoWithFields("QR code received from client channel", map[string]interface{}{
		"session_id": sessionID,
	})

	// Process QR code: save to database and display in terminal
	if qrCode != "" {
		h.updateSessionQRCode(sessionID, qrCode)
		h.qrGen.DisplayQRCodeInTerminal(qrCode, sessionID)
	}
}
