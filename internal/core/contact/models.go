package contact

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)



type Contact struct {

	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`


	ZpJID       string `json:"zp_jid"`
	ZpName      string `json:"zp_name"`
	ZpPushName  string `json:"zp_push_name"`
	ZpShortName string `json:"zp_short_name"`
	ZpAvatar    string `json:"zp_avatar"`


	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email,omitempty"`
	IsGroup     bool   `json:"is_group"`
	IsBlocked   bool   `json:"is_blocked"`
	IsBusiness  bool   `json:"is_business"`


	CwContactID      *int `json:"cw_contact_id,omitempty"`
	CwConversationID *int `json:"cw_conversation_id,omitempty"`


	SyncStatus string     `json:"sync_status"`
	SyncedAt   *time.Time `json:"synced_at,omitempty"`


	LastSeen   *time.Time `json:"last_seen,omitempty"`
	IsOnline   bool       `json:"is_online"`
	LastStatus string     `json:"last_status,omitempty"`


	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


type ContactType string

const (
	ContactTypeIndividual ContactType = "individual"
	ContactTypeGroup      ContactType = "group"
	ContactTypeBusiness   ContactType = "business"
)


type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusFailed  SyncStatus = "failed"
)


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


type UpdateSyncStatusRequest struct {
	ID               uuid.UUID  `json:"id" validate:"required"`
	SyncStatus       SyncStatus `json:"sync_status" validate:"required"`
	CwContactID      *int       `json:"cw_contact_id,omitempty"`
	CwConversationID *int       `json:"cw_conversation_id,omitempty"`
	SyncedAt         *time.Time `json:"synced_at,omitempty"`
}


type ListContactsRequest struct {
	SessionID string `json:"session_id,omitempty" validate:"omitempty,uuid"`
	IsGroup   *bool  `json:"is_group,omitempty"`
	IsBlocked *bool  `json:"is_blocked,omitempty"`
	Search    string `json:"search,omitempty"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}


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


func IsValidSyncStatus(status string) bool {
	switch SyncStatus(status) {
	case SyncStatusPending, SyncStatusSynced, SyncStatusFailed:
		return true
	default:
		return false
	}
}


func (c *Contact) IsSynced() bool {
	return c.SyncStatus == string(SyncStatusSynced) && c.CwContactID != nil
}


func (c *Contact) IsPending() bool {
	return c.SyncStatus == string(SyncStatusPending)
}


func (c *Contact) IsFailed() bool {
	return c.SyncStatus == string(SyncStatusFailed)
}


func (c *Contact) HasChatwootData() bool {
	return c.CwContactID != nil
}


func (c *Contact) GetContactType() ContactType {
	if c.IsGroup {
		return ContactTypeGroup
	}
	if c.IsBusiness {
		return ContactTypeBusiness
	}
	return ContactTypeIndividual
}


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


func (c *Contact) GetCleanPhoneNumber() string {

	phone := ""
	for _, char := range c.PhoneNumber {
		if char >= '0' && char <= '9' {
			phone += string(char)
		}
	}
	return phone
}


func (c *Contact) IsOnlineNow() bool {
	return c.IsOnline
}


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
