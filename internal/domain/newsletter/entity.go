package newsletter

import (
	"errors"
	"strings"
	"time"
)

// Domain errors
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

// NewsletterRole represents the user's role in a newsletter
type NewsletterRole string

const (
	NewsletterRoleSubscriber NewsletterRole = "subscriber"
	NewsletterRoleGuest      NewsletterRole = "guest"
	NewsletterRoleAdmin      NewsletterRole = "admin"
	NewsletterRoleOwner      NewsletterRole = "owner"
)

// NewsletterState represents the state of a newsletter
type NewsletterState string

const (
	NewsletterStateActive       NewsletterState = "active"
	NewsletterStateSuspended    NewsletterState = "suspended"
	NewsletterStateGeoSuspended NewsletterState = "geosuspended"
)

// NewsletterMuteState represents the mute status
type NewsletterMuteState string

const (
	NewsletterMuteOn  NewsletterMuteState = "on"
	NewsletterMuteOff NewsletterMuteState = "off"
)

// NewsletterVerificationState represents verification status
type NewsletterVerificationState string

const (
	NewsletterVerificationStateVerified   NewsletterVerificationState = "verified"
	NewsletterVerificationStateUnverified NewsletterVerificationState = "unverified"
)

// ProfilePictureInfo represents newsletter profile picture information
type ProfilePictureInfo struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Direct string `json:"direct"`
}

// NewsletterInfo represents a WhatsApp newsletter/channel
type NewsletterInfo struct {
	ID                string                      `json:"id"`
	Name              string                      `json:"name"`
	Description       string                      `json:"description"`
	InviteCode        string                      `json:"inviteCode"`
	SubscriberCount   int                         `json:"subscriberCount"`
	State             NewsletterState             `json:"state"`
	Role              NewsletterRole              `json:"role"`
	Muted             bool                        `json:"muted"`
	MuteState         NewsletterMuteState         `json:"muteState"`
	Verified          bool                        `json:"verified"`
	VerificationState NewsletterVerificationState `json:"verificationState"`
	CreationTime      time.Time                   `json:"creationTime"`
	UpdateTime        time.Time                   `json:"updateTime"`
	Picture           *ProfilePictureInfo         `json:"picture,omitempty"`
	Preview           *ProfilePictureInfo         `json:"preview,omitempty"`
}

// CreateNewsletterRequest represents the data needed to create a newsletter
type CreateNewsletterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Picture     []byte `json:"picture,omitempty"`
}

// GetNewsletterInfoRequest represents a request to get newsletter information
type GetNewsletterInfoRequest struct {
	JID string `json:"jid"`
}

// GetNewsletterInfoWithInviteRequest represents a request to get newsletter info via invite
type GetNewsletterInfoWithInviteRequest struct {
	InviteKey string `json:"inviteKey"`
}

// FollowNewsletterRequest represents a request to follow a newsletter
type FollowNewsletterRequest struct {
	JID string `json:"jid"`
}

// UnfollowNewsletterRequest represents a request to unfollow a newsletter
type UnfollowNewsletterRequest struct {
	JID string `json:"jid"`
}

// Business logic methods

// Validate validates the newsletter information
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

// IsAdmin checks if the current user is an admin of the newsletter
func (n *NewsletterInfo) IsAdmin() bool {
	return n.Role == NewsletterRoleAdmin || n.Role == NewsletterRoleOwner
}

// IsOwner checks if the current user is the owner of the newsletter
func (n *NewsletterInfo) IsOwner() bool {
	return n.Role == NewsletterRoleOwner
}

// IsActive checks if the newsletter is active
func (n *NewsletterInfo) IsActive() bool {
	return n.State == NewsletterStateActive
}

// IsMuted checks if the newsletter is muted
func (n *NewsletterInfo) IsMuted() bool {
	return n.Muted || n.MuteState == NewsletterMuteOn
}

// IsVerified checks if the newsletter is verified
func (n *NewsletterInfo) IsVerified() bool {
	return n.Verified || n.VerificationState == NewsletterVerificationStateVerified
}

// CanManage checks if the current user can manage the newsletter
func (n *NewsletterInfo) CanManage() bool {
	return n.IsAdmin() && n.IsActive()
}

// GetDisplayName returns the display name for the newsletter
func (n *NewsletterInfo) GetDisplayName() string {
	if n.Name != "" {
		return n.Name
	}
	return n.ID
}

// HasPicture checks if the newsletter has a profile picture
func (n *NewsletterInfo) HasPicture() bool {
	return n.Picture != nil && n.Picture.URL != ""
}

// Validate validates the create newsletter request
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

// Validate validates the get newsletter info request
func (req *GetNewsletterInfoRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, "@newsletter") {
		return ErrInvalidNewsletterJID
	}
	return nil
}

// Validate validates the get newsletter info with invite request
func (req *GetNewsletterInfoWithInviteRequest) Validate() error {
	if req.InviteKey == "" {
		return ErrInvalidInviteKey
	}
	// Remove common prefixes if present
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "https://whatsapp.com/channel/")
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "whatsapp.com/channel/")
	req.InviteKey = strings.TrimPrefix(req.InviteKey, "channel/")

	if req.InviteKey == "" {
		return ErrInvalidInviteKey
	}
	return nil
}

// Validate validates the follow newsletter request
func (req *FollowNewsletterRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, "@newsletter") {
		return ErrInvalidNewsletterJID
	}
	return nil
}

