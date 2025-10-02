package group

import (
	"context"
)


type Repository interface {

	Create(ctx context.Context, group *Group) error
	GetByID(ctx context.Context, id string) (*Group, error)
	GetByGroupJID(ctx context.Context, sessionID, groupJID string) (*Group, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id string) error


	ListBySession(ctx context.Context, sessionID string) ([]*Group, error)
	ListJoinedGroups(ctx context.Context, sessionID string) ([]*Group, error)


	AddParticipant(ctx context.Context, groupID string, participant *Participant) error
	RemoveParticipant(ctx context.Context, groupID, participantJID string) error
	UpdateParticipant(ctx context.Context, groupID string, participant *Participant) error
	GetParticipants(ctx context.Context, groupID string) ([]Participant, error)


	UpdateSettings(ctx context.Context, groupID string, settings *GroupSettings) error


	SaveInviteLink(ctx context.Context, invite *InviteLink) error
	GetInviteLink(ctx context.Context, groupJID string) (*InviteLink, error)
	RevokeInviteLink(ctx context.Context, groupJID string) error


	SaveGroupRequest(ctx context.Context, request *GroupRequest) error
	GetGroupRequests(ctx context.Context, groupJID string) ([]*GroupRequest, error)
	UpdateGroupRequest(ctx context.Context, groupJID, requesterJID, status string) error
}


type WhatsAppGateway interface {

	CreateGroup(ctx context.Context, sessionID, name string, participants []string, description string) (*GroupInfo, error)
	GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*GroupInfo, error)
	ListJoinedGroups(ctx context.Context, sessionID string) ([]*GroupInfo, error)


	AddParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	RemoveParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	PromoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	DemoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error


	SetGroupName(ctx context.Context, sessionID, groupJID, name string) error
	SetGroupDescription(ctx context.Context, sessionID, groupJID, description string) error
	SetGroupPhoto(ctx context.Context, sessionID, groupJID string, photoData []byte) error


	SetGroupAnnounce(ctx context.Context, sessionID, groupJID string, announce bool) error
	SetGroupRestrict(ctx context.Context, sessionID, groupJID string, restrict bool) error
	SetGroupLocked(ctx context.Context, sessionID, groupJID string, locked bool) error


	GetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (*InviteLink, error)
	RevokeGroupInviteLink(ctx context.Context, sessionID, groupJID string) error
	JoinGroupViaLink(ctx context.Context, sessionID, inviteLink string) (*GroupInfo, error)


	LeaveGroup(ctx context.Context, sessionID, groupJID string) error
	JoinGroupWithInvite(ctx context.Context, sessionID, groupJID, inviteCode string) (*GroupInfo, error)


	GetGroupRequestParticipants(ctx context.Context, sessionID, groupJID string) ([]*GroupRequest, error)
	ApproveGroupRequest(ctx context.Context, sessionID, groupJID string, requesterJIDs []string) error
	RejectGroupRequest(ctx context.Context, sessionID, groupJID string, requesterJIDs []string) error


	GetGroupInfoFromInviteLink(ctx context.Context, sessionID, inviteLink string) (*GroupInfo, error)
	GetGroupInfoFromInvite(ctx context.Context, sessionID, groupJID, inviteCode string) (*GroupInfo, error)
}


type Service interface {

	ValidateGroupCreation(req *CreateGroupRequest) error
	ValidateGroupName(name string) error
	ValidateGroupDescription(description string) error
	ValidateParticipants(participants []string) error
	ValidateInviteLink(inviteLink string) error
	ValidateJID(jid string) error


	CanPerformAction(userJID, groupJID string, action GroupAction, groupInfo *GroupInfo) error
	IsGroupAdmin(userJID, groupJID string, groupInfo *GroupInfo) bool
	IsGroupOwner(userJID, groupJID string, groupInfo *GroupInfo) bool


	ProcessParticipantChanges(req *UpdateParticipantsRequest, currentGroup *GroupInfo) error
	ProcessSettingsChanges(req *UpdateGroupSettingsRequest, currentGroup *GroupInfo) error


	NormalizeJID(jid string) string
	ExtractPhoneNumber(jid string) string
	FormatGroupJID(groupID string) string
}


