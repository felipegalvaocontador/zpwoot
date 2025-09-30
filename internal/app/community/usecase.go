package community

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/community"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type UseCase interface {
	LinkGroup(ctx context.Context, sessionID string, req *LinkGroupRequest) (*LinkGroupResponse, error)

	UnlinkGroup(ctx context.Context, sessionID string, req *UnlinkGroupRequest) (*UnlinkGroupResponse, error)

	GetCommunityInfo(ctx context.Context, sessionID string, req *GetCommunityInfoRequest) (*CommunityInfoResponse, error)

	GetSubGroups(ctx context.Context, sessionID string, req *GetSubGroupsRequest) (*SubGroupsResponse, error)
}

type useCaseImpl struct {
	communityManager ports.CommunityManager
	communityService community.Service
	sessionRepo      ports.SessionRepository
	logger           logger.Logger
}

func NewUseCase(
	communityManager ports.CommunityManager,
	communityService community.Service,
	sessionRepo ports.SessionRepository,
	logger logger.Logger,
) UseCase {
	return &useCaseImpl{
		communityManager: communityManager,
		communityService: communityService,
		sessionRepo:      sessionRepo,
		logger:           logger,
	}
}

func (uc *useCaseImpl) LinkGroup(ctx context.Context, sessionID string, req *LinkGroupRequest) (*LinkGroupResponse, error) {
	err := uc.validateLinkGroupRequest(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	communityJID, groupJID, err := uc.prepareLinkGroupData(req)
	if err != nil {
		return nil, err
	}

	linkInfo, err := uc.executeLinkGroup(ctx, sessionID, communityJID, groupJID)
	if err != nil {
		return nil, err
	}

	return NewLinkGroupResponse(linkInfo), nil
}

func (uc *useCaseImpl) validateLinkGroupRequest(ctx context.Context, sessionID string, req *LinkGroupRequest) error {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid link group request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("validation failed: %w", err)
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return fmt.Errorf("session is not connected")
	}

	return nil
}

func (uc *useCaseImpl) prepareLinkGroupData(req *LinkGroupRequest) (string, string, error) {
	if err := uc.communityService.ValidateLinkRequest(req.CommunityJID, req.GroupJID); err != nil {
		uc.logger.ErrorWithFields("Invalid link request", map[string]interface{}{
			"community_jid": req.CommunityJID,
			"group_jid":     req.GroupJID,
			"error":         err.Error(),
		})
		return "", "", fmt.Errorf("invalid link request: %w", err)
	}

	communityJID := uc.communityService.FormatCommunityJID(req.CommunityJID)
	groupJID := uc.communityService.FormatGroupJID(req.GroupJID)

	return communityJID, groupJID, nil
}

