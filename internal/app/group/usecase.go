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

	// Advanced group methods
	GetGroupInfoFromLink(ctx context.Context, sessionID string, req *GetGroupInfoFromLinkRequest) (*GroupInfoFromLinkResponse, error)
	GetGroupInfoFromInvite(ctx context.Context, sessionID string, req *GetGroupInfoFromInviteRequest) (*GroupInfoFromInviteResponse, error)
	JoinGroupWithInvite(ctx context.Context, sessionID string, req *JoinGroupWithInviteRequest) (*JoinGroupWithInviteResponse, error)
}

type useCaseImpl struct {
	wameowMgr    ports.WameowManager
	groupService *group.Service
}

func NewUseCase(
	groupRepo ports.GroupRepository, // Kept for interface compatibility but not used
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

	// Validate through domain service
	if err := uc.groupService.ValidateGroupCreation(domainReq); err != nil {
		return nil, err
	}

	// Create group via wameow manager
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
	// Get group info via wameow manager
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
	// List groups via wameow manager
	groups, err := uc.wameowMgr.ListJoinedGroups(sessionID)
	if err != nil {
		return nil, err
	}

	var groupList []GroupInfo
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

	// Validate through domain service
	if err := uc.groupService.ValidateParticipantUpdate(domainReq); err != nil {
		return nil, err
	}

	// Update participants via wameow manager
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
	// Validate through domain service
	if err := uc.groupService.ValidateGroupName(req.Name); err != nil {
		return nil, err
	}

	// Set group name via wameow manager
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
	// Validate through domain service
	if err := uc.groupService.ValidateGroupDescription(req.Description); err != nil {
		return nil, err
	}

	// Set group description via wameow manager
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
	// Validate photo data (basic validation)
	if req.Photo == "" {
		return nil, group.ErrInvalidGroupJID // Use appropriate error
	}

	// Convert base64 photo to bytes (simplified)
	photoBytes := []byte(req.Photo) // In real implementation, decode base64

	// Set group photo via wameow manager
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
	// Get invite link via wameow manager
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
	// Validate invite link
	if err := uc.groupService.ValidateInviteLink(req.InviteLink); err != nil {
		return nil, err
	}

	// Join group via wameow manager
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
	// Leave group via wameow manager
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
	// Update group settings via wameow manager
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

// Helper functions for conversion
func convertParticipants(participants []ports.GroupParticipant) []GroupParticipant {
	var result []GroupParticipant
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

// isUserAdmin checks if the current user is an admin of the group
func (uc *useCaseImpl) isUserAdmin(group *ports.GroupInfo, sessionID string) bool {
	// Get current user's JID from session
	userJID, err := uc.wameowMgr.GetUserJID(sessionID)
	if err != nil {
		return false
	}

	// Check if user is in the participants list as admin
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

	// Convert to interface{} slice for JSON response
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

// ============================================================================
// ADVANCED GROUP METHODS
// ============================================================================

// GetGroupInfoFromLink gets group information from an invite link
func (uc *useCaseImpl) GetGroupInfoFromLink(ctx context.Context, sessionID string, req *GetGroupInfoFromLinkRequest) (*GroupInfoFromLinkResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get group info from link via wameow manager
	groupInfo, err := uc.wameowMgr.GetGroupInfoFromLink(sessionID, req.InviteLink)
	if err != nil {
		return nil, err
	}

	return NewGroupInfoFromLinkResponse(groupInfo), nil
}

// GetGroupInfoFromInvite gets group information from an invite
func (uc *useCaseImpl) GetGroupInfoFromInvite(ctx context.Context, sessionID string, req *GetGroupInfoFromInviteRequest) (*GroupInfoFromInviteResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get group info from invite via wameow manager
	groupInfo, err := uc.wameowMgr.GetGroupInfoFromInvite(sessionID, req.GroupJID, req.Inviter, req.Code, req.Expiration)
	if err != nil {
		return nil, err
	}

	return NewGroupInfoFromInviteResponse(groupInfo, req.Code, req.Inviter), nil
}

// JoinGroupWithInvite joins a group using a specific invite
func (uc *useCaseImpl) JoinGroupWithInvite(ctx context.Context, sessionID string, req *JoinGroupWithInviteRequest) (*JoinGroupWithInviteResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Join group with invite via wameow manager
	err := uc.wameowMgr.JoinGroupWithInvite(sessionID, req.GroupJID, req.Inviter, req.Code, req.Expiration)
	if err != nil {
		return NewJoinGroupWithInviteResponse(req.GroupJID, false, "Failed to join group: "+err.Error()), nil
	}

	return NewJoinGroupWithInviteResponse(req.GroupJID, true, "Successfully joined group"), nil
}
