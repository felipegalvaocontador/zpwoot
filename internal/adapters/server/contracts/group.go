package contracts

import (
	"time"
)

type CreateGroupRequest struct {
	Name         string   `json:"name" validate:"required,min=1,max=25"`
	Description  string   `json:"description,omitempty" validate:"max=512"`
	Participants []string `json:"participants" validate:"required,min=1,max=256"`
}

type UpdateParticipantsRequest struct {
	GroupJID     string   `json:"group_jid" validate:"required"`
	Action       string   `json:"action" validate:"required,oneof=add remove promote demote"`
	Participants []string `json:"participants" validate:"required,min=1"`
}

type SetGroupNameRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
	Name     string `json:"name" validate:"required,min=1,max=25"`
}

type SetGroupDescriptionRequest struct {
	GroupJID    string `json:"group_jid" validate:"required"`
	Description string `json:"description" validate:"max=512"`
}

type SetGroupPhotoRequest struct {
	GroupJID  string `json:"group_jid" validate:"required"`
	PhotoData []byte `json:"photo_data" validate:"required"`
	MimeType  string `json:"mime_type" validate:"required"`
}

type UpdateGroupSettingsRequest struct {
	GroupJID         string `json:"group_jid" validate:"required"`
	Announce         *bool  `json:"announce,omitempty"`
	Restrict         *bool  `json:"restrict,omitempty"`
	JoinApprovalMode string `json:"join_approval_mode,omitempty" validate:"omitempty,oneof=auto admin_approval"`
	MemberAddMode    string `json:"member_add_mode,omitempty" validate:"omitempty,oneof=all_members only_admins"`
	Locked           *bool  `json:"locked,omitempty"`
}

type GetInviteLinkRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}

type JoinGroupViaLinkRequest struct {
	InviteLink string `json:"invite_link" validate:"required"`
}

type LeaveGroupRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}

type GetGroupInfoFromInviteRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type JoinGroupWithInviteRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type GroupRequestActionRequest struct {
	GroupJID      string   `json:"group_jid" validate:"required"`
	RequesterJIDs []string `json:"requester_jids" validate:"required,min=1"`
	Action        string   `json:"action" validate:"required,oneof=approve reject"`
}

type CreateGroupResponse struct {
	GroupJID     string    `json:"group_jid"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Participants []string  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
}

type ListGroupsResponse struct {
	Groups  []GroupInfo `json:"groups"`
	Count   int         `json:"count"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
}

type GroupInfo struct {
	GroupJID     string    `json:"group_jid"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Owner        string    `json:"owner"`
	Participants int       `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
}

type GetGroupInfoResponse struct {
	GroupJID     string            `json:"group_jid"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Owner        string            `json:"owner"`
	Participants []ParticipantInfo `json:"participants"`
	Settings     GroupSettings     `json:"settings"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
}

type ParticipantInfo struct {
	JID      string    `json:"jid"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	Status   string    `json:"status"`
}

type GroupSettings struct {
	Announce         bool   `json:"announce"`
	Restrict         bool   `json:"restrict"`
	JoinApprovalMode string `json:"join_approval_mode"`
	MemberAddMode    string `json:"member_add_mode"`
	Locked           bool   `json:"locked"`
}

type UpdateParticipantsResponse struct {
	GroupJID     string   `json:"group_jid"`
	Action       string   `json:"action"`
	Participants []string `json:"participants"`
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
}

type SetGroupNameResponse struct {
	GroupJID string `json:"group_jid"`
	Name     string `json:"name"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

type SetGroupDescriptionResponse struct {
	GroupJID    string `json:"group_jid"`
	Description string `json:"description"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
}

type SetGroupPhotoResponse struct {
	GroupJID string `json:"group_jid"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

type UpdateGroupSettingsResponse struct {
	GroupJID string        `json:"group_jid"`
	Settings GroupSettings `json:"settings"`
	Success  bool          `json:"success"`
	Message  string        `json:"message"`
}

type GetInviteLinkResponse struct {
	GroupJID   string `json:"group_jid"`
	InviteLink string `json:"invite_link"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

type JoinGroupResponse struct {
	GroupJID string `json:"group_jid"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

type LeaveGroupResponse struct {
	GroupJID string `json:"group_jid"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

type GetGroupRequestParticipantsResponse struct {
	GroupJID     string             `json:"group_jid"`
	Participants []GroupRequestInfo `json:"participants"`
	Count        int                `json:"count"`
	Success      bool               `json:"success"`
	Message      string             `json:"message"`
}

type GroupRequestInfo struct {
	RequesterJID string    `json:"requester_jid"`
	RequestedAt  time.Time `json:"requested_at"`
	Status       string    `json:"status"`
}

type UpdateGroupRequestParticipantsResponse struct {
	GroupJID      string   `json:"group_jid"`
	Action        string   `json:"action"`
	RequesterJIDs []string `json:"requester_jids"`
	Success       bool     `json:"success"`
	Message       string   `json:"message"`
}

type GetGroupInfoFromInviteResponse struct {
	GroupJID  string    `json:"group_jid"`
	Code      string    `json:"code"`
	GroupInfo GroupInfo `json:"group_info"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
}

type JoinGroupWithInviteResponse struct {
	GroupJID string `json:"group_jid"`
	Code     string `json:"code"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

type SetGroupJoinApprovalModeResponse struct {
	GroupJID         string `json:"group_jid"`
	JoinApprovalMode string `json:"join_approval_mode"`
	Success          bool   `json:"success"`
	Message          string `json:"message"`
}

type SetGroupMemberAddModeResponse struct {
	GroupJID      string `json:"group_jid"`
	MemberAddMode string `json:"member_add_mode"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}

type GetGroupInfoFromLinkResponse struct {
	InviteLink string    `json:"invite_link"`
	GroupInfo  GroupInfo `json:"group_info"`
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
}
