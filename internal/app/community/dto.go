package community

import (
	"fmt"
	"time"

	"zpwoot/internal/domain/community"
)

// LinkGroupRequest represents the request for linking a group to a community
type LinkGroupRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
	GroupJID     string `json:"groupJid" validate:"required"`
}

// Validate validates the link group request
func (r *LinkGroupRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

// UnlinkGroupRequest represents the request for unlinking a group from a community
type UnlinkGroupRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
	GroupJID     string `json:"groupJid" validate:"required"`
}

// Validate validates the unlink group request
func (r *UnlinkGroupRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

// GetCommunityInfoRequest represents the request for getting community information
type GetCommunityInfoRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
}

// Validate validates the get community info request
func (r *GetCommunityInfoRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	return nil
}

// GetCommunityJID returns the community JID
func (r *GetCommunityInfoRequest) GetCommunityJID() string {
	return r.CommunityJID
}

// GetSubGroupsRequest represents the request for getting community sub-groups
type GetSubGroupsRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
}

// Validate validates the get sub-groups request
func (r *GetSubGroupsRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	return nil
}

// GetCommunityJID returns the community JID
func (r *GetSubGroupsRequest) GetCommunityJID() string {
	return r.CommunityJID
}

// LinkGroupResponse represents the response for linking a group to a community
type LinkGroupResponse struct {
	LinkedAt     time.Time `json:"linkedAt"`
	Message      string    `json:"message"`
	CommunityJID string    `json:"communityJid"`
	GroupJID     string    `json:"groupJid"`
	Success      bool      `json:"success"`
}

// NewLinkGroupResponse creates a new link group response
func NewLinkGroupResponse(linkInfo *community.GroupLinkInfo) *LinkGroupResponse {
	return &LinkGroupResponse{
		Success:      linkInfo.Success,
		Message:      linkInfo.Message,
		CommunityJID: linkInfo.CommunityJID,
		GroupJID:     linkInfo.GroupJID,
		LinkedAt:     linkInfo.LinkedAt,
	}
}

// UnlinkGroupResponse represents the response for unlinking a group from a community
type UnlinkGroupResponse struct {
	UnlinkedAt   time.Time `json:"unlinkedAt"`
	Message      string    `json:"message"`
	CommunityJID string    `json:"communityJid"`
	GroupJID     string    `json:"groupJid"`
	Success      bool      `json:"success"`
}

// NewUnlinkGroupResponse creates a new unlink group response
func NewUnlinkGroupResponse(communityJID, groupJID string, success bool, message string) *UnlinkGroupResponse {
	return &UnlinkGroupResponse{
		Success:      success,
		Message:      message,
		CommunityJID: communityJID,
		GroupJID:     groupJID,
		UnlinkedAt:   time.Now(),
	}
}

// CommunityInfoResponse represents the response for community information
type CommunityInfoResponse struct {
	ID               string `json:"id"`
	JID              string `json:"jid"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ParticipantCount int    `json:"participantCount"`
	GroupCount       int    `json:"groupCount"`
	IsOwner          bool   `json:"isOwner"`
	IsAdmin          bool   `json:"isAdmin"`
	IsMuted          bool   `json:"isMuted"`
	IsAnnouncement   bool   `json:"isAnnouncement"`
}

// NewCommunityInfoResponse creates a new community info response
func NewCommunityInfoResponse(info *community.CommunityInfo) *CommunityInfoResponse {
	return &CommunityInfoResponse{
		ID:               info.ID,
		JID:              info.JID,
		Name:             info.Name,
		Description:      info.Description,
		ParticipantCount: info.ParticipantCount,
		GroupCount:       info.GroupCount,
		IsOwner:          info.IsOwner,
		IsAdmin:          info.IsAdmin,
		IsMuted:          info.IsMuted,
		IsAnnouncement:   info.IsAnnouncement,
	}
}

// LinkedGroupInfo represents information about a linked group
type LinkedGroupInfo struct {
	LinkedAt         time.Time `json:"linkedAt"`
	JID              string    `json:"jid"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	ParticipantCount int       `json:"participantCount"`
	IsOwner          bool      `json:"isOwner"`
	IsAdmin          bool      `json:"isAdmin"`
}

// NewLinkedGroupInfo creates a new linked group info
func NewLinkedGroupInfo(group *community.LinkedGroup) *LinkedGroupInfo {
	return &LinkedGroupInfo{
		JID:              group.JID,
		Name:             group.Name,
		Description:      group.Description,
		ParticipantCount: group.ParticipantCount,
		IsOwner:          group.IsOwner,
		IsAdmin:          group.IsAdmin,
		LinkedAt:         group.LinkedAt,
	}
}

// SubGroupsResponse represents the response for community sub-groups
type SubGroupsResponse struct {
	CommunityJID string             `json:"communityJid"`
	Groups       []*LinkedGroupInfo `json:"groups"`
	TotalCount   int                `json:"totalCount"`
}

