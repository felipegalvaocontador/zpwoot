package contact

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Contact representa um contato no sistema zpwoot
// Mapeia contatos entre WhatsApp e Chatwoot
type Contact struct {
	// Identificadores únicos
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`

	// WhatsApp Contact Identifiers
	ZpJID       string `json:"zp_jid"`        // WhatsApp JID (ex: 5511999999999@s.whatsapp.net)
	ZpName      string `json:"zp_name"`       // Nome no WhatsApp
	ZpPushName  string `json:"zp_push_name"`  // Push name do WhatsApp
	ZpShortName string `json:"zp_short_name"` // Nome curto
	ZpAvatar    string `json:"zp_avatar"`     // URL do avatar

	// Contact Information
	PhoneNumber string `json:"phone_number"` // Número de telefone limpo
	Email       string `json:"email,omitempty"`
	IsGroup     bool   `json:"is_group"`
	IsBlocked   bool   `json:"is_blocked"`
	IsBusiness  bool   `json:"is_business"`

	// Chatwoot Contact Identifiers
	CwContactID      *int `json:"cw_contact_id,omitempty"`
	CwConversationID *int `json:"cw_conversation_id,omitempty"`

	// Sync Status
	SyncStatus string     `json:"sync_status"` // pending, synced, failed
	SyncedAt   *time.Time `json:"synced_at,omitempty"`

	// Metadata
	LastSeen   *time.Time `json:"last_seen,omitempty"`
	IsOnline   bool       `json:"is_online"`
	LastStatus string     `json:"last_status,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContactType constantes para tipos de contato
type ContactType string

const (
	ContactTypeIndividual ContactType = "individual"
	ContactTypeGroup      ContactType = "group"
	ContactTypeBusiness   ContactType = "business"
)

// SyncStatus constantes para status de sincronização
type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusFailed  SyncStatus = "failed"
)

// CreateContactRequest dados para criação de contato
type CreateContactRequest struct {
	SessionID   uuid.UUID `json:"session_id" validate:"required"`
	ZpJID       string    `json:"zp_jid" validate:"required"`
	ZpName      string    `json:"zp_name"`
	ZpPushName  string    `json:"zp_push_name"`
	ZpShortName string    `json:"zp_short_name"`
	ZpAvatar    string    `json:"zp_avatar"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email,omitempty" validate:"omitempty,email"`
	IsGroup     bool      `json:"is_group"`
	IsBusiness  bool      `json:"is_business"`
}

// UpdateContactRequest dados para atualização de contato
type UpdateContactRequest struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	ZpName      string    `json:"zp_name"`
	ZpPushName  string    `json:"zp_push_name"`
	ZpShortName string    `json:"zp_short_name"`
	ZpAvatar    string    `json:"zp_avatar"`
	Email       string    `json:"email,omitempty" validate:"omitempty,email"`
	IsBlocked   bool      `json:"is_blocked"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
	IsOnline    bool      `json:"is_online"`
	LastStatus  string    `json:"last_status,omitempty"`
}

// UpdateSyncStatusRequest dados para atualização de status de sync
type UpdateSyncStatusRequest struct {
	ID               uuid.UUID  `json:"id" validate:"required"`
	SyncStatus       SyncStatus `json:"sync_status" validate:"required"`
	CwContactID      *int       `json:"cw_contact_id,omitempty"`
	CwConversationID *int       `json:"cw_conversation_id,omitempty"`
	SyncedAt         *time.Time `json:"synced_at,omitempty"`
}

// ListContactsRequest dados para listagem de contatos
type ListContactsRequest struct {
	SessionID string `json:"session_id,omitempty" validate:"omitempty,uuid"`
	IsGroup   *bool  `json:"is_group,omitempty"`
	IsBlocked *bool  `json:"is_blocked,omitempty"`
	Search    string `json:"search,omitempty"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}

// ContactStats estatísticas de contatos
type ContactStats struct {
	TotalContacts      int64            `json:"total_contacts"`
	ContactsByType     map[string]int64 `json:"contacts_by_type"`
	ContactsByStatus   map[string]int64 `json:"contacts_by_status"`
	SyncedContacts     int64            `json:"synced_contacts"`
	PendingContacts    int64            `json:"pending_contacts"`
	FailedContacts     int64            `json:"failed_contacts"`
	BlockedContacts    int64            `json:"blocked_contacts"`
	BusinessContacts   int64            `json:"business_contacts"`
	GroupContacts      int64            `json:"group_contacts"`
	IndividualContacts int64            `json:"individual_contacts"`
	OnlineContacts     int64            `json:"online_contacts"`
	ContactsToday      int64            `json:"contacts_today"`
	ContactsThisWeek   int64            `json:"contacts_this_week"`
	ContactsThisMonth  int64            `json:"contacts_this_month"`
}

// IsValidSyncStatus verifica se o status de sync é válido
func IsValidSyncStatus(status string) bool {
	switch SyncStatus(status) {
	case SyncStatusPending, SyncStatusSynced, SyncStatusFailed:
		return true
	default:
		return false
	}
}

// IsSynced verifica se o contato está sincronizado
func (c *Contact) IsSynced() bool {
	return c.SyncStatus == string(SyncStatusSynced) && c.CwContactID != nil
}

// IsPending verifica se o contato está pendente de sincronização
func (c *Contact) IsPending() bool {
	return c.SyncStatus == string(SyncStatusPending)
}

// IsFailed verifica se a sincronização falhou
func (c *Contact) IsFailed() bool {
	return c.SyncStatus == string(SyncStatusFailed)
}

// HasChatwootData verifica se o contato tem dados do Chatwoot
func (c *Contact) HasChatwootData() bool {
	return c.CwContactID != nil
}

// GetContactType retorna o tipo do contato
func (c *Contact) GetContactType() ContactType {
	if c.IsGroup {
		return ContactTypeGroup
	}
	if c.IsBusiness {
		return ContactTypeBusiness
	}
	return ContactTypeIndividual
}

// GetDisplayName retorna o nome de exibição do contato
func (c *Contact) GetDisplayName() string {
	if c.ZpName != "" {
		return c.ZpName
	}
	if c.ZpPushName != "" {
		return c.ZpPushName
	}
	if c.ZpShortName != "" {
		return c.ZpShortName
	}
	return c.PhoneNumber
}

// GetCleanPhoneNumber retorna o número de telefone limpo (apenas dígitos)
func (c *Contact) GetCleanPhoneNumber() string {
	// Remove caracteres não numéricos
	phone := ""
	for _, char := range c.PhoneNumber {
		if char >= '0' && char <= '9' {
			phone += string(char)
		}
	}
	return phone
}

// IsOnlineNow verifica se o contato está online agora
func (c *Contact) IsOnlineNow() bool {
	return c.IsOnline
}

// GetLastSeenString retorna string formatada do último visto
func (c *Contact) GetLastSeenString() string {
	if c.LastSeen == nil {
		return "Nunca visto"
	}
	
	now := time.Now()
	diff := now.Sub(*c.LastSeen)
	
	if diff < time.Minute {
		return "Agora"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d minutos atrás", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d horas atrás", int(diff.Hours()))
	} else {
		return fmt.Sprintf("%d dias atrás", int(diff.Hours()/24))
	}
}
