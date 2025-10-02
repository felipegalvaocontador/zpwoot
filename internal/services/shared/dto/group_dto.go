package dto

import (
	"time"
)

// ===== REQUEST DTOs =====

// CreateGroupRequest representa uma solicitação de criação de grupo
type CreateGroupRequest struct {
	Name         string   `json:"name" validate:"required,min=1,max=25"`
	Description  string   `json:"description,omitempty" validate:"max=512"`
	Participants []string `json:"participants" validate:"required,min=1,max=256"`
}

// UpdateParticipantsRequest representa uma solicitação de atualização de participantes
type UpdateParticipantsRequest struct {
	GroupJID     string   `json:"group_jid" validate:"required"`
	Action       string   `json:"action" validate:"required,oneof=add remove promote demote"`
	Participants []string `json:"participants" validate:"required,min=1"`
}

// SetGroupNameRequest representa uma solicitação de alteração de nome
type SetGroupNameRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
	Name     string `json:"name" validate:"required,min=1,max=25"`
}

// SetGroupDescriptionRequest representa uma solicitação de alteração de descrição
type SetGroupDescriptionRequest struct {
	GroupJID    string `json:"group_jid" validate:"required"`
	Description string `json:"description" validate:"max=512"`
}

// SetGroupPhotoRequest representa uma solicitação de alteração de foto
type SetGroupPhotoRequest struct {
	GroupJID  string `json:"group_jid" validate:"required"`
	PhotoData []byte `json:"photo_data" validate:"required"`
	MimeType  string `json:"mime_type" validate:"required"`
}

// UpdateGroupSettingsRequest representa uma solicitação de atualização de configurações
type UpdateGroupSettingsRequest struct {
	GroupJID         string `json:"group_jid" validate:"required"`
	Announce         *bool  `json:"announce,omitempty"`
	Restrict         *bool  `json:"restrict,omitempty"`
	JoinApprovalMode string `json:"join_approval_mode,omitempty" validate:"omitempty,oneof=auto admin_approval"`
	MemberAddMode    string `json:"member_add_mode,omitempty" validate:"omitempty,oneof=all_members only_admins"`
	Locked           *bool  `json:"locked,omitempty"`
}

// GetInviteLinkRequest representa uma solicitação de link de convite
type GetInviteLinkRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}

// JoinGroupViaLinkRequest representa uma solicitação para entrar em grupo via link
type JoinGroupViaLinkRequest struct {
	InviteLink string `json:"invite_link" validate:"required"`
}

// LeaveGroupRequest representa uma solicitação para sair do grupo
type LeaveGroupRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}

// GetGroupInfoFromInviteRequest representa uma solicitação de info via convite
type GetGroupInfoFromInviteRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

// JoinGroupWithInviteRequest representa uma solicitação para entrar com convite
type JoinGroupWithInviteRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

// GroupRequestActionRequest representa uma ação em solicitação de grupo
type GroupRequestActionRequest struct {
	GroupJID      string   `json:"group_jid" validate:"required"`
	RequesterJIDs []string `json:"requester_jids" validate:"required,min=1"`
	Action        string   `json:"action" validate:"required,oneof=approve reject"`
}

// ===== RESPONSE DTOs =====

// CreateGroupResponse representa a resposta de criação de grupo
type CreateGroupResponse struct {
	GroupJID     string    `json:"group_jid"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Participants []string  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
}

// ListGroupsResponse representa a resposta de listagem de grupos
type ListGroupsResponse struct {
	Groups  []GroupInfo `json:"groups"`
	Count   int         `json:"count"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
}

