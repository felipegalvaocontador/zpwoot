package group

import (
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/types"

	"zpwoot/internal/domain/group"
)

type CreateGroupRequest struct {
	Name         string   `json:"name" validate:"required,min=1,max=25" example:"My Group"`
	Description  string   `json:"description,omitempty" validate:"max=512" example:"Group description"`
	Participants []string `json:"participants" validate:"required,min=1" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
} // @name CreateGroupRequest

type CreateGroupResponse struct {
	CreatedAt    time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	GroupJID     string    `json:"groupJid" example:"120363123456789012@g.us"`
	Name         string    `json:"name" example:"My Group"`
	Description  string    `json:"description,omitempty" example:"Group description"`
	Participants []string  `json:"participants" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
} // @name CreateGroupResponse

type GetGroupInfoRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
} // @name GetGroupInfoRequest

type GetGroupInfoResponse struct {
	CreatedAt    time.Time          `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    time.Time          `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
	GroupJID     string             `json:"groupJid" example:"120363123456789012@g.us"`
	Name         string             `json:"name" example:"My Group"`
	Description  string             `json:"description,omitempty" example:"Group description"`
	Owner        string             `json:"owner" example:"5511999999999@s.whatsapp.net"`
	Participants []GroupParticipant `json:"participants"`
	Settings     GroupSettings      `json:"settings"`
} // @name GetGroupInfoResponse

type GroupParticipant struct {
	JID          string `json:"jid" example:"5511999999999@s.whatsapp.net"`
	IsAdmin      bool   `json:"isAdmin" example:"false"`
	IsSuperAdmin bool   `json:"isSuperAdmin" example:"false"`
} // @name GroupParticipant

type GroupSettings struct {
	Announce bool `json:"announce" example:"false"` // Only admins can send messages
	Locked   bool `json:"locked" example:"false"`   // Only admins can edit group info
} // @name GroupSettings

type ListGroupsResponse struct {
	Groups []GroupInfo `json:"groups"`
	Total  int         `json:"total" example:"5"`
} // @name ListGroupsResponse

