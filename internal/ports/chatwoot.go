package ports

import (
	"context"
	"io"
	"time"
)

type ChatwootClient interface {
	CreateInbox(name, webhookURL string) (*ChatwootInbox, error)
	ListInboxes() ([]ChatwootInbox, error)
	GetInbox(inboxID int) (*ChatwootInbox, error)
	UpdateInbox(inboxID int, updates map[string]interface{}) error
	DeleteInbox(inboxID int) error

	CreateContact(phone, name string, inboxID int) (*ChatwootContact, error)
	FindContact(phone string, inboxID int) (*ChatwootContact, error)
	GetContact(contactID int) (*ChatwootContact, error)
	UpdateContactAttributes(contactID int, attributes map[string]interface{}) error
	MergeContacts(baseContactID, mergeContactID int) error

	CreateConversation(contactID, inboxID int) (*ChatwootConversation, error)
	GetConversation(contactID, inboxID int) (*ChatwootConversation, error)
	GetConversationByID(conversationID int) (*ChatwootConversation, error)
	GetConversationSenderPhone(conversationID int) (string, error)
	ListContactConversations(contactID int) ([]ChatwootConversation, error)
	UpdateConversationStatus(conversationID int, status string) error

	SendMessage(conversationID int, content string) (*ChatwootMessage, error)
	SendMessageWithType(conversationID int, content string, messageType string) (*ChatwootMessage, error)
	SendMediaMessage(conversationID int, content string, attachment io.Reader, filename string) (*ChatwootMessage, error)
	GetMessages(conversationID int, before int) ([]ChatwootMessage, error)

	GetAccount() (*ChatwootAccount, error)
	UpdateAccount(updates map[string]interface{}) error
}

type ChatwootManager interface {
	GetClient(sessionID string) (ChatwootClient, error)
	IsEnabled(sessionID string) bool

	InitInstanceChatwoot(sessionID, inboxName, webhookURL string, autoCreate bool) error

	SetConfig(sessionID string, config *ChatwootConfig) error
	GetConfig(sessionID string) (*ChatwootConfig, error)

	Cleanup(sessionID string) error
}

type WebhookHandler interface {
	ProcessWebhook(ctx context.Context, webhook *ChatwootWebhookPayload, sessionID string) error
}

type ChatwootIntegration interface {
	CreateContact(phoneNumber, name string) (*ChatwootContact, error)
	CreateConversation(contactID string, sessionID string) (*ChatwootConversation, error)
	SendMessage(conversationID, content, messageType string) error
	GetContact(phoneNumber string) (*ChatwootContact, error)
	GetConversation(conversationID string) (*ChatwootConversation, error)
	UpdateContactAttributes(contactID string, attributes map[string]interface{}) error
}

type ChatwootIntegrationExtended interface {
	ChatwootIntegration

	CreateInbox(name, channelType string) (*ChatwootInbox, error)
	GetInbox(inboxID int) (*ChatwootInbox, error)
	UpdateInbox(inboxID int, updates map[string]interface{}) error
	DeleteInbox(inboxID int) error

	GetAccount() (*ChatwootAccount, error)
	UpdateAccount(updates map[string]interface{}) error

	GetAgents() ([]*ChatwootAgent, error)
	GetAgent(agentID int) (*ChatwootAgent, error)
	AssignConversation(conversationID, agentID int) error
	UnassignConversation(conversationID int) error

	CreateLabel(name, description, color string) (*ChatwootLabel, error)
	GetLabels() ([]*ChatwootLabel, error)
	AddLabelToConversation(conversationID int, labelID int) error
	RemoveLabelFromConversation(conversationID int, labelID int) error

	CreateCustomAttribute(name, attributeType, description string) (*ChatwootCustomAttribute, error)
	GetCustomAttributes() ([]*ChatwootCustomAttribute, error)
	UpdateContactCustomAttribute(contactID int, attributeKey string, value interface{}) error

	SetConfig(url string, events []string) (*ChatwootWebhook, error)
	GetWebhooks() ([]*ChatwootWebhook, error)
	UpdateWebhook(webhookID int, updates map[string]interface{}) error
	DeleteWebhook(webhookID int) error

	GetConversationMetrics(from, to int64) (*ConversationMetrics, error)
	GetAgentMetrics(agentID int, from, to int64) (*AgentMetrics, error)
	GetAccountMetrics(from, to int64) (*AccountMetrics, error)

	BulkCreateContacts(contacts []*ChatwootContact) ([]*ChatwootContact, error)
	BulkUpdateContacts(updates []ContactUpdate) error
	BulkDeleteContacts(contactIDs []int) error
}

type ChatwootInbox struct {
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
	Name                 string                 `json:"name"`
	ChannelType          string                 `json:"channel_type"`
	WebhookURL           string                 `json:"webhook_url,omitempty"`
	UpdatedAt            string                 `json:"updated_at"`
	GreetingMessage      string                 `json:"greeting_message"`
	OutOfOfficeMessage   string                 `json:"out_of_office_message"`
	Timezone             string                 `json:"timezone"`
	CreatedAt            string                 `json:"created_at"`
	ID                   int                    `json:"id"`
	WorkingHoursEnabled  bool                   `json:"working_hours_enabled"`
	EnableAutoAssignment bool                   `json:"enable_auto_assignment"`
	GreetingEnabled      bool                   `json:"greeting_enabled"`
}

