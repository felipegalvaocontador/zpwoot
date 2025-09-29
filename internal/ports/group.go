package ports

import (
	"context"

	"zpwoot/internal/domain/group"
)

// GroupRepository defines the interface for group data operations
type GroupRepository interface {
	// CreateGroup creates a new WhatsApp group
	CreateGroup(ctx context.Context, sessionID string, req *group.CreateGroupRequest) (*group.GroupInfo, error)

	// GetGroupInfo retrieves information about a specific group
	GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*group.GroupInfo, error)

	// ListJoinedGroups lists all groups the session is a member of
	ListJoinedGroups(ctx context.Context, sessionID string) ([]*group.GroupInfo, error)

	// UpdateGroupParticipants adds, removes, promotes, or demotes group participants
	UpdateGroupParticipants(ctx context.Context, sessionID string, req *group.UpdateParticipantsRequest) (*group.UpdateParticipantsResult, error)

	// SetGroupName updates the group name
	SetGroupName(ctx context.Context, sessionID string, req *group.SetGroupNameRequest) error

	// SetGroupDescription updates the group description
	SetGroupDescription(ctx context.Context, sessionID string, req *group.SetGroupDescriptionRequest) error

	// SetGroupPhoto updates the group photo
	SetGroupPhoto(ctx context.Context, sessionID string, req *group.SetGroupPhotoRequest) error

	// GetGroupInviteLink retrieves or generates a group invite link
	GetGroupInviteLink(ctx context.Context, sessionID string, req *group.GetInviteLinkRequest) (*group.InviteLinkResponse, error)

	// ResetGroupInviteLink resets the group invite link
	ResetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (*group.InviteLinkResponse, error)

	// JoinGroupViaLink joins a group using an invite link
	JoinGroupViaLink(ctx context.Context, sessionID string, req *group.JoinGroupRequest) (*group.GroupInfo, error)

	// LeaveGroup leaves a group
	LeaveGroup(ctx context.Context, sessionID string, req *group.LeaveGroupRequest) error

	// UpdateGroupSettings updates group settings (announce, locked)
	UpdateGroupSettings(ctx context.Context, sessionID string, req *group.UpdateGroupSettingsRequest) error

	// IsGroupAdmin checks if the session user is an admin of the group
	IsGroupAdmin(ctx context.Context, sessionID, groupJID string) (bool, error)

	// GetGroupParticipants retrieves all participants of a group
	GetGroupParticipants(ctx context.Context, sessionID, groupJID string) ([]group.GroupParticipant, error)
}

// GroupService defines the interface for group business logic
type GroupService interface {
	// ValidateGroupCreation validates group creation parameters
	ValidateGroupCreation(req *group.CreateGroupRequest) error

	// ValidateParticipantUpdate validates participant update parameters
	ValidateParticipantUpdate(req *group.UpdateParticipantsRequest) error

	// ValidateGroupName validates group name
	ValidateGroupName(name string) error

	// ValidateGroupDescription validates group description
	ValidateGroupDescription(description string) error

	// ValidateInviteLink validates invite link format
	ValidateInviteLink(link string) error

	// CanPerformAction checks if user can perform a specific action on the group
	CanPerformAction(userJID, groupJID, action string, groupInfo *group.GroupInfo) error

	// ProcessParticipantChanges processes and validates participant changes
	ProcessParticipantChanges(req *group.UpdateParticipantsRequest, currentGroup *group.GroupInfo) error
}