type GroupInfo struct {
	CreatedAt        time.Time `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	GroupJID         string    `json:"groupJid" example:"120363123456789012@g.us"`
	Name             string    `json:"name" example:"My Group"`
	Description      string    `json:"description,omitempty" example:"Group description"`
	ParticipantCount int       `json:"participantCount" example:"10"`
	IsAdmin          bool      `json:"isAdmin" example:"true"`
} // @name GroupInfo

type UpdateGroupParticipantsRequest struct {
	GroupJID     string   `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Action       string   `json:"action" validate:"required,oneof=add remove promote demote" example:"add"`
	Participants []string `json:"participants" validate:"required,min=1" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
} // @name UpdateGroupParticipantsRequest

type UpdateGroupParticipantsResponse struct {
	GroupJID     string   `json:"groupJid" example:"120363123456789012@g.us"`
	Participants []string `json:"participants" example:"5511999999999@s.whatsapp.net,5511888888888@s.whatsapp.net"`
	Action       string   `json:"action" example:"add"`
	Success      []string `json:"success" example:"5511999999999@s.whatsapp.net"`
	Failed       []string `json:"failed,omitempty" example:"5511888888888@s.whatsapp.net"`
} // @name UpdateGroupParticipantsResponse

type SetGroupNameRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Name     string `json:"name" validate:"required,min=1,max=25" example:"New Group Name"`
} // @name SetGroupNameRequest

func (r *SetGroupNameRequest) GetGroupJID() string {
	return r.GroupJID
}

type SetGroupDescriptionRequest struct {
	GroupJID    string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Description string `json:"description" validate:"max=512" example:"New group description"`
} // @name SetGroupDescriptionRequest

type GetGroupInviteLinkRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Reset    bool   `json:"reset,omitempty" example:"false"`
} // @name GetGroupInviteLinkRequest

type GetGroupInviteLinkResponse struct {
	GroupJID   string `json:"groupJid" example:"120363123456789012@g.us"`
	InviteLink string `json:"inviteLink" example:"https://chat.whatsapp.com/ABC123DEF456"`
} // @name GetGroupInviteLinkResponse

type JoinGroupRequest struct {
	InviteLink string `json:"inviteLink" validate:"required" example:"https://chat.whatsapp.com/ABC123DEF456"`
} // @name JoinGroupRequest

type JoinGroupResponse struct {
	GroupJID string `json:"groupJid" example:"120363123456789012@g.us"`
	Name     string `json:"name" example:"My Group"`
	Status   string `json:"status" example:"joined"`
} // @name JoinGroupResponse

type LeaveGroupRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
} // @name LeaveGroupRequest

type LeaveGroupResponse struct {
	GroupJID string `json:"groupJid" example:"120363123456789012@g.us"`
	Status   string `json:"status" example:"left"`
} // @name LeaveGroupResponse

type SetGroupPhotoRequest struct {
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Photo    string `json:"photo" validate:"required" example:"data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ..."`
} // @name SetGroupPhotoRequest

type UpdateGroupSettingsRequest struct {
	Announce *bool  `json:"announce,omitempty" example:"true"`
	Locked   *bool  `json:"locked,omitempty" example:"false"`
	GroupJID string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
} // @name UpdateGroupSettingsRequest

type GroupActionResponse struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	GroupJID  string    `json:"groupJid" example:"120363123456789012@g.us"`
	Status    string    `json:"status" example:"success"`
	Message   string    `json:"message" example:"Action completed successfully"`
} // @name GroupActionResponse

func (r *CreateGroupRequest) ToDomain() *group.CreateGroupRequest {
	return &group.CreateGroupRequest{
		Name:         r.Name,
		Participants: r.Participants,
		Description:  r.Description,
	}
}

func FromDomainGroupInfo(g *group.GroupInfo) *GetGroupInfoResponse {
	participants := make([]GroupParticipant, len(g.Participants))
	for i, p := range g.Participants {
		participants[i] = GroupParticipant{
			JID:          p.JID,
			IsAdmin:      p.IsAdmin,
			IsSuperAdmin: p.IsSuperAdmin,
		}
	}

	return &GetGroupInfoResponse{
		GroupJID:     g.GroupJID,
		Name:         g.Name,
		Description:  g.Description,
		Owner:        g.Owner,
		Participants: participants,
		Settings: GroupSettings{
			Announce: g.Settings.Announce,
			Locked:   g.Settings.Locked,
		},
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func FromDomainGroupList(groups []*group.GroupInfo) []GroupInfo {
	result := make([]GroupInfo, len(groups))
	for i, g := range groups {
		result[i] = GroupInfo{
			GroupJID:         g.GroupJID,
			Name:             g.Name,
			Description:      g.Description,
			ParticipantCount: len(g.Participants),
			IsAdmin:          g.IsCurrentUserAdmin(),
			CreatedAt:        g.CreatedAt,
		}
	}
	return result
}


type GetGroupInfoFromLinkRequest struct {
	InviteLink string `json:"inviteLink" validate:"required" example:"https://chat.whatsapp.com/ABC123DEF456"`
} // @name GetGroupInfoFromLinkRequest

func (r *GetGroupInfoFromLinkRequest) Validate() error {
	if r.InviteLink == "" {
		return fmt.Errorf("invite link is required")
	}
	return nil
}

type GetGroupInfoFromInviteRequest struct {
	GroupJID   string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Inviter    string `json:"inviter,omitempty" example:"5511999999999@s.whatsapp.net"`
	Code       string `json:"code" validate:"required" example:"ABC123DEF456"`
	Expiration int64  `json:"expiration,omitempty" example:"1640995200"`
} // @name GetGroupInfoFromInviteRequest

func (r *GetGroupInfoFromInviteRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	if r.Code == "" {
		return fmt.Errorf("invite code is required")
	}
	return nil
}

type JoinGroupWithInviteRequest struct {
	GroupJID   string `json:"groupJid" validate:"required" example:"120363123456789012@g.us"`
	Inviter    string `json:"inviter,omitempty" example:"5511999999999@s.whatsapp.net"`
	Code       string `json:"code" validate:"required" example:"ABC123DEF456"`
	Expiration int64  `json:"expiration,omitempty" example:"1640995200"`
} // @name JoinGroupWithInviteRequest

func (r *JoinGroupWithInviteRequest) Validate() error {
	if r.GroupJID == "" {
		return fmt.Errorf("group JID is required")
	}
	if r.Code == "" {
		return fmt.Errorf("invite code is required")
	}
	return nil
}

type GroupInfoFromLinkResponse struct {
	JID              string `json:"jid" example:"120363123456789012@g.us"`
	Name             string `json:"name" example:"My Group"`
	Description      string `json:"description" example:"Group description"`
	CreatedAt        string `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	ParticipantCount int    `json:"participantCount" example:"10"`
	IsAnnouncement   bool   `json:"isAnnouncement" example:"false"`
	IsLocked         bool   `json:"isLocked" example:"false"`
} // @name GroupInfoFromLinkResponse

