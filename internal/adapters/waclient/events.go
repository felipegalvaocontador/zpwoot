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
	"zpwoot/platform/logger"
)

// EventHandler processa eventos do WhatsApp baseado no legacy
type EventHandler struct {
	gateway     *Gateway
	sessionName string
	logger      *logger.Logger

	// QR code generator (reutilizado para evitar criação desnecessária)
	qrGenerator *QRGenerator

	// Callbacks externos
	webhookHandler  WebhookEventHandler
	chatwootManager ChatwootManager
}

// WebhookEventHandler interface para processar eventos via webhook
type WebhookEventHandler interface {
	HandleWhatsmeowEvent(evt interface{}, sessionID string) error
}

// ChatwootManager interface para integração com Chatwoot
type ChatwootManager interface {
	IsEnabled(sessionID string) bool
	ProcessWhatsAppMessage(sessionID, messageID, from, content, messageType string, timestamp time.Time, fromMe bool) error
}

// NewEventHandler cria novo event handler
func NewEventHandler(gateway *Gateway, sessionName string, logger *logger.Logger) *EventHandler {
	return &EventHandler{
		gateway:     gateway,
		sessionName: sessionName,
		logger:      logger,
		qrGenerator: NewQRGenerator(logger), // Inicializar QR generator uma vez
	}
}

// SetWebhookHandler configura webhook handler
func (h *EventHandler) SetWebhookHandler(handler WebhookEventHandler) {
	h.webhookHandler = handler
}

// SetChatwootManager configura Chatwoot manager
func (h *EventHandler) SetChatwootManager(manager ChatwootManager) {
	h.chatwootManager = manager
}

// HandleEvent processa eventos do WhatsApp
func (h *EventHandler) HandleEvent(evt interface{}, sessionID string) {
	// Entregar para webhook primeiro
	h.deliverToWebhook(evt, sessionID)

	// Processar evento internamente
	h.handleEventInternal(evt, sessionID)
}

// handleEventInternal processa eventos internamente
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

// deliverToWebhook entrega evento para webhook se configurado
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

// ===== EVENT HANDLERS =====

// handleConnected processa evento de conexão
func (h *EventHandler) handleConnected(evt *events.Connected, sessionID string) {
	h.logger.InfoWithFields("WhatsApp connected", map[string]interface{}{
		"session_id": sessionID,
	})

	// Atualizar status da sessão no banco de dados
	if err := h.gateway.UpdateSessionStatus(sessionID, "connected"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "connected",
			"error":      err.Error(),
		})
	}
}

// handleDisconnected processa evento de desconexão
func (h *EventHandler) handleDisconnected(evt *events.Disconnected, sessionID string) {
	h.logger.WarnWithFields("WhatsApp disconnected", map[string]interface{}{
		"session_id": sessionID,
	})

	// Atualizar status da sessão no banco de dados
	if err := h.gateway.UpdateSessionStatus(sessionID, "disconnected"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "disconnected",
			"error":      err.Error(),
		})
	}
}

// handleLoggedOut processa evento de logout
func (h *EventHandler) handleLoggedOut(evt *events.LoggedOut, sessionID string) {
	h.logger.WarnWithFields("WhatsApp logged out", map[string]interface{}{
		"session_id": sessionID,
		"reason":     evt.Reason,
	})

	// Atualizar status da sessão no banco de dados
	if err := h.gateway.UpdateSessionStatus(sessionID, "logged_out"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "logged_out",
			"error":      err.Error(),
		})
	}
}

// handleQREvent processa evento de QR code (sem dados do QR)
func (h *EventHandler) handleQREvent(sessionID string) {
	h.logger.InfoWithFields("QR code event received", map[string]interface{}{
		"session_id": sessionID,
	})

	// Atualizar status da sessão no banco de dados
	if err := h.gateway.UpdateSessionStatus(sessionID, "qr_code"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "qr_code",
			"error":      err.Error(),
		})
	}
}

// handleQRCodeEvent processa evento de QR code customizado (com dados do QR)
func (h *EventHandler) handleQRCodeEvent(evt *QRCodeEvent, sessionID string) {
	h.logger.InfoWithFields("QR code event with data received", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": evt.SessionName,
		"qr_length":    len(evt.QRCode),
		"expires_at":   evt.ExpiresAt,
	})

	// Exibir QR code no terminal usando o QR generator reutilizado
	h.qrGenerator.DisplayQRCodeInTerminal(evt.QRCode, evt.SessionName)

	// Atualizar status da sessão no banco de dados
	if err := h.gateway.UpdateSessionStatus(sessionID, "qr_code"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status", map[string]interface{}{
			"session_id": sessionID,
			"status":     "qr_code",
			"error":      err.Error(),
		})
	}

	// Atualizar QR code no banco de dados
	if err := h.gateway.UpdateSessionQRCode(sessionID, evt.QRCode, evt.ExpiresAt); err != nil {
		h.logger.ErrorWithFields("Failed to update QR code in database", map[string]interface{}{
			"session_id": sessionID,
			"qr_length":  len(evt.QRCode),
			"error":      err.Error(),
		})
	}
}

