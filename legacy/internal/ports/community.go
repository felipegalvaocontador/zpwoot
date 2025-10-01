package ports

import (
	"context"

	"zpwoot/internal/domain/community"
)

type CommunityManager interface {
	LinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error

	UnlinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error

	GetCommunityInfo(ctx context.Context, sessionID string, communityJID string) (*community.CommunityInfo, error)

	GetSubGroups(ctx context.Context, sessionID string, communityJID string) ([]*community.LinkedGroup, error)

	GetLinkedGroupsParticipants(ctx context.Context, sessionID string, communityJID string) ([]string, error)
}

type CommunityService interface {
	ValidateCommunityJID(jid string) error
	ValidateGroupJID(jid string) error
	ValidateLinkRequest(communityJID, groupJID string) error

	FormatCommunityJID(jid string) string
	FormatGroupJID(jid string) string
	SanitizeCommunityName(name string) string
	SanitizeCommunityDescription(description string) string

	ProcessCommunityInfo(info *community.CommunityInfo) error
	ProcessLinkedGroups(groups []*community.LinkedGroup) error
	ProcessGroupLinkResult(result *community.GroupLinkInfo) error

	CanLinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanUnlinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanViewCommunity(communityJID string, userJID string) error

	GenerateCommunityEvent(eventType community.CommunityEventType, communityJID, actorJID, targetJID string, data map[string]interface{}) *community.CommunityEvent
}

type CommunityRepository interface {
	GetCommunity(ctx context.Context, id string) (*community.Community, error)

	SaveCommunity(ctx context.Context, community *community.Community) error

	DeleteCommunity(ctx context.Context, id string) error

	ListCommunities(ctx context.Context, userID string) ([]*community.Community, error)

	GetLinkedGroups(ctx context.Context, communityID string) ([]*community.LinkedGroup, error)

	SaveLinkedGroup(ctx context.Context, communityID string, linkedGroup *community.LinkedGroup) error

	RemoveLinkedGroup(ctx context.Context, communityID string, groupID string) error

	GetCommunityEvents(ctx context.Context, communityID string, limit int, offset int) ([]*community.CommunityEvent, error)

	SaveCommunityEvent(ctx context.Context, event *community.CommunityEvent) error
}

type CommunityIntegration interface {
	OnGroupLinked(ctx context.Context, communityJID, groupJID string, event *community.CommunityEvent) error

	OnGroupUnlinked(ctx context.Context, communityJID, groupJID string, event *community.CommunityEvent) error

	OnCommunityInfoUpdated(ctx context.Context, communityJID string, oldInfo, newInfo *community.CommunityInfo) error

	OnParticipantJoined(ctx context.Context, communityJID string, participantJID string, event *community.CommunityEvent) error

	OnParticipantLeft(ctx context.Context, communityJID string, participantJID string, event *community.CommunityEvent) error
}

type CommunityNotifier interface {
	NotifyGroupLinked(ctx context.Context, communityJID, groupJID string, actorJID string) error

	NotifyGroupUnlinked(ctx context.Context, communityJID, groupJID string, actorJID string) error

	NotifyCommunityUpdated(ctx context.Context, communityJID string, actorJID string, changes map[string]interface{}) error

	NotifyParticipantActivity(ctx context.Context, communityJID string, participantJID string, activity string) error
}

type CommunityAnalytics interface {
	TrackGroupLink(ctx context.Context, communityJID, groupJID string, actorJID string) error

	TrackGroupUnlink(ctx context.Context, communityJID, groupJID string, actorJID string) error

	TrackCommunityActivity(ctx context.Context, communityJID string, activityType string, metadata map[string]interface{}) error

	GetCommunityStats(ctx context.Context, communityJID string) (*community.CommunityStats, error)

	GetCommunityInsights(ctx context.Context, communityJID string, timeRange string) (map[string]interface{}, error)
}

type CommunityValidator interface {
	ValidateCommunityCreation(ctx context.Context, params map[string]interface{}) error

	ValidateGroupLinking(ctx context.Context, communityJID, groupJID string, actorJID string) error

	ValidateCommunityPermissions(ctx context.Context, communityJID string, userJID string, operation string) error

	ValidateCommunitySettings(ctx context.Context, communityJID string, settings *community.CommunitySettings) error
}

type CommunityCache interface {
	GetCommunityInfo(ctx context.Context, communityJID string) (*community.CommunityInfo, error)

	SetCommunityInfo(ctx context.Context, communityJID string, info *community.CommunityInfo, ttl int) error

	GetLinkedGroups(ctx context.Context, communityJID string) ([]*community.LinkedGroup, error)

	SetLinkedGroups(ctx context.Context, communityJID string, groups []*community.LinkedGroup, ttl int) error

	InvalidateCommunity(ctx context.Context, communityJID string) error

	InvalidateGroup(ctx context.Context, groupJID string) error
}
