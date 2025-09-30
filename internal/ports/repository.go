package ports

import (
	"context"
)

// ChatwootRepository defines the interface for Chatwoot data operations
type ChatwootRepository interface {
	CreateConfig(ctx context.Context, config *ChatwootConfig) error
	GetConfig(ctx context.Context) (*ChatwootConfig, error)
	GetConfigBySessionID(ctx context.Context, sessionID string) (*ChatwootConfig, error)
	UpdateConfig(ctx context.Context, config *ChatwootConfig) error
	DeleteConfig(ctx context.Context) error

	CreateContact(ctx context.Context, contact *ChatwootContact) error
	GetContactByID(ctx context.Context, id int) (*ChatwootContact, error)
	GetContactByPhone(ctx context.Context, phoneNumber string) (*ChatwootContact, error)
	UpdateContact(ctx context.Context, contact *ChatwootContact) error
	DeleteContact(ctx context.Context, id int) error
	ListContacts(ctx context.Context, limit, offset int) ([]*ChatwootContact, int, error)

	CreateConversation(ctx context.Context, conversation *ChatwootConversation) error
	GetConversationByID(ctx context.Context, id int) (*ChatwootConversation, error)
	GetConversationByContactID(ctx context.Context, contactID int) (*ChatwootConversation, error)
	GetConversationBySessionID(ctx context.Context, sessionID string) (*ChatwootConversation, error)
	UpdateConversation(ctx context.Context, conversation *ChatwootConversation) error
	DeleteConversation(ctx context.Context, id int) error
	ListConversations(ctx context.Context, limit, offset int) ([]*ChatwootConversation, int, error)
	GetActiveConversations(ctx context.Context) ([]*ChatwootConversation, error)

	CreateMessage(ctx context.Context, message *ChatwootMessage) error
	GetMessageByID(ctx context.Context, id int) (*ChatwootMessage, error)
	GetMessagesByConversationID(ctx context.Context, conversationID int, limit, offset int) ([]*ChatwootMessage, error)
	UpdateMessage(ctx context.Context, message *ChatwootMessage) error
	DeleteMessage(ctx context.Context, id int) error

	CreateSyncRecord(ctx context.Context, record *SyncRecord) error
	GetSyncRecord(ctx context.Context, sessionID, recordType, externalID string) (*SyncRecord, error)
	UpdateSyncRecord(ctx context.Context, record *SyncRecord) error
	DeleteSyncRecord(ctx context.Context, id string) error
	GetSyncRecordsBySession(ctx context.Context, sessionID string) ([]*SyncRecord, error)

	GetContactCount(ctx context.Context) (int, error)
	GetConversationCount(ctx context.Context) (int, error)
	GetActiveConversationCount(ctx context.Context) (int, error)
	GetMessageCount(ctx context.Context) (int, error)
	GetMessageCountByType(ctx context.Context, messageType string) (int, error)
	GetStatsForPeriod(ctx context.Context, from, to int64) (*ChatwootStats, error)
}

// SyncRecord represents a synchronization record between Wameow and Chatwoot
type SyncRecord struct {
	ID           string `json:"id" db:"id"`
	SessionID    string `json:"session_id" db:"session_id"`
	RecordType   string `json:"record_type" db:"record_type"`
	ExternalID   string `json:"external_id" db:"external_id"`
	PhoneNumber  string `json:"phone_number,omitempty" db:"phone_number"`
	SyncStatus   string `json:"sync_status" db:"sync_status"`
	ErrorMessage string `json:"error_message,omitempty" db:"error_message"`
	ChatwootID   int    `json:"chatwoot_id" db:"chatwoot_id"`
	LastSyncAt   int64  `json:"last_sync_at" db:"last_sync_at"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
	UpdatedAt    int64  `json:"updated_at" db:"updated_at"`
}

// ChatwootStats represents statistics for Chatwoot operations
type ChatwootStats struct {
	TotalContacts       int   `json:"total_contacts"`
	TotalConversations  int   `json:"total_conversations"`
	ActiveConversations int   `json:"active_conversations"`
	MessagesSent        int64 `json:"messages_sent"`
	MessagesReceived    int64 `json:"messages_received"`
	LastSyncAt          int64 `json:"last_sync_at"`
	SyncErrors          int   `json:"sync_errors"`
	From                int64 `json:"from"`
	To                  int64 `json:"to"`
}
