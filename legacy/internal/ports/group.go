package ports

import (
	"context"

	"zpwoot/internal/domain/group"
)

type GroupRepository interface {
	CreateGroup(ctx context.Context, sessionID string, req *group.CreateGroupRequest) (*group.GroupInfo, error)

	GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*group.GroupInfo, error)

	ListJoinedGroups(ctx context.Context, sessionID string) ([]*group.GroupInfo, error)

	UpdateGroupParticipants(ctx context.Context, sessionID string, req *group.UpdateParticipantsRequest) (*group.UpdateParticipantsResult, error)

	SetGroupName(ctx context.Context, sessionID string, req *group.SetGroupNameRequest) error

	SetGroupDescription(ctx context.Context, sessionID string, req *group.SetGroupDescriptionRequest) error

	SetGroupPhoto(ctx context.Context, sessionID string, req *group.SetGroupPhotoRequest) error

	GetGroupInviteLink(ctx context.Context, sessionID string, req *group.GetInviteLinkRequest) (*group.InviteLinkResponse, error)

	ResetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (*group.InviteLinkResponse, error)

	JoinGroupViaLink(ctx context.Context, sessionID string, req *group.JoinGroupRequest) (*group.GroupInfo, error)

	LeaveGroup(ctx context.Context, sessionID string, req *group.LeaveGroupRequest) error

	UpdateGroupSettings(ctx context.Context, sessionID string, req *group.UpdateGroupSettingsRequest) error

	IsGroupAdmin(ctx context.Context, sessionID, groupJID string) (bool, error)

	GetGroupParticipants(ctx context.Context, sessionID, groupJID string) ([]group.GroupParticipant, error)
}

type GroupService interface {
	ValidateGroupCreation(req *group.CreateGroupRequest) error

	ValidateParticipantUpdate(req *group.UpdateParticipantsRequest) error

	ValidateGroupName(name string) error

	ValidateGroupDescription(description string) error

	ValidateInviteLink(link string) error

	CanPerformAction(userJID, groupJID, action string, groupInfo *group.GroupInfo) error

	ProcessParticipantChanges(req *group.UpdateParticipantsRequest, currentGroup *group.GroupInfo) error
}
