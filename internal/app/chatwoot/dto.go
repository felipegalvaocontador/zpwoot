package chatwoot

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/ports"
)

type CreateChatwootConfigRequest struct {
	// Core configuration - matching zpChatwoot table
	URL       string  `json:"url" validate:"required,url" example:"http://localhost:3001"`
	Token     string  `json:"token" validate:"required" example:"WAF6y4K5s6sdR9uVpsdE7BCt"`
	AccountID string  `json:"accountId" validate:"required" example:"1"`
	InboxID   *string `json:"inboxId,omitempty" example:"1"`
	Enabled   *bool   `json:"enabled,omitempty" example:"true"`

	// Advanced configuration - shorter names matching DB columns
	InboxName      *string  `json:"inboxName,omitempty" example:"WhatsApp zpwoot"`
	AutoCreate     *bool    `json:"autoCreate,omitempty" example:"true"`
	SignMsg        *bool    `json:"signMsg,omitempty" example:"false"`
	SignDelimiter  *string  `json:"signDelimiter,omitempty" example:"\\n\\n"`
	ReopenConv     *bool    `json:"reopenConv,omitempty" example:"true"`
	ConvPending    *bool    `json:"convPending,omitempty" example:"false"`
	ImportContacts *bool    `json:"importContacts,omitempty" example:"false"`
	ImportMessages *bool    `json:"importMessages,omitempty" example:"false"`
	ImportDays     *int     `json:"importDays,omitempty" example:"60"`
	MergeBrazil    *bool    `json:"mergeBrazil,omitempty" example:"true"`
	Organization   *string  `json:"organization,omitempty" example:"zpwoot Bot"`
	Logo           *string  `json:"logo,omitempty" example:"https://zpwoot.com/logo.png"`
	Number         *string  `json:"number,omitempty" example:"5511999999999"`
	IgnoreJids     []string `json:"ignoreJids,omitempty" example:"[\"5511888888888@s.whatsapp.net\"]"`
} // @name CreateChatwootConfigRequest

type CreateChatwootConfigResponse struct {
	ID        string    `json:"id" example:"chatwoot-config-123"`
	URL       string    `json:"url" example:"https://chatwoot.example.com"`
	AccountID string    `json:"accountId" example:"1"`
	InboxID   *string   `json:"inboxId,omitempty" example:"1"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
} // @name CreateChatwootConfigResponse

type UpdateChatwootConfigRequest struct {
	URL       *string `json:"url,omitempty" validate:"omitempty,url" example:"http://localhost:3001"`
	Token     *string `json:"token,omitempty" example:"new-token-123"`
	AccountID *string `json:"accountId,omitempty" example:"2"`
	InboxID   *string `json:"inboxId,omitempty" example:"2"`
	Enabled   *bool   `json:"enabled,omitempty" example:"false"`

	// Advanced configuration updates
	InboxName      *string  `json:"inboxName,omitempty" example:"WhatsApp zpwoot Updated"`
	AutoCreate     *bool    `json:"autoCreate,omitempty" example:"false"`
	SignMsg        *bool    `json:"signMsg,omitempty" example:"true"`
	SignDelimiter  *string  `json:"signDelimiter,omitempty" example:"\\n---\\n"`
	ReopenConv     *bool    `json:"reopenConv,omitempty" example:"false"`
	ConvPending    *bool    `json:"convPending,omitempty" example:"true"`
	ImportContacts *bool    `json:"importContacts,omitempty" example:"true"`
	ImportMessages *bool    `json:"importMessages,omitempty" example:"true"`
	ImportDays     *int     `json:"importDays,omitempty" example:"30"`
	MergeBrazil    *bool    `json:"mergeBrazil,omitempty" example:"false"`
	Organization   *string  `json:"organization,omitempty" example:"Updated Bot"`
	Logo           *string  `json:"logo,omitempty" example:"https://new-logo.com/logo.png"`
	Number         *string  `json:"number,omitempty" example:"5511888888888"`
	IgnoreJids     []string `json:"ignoreJids,omitempty" example:"[\"5511777777777@s.whatsapp.net\"]"`
}

type ChatwootConfigResponse struct {
	ID        string    `json:"id" example:"chatwoot-config-123"`
	URL       string    `json:"url" example:"https://chatwoot.example.com"`
	AccountID string    `json:"accountId" example:"1"`
	InboxID   *string   `json:"inboxId,omitempty" example:"1"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
} // @name ChatwootConfigResponse

