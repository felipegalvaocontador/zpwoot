package services

import (
	"context"
	"fmt"

	"zpwoot/internal/core/group"
	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/services/shared/validation"
	"zpwoot/platform/logger"
)


type GroupService struct {
	groupCore       group.Service
	groupRepo       group.Repository
	whatsappGateway group.WhatsAppGateway
	logger          *logger.Logger
	validator       *validation.Validator
}


func NewGroupService(
	groupCore group.Service,
	groupRepo group.Repository,
	whatsappGateway group.WhatsAppGateway,
	logger *logger.Logger,
	validator *validation.Validator,
) *GroupService {
	return &GroupService{
		groupCore:       groupCore,
		groupRepo:       groupRepo,
		whatsappGateway: whatsappGateway,
		logger:          logger,
		validator:       validator,
	}
}


func (s *GroupService) CreateGroup(ctx context.Context, sessionID string, req *contracts.CreateGroupRequest) (*contracts.CreateGroupResponse, error) {
	s.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id":      sessionID,
		"group_name":      req.Name,
		"participants":    len(req.Participants),
		"has_description": req.Description != "",
	})


	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}


	domainReq := &group.CreateGroupRequest{
		Name:         req.Name,
		Description:  req.Description,
		Participants: req.Participants,
	}


	if err := s.groupCore.ValidateGroupCreation(domainReq); err != nil {
		return nil, fmt.Errorf("group validation failed: %w", err)
	}


	groupInfo, err := s.whatsappGateway.CreateGroup(ctx, sessionID, req.Name, req.Participants, req.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create group in WhatsApp: %w", err)
	}


	groupModel := s.convertGroupInfoToModel(groupInfo, sessionID)
	if err := s.groupRepo.Create(ctx, groupModel); err != nil {
		s.logger.ErrorWithFields("Failed to save group to database", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupInfo.GroupJID,
			"error":      err.Error(),
		})

	}

	response := &contracts.CreateGroupResponse{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Participants: req.Participants,
		CreatedAt:    groupInfo.CreatedAt,
		Success:      true,
		Message:      "Group created successfully",
	}

	s.logger.InfoWithFields("Group created successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupInfo.GroupJID,
		"group_name": groupInfo.Name,
	})

	return response, nil
}


func (s *GroupService) ListGroups(ctx context.Context, sessionID string) (*contracts.ListGroupsResponse, error) {
	s.logger.InfoWithFields("Listing groups", map[string]interface{}{
		"session_id": sessionID,
	})


	groupInfos, err := s.whatsappGateway.ListJoinedGroups(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups from WhatsApp: %w", err)
	}


	groups := make([]contracts.GroupInfo, len(groupInfos))
	for i, groupInfo := range groupInfos {
		groups[i] = contracts.GroupInfo{
			GroupJID:     groupInfo.GroupJID,
			Name:         groupInfo.Name,
			Description:  groupInfo.Description,
			Owner:        groupInfo.Owner,
			Participants: len(groupInfo.Participants),
			CreatedAt:    groupInfo.CreatedAt,
		}
	}

	response := &contracts.ListGroupsResponse{
		Groups:  groups,
		Count:   len(groups),
		Success: true,
		Message: "Groups retrieved successfully",
	}

	s.logger.InfoWithFields("Groups listed successfully", map[string]interface{}{
		"session_id":   sessionID,
		"group_count":  len(groups),
	})

	return response, nil
}


func (s *GroupService) GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*contracts.GetGroupInfoResponse, error) {
	s.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})


	groupInfo, err := s.whatsappGateway.GetGroupInfo(ctx, sessionID, groupJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info from WhatsApp: %w", err)
	}


	participants := make([]contracts.ParticipantInfo, len(groupInfo.Participants))
	for i, p := range groupInfo.Participants {
		participants[i] = contracts.ParticipantInfo{
			JID:      p.JID,
			Role:     string(p.Role),
			JoinedAt: p.JoinedAt,
			Status:   string(p.Status),
		}
	}

	response := &contracts.GetGroupInfoResponse{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Owner:        groupInfo.Owner,
		Participants: participants,
		Settings: contracts.GroupSettings{
			Announce:         groupInfo.Settings.Announce,
			Restrict:         groupInfo.Settings.Restrict,
			JoinApprovalMode: groupInfo.Settings.JoinApprovalMode,
			MemberAddMode:    groupInfo.Settings.MemberAddMode,
			Locked:           groupInfo.Settings.Locked,
		},
		CreatedAt: groupInfo.CreatedAt,
		UpdatedAt: groupInfo.UpdatedAt,
		Success:   true,
		Message:   "Group info retrieved successfully",
	}

	s.logger.InfoWithFields("Group info retrieved successfully", map[string]interface{}{
		"session_id":      sessionID,
		"group_jid":       groupJID,
		"group_name":      groupInfo.Name,
		"participant_count": len(participants),
	})

	return response, nil
}


func (s *GroupService) UpdateGroupParticipants(ctx context.Context, sessionID string, req *contracts.UpdateParticipantsRequest) (*contracts.UpdateParticipantsResponse, error) {
	s.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})


	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}


	groupInfo, err := s.whatsappGateway.GetGroupInfo(ctx, sessionID, req.GroupJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}


	domainReq := &group.UpdateParticipantsRequest{
		GroupJID:     req.GroupJID,
		Action:       req.Action,
		Participants: req.Participants,
	}

	if err := s.groupCore.ProcessParticipantChanges(domainReq, groupInfo); err != nil {
		return nil, fmt.Errorf("participant changes validation failed: %w", err)
	}


	switch req.Action {
	case "add":
		err = s.whatsappGateway.AddParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	case "remove":
		err = s.whatsappGateway.RemoveParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	case "promote":
		err = s.whatsappGateway.PromoteParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	case "demote":
		err = s.whatsappGateway.DemoteParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	default:
		return nil, fmt.Errorf("invalid action: %s", req.Action)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to %s participants: %w", req.Action, err)
	}

	response := &contracts.UpdateParticipantsResponse{
		GroupJID:     req.GroupJID,
		Action:       req.Action,
		Participants: req.Participants,
		Success:      true,
		Message:      fmt.Sprintf("Participants %s successfully", req.Action),
	}

	s.logger.InfoWithFields("Group participants updated successfully", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	return response, nil
}


func (s *GroupService) SetGroupName(ctx context.Context, sessionID string, req *contracts.SetGroupNameRequest) (*contracts.SetGroupNameResponse, error) {
	s.logger.InfoWithFields("Setting group name", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"new_name":   req.Name,
	})


	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}


	if err := s.groupCore.ValidateGroupName(req.Name); err != nil {
		return nil, fmt.Errorf("group name validation failed: %w", err)
	}


	if err := s.whatsappGateway.SetGroupName(ctx, sessionID, req.GroupJID, req.Name); err != nil {
		return nil, fmt.Errorf("failed to set group name: %w", err)
	}

	response := &contracts.SetGroupNameResponse{
		GroupJID: req.GroupJID,
		Name:     req.Name,
		Success:  true,
		Message:  "Group name updated successfully",
	}

	s.logger.InfoWithFields("Group name updated successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"new_name":   req.Name,
	})

	return response, nil
}


func (s *GroupService) convertGroupInfoToModel(groupInfo *group.GroupInfo, sessionID string) *group.Group {


	return &group.Group{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Owner:        groupInfo.Owner,
		Settings:     groupInfo.Settings,
		Participants: groupInfo.Participants,
		CreatedAt:    groupInfo.CreatedAt,
		UpdatedAt:    groupInfo.UpdatedAt,
	}
}