type ChatwootContactInbox struct {
	SourceID string `json:"source_id"`
	InboxID  int    `json:"inbox_id"`
}

type ChatwootSender struct {
	Name          string `json:"name"`
	AvailableName string `json:"available_name"`
	AvatarURL     string `json:"avatar_url"`
	Type          string `json:"type"`
	Identifier    string `json:"identifier,omitempty"`
	Email         string `json:"email,omitempty"`
	ID            int    `json:"id"`
}

type ChatwootAccount struct {
	Name         string `json:"name"`
	Locale       string `json:"locale"`
	Domain       string `json:"domain,omitempty"`
	SupportEmail string `json:"support_email,omitempty"`
	ID           int    `json:"id"`
}

type ChatwootAgent struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url,omitempty"`
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Confirmed bool   `json:"confirmed"`
	Available bool   `json:"available"`
}

type ChatwootLabel struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	ID            int    `json:"id"`
	ShowOnSidebar bool   `json:"show_on_sidebar"`
}

type ChatwootCustomAttribute struct {
	AttributeKey   string `json:"attribute_key"`
	AttributeType  string `json:"attribute_type"`
	Description    string `json:"description"`
	DefaultValue   string `json:"default_value,omitempty"`
	AttributeModel string `json:"attribute_model"`
	ID             int    `json:"id"`
}

type ChatwootWebhook struct {
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	ID        int      `json:"id"`
	AccountID int      `json:"account_id"`
}

type ConversationMetrics struct {
	TotalConversations    int     `json:"total_conversations"`
	OpenConversations     int     `json:"open_conversations"`
	ResolvedConversations int     `json:"resolved_conversations"`
	AverageResolutionTime float64 `json:"average_resolution_time"`
	AverageResponseTime   float64 `json:"average_response_time"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

type AgentMetrics struct {
	AgentID               int     `json:"agent_id"`
	ConversationsHandled  int     `json:"conversations_handled"`
	ConversationsResolved int     `json:"conversations_resolved"`
	AverageResponseTime   float64 `json:"average_response_time"`
	MessagesSent          int     `json:"messages_sent"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

type AccountMetrics struct {
	TotalContacts         int     `json:"total_contacts"`
	TotalConversations    int     `json:"total_conversations"`
	TotalMessages         int     `json:"total_messages"`
	ActiveAgents          int     `json:"active_agents"`
	AverageResolutionTime float64 `json:"average_resolution_time"`
	CustomerSatisfaction  float64 `json:"customer_satisfaction"`
	From                  int64   `json:"from"`
	To                    int64   `json:"to"`
}

type ContactUpdate struct {
	Updates map[string]interface{} `json:"updates"`
	ID      int                    `json:"id"`
}

type ZpMessage struct {
	UpdatedAt        time.Time  `json:"updated_at"`
	CreatedAt        time.Time  `json:"created_at"`
	ZpTimestamp      time.Time  `json:"zp_timestamp"`
	CwMessageID      *int       `json:"cw_message_id,omitempty"`
	SyncedAt         *time.Time `json:"synced_at,omitempty"`
	CwConversationID *int       `json:"cw_conversation_id,omitempty"`
	ZpSender         string     `json:"zp_sender"`
	ZpType           string     `json:"zp_type"`
	Content          string     `json:"content"`
	ZpChat           string     `json:"zp_chat"`
	SyncStatus       string     `json:"sync_status"`
	ID               string     `json:"id"`
	ZpMessageID      string     `json:"zp_message_id"`
	SessionID        string     `json:"session_id"`
	ZpFromMe         bool       `json:"zp_from_me"`
}

type ChatwootMessageRepository interface {
	CreateMessage(ctx context.Context, message *ZpMessage) error
	GetMessageByZpID(ctx context.Context, sessionID, zpMessageID string) (*ZpMessage, error)
	GetMessageByCwID(ctx context.Context, cwMessageID int) (*ZpMessage, error)
	UpdateSyncStatus(ctx context.Context, id string, status string, cwMessageID, cwConversationID *int) error
	GetMessagesBySession(ctx context.Context, sessionID string, limit, offset int) ([]*ZpMessage, error)
	GetMessagesByChat(ctx context.Context, sessionID, chatJID string, limit, offset int) ([]*ZpMessage, error)
	GetPendingSyncMessages(ctx context.Context, sessionID string, limit int) ([]*ZpMessage, error)
	DeleteMessage(ctx context.Context, id string) error
}

type ChatwootMessageMapper interface {
	CreateMapping(ctx context.Context, sessionID, zpMessageID, zpSender, zpChat, zpType, content string, zpTimestamp time.Time, zpFromMe bool) (*ZpMessage, error)
	UpdateMapping(ctx context.Context, sessionID, zpMessageID string, cwMessageID, cwConversationID int) error
	GetMappingByZpID(ctx context.Context, sessionID, zpMessageID string) (*ZpMessage, error)
	GetMappingByCwID(ctx context.Context, cwMessageID int) (*ZpMessage, error)
	IsMessageMapped(ctx context.Context, sessionID, zpMessageID string) bool
	MarkAsFailed(ctx context.Context, sessionID, zpMessageID string) error
}

type ChatwootWebhookPayload struct {
	Data    map[string]interface{} `json:"data"`
	Account ChatwootAccount        `json:"account"`
	Event   string                 `json:"event"`
}
