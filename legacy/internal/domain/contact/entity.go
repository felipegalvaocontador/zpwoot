package contact

import (
	"errors"
	"time"
)

var (
	ErrInvalidSessionID   = errors.New("invalid session ID")
	ErrInvalidJID         = errors.New("invalid JID")
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	ErrInvalidLimit       = errors.New("invalid limit: must be between 1 and 100")
	ErrInvalidOffset      = errors.New("invalid offset: must be >= 0")

	ErrSessionNotFound  = errors.New("session not found")
	ErrContactNotFound  = errors.New("contact not found")
	ErrProfileNotFound  = errors.New("profile not found")
	ErrBusinessNotFound = errors.New("business profile not found")

	ErrSyncFailed       = errors.New("contact sync failed")
	ErrAPIUnavailable   = errors.New("WhatsApp API unavailable")
	ErrRateLimited      = errors.New("rate limited")
	ErrPermissionDenied = errors.New("permission denied")
)

type CheckWhatsAppRequest struct {
	SessionID    string   `json:"session_id"`
	PhoneNumbers []string `json:"phone_numbers"`
}

type WhatsAppStatus struct {
	PhoneNumber  string `json:"phone_number"`
	JID          string `json:"jid,omitempty"`
	VerifiedName string `json:"verified_name,omitempty"`
	IsOnWhatsApp bool   `json:"is_on_whatsapp"`
	IsBusiness   bool   `json:"is_business,omitempty"`
}

type CheckWhatsAppResponse struct {
	Results []WhatsAppStatus `json:"results"`
	Total   int              `json:"total"`
	Checked int              `json:"checked"`
}

type GetProfilePictureRequest struct {
	SessionID string `json:"session_id"`
	JID       string `json:"jid"`
	Preview   bool   `json:"preview,omitempty"`
}

type ProfilePictureInfo struct {
	UpdatedAt  time.Time `json:"updated_at"`
	JID        string    `json:"jid"`
	URL        string    `json:"url,omitempty"`
	ID         string    `json:"id,omitempty"`
	Type       string    `json:"type,omitempty"`
	DirectPath string    `json:"direct_path,omitempty"`
	HasPicture bool      `json:"has_picture"`
}

type GetUserInfoRequest struct {
	SessionID string   `json:"session_id"`
	JIDs      []string `json:"jids"`
}

type UserInfo struct {
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	JID          string     `json:"jid"`
	PhoneNumber  string     `json:"phone_number,omitempty"`
	Name         string     `json:"name,omitempty"`
	Status       string     `json:"status,omitempty"`
	PictureID    string     `json:"picture_id,omitempty"`
	VerifiedName string     `json:"verified_name,omitempty"`
	IsBusiness   bool       `json:"is_business"`
	IsContact    bool       `json:"is_contact"`
	IsOnline     bool       `json:"is_online"`
}

type GetUserInfoResponse struct {
	Users []UserInfo `json:"users"`
	Total int        `json:"total"`
	Found int        `json:"found"`
}

type ListContactsRequest struct {
	SessionID string `json:"session_id"`
	Search    string `json:"search,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

type Contact struct {
	AddedAt     time.Time `json:"added_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	JID         string    `json:"jid"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Name        string    `json:"name,omitempty"`
	ShortName   string    `json:"short_name,omitempty"`
	PushName    string    `json:"push_name,omitempty"`
	IsBusiness  bool      `json:"is_business"`
	IsContact   bool      `json:"is_contact"`
	IsBlocked   bool      `json:"is_blocked"`
}

type ListContactsResponse struct {
	Contacts []Contact `json:"contacts"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
	HasMore  bool      `json:"has_more"`
}

type SyncContactsRequest struct {
	SessionID string `json:"session_id"`
	Force     bool   `json:"force,omitempty"`
}

type SyncContactsResponse struct {
	SyncedAt time.Time `json:"synced_at"`
	Synced   int       `json:"synced"`
	Added    int       `json:"added"`
	Updated  int       `json:"updated"`
	Removed  int       `json:"removed"`
	Total    int       `json:"total"`
}

type GetBusinessProfileRequest struct {
	SessionID string `json:"session_id"`
	JID       string `json:"jid"`
}

type BusinessProfile struct {
	JID         string `json:"jid"`
	Name        string `json:"name,omitempty"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
	Address     string `json:"address,omitempty"`
	Verified    bool   `json:"verified"`
}

type GetBusinessProfileResponse struct {
	UpdatedAt time.Time       `json:"updated_at"`
	Profile   BusinessProfile `json:"profile"`
	Found     bool            `json:"found"`
}

type GetContactStatsRequest struct {
	SessionID string `json:"session_id"`
}

type ContactStats struct {
	LastSyncAt       *time.Time `json:"last_sync_at,omitempty"`
	TotalContacts    int        `json:"total_contacts"`
	WhatsAppContacts int        `json:"whatsapp_contacts"`
	BusinessContacts int        `json:"business_contacts"`
	BlockedContacts  int        `json:"blocked_contacts"`
	SyncRate         float64    `json:"sync_rate"`
}

type GetContactStatsResponse struct {
	Stats     ContactStats `json:"stats"`
	UpdatedAt time.Time    `json:"updated_at"`
	SessionID string       `json:"session_id"`
}
