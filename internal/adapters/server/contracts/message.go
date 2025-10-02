package contracts

import (
	"time"
)

// ===== MESSAGE REQUESTS =====

// CreateMessageRequest DTO para criação de mensagem
type CreateMessageRequest struct {
	ZpMessageID string `json:"zp_message_id" validate:"required" example:"3EB0C767D71D"`
	ZpSender    string `json:"zp_sender" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ZpChat      string `json:"zp_chat" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ZpTimestamp string `json:"zp_timestamp" validate:"required" example:"2024-01-01T12:00:00Z"`
	ZpFromMe    bool   `json:"zp_from_me" example:"false"`
	ZpType      string `json:"zp_type" validate:"required" example:"text"`
	Content     string `json:"content,omitempty" example:"Hello World"`
} // @name CreateMessageRequest

// ListMessagesRequest DTO para listagem de mensagens
type ListMessagesRequest struct {
	PaginationRequest
	ChatJID     string `json:"chat_jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	MessageType string `json:"message_type,omitempty" example:"text"`
	FromMe      *bool  `json:"from_me,omitempty" example:"false"`
	DateFrom    string `json:"date_from,omitempty" example:"2024-01-01"`
	DateTo      string `json:"date_to,omitempty" example:"2024-01-31"`
} // @name ListMessagesRequest

// SendTextMessageRequest DTO para envio de mensagem de texto
type SendTextMessageRequest struct {
	To           string   `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Text         string   `json:"text" validate:"required" example:"Hello, World!"`
	ReplyTo      string   `json:"reply_to,omitempty" example:"3EB0C767D71D"`
	MentionedJID []string `json:"mentioned_jid,omitempty" example:"[\"5511888888888@s.whatsapp.net\"]"`
	LinkPreview  bool     `json:"link_preview" example:"true"`
} // @name SendTextMessageRequest

// SendMediaMessageRequest DTO para envio de mensagem de mídia
type SendMediaMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MediaURL string `json:"media_url" validate:"required,url" example:"https://example.com/image.jpg"`
	Type     string `json:"type" validate:"required,oneof=image audio video document" example:"image"`
	Caption  string `json:"caption,omitempty" example:"Check this out!"`
	Filename string `json:"filename,omitempty" example:"image.jpg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendMediaMessageRequest

// ===== MESSAGE RESPONSES =====

// CreateMessageResponse DTO de resposta para criação de mensagem
type CreateMessageResponse struct {
	BaseResponse
	Message *MessageInfo `json:"message"`
} // @name CreateMessageResponse

// ListMessagesResponse DTO de resposta para listagem de mensagens
type ListMessagesResponse struct {
	ListResponse
	Messages []MessageInfo `json:"messages"`
} // @name ListMessagesResponse

// SendMessageResponse DTO de resposta para envio de mensagem
type SendMessageResponse struct {
	BaseResponse
	MessageID   string     `json:"message_id" example:"3EB0C767D71D"`
	Timestamp   time.Time  `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Status      string     `json:"status" example:"sent"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty" example:"2024-01-01T12:00:05Z"`
	ReadAt      *time.Time `json:"read_at,omitempty" example:"2024-01-01T12:00:10Z"`
} // @name SendMessageResponse

// ===== NESTED TYPES =====

// MessageInfo informações completas de uma mensagem
type MessageInfo struct {
	ID               string     `json:"id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	SessionID        string     `json:"session_id" example:"session-123"`
	ZpMessageID      string     `json:"zp_message_id" example:"3EB0C767D71D"`
	ZpSender         string     `json:"zp_sender" example:"5511999999999@s.whatsapp.net"`
	ZpChat           string     `json:"zp_chat" example:"5511999999999@s.whatsapp.net"`
	ZpTimestamp      time.Time  `json:"zp_timestamp" example:"2024-01-01T12:00:00Z"`
	ZpFromMe         bool       `json:"zp_from_me" example:"false"`
	ZpType           string     `json:"zp_type" example:"text"`
	Content          string     `json:"content,omitempty" example:"Hello World"`
	MediaURL         string     `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
	MediaType        string     `json:"media_type,omitempty" example:"image"`
	CwMessageID      *int       `json:"cw_message_id,omitempty" example:"123"`
	CwConversationID *int       `json:"cw_conversation_id,omitempty" example:"456"`
	SyncStatus       string     `json:"sync_status" example:"synced"`
	SyncedAt         *time.Time `json:"synced_at,omitempty" example:"2024-01-01T12:00:05Z"`
	CreatedAt        time.Time  `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt        time.Time  `json:"updated_at" example:"2024-01-01T12:00:00Z"`
} // @name MessageInfo

// MessageDTO alias para MessageInfo (compatibilidade com services)
type MessageDTO = MessageInfo

// MessageStats estatísticas de mensagens
type MessageStats struct {
	TotalMessages     int64            `json:"total_messages" example:"1000"`
	MessagesByType    map[string]int64 `json:"messages_by_type"`
	MessagesByStatus  map[string]int64 `json:"messages_by_status"`
	SyncedMessages    int64            `json:"synced_messages" example:"950"`
	PendingMessages   int64            `json:"pending_messages" example:"30"`
	FailedMessages    int64            `json:"failed_messages" example:"20"`
	MessagesToday     int64            `json:"messages_today" example:"50"`
	MessagesThisWeek  int64            `json:"messages_this_week" example:"300"`
	MessagesThisMonth int64            `json:"messages_this_month" example:"1000"`
	AveragePerDay     float64          `json:"average_per_day" example:"33.3"`
	PeakHour          int              `json:"peak_hour" example:"14"`
} // @name MessageStats
