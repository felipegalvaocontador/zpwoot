package ports

import (
	"context"
	"time"

	"zpwoot/internal/domain/message"
)

// MessageRepository defines the interface for message data operations
type MessageRepository interface {
	// SaveMessage saves a message to local storage
	SaveMessage(ctx context.Context, sessionID string, message *MessageEntity) error

	// GetMessage retrieves a message by ID
	GetMessage(ctx context.Context, sessionID, messageID string) (*MessageEntity, error)

	// ListMessages lists messages with pagination and filtering
	ListMessages(ctx context.Context, sessionID string, req *ListMessagesRequest) (*ListMessagesResponse, error)

	// UpdateMessage updates an existing message
	UpdateMessage(ctx context.Context, sessionID string, message *MessageEntity) error

	// DeleteMessage removes a message from local storage
	DeleteMessage(ctx context.Context, sessionID, messageID string) error

	// GetMessagesByChat gets messages for a specific chat
	GetMessagesByChat(ctx context.Context, sessionID, chatJID string, limit, offset int) ([]*MessageEntity, error)

	// GetMessageStats returns message statistics
	GetMessageStats(ctx context.Context, sessionID string) (*MessageStats, error)

	// BulkSaveMessages saves multiple messages in a single operation
	BulkSaveMessages(ctx context.Context, sessionID string, messages []*MessageEntity) error

	// SearchMessages searches messages by content
	SearchMessages(ctx context.Context, sessionID, query string, limit, offset int) ([]*MessageEntity, error)

	// GetMessagesByType returns messages filtered by type
	GetMessagesByType(ctx context.Context, sessionID, messageType string) ([]*MessageEntity, error)

	// GetUnreadMessages gets unread messages for a session
	GetUnreadMessages(ctx context.Context, sessionID string) ([]*MessageEntity, error)

	// MarkMessageAsRead marks a message as read in local storage
	MarkMessageAsRead(ctx context.Context, sessionID, messageID string) error
}

// MessageManager defines the interface for WhatsApp message operations
type MessageManager interface {
	// SendMessage sends a text message
	SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string, contextInfo *message.ContextInfo) (*message.SendResult, error)

	// SendMediaMessage sends a media message (image, video, audio, document)
	SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error

	// SendButtonMessage sends a message with interactive buttons
	SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error)

	// SendListMessage sends a message with interactive list
	SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error)

	// SendReaction sends a reaction to a message
	SendReaction(sessionID, to, messageID, reaction string) error

	// SendPresence sends presence information (typing, recording, etc.)
	SendPresence(sessionID, to, presence string) error

	// EditMessage edits an existing message
	EditMessage(sessionID, to, messageID, newText string) error

	// MarkRead marks a message as read
	MarkRead(sessionID, to, messageID string) error

	// RevokeMessage revokes/deletes a message
	RevokeMessage(sessionID, to, messageID string) (*message.SendResult, error)

	// ForwardMessage forwards a message to another chat
	ForwardMessage(sessionID, fromChat, toChat, messageID string) (*message.SendResult, error)

	// DownloadMedia downloads media from a message
	DownloadMedia(sessionID, messageID string) ([]byte, error)

	// GetMessageInfo gets detailed information about a message
	GetMessageInfo(sessionID, messageID string) (*MessageInfo, error)

	// SendLocation sends a location message
	SendLocation(sessionID, to string, latitude, longitude float64, name, address string) (*message.SendResult, error)

	// SendContact sends a contact message
	SendContact(sessionID, to, contactName, contactPhone string) (*message.SendResult, error)

	// SendSticker sends a sticker message
	SendSticker(sessionID, to string, sticker []byte) (*message.SendResult, error)
}