// Validate validates the unfollow newsletter request
func (req *UnfollowNewsletterRequest) Validate() error {
	if req.JID == "" {
		return ErrInvalidNewsletterJID
	}
	if !strings.Contains(req.JID, "@newsletter") {
		return ErrInvalidNewsletterJID
	}
	return nil
}

// Helper functions

// IsValidNewsletterJID checks if a JID is a valid newsletter JID
func IsValidNewsletterJID(jid string) bool {
	return strings.Contains(jid, "@newsletter")
}

// IsValidNewsletterRole checks if a role is valid
func IsValidNewsletterRole(role string) bool {
	switch NewsletterRole(role) {
	case NewsletterRoleSubscriber, NewsletterRoleGuest, NewsletterRoleAdmin, NewsletterRoleOwner:
		return true
	default:
		return false
	}
}

// IsValidNewsletterState checks if a state is valid
func IsValidNewsletterState(state string) bool {
	switch NewsletterState(state) {
	case NewsletterStateActive, NewsletterStateSuspended, NewsletterStateGeoSuspended:
		return true
	default:
		return false
	}
}

// ParseNewsletterRole parses a string to NewsletterRole
func ParseNewsletterRole(role string) (NewsletterRole, error) {
	if !IsValidNewsletterRole(role) {
		return "", ErrInvalidNewsletterRole
	}
	return NewsletterRole(role), nil
}

// ParseNewsletterState parses a string to NewsletterState
func ParseNewsletterState(state string) (NewsletterState, error) {
	if !IsValidNewsletterState(state) {
		return "", ErrInvalidNewsletterState
	}
	return NewsletterState(state), nil
}

// NewsletterMessage represents a message in a newsletter
type NewsletterMessage struct {
	ID          string    `json:"id"`
	ServerID    string    `json:"serverId"`
	FromJID     string    `json:"fromJid"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Body        string    `json:"body,omitempty"`
	ViewsCount  int       `json:"viewsCount"`
	SharesCount int       `json:"sharesCount"`
	Reactions   []string  `json:"reactions,omitempty"`
}

// GetNewsletterMessagesRequest represents the request for getting newsletter messages
type GetNewsletterMessagesRequest struct {
	JID    string `json:"jid" validate:"required"`
	Count  int    `json:"count,omitempty"`
	Before string `json:"before,omitempty"` // MessageServerID
}

// GetNewsletterMessagesResponse represents the response for getting newsletter messages
type GetNewsletterMessagesResponse struct {
	Messages []*NewsletterMessage `json:"messages"`
	Total    int                  `json:"total"`
	HasMore  bool                 `json:"hasMore"`
}

// GetNewsletterMessageUpdatesRequest represents the request for getting newsletter message updates
type GetNewsletterMessageUpdatesRequest struct {
	JID   string `json:"jid" validate:"required"`
	Count int    `json:"count,omitempty"`
	Since string `json:"since,omitempty"` // ISO timestamp
	After string `json:"after,omitempty"` // MessageServerID
}

// GetNewsletterMessageUpdatesResponse represents the response for getting newsletter message updates
type GetNewsletterMessageUpdatesResponse struct {
	Updates []*NewsletterMessage `json:"updates"`
	Total   int                  `json:"total"`
	HasMore bool                 `json:"hasMore"`
}

// NewsletterMarkViewedRequest represents the request for marking newsletter messages as viewed
type NewsletterMarkViewedRequest struct {
	JID       string   `json:"jid" validate:"required"`
	ServerIDs []string `json:"serverIds" validate:"required,min=1"`
}

// NewsletterSendReactionRequest represents the request for sending a reaction to a newsletter message
type NewsletterSendReactionRequest struct {
	JID       string `json:"jid" validate:"required"`
	ServerID  string `json:"serverId" validate:"required"`
	Reaction  string `json:"reaction"`            // Empty string to remove reaction
	MessageID string `json:"messageId,omitempty"` // Optional, will be generated if empty
}

// NewsletterToggleMuteRequest represents the request for toggling newsletter mute status
type NewsletterToggleMuteRequest struct {
	JID  string `json:"jid" validate:"required"`
	Mute bool   `json:"mute"`
}

// NewsletterSubscribeLiveUpdatesRequest represents the request for subscribing to live updates
type NewsletterSubscribeLiveUpdatesRequest struct {
	JID string `json:"jid" validate:"required"`
}

// NewsletterSubscribeLiveUpdatesResponse represents the response for subscribing to live updates
type NewsletterSubscribeLiveUpdatesResponse struct {
	Duration int64 `json:"duration"` // Duration in seconds
}

// AcceptTOSNoticeRequest represents the request for accepting terms of service notice
type AcceptTOSNoticeRequest struct {
	NoticeID string `json:"noticeId" validate:"required"`
	Stage    string `json:"stage" validate:"required"`
}

// NewsletterActionResponse represents a generic response for newsletter actions
type NewsletterActionResponse struct {
	JID     string `json:"jid"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// UploadNewsletterRequest represents the request for uploading newsletter media
type UploadNewsletterRequest struct {
	Data      []byte `json:"data" validate:"required"`
	MimeType  string `json:"mimeType" validate:"required"`
	MediaType string `json:"mediaType" validate:"required"` // image, video, audio, document
}

// UploadNewsletterResponse represents the response for uploading newsletter media
type UploadNewsletterResponse struct {
	URL        string `json:"url"`
	DirectPath string `json:"directPath"`
	Handle     string `json:"handle"`
	ObjectID   string `json:"objectId"`
	FileSHA256 string `json:"fileSha256"`
	FileLength uint64 `json:"fileLength"`
}