// handlePairSuccess processa evento de pareamento bem-sucedido
func (h *EventHandler) handlePairSuccess(evt *events.PairSuccess, sessionID string) {
	deviceJID := evt.ID.String()

	h.logger.InfoWithFields("WhatsApp pairing successful", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": deviceJID,
	})

	// Atualizar device JID da sessão no banco de dados
	if err := h.gateway.UpdateSessionDeviceJID(sessionID, deviceJID); err != nil {
		h.logger.ErrorWithFields("Failed to update session device JID", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": deviceJID,
			"error":      err.Error(),
		})
	}

	// Atualizar status da sessão para conectada
	if err := h.gateway.UpdateSessionStatus(sessionID, "connected"); err != nil {
		h.logger.ErrorWithFields("Failed to update session status after pairing", map[string]interface{}{
			"session_id": sessionID,
			"status":     "connected",
			"error":      err.Error(),
		})
	}
}

// handlePairError processa evento de erro de pareamento
func (h *EventHandler) handlePairError(evt *events.PairError, sessionID string) {
	h.logger.ErrorWithFields("WhatsApp pairing failed", map[string]interface{}{
		"session_id": sessionID,
		"error":      evt.Error.Error(),
	})

	// TODO: Atualizar status da sessão no banco de dados
}

// handleMessage processa evento de mensagem baseado no legacy
func (h *EventHandler) handleMessage(evt *events.Message, sessionID string) {
	h.logger.InfoWithFields("Message received", map[string]interface{}{
		"session_id": sessionID,
		"message_id": evt.Info.ID,
		"from":       evt.Info.Sender.String(),
		"chat":       evt.Info.Chat.String(),
		"from_me":    evt.Info.IsFromMe,
		"type":       evt.Info.Type,
	})

	// Salvar mensagem no banco de dados
	if err := h.saveMessageToDatabase(evt, sessionID); err != nil {
		h.logger.ErrorWithFields("Failed to save message to database", map[string]interface{}{
			"session_id": sessionID,
			"message_id": evt.Info.ID,
			"error":      err.Error(),
		})
	}

	// Processar para Chatwoot se habilitado
	if h.chatwootManager != nil && h.chatwootManager.IsEnabled(sessionID) {
		h.processMessageForChatwoot(evt, sessionID)
	}
}

// handleReceipt processa evento de recibo
func (h *EventHandler) handleReceipt(evt *events.Receipt, sessionID string) {
	h.logger.DebugWithFields("Receipt received", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Type,
		"sender":     evt.Sender.String(),
		"timestamp":  evt.Timestamp,
	})

	// TODO: Atualizar status das mensagens
}

// handleOtherEvents processa outros eventos
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