// MessageService defines the interface for message business logic
type MessageService interface {
	// SendMessage handles message sending with validation and processing
	SendMessage(ctx context.Context, req *message.SendMessageRequest) (*message.SendMessageResponse, error)

	// ListMessages lists and filters messages with business logic
	ListMessages(ctx context.Context, req *ListMessagesRequest) (*ListMessagesResponse, error)

	// GetMessageStats calculates and returns message statistics
	GetMessageStats(ctx context.Context, req *GetMessageStatsRequest) (*GetMessageStatsResponse, error)

	// ProcessIncomingMessage processes incoming messages with business rules
	ProcessIncomingMessage(ctx context.Context, sessionID string, message *MessageEntity) error

	// ValidateMessage validates message content and format
	ValidateMessage(message *MessageEntity) error

	// ValidateRecipient validates recipient JID format
	ValidateRecipient(jid string) error

	// FormatMessage formats message content according to business rules
	FormatMessage(content string, messageType string) (string, error)

	// CheckMessageLimits checks if message sending limits are respected
	CheckMessageLimits(ctx context.Context, sessionID, recipientJID string) error

	// ScheduleMessage schedules a message for later sending
	ScheduleMessage(ctx context.Context, req *ScheduleMessageRequest) (*ScheduleMessageResponse, error)

	// CancelScheduledMessage cancels a scheduled message
	CancelScheduledMessage(ctx context.Context, sessionID, scheduleID string) error

	// GetScheduledMessages gets list of scheduled messages
	GetScheduledMessages(ctx context.Context, sessionID string) ([]*ScheduledMessage, error)

	// ProcessMessageDelivery processes message delivery status updates
	ProcessMessageDelivery(ctx context.Context, sessionID, messageID, status string) error
}

// MessageEntity represents a message entity for repository operations
type MessageEntity struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	ChatJID   string    `json:"chat_jid"`
	FromJID   string    `json:"from_jid"`
	ToJID     string    `json:"to_jid"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	MediaURL  string    `json:"media_url,omitempty"`
	Caption   string    `json:"caption,omitempty"`
	IsRead    bool      `json:"is_read"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListMessagesRequest represents a request to list messages
type ListMessagesRequest struct {
	SessionID string     `json:"session_id"`
	ChatJID   string     `json:"chat_jid,omitempty"`
	Type      string     `json:"type,omitempty"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	FromDate  *time.Time `json:"from_date,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
}

// ListMessagesResponse represents a response with message list
type ListMessagesResponse struct {
	Messages []*MessageEntity `json:"messages"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
	HasMore  bool             `json:"has_more"`
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalMessages    int64 `json:"total_messages"`
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	MediaMessages    int64 `json:"media_messages"`
	TextMessages     int64 `json:"text_messages"`
	UnreadMessages   int64 `json:"unread_messages"`
}

// GetMessageStatsRequest represents a request to get message statistics
type GetMessageStatsRequest struct {
	SessionID string     `json:"session_id"`
	ChatJID   string     `json:"chat_jid,omitempty"`
	FromDate  *time.Time `json:"from_date,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
}

// GetMessageStatsResponse represents a response with message statistics
type GetMessageStatsResponse struct {
	Stats     *MessageStats `json:"stats"`
	SessionID string        `json:"session_id"`
	ChatJID   string        `json:"chat_jid,omitempty"`
}

// ScheduleMessageRequest represents a request to schedule a message
type ScheduleMessageRequest struct {
	SessionID   string                      `json:"session_id"`
	To          string                      `json:"to"`
	Message     *message.SendMessageRequest `json:"message"`
	ScheduledAt time.Time                   `json:"scheduled_at"`
}

// ScheduleMessageResponse represents a response for scheduled message
type ScheduleMessageResponse struct {
	ScheduleID  string    `json:"schedule_id"`
	SessionID   string    `json:"session_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"`
}

// ScheduledMessage represents a scheduled message
type ScheduledMessage struct {
	ID          string                      `json:"id"`
	SessionID   string                      `json:"session_id"`
	To          string                      `json:"to"`
	Message     *message.SendMessageRequest `json:"message"`
	ScheduledAt time.Time                   `json:"scheduled_at"`
	Status      string                      `json:"status"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}
