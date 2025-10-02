package community

import (
	"fmt"
	"time"

	"zpwoot/internal/domain/community"
)

type LinkGroupRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
	GroupJID     string `json:"groupJid" validate:"required"`
}

func (r *LinkGroupRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

type UnlinkGroupRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
	GroupJID     string `json:"groupJid" validate:"required"`
}

func (r *UnlinkGroupRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	return nil
}

type GetCommunityInfoRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
}

func (r *GetCommunityInfoRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	return nil
}

func (r *GetCommunityInfoRequest) GetCommunityJID() string {
	return r.CommunityJID
}

type GetSubGroupsRequest struct {
	CommunityJID string `json:"communityJid" validate:"required"`
}

func (r *GetSubGroupsRequest) Validate() error {
	if r.CommunityJID == "" {
		return fmt.Errorf("community JID is required")
	}
	return nil
}

func (r *GetSubGroupsRequest) GetCommunityJID() string {
	return r.CommunityJID
}

type LinkGroupResponse struct {
	LinkedAt     time.Time `json:"linkedAt"`
	Message      string    `json:"message"`
	CommunityJID string    `json:"communityJid"`
	GroupJID     string    `json:"groupJid"`
	Success      bool      `json:"success"`
}

func NewLinkGroupResponse(linkInfo *community.GroupLinkInfo) *LinkGroupResponse {
	return &LinkGroupResponse{
		Success:      linkInfo.Success,
		Message:      linkInfo.Message,
		CommunityJID: linkInfo.CommunityJID,
		GroupJID:     linkInfo.GroupJID,
		LinkedAt:     linkInfo.LinkedAt,
	}
}

type UnlinkGroupResponse struct {
	UnlinkedAt   time.Time `json:"unlinkedAt"`
	Message      string    `json:"message"`
	CommunityJID string    `json:"communityJid"`
	GroupJID     string    `json:"groupJid"`
	Success      bool      `json:"success"`
}

func NewUnlinkGroupResponse(communityJID, groupJID string, success bool, message string) *UnlinkGroupResponse {
	return &UnlinkGroupResponse{
		Success:      success,
		Message:      message,
		CommunityJID: communityJID,
		GroupJID:     groupJID,
		UnlinkedAt:   time.Now(),
	}
}

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

type LinkedGroupInfo struct {
	LinkedAt         time.Time `json:"linkedAt"`
	JID              string    `json:"jid"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	ParticipantCount int       `json:"participantCount"`
	IsOwner          bool      `json:"isOwner"`
	IsAdmin          bool      `json:"isAdmin"`
}

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

type SubGroupsResponse struct {
	CommunityJID string             `json:"communityJid"`
	Groups       []*LinkedGroupInfo `json:"groups"`
	TotalCount   int                `json:"totalCount"`
}

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

type CommunityActionResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func NewCommunityActionResponse(success bool, message string) *CommunityActionResponse {
	return &CommunityActionResponse{
		Success: success,
		Message: message,
	}
}

type CommunityErrorResponse struct {
	Context map[string]interface{} `json:"context,omitempty"`
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details string                 `json:"details,omitempty"`
	Success bool                   `json:"success"`
}

func NewCommunityErrorResponse(err error) *CommunityErrorResponse {
	response := &CommunityErrorResponse{
		Success: false,
		Error:   err.Error(),
	}

	if communityErr, ok := community.GetCommunityError(err); ok {
		response.Code = communityErr.Code
		response.Details = communityErr.Details
		response.Context = communityErr.Context
	}

	return response
}

type CommunityStatsResponse struct {
	TotalParticipants int `json:"totalParticipants"`
	TotalGroups       int `json:"totalGroups"`
	ActiveGroups      int `json:"activeGroups"`
	RecentActivity    int `json:"recentActivity"`
}

func NewCommunityStatsResponse(stats *community.CommunityStats) *CommunityStatsResponse {
	return &CommunityStatsResponse{
		TotalParticipants: stats.TotalParticipants,
		TotalGroups:       stats.TotalGroups,
		ActiveGroups:      stats.ActiveGroups,
		RecentActivity:    stats.RecentActivity,
	}
}

type CommunityEventResponse struct {
	Timestamp    time.Time              `json:"timestamp"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	CommunityJID string                 `json:"communityJid"`
	ActorJID     string                 `json:"actorJid"`
	TargetJID    string                 `json:"targetJid,omitempty"`
}

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

type CommunityParticipantResponse struct {
	JoinedAt time.Time `json:"joinedAt"`
	JID      string    `json:"jid"`
	Name     string    `json:"name"`
	IsOwner  bool      `json:"isOwner"`
	IsAdmin  bool      `json:"isAdmin"`
}

func NewCommunityParticipantResponse(participant *community.CommunityParticipant) *CommunityParticipantResponse {
	return &CommunityParticipantResponse{
		JID:      participant.JID,
		Name:     participant.Name,
		IsOwner:  participant.IsOwner,
		IsAdmin:  participant.IsAdmin,
		JoinedAt: participant.JoinedAt,
	}
}

type CommunitySettingsResponse struct {
	WhoCanAddGroups    string `json:"whoCanAddGroups"`
	WhoCanSendMessages string `json:"whoCanSendMessages"`
	WhoCanEditInfo     string `json:"whoCanEditInfo"`
	IsAnnouncement     bool   `json:"isAnnouncement"`
	IsMuted            bool   `json:"isMuted"`
}

func NewCommunitySettingsResponse(settings *community.CommunitySettings) *CommunitySettingsResponse {
	return &CommunitySettingsResponse{
		IsAnnouncement:     settings.IsAnnouncement,
		IsMuted:            settings.IsMuted,
		WhoCanAddGroups:    settings.WhoCanAddGroups,
		WhoCanSendMessages: settings.WhoCanSendMessages,
		WhoCanEditInfo:     settings.WhoCanEditInfo,
	}
}
