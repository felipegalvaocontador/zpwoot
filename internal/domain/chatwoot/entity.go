package chatwoot

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChatwootConfig struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SessionID uuid.UUID `json:"sessionId" db:"sessionId"`
	URL       string    `json:"url" db:"url"`
	Token     string    `json:"token" db:"token"`
	AccountID string    `json:"accountId" db:"accountId"`
	InboxID   *string   `json:"inboxId,omitempty" db:"inboxId"`
	Enabled   bool      `json:"enabled" db:"enabled"`

	// Advanced configuration - shorter names matching DB
	InboxName      *string  `json:"inboxName,omitempty" db:"inboxName"`
	AutoCreate     bool     `json:"autoCreate" db:"autoCreate"`
	SignMsg        bool     `json:"signMsg" db:"signMsg"`
	SignDelimiter  string   `json:"signDelimiter" db:"signDelimiter"`
	ReopenConv     bool     `json:"reopenConv" db:"reopenConv"`
	ConvPending    bool     `json:"convPending" db:"convPending"`
	ImportContacts bool     `json:"importContacts" db:"importContacts"`
	ImportMessages bool     `json:"importMessages" db:"importMessages"`
	ImportDays     int      `json:"importDays" db:"importDays"`
	MergeBrazil    bool     `json:"mergeBrazil" db:"mergeBrazil"`
	Organization   *string  `json:"organization,omitempty" db:"organization"`
	Logo           *string  `json:"logo,omitempty" db:"logo"`
	Number         *string  `json:"number,omitempty" db:"number"`
	IgnoreJids     []string `json:"ignoreJids,omitempty" db:"ignoreJids"`

	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedAt"`
}

var (
	ErrConfigNotFound       = errors.New("chatwoot config not found")
	ErrContactNotFound      = errors.New("chatwoot contact not found")
	ErrConversationNotFound = errors.New("chatwoot conversation not found")
	ErrMessageNotFound      = errors.New("chatwoot message not found")
	ErrInvalidAPIKey        = errors.New("invalid chatwoot API key")
	ErrInvalidAccountID     = errors.New("invalid chatwoot account ID")
	ErrChatwootAPIError     = errors.New("chatwoot API error")
)

// Domain DTOs - used by domain service
type CreateChatwootConfigRequest struct {
	SessionID uuid.UUID `json:"sessionId" validate:"required"`
	URL       string    `json:"url" validate:"required,url"`
	Token     string    `json:"token" validate:"required"`
	AccountID string    `json:"accountId" validate:"required"`
	InboxID   *string   `json:"inboxId,omitempty"`
	Enabled   *bool     `json:"enabled,omitempty"`

	// Advanced configuration
	InboxName      *string  `json:"inboxName,omitempty"`
	AutoCreate     *bool    `json:"autoCreate,omitempty"`
	SignMsg        *bool    `json:"signMsg,omitempty"`
	SignDelimiter  *string  `json:"signDelimiter,omitempty"`
	ReopenConv     *bool    `json:"reopenConv,omitempty"`
	ConvPending    *bool    `json:"convPending,omitempty"`
	ImportContacts *bool    `json:"importContacts,omitempty"`
	ImportMessages *bool    `json:"importMessages,omitempty"`
	ImportDays     *int     `json:"importDays,omitempty"`
	MergeBrazil    *bool    `json:"mergeBrazil,omitempty"`
	Organization   *string  `json:"organization,omitempty"`
	Logo           *string  `json:"logo,omitempty"`
	Number         *string  `json:"number,omitempty"`
	IgnoreJids     []string `json:"ignoreJids,omitempty"`
}

type GetChatwootConfigBySessionRequest struct {
	SessionID string `json:"sessionId" validate:"required"`
}

