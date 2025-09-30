package wameow

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/types"
	"zpwoot/internal/domain/community"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// CommunityAdapter implements the CommunityManager interface
type CommunityAdapter struct {
	wameowManager ports.WameowManager
	logger        logger.Logger
}

// NewCommunityAdapter creates a new community adapter
func NewCommunityAdapter(wameowManager ports.WameowManager, logger logger.Logger) *CommunityAdapter {
	return &CommunityAdapter{
		wameowManager: wameowManager,
		logger:        logger,
	}
}

// LinkGroup links a group to a community
func (ca *CommunityAdapter) LinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error {
	// Cast to Manager to access internal client
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.LinkGroup(ctx, communityJID, groupJID)
}

// UnlinkGroup unlinks a group from a community
func (ca *CommunityAdapter) UnlinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error {
	// Cast to Manager to access internal client
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.UnlinkGroup(ctx, communityJID, groupJID)
}

// GetCommunityInfo gets information about a community
func (ca *CommunityAdapter) GetCommunityInfo(ctx context.Context, sessionID string, communityJID string) (*community.CommunityInfo, error) {
	// Cast to Manager to access internal client
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// For now, we'll simulate getting community info since whatsmeow might not have direct community info method
	// This would need to be implemented based on actual whatsmeow capabilities

	// Try to get group info as communities are essentially special groups
	groupInfo, err := client.GetGroupInfo(ctx, communityJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get community info: %w", err)
	}

	// Convert group info to community info
	communityInfo := &community.CommunityInfo{
		ID:               extractIDFromJID(communityJID),
		JID:              communityJID,
		Name:             groupInfo.Name,
		Description:      groupInfo.Topic,
		ParticipantCount: len(groupInfo.Participants),
		GroupCount:       0,     // Would need to be calculated from linked groups
		IsOwner:          false, // Would need to be determined from user's role
		IsAdmin:          false, // Would need to be determined from user's role
		IsMuted:          false, // Would need to be determined from user's settings
		IsAnnouncement:   groupInfo.IsAnnounce,
	}

	return communityInfo, nil
}

// GetSubGroups gets all sub-groups (linked groups) of a community
func (ca *CommunityAdapter) GetSubGroups(ctx context.Context, sessionID string, communityJID string) ([]*community.LinkedGroup, error) {
	// Cast to Manager to access internal client
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Get sub-groups using whatsmeow
	subGroupsInfo, err := client.GetSubGroups(ctx, communityJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-groups: %w", err)
	}

	// Convert whatsmeow group link targets to domain linked groups
	linkedGroups := make([]*community.LinkedGroup, len(subGroupsInfo))
	for i, groupTarget := range subGroupsInfo {
		linkedGroups[i] = convertToLinkedGroupFromTarget(groupTarget)
	}

	return linkedGroups, nil
}

// Helper functions

// extractIDFromJID extracts the ID part from a JID
func extractIDFromJID(jid string) string {
	if len(jid) > 5 && jid[len(jid)-5:] == "@g.us" {
		return jid[:len(jid)-5]
	}
	return jid
}

// convertToLinkedGroupFromTarget converts whatsmeow GroupLinkTarget to domain LinkedGroup
func convertToLinkedGroupFromTarget(groupTarget *types.GroupLinkTarget) *community.LinkedGroup {
	return &community.LinkedGroup{
		JID:              groupTarget.JID.String(),
		Name:             groupTarget.GroupName.Name,
		Description:      "",         // GroupLinkTarget doesn't have description
		ParticipantCount: 0,          // GroupLinkTarget doesn't have participant count
		IsOwner:          false,      // Would need to be determined from user's role
		IsAdmin:          false,      // Would need to be determined from user's role
		LinkedAt:         time.Now(), // Use current time since GroupLinkTarget doesn't have creation time
	}
}

// GetLinkedGroupsParticipants gets participants from all linked groups in a community
func (ca *CommunityAdapter) GetLinkedGroupsParticipants(ctx context.Context, sessionID string, communityJID string) ([]string, error) {
	// Cast to Manager to access internal client
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Get linked groups participants using whatsmeow
	participants, err := client.GetLinkedGroupsParticipants(ctx, communityJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked groups participants: %w", err)
	}

	// Convert JIDs to strings
	participantJIDs := make([]string, len(participants))
	for i, jid := range participants {
		participantJIDs[i] = jid.String()
	}

	return participantJIDs, nil
}
