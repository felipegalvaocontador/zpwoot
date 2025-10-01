package group

import (
	"context"
	"time"

	"zpwoot/internal/domain/group"
	"zpwoot/internal/ports"
)

type UseCase interface {
	CreateGroup(ctx context.Context, sessionID string, req *CreateGroupRequest) (*CreateGroupResponse, error)
	GetGroupInfo(ctx context.Context, sessionID string, req *GetGroupInfoRequest) (*GetGroupInfoResponse, error)
	ListGroups(ctx context.Context, sessionID string) (*ListGroupsResponse, error)
	UpdateGroupParticipants(ctx context.Context, sessionID string, req *UpdateGroupParticipantsRequest) (*UpdateGroupParticipantsResponse, error)
	SetGroupName(ctx context.Context, sessionID string, req *SetGroupNameRequest) (*GroupActionResponse, error)
	SetGroupDescription(ctx context.Context, sessionID string, req *SetGroupDescriptionRequest) (*GroupActionResponse, error)
	SetGroupPhoto(ctx context.Context, sessionID string, req *SetGroupPhotoRequest) (*GroupActionResponse, error)
	GetGroupInviteLink(ctx context.Context, sessionID string, req *GetGroupInviteLinkRequest) (*GetGroupInviteLinkResponse, error)
	JoinGroup(ctx context.Context, sessionID string, req *JoinGroupRequest) (*JoinGroupResponse, error)
	LeaveGroup(ctx context.Context, sessionID string, req *LeaveGroupRequest) (*LeaveGroupResponse, error)
	UpdateGroupSettings(ctx context.Context, sessionID string, req *UpdateGroupSettingsRequest) (*GroupActionResponse, error)
	GetGroupRequestParticipants(ctx context.Context, sessionID string, groupJID string) ([]interface{}, error)
	UpdateGroupRequestParticipants(ctx context.Context, sessionID string, groupJID string, participants []string, action string) ([]string, []string, error)
	SetGroupJoinApprovalMode(ctx context.Context, sessionID string, groupJID string, requireApproval bool) error
	SetGroupMemberAddMode(ctx context.Context, sessionID string, groupJID string, mode string) error

	GetGroupInfoFromLink(ctx context.Context, sessionID string, req *GetGroupInfoFromLinkRequest) (*GroupInfoFromLinkResponse, error)
	GetGroupInfoFromInvite(ctx context.Context, sessionID string, req *GetGroupInfoFromInviteRequest) (*GroupInfoFromInviteResponse, error)
	JoinGroupWithInvite(ctx context.Context, sessionID string, req *JoinGroupWithInviteRequest) (*JoinGroupWithInviteResponse, error)
}

type useCaseImpl struct {
	wameowMgr    ports.WameowManager
	groupService *group.Service
}

func NewUseCase(
	groupRepo ports.GroupRepository,
	wameowMgr ports.WameowManager,
	groupService *group.Service,
) UseCase {
	return &useCaseImpl{
		wameowMgr:    wameowMgr,
		groupService: groupService,
	}
}

func (uc *useCaseImpl) CreateGroup(ctx context.Context, sessionID string, req *CreateGroupRequest) (*CreateGroupResponse, error) {
	domainReq := req.ToDomain()

	if err := uc.groupService.ValidateGroupCreation(domainReq); err != nil {
		return nil, err
	}

	groupInfo, err := uc.wameowMgr.CreateGroup(sessionID, domainReq.Name, domainReq.Participants, domainReq.Description)
	if err != nil {
		return nil, err
	}

	return &CreateGroupResponse{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Participants: domainReq.Participants,
		CreatedAt:    groupInfo.CreatedAt,
	}, nil
}

func (uc *useCaseImpl) GetGroupInfo(ctx context.Context, sessionID string, req *GetGroupInfoRequest) (*GetGroupInfoResponse, error) {
	groupInfo, err := uc.wameowMgr.GetGroupInfo(sessionID, req.GroupJID)
	if err != nil {
		return nil, err
	}

	return &GetGroupInfoResponse{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Owner:        groupInfo.Owner,
		Participants: convertParticipants(groupInfo.Participants),
		Settings:     convertSettings(groupInfo.Settings),
		CreatedAt:    groupInfo.CreatedAt,
		UpdatedAt:    groupInfo.UpdatedAt,
	}, nil
}

