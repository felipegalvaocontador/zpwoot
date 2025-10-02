package ports

import (
	"context"
	"time"

	"zpwoot/internal/domain/message"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, sessionID string, message *MessageEntity) error

	GetMessage(ctx context.Context, sessionID, messageID string) (*MessageEntity, error)

	ListMessages(ctx context.Context, sessionID string, req *ListMessagesRequest) (*ListMessagesResponse, error)

	UpdateMessage(ctx context.Context, sessionID string, message *MessageEntity) error

	DeleteMessage(ctx context.Context, sessionID, messageID string) error

	GetMessagesByChat(ctx context.Context, sessionID, chatJID string, limit, offset int) ([]*MessageEntity, error)

	GetMessageStats(ctx context.Context, sessionID string) (*MessageStats, error)

	BulkSaveMessages(ctx context.Context, sessionID string, messages []*MessageEntity) error

	SearchMessages(ctx context.Context, sessionID, query string, limit, offset int) ([]*MessageEntity, error)

	GetMessagesByType(ctx context.Context, sessionID, messageType string) ([]*MessageEntity, error)

	GetUnreadMessages(ctx context.Context, sessionID string) ([]*MessageEntity, error)

	MarkMessageAsRead(ctx context.Context, sessionID, messageID string) error
}

type MessageManager interface {
	SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string, contextInfo *message.ContextInfo) (*message.SendResult, error)

	SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error

	SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error)

	SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error)

	SendReaction(sessionID, to, messageID, reaction string) error

	SendPresence(sessionID, to, presence string) error

	EditMessage(sessionID, to, messageID, newText string) error

	MarkRead(sessionID, to, messageID string) error

	RevokeMessage(sessionID, to, messageID string) (*message.SendResult, error)

	ForwardMessage(sessionID, fromChat, toChat, messageID string) (*message.SendResult, error)

	DownloadMedia(sessionID, messageID string) ([]byte, error)

	GetMessageInfo(sessionID, messageID string) (*MessageInfo, error)

	SendLocation(sessionID, to string, latitude, longitude float64, name, address string) (*message.SendResult, error)

	SendContact(sessionID, to, contactName, contactPhone string) (*message.SendResult, error)

	SendSticker(sessionID, to string, sticker []byte) (*message.SendResult, error)
}

type MessageService interface {
	SendMessage(ctx context.Context, req *message.SendMessageRequest) (*message.SendMessageResponse, error)

	ListMessages(ctx context.Context, req *ListMessagesRequest) (*ListMessagesResponse, error)

	GetMessageStats(ctx context.Context, req *GetMessageStatsRequest) (*GetMessageStatsResponse, error)

	ProcessIncomingMessage(ctx context.Context, sessionID string, message *MessageEntity) error

	ValidateMessage(message *MessageEntity) error

	ValidateRecipient(jid string) error

	FormatMessage(content string, messageType string) (string, error)

	CheckMessageLimits(ctx context.Context, sessionID, recipientJID string) error

	ScheduleMessage(ctx context.Context, req *ScheduleMessageRequest) (*ScheduleMessageResponse, error)

	CancelScheduledMessage(ctx context.Context, sessionID, scheduleID string) error

	GetScheduledMessages(ctx context.Context, sessionID string) ([]*ScheduledMessage, error)

	ProcessMessageDelivery(ctx context.Context, sessionID, messageID, status string) error
}

type MessageEntity struct {
	Timestamp time.Time `json:"timestamp"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
	ToJID     string    `json:"to_jid"`
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	MediaURL  string    `json:"media_url,omitempty"`
	Caption   string    `json:"caption,omitempty"`
	FromJID   string    `json:"from_jid"`
	ChatJID   string    `json:"chat_jid"`
	SessionID string    `json:"session_id"`
	IsRead    bool      `json:"is_read"`
}

type ListMessagesRequest struct {
	FromDate  *time.Time `json:"from_date,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
	SessionID string     `json:"session_id"`
	ChatJID   string     `json:"chat_jid,omitempty"`
	Type      string     `json:"type,omitempty"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type ListMessagesResponse struct {
	Messages []*MessageEntity `json:"messages"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
	HasMore  bool             `json:"has_more"`
}

type MessageStats struct {
	TotalMessages    int64 `json:"total_messages"`
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	MediaMessages    int64 `json:"media_messages"`
	TextMessages     int64 `json:"text_messages"`
	UnreadMessages   int64 `json:"unread_messages"`
}

type GetMessageStatsRequest struct {
	FromDate  *time.Time `json:"from_date,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
	SessionID string     `json:"session_id"`
	ChatJID   string     `json:"chat_jid,omitempty"`
}

type GetMessageStatsResponse struct {
	Stats     *MessageStats `json:"stats"`
	SessionID string        `json:"session_id"`
	ChatJID   string        `json:"chat_jid,omitempty"`
}

type ScheduleMessageRequest struct {
	ScheduledAt time.Time                   `json:"scheduled_at"`
	Message     *message.SendMessageRequest `json:"message"`
	SessionID   string                      `json:"session_id"`
	To          string                      `json:"to"`
}

type ScheduleMessageResponse struct {
	ScheduleID  string    `json:"schedule_id"`
	SessionID   string    `json:"session_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"`
}

type ScheduledMessage struct {
	ScheduledAt time.Time                   `json:"scheduled_at"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
	Message     *message.SendMessageRequest `json:"message"`
	ID          string                      `json:"id"`
	SessionID   string                      `json:"session_id"`
	To          string                      `json:"to"`
	Status      string                      `json:"status"`
}
