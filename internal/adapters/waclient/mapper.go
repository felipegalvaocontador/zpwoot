package waclient

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"zpwoot/internal/core/session"
)

type MessageMapper struct{}

func NewMessageMapper() *MessageMapper {
	return &MessageMapper{}
}

func (m *MessageMapper) EventToWhatsAppMessage(evt *events.Message) *session.WhatsAppMessage {
	if evt == nil {
		return nil
	}

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
			"push_name":    evt.Info.PushName,
			"message_type": evt.Info.Type,
			"category":     evt.Info.Category,
		},
	}
}

func (m *MessageMapper) extractMessageContent(message *waE2E.Message) (string, string) {
	if message == nil {
		return "", "unknown"
	}

	if content, msgType := m.extractTextContent(message); msgType != "" {
		return content, msgType
	}

	if content, msgType := m.extractMediaContent(message); msgType != "" {
		return content, msgType
	}

	if content, msgType := m.extractSpecialContent(message); msgType != "" {
		return content, msgType
	}

	return "[Unsupported message type]", "unknown"
}

func (m *MessageMapper) extractTextContent(message *waE2E.Message) (string, string) {
	if message.Conversation != nil {
		return *message.Conversation, "text"
	}

	if message.ExtendedTextMessage != nil && message.ExtendedTextMessage.Text != nil {
		return *message.ExtendedTextMessage.Text, "text"
	}

	return "", ""
}

func (m *MessageMapper) extractMediaContent(message *waE2E.Message) (string, string) {
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

	return "", ""
}

func (m *MessageMapper) extractSpecialContent(message *waE2E.Message) (string, string) {
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

	return "", ""
}

func (m *MessageMapper) JIDToPhoneNumber(jid string) string {
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}

func (m *MessageMapper) PhoneNumberToJID(phoneNumber string) types.JID {
	validator := NewValidator()
	cleanNumber := validator.CleanPhoneNumber(phoneNumber)

	return types.JID{
		User:   cleanNumber,
		Server: types.DefaultUserServer,
	}
}

func (m *MessageMapper) FormatJID(jid types.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	return jid.String()
}

func (m *MessageMapper) IsGroupJID(jid string) bool {
	return strings.Contains(jid, "@g.us")
}

func (m *MessageMapper) IsBroadcastJID(jid string) bool {
	return strings.Contains(jid, "@broadcast")
}

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