// GroupInfo representa informações básicas de um grupo
type GroupInfo struct {
	GroupJID     string    `json:"group_jid"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Owner        string    `json:"owner"`
	Participants int       `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetGroupInfoResponse representa a resposta de informações detalhadas do grupo
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

// ParticipantInfo representa informações de um participante
type ParticipantInfo struct {
	JID      string    `json:"jid"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	Status   string    `json:"status"`
}

// GroupSettings representa as configurações de um grupo
type GroupSettings struct {
	Announce         bool   `json:"announce"`
	Restrict         bool   `json:"restrict"`
	JoinApprovalMode string `json:"join_approval_mode"`
	MemberAddMode    string `json:"member_add_mode"`
	Locked           bool   `json:"locked"`
}

// UpdateParticipantsResponse representa a resposta de atualização de participantes
type UpdateParticipantsResponse struct {
	GroupJID     string   `json:"group_jid"`
	Action       string   `json:"action"`
	Participants []string `json:"participants"`
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
}

// SetGroupNameResponse representa a resposta de alteração de nome
type SetGroupNameResponse struct {
	GroupJID string `json:"group_jid"`
	Name     string `json:"name"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// SetGroupDescriptionResponse representa a resposta de alteração de descrição
type SetGroupDescriptionResponse struct {
	GroupJID    string `json:"group_jid"`
	Description string `json:"description"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
}

// SetGroupPhotoResponse representa a resposta de alteração de foto
type SetGroupPhotoResponse struct {
	GroupJID string `json:"group_jid"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// UpdateGroupSettingsResponse representa a resposta de atualização de configurações
type UpdateGroupSettingsResponse struct {
	GroupJID string        `json:"group_jid"`
	Settings GroupSettings `json:"settings"`
	Success  bool          `json:"success"`
	Message  string        `json:"message"`
}

// GetInviteLinkResponse representa a resposta de link de convite
type GetInviteLinkResponse struct {
	GroupJID   string `json:"group_jid"`
	InviteLink string `json:"invite_link"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// JoinGroupResponse representa a resposta de entrada em grupo
type JoinGroupResponse struct {
	GroupJID string `json:"group_jid"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// LeaveGroupResponse representa a resposta de saída do grupo
type LeaveGroupResponse struct {
	GroupJID string `json:"group_jid"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// GetGroupRequestParticipantsResponse representa a resposta de solicitações pendentes
type GetGroupRequestParticipantsResponse struct {
	GroupJID     string               `json:"group_jid"`
	Participants []GroupRequestInfo   `json:"participants"`
	Count        int                  `json:"count"`
	Success      bool                 `json:"success"`
	Message      string               `json:"message"`
}

// GroupRequestInfo representa informações de uma solicitação
type GroupRequestInfo struct {
	RequesterJID string    `json:"requester_jid"`
	RequestedAt  time.Time `json:"requested_at"`
	Status       string    `json:"status"`
}

// UpdateGroupRequestParticipantsResponse representa a resposta de ação em solicitações
type UpdateGroupRequestParticipantsResponse struct {
	GroupJID      string   `json:"group_jid"`
	Action        string   `json:"action"`
	RequesterJIDs []string `json:"requester_jids"`
	Success       bool     `json:"success"`
	Message       string   `json:"message"`
}

// GetGroupInfoFromInviteResponse representa a resposta de info via convite
type GetGroupInfoFromInviteResponse struct {
	GroupJID    string    `json:"group_jid"`
	Code        string    `json:"code"`
	GroupInfo   GroupInfo `json:"group_info"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
}

// JoinGroupWithInviteResponse representa a resposta de entrada com convite
type JoinGroupWithInviteResponse struct {
	GroupJID string `json:"group_jid"`
	Code     string `json:"code"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// SetGroupJoinApprovalModeResponse representa a resposta de modo de aprovação
type SetGroupJoinApprovalModeResponse struct {
	GroupJID         string `json:"group_jid"`
	JoinApprovalMode string `json:"join_approval_mode"`
	Success          bool   `json:"success"`
	Message          string `json:"message"`
}

// SetGroupMemberAddModeResponse representa a resposta de modo de adição
type SetGroupMemberAddModeResponse struct {
	GroupJID      string `json:"group_jid"`
	MemberAddMode string `json:"member_add_mode"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}

// GetGroupInfoFromLinkResponse representa a resposta de info via link
type GetGroupInfoFromLinkResponse struct {
	InviteLink string    `json:"invite_link"`
	GroupInfo  GroupInfo `json:"group_info"`
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
}