func (uc *useCaseImpl) ListGroups(ctx context.Context, sessionID string) (*ListGroupsResponse, error) {
	groups, err := uc.wameowMgr.ListJoinedGroups(sessionID)
	if err != nil {
		return nil, err
	}

	groupList := make([]GroupInfo, 0, len(groups))
	for _, group := range groups {
		groupList = append(groupList, GroupInfo{
			GroupJID:         group.GroupJID,
			Name:             group.Name,
			Description:      group.Description,
			ParticipantCount: len(group.Participants),
			IsAdmin:          uc.isUserAdmin(group, sessionID),
			CreatedAt:        group.CreatedAt,
		})
	}

	return &ListGroupsResponse{
		Groups: groupList,
		Total:  len(groups),
	}, nil
}

func (uc *useCaseImpl) UpdateGroupParticipants(ctx context.Context, sessionID string, req *UpdateGroupParticipantsRequest) (*UpdateGroupParticipantsResponse, error) {
	domainReq := &group.UpdateParticipantsRequest{
		GroupJID:     req.GroupJID,
		Participants: req.Participants,
		Action:       req.Action,
	}

	if err := uc.groupService.ValidateParticipantUpdate(domainReq); err != nil {
		return nil, err
	}

	success, failed, err := uc.wameowMgr.UpdateGroupParticipants(sessionID, req.GroupJID, req.Participants, req.Action)
	if err != nil {
		return nil, err
	}

	return &UpdateGroupParticipantsResponse{
		GroupJID:     req.GroupJID,
		Participants: req.Participants,
		Action:       req.Action,
		Success:      success,
		Failed:       failed,
	}, nil
}

func (uc *useCaseImpl) SetGroupName(ctx context.Context, sessionID string, req *SetGroupNameRequest) (*GroupActionResponse, error) {
	if err := uc.groupService.ValidateGroupName(req.Name); err != nil {
		return nil, err
	}

	err := uc.wameowMgr.SetGroupName(sessionID, req.GroupJID, req.Name)
	if err != nil {
		return nil, err
	}

	return &GroupActionResponse{
		GroupJID:  req.GroupJID,
		Status:    "success",
		Message:   "Group name updated successfully",
		Timestamp: time.Now(),
	}, nil
}

func (uc *useCaseImpl) SetGroupDescription(ctx context.Context, sessionID string, req *SetGroupDescriptionRequest) (*GroupActionResponse, error) {
	if err := uc.groupService.ValidateGroupDescription(req.Description); err != nil {
		return nil, err
	}

	err := uc.wameowMgr.SetGroupDescription(sessionID, req.GroupJID, req.Description)
	if err != nil {
		return nil, err
	}

	return &GroupActionResponse{
		GroupJID:  req.GroupJID,
		Status:    "success",
		Message:   "Group description updated successfully",
		Timestamp: time.Now(),
	}, nil
}

func (uc *useCaseImpl) SetGroupPhoto(ctx context.Context, sessionID string, req *SetGroupPhotoRequest) (*GroupActionResponse, error) {
	if req.Photo == "" {
		return nil, group.ErrInvalidGroupJID
	}

	photoBytes := []byte(req.Photo)

	err := uc.wameowMgr.SetGroupPhoto(sessionID, req.GroupJID, photoBytes)
	if err != nil {
		return nil, err
	}

	return &GroupActionResponse{
		GroupJID:  req.GroupJID,
		Status:    "success",
		Message:   "Group photo updated successfully",
		Timestamp: time.Now(),
	}, nil
}

func (uc *useCaseImpl) GetGroupInviteLink(ctx context.Context, sessionID string, req *GetGroupInviteLinkRequest) (*GetGroupInviteLinkResponse, error) {
	inviteLink, err := uc.wameowMgr.GetGroupInviteLink(sessionID, req.GroupJID, req.Reset)
	if err != nil {
		return nil, err
	}

	return &GetGroupInviteLinkResponse{
		GroupJID:   req.GroupJID,
		InviteLink: inviteLink,
	}, nil
}

func (uc *useCaseImpl) JoinGroup(ctx context.Context, sessionID string, req *JoinGroupRequest) (*JoinGroupResponse, error) {
	if err := uc.groupService.ValidateInviteLink(req.InviteLink); err != nil {
		return nil, err
	}

	groupInfo, err := uc.wameowMgr.JoinGroupViaLink(sessionID, req.InviteLink)
	if err != nil {
		return nil, err
	}

	return &JoinGroupResponse{
		GroupJID: groupInfo.GroupJID,
		Name:     groupInfo.Name,
		Status:   "joined",
	}, nil
}

