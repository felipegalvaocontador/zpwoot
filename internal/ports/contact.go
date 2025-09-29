package ports

import (
	"context"
	"time"

	"zpwoot/internal/domain/contact"
)

// ContactRepository defines the interface for contact data operations
type ContactRepository interface {
	// SaveContact saves a contact to local storage
	SaveContact(ctx context.Context, sessionID string, contact *contact.Contact) error

	// GetContact retrieves a contact by JID
	GetContact(ctx context.Context, sessionID, jid string) (*contact.Contact, error)

	// ListContacts lists contacts with pagination and filtering
	ListContacts(ctx context.Context, sessionID string, req *ListContactsRequest) (*ListContactsResponse, error)

	// UpdateContact updates an existing contact
	UpdateContact(ctx context.Context, sessionID string, contact *contact.Contact) error

	// DeleteContact removes a contact from local storage
	DeleteContact(ctx context.Context, sessionID, jid string) error

	// GetContactStats returns contact statistics
	GetContactStats(ctx context.Context, sessionID string) (*ContactStats, error)

	// BulkSaveContacts saves multiple contacts in a single operation
	BulkSaveContacts(ctx context.Context, sessionID string, contacts []*contact.Contact) error

	// SearchContacts searches contacts by name or phone number
	SearchContacts(ctx context.Context, sessionID, query string, limit, offset int) ([]*contact.Contact, error)

	// GetContactsByType returns contacts filtered by type (business, regular, etc.)
	GetContactsByType(ctx context.Context, sessionID, contactType string) ([]*contact.Contact, error)
}

// ContactManager defines the interface for WhatsApp contact operations
type ContactManager interface {
	// IsOnWhatsApp checks if phone numbers are registered on WhatsApp
	IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]interface{}, error)

	// GetProfilePictureInfo gets profile picture information for a contact
	GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (map[string]interface{}, error)

	// GetUserInfo gets detailed information about WhatsApp users
	GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]map[string]interface{}, error)

	// GetBusinessProfile gets business profile information for a contact
	GetBusinessProfile(ctx context.Context, sessionID, jid string) (map[string]interface{}, error)

	// GetAllContacts gets all contacts from the WhatsApp account
	GetAllContacts(ctx context.Context, sessionID string) (map[string]interface{}, error)

	// SyncContacts synchronizes contacts from the device with WhatsApp
	SyncContacts(ctx context.Context, sessionID string, force bool) (*SyncContactsResponse, error)

	// BlockContact blocks a contact
	BlockContact(ctx context.Context, sessionID, jid string) error

	// UnblockContact unblocks a contact
	UnblockContact(ctx context.Context, sessionID, jid string) error

	// GetBlockedContacts gets list of blocked contacts
	GetBlockedContacts(ctx context.Context, sessionID string) ([]string, error)
}

// ContactService defines the interface for contact business logic
type ContactService interface {
	// CheckWhatsApp validates and checks if phone numbers are on WhatsApp
	CheckWhatsApp(ctx context.Context, req *CheckWhatsAppRequest) (*CheckWhatsAppResponse, error)

	// GetProfilePicture gets and processes profile picture information
	GetProfilePicture(ctx context.Context, req *GetProfilePictureRequest) (*ProfilePictureInfo, error)

	// GetUserInfo gets and validates user information
	GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error)

	// ListContacts lists and filters contacts with business logic
	ListContacts(ctx context.Context, req *ListContactsRequest) (*ListContactsResponse, error)

	// SyncContacts handles contact synchronization with validation
	SyncContacts(ctx context.Context, req *SyncContactsRequest) (*SyncContactsResponse, error)

	// GetBusinessProfile gets and validates business profile information
	GetBusinessProfile(ctx context.Context, req *GetBusinessProfileRequest) (*GetBusinessProfileResponse, error)

	// GetContactStats calculates and returns contact statistics
	GetContactStats(ctx context.Context, req *GetContactStatsRequest) (*GetContactStatsResponse, error)

	// ValidatePhoneNumber validates phone number format
	ValidatePhoneNumber(phoneNumber string) error

	// ValidateJID validates WhatsApp JID format
	ValidateJID(jid string) error

	// NormalizePhoneNumber normalizes phone number to standard format
	NormalizePhoneNumber(phoneNumber string) (string, error)

	// FormatJID formats a phone number to WhatsApp JID
	FormatJID(phoneNumber string) (string, error)

	// ExtractPhoneFromJID extracts phone number from WhatsApp JID
	ExtractPhoneFromJID(jid string) (string, error)

	// IsBusinessContact checks if a contact is a business contact
	IsBusinessContact(contact *contact.Contact) bool

	// ProcessContactInfo processes and enriches contact information
	ProcessContactInfo(ctx context.Context, sessionID string, contact *contact.Contact) (*contact.Contact, error)
}

