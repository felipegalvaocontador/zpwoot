package ports

import (
	"context"

	"zpwoot/internal/domain/community"
)

// CommunityManager defines the interface for community management operations
type CommunityManager interface {
	// LinkGroup links a group to a community
	LinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error

	// UnlinkGroup unlinks a group from a community
	UnlinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error

	// GetCommunityInfo gets information about a community
	GetCommunityInfo(ctx context.Context, sessionID string, communityJID string) (*community.CommunityInfo, error)

	// GetSubGroups gets all sub-groups (linked groups) of a community
	GetSubGroups(ctx context.Context, sessionID string, communityJID string) ([]*community.LinkedGroup, error)

	// GetLinkedGroupsParticipants gets participants from all linked groups in a community
	GetLinkedGroupsParticipants(ctx context.Context, sessionID string, communityJID string) ([]string, error)
}

// CommunityService defines the interface for community domain services
type CommunityService interface {
	// Validation methods
	ValidateCommunityJID(jid string) error
	ValidateGroupJID(jid string) error
	ValidateLinkRequest(communityJID, groupJID string) error

	// Formatting methods
	FormatCommunityJID(jid string) string
	FormatGroupJID(jid string) string
	SanitizeCommunityName(name string) string
	SanitizeCommunityDescription(description string) string

	// Processing methods
	ProcessCommunityInfo(info *community.CommunityInfo) error
	ProcessLinkedGroups(groups []*community.LinkedGroup) error
	ProcessGroupLinkResult(result *community.GroupLinkInfo) error

	// Business logic methods
	CanLinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanUnlinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanViewCommunity(communityJID string, userJID string) error

	// Utility methods
	GenerateCommunityEvent(eventType community.CommunityEventType, communityJID, actorJID, targetJID string, data map[string]interface{}) *community.CommunityEvent
}

// CommunityRepository defines the interface for community data persistence
// Note: For now, communities are managed entirely through WhatsApp
// This interface is kept for future extensions if we need local storage
type CommunityRepository interface {
	// GetCommunity gets a community by ID
	GetCommunity(ctx context.Context, id string) (*community.Community, error)

	// SaveCommunity saves a community
	SaveCommunity(ctx context.Context, community *community.Community) error

	// DeleteCommunity deletes a community
	DeleteCommunity(ctx context.Context, id string) error

	// ListCommunities lists communities for a user
	ListCommunities(ctx context.Context, userID string) ([]*community.Community, error)

	// GetLinkedGroups gets linked groups for a community
	GetLinkedGroups(ctx context.Context, communityID string) ([]*community.LinkedGroup, error)

	// SaveLinkedGroup saves a linked group relationship
	SaveLinkedGroup(ctx context.Context, communityID string, linkedGroup *community.LinkedGroup) error

	// RemoveLinkedGroup removes a linked group relationship
	RemoveLinkedGroup(ctx context.Context, communityID string, groupID string) error

	// GetCommunityEvents gets events for a community
	GetCommunityEvents(ctx context.Context, communityID string, limit int, offset int) ([]*community.CommunityEvent, error)

	// SaveCommunityEvent saves a community event
	SaveCommunityEvent(ctx context.Context, event *community.CommunityEvent) error
}

// CommunityIntegration defines the interface for community integrations
type CommunityIntegration interface {
	// OnGroupLinked is called when a group is linked to a community
	OnGroupLinked(ctx context.Context, communityJID, groupJID string, event *community.CommunityEvent) error

	// OnGroupUnlinked is called when a group is unlinked from a community
	OnGroupUnlinked(ctx context.Context, communityJID, groupJID string, event *community.CommunityEvent) error

	// OnCommunityInfoUpdated is called when community information is updated
	OnCommunityInfoUpdated(ctx context.Context, communityJID string, oldInfo, newInfo *community.CommunityInfo) error

	// OnParticipantJoined is called when a participant joins a community
	OnParticipantJoined(ctx context.Context, communityJID string, participantJID string, event *community.CommunityEvent) error

	// OnParticipantLeft is called when a participant leaves a community
	OnParticipantLeft(ctx context.Context, communityJID string, participantJID string, event *community.CommunityEvent) error
}

// CommunityNotifier defines the interface for community notifications
type CommunityNotifier interface {
	// NotifyGroupLinked notifies about a group being linked to a community
	NotifyGroupLinked(ctx context.Context, communityJID, groupJID string, actorJID string) error

	// NotifyGroupUnlinked notifies about a group being unlinked from a community
	NotifyGroupUnlinked(ctx context.Context, communityJID, groupJID string, actorJID string) error

	// NotifyCommunityUpdated notifies about community information updates
	NotifyCommunityUpdated(ctx context.Context, communityJID string, actorJID string, changes map[string]interface{}) error

	// NotifyParticipantActivity notifies about participant activity in a community
	NotifyParticipantActivity(ctx context.Context, communityJID string, participantJID string, activity string) error
}

// CommunityAnalytics defines the interface for community analytics
type CommunityAnalytics interface {
	// TrackGroupLink tracks a group link event
	TrackGroupLink(ctx context.Context, communityJID, groupJID string, actorJID string) error

	// TrackGroupUnlink tracks a group unlink event
	TrackGroupUnlink(ctx context.Context, communityJID, groupJID string, actorJID string) error

	// TrackCommunityActivity tracks general community activity
	TrackCommunityActivity(ctx context.Context, communityJID string, activityType string, metadata map[string]interface{}) error

	// GetCommunityStats gets statistics for a community
	GetCommunityStats(ctx context.Context, communityJID string) (*community.CommunityStats, error)

	// GetCommunityInsights gets insights for a community
	GetCommunityInsights(ctx context.Context, communityJID string, timeRange string) (map[string]interface{}, error)
}

// CommunityValidator defines the interface for community validation
type CommunityValidator interface {
	// ValidateCommunityCreation validates community creation parameters
	ValidateCommunityCreation(ctx context.Context, params map[string]interface{}) error

	// ValidateGroupLinking validates group linking parameters
	ValidateGroupLinking(ctx context.Context, communityJID, groupJID string, actorJID string) error

	// ValidateCommunityPermissions validates user permissions for community operations
	ValidateCommunityPermissions(ctx context.Context, communityJID string, userJID string, operation string) error

	// ValidateCommunitySettings validates community settings
	ValidateCommunitySettings(ctx context.Context, communityJID string, settings *community.CommunitySettings) error
}

// CommunityCache defines the interface for community caching
type CommunityCache interface {
	// GetCommunityInfo gets cached community information
	GetCommunityInfo(ctx context.Context, communityJID string) (*community.CommunityInfo, error)

	// SetCommunityInfo caches community information
	SetCommunityInfo(ctx context.Context, communityJID string, info *community.CommunityInfo, ttl int) error

	// GetLinkedGroups gets cached linked groups
	GetLinkedGroups(ctx context.Context, communityJID string) ([]*community.LinkedGroup, error)

	// SetLinkedGroups caches linked groups
	SetLinkedGroups(ctx context.Context, communityJID string, groups []*community.LinkedGroup, ttl int) error

	// InvalidateCommunity invalidates all cached data for a community
	InvalidateCommunity(ctx context.Context, communityJID string) error

	// InvalidateGroup invalidates cached data for a specific group
	InvalidateGroup(ctx context.Context, groupJID string) error
}
