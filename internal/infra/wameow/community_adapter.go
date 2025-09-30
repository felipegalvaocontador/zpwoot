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

type CommunityAdapter struct {
	wameowManager ports.WameowManager
	logger        logger.Logger
}

func NewCommunityAdapter(wameowManager ports.WameowManager, logger logger.Logger) *CommunityAdapter {
	return &CommunityAdapter{
		wameowManager: wameowManager,
		logger:        logger,
	}
}

func (ca *CommunityAdapter) LinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error {
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

func (ca *CommunityAdapter) UnlinkGroup(ctx context.Context, sessionID string, communityJID, groupJID string) error {
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

func (ca *CommunityAdapter) GetCommunityInfo(ctx context.Context, sessionID string, communityJID string) (*community.CommunityInfo, error) {
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	groupInfo, err := client.GetGroupInfo(ctx, communityJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get community info: %w", err)
	}

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

func (ca *CommunityAdapter) GetSubGroups(ctx context.Context, sessionID string, communityJID string) ([]*community.LinkedGroup, error) {
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	subGroupsInfo, err := client.GetSubGroups(ctx, communityJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-groups: %w", err)
	}

	linkedGroups := make([]*community.LinkedGroup, len(subGroupsInfo))
	for i, groupTarget := range subGroupsInfo {
		linkedGroups[i] = convertToLinkedGroupFromTarget(groupTarget)
	}

	return linkedGroups, nil
}

func extractIDFromJID(jid string) string {
	if len(jid) > 5 && jid[len(jid)-5:] == "@g.us" {
		return jid[:len(jid)-5]
	}
	return jid
}

func convertToLinkedGroupFromTarget(groupTarget *types.GroupLinkTarget) *community.LinkedGroup {
	return &community.LinkedGroup{
		JID:              groupTarget.JID.String(),
		Name:             groupTarget.Name,
		Description:      "",         // GroupLinkTarget doesn't have description
		ParticipantCount: 0,          // GroupLinkTarget doesn't have participant count
		IsOwner:          false,      // Would need to be determined from user's role
		IsAdmin:          false,      // Would need to be determined from user's role
		LinkedAt:         time.Now(), // Use current time since GroupLinkTarget doesn't have creation time
	}
}

func (ca *CommunityAdapter) GetLinkedGroupsParticipants(ctx context.Context, sessionID string, communityJID string) ([]string, error) {
	manager, ok := ca.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	participants, err := client.GetLinkedGroupsParticipants(ctx, communityJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked groups participants: %w", err)
	}

	participantJIDs := make([]string, len(participants))
	for i, jid := range participants {
		participantJIDs[i] = jid.String()
	}

	return participantJIDs, nil
}
