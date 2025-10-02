package contracts

import (
	"time"
)




type CheckWhatsAppRequest struct {
	PhoneNumbers []string `json:"phone_numbers" validate:"required,min=1,max=50"`
}


type GetProfilePictureRequest struct {
	JID     string `json:"jid" validate:"required"`
	Preview bool   `json:"preview,omitempty"`
}


type GetProfilePictureInfoRequest struct {
	JID     string `json:"jid" validate:"required"`
	Preview bool   `json:"preview,omitempty"`
}


type GetUserInfoRequest struct {
	JIDs []string `json:"jids" validate:"required,min=1,max=20"`
}


type GetDetailedUserInfoRequest struct {
	JIDs []string `json:"jids" validate:"required,min=1,max=20"`
}


type ListContactsRequest struct {
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}


type SyncContactsRequest struct {
	Force bool `json:"force,omitempty"`
}


type GetBusinessProfileRequest struct {
	JID string `json:"jid" validate:"required"`
}


type BlockContactRequest struct {
	JID string `json:"jid" validate:"required"`
}


type UnblockContactRequest struct {
	JID string `json:"jid" validate:"required"`
}




type CheckWhatsAppResponse struct {
	Results []WhatsAppCheckResult `json:"results"`
	Total   int                   `json:"total"`
	Found   int                   `json:"found"`
	Success bool                  `json:"success"`
	Message string                `json:"message"`
}


type WhatsAppCheckResult struct {
	PhoneNumber  string `json:"phone_number"`
	IsOnWhatsApp bool   `json:"is_on_whatsapp"`
	JID          string `json:"jid,omitempty"`
}


type GetProfilePictureResponse struct {
	JID        string `json:"jid"`
	HasPicture bool   `json:"has_picture"`
	URL        string `json:"url,omitempty"`
	Data       []byte `json:"data,omitempty"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}


type GetProfilePictureInfoResponse struct {
	JID         string     `json:"jid"`
	HasPicture  bool       `json:"has_picture"`
	URL         string     `json:"url,omitempty"`
	ID          string     `json:"id,omitempty"`
	Type        string     `json:"type,omitempty"`
	DirectPath  string     `json:"direct_path,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Success     bool       `json:"success"`
	Message     string     `json:"message"`
}


type GetUserInfoResponse struct {
	Users   []UserInfo `json:"users"`
	Total   int        `json:"total"`
	Found   int        `json:"found"`
	Success bool       `json:"success"`
	Message string     `json:"message"`
}


type UserInfo struct {
	JID          string     `json:"jid"`
	PhoneNumber  string     `json:"phone_number"`
	Name         string     `json:"name,omitempty"`
	Status       string     `json:"status,omitempty"`
	PictureID    string     `json:"picture_id,omitempty"`
	IsBusiness   bool       `json:"is_business"`
	VerifiedName string     `json:"verified_name,omitempty"`
	IsContact    bool       `json:"is_contact"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	IsOnline     bool       `json:"is_online"`
}


type ListContactsResponse struct {
	Contacts []ContactDetails `json:"contacts"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
	Success  bool             `json:"success"`
	Message  string           `json:"message"`
}


type ContactDetails struct {
	JID          string `json:"jid"`
	PhoneNumber  string `json:"phone_number"`
	Name         string `json:"name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBusiness   bool   `json:"is_business"`
	IsContact    bool   `json:"is_contact"`
	IsBlocked    bool   `json:"is_blocked"`
}


type SyncContactsResponse struct {
	TotalContacts int    `json:"total_contacts"`
	SyncedCount   int    `json:"synced_count"`
	NewCount      int    `json:"new_count"`
	UpdatedCount  int    `json:"updated_count"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}


type GetBusinessProfileResponse struct {
	JID          string `json:"jid"`
	IsBusiness   bool   `json:"is_business"`
	BusinessName string `json:"business_name,omitempty"`
	Category     string `json:"category,omitempty"`
	Description  string `json:"description,omitempty"`
	Website      string `json:"website,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
	Success      bool   `json:"success"`
	Message      string `json:"message"`
}


type BlockContactResponse struct {
	JID     string `json:"jid"`
	Blocked bool   `json:"blocked"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}


type UnblockContactResponse struct {
	JID     string `json:"jid"`
	Blocked bool   `json:"blocked"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}


type GetAllContactsResponse struct {
	Contacts []ContactDetails `json:"contacts"`
	Total    int              `json:"total"`
	Success  bool             `json:"success"`
	Message  string           `json:"message"`
}




type GetDetailedUserInfoResponse struct {
	Users   []DetailedUserInfo `json:"users"`
	Total   int                `json:"total"`
	Found   int                `json:"found"`
	Success bool               `json:"success"`
	Message string             `json:"message"`
}


type DetailedUserInfo struct {
	JID             string     `json:"jid"`
	PhoneNumber     string     `json:"phone_number"`
	Name            string     `json:"name,omitempty"`
	Status          string     `json:"status,omitempty"`
	StatusTimestamp *time.Time `json:"status_timestamp,omitempty"`
	PictureID       string     `json:"picture_id,omitempty"`
	PictureURL      string     `json:"picture_url,omitempty"`
	IsBusiness      bool       `json:"is_business"`
	BusinessProfile *BusinessProfileInfo `json:"business_profile,omitempty"`
	VerifiedName    string     `json:"verified_name,omitempty"`
	IsContact       bool       `json:"is_contact"`
	LastSeen        *time.Time `json:"last_seen,omitempty"`
	IsOnline        bool       `json:"is_online"`
	IsBlocked       bool       `json:"is_blocked"`
	Privacy         PrivacySettings `json:"privacy"`
}


type BusinessProfileInfo struct {
	BusinessName string `json:"business_name,omitempty"`
	Category     string `json:"category,omitempty"`
	Description  string `json:"description,omitempty"`
	Website      string `json:"website,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
}


type PrivacySettings struct {
	LastSeen       string `json:"last_seen"`
	ProfilePhoto   string `json:"profile_photo"`
	Status         string `json:"status"`
	ReadReceipts   bool   `json:"read_receipts"`
	Groups         string `json:"groups"`
	CallsAdd       string `json:"calls_add"`
}




type ValidateContactRequest struct {
	JID string `json:"jid" validate:"required"`
}


type ValidateContactResponse struct {
	JID       string `json:"jid"`
	IsValid   bool   `json:"is_valid"`
	Exists    bool   `json:"exists"`
	IsContact bool   `json:"is_contact"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}
