package messaging

import (
	"context"

	"github.com/google/uuid"
)

// Repository interface para persistência de mensagens
type Repository interface {
	// CRUD básico
	Create(ctx context.Context, message *Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*Message, error)
	Update(ctx context.Context, message *Message) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Consultas por identificadores WhatsApp
	GetByZpMessageID(ctx context.Context, sessionID uuid.UUID, zpMessageID string) (*Message, error)
	ExistsByZpMessageID(ctx context.Context, sessionID uuid.UUID, zpMessageID string) (bool, error)

	// Consultas por identificadores Chatwoot
	GetByCwMessageID(ctx context.Context, cwMessageID int) (*Message, error)
	GetByCwConversationID(ctx context.Context, cwConversationID int, limit, offset int) ([]*Message, error)

	// Listagem e filtros
	List(ctx context.Context, limit, offset int) ([]*Message, error)
	ListBySession(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*Message, error)
	ListByChat(ctx context.Context, sessionID uuid.UUID, chatJID string, limit, offset int) ([]*Message, error)
	ListBySyncStatus(ctx context.Context, status SyncStatus, limit, offset int) ([]*Message, error)

	// Operações de sincronização
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, status SyncStatus, cwMessageID, cwConversationID *int) error
	GetPendingSyncMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*Message, error)
	GetFailedSyncMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*Message, error)
	MarkAsSynced(ctx context.Context, id uuid.UUID, cwMessageID, cwConversationID int) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, errorReason string) error

	// Estatísticas e contadores
	Count(ctx context.Context) (int64, error)
	CountBySession(ctx context.Context, sessionID uuid.UUID) (int64, error)
	CountByChat(ctx context.Context, sessionID uuid.UUID, chatJID string) (int64, error)
	CountBySyncStatus(ctx context.Context, status SyncStatus) (int64, error)
	CountByType(ctx context.Context, messageType MessageType) (int64, error)

	// Estatísticas avançadas
	GetStats(ctx context.Context) (*MessageStats, error)
	GetStatsBySession(ctx context.Context, sessionID uuid.UUID) (*MessageStats, error)
	GetStatsForPeriod(ctx context.Context, sessionID uuid.UUID, from, to int64) (*MessageStats, error)

	// Limpeza e manutenção
	DeleteOldMessages(ctx context.Context, olderThanDays int) (int64, error)
	DeleteBySession(ctx context.Context, sessionID uuid.UUID) (int64, error)
	CleanupFailedMessages(ctx context.Context, olderThanHours int) (int64, error)
}

// MessageService interface para lógica de negócio de mensagens
type MessageService interface {
	// Operações principais
	CreateMessage(ctx context.Context, req *CreateMessageRequest) (*Message, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*Message, error)
	UpdateMessage(ctx context.Context, message *Message) error
	DeleteMessage(ctx context.Context, id uuid.UUID) error

	// Sincronização com Chatwoot
	SyncMessage(ctx context.Context, id uuid.UUID, cwMessageID, cwConversationID int) error
	SyncPendingMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*Message, error)
	RetryFailedMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*Message, error)

	// Consultas e listagem
	ListMessages(ctx context.Context, req *ListMessagesRequest) ([]*Message, int64, error)
	GetMessagesByChat(ctx context.Context, sessionID uuid.UUID, chatJID string, limit, offset int) ([]*Message, error)
	SearchMessages(ctx context.Context, sessionID uuid.UUID, query string, limit, offset int) ([]*Message, error)

	// Estatísticas
	GetMessageStats(ctx context.Context, sessionID *uuid.UUID) (*MessageStats, error)
	GetDashboardStats(ctx context.Context) (*MessageStats, error)

	// Manutenção
	CleanupOldMessages(ctx context.Context, sessionID uuid.UUID, olderThanDays int) (int64, error)
	ReprocessFailedMessages(ctx context.Context, sessionID uuid.UUID) (int64, error)
}

// MessageGateway interface para integração com sistemas externos
type MessageGateway interface {
	// Envio de mensagens
	SendTextMessage(ctx context.Context, sessionID uuid.UUID, to, content string) (*Message, error)
	SendMediaMessage(ctx context.Context, sessionID uuid.UUID, to, mediaURL, caption string, mediaType MessageType) (*Message, error)
	
	// Recebimento de mensagens (webhook/callback)
	ProcessIncomingMessage(ctx context.Context, sessionID uuid.UUID, whatsappMessage interface{}) (*Message, error)
	
	// Status de entrega
	UpdateMessageStatus(ctx context.Context, messageID string, status string) error
	
	// Integração com Chatwoot
	SyncToChatwoot(ctx context.Context, message *Message) error
	SyncFromChatwoot(ctx context.Context, cwMessageID int) (*Message, error)
}

// EventPublisher interface para publicação de eventos de mensagem
type EventPublisher interface {
	// Eventos de mensagem
	PublishMessageCreated(ctx context.Context, message *Message) error
	PublishMessageSynced(ctx context.Context, message *Message) error
	PublishMessageFailed(ctx context.Context, message *Message, reason string) error
	PublishMessageDeleted(ctx context.Context, messageID uuid.UUID) error
	
	// Eventos de sincronização
	PublishSyncStarted(ctx context.Context, sessionID uuid.UUID, messageCount int) error
	PublishSyncCompleted(ctx context.Context, sessionID uuid.UUID, syncedCount, failedCount int) error
	PublishSyncFailed(ctx context.Context, sessionID uuid.UUID, reason string) error
}

// MessageValidator interface para validação de mensagens
type MessageValidator interface {
	ValidateCreateRequest(req *CreateMessageRequest) error
	ValidateMessage(message *Message) error
	ValidateMessageType(messageType string) error
	ValidateSyncStatus(status string) error
	ValidateContent(content string, messageType MessageType) error
}