func (uc *useCaseImpl) LeaveGroup(ctx context.Context, sessionID string, req *LeaveGroupRequest) (*LeaveGroupResponse, error) {
	err := uc.wameowMgr.LeaveGroup(sessionID, req.GroupJID)
	if err != nil {
		return nil, err
	}

	return &LeaveGroupResponse{
		GroupJID: req.GroupJID,
		Status:   "left",
	}, nil
}

func (uc *useCaseImpl) UpdateGroupSettings(ctx context.Context, sessionID string, req *UpdateGroupSettingsRequest) (*GroupActionResponse, error) {
	err := uc.wameowMgr.UpdateGroupSettings(sessionID, req.GroupJID, req.Announce, req.Locked)
	if err != nil {
		return nil, err
	}

	return &GroupActionResponse{
		GroupJID:  req.GroupJID,
		Status:    "success",
		Message:   "Group settings updated successfully",
		Timestamp: time.Now(),
	}, nil
}

func convertParticipants(participants []ports.GroupParticipant) []GroupParticipant {
	result := make([]GroupParticipant, 0, len(participants))
	for _, p := range participants {
		result = append(result, GroupParticipant{
			JID:          p.JID,
			IsAdmin:      p.IsAdmin,
			IsSuperAdmin: p.IsSuperAdmin,
		})
	}
	return result
}

func convertSettings(settings ports.GroupSettings) GroupSettings {
	return GroupSettings{
		Announce: settings.Announce,
		Locked:   settings.Locked,
	}
}

func (uc *useCaseImpl) isUserAdmin(group *ports.GroupInfo, sessionID string) bool {
	userJID, err := uc.wameowMgr.GetUserJID(sessionID)
	if err != nil {
		return false
	}

	for _, participant := range group.Participants {
		if participant.JID == userJID && (participant.IsAdmin || participant.IsSuperAdmin) {
			return true
		}
	}

	return false
}

func (uc *useCaseImpl) GetGroupRequestParticipants(ctx context.Context, sessionID string, groupJID string) ([]interface{}, error) {
	participants, err := uc.wameowMgr.GetGroupRequestParticipants(sessionID, groupJID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(participants))
	for i, p := range participants {
		result[i] = map[string]interface{}{
			"jid":          p.JID.String(),
			"requested_at": p.RequestedAt.Format(time.RFC3339),
		}
	}

	return result, nil
}

func (uc *useCaseImpl) UpdateGroupRequestParticipants(ctx context.Context, sessionID string, groupJID string, participants []string, action string) ([]string, []string, error) {
	return uc.wameowMgr.UpdateGroupRequestParticipants(sessionID, groupJID, participants, action)
}

func (uc *useCaseImpl) SetGroupJoinApprovalMode(ctx context.Context, sessionID string, groupJID string, requireApproval bool) error {
	return uc.wameowMgr.SetGroupJoinApprovalMode(sessionID, groupJID, requireApproval)
}

func (uc *useCaseImpl) SetGroupMemberAddMode(ctx context.Context, sessionID string, groupJID string, mode string) error {
	return uc.wameowMgr.SetGroupMemberAddMode(sessionID, groupJID, mode)
}

func (uc *useCaseImpl) GetGroupInfoFromLink(ctx context.Context, sessionID string, req *GetGroupInfoFromLinkRequest) (*GroupInfoFromLinkResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	groupInfo, err := uc.wameowMgr.GetGroupInfoFromLink(sessionID, req.InviteLink)
	if err != nil {
		return nil, err
	}

	return NewGroupInfoFromLinkResponse(groupInfo), nil
}

func (uc *useCaseImpl) GetGroupInfoFromInvite(ctx context.Context, sessionID string, req *GetGroupInfoFromInviteRequest) (*GroupInfoFromInviteResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	groupInfo, err := uc.wameowMgr.GetGroupInfoFromInvite(sessionID, req.GroupJID, req.Inviter, req.Code, req.Expiration)
	if err != nil {
		return nil, err
	}

	return NewGroupInfoFromInviteResponse(groupInfo, req.Code, req.Inviter), nil
}

func (uc *useCaseImpl) JoinGroupWithInvite(ctx context.Context, sessionID string, req *JoinGroupWithInviteRequest) (*JoinGroupWithInviteResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := uc.wameowMgr.JoinGroupWithInvite(sessionID, req.GroupJID, req.Inviter, req.Code, req.Expiration)
	if err != nil {
		return NewJoinGroupWithInviteResponse(req.GroupJID, false, "Failed to join group: "+err.Error()), err
	}

	return NewJoinGroupWithInviteResponse(req.GroupJID, true, "Successfully joined group"), nil
}
