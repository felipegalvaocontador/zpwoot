package dto

import (
	"fmt"
	"time"
)

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