type SyncContactRequest struct {
	PhoneNumber string                 `json:"phoneNumber" validate:"required" example:"+5511999999999"`
	Name        string                 `json:"name" validate:"required" example:"John Doe"`
	Email       string                 `json:"email,omitempty" validate:"omitempty,email" example:"john@example.com"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

type SyncContactResponse struct {
	ID          int                    `json:"id" example:"123"`
	PhoneNumber string                 `json:"phoneNumber" example:"+5511999999999"`
	Name        string                 `json:"name" example:"John Doe"`
	Email       string                 `json:"email,omitempty" example:"john@example.com"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt   time.Time              `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time              `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
}

type SyncConversationRequest struct {
	ContactID   int    `json:"contactId" validate:"required" example:"123"`
	SessionID   string `json:"sessionId" validate:"required" example:"session-123"`
	PhoneNumber string `json:"phoneNumber" validate:"required" example:"+5511999999999"`
}

type SyncConversationResponse struct {
	ID          int       `json:"id" example:"456"`
	ContactID   int       `json:"contactId" example:"123"`
	SessionID   string    `json:"sessionId" example:"session-123"`
	PhoneNumber string    `json:"phoneNumber" example:"+5511999999999"`
	Status      string    `json:"status" example:"open"`
	CreatedAt   time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
}

type SendMessageToChatwootRequest struct {
	ConversationID int                    `json:"conversationId" validate:"required" example:"456"`
	Content        string                 `json:"content" validate:"required" example:"Hello from Wameow!"`
	MessageType    string                 `json:"messageType" validate:"required,oneof=incoming outgoing" example:"incoming"`
	ContentType    string                 `json:"contentType,omitempty" example:"text"`
	Attachments    []ChatwootAttachment   `json:"attachments,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type ChatwootAttachment struct {
	URL      string `json:"url" example:"https://example.com/image.jpg"`
	FileType string `json:"fileType" example:"image"`
	FileName string `json:"fileName" example:"image.jpg"`
}

type SendMessageToChatwootResponse struct {
	ID             int                    `json:"id" example:"789"`
	ConversationID int                    `json:"conversationId" example:"456"`
	Content        string                 `json:"content" example:"Hello from Wameow!"`
	MessageType    string                 `json:"messageType" example:"incoming"`
	ContentType    string                 `json:"contentType" example:"text"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"createdAt" example:"2024-01-01T00:00:00Z"`
}

// WebhookRequest represents the complete webhook payload from Chatwoot
type WebhookRequest struct {
	Account           ChatwootAccount        `json:"account"`
	Conversation      ChatwootConversation   `json:"conversation"`
	Message           *ChatwootMessage       `json:"message,omitempty"`
	Contact           ChatwootContact        `json:"contact"`
	Event             string                 `json:"event"`
	Private           bool                   `json:"private"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
	Meta              *Meta                  `json:"meta,omitempty"`
} // @name WebhookRequest

// Meta represents metadata in webhook payload
type Meta struct {
	Sender *Sender `json:"sender,omitempty"`
} // @name Meta

// Sender represents the sender information in webhook
type Sender struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Identifier    string `json:"identifier"`
	AvailableName string `json:"available_name"`
	AvatarURL     string `json:"avatar_url"`
	Type          string `json:"type"` // contact, user, agent_bot
	Email         string `json:"email,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
} // @name Sender

type ChatwootWebhookPayload struct {
	Event        string               `json:"event" example:"message_created"`
	Account      ChatwootAccount      `json:"account"`
	Conversation ChatwootConversation `json:"conversation,omitempty"`

	// Real Chatwoot webhook fields (top-level)
	ID          int                    `json:"id,omitempty"`
	Content     string                 `json:"content,omitempty"`
	ContentType string                 `json:"content_type,omitempty"`
	MessageType string                 `json:"message_type,omitempty"`
	Private     bool                   `json:"private,omitempty"`
	SourceID    *string                `json:"source_id,omitempty"`
	Sender      Sender                 `json:"sender,omitempty"`
	Inbox       map[string]interface{} `json:"inbox,omitempty"`

	// Legacy/nested message (some deployments send under "message")
	Message ChatwootMessage `json:"message,omitempty"`

	// Optional/extra attributes
	ContentAttributes    map[string]interface{} `json:"content_attributes,omitempty"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
	CreatedAt            interface{}            `json:"created_at,omitempty"`
	Data                 map[string]interface{} `json:"data,omitempty"`
}

type ChatwootAccount struct {
	ID   int    `json:"id" example:"1"`
	Name string `json:"name" example:"My Company"`
}