type EventHandler interface {

	OnGroupCreated(ctx context.Context, sessionID string, groupInfo *GroupInfo) error
	OnGroupInfoChanged(ctx context.Context, sessionID string, groupJID string, changes map[string]interface{}) error
	OnGroupSettingsChanged(ctx context.Context, sessionID string, groupJID string, settings *GroupSettings) error


	OnParticipantAdded(ctx context.Context, sessionID, groupJID string, participants []string, addedBy string) error
	OnParticipantRemoved(ctx context.Context, sessionID, groupJID string, participants []string, removedBy string) error
	OnParticipantPromoted(ctx context.Context, sessionID, groupJID string, participants []string, promotedBy string) error
	OnParticipantDemoted(ctx context.Context, sessionID, groupJID string, participants []string, demotedBy string) error
	OnParticipantLeft(ctx context.Context, sessionID, groupJID, participantJID string) error


	OnInviteLinkGenerated(ctx context.Context, sessionID, groupJID string, inviteLink *InviteLink) error
	OnInviteLinkRevoked(ctx context.Context, sessionID, groupJID string) error
	OnGroupJoined(ctx context.Context, sessionID string, groupInfo *GroupInfo, joinMethod string) error


	OnJoinRequestReceived(ctx context.Context, sessionID, groupJID, requesterJID string) error
	OnJoinRequestApproved(ctx context.Context, sessionID, groupJID string, approvedJIDs []string, approvedBy string) error
	OnJoinRequestRejected(ctx context.Context, sessionID, groupJID string, rejectedJIDs []string, rejectedBy string) error
}


type QRGenerator interface {
	GenerateGroupInviteQR(inviteLink string) ([]byte, error)
	GenerateGroupInfoQR(groupInfo *GroupInfo) ([]byte, error)
}


type Validator interface {
	ValidateGroupName(name string) error
	ValidateGroupDescription(description string) error
	ValidateParticipantJID(jid string) error
	ValidateInviteLink(link string) error
	ValidateGroupSettings(settings *GroupSettings) error
}




type CreateGroupRequest struct {
	Name         string   `json:"name" validate:"required,min=1,max=25"`
	Description  string   `json:"description,omitempty" validate:"max=512"`
	Participants []string `json:"participants" validate:"required,min=1,max=256"`
}


type UpdateParticipantsRequest struct {
	GroupJID     string   `json:"group_jid" validate:"required"`
	Action       string   `json:"action" validate:"required,oneof=add remove promote demote"`
	Participants []string `json:"participants" validate:"required,min=1"`
}


type UpdateGroupSettingsRequest struct {
	GroupJID         string `json:"group_jid" validate:"required"`
	Announce         *bool  `json:"announce,omitempty"`
	Restrict         *bool  `json:"restrict,omitempty"`
	JoinApprovalMode string `json:"join_approval_mode,omitempty" validate:"omitempty,oneof=auto admin_approval"`
	MemberAddMode    string `json:"member_add_mode,omitempty" validate:"omitempty,oneof=all_members only_admins"`
	Locked           *bool  `json:"locked,omitempty"`
}


type SetGroupNameRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
	Name     string `json:"name" validate:"required,min=1,max=25"`
}


type SetGroupDescriptionRequest struct {
	GroupJID    string `json:"group_jid" validate:"required"`
	Description string `json:"description" validate:"max=512"`
}


type SetGroupPhotoRequest struct {
	GroupJID  string `json:"group_jid" validate:"required"`
	PhotoData []byte `json:"photo_data" validate:"required"`
	MimeType  string `json:"mime_type" validate:"required"`
}


type GetInviteLinkRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}


type JoinGroupRequest struct {
	InviteLink string `json:"invite_link" validate:"required"`
}


type LeaveGroupRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}


type GroupRequestAction struct {
	GroupJID      string   `json:"group_jid" validate:"required"`
	RequesterJIDs []string `json:"requester_jids" validate:"required,min=1"`
	Action        string   `json:"action" validate:"required,oneof=approve reject"`
}
