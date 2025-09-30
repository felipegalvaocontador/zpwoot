package ports

import (
	"context"

	"zpwoot/internal/domain/newsletter"
)

// NewsletterRepository defines the interface for newsletter data persistence
type NewsletterRepository interface {
	// Note: For now, newsletters are managed entirely through WhatsApp
	// This interface is kept for future extensions if we need local storage
	GetNewsletter(ctx context.Context, id string) (*newsletter.NewsletterInfo, error)
	SaveNewsletter(ctx context.Context, newsletter *newsletter.NewsletterInfo) error
	DeleteNewsletter(ctx context.Context, id string) error
	ListNewsletters(ctx context.Context, userID string) ([]*newsletter.NewsletterInfo, error)
}

// NewsletterManager defines the interface for WhatsApp newsletter operations
type NewsletterManager interface {
	// CreateNewsletter creates a new WhatsApp newsletter/channel
	CreateNewsletter(ctx context.Context, sessionID string, name, description string) (*newsletter.NewsletterInfo, error)

	// GetNewsletterInfo gets information about a newsletter by JID
	GetNewsletterInfo(ctx context.Context, sessionID string, jid string) (*newsletter.NewsletterInfo, error)

	// GetNewsletterInfoWithInvite gets newsletter information using an invite key
	GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, inviteKey string) (*newsletter.NewsletterInfo, error)

	// FollowNewsletter makes the user follow (subscribe to) a newsletter
	FollowNewsletter(ctx context.Context, sessionID string, jid string) error

	// UnfollowNewsletter makes the user unfollow (unsubscribe from) a newsletter
	UnfollowNewsletter(ctx context.Context, sessionID string, jid string) error

	// GetSubscribedNewsletters gets all newsletters the user is subscribed to
	GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]*newsletter.NewsletterInfo, error)

	// GetNewsletterMessages gets messages from a newsletter
	GetNewsletterMessages(ctx context.Context, sessionID string, jid string, count int, before string) ([]*newsletter.NewsletterMessage, error)

	// GetNewsletterMessageUpdates gets message updates from a newsletter (view counts, reactions)
	GetNewsletterMessageUpdates(ctx context.Context, sessionID string, jid string, count int, since string, after string) ([]*newsletter.NewsletterMessage, error)

	// NewsletterMarkViewed marks newsletter messages as viewed
	NewsletterMarkViewed(ctx context.Context, sessionID string, jid string, serverIDs []string) error

	// NewsletterSendReaction sends a reaction to a newsletter message
	NewsletterSendReaction(ctx context.Context, sessionID string, jid string, serverID string, reaction string, messageID string) error

	// NewsletterSubscribeLiveUpdates subscribes to live updates from a newsletter
	NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID string, jid string) (int64, error)

	// NewsletterToggleMute toggles mute status of a newsletter
	NewsletterToggleMute(ctx context.Context, sessionID string, jid string, mute bool) error

	// AcceptTOSNotice accepts a terms of service notice
	AcceptTOSNotice(ctx context.Context, sessionID string, noticeID string, stage string) error

	// UploadNewsletter uploads media for newsletters
	UploadNewsletter(ctx context.Context, sessionID string, data []byte, mimeType string, mediaType string) (*newsletter.UploadNewsletterResponse, error)

	// UploadNewsletterReader uploads media for newsletters from a reader
	UploadNewsletterReader(ctx context.Context, sessionID string, data []byte, mimeType string, mediaType string) (*newsletter.UploadNewsletterResponse, error)
}

// NewsletterService defines the interface for newsletter domain services
type NewsletterService interface {
	// Validation methods
	ValidateNewsletterCreation(req *newsletter.CreateNewsletterRequest) error
	ValidateNewsletterName(name string) error
	ValidateNewsletterDescription(description string) error
	ValidateNewsletterJID(jid string) error
	ValidateInviteKey(inviteKey string) error
	ValidateGetNewsletterInfoRequest(req *newsletter.GetNewsletterInfoRequest) error
	ValidateGetNewsletterInfoWithInviteRequest(req *newsletter.GetNewsletterInfoWithInviteRequest) error
	ValidateFollowNewsletterRequest(req *newsletter.FollowNewsletterRequest) error
	ValidateUnfollowNewsletterRequest(req *newsletter.UnfollowNewsletterRequest) error

	// Business logic methods
	ProcessNewsletterInfo(info *newsletter.NewsletterInfo) error
	CanUserManageNewsletter(info *newsletter.NewsletterInfo, userRole newsletter.NewsletterRole) bool
	CanUserFollowNewsletter(info *newsletter.NewsletterInfo) bool

	// Utility methods
	SanitizeNewsletterName(name string) string
	SanitizeNewsletterDescription(description string) string
	CleanInviteKey(inviteKey string) string
	FormatNewsletterJID(jid string) string
	ExtractNewsletterID(jid string) string
	IsNewsletterOwner(info *newsletter.NewsletterInfo) bool
	IsNewsletterAdmin(info *newsletter.NewsletterInfo) bool
	GetNewsletterPermissions(info *newsletter.NewsletterInfo) map[string]bool
}