// processMessageForChatwoot processa mensagem para Chatwoot
func (h *EventHandler) processMessageForChatwoot(evt *events.Message, sessionID string) {
	messageID := evt.Info.ID
	from := evt.Info.Sender.String()
	timestamp := evt.Info.Timestamp
	fromMe := evt.Info.IsFromMe

	// Extrair conteúdo e tipo da mensagem
	content, messageType := h.extractMessageContentString(evt.Message)

	// Extrair número de telefone do contato
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

// extractMessageContentString extrai conteúdo e tipo da mensagem como string
func (h *EventHandler) extractMessageContentString(message *waE2E.Message) (string, string) {
	if message == nil {
		return "", "unknown"
	}

	// Texto simples
	if message.Conversation != nil {
		return *message.Conversation, "text"
	}

	// Texto estendido
	if message.ExtendedTextMessage != nil && message.ExtendedTextMessage.Text != nil {
		return *message.ExtendedTextMessage.Text, "text"
	}

	// Imagem
	if message.ImageMessage != nil {
		caption := ""
		if message.ImageMessage.Caption != nil {
			caption = *message.ImageMessage.Caption
		}
		return caption, "image"
	}

	// Áudio
	if message.AudioMessage != nil {
		return "[Audio]", "audio"
	}

	// Vídeo
	if message.VideoMessage != nil {
		caption := ""
		if message.VideoMessage.Caption != nil {
			caption = *message.VideoMessage.Caption
		}
		return caption, "video"
	}

	// Documento
	if message.DocumentMessage != nil {
		filename := ""
		if message.DocumentMessage.FileName != nil {
			filename = *message.DocumentMessage.FileName
		}
		return fmt.Sprintf("[Document: %s]", filename), "document"
	}

	// Sticker
	if message.StickerMessage != nil {
		return "[Sticker]", "sticker"
	}

	// Localização
	if message.LocationMessage != nil {
		return "[Location]", "location"
	}

	// Contato
	if message.ContactMessage != nil {
		name := ""
		if message.ContactMessage.DisplayName != nil {
			name = *message.ContactMessage.DisplayName
		}
		return fmt.Sprintf("[Contact: %s]", name), "contact"
	}

	return "[Unknown message type]", "unknown"
}

// extractContactNumber extrai número de telefone do JID
func (h *EventHandler) extractContactNumber(jid string) string {
	// JID format: number@s.whatsapp.net
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}

// ===== OTHER EVENT HANDLERS =====

// handlePresence processa evento de presença
func (h *EventHandler) handlePresence(evt *events.Presence, sessionID string) {
	h.logger.DebugWithFields("Presence update", map[string]interface{}{
		"session_id": sessionID,
		"from":       evt.From.String(),
		"presence":   evt.Unavailable,
	})
}

// handleChatPresence processa evento de presença em chat
func (h *EventHandler) handleChatPresence(evt *events.ChatPresence, sessionID string) {
	h.logger.DebugWithFields("Chat presence update", map[string]interface{}{
		"session_id": sessionID,
		"chat":       evt.Chat.String(),
		"state":      evt.State,
	})
}

// handleHistorySync processa evento de sincronização de histórico
func (h *EventHandler) handleHistorySync(evt *events.HistorySync, sessionID string) {
	h.logger.InfoWithFields("History sync", map[string]interface{}{
		"session_id": sessionID,
		"type":       evt.Data.SyncType,
		"progress":   evt.Data.Progress,
	})
}

// handleAppState processa evento de estado da aplicação
func (h *EventHandler) handleAppState(evt *events.AppState) {
	h.logger.DebugWithFields("App state update", map[string]interface{}{
		"name": "app_state",
	})
}

// handleAppStateSyncComplete processa evento de sincronização completa
func (h *EventHandler) handleAppStateSyncComplete(evt *events.AppStateSyncComplete, sessionID string) {
	h.logger.InfoWithFields("App state sync complete", map[string]interface{}{
		"session_id": sessionID,
		"name":       evt.Name,
	})
}

// handleKeepAliveTimeout processa evento de timeout de keep alive
func (h *EventHandler) handleKeepAliveTimeout(evt *events.KeepAliveTimeout, sessionID string) {
	h.logger.WarnWithFields("Keep alive timeout", map[string]interface{}{
		"session_id": sessionID,
	})
}

// handleKeepAliveRestored processa evento de keep alive restaurado
func (h *EventHandler) handleKeepAliveRestored(evt *events.KeepAliveRestored, sessionID string) {
	h.logger.InfoWithFields("Keep alive restored", map[string]interface{}{
		"session_id": sessionID,
	})
}

// handleContact processa evento de contato
func (h *EventHandler) handleContact(evt *events.Contact, sessionID string) {
	h.logger.DebugWithFields("Contact update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handleGroupInfo processa evento de informações de grupo
func (h *EventHandler) handleGroupInfo(evt *events.GroupInfo, sessionID string) {
	h.logger.DebugWithFields("Group info update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handlePicture processa evento de foto
func (h *EventHandler) handlePicture(evt *events.Picture, sessionID string) {
	h.logger.DebugWithFields("Picture update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// handleBusinessName processa evento de nome comercial
func (h *EventHandler) handleBusinessName(evt *events.BusinessName, sessionID string) {
	h.logger.DebugWithFields("Business name update", map[string]interface{}{
		"session_id": sessionID,
		"jid":        evt.JID.String(),
	})
}

// ===== MESSAGE PROCESSING =====

// saveMessageToDatabase salva mensagem recebida no banco de dados baseado no legacy
func (h *EventHandler) saveMessageToDatabase(evt *events.Message, sessionID string) error {
	// Converter mensagem whatsmeow para formato interno
	message, err := h.convertWhatsmeowMessage(evt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Salvar via gateway (que tem acesso aos repositórios)
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

// convertWhatsmeowMessage converte mensagem whatsmeow para formato interno
func (h *EventHandler) convertWhatsmeowMessage(evt *events.Message, sessionID string) (*messaging.Message, error) {
	// Extrair conteúdo da mensagem baseado no tipo
	contentMap := h.extractMessageContent(evt.Message)

	// Converter content map para string JSON
	contentStr := fmt.Sprintf("%v", contentMap)

	// Parse sessionID para UUID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Criar mensagem no formato interno
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

// extractMessageContent extrai conteúdo da mensagem baseado no tipo
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
