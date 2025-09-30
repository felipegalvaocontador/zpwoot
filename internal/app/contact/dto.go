package contact

import "time"

type CheckWhatsAppRequest struct {
	SessionID    string   `json:"sessionId,omitempty"`
	PhoneNumbers []string `json:"phoneNumbers" validate:"required,min=1,max=50" example:"[\"+5511999999999\", \"+5511888888888\"]"`
}

type WhatsAppStatus struct {
	PhoneNumber  string `json:"phoneNumber" example:"+5511999999999"`
	JID          string `json:"jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	VerifiedName string `json:"verifiedName,omitempty" example:"Company Name"`
	IsOnWhatsApp bool   `json:"isOnWhatsapp" example:"true"`
	IsBusiness   bool   `json:"isBusiness" example:"false"`
}

type CheckWhatsAppResponse struct {
	Results []WhatsAppStatus `json:"results"`
	Total   int              `json:"total" example:"2"`
	Checked int              `json:"checked" example:"2"`
}

type GetProfilePictureRequest struct {
	SessionID string `json:"sessionId,omitempty"`
	JID       string `json:"jid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Preview   bool   `json:"preview" example:"false"`
}

type ProfilePictureResponse struct {
	UpdatedAt  time.Time `json:"updatedAt,omitempty" example:"2024-01-01T12:00:00Z"`
	JID        string    `json:"jid" example:"5511999999999@s.whatsapp.net"`
	URL        string    `json:"url,omitempty" example:"https://pps.whatsapp.net/v/..."`
	ID         string    `json:"id,omitempty" example:"1234567890"`
	Type       string    `json:"type,omitempty" example:"image"`
	DirectPath string    `json:"directPath,omitempty"`
	HasPicture bool      `json:"hasPicture" example:"true"`
}

type GetUserInfoRequest struct {
	SessionID string   `json:"sessionId,omitempty"`
	JIDs      []string `json:"jids" validate:"required,min=1,max=20" example:"[\"5511999999999@s.whatsapp.net\", \"5511888888888@s.whatsapp.net\"]"`
}

type UserInfo struct {
	LastSeen     *time.Time `json:"lastSeen,omitempty" example:"2024-01-01T12:00:00Z"`
	JID          string     `json:"jid" example:"5511999999999@s.whatsapp.net"`
	PhoneNumber  string     `json:"phoneNumber" example:"+5511999999999"`
	Name         string     `json:"name,omitempty" example:"John Doe"`
	Status       string     `json:"status,omitempty" example:"Hey there! I am using WhatsApp."`
	PictureID    string     `json:"pictureId,omitempty" example:"1234567890"`
	VerifiedName string     `json:"verifiedName,omitempty" example:"Company Name"`
	IsBusiness   bool       `json:"isBusiness" example:"false"`
	IsContact    bool       `json:"isContact" example:"true"`
	IsOnline     bool       `json:"isOnline" example:"false"`
}

type GetUserInfoResponse struct {
	Users []UserInfo `json:"users"`
	Total int        `json:"total" example:"2"`
	Found int        `json:"found" example:"2"`
}

type ListContactsRequest struct {
	SessionID string `json:"sessionId,omitempty"`
	Search    string `json:"search,omitempty" example:"John"`
	Limit     int    `json:"limit" validate:"min=1,max=100" example:"50"`
	Offset    int    `json:"offset" validate:"min=0" example:"0"`
}

type Contact struct {
	AddedAt     time.Time `json:"addedAt,omitempty" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty" example:"2024-01-01T12:00:00Z"`
	JID         string    `json:"jid" example:"5511999999999@s.whatsapp.net"`
	PhoneNumber string    `json:"phoneNumber" example:"+5511999999999"`
	Name        string    `json:"name,omitempty" example:"John Doe"`
	ShortName   string    `json:"shortName,omitempty" example:"John"`
	PushName    string    `json:"pushName,omitempty" example:"John"`
	IsBusiness  bool      `json:"isBusiness" example:"false"`
	IsContact   bool      `json:"isContact" example:"true"`
	IsBlocked   bool      `json:"isBlocked" example:"false"`
}

type ListContactsResponse struct {
	Contacts []Contact `json:"contacts"`
	Total    int       `json:"total" example:"150"`
	Limit    int       `json:"limit" example:"50"`
	Offset   int       `json:"offset" example:"0"`
	HasMore  bool      `json:"hasMore" example:"true"`
}

type SyncContactsRequest struct {
	SessionID string `json:"sessionId,omitempty"`
	Force     bool   `json:"force" example:"false"`
}

type SyncContactsResponse struct {
	SyncedAt time.Time `json:"syncedAt" example:"2024-01-01T12:00:00Z"`
	Message  string    `json:"message" example:"Contacts synchronized successfully"`
	Synced   int       `json:"synced" example:"25"`
	Added    int       `json:"added" example:"5"`
	Updated  int       `json:"updated" example:"3"`
	Removed  int       `json:"removed" example:"1"`
	Total    int       `json:"total" example:"150"`
}

type GetBusinessProfileRequest struct {
	SessionID string `json:"sessionId,omitempty"`
	JID       string `json:"jid" validate:"required" example:"5511999999999@s.whatsapp.net"`
}

type BusinessProfile struct {
	JID         string `json:"jid" example:"5511999999999@s.whatsapp.net"`
	Name        string `json:"name,omitempty" example:"My Business"`
	Category    string `json:"category,omitempty" example:"Retail"`
	Description string `json:"description,omitempty" example:"We sell amazing products"`
	Website     string `json:"website,omitempty" example:"https://mybusiness.com"`
	Email       string `json:"email,omitempty" example:"contact@mybusiness.com"`
	Address     string `json:"address,omitempty" example:"123 Main St, City"`
	Verified    bool   `json:"verified" example:"true"`
}

type BusinessProfileResponse struct {
	UpdatedAt time.Time       `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
	Profile   BusinessProfile `json:"profile"`
	Found     bool            `json:"found" example:"true"`
}

type ContactStats struct {
	LastSyncAt       *time.Time `json:"lastSyncAt,omitempty" example:"2024-01-01T12:00:00Z"`
	TotalContacts    int        `json:"totalContacts" example:"150"`
	WhatsAppContacts int        `json:"whatsappContacts" example:"120"`
	BusinessContacts int        `json:"businessContacts" example:"10"`
	BlockedContacts  int        `json:"blockedContacts" example:"2"`
	SyncRate         float64    `json:"syncRate" example:"0.8"`
}

type GetContactStatsRequest struct {
	SessionID string `json:"sessionId,omitempty"`
}

type GetContactStatsResponse struct {
	Stats     ContactStats `json:"stats"`
	UpdatedAt time.Time    `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
	SessionID string       `json:"sessionId" example:"session-123"`
}
