package messaging

import (
	"time"

	"github.com/google/uuid"
)

// Message representa uma mensagem no sistema zpwoot
// Mapeia mensagens entre WhatsApp e Chatwoot
type Message struct {
	// Identificadores únicos
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`

	// WhatsApp Message Identifiers (from whatsmeow)
	ZpMessageID string    `json:"zp_message_id"`
	ZpSender    string    `json:"zp_sender"`
	ZpChat      string    `json:"zp_chat"`
	ZpTimestamp time.Time `json:"zp_timestamp"`
	ZpFromMe    bool      `json:"zp_from_me"`
	ZpType      string    `json:"zp_type"` // text, image, audio, video, document, contact, etc.
	Content     string    `json:"content,omitempty"`

	// Chatwoot Message Identifiers
	CwMessageID      *int `json:"cw_message_id,omitempty"`
	CwConversationID *int `json:"cw_conversation_id,omitempty"`

	// Sync Status
	SyncStatus string     `json:"sync_status"` // pending, synced, failed
	SyncedAt   *time.Time `json:"synced_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageType constantes para tipos de mensagem
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeDocument MessageType = "document"
	MessageTypeContact  MessageType = "contact"
	MessageTypeLocation MessageType = "location"
	MessageTypeSticker  MessageType = "sticker"
)

// SyncStatus constantes para status de sincronização
type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusFailed  SyncStatus = "failed"
)

// CreateMessageRequest dados para criação de mensagem
type CreateMessageRequest struct {
	SessionID   uuid.UUID   `json:"session_id" validate:"required"`
	ZpMessageID string      `json:"zp_message_id" validate:"required"`
	ZpSender    string      `json:"zp_sender" validate:"required"`
	ZpChat      string      `json:"zp_chat" validate:"required"`
	ZpTimestamp time.Time   `json:"zp_timestamp" validate:"required"`
	ZpFromMe    bool        `json:"zp_from_me"`
	ZpType      MessageType `json:"zp_type" validate:"required"`
	Content     string      `json:"content,omitempty"`
}

// UpdateSyncStatusRequest dados para atualização de status de sync
type UpdateSyncStatusRequest struct {
	ID               uuid.UUID   `json:"id" validate:"required"`
	SyncStatus       SyncStatus  `json:"sync_status" validate:"required"`
	CwMessageID      *int        `json:"cw_message_id,omitempty"`
	CwConversationID *int        `json:"cw_conversation_id,omitempty"`
	SyncedAt         *time.Time  `json:"synced_at,omitempty"`
}

// ListMessagesRequest dados para listagem de mensagens
type ListMessagesRequest struct {
	SessionID string `json:"session_id,omitempty"`
	ChatJID   string `json:"chat_jid,omitempty"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}

// MessageStats estatísticas de mensagens
type MessageStats struct {
	TotalMessages     int64            `json:"total_messages"`
	MessagesByType    map[string]int64 `json:"messages_by_type"`
	MessagesByStatus  map[string]int64 `json:"messages_by_status"`
	SyncedMessages    int64            `json:"synced_messages"`
	PendingMessages   int64            `json:"pending_messages"`
	FailedMessages    int64            `json:"failed_messages"`
	MessagesToday     int64            `json:"messages_today"`
	MessagesThisWeek  int64            `json:"messages_this_week"`
	MessagesThisMonth int64            `json:"messages_this_month"`
}

// IsValidMessageType verifica se o tipo de mensagem é válido
func IsValidMessageType(msgType string) bool {
	switch MessageType(msgType) {
	case MessageTypeText, MessageTypeImage, MessageTypeAudio, 
		 MessageTypeVideo, MessageTypeDocument, MessageTypeContact,
		 MessageTypeLocation, MessageTypeSticker:
		return true
	default:
		return false
	}
}

// IsValidSyncStatus verifica se o status de sync é válido
func IsValidSyncStatus(status string) bool {
	switch SyncStatus(status) {
	case SyncStatusPending, SyncStatusSynced, SyncStatusFailed:
		return true
	default:
		return false
	}
}

// IsSynced verifica se a mensagem está sincronizada
func (m *Message) IsSynced() bool {
	return m.SyncStatus == string(SyncStatusSynced) && m.CwMessageID != nil
}

// IsPending verifica se a mensagem está pendente de sincronização
func (m *Message) IsPending() bool {
	return m.SyncStatus == string(SyncStatusPending)
}

// IsFailed verifica se a sincronização falhou
func (m *Message) IsFailed() bool {
	return m.SyncStatus == string(SyncStatusFailed)
}

// HasChatwootData verifica se a mensagem tem dados do Chatwoot
func (m *Message) HasChatwootData() bool {
	return m.CwMessageID != nil && m.CwConversationID != nil
}

// GetMessageTypeString retorna o tipo da mensagem como string
func (m *Message) GetMessageTypeString() string {
	return m.ZpType
}

// GetSyncStatusString retorna o status de sync como string
func (m *Message) GetSyncStatusString() string {
	return m.SyncStatus
}
