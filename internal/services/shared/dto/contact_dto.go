package dto

import (
	"time"
)

// ===== REQUEST DTOs =====

// CheckWhatsAppRequest representa uma solicitação de verificação de números no WhatsApp
type CheckWhatsAppRequest struct {
	PhoneNumbers []string `json:"phone_numbers" validate:"required,min=1,max=50"`
}

// GetProfilePictureRequest representa uma solicitação de foto de perfil
type GetProfilePictureRequest struct {
	JID     string `json:"jid" validate:"required"`
	Preview bool   `json:"preview,omitempty"`
}

// GetProfilePictureInfoRequest representa uma solicitação de informações da foto de perfil
type GetProfilePictureInfoRequest struct {
	JID     string `json:"jid" validate:"required"`
	Preview bool   `json:"preview,omitempty"`
}

// GetUserInfoRequest representa uma solicitação de informações do usuário
type GetUserInfoRequest struct {
	JIDs []string `json:"jids" validate:"required,min=1,max=20"`
}

// GetDetailedUserInfoRequest representa uma solicitação de informações detalhadas
type GetDetailedUserInfoRequest struct {
	JIDs []string `json:"jids" validate:"required,min=1,max=20"`
}

// ListContactsRequest representa uma solicitação de listagem de contatos
type ListContactsRequest struct {
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// SyncContactsRequest representa uma solicitação de sincronização de contatos
type SyncContactsRequest struct {
	Force bool `json:"force,omitempty"`
}

// GetBusinessProfileRequest representa uma solicitação de perfil de negócio
type GetBusinessProfileRequest struct {
	JID string `json:"jid" validate:"required"`
}

// BlockContactRequest representa uma solicitação de bloqueio de contato
type BlockContactRequest struct {
	JID string `json:"jid" validate:"required"`
}

// UnblockContactRequest representa uma solicitação de desbloqueio de contato
type UnblockContactRequest struct {
	JID string `json:"jid" validate:"required"`
}

// ===== RESPONSE DTOs =====

// CheckWhatsAppResponse representa a resposta de verificação de números
type CheckWhatsAppResponse struct {
	Results []WhatsAppCheckResult `json:"results"`
	Total   int                   `json:"total"`
	Found   int                   `json:"found"`
	Success bool                  `json:"success"`
	Message string                `json:"message"`
}

// WhatsAppCheckResult representa o resultado de verificação de um número
type WhatsAppCheckResult struct {
	PhoneNumber  string `json:"phone_number"`
	IsOnWhatsApp bool   `json:"is_on_whatsapp"`
	JID          string `json:"jid,omitempty"`
}

// GetProfilePictureResponse representa a resposta de foto de perfil
type GetProfilePictureResponse struct {
	JID        string `json:"jid"`
	HasPicture bool   `json:"has_picture"`
	URL        string `json:"url,omitempty"`
	Data       []byte `json:"data,omitempty"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// GetProfilePictureInfoResponse representa a resposta de informações da foto
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

// GetUserInfoResponse representa a resposta de informações do usuário
type GetUserInfoResponse struct {
	Users   []UserInfo `json:"users"`
	Total   int        `json:"total"`
	Found   int        `json:"found"`
	Success bool       `json:"success"`
	Message string     `json:"message"`
}

// UserInfo representa informações de um usuário
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

// ListContactsResponse representa a resposta de listagem de contatos
type ListContactsResponse struct {
	Contacts []ContactDetails `json:"contacts"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
	Success  bool             `json:"success"`
	Message  string           `json:"message"`
}

// ContactDetails representa informações de um contato
type ContactDetails struct {
	JID          string `json:"jid"`
	PhoneNumber  string `json:"phone_number"`
	Name         string `json:"name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBusiness   bool   `json:"is_business"`
	IsContact    bool   `json:"is_contact"`
	IsBlocked    bool   `json:"is_blocked"`
}

// SyncContactsResponse representa a resposta de sincronização
type SyncContactsResponse struct {
	TotalContacts int    `json:"total_contacts"`
	SyncedCount   int    `json:"synced_count"`
	NewCount      int    `json:"new_count"`
	UpdatedCount  int    `json:"updated_count"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}

// GetBusinessProfileResponse representa a resposta de perfil de negócio
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

// BlockContactResponse representa a resposta de bloqueio
type BlockContactResponse struct {
	JID     string `json:"jid"`
	Blocked bool   `json:"blocked"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UnblockContactResponse representa a resposta de desbloqueio
type UnblockContactResponse struct {
	JID     string `json:"jid"`
	Blocked bool   `json:"blocked"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetAllContactsResponse representa a resposta de todos os contatos
type GetAllContactsResponse struct {
	Contacts []ContactDetails `json:"contacts"`
	Total    int              `json:"total"`
	Success  bool             `json:"success"`
	Message  string           `json:"message"`
}

// ===== DETAILED INFO DTOs =====

// GetDetailedUserInfoResponse representa a resposta de informações detalhadas
type GetDetailedUserInfoResponse struct {
	Users   []DetailedUserInfo `json:"users"`
	Total   int                `json:"total"`
	Found   int                `json:"found"`
	Success bool               `json:"success"`
	Message string             `json:"message"`
}

// DetailedUserInfo representa informações detalhadas de um usuário
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

// BusinessProfileInfo representa informações de perfil de negócio
type BusinessProfileInfo struct {
	BusinessName string `json:"business_name,omitempty"`
	Category     string `json:"category,omitempty"`
	Description  string `json:"description,omitempty"`
	Website      string `json:"website,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
}

// PrivacySettings representa configurações de privacidade
type PrivacySettings struct {
	LastSeen       string `json:"last_seen"`        // "everyone", "contacts", "nobody"
	ProfilePhoto   string `json:"profile_photo"`    // "everyone", "contacts", "nobody"
	Status         string `json:"status"`           // "everyone", "contacts", "nobody"
	ReadReceipts   bool   `json:"read_receipts"`
	Groups         string `json:"groups"`           // "everyone", "contacts", "nobody"
	CallsAdd       string `json:"calls_add"`        // "everyone", "contacts", "nobody"
}

// ===== VALIDATION DTOs =====

// ValidateContactRequest representa uma solicitação de validação de contato
type ValidateContactRequest struct {
	JID string `json:"jid" validate:"required"`
}

// ValidateContactResponse representa a resposta de validação
type ValidateContactResponse struct {
	JID       string `json:"jid"`
	IsValid   bool   `json:"is_valid"`
	Exists    bool   `json:"exists"`
	IsContact bool   `json:"is_contact"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}