// NewsletterUseCase defines the interface for newsletter application use cases
type NewsletterUseCase interface {
	// CreateNewsletter creates a new newsletter
	CreateNewsletter(ctx context.Context, sessionID string, req *newsletter.CreateNewsletterRequest) (*newsletter.NewsletterInfo, error)

	// GetNewsletterInfo gets newsletter information by JID
	GetNewsletterInfo(ctx context.Context, sessionID string, req *newsletter.GetNewsletterInfoRequest) (*newsletter.NewsletterInfo, error)

	// GetNewsletterInfoWithInvite gets newsletter information using invite key
	GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, req *newsletter.GetNewsletterInfoWithInviteRequest) (*newsletter.NewsletterInfo, error)

	// FollowNewsletter follows a newsletter
	FollowNewsletter(ctx context.Context, sessionID string, req *newsletter.FollowNewsletterRequest) error

	// UnfollowNewsletter unfollows a newsletter
	UnfollowNewsletter(ctx context.Context, sessionID string, req *newsletter.UnfollowNewsletterRequest) error

	// GetSubscribedNewsletters gets all subscribed newsletters
	GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]*newsletter.NewsletterInfo, error)
}

// NewsletterHandler defines the interface for newsletter HTTP handlers
type NewsletterHandler interface {
	// CreateNewsletter handles POST /newsletters/create
	CreateNewsletter(c interface{}) error

	// GetNewsletterInfo handles GET /newsletters/info
	GetNewsletterInfo(c interface{}) error

	// GetNewsletterInfoWithInvite handles POST /newsletters/info-from-invite
	GetNewsletterInfoWithInvite(c interface{}) error

	// FollowNewsletter handles POST /newsletters/follow
	FollowNewsletter(c interface{}) error

	// UnfollowNewsletter handles POST /newsletters/unfollow
	UnfollowNewsletter(c interface{}) error

	// GetSubscribedNewsletters handles GET /newsletters
	GetSubscribedNewsletters(c interface{}) error
}

// NewsletterEventHandler defines the interface for newsletter event handling
type NewsletterEventHandler interface {
	// OnNewsletterCreated handles newsletter creation events
	OnNewsletterCreated(ctx context.Context, sessionID string, newsletter *newsletter.NewsletterInfo) error

	// OnNewsletterFollowed handles newsletter follow events
	OnNewsletterFollowed(ctx context.Context, sessionID string, jid string) error

	// OnNewsletterUnfollowed handles newsletter unfollow events
	OnNewsletterUnfollowed(ctx context.Context, sessionID string, jid string) error

	// OnNewsletterUpdated handles newsletter update events
	OnNewsletterUpdated(ctx context.Context, sessionID string, newsletter *newsletter.NewsletterInfo) error
}

// NewsletterValidator defines the interface for newsletter validation
type NewsletterValidator interface {
	// ValidateJID validates a newsletter JID
	ValidateJID(jid string) error

	// ValidateInviteKey validates an invite key
	ValidateInviteKey(inviteKey string) error

	// ValidateName validates a newsletter name
	ValidateName(name string) error

	// ValidateDescription validates a newsletter description
	ValidateDescription(description string) error

	// IsValidNewsletterJID checks if a JID is a valid newsletter JID
	IsValidNewsletterJID(jid string) bool

	// CleanInviteKey cleans and normalizes an invite key
	CleanInviteKey(inviteKey string) string
}

// NewsletterConverter defines the interface for newsletter data conversion
type NewsletterConverter interface {
	// ToNewsletterInfo converts external data to domain NewsletterInfo
	ToNewsletterInfo(data interface{}) (*newsletter.NewsletterInfo, error)

	// FromNewsletterInfo converts domain NewsletterInfo to external format
	FromNewsletterInfo(info *newsletter.NewsletterInfo) (interface{}, error)

	// ToNewsletterInfoList converts external data list to domain NewsletterInfo list
	ToNewsletterInfoList(data interface{}) ([]*newsletter.NewsletterInfo, error)

	// FromNewsletterInfoList converts domain NewsletterInfo list to external format
	FromNewsletterInfoList(infos []*newsletter.NewsletterInfo) (interface{}, error)
}

// NewsletterCache defines the interface for newsletter caching
type NewsletterCache interface {
	// GetNewsletter gets a newsletter from cache
	GetNewsletter(ctx context.Context, sessionID, jid string) (*newsletter.NewsletterInfo, error)

	// SetNewsletter sets a newsletter in cache
	SetNewsletter(ctx context.Context, sessionID string, info *newsletter.NewsletterInfo) error

	// DeleteNewsletter deletes a newsletter from cache
	DeleteNewsletter(ctx context.Context, sessionID, jid string) error

	// GetSubscribedNewsletters gets subscribed newsletters from cache
	GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]*newsletter.NewsletterInfo, error)

	// SetSubscribedNewsletters sets subscribed newsletters in cache
	SetSubscribedNewsletters(ctx context.Context, sessionID string, infos []*newsletter.NewsletterInfo) error

	// ClearSubscribedNewsletters clears subscribed newsletters cache
	ClearSubscribedNewsletters(ctx context.Context, sessionID string) error
}

// NewsletterMetrics defines the interface for newsletter metrics
type NewsletterMetrics interface {
	// IncrementNewsletterCreated increments newsletter creation counter
	IncrementNewsletterCreated(sessionID string)

	// IncrementNewsletterFollowed increments newsletter follow counter
	IncrementNewsletterFollowed(sessionID string)

	// IncrementNewsletterUnfollowed increments newsletter unfollow counter
	IncrementNewsletterUnfollowed(sessionID string)

	// IncrementNewsletterInfoRequested increments newsletter info request counter
	IncrementNewsletterInfoRequested(sessionID string)

	// RecordNewsletterOperationDuration records operation duration
	RecordNewsletterOperationDuration(operation string, duration float64)

	// RecordNewsletterError records newsletter operation errors
	RecordNewsletterError(operation, errorType string)
}