// NewSubGroupsResponse creates a new sub-groups response
func NewSubGroupsResponse(communityJID string, groups []*community.LinkedGroup) *SubGroupsResponse {
	groupInfos := make([]*LinkedGroupInfo, len(groups))
	for i, group := range groups {
		groupInfos[i] = NewLinkedGroupInfo(group)
	}

	return &SubGroupsResponse{
		CommunityJID: communityJID,
		Groups:       groupInfos,
		TotalCount:   len(groupInfos),
	}
}

// CommunityActionResponse represents a generic response for community actions
type CommunityActionResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// NewCommunityActionResponse creates a new community action response
func NewCommunityActionResponse(success bool, message string) *CommunityActionResponse {
	return &CommunityActionResponse{
		Success: success,
		Message: message,
	}
}

// CommunityErrorResponse represents an error response for community operations
type CommunityErrorResponse struct {
	Context map[string]interface{} `json:"context,omitempty"`
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details string                 `json:"details,omitempty"`
	Success bool                   `json:"success"`
}

// NewCommunityErrorResponse creates a new community error response
func NewCommunityErrorResponse(err error) *CommunityErrorResponse {
	response := &CommunityErrorResponse{
		Success: false,
		Error:   err.Error(),
	}

	// If it's a community-specific error, add additional details
	if communityErr, ok := community.GetCommunityError(err); ok {
		response.Code = communityErr.Code
		response.Details = communityErr.Details
		response.Context = communityErr.Context
	}

	return response
}

// CommunityStatsResponse represents community statistics
type CommunityStatsResponse struct {
	TotalParticipants int `json:"totalParticipants"`
	TotalGroups       int `json:"totalGroups"`
	ActiveGroups      int `json:"activeGroups"`
	RecentActivity    int `json:"recentActivity"`
}

// NewCommunityStatsResponse creates a new community stats response
func NewCommunityStatsResponse(stats *community.CommunityStats) *CommunityStatsResponse {
	return &CommunityStatsResponse{
		TotalParticipants: stats.TotalParticipants,
		TotalGroups:       stats.TotalGroups,
		ActiveGroups:      stats.ActiveGroups,
		RecentActivity:    stats.RecentActivity,
	}
}

// CommunityEventResponse represents a community event
type CommunityEventResponse struct {
	Timestamp    time.Time              `json:"timestamp"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	CommunityJID string                 `json:"communityJid"`
	ActorJID     string                 `json:"actorJid"`
	TargetJID    string                 `json:"targetJid,omitempty"`
}

// NewCommunityEventResponse creates a new community event response
func NewCommunityEventResponse(event *community.CommunityEvent) *CommunityEventResponse {
	return &CommunityEventResponse{
		ID:           event.ID,
		Type:         string(event.Type),
		CommunityJID: event.CommunityJID,
		ActorJID:     event.ActorJID,
		TargetJID:    event.TargetJID,
		Data:         event.Data,
		Timestamp:    event.Timestamp,
	}
}

// CommunityParticipantResponse represents a community participant
type CommunityParticipantResponse struct {
	JoinedAt time.Time `json:"joinedAt"`
	JID      string    `json:"jid"`
	Name     string    `json:"name"`
	IsOwner  bool      `json:"isOwner"`
	IsAdmin  bool      `json:"isAdmin"`
}

// NewCommunityParticipantResponse creates a new community participant response
func NewCommunityParticipantResponse(participant *community.CommunityParticipant) *CommunityParticipantResponse {
	return &CommunityParticipantResponse{
		JID:      participant.JID,
		Name:     participant.Name,
		IsOwner:  participant.IsOwner,
		IsAdmin:  participant.IsAdmin,
		JoinedAt: participant.JoinedAt,
	}
}

// CommunitySettingsResponse represents community settings
type CommunitySettingsResponse struct {
	WhoCanAddGroups    string `json:"whoCanAddGroups"`
	WhoCanSendMessages string `json:"whoCanSendMessages"`
	WhoCanEditInfo     string `json:"whoCanEditInfo"`
	IsAnnouncement     bool   `json:"isAnnouncement"`
	IsMuted            bool   `json:"isMuted"`
}

// NewCommunitySettingsResponse creates a new community settings response
func NewCommunitySettingsResponse(settings *community.CommunitySettings) *CommunitySettingsResponse {
	return &CommunitySettingsResponse{
		IsAnnouncement:     settings.IsAnnouncement,
		IsMuted:            settings.IsMuted,
		WhoCanAddGroups:    settings.WhoCanAddGroups,
		WhoCanSendMessages: settings.WhoCanSendMessages,
		WhoCanEditInfo:     settings.WhoCanEditInfo,
	}
}
