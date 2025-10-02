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
	Content      string   `json:"content,omitempty" example:"Hello, World!"` // Alias para Text (compatibilidade)
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

// UpdateSyncStatusRequest DTO para atualização de status de sincronização
type UpdateSyncStatusRequest struct {
	SyncStatus       string `json:"sync_status" validate:"required,oneof=pending synced failed" example:"synced"`
	CwMessageID      *int   `json:"cw_message_id,omitempty" example:"123"`
	CwConversationID *int   `json:"cw_conversation_id,omitempty" example:"456"`
} // @name UpdateSyncStatusRequest

// SendImageMessageRequest DTO para envio de imagem
type SendImageMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_image_data"`
	Caption  string `json:"caption,omitempty" example:"Check this image!"`
	Filename string `json:"filename,omitempty" example:"image.jpg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendImageMessageRequest

// SendAudioMessageRequest DTO para envio de áudio
type SendAudioMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_audio_data"`
	Caption  string `json:"caption,omitempty" example:"Audio message"`
	Filename string `json:"filename,omitempty" example:"audio.mp3"`
	MimeType string `json:"mime_type,omitempty" example:"audio/mpeg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendAudioMessageRequest

// SendVideoMessageRequest DTO para envio de vídeo
type SendVideoMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_video_data"`
	Caption  string `json:"caption,omitempty" example:"Check this video!"`
	Filename string `json:"filename,omitempty" example:"video.mp4"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendVideoMessageRequest

// SendDocumentMessageRequest DTO para envio de documento
type SendDocumentMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_document_data"`
	Caption  string `json:"caption,omitempty" example:"Document"`
	Filename string `json:"filename" validate:"required" example:"document.pdf"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendDocumentMessageRequest

// SendStickerMessageRequest DTO para envio de sticker
type SendStickerMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_sticker_data"`
	MimeType string `json:"mime_type,omitempty" example:"image/webp"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendStickerMessageRequest

// SendLocationMessageRequest DTO para envio de localização
type SendLocationMessageRequest struct {
	To        string  `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Latitude  float64 `json:"latitude" validate:"required" example:"-23.5505"`
	Longitude float64 `json:"longitude" validate:"required" example:"-46.6333"`
	Name      string  `json:"name,omitempty" example:"São Paulo"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP, Brazil"`
	ReplyTo   string  `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendLocationMessageRequest

// SendContactMessageRequest DTO para envio de contato
type SendContactMessageRequest struct {
	To           string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Name         string `json:"name" validate:"required" example:"John Doe"`
	Phone        string `json:"phone" validate:"required" example:"+5511888888888"`
	ContactName  string `json:"contact_name,omitempty" example:"John Doe"` // Alias para Name
	ContactPhone string `json:"contact_phone,omitempty" example:"+5511888888888"` // Alias para Phone
	ReplyTo      string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendContactMessageRequest

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
	To          string     `json:"to" example:"5511999999999@s.whatsapp.net"`
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

// ===== ADDITIONAL REQUEST TYPES =====

// SendContactListMessageRequest DTO para envio de lista de contatos
type SendContactListMessageRequest struct {
	To       string        `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Contacts []ContactInfo `json:"contacts" validate:"required,min=1"`
	ReplyTo  string        `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendContactListMessageRequest

// ContactInfo informações de contato
type ContactInfo struct {
	Name  string `json:"name" validate:"required" example:"John Doe"`
	Phone string `json:"phone" validate:"required" example:"+5511888888888"`
} // @name ContactInfo

// ContactResult resultado de operação com contato
type ContactResult struct {
	Name        string `json:"name" example:"John Doe"`
	Phone       string `json:"phone" example:"+5511888888888"`
	ContactName string `json:"contact_name,omitempty" example:"John Doe"` // Alias para Name
	MessageID   string `json:"message_id,omitempty" example:"3EB0C767D71D"`
	Status      string `json:"status" example:"sent"`
	Success     bool   `json:"success" example:"true"`
	Error       string `json:"error,omitempty" example:""`
} // @name ContactResult

// SendContactListResponse resposta para envio de lista de contatos
type SendContactListResponse struct {
	BaseResponse
	SessionID       string          `json:"session_id,omitempty" example:"session-123"`
	RemoteJID       string          `json:"remote_jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	ContactCount    int             `json:"contact_count" example:"3"`
	ContactResults  []ContactResult `json:"contact_results"`
	Results         []ContactResult `json:"results"` // Alias para ContactResults
	SentAt          string          `json:"sent_at,omitempty" example:"2024-01-01T12:00:00Z"`
} // @name SendContactListResponse

// SendBusinessProfileMessageRequest DTO para envio de perfil business
type SendBusinessProfileMessageRequest struct {
	To      string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ReplyTo string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendBusinessProfileMessageRequest

// SendButtonMessageRequest DTO para envio de mensagem com botões
type SendButtonMessageRequest struct {
	To      string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Text    string       `json:"text" validate:"required" example:"Choose an option:"`
	Footer  string       `json:"footer,omitempty" example:"Powered by ZPWoot"`
	Buttons []ButtonInfo `json:"buttons" validate:"required,min=1,max=3"`
	ReplyTo string       `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendButtonMessageRequest

// SendListMessageRequest DTO para envio de mensagem com lista
type SendListMessageRequest struct {
	To         string            `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body       string            `json:"body" validate:"required" example:"Choose from the list:"`
	ButtonText string            `json:"button_text" validate:"required" example:"View Options"`
	Sections   []ListSectionInfo `json:"sections" validate:"required,min=1,max=10"`
	ReplyTo    string            `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendListMessageRequest

// SendPollMessageRequest DTO para envio de enquete
type SendPollMessageRequest struct {
	To              string           `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Question        string           `json:"question" validate:"required" example:"What's your favorite color?"`
	Options         []PollOptionInfo `json:"options" validate:"required,min=2,max=12"`
	SelectableCount int              `json:"selectable_count" validate:"min=1" example:"1"`
	ReplyTo         string           `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendPollMessageRequest

// SendReactionMessageRequest DTO para envio de reação
type SendReactionMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"message_id" validate:"required" example:"3EB0C767D71D"`
	Reaction  string `json:"reaction" validate:"required" example:"👍"`
} // @name SendReactionMessageRequest

// SendPresenceMessageRequest DTO para envio de presença
type SendPresenceMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Presence string `json:"presence" validate:"required,oneof=typing recording online offline paused" example:"typing"`
} // @name SendPresenceMessageRequest

// ===== AUXILIARY TYPES =====

// ButtonInfo informações de botão
type ButtonInfo struct {
	ID   string `json:"id" validate:"required" example:"btn-1"`
	Text string `json:"text" validate:"required" example:"Option 1"`
	Type string `json:"type,omitempty" example:"reply"`
} // @name ButtonInfo

// ListSectionInfo informações de seção de lista
type ListSectionInfo struct {
	Title string        `json:"title" validate:"required" example:"Section 1"`
	Rows  []ListRowInfo `json:"rows" validate:"required,min=1,max=10"`
} // @name ListSectionInfo

// ListRowInfo informações de linha de lista
type ListRowInfo struct {
	ID          string `json:"id" validate:"required" example:"row-1"`
	Title       string `json:"title" validate:"required" example:"Option 1"`
	Description string `json:"description,omitempty" example:"Description of option 1"`
} // @name ListRowInfo

// PollOptionInfo informações de opção de enquete
type PollOptionInfo struct {
	Name string `json:"name" validate:"required" example:"Option 1"`
} // @name PollOptionInfo
