package newsletter

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidNewsletterJID   = errors.New("invalid newsletter JID")
	ErrInvalidNewsletterName  = errors.New("invalid newsletter name")
	ErrNewsletterNameTooLong  = errors.New("newsletter name too long (max 64 characters)")
	ErrDescriptionTooLong     = errors.New("description too long (max 256 characters)")
	ErrInvalidInviteKey       = errors.New("invalid invite key")
	ErrNewsletterNotFound     = errors.New("newsletter not found")
	ErrNotNewsletterAdmin     = errors.New("user is not a newsletter admin")
	ErrEmptyNewsletterName    = errors.New("newsletter name cannot be empty")
	ErrInvalidNewsletterRole  = errors.New("invalid newsletter role")
	ErrInvalidNewsletterState = errors.New("invalid newsletter state")
)

const (
	NewsletterJIDSuffix = "@newsletter"
)

type NewsletterRole string

const (
	NewsletterRoleSubscriber NewsletterRole = "subscriber"
	NewsletterRoleGuest      NewsletterRole = "guest"
	NewsletterRoleAdmin      NewsletterRole = "admin"
	NewsletterRoleOwner      NewsletterRole = "owner"
)

type NewsletterState string

const (
	NewsletterStateActive       NewsletterState = "active"
	NewsletterStateSuspended    NewsletterState = "suspended"
	NewsletterStateGeoSuspended NewsletterState = "geosuspended"
)

type NewsletterMuteState string

const (
	NewsletterMuteOn  NewsletterMuteState = "on"
	NewsletterMuteOff NewsletterMuteState = "off"
)

type NewsletterVerificationState string

const (
	NewsletterVerificationStateVerified   NewsletterVerificationState = "verified"
	NewsletterVerificationStateUnverified NewsletterVerificationState = "unverified"
)

type ProfilePictureInfo struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Direct string `json:"direct"`
}

type NewsletterInfo struct {
	CreationTime      time.Time                   `json:"creationTime"`
	UpdateTime        time.Time                   `json:"updateTime"`
	Preview           *ProfilePictureInfo         `json:"preview,omitempty"`
	Picture           *ProfilePictureInfo         `json:"picture,omitempty"`
	Role              NewsletterRole              `json:"role"`
	State             NewsletterState             `json:"state"`
	ID                string                      `json:"id"`
	MuteState         NewsletterMuteState         `json:"muteState"`
	VerificationState NewsletterVerificationState `json:"verificationState"`
	InviteCode        string                      `json:"inviteCode"`
	Description       string                      `json:"description"`
	Name              string                      `json:"name"`
	SubscriberCount   int                         `json:"subscriberCount"`
	Muted             bool                        `json:"muted"`
	Verified          bool                        `json:"verified"`
}

type CreateNewsletterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Picture     []byte `json:"picture,omitempty"`
}

type GetNewsletterInfoRequest struct {
	JID string `json:"jid"`
}

type GetNewsletterInfoWithInviteRequest struct {
	InviteKey string `json:"inviteKey"`
}

type FollowNewsletterRequest struct {
	JID string `json:"jid"`
}

type UnfollowNewsletterRequest struct {
	JID string `json:"jid"`
}

func (n *NewsletterInfo) Validate() error {
	if n.ID == "" {
		return ErrInvalidNewsletterJID
	}
	if n.Name == "" {
		return ErrEmptyNewsletterName
	}
	if len(n.Name) > 64 {
		return ErrNewsletterNameTooLong
	}
	if len(n.Description) > 256 {
		return ErrDescriptionTooLong
	}
	return nil
}

func (n *NewsletterInfo) IsAdmin() bool {
	return n.Role == NewsletterRoleAdmin || n.Role == NewsletterRoleOwner
}

func (n *NewsletterInfo) IsOwner() bool {
	return n.Role == NewsletterRoleOwner
}

func (n *NewsletterInfo) IsActive() bool {
	return n.State == NewsletterStateActive
}

func (n *NewsletterInfo) IsMuted() bool {
	return n.Muted || n.MuteState == NewsletterMuteOn
}

func (n *NewsletterInfo) IsVerified() bool {
	return n.Verified || n.VerificationState == NewsletterVerificationStateVerified
}

func (n *NewsletterInfo) CanManage() bool {
	return n.IsAdmin() && n.IsActive()
}

func (n *NewsletterInfo) GetDisplayName() string {
	if n.Name != "" {
		return n.Name
	}
	return n.ID
}

func (n *NewsletterInfo) HasPicture() bool {
	return n.Picture != nil && n.Picture.URL != ""
}

func (req *CreateNewsletterRequest) Validate() error {
	if req.Name == "" {
		return ErrEmptyNewsletterName
	}
	if len(req.Name) > 64 {
		return ErrNewsletterNameTooLong
	}
	if len(req.Description) > 256 {
		return ErrDescriptionTooLong
	}
	return nil
}

