package ports

import (
	"context"
	"time"

	"zpwoot/internal/domain/contact"
)

type ContactRepository interface {
	SaveContact(ctx context.Context, sessionID string, contact *contact.Contact) error

	GetContact(ctx context.Context, sessionID, jid string) (*contact.Contact, error)

	ListContacts(ctx context.Context, sessionID string, req *ListContactsRequest) (*ListContactsResponse, error)

	UpdateContact(ctx context.Context, sessionID string, contact *contact.Contact) error

	DeleteContact(ctx context.Context, sessionID, jid string) error

	GetContactStats(ctx context.Context, sessionID string) (*ContactStats, error)

	BulkSaveContacts(ctx context.Context, sessionID string, contacts []*contact.Contact) error

	SearchContacts(ctx context.Context, sessionID, query string, limit, offset int) ([]*contact.Contact, error)

	GetContactsByType(ctx context.Context, sessionID, contactType string) ([]*contact.Contact, error)
}

type ContactManager interface {
	IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]interface{}, error)

	GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (map[string]interface{}, error)

	GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]map[string]interface{}, error)

	GetBusinessProfile(ctx context.Context, sessionID, jid string) (map[string]interface{}, error)

	GetAllContacts(ctx context.Context, sessionID string) (map[string]interface{}, error)

	SyncContacts(ctx context.Context, sessionID string, force bool) (*SyncContactsResponse, error)

	BlockContact(ctx context.Context, sessionID, jid string) error

	UnblockContact(ctx context.Context, sessionID, jid string) error

	GetBlockedContacts(ctx context.Context, sessionID string) ([]string, error)
}

type ContactService interface {
	CheckWhatsApp(ctx context.Context, req *CheckWhatsAppRequest) (*CheckWhatsAppResponse, error)

	GetProfilePicture(ctx context.Context, req *GetProfilePictureRequest) (*ProfilePictureInfo, error)

	GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error)

	ListContacts(ctx context.Context, req *ListContactsRequest) (*ListContactsResponse, error)

	SyncContacts(ctx context.Context, req *SyncContactsRequest) (*SyncContactsResponse, error)

	GetBusinessProfile(ctx context.Context, req *GetBusinessProfileRequest) (*GetBusinessProfileResponse, error)

	GetContactStats(ctx context.Context, req *GetContactStatsRequest) (*GetContactStatsResponse, error)

	ValidatePhoneNumber(phoneNumber string) error

	ValidateJID(jid string) error

	NormalizePhoneNumber(phoneNumber string) (string, error)

	FormatJID(phoneNumber string) (string, error)

	ExtractPhoneFromJID(jid string) (string, error)

	IsBusinessContact(contact *contact.Contact) bool

	ProcessContactInfo(ctx context.Context, sessionID string, contact *contact.Contact) (*contact.Contact, error)
}

type ListContactsRequest struct {
	IsBusiness  *bool  `json:"is_business,omitempty"`
	IsBlocked   *bool  `json:"is_blocked,omitempty"`
	SessionID   string `json:"session_id"`
	Query       string `json:"query,omitempty"`
	ContactType string `json:"contact_type,omitempty"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

type ListContactsResponse struct {
	Contacts []*contact.Contact `json:"contacts"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	HasMore  bool               `json:"has_more"`
}

type ContactStats struct {
	TotalContacts    int64 `json:"total_contacts"`
	BusinessContacts int64 `json:"business_contacts"`
	BlockedContacts  int64 `json:"blocked_contacts"`
	RecentContacts   int64 `json:"recent_contacts"`
}

type CheckWhatsAppRequest struct {
	SessionID    string   `json:"session_id"`
	PhoneNumbers []string `json:"phone_numbers"`
}

type CheckWhatsAppResponse struct {
	Results map[string]bool `json:"results"`
}

type GetProfilePictureRequest struct {
	SessionID string `json:"session_id"`
	JID       string `json:"jid"`
	Preview   bool   `json:"preview"`
}

type ProfilePictureInfo struct {
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	DirectURL string    `json:"direct_url,omitempty"`
}

type GetUserInfoRequest struct {
	SessionID string   `json:"session_id"`
	JIDs      []string `json:"jids"`
}

type GetUserInfoResponse struct {
	Users []UserInfo `json:"users"`
}

type UserInfo struct {
	JID        string `json:"jid"`
	Name       string `json:"name"`
	ShortName  string `json:"short_name,omitempty"`
	PushName   string `json:"push_name,omitempty"`
	IsBusiness bool   `json:"is_business"`
	IsContact  bool   `json:"is_contact"`
}

type SyncContactsRequest struct {
	SessionID string `json:"session_id"`
	Force     bool   `json:"force"`
}

type SyncContactsResponse struct {
	Contacts     []*contact.Contact `json:"contacts"`
	SyncedCount  int                `json:"synced_count"`
	NewCount     int                `json:"new_count"`
	UpdatedCount int                `json:"updated_count"`
}

type GetBusinessProfileRequest struct {
	SessionID string `json:"session_id"`
	JID       string `json:"jid"`
}

type GetBusinessProfileResponse struct {
	BusinessProfile *BusinessProfile `json:"business_profile"`
}

type BusinessProfile struct {
	JID         string `json:"jid"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
	Address     string `json:"address,omitempty"`
}

type GetContactStatsRequest struct {
	SessionID string `json:"session_id"`
}

type GetContactStatsResponse struct {
	Stats     *ContactStats `json:"stats"`
	SessionID string        `json:"session_id"`
}
