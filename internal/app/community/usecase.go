package community

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/community"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// UseCase defines the interface for community use cases
type UseCase interface {
	// LinkGroup links a group to a community
	LinkGroup(ctx context.Context, sessionID string, req *LinkGroupRequest) (*LinkGroupResponse, error)

	// UnlinkGroup unlinks a group from a community
	UnlinkGroup(ctx context.Context, sessionID string, req *UnlinkGroupRequest) (*UnlinkGroupResponse, error)

	// GetCommunityInfo gets community information
	GetCommunityInfo(ctx context.Context, sessionID string, req *GetCommunityInfoRequest) (*CommunityInfoResponse, error)

	// GetSubGroups gets all sub-groups of a community
	GetSubGroups(ctx context.Context, sessionID string, req *GetSubGroupsRequest) (*SubGroupsResponse, error)
}

// useCaseImpl implements the UseCase interface
type useCaseImpl struct {
	communityManager ports.CommunityManager
	communityService community.Service
	sessionRepo      ports.SessionRepository
	logger           logger.Logger
}

// NewUseCase creates a new community use case
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

// LinkGroup links a group to a community
func (uc *useCaseImpl) LinkGroup(ctx context.Context, sessionID string, req *LinkGroupRequest) (*LinkGroupResponse, error) {
	// Validate request and session
	err := uc.validateLinkGroupRequest(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	// Format JIDs and validate link request
	communityJID, groupJID, err := uc.prepareLinkGroupData(req)
	if err != nil {
		return nil, err
	}

	// Execute the link operation
	linkInfo, err := uc.executeLinkGroup(ctx, sessionID, communityJID, groupJID)
	if err != nil {
		return nil, err
	}

	return NewLinkGroupResponse(linkInfo), nil
}

// validateLinkGroupRequest validates the request and session
func (uc *useCaseImpl) validateLinkGroupRequest(ctx context.Context, sessionID string, req *LinkGroupRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid link group request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
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

// prepareLinkGroupData validates and formats JIDs for link operation
func (uc *useCaseImpl) prepareLinkGroupData(req *LinkGroupRequest) (string, string, error) {
	// Validate link request using domain service
	if err := uc.communityService.ValidateLinkRequest(req.CommunityJID, req.GroupJID); err != nil {
		uc.logger.ErrorWithFields("Invalid link request", map[string]interface{}{
			"community_jid": req.CommunityJID,
			"group_jid":     req.GroupJID,
			"error":         err.Error(),
		})
		return "", "", fmt.Errorf("invalid link request: %w", err)
	}

	// Format JIDs
	communityJID := uc.communityService.FormatCommunityJID(req.CommunityJID)
	groupJID := uc.communityService.FormatGroupJID(req.GroupJID)

	return communityJID, groupJID, nil
}

// executeLinkGroup executes the link group operation
func (uc *useCaseImpl) executeLinkGroup(ctx context.Context, sessionID, communityJID, groupJID string) (*community.GroupLinkInfo, error) {
	uc.logger.InfoWithFields("Linking group to community", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	// Link group via community manager
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

	// Create and process link info
	linkInfo := &community.GroupLinkInfo{
		CommunityJID: communityJID,
		GroupJID:     groupJID,
		Success:      true,
		Message:      "Group linked successfully",
	}

	// Process the result
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

// UnlinkGroup unlinks a group from a community
func (uc *useCaseImpl) UnlinkGroup(ctx context.Context, sessionID string, req *UnlinkGroupRequest) (*UnlinkGroupResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid unlink group request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
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

	// Validate unlink request using domain service
	if err := uc.communityService.ValidateLinkRequest(req.CommunityJID, req.GroupJID); err != nil {
		uc.logger.ErrorWithFields("Invalid unlink request", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": req.CommunityJID,
			"group_jid":     req.GroupJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("invalid unlink request: %w", err)
	}

	// Format JIDs
	communityJID := uc.communityService.FormatCommunityJID(req.CommunityJID)
	groupJID := uc.communityService.FormatGroupJID(req.GroupJID)

	uc.logger.InfoWithFields("Unlinking group from community", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	// Unlink group via community manager
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

// GetCommunityInfo gets community information
func (uc *useCaseImpl) GetCommunityInfo(ctx context.Context, sessionID string, req *GetCommunityInfoRequest) (*CommunityInfoResponse, error) {
	// Validate request and session
	communityJID, err := uc.validateCommunityInfoRequest(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	// Retrieve and process community info
	communityInfo, err := uc.retrieveCommunityInfo(ctx, sessionID, communityJID)
	if err != nil {
		return nil, err
	}

	return NewCommunityInfoResponse(communityInfo), nil
}

// CommunityRequestValidator interface for validating community requests
type CommunityRequestValidator interface {
	Validate() error
	GetCommunityJID() string
}

// validateCommunityRequest validates the request and returns formatted community JID
func (uc *useCaseImpl) validateCommunityRequest(ctx context.Context, sessionID string, req CommunityRequestValidator, operationName string) (string, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields(fmt.Sprintf("Invalid %s request", operationName), map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Validate session
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

	// Validate community JID
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

// validateCommunityInfoRequest validates the request and returns formatted community JID
func (uc *useCaseImpl) validateCommunityInfoRequest(ctx context.Context, sessionID string, req *GetCommunityInfoRequest) (string, error) {
	return uc.validateCommunityRequest(ctx, sessionID, req, "get community info")
}

// retrieveCommunityInfo retrieves and processes community information
func (uc *useCaseImpl) retrieveCommunityInfo(ctx context.Context, sessionID, communityJID string) (*community.CommunityInfo, error) {
	uc.logger.InfoWithFields("Getting community info", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
	})

	// Get community info via community manager
	communityInfo, err := uc.communityManager.GetCommunityInfo(ctx, sessionID, communityJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get community info", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to get community info: %w", err)
	}

	// Process community info
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

// GetSubGroups gets all sub-groups of a community
func (uc *useCaseImpl) GetSubGroups(ctx context.Context, sessionID string, req *GetSubGroupsRequest) (*SubGroupsResponse, error) {
	// Validate request and session
	communityJID, err := uc.validateSubGroupsRequest(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	// Retrieve and process sub-groups
	subGroups, err := uc.retrieveSubGroups(ctx, sessionID, communityJID)
	if err != nil {
		return nil, err
	}

	return NewSubGroupsResponse(communityJID, subGroups), nil
}

// validateSubGroupsRequest validates the request and returns formatted community JID
func (uc *useCaseImpl) validateSubGroupsRequest(ctx context.Context, sessionID string, req *GetSubGroupsRequest) (string, error) {
	return uc.validateCommunityRequest(ctx, sessionID, req, "get sub-groups")
}

// retrieveSubGroups retrieves and processes community sub-groups
func (uc *useCaseImpl) retrieveSubGroups(ctx context.Context, sessionID, communityJID string) ([]*community.LinkedGroup, error) {
	uc.logger.InfoWithFields("Getting community sub-groups", map[string]interface{}{
		"session_id":    sessionID,
		"community_jid": communityJID,
	})

	// Get sub-groups via community manager
	subGroups, err := uc.communityManager.GetSubGroups(ctx, sessionID, communityJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get community sub-groups", map[string]interface{}{
			"session_id":    sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to get community sub-groups: %w", err)
	}

	// Process linked groups
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