func (req *GetNewsletterInfoRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, NewsletterJIDSuffix) {
		return ErrInvalidNewsletterJID
	}
	return nil
}

func (req *GetNewsletterInfoWithInviteRequest) Validate() error {
	if req.InviteKey == "" {
		return ErrInvalidInviteKey
	}
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "https://whatsapp.com/channel/")
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "whatsapp.com/channel/")
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "channel/")

	if req.InviteKey == "" {
		return ErrInvalidInviteKey
	}
	return nil
}

func (req *FollowNewsletterRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, NewsletterJIDSuffix) {
		return ErrInvalidNewsletterJID
	}
	return nil
}

func (req *UnfollowNewsletterRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, NewsletterJIDSuffix) {
		return ErrInvalidNewsletterJID
	}
	return nil
}

func IsValidNewsletterJID(jid string) bool {
	return strings.Contains(jid, NewsletterJIDSuffix)
}

func IsValidNewsletterRole(role string) bool {
	switch NewsletterRole(role) {
	case NewsletterRoleSubscriber, NewsletterRoleGuest, NewsletterRoleAdmin, NewsletterRoleOwner:
		return true
	default:
		return false
	}
}

func IsValidNewsletterState(state string) bool {
	switch NewsletterState(state) {
	case NewsletterStateActive, NewsletterStateSuspended, NewsletterStateGeoSuspended:
		return true
	default:
		return false
	}
}

func ParseNewsletterRole(role string) (NewsletterRole, error) {
	if !IsValidNewsletterRole(role) {
		return "", ErrInvalidNewsletterRole
	}
	return NewsletterRole(role), nil
}

func ParseNewsletterState(state string) (NewsletterState, error) {
	if !IsValidNewsletterState(state) {
		return "", ErrInvalidNewsletterState
	}
	return NewsletterState(state), nil
}

type NewsletterMessage struct {
	Timestamp   time.Time `json:"timestamp"`
	ID          string    `json:"id"`
	ServerID    string    `json:"serverId"`
	FromJID     string    `json:"fromJid"`
	Type        string    `json:"type"`
	Body        string    `json:"body,omitempty"`
	Reactions   []string  `json:"reactions,omitempty"`
	ViewsCount  int       `json:"viewsCount"`
	SharesCount int       `json:"sharesCount"`
}

type GetNewsletterMessagesRequest struct {
	JID    string `json:"jid" validate:"required"`
	Before string `json:"before,omitempty"`
	Count  int    `json:"count,omitempty"`
}

type GetNewsletterMessagesResponse struct {
	Messages []*NewsletterMessage `json:"messages"`
	Total    int                  `json:"total"`
	HasMore  bool                 `json:"hasMore"`
}

type GetNewsletterMessageUpdatesRequest struct {
	JID   string `json:"jid" validate:"required"`
	Since string `json:"since,omitempty"`
	After string `json:"after,omitempty"`
	Count int    `json:"count,omitempty"`
}

type GetNewsletterMessageUpdatesResponse struct {
	Updates []*NewsletterMessage `json:"updates"`
	Total   int                  `json:"total"`
	HasMore bool                 `json:"hasMore"`
}

type NewsletterMarkViewedRequest struct {
	JID       string   `json:"jid" validate:"required"`
	ServerIDs []string `json:"serverIds" validate:"required,min=1"`
}

type NewsletterSendReactionRequest struct {
	JID       string `json:"jid" validate:"required"`
	ServerID  string `json:"serverId" validate:"required"`
	Reaction  string `json:"reaction"`            // Empty string to remove reaction
	MessageID string `json:"messageId,omitempty"` // Optional, will be generated if empty
}

type NewsletterToggleMuteRequest struct {
	JID  string `json:"jid" validate:"required"`
	Mute bool   `json:"mute"`
}

type NewsletterSubscribeLiveUpdatesRequest struct {
	JID string `json:"jid" validate:"required"`
}

type NewsletterSubscribeLiveUpdatesResponse struct {
	Duration int64 `json:"duration"` // Duration in seconds
}

type AcceptTOSNoticeRequest struct {
	NoticeID string `json:"noticeId" validate:"required"`
	Stage    string `json:"stage" validate:"required"`
}

type NewsletterActionResponse struct {
	JID     string `json:"jid"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UploadNewsletterRequest struct {
	MimeType  string `json:"mimeType" validate:"required"`
	MediaType string `json:"mediaType" validate:"required"`
	Data      []byte `json:"data" validate:"required"`
}

type UploadNewsletterResponse struct {
	URL        string `json:"url"`
	DirectPath string `json:"directPath"`
	Handle     string `json:"handle"`
	ObjectID   string `json:"objectId"`
	FileSHA256 string `json:"fileSha256"`
	FileLength uint64 `json:"fileLength"`
}