func (uc *useCaseImpl) executeLinkGroup(ctx context.Context, sessionID, communityJID, groupJID string) (*community.GroupLinkInfo, error) {
	uc.logger.InfoWithFields("Linking group to community", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	err := uc.communityManager.LinkGroup(ctx, sessionID, communityJID, groupJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to link group to community", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"group_jid":     groupJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to link group to community: %w", err)
	}

	uc.logger.InfoWithFields("Group linked to community successfully", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	linkInfo := &community.GroupLinkInfo{
		CommunityJID: communityJID,
		GroupJID:     groupJID,
		Success:      true,
		Message:      "Group linked successfully",
	}

	if err := uc.communityService.ProcessGroupLinkResult(linkInfo); err != nil {
		uc.logger.ErrorWithFields("Failed to process link result", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"group_jid":     groupJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to process link result: %w", err)
	}

	return linkInfo, nil
}

func (uc *useCaseImpl) UnlinkGroup(ctx context.Context, sessionID string, req *UnlinkGroupRequest) (*UnlinkGroupResponse, error) {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid unlink group request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	if validationErr := uc.communityService.ValidateLinkRequest(req.CommunityJID, req.GroupJID); validationErr != nil {
		uc.logger.ErrorWithFields("Invalid unlink request", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": req.CommunityJID,
			"group_jid":     req.GroupJID,
			"error":         validationErr.Error(),
		})
		return nil, fmt.Errorf("invalid unlink request: %w", validationErr)
	}

	communityJID := uc.communityService.FormatCommunityJID(req.CommunityJID)
	groupJID := uc.communityService.FormatGroupJID(req.GroupJID)

	uc.logger.InfoWithFields("Unlinking group from community", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	err = uc.communityManager.UnlinkGroup(ctx, sessionID, communityJID, groupJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to unlink group from community", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"group_jid":     groupJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to unlink group from community: %w", err)
	}

	uc.logger.InfoWithFields("Group unlinked from community successfully", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	return NewUnlinkGroupResponse(communityJID, groupJID, true, "Group unlinked successfully"), nil
}

func (uc *useCaseImpl) GetCommunityInfo(ctx context.Context, sessionID string, req *GetCommunityInfoRequest) (*CommunityInfoResponse, error) {
	communityJID, err := uc.validateCommunityInfoRequest(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	communityInfo, err := uc.retrieveCommunityInfo(ctx, sessionID, communityJID)
	if err != nil {
		return nil, err
	}

	return NewCommunityInfoResponse(communityInfo), nil
}

type CommunityRequestValidator interface {
	Validate() error
	GetCommunityJID() string
}

func (uc *useCaseImpl) validateCommunityRequest(ctx context.Context, sessionID string, req CommunityRequestValidator, operationName string) (string, error) {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields(fmt.Sprintf("Invalid %s request", operationName), map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("validation failed: %w", err)
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return "", fmt.Errorf("session is not connected")
	}

	communityJID := req.GetCommunityJID()
	if err := uc.communityService.ValidateCommunityJID(communityJID); err != nil {
		uc.logger.ErrorWithFields("Invalid community JID", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return "", fmt.Errorf("invalid community JID: %w", err)
	}

	return uc.communityService.FormatCommunityJID(communityJID), nil
}

func (uc *useCaseImpl) validateCommunityInfoRequest(ctx context.Context, sessionID string, req *GetCommunityInfoRequest) (string, error) {
	return uc.validateCommunityRequest(ctx, sessionID, req, "get community info")
}

func (uc *useCaseImpl) retrieveCommunityInfo(ctx context.Context, sessionID, communityJID string) (*community.CommunityInfo, error) {
	uc.logger.InfoWithFields("Getting community info", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
	})

	communityInfo, err := uc.communityManager.GetCommunityInfo(ctx, sessionID, communityJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get community info", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to get community info: %w", err)
	}

	if err := uc.communityService.ProcessCommunityInfo(communityInfo); err != nil {
		uc.logger.ErrorWithFields("Failed to process community info", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to process community info: %w", err)
	}

	uc.logger.InfoWithFields("Community info retrieved successfully", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"name":          communityInfo.Name,
	})

	return communityInfo, nil
}

func (uc *useCaseImpl) GetSubGroups(ctx context.Context, sessionID string, req *GetSubGroupsRequest) (*SubGroupsResponse, error) {
	communityJID, err := uc.validateSubGroupsRequest(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	subGroups, err := uc.retrieveSubGroups(ctx, sessionID, communityJID)
	if err != nil {
		return nil, err
	}

	return NewSubGroupsResponse(communityJID, subGroups), nil
}

func (uc *useCaseImpl) validateSubGroupsRequest(ctx context.Context, sessionID string, req *GetSubGroupsRequest) (string, error) {
	return uc.validateCommunityRequest(ctx, sessionID, req, "get sub-groups")
}

func (uc *useCaseImpl) retrieveSubGroups(ctx context.Context, sessionID, communityJID string) ([]*community.LinkedGroup, error) {
	uc.logger.InfoWithFields("Getting community sub-groups", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
	})

	subGroups, err := uc.communityManager.GetSubGroups(ctx, sessionID, communityJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get community sub-groups", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to get community sub-groups: %w", err)
	}

	if err := uc.communityService.ProcessLinkedGroups(subGroups); err != nil {
		uc.logger.ErrorWithFields("Failed to process linked groups", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to process linked groups: %w", err)
	}

	uc.logger.InfoWithFields("Community sub-groups retrieved successfully", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"count":         len(subGroups),
	})

	return subGroups, nil
}
