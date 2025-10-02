package messaging

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`

	ZpMessageID string    `json:"zp_message_id"`
	ZpSender    string    `json:"zp_sender"`
	ZpChat      string    `json:"zp_chat"`
	ZpTimestamp time.Time `json:"zp_timestamp"`
	ZpFromMe    bool      `json:"zp_from_me"`
	ZpType      string    `json:"zp_type"`
	Content     string    `json:"content,omitempty"`

	CwMessageID      *int `json:"cw_message_id,omitempty"`
	CwConversationID *int `json:"cw_conversation_id,omitempty"`

	SyncStatus string     `json:"sync_status"`
	SyncedAt   *time.Time `json:"synced_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusFailed  SyncStatus = "failed"
)

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

type UpdateSyncStatusRequest struct {
	ID               uuid.UUID  `json:"id" validate:"required"`
	SyncStatus       SyncStatus `json:"sync_status" validate:"required"`
	CwMessageID      *int       `json:"cw_message_id,omitempty"`
	CwConversationID *int       `json:"cw_conversation_id,omitempty"`
	SyncedAt         *time.Time `json:"synced_at,omitempty"`
}

type ListMessagesRequest struct {
	SessionID string `json:"session_id,omitempty"`
	ChatJID   string `json:"chat_jid,omitempty"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}

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

func IsValidSyncStatus(status string) bool {
	switch SyncStatus(status) {
	case SyncStatusPending, SyncStatusSynced, SyncStatusFailed:
		return true
	default:
		return false
	}
}

func (m *Message) IsSynced() bool {
	return m.SyncStatus == string(SyncStatusSynced) && m.CwMessageID != nil
}

func (m *Message) IsPending() bool {
	return m.SyncStatus == string(SyncStatusPending)
}

func (m *Message) IsFailed() bool {
	return m.SyncStatus == string(SyncStatusFailed)
}

func (m *Message) HasChatwootData() bool {
	return m.CwMessageID != nil && m.CwConversationID != nil
}

func (m *Message) GetMessageTypeString() string {
	return m.ZpType
}

func (m *Message) GetSyncStatusString() string {
	return m.SyncStatus
}
