package ports

import (
	"context"

	"zpwoot/internal/domain/newsletter"
)

type NewsletterRepository interface {
	GetNewsletter(ctx context.Context, id string) (*newsletter.NewsletterInfo, error)
	SaveNewsletter(ctx context.Context, newsletter *newsletter.NewsletterInfo) error
	DeleteNewsletter(ctx context.Context, id string) error
	ListNewsletters(ctx context.Context, userID string) ([]*newsletter.NewsletterInfo, error)
}

type NewsletterManager interface {
	CreateNewsletter(ctx context.Context, sessionID string, name, description string) (*newsletter.NewsletterInfo, error)

	GetNewsletterInfo(ctx context.Context, sessionID string, jid string) (*newsletter.NewsletterInfo, error)

	GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, inviteKey string) (*newsletter.NewsletterInfo, error)

	FollowNewsletter(ctx context.Context, sessionID string, jid string) error

	UnfollowNewsletter(ctx context.Context, sessionID string, jid string) error

	GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]*newsletter.NewsletterInfo, error)

	GetNewsletterMessages(ctx context.Context, sessionID string, jid string, count int, before string) ([]*newsletter.NewsletterMessage, error)

	GetNewsletterMessageUpdates(ctx context.Context, sessionID string, jid string, count int, since string, after string) ([]*newsletter.NewsletterMessage, error)

	NewsletterMarkViewed(ctx context.Context, sessionID string, jid string, serverIDs []string) error

	NewsletterSendReaction(ctx context.Context, sessionID string, jid string, serverID string, reaction string, messageID string) error

	NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID string, jid string) (int64, error)

	NewsletterToggleMute(ctx context.Context, sessionID string, jid string, mute bool) error

	AcceptTOSNotice(ctx context.Context, sessionID string, noticeID string, stage string) error

	UploadNewsletter(ctx context.Context, sessionID string, data []byte, mimeType string, mediaType string) (*newsletter.UploadNewsletterResponse, error)

	UploadNewsletterReader(ctx context.Context, sessionID string, data []byte, mimeType string, mediaType string) (*newsletter.UploadNewsletterResponse, error)
}

type NewsletterService interface {
	ValidateNewsletterCreation(req *newsletter.CreateNewsletterRequest) error
	ValidateNewsletterName(name string) error
	ValidateNewsletterDescription(description string) error
	ValidateNewsletterJID(jid string) error
	ValidateInviteKey(inviteKey string) error
	ValidateGetNewsletterInfoRequest(req *newsletter.GetNewsletterInfoRequest) error
	ValidateGetNewsletterInfoWithInviteRequest(req *newsletter.GetNewsletterInfoWithInviteRequest) error
	ValidateFollowNewsletterRequest(req *newsletter.FollowNewsletterRequest) error
	ValidateUnfollowNewsletterRequest(req *newsletter.UnfollowNewsletterRequest) error

	ProcessNewsletterInfo(info *newsletter.NewsletterInfo) error
	CanUserManageNewsletter(info *newsletter.NewsletterInfo, userRole newsletter.NewsletterRole) bool
	CanUserFollowNewsletter(info *newsletter.NewsletterInfo) bool

	SanitizeNewsletterName(name string) string
	SanitizeNewsletterDescription(description string) string
	CleanInviteKey(inviteKey string) string
	FormatNewsletterJID(jid string) string
	ExtractNewsletterID(jid string) string
	IsNewsletterOwner(info *newsletter.NewsletterInfo) bool
	IsNewsletterAdmin(info *newsletter.NewsletterInfo) bool
	GetNewsletterPermissions(info *newsletter.NewsletterInfo) map[string]bool
}

type NewsletterUseCase interface {
	CreateNewsletter(ctx context.Context, sessionID string, req *newsletter.CreateNewsletterRequest) (*newsletter.NewsletterInfo, error)

	GetNewsletterInfo(ctx context.Context, sessionID string, req *newsletter.GetNewsletterInfoRequest) (*newsletter.NewsletterInfo, error)

	GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, req *newsletter.GetNewsletterInfoWithInviteRequest) (*newsletter.NewsletterInfo, error)

	FollowNewsletter(ctx context.Context, sessionID string, req *newsletter.FollowNewsletterRequest) error

	UnfollowNewsletter(ctx context.Context, sessionID string, req *newsletter.UnfollowNewsletterRequest) error

	GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]*newsletter.NewsletterInfo, error)
}

type NewsletterHandler interface {
	CreateNewsletter(c interface{}) error

	GetNewsletterInfo(c interface{}) error

	GetNewsletterInfoWithInvite(c interface{}) error

	FollowNewsletter(c interface{}) error

	UnfollowNewsletter(c interface{}) error

	GetSubscribedNewsletters(c interface{}) error
}

type NewsletterEventHandler interface {
	OnNewsletterCreated(ctx context.Context, sessionID string, newsletter *newsletter.NewsletterInfo) error

	OnNewsletterFollowed(ctx context.Context, sessionID string, jid string) error

	OnNewsletterUnfollowed(ctx context.Context, sessionID string, jid string) error

	OnNewsletterUpdated(ctx context.Context, sessionID string, newsletter *newsletter.NewsletterInfo) error
}

type NewsletterValidator interface {
	ValidateJID(jid string) error

	ValidateInviteKey(inviteKey string) error

	ValidateName(name string) error

	ValidateDescription(description string) error

	IsValidNewsletterJID(jid string) bool

	CleanInviteKey(inviteKey string) string
}

type NewsletterConverter interface {
	ToNewsletterInfo(data interface{}) (*newsletter.NewsletterInfo, error)

	FromNewsletterInfo(info *newsletter.NewsletterInfo) (interface{}, error)

	ToNewsletterInfoList(data interface{}) ([]*newsletter.NewsletterInfo, error)

	FromNewsletterInfoList(infos []*newsletter.NewsletterInfo) (interface{}, error)
}

type NewsletterCache interface {
	GetNewsletter(ctx context.Context, sessionID, jid string) (*newsletter.NewsletterInfo, error)

	SetNewsletter(ctx context.Context, sessionID string, info *newsletter.NewsletterInfo) error

	DeleteNewsletter(ctx context.Context, sessionID, jid string) error

	GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]*newsletter.NewsletterInfo, error)

	SetSubscribedNewsletters(ctx context.Context, sessionID string, infos []*newsletter.NewsletterInfo) error

	ClearSubscribedNewsletters(ctx context.Context, sessionID string) error
}

type NewsletterMetrics interface {
	IncrementNewsletterCreated(sessionID string)

	IncrementNewsletterFollowed(sessionID string)

	IncrementNewsletterUnfollowed(sessionID string)

	IncrementNewsletterInfoRequested(sessionID string)

	RecordNewsletterOperationDuration(operation string, duration float64)

	RecordNewsletterError(operation, errorType string)
}