// ListContactsRequest represents a request to list contacts
type ListContactsRequest struct {
	SessionID   string `json:"session_id"`
	Query       string `json:"query,omitempty"`
	ContactType string `json:"contact_type,omitempty"`
	IsBusiness  *bool  `json:"is_business,omitempty"`
	IsBlocked   *bool  `json:"is_blocked,omitempty"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

// ListContactsResponse represents a response with contact list
type ListContactsResponse struct {
	Contacts []*contact.Contact `json:"contacts"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	HasMore  bool               `json:"has_more"`
}

// ContactStats represents contact statistics
type ContactStats struct {
	TotalContacts    int64 `json:"total_contacts"`
	BusinessContacts int64 `json:"business_contacts"`
	BlockedContacts  int64 `json:"blocked_contacts"`
	RecentContacts   int64 `json:"recent_contacts"`
}

// CheckWhatsAppRequest represents a request to check WhatsApp status
type CheckWhatsAppRequest struct {
	SessionID    string   `json:"session_id"`
	PhoneNumbers []string `json:"phone_numbers"`
}

// CheckWhatsAppResponse represents a response with WhatsApp status
type CheckWhatsAppResponse struct {
	Results map[string]bool `json:"results"`
}

// GetProfilePictureRequest represents a request to get profile picture
type GetProfilePictureRequest struct {
	SessionID string `json:"session_id"`
	JID       string `json:"jid"`
	Preview   bool   `json:"preview"`
}

// ProfilePictureInfo represents profile picture information
type ProfilePictureInfo struct {
	URL       string    `json:"url"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	DirectURL string    `json:"direct_url,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetUserInfoRequest represents a request to get user information
type GetUserInfoRequest struct {
	SessionID string   `json:"session_id"`
	JIDs      []string `json:"jids"`
}

// GetUserInfoResponse represents a response with user information
type GetUserInfoResponse struct {
	Users []UserInfo `json:"users"`
}

// UserInfo represents user information
type UserInfo struct {
	JID        string `json:"jid"`
	Name       string `json:"name"`
	ShortName  string `json:"short_name,omitempty"`
	PushName   string `json:"push_name,omitempty"`
	IsBusiness bool   `json:"is_business"`
	IsContact  bool   `json:"is_contact"`
}

// SyncContactsRequest represents a request to sync contacts
type SyncContactsRequest struct {
	SessionID string `json:"session_id"`
	Force     bool   `json:"force"`
}

// SyncContactsResponse represents a response for contact sync
type SyncContactsResponse struct {
	SyncedCount  int                `json:"synced_count"`
	NewCount     int                `json:"new_count"`
	UpdatedCount int                `json:"updated_count"`
	Contacts     []*contact.Contact `json:"contacts"`
}

// GetBusinessProfileRequest represents a request to get business profile
type GetBusinessProfileRequest struct {
	SessionID string `json:"session_id"`
	JID       string `json:"jid"`
}

// GetBusinessProfileResponse represents a response with business profile
type GetBusinessProfileResponse struct {
	BusinessProfile *BusinessProfile `json:"business_profile"`
}

// BusinessProfile represents business profile information
type BusinessProfile struct {
	JID         string `json:"jid"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
	Address     string `json:"address,omitempty"`
}

// GetContactStatsRequest represents a request to get contact statistics
type GetContactStatsRequest struct {
	SessionID string `json:"session_id"`
}

// GetContactStatsResponse represents a response with contact statistics
type GetContactStatsResponse struct {
	Stats     *ContactStats `json:"stats"`
	SessionID string        `json:"session_id"`
}