type ChatwootConversation struct {
	ID                   int                    `json:"id" example:"456"`
	Status               string                 `json:"status" example:"open"`
	ContactID            int                    `json:"contact_id" example:"123"`
	InboxID              int                    `json:"inbox_id" example:"1"`
	AgentLastSeenAt      interface{}            `json:"agent_last_seen_at,omitempty"`
	ContactLastSeenAt    interface{}            `json:"contact_last_seen_at,omitempty"`
	Timestamp            int64                  `json:"timestamp" example:"1640995200"`
	UnreadCount          int                    `json:"unread_count" example:"0"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
	CustomAttributes     map[string]interface{} `json:"custom_attributes,omitempty"`
	Labels               []string               `json:"labels,omitempty"`
}

type ChatwootContact struct {
	ID                   int                    `json:"id" example:"123"`
	Name                 string                 `json:"name" example:"John Doe"`
	PhoneNumber          string                 `json:"phone_number" example:"+5511999999999"`
	Email                string                 `json:"email,omitempty" example:"john@example.com"`
	Identifier           string                 `json:"identifier,omitempty"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
	CustomAttributes     map[string]interface{} `json:"custom_attributes,omitempty"`
}

type ChatwootMessage struct {
	ID                int                    `json:"id" example:"789"`
	Content           string                 `json:"content" example:"Hello!"`
	MessageType       string                 `json:"message_type" example:"incoming"`
	ContentType       string                 `json:"content_type" example:"text"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
	CreatedAt         string                 `json:"created_at" example:"2024-01-01T00:00:00Z"`
	Private           bool                   `json:"private" example:"false"`
	SourceID          string                 `json:"source_id,omitempty"`
	Sender            *Sender                `json:"sender,omitempty"`
	ConversationID    int                    `json:"conversation_id" example:"456"`
	Attachments       []ChatwootAttachment   `json:"attachments,omitempty"`
}

type TestChatwootConnectionResponse struct {
	Success     bool   `json:"success" example:"true"`
	AccountName string `json:"accountName,omitempty" example:"My Company"`
	InboxName   string `json:"inboxName,omitempty" example:"Wameow Inbox"`
	Error       string `json:"error,omitempty"`
} // @name TestChatwootConnectionResponse

type ChatwootStatsResponse struct {
	TotalContacts       int `json:"totalContacts" example:"150"`
	TotalConversations  int `json:"totalConversations" example:"89"`
	ActiveConversations int `json:"activeConversations" example:"12"`
	MessagesSent        int `json:"messagesSent" example:"1250"`
	MessagesReceived    int `json:"messagesReceived" example:"890"`
} // @name ChatwootStatsResponse

// Evolution API specific DTOs
type ChatwootConfigEvolutionRequest struct {
	Enabled                 *bool    `json:"enabled,omitempty" example:"true"`
	AccountID               string   `json:"accountId" validate:"required" example:"1"`
	Token                   string   `json:"token" validate:"required" example:"WAF6y4K5s6sdR9uVpsdE7BCt"`
	URL                     string   `json:"url" validate:"required,url" example:"http://localhost:3001"`
	NameInbox               string   `json:"nameInbox,omitempty" example:"WhatsApp Inbox"`
	SignMsg                 *bool    `json:"signMsg,omitempty" example:"false"`
	SignDelimiter           string   `json:"signDelimiter,omitempty" example:"\\n\\n"`
	Number                  string   `json:"number,omitempty" example:"5511999999999"`
	ReopenConversation      *bool    `json:"reopenConversation,omitempty" example:"true"`
	ConversationPending     *bool    `json:"conversationPending,omitempty" example:"false"`
	MergeBrazilContacts     *bool    `json:"mergeBrazilContacts,omitempty" example:"true"`
	ImportContacts          *bool    `json:"importContacts,omitempty" example:"false"`
	ImportMessages          *bool    `json:"importMessages,omitempty" example:"false"`
	DaysLimitImportMessages *int     `json:"daysLimitImportMessages,omitempty" example:"60"`
	AutoCreate              *bool    `json:"autoCreate,omitempty" example:"true"`
	Organization            string   `json:"organization,omitempty" example:"My Company"`
	Logo                    string   `json:"logo,omitempty" example:"https://example.com/logo.png"`
	IgnoreJids              []string `json:"ignoreJids,omitempty" example:"['123456@s.whatsapp.net']"`
} // @name ChatwootConfigEvolutionRequest

type ChatwootConfigEvolutionResponse struct {
	Enabled                 bool     `json:"enabled" example:"true"`
	AccountID               string   `json:"accountId" example:"1"`
	Token                   string   `json:"token" example:"WAF6y4K5s6sdR9uVpsdE7BCt"`
	URL                     string   `json:"url" example:"http://localhost:3001"`
	NameInbox               string   `json:"nameInbox" example:"WhatsApp Inbox"`
	SignMsg                 bool     `json:"signMsg" example:"false"`
	SignDelimiter           string   `json:"signDelimiter" example:"\\n\\n"`
	Number                  string   `json:"number" example:"5511999999999"`
	ReopenConversation      bool     `json:"reopenConversation" example:"true"`
	ConversationPending     bool     `json:"conversationPending" example:"false"`
	MergeBrazilContacts     bool     `json:"mergeBrazilContacts" example:"true"`
	ImportContacts          bool     `json:"importContacts" example:"false"`
	ImportMessages          bool     `json:"importMessages" example:"false"`
	DaysLimitImportMessages int      `json:"daysLimitImportMessages" example:"60"`
	AutoCreate              bool     `json:"autoCreate" example:"true"`
	Organization            string   `json:"organization" example:"My Company"`
	Logo                    string   `json:"logo" example:"https://example.com/logo.png"`
	IgnoreJids              []string `json:"ignoreJids" example:"['123456@s.whatsapp.net']"`
	WebhookURL              string   `json:"webhook_url" example:"http://localhost:8080/chatwoot/webhook/session-id"`
	InboxID                 *string  `json:"inboxId,omitempty" example:"1"`
} // @name ChatwootConfigEvolutionResponse

type InitInstanceChatwootRequest struct {
	InboxName    string `json:"inboxName" validate:"required" example:"WhatsApp Inbox"`
	WebhookURL   string `json:"webhookUrl" validate:"required,url" example:"https://myapp.com/webhook"`
	AutoCreate   bool   `json:"autoCreate" example:"true"`
	Number       string `json:"number,omitempty" example:"5511999999999"`
	Organization string `json:"organization,omitempty" example:"My Company"`
	Logo         string `json:"logo,omitempty" example:"https://example.com/logo.png"`
} // @name InitInstanceChatwootRequest

type ImportHistoryRequest struct {
	DaysLimit int `json:"daysLimit" validate:"min=1,max=365" example:"60"`
} // @name ImportHistoryRequest

func (r *CreateChatwootConfigRequest) ToCreateChatwootConfigRequest(sessionID string) (*chatwoot.CreateChatwootConfigRequest, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	return &chatwoot.CreateChatwootConfigRequest{
		SessionID:      sessionUUID,
		URL:            r.URL,
		Token:          r.Token,
		AccountID:      r.AccountID,
		InboxID:        r.InboxID,
		Enabled:        r.Enabled,
		InboxName:      r.InboxName,
		AutoCreate:     r.AutoCreate,
		SignMsg:        r.SignMsg,
		SignDelimiter:  r.SignDelimiter,
		ReopenConv:     r.ReopenConv,
		ConvPending:    r.ConvPending,
		ImportContacts: r.ImportContacts,
		ImportMessages: r.ImportMessages,
		ImportDays:     r.ImportDays,
		MergeBrazil:    r.MergeBrazil,
		Organization:   r.Organization,
		Logo:           r.Logo,
		Number:         r.Number,
		IgnoreJids:     r.IgnoreJids,
	}, nil
}

func (r *UpdateChatwootConfigRequest) ToUpdateChatwootConfigRequest() *chatwoot.UpdateChatwootConfigRequest {
	return &chatwoot.UpdateChatwootConfigRequest{
		URL:            r.URL,
		Token:          r.Token,
		AccountID:      r.AccountID,
		InboxID:        r.InboxID,
		Enabled:        r.Enabled,
		InboxName:      r.InboxName,
		AutoCreate:     r.AutoCreate,
		SignMsg:        r.SignMsg,
		SignDelimiter:  r.SignDelimiter,
		ReopenConv:     r.ReopenConv,
		ConvPending:    r.ConvPending,
		ImportContacts: r.ImportContacts,
		ImportMessages: r.ImportMessages,
		ImportDays:     r.ImportDays,
		MergeBrazil:    r.MergeBrazil,
		Organization:   r.Organization,
		Logo:           r.Logo,
		Number:         r.Number,
		IgnoreJids:     r.IgnoreJids,
	}
}

func FromChatwootConfig(c *ports.ChatwootConfig) *ChatwootConfigResponse {
	return &ChatwootConfigResponse{
		ID:        c.ID.String(),
		URL:       c.URL,
		AccountID: c.AccountID,
		InboxID:   c.InboxID,
		Active:    c.Enabled,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
