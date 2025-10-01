package waclient

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"zpwoot/internal/core/session"
)

// MessageMapper mapeia mensagens entre diferentes formatos baseado no legacy
type MessageMapper struct{}

// NewMessageMapper cria novo mapper de mensagens
func NewMessageMapper() *MessageMapper {
	return &MessageMapper{}
}

// EventToWhatsAppMessage converte evento de mensagem para WhatsAppMessage
func (m *MessageMapper) EventToWhatsAppMessage(evt *events.Message) *session.WhatsAppMessage {
	if evt == nil {
		return nil
	}

	// Extrair conteúdo e tipo da mensagem
	content, messageType := m.extractMessageContent(evt.Message)

	return &session.WhatsAppMessage{
		ID:        evt.Info.ID,
		From:      evt.Info.Sender.String(),
		To:        evt.Info.Chat.String(),
		Chat:      evt.Info.Chat.String(),
		Type:      messageType,
		Content:   content,
		Timestamp: evt.Info.Timestamp,
		FromMe:    evt.Info.IsFromMe,
		Metadata: map[string]interface{}{
			"push_name":      evt.Info.PushName,
			"message_type":   evt.Info.Type,
			"category":       evt.Info.Category,
		},
	}
}

// extractMessageContent extrai conteúdo e tipo da mensagem
func (m *MessageMapper) extractMessageContent(message *waE2E.Message) (string, string) {
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

// JIDToPhoneNumber converte JID para número de telefone
func (m *MessageMapper) JIDToPhoneNumber(jid string) string {
	// JID format: number@s.whatsapp.net
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}

// PhoneNumberToJID converte número de telefone para JID
func (m *MessageMapper) PhoneNumberToJID(phoneNumber string) types.JID {
	// Remove caracteres não numéricos
	cleanNumber := strings.ReplaceAll(phoneNumber, "+", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")

	return types.JID{
		User:   cleanNumber,
		Server: types.DefaultUserServer,
	}
}

// FormatJID formata JID para exibição
func (m *MessageMapper) FormatJID(jid types.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	return jid.String()
}

// IsGroupJID verifica se JID é de grupo
func (m *MessageMapper) IsGroupJID(jid string) bool {
	return strings.Contains(jid, "@g.us")
}

// IsBroadcastJID verifica se JID é de broadcast
func (m *MessageMapper) IsBroadcastJID(jid string) bool {
	return strings.Contains(jid, "@broadcast")
}

// ExtractGroupID extrai ID do grupo do JID
func (m *MessageMapper) ExtractGroupID(jid string) string {
	if !m.IsGroupJID(jid) {
		return ""
	}
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// MessageTypeToString converte tipo de mensagem para string
func (m *MessageMapper) MessageTypeToString(msgType string) string {
	switch msgType {
	case "text":
		return "Text"
	case "image":
		return "Image"
	case "audio":
		return "Audio"
	case "video":
		return "Video"
	case "document":
		return "Document"
	case "sticker":
		return "Sticker"
	case "location":
		return "Location"
	case "contact":
		return "Contact"
	default:
		return "Unknown"
	}
}