type UpdateChatwootConfigRequest struct {
	URL       *string `json:"url,omitempty" validate:"omitempty,url"`
	Token     *string `json:"token,omitempty"`
	AccountID *string `json:"accountId,omitempty"`
	InboxID   *string `json:"inboxId,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`

	// Advanced configuration updates
	InboxName      *string  `json:"inboxName,omitempty"`
	AutoCreate     *bool    `json:"autoCreate,omitempty"`
	SignMsg        *bool    `json:"signMsg,omitempty"`
	SignDelimiter  *string  `json:"signDelimiter,omitempty"`
	ReopenConv     *bool    `json:"reopenConv,omitempty"`
	ConvPending    *bool    `json:"convPending,omitempty"`
	ImportContacts *bool    `json:"importContacts,omitempty"`
	ImportMessages *bool    `json:"importMessages,omitempty"`
	ImportDays     *int     `json:"importDays,omitempty"`
	MergeBrazil    *bool    `json:"mergeBrazil,omitempty"`
	Organization   *string  `json:"organization,omitempty"`
	Logo           *string  `json:"logo,omitempty"`
	Number         *string  `json:"number,omitempty"`
	IgnoreJids     []string `json:"ignoreJids,omitempty"`
}

type ChatwootContact struct {
	ID                   int                    `json:"id"`
	Name                 string                 `json:"name"`
	PhoneNumber          string                 `json:"phone_number"`
	Email                string                 `json:"email"`
	Identifier           string                 `json:"identifier,omitempty"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
	CustomAttributes     map[string]interface{} `json:"custom_attributes,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

type ChatwootConversation struct {
	ID        int               `json:"id"`
	ContactID int               `json:"contact_id"`
	InboxID   int               `json:"inbox_id"`
	Status    string            `json:"status"`
	Messages  []ChatwootMessage `json:"messages,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type ChatwootMessage struct {
	ID                int                    `json:"id"`
	ConversationID    int                    `json:"conversation_id"`
	Content           string                 `json:"content"`
	MessageType       string                 `json:"message_type"`
	ContentType       string                 `json:"content_type"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
	Private           bool                   `json:"private"`
	SourceID          string                 `json:"source_id,omitempty"`
	Sender            *ChatwootSender        `json:"sender,omitempty"`
	Attachments       []ChatwootAttachment   `json:"attachments"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreatedAt         time.Time              `json:"created_at"`
}

type ChatwootAttachment struct {
	ID       int    `json:"id"`
	FileType string `json:"file_type"`
	FileURL  string `json:"data_url"`
	FileName string `json:"file_name"`
}

type ChatwootSender struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	AvailableName string `json:"available_name"`
	AvatarURL     string `json:"avatar_url"`
	Type          string `json:"type"`
	Identifier    string `json:"identifier,omitempty"`
	Email         string `json:"email,omitempty"`
}

// ChatwootWebhookPayload representa o payload REAL enviado pelo Chatwoot
// Baseado na análise dos logs do Sidekiq
type ChatwootWebhookPayload struct {
	// Metadados do evento
	Event   string          `json:"event"`
	Account ChatwootAccount `json:"account"`

	// Dados da conversa
	Conversation ChatwootConversation `json:"conversation"`

	// Dados do remetente (estrutura complexa do Chatwoot)
	Sender struct {
		Account              ChatwootAccount        `json:"account"`
		AdditionalAttributes map[string]interface{} `json:"additional_attributes"`
		Avatar               string                 `json:"avatar"`
		CustomAttributes     map[string]interface{} `json:"custom_attributes"`
		Email                *string                `json:"email"`
		ID                   int                    `json:"id"`
		Identifier           *string                `json:"identifier"`
		Name                 string                 `json:"name"`
		PhoneNumber          string                 `json:"phone_number"`
		Thumbnail            string                 `json:"thumbnail"`
		Blocked              bool                   `json:"blocked"`
	} `json:"sender"`

	// Dados da mensagem (campos diretos no payload)
	ID          int         `json:"id"`
	Content     string      `json:"content"`
	ContentType string      `json:"content_type"`
	MessageType string      `json:"message_type"`
	Private     bool        `json:"private"`
	SourceID    *string     `json:"source_id"`
	CreatedAt   interface{} `json:"created_at"`

	// Campos adicionais do Chatwoot
	AdditionalAttributes map[string]interface{} `json:"additional_attributes"`
	ContentAttributes    map[string]interface{} `json:"content_attributes"`
	Inbox                map[string]interface{} `json:"inbox"`

	// Campos legados para compatibilidade
	Contact  ChatwootContact        `json:"contact,omitempty"`
	Message  *ChatwootMessage       `json:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ChatwootAccount struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SyncContactRequest struct {
	PhoneNumber string                 `json:"phone_number" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Email       string                 `json:"email,omitempty" validate:"omitempty,email"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

type SyncConversationRequest struct {
	ContactID int    `json:"contact_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
}

type SendMessageToChatwootRequest struct {
	ConversationID int                    `json:"conversation_id" validate:"required"`
	Content        string                 `json:"content" validate:"required"`
	MessageType    string                 `json:"message_type" validate:"required,oneof=incoming outgoing"`
	ContentType    string                 `json:"content_type,omitempty"`
	Attachments    []ChatwootAttachment   `json:"attachments,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

func NewChatwootConfig(url, apiKey, accountID string, inboxID *string) *ChatwootConfig {
	return &ChatwootConfig{
		ID:        uuid.New(),
		URL:       url,
		Token:     apiKey,
		AccountID: accountID,
		InboxID:   inboxID,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *ChatwootConfig) Update(req *UpdateChatwootConfigRequest) {
	if req.URL != nil {
		c.URL = *req.URL
	}
	if req.Token != nil {
		c.Token = *req.Token
	}
	if req.AccountID != nil {
		c.AccountID = *req.AccountID
	}
	if req.InboxID != nil {
		c.InboxID = req.InboxID
	}
	if req.Enabled != nil {
		c.Enabled = *req.Enabled
	}
	c.UpdatedAt = time.Now()
}

func (c *ChatwootConfig) IsConfigured() bool {
	return c.URL != "" && c.Token != "" && c.AccountID != ""
}

func (c *ChatwootConfig) GetBaseURL() string {
	return c.URL + "/accounts/" + c.AccountID
}

type ChatwootEventType string

const (
	ChatwootEventConversationCreated       ChatwootEventType = "conversation_created"
	ChatwootEventConversationUpdated       ChatwootEventType = "conversation_updated"
	ChatwootEventConversationResolved      ChatwootEventType = "conversation_resolved"
	ChatwootEventMessageCreated            ChatwootEventType = "message_created"
	ChatwootEventMessageUpdated            ChatwootEventType = "message_updated"
	ChatwootEventContactCreated            ChatwootEventType = "contact_created"
	ChatwootEventContactUpdated            ChatwootEventType = "contact_updated"
	ChatwootEventConversationStatusChanged ChatwootEventType = "conversation_status_changed"
	// Additional events that Chatwoot might send
	ChatwootEventConversationOpened     ChatwootEventType = "conversation_opened"
	ChatwootEventConversationReopened   ChatwootEventType = "conversation_reopened"
	ChatwootEventConversationSnoozed    ChatwootEventType = "conversation_snoozed"
	ChatwootEventConversationUnsnoozed  ChatwootEventType = "conversation_unsnoozed"
	ChatwootEventConversationAssigned   ChatwootEventType = "conversation_assigned"
	ChatwootEventConversationUnassigned ChatwootEventType = "conversation_unassigned"
	ChatwootEventMessageDeleted         ChatwootEventType = "message_deleted"
	ChatwootEventContactMerged          ChatwootEventType = "contact_merged"
	ChatwootEventContactDeleted         ChatwootEventType = "contact_deleted"
	// Typing events
	ChatwootEventConversationTypingOn  ChatwootEventType = "conversation_typing_on"
	ChatwootEventConversationTypingOff ChatwootEventType = "conversation_typing_off"
)

func IsValidChatwootEvent(eventType string) bool {
	switch ChatwootEventType(eventType) {
	case ChatwootEventConversationCreated,
		ChatwootEventConversationUpdated,
		ChatwootEventConversationResolved,
		ChatwootEventMessageCreated,
		ChatwootEventMessageUpdated,
		ChatwootEventContactCreated,
		ChatwootEventContactUpdated,
		ChatwootEventConversationStatusChanged,
		ChatwootEventConversationOpened,
		ChatwootEventConversationReopened,
		ChatwootEventConversationSnoozed,
		ChatwootEventConversationUnsnoozed,
		ChatwootEventConversationAssigned,
		ChatwootEventConversationUnassigned,
		ChatwootEventMessageDeleted,
		ChatwootEventContactMerged,
		ChatwootEventContactDeleted,
		ChatwootEventConversationTypingOn,
		ChatwootEventConversationTypingOff:
		return true
	default:
		return false
	}
}
