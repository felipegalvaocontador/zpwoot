package dto

import (
	"fmt"
	"time"
)

// ===== REQUEST DTOs =====

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

// SendTextMessageRequest DTO para envio de mensagem de texto
type SendTextMessageRequest struct {
	To      string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Content string `json:"content" validate:"required" example:"Hello World"`
	ReplyTo string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendTextMessageRequest

// SendMediaMessageRequest DTO para envio de mídia genérica
type SendMediaMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MediaURL string `json:"media_url" validate:"required,url" example:"https://example.com/media.jpg"`
	Caption  string `json:"caption,omitempty" example:"Media caption"`
	Type     string `json:"type" validate:"required,oneof=image video audio document" example:"image"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendMediaMessageRequest

// SendImageMessageRequest DTO para envio de imagem
type SendImageMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"https://example.com/image.jpg"`
	Caption  string `json:"caption,omitempty" example:"Beautiful image"`
	MimeType string `json:"mime_type,omitempty" example:"image/jpeg"`
	Filename string `json:"filename,omitempty" example:"image.jpg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendImageMessageRequest

// SendAudioMessageRequest DTO para envio de áudio
type SendAudioMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"https://example.com/audio.mp3"`
	Caption  string `json:"caption,omitempty" example:"Voice message"`
	MimeType string `json:"mime_type,omitempty" example:"audio/mpeg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendAudioMessageRequest

// SendVideoMessageRequest DTO para envio de vídeo
type SendVideoMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"https://example.com/video.mp4"`
	Caption  string `json:"caption,omitempty" example:"Amazing video"`
	MimeType string `json:"mime_type,omitempty" example:"video/mp4"`
	Filename string `json:"filename,omitempty" example:"video.mp4"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendVideoMessageRequest

// SendDocumentMessageRequest DTO para envio de documento
type SendDocumentMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"https://example.com/document.pdf"`
	Caption  string `json:"caption,omitempty" example:"Important document"`
	MimeType string `json:"mime_type,omitempty" example:"application/pdf"`
	Filename string `json:"filename" validate:"required" example:"document.pdf"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendDocumentMessageRequest

// SendStickerMessageRequest DTO para envio de sticker
type SendStickerMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"https://example.com/sticker.webp"`
	MimeType string `json:"mime_type,omitempty" example:"image/webp"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendStickerMessageRequest

// SendLocationMessageRequest DTO para envio de localização
type SendLocationMessageRequest struct {
	To        string  `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Latitude  float64 `json:"latitude" validate:"required" example:"-23.5505"`
	Longitude float64 `json:"longitude" validate:"required" example:"-46.6333"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP, Brasil"`
	ReplyTo   string  `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendLocationMessageRequest

// SendContactMessageRequest DTO para envio de contato
type SendContactMessageRequest struct {
	To           string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ContactName  string `json:"contact_name" validate:"required" example:"John Doe"`
	ContactPhone string `json:"contact_phone" validate:"required" example:"+5511888888888"`
	ReplyTo      string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendContactMessageRequest

// ContactInfo DTO para informações de contato
type ContactInfo struct {
	Name  string `json:"name" validate:"required" example:"John Doe"`
	Phone string `json:"phone" validate:"required" example:"+5511888888888"`
} // @name ContactInfo

// SendContactListMessageRequest DTO para envio de lista de contatos
type SendContactListMessageRequest struct {
	To       string        `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Contacts []ContactInfo `json:"contacts" validate:"required,min=1,max=5"`
	ReplyTo  string        `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendContactListMessageRequest

// SendBusinessProfileMessageRequest DTO para envio de perfil business
type SendBusinessProfileMessageRequest struct {
	To          string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	BusinessJID string `json:"business_jid" validate:"required" example:"5511777777777@s.whatsapp.net"`
	ReplyTo     string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendBusinessProfileMessageRequest

// ButtonInfo DTO para informações de botão
type ButtonInfo struct {
	ID   string `json:"id" validate:"required" example:"btn_1"`
	Text string `json:"text" validate:"required" example:"Click Me"`
} // @name ButtonInfo

// SendButtonMessageRequest DTO para envio de mensagem com botões
type SendButtonMessageRequest struct {
	To      string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body    string       `json:"body" validate:"required" example:"Choose an option:"`
	Buttons []ButtonInfo `json:"buttons" validate:"required,min=1,max=3"`
	ReplyTo string       `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendButtonMessageRequest

// ListRowInfo DTO para informações de linha da lista
type ListRowInfo struct {
	ID          string `json:"id" validate:"required" example:"row_1"`
	Title       string `json:"title" validate:"required" example:"Option 1"`
	Description string `json:"description,omitempty" example:"Description for option 1"`
} // @name ListRowInfo

// ListSectionInfo DTO para informações de seção da lista
type ListSectionInfo struct {
	Title string        `json:"title" validate:"required" example:"Section 1"`
	Rows  []ListRowInfo `json:"rows" validate:"required,min=1"`
} // @name ListSectionInfo

// SendListMessageRequest DTO para envio de mensagem com lista
type SendListMessageRequest struct {
	To         string            `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body       string            `json:"body" validate:"required" example:"Choose from the list:"`
	ButtonText string            `json:"button_text" validate:"required" example:"View Options"`
	Sections   []ListSectionInfo `json:"sections" validate:"required,min=1,max=10"`
	ReplyTo    string            `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendListMessageRequest

// PollOptionInfo DTO para informações de opção de poll
type PollOptionInfo struct {
	Name string `json:"name" validate:"required" example:"Option 1"`
} // @name PollOptionInfo

// SendPollMessageRequest DTO para envio de poll
type SendPollMessageRequest struct {
	To                string           `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Name              string           `json:"name" validate:"required" example:"What's your favorite color?"`
	Options           []PollOptionInfo `json:"options" validate:"required,min=2,max=12"`
	SelectableCount   int              `json:"selectable_count,omitempty" example:"1"`
	AllowMultipleVote bool             `json:"allow_multiple_vote,omitempty" example:"false"`
	ReplyTo           string           `json:"reply_to,omitempty" example:"3EB0C767D71D"`
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

// EditMessageRequest DTO para edição de mensagem
type EditMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"message_id" validate:"required" example:"3EB0C767D71D"`
	NewBody   string `json:"new_body" validate:"required" example:"Updated message text"`
} // @name EditMessageRequest

// RevokeMessageRequest DTO para revogação de mensagem
type RevokeMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"message_id" validate:"required" example:"3EB0C767D71D"`
} // @name RevokeMessageRequest

// MarkAsReadRequest DTO para marcar mensagens como lidas
type MarkAsReadRequest struct {
	ChatJID    string   `json:"chat_jid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageIDs []string `json:"message_ids" validate:"required,min=1" example:"[\"3EB0C767D71D\"]"`
} // @name MarkAsReadRequest

// ListMessagesRequest DTO para listagem de mensagens
type ListMessagesRequest struct {
	ChatJID string `json:"chat_jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	Limit   int    `json:"limit" validate:"min=1,max=100" example:"50"`
	Offset  int    `json:"offset" validate:"min=0" example:"0"`
} // @name ListMessagesRequest

// UpdateSyncStatusRequest DTO para atualização de status de sync
type UpdateSyncStatusRequest struct {
	SyncStatus       string `json:"sync_status" validate:"required,oneof=pending synced failed" example:"synced"`
	CwMessageID      *int   `json:"cw_message_id,omitempty" example:"123"`
	CwConversationID *int   `json:"cw_conversation_id,omitempty" example:"456"`
} // @name UpdateSyncStatusRequest

// ===== RESPONSE DTOs =====

// CreateMessageResponse DTO de resposta para criação de mensagem
type CreateMessageResponse struct {
	ID          string    `json:"id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	SessionID   string    `json:"session_id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	ZpMessageID string    `json:"zp_message_id" example:"3EB0C767D71D"`
	SyncStatus  string    `json:"sync_status" example:"pending"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
} // @name CreateMessageResponse

// SendMessageResponse DTO de resposta para envio de mensagem
type SendMessageResponse struct {
	MessageID string    `json:"message_id" example:"3EB0C767D71D"`
	To        string    `json:"to" example:"5511999999999@s.whatsapp.net"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} // @name SendMessageResponse

// ListMessagesResponse DTO de resposta para listagem de mensagens
type ListMessagesResponse struct {
	Messages []*MessageDTO `json:"messages"`
	Total    int64         `json:"total" example:"150"`
	Limit    int           `json:"limit" example:"50"`
	Offset   int           `json:"offset" example:"0"`
} // @name ListMessagesResponse

// MarkAsReadResponse DTO de resposta para marcar como lido
type MarkAsReadResponse struct {
	ChatJID     string `json:"chat_jid" example:"5511999999999@s.whatsapp.net"`
	MarkedCount int    `json:"marked_count" example:"5"`
	Status      string `json:"status" example:"success"`
} // @name MarkAsReadResponse

// ContactResult DTO para resultado de envio de contato
type ContactResult struct {
	ContactName string `json:"contact_name" example:"John Doe"`
	MessageID   string `json:"message_id" example:"3EB0C767D71D"`
	Status      string `json:"status" example:"sent"`
} // @name ContactResult

// SendContactListResponse DTO de resposta para envio de lista de contatos
type SendContactListResponse struct {
	SessionID      string          `json:"session_id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	RemoteJID      string          `json:"remote_jid" example:"5511999999999@s.whatsapp.net"`
	ContactCount   int             `json:"contact_count" example:"3"`
	ContactResults []ContactResult `json:"contact_results"`
	SentAt         time.Time       `json:"sent_at" example:"2024-01-01T12:00:00Z"`
} // @name SendContactListResponse

// PollVoteInfo DTO para informações de voto em poll
type PollVoteInfo struct {
	OptionName string   `json:"option_name" example:"Option 1"`
	Voters     []string `json:"voters" example:"[\"5511888888888@s.whatsapp.net\"]"`
	VoteCount  int      `json:"vote_count" example:"5"`
} // @name PollVoteInfo

// GetPollResultsResponse DTO de resposta para resultados de poll
type GetPollResultsResponse struct {
	MessageID   string         `json:"message_id" example:"3EB0C767D71D"`
	PollName    string         `json:"poll_name" example:"What's your favorite color?"`
	TotalVotes  int            `json:"total_votes" example:"10"`
	VoteResults []PollVoteInfo `json:"vote_results"`
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-01T12:00:00Z"`
} // @name GetPollResultsResponse

// ===== DATA DTOs =====

// MessageDTO representa uma mensagem na camada de aplicação
type MessageDTO struct {
	// Identificadores únicos
	ID        string `json:"id"`
	SessionID string `json:"session_id"`

	// WhatsApp Message Identifiers
	ZpMessageID string    `json:"zp_message_id"`
	ZpSender    string    `json:"zp_sender"`
	ZpChat      string    `json:"zp_chat"`
	ZpTimestamp time.Time `json:"zp_timestamp"`
	ZpFromMe    bool      `json:"zp_from_me"`
	ZpType      string    `json:"zp_type"`
	Content     string    `json:"content,omitempty"`

	// Chatwoot Message Identifiers
	CwMessageID      *int `json:"cw_message_id,omitempty"`
	CwConversationID *int `json:"cw_conversation_id,omitempty"`

	// Sync Status
	SyncStatus string     `json:"sync_status"`
	SyncedAt   *time.Time `json:"synced_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageStatsDTO representa estatísticas de mensagens
type MessageStatsDTO struct {
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

// MessageSyncDTO representa dados de sincronização de mensagem
type MessageSyncDTO struct {
	MessageID        string    `json:"message_id"`
	ZpMessageID      string    `json:"zp_message_id"`
	CwMessageID      *int      `json:"cw_message_id,omitempty"`
	CwConversationID *int      `json:"cw_conversation_id,omitempty"`
	SyncStatus       string    `json:"sync_status"`
	SyncedAt         *time.Time `json:"synced_at,omitempty"`
	LastError        string    `json:"last_error,omitempty"`
}

// MessageSearchDTO representa critérios de busca de mensagens
type MessageSearchDTO struct {
	SessionID string `json:"session_id,omitempty"`
	ChatJID   string `json:"chat_jid,omitempty"`
	FromDate  string `json:"from_date,omitempty"`
	ToDate    string `json:"to_date,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	Query     string `json:"query,omitempty"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// MessageBulkOperationDTO representa operação em lote em mensagens
type MessageBulkOperationDTO struct {
	MessageIDs []string `json:"message_ids"`
	Operation  string   `json:"operation"` // sync, delete, retry, etc.
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// MessageBulkResultDTO representa resultado de operação em lote
type MessageBulkResultDTO struct {
	TotalProcessed int      `json:"total_processed"`
	Successful     int      `json:"successful"`
	Failed         int      `json:"failed"`
	Errors         []string `json:"errors,omitempty"`
}

// IsSynced verifica se a mensagem está sincronizada
func (m *MessageDTO) IsSynced() bool {
	return m.SyncStatus == "synced" && m.CwMessageID != nil
}

// IsPending verifica se a mensagem está pendente de sincronização
func (m *MessageDTO) IsPending() bool {
	return m.SyncStatus == "pending"
}

// IsFailed verifica se a sincronização falhou
func (m *MessageDTO) IsFailed() bool {
	return m.SyncStatus == "failed"
}

// HasChatwootData verifica se a mensagem tem dados do Chatwoot
func (m *MessageDTO) HasChatwootData() bool {
	return m.CwMessageID != nil && m.CwConversationID != nil
}

// GetTypeDisplayName retorna nome amigável do tipo de mensagem
func (m *MessageDTO) GetTypeDisplayName() string {
	switch m.ZpType {
	case "text":
		return "Texto"
	case "image":
		return "Imagem"
	case "audio":
		return "Áudio"
	case "video":
		return "Vídeo"
	case "document":
		return "Documento"
	case "contact":
		return "Contato"
	case "location":
		return "Localização"
	case "sticker":
		return "Sticker"
	default:
		return "Desconhecido"
	}
}

// GetStatusDisplayName retorna nome amigável do status de sincronização
func (m *MessageDTO) GetStatusDisplayName() string {
	switch m.SyncStatus {
	case "pending":
		return "Pendente"
	case "synced":
		return "Sincronizado"
	case "failed":
		return "Falhou"
	default:
		return "Desconhecido"
	}
}

// GetDirectionDisplayName retorna direção da mensagem
func (m *MessageDTO) GetDirectionDisplayName() string {
	if m.ZpFromMe {
		return "Enviada"
	}
	return "Recebida"
}

// GetContentPreview retorna preview do conteúdo da mensagem
func (m *MessageDTO) GetContentPreview(maxLength int) string {
	if m.Content == "" {
		return fmt.Sprintf("[%s]", m.GetTypeDisplayName())
	}

	if len(m.Content) <= maxLength {
		return m.Content
	}

	return m.Content[:maxLength] + "..."
}