func NewGroupInfoFromLinkResponse(groupInfo *types.GroupInfo) *GroupInfoFromLinkResponse {
	return &GroupInfoFromLinkResponse{
		JID:              groupInfo.JID.String(),
		Name:             groupInfo.Name,
		Description:      groupInfo.Topic,
		ParticipantCount: len(groupInfo.Participants),
		IsAnnouncement:   groupInfo.IsAnnounce,
		IsLocked:         groupInfo.IsLocked,
		CreatedAt:        groupInfo.GroupCreated.Format(time.RFC3339),
	}
}

type GroupInfoFromInviteResponse struct {
	JID              string `json:"jid" example:"120363123456789012@g.us"`
	Name             string `json:"name" example:"My Group"`
	Description      string `json:"description" example:"Group description"`
	CreatedAt        string `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	InviteCode       string `json:"inviteCode" example:"ABC123DEF456"`
	Inviter          string `json:"inviter,omitempty" example:"5511999999999@s.whatsapp.net"`
	ParticipantCount int    `json:"participantCount" example:"10"`
	IsAnnouncement   bool   `json:"isAnnouncement" example:"false"`
	IsLocked         bool   `json:"isLocked" example:"false"`
} // @name GroupInfoFromInviteResponse

func NewGroupInfoFromInviteResponse(groupInfo *types.GroupInfo, code, inviter string) *GroupInfoFromInviteResponse {
	return &GroupInfoFromInviteResponse{
		JID:              groupInfo.JID.String(),
		Name:             groupInfo.Name,
		Description:      groupInfo.Topic,
		ParticipantCount: len(groupInfo.Participants),
		IsAnnouncement:   groupInfo.IsAnnounce,
		IsLocked:         groupInfo.IsLocked,
		CreatedAt:        groupInfo.GroupCreated.Format(time.RFC3339),
		InviteCode:       code,
		Inviter:          inviter,
	}
}

type JoinGroupWithInviteResponse struct {
	Message  string `json:"message" example:"Successfully joined group"`
	GroupJID string `json:"groupJid" example:"120363123456789012@g.us"`
	JoinedAt string `json:"joinedAt" example:"2024-01-01T00:00:00Z"`
	Success  bool   `json:"success" example:"true"`
} // @name JoinGroupWithInviteResponse

func NewJoinGroupWithInviteResponse(groupJID string, success bool, message string) *JoinGroupWithInviteResponse {
	return &JoinGroupWithInviteResponse{
		Success:  success,
		Message:  message,
		GroupJID: groupJID,
		JoinedAt: time.Now().Format(time.RFC3339),
	}
}
