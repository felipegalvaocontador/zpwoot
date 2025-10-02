package group

import (
	"context"
)

// Repository define as operações de persistência para grupos
type Repository interface {
	// CRUD básico
	Create(ctx context.Context, group *Group) error
	GetByID(ctx context.Context, id string) (*Group, error)
	GetByGroupJID(ctx context.Context, sessionID, groupJID string) (*Group, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id string) error

	// Listagem
	ListBySession(ctx context.Context, sessionID string) ([]*Group, error)
	ListJoinedGroups(ctx context.Context, sessionID string) ([]*Group, error)

	// Participantes
	AddParticipant(ctx context.Context, groupID string, participant *Participant) error
	RemoveParticipant(ctx context.Context, groupID, participantJID string) error
	UpdateParticipant(ctx context.Context, groupID string, participant *Participant) error
	GetParticipants(ctx context.Context, groupID string) ([]Participant, error)

	// Configurações
	UpdateSettings(ctx context.Context, groupID string, settings *GroupSettings) error

	// Convites
	SaveInviteLink(ctx context.Context, invite *InviteLink) error
	GetInviteLink(ctx context.Context, groupJID string) (*InviteLink, error)
	RevokeInviteLink(ctx context.Context, groupJID string) error

	// Solicitações
	SaveGroupRequest(ctx context.Context, request *GroupRequest) error
	GetGroupRequests(ctx context.Context, groupJID string) ([]*GroupRequest, error)
	UpdateGroupRequest(ctx context.Context, groupJID, requesterJID, status string) error
}

// WhatsAppGateway define as operações do WhatsApp para grupos
type WhatsAppGateway interface {
	// Operações básicas de grupo
	CreateGroup(ctx context.Context, sessionID, name string, participants []string, description string) (*GroupInfo, error)
	GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*GroupInfo, error)
	ListJoinedGroups(ctx context.Context, sessionID string) ([]*GroupInfo, error)

	// Gerenciamento de participantes
	AddParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	RemoveParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	PromoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error
	DemoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error

	// Configurações do grupo
	SetGroupName(ctx context.Context, sessionID, groupJID, name string) error
	SetGroupDescription(ctx context.Context, sessionID, groupJID, description string) error
	SetGroupPhoto(ctx context.Context, sessionID, groupJID string, photoData []byte) error

	// Configurações avançadas
	SetGroupAnnounce(ctx context.Context, sessionID, groupJID string, announce bool) error
	SetGroupRestrict(ctx context.Context, sessionID, groupJID string, restrict bool) error
	SetGroupLocked(ctx context.Context, sessionID, groupJID string, locked bool) error

	// Links de convite
	GetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (*InviteLink, error)
	RevokeGroupInviteLink(ctx context.Context, sessionID, groupJID string) error
	JoinGroupViaLink(ctx context.Context, sessionID, inviteLink string) (*GroupInfo, error)

	// Ações de grupo
	LeaveGroup(ctx context.Context, sessionID, groupJID string) error
	JoinGroupWithInvite(ctx context.Context, sessionID, groupJID, inviteCode string) (*GroupInfo, error)

	// Solicitações de entrada
	GetGroupRequestParticipants(ctx context.Context, sessionID, groupJID string) ([]*GroupRequest, error)
	ApproveGroupRequest(ctx context.Context, sessionID, groupJID string, requesterJIDs []string) error
	RejectGroupRequest(ctx context.Context, sessionID, groupJID string, requesterJIDs []string) error

	// Informações de convite
	GetGroupInfoFromInviteLink(ctx context.Context, sessionID, inviteLink string) (*GroupInfo, error)
	GetGroupInfoFromInvite(ctx context.Context, sessionID, groupJID, inviteCode string) (*GroupInfo, error)
}

// Service define a lógica de negócio para grupos
type Service interface {
	// Validações
	ValidateGroupCreation(req *CreateGroupRequest) error
	ValidateGroupName(name string) error
	ValidateGroupDescription(description string) error
	ValidateParticipants(participants []string) error
	ValidateInviteLink(inviteLink string) error
	ValidateJID(jid string) error

	// Permissões
	CanPerformAction(userJID, groupJID string, action GroupAction, groupInfo *GroupInfo) error
	IsGroupAdmin(userJID, groupJID string, groupInfo *GroupInfo) bool
	IsGroupOwner(userJID, groupJID string, groupInfo *GroupInfo) bool

	// Processamento de mudanças
	ProcessParticipantChanges(req *UpdateParticipantsRequest, currentGroup *GroupInfo) error
	ProcessSettingsChanges(req *UpdateGroupSettingsRequest, currentGroup *GroupInfo) error

	// Utilitários
	NormalizeJID(jid string) string
	ExtractPhoneNumber(jid string) string
	FormatGroupJID(groupID string) string
}

// EventHandler define como eventos de grupo são processados
type EventHandler interface {
	// Eventos de grupo
	OnGroupCreated(ctx context.Context, sessionID string, groupInfo *GroupInfo) error
	OnGroupInfoChanged(ctx context.Context, sessionID string, groupJID string, changes map[string]interface{}) error
	OnGroupSettingsChanged(ctx context.Context, sessionID string, groupJID string, settings *GroupSettings) error

	// Eventos de participantes
	OnParticipantAdded(ctx context.Context, sessionID, groupJID string, participants []string, addedBy string) error
	OnParticipantRemoved(ctx context.Context, sessionID, groupJID string, participants []string, removedBy string) error
	OnParticipantPromoted(ctx context.Context, sessionID, groupJID string, participants []string, promotedBy string) error
	OnParticipantDemoted(ctx context.Context, sessionID, groupJID string, participants []string, demotedBy string) error
	OnParticipantLeft(ctx context.Context, sessionID, groupJID, participantJID string) error

	// Eventos de convite
	OnInviteLinkGenerated(ctx context.Context, sessionID, groupJID string, inviteLink *InviteLink) error
	OnInviteLinkRevoked(ctx context.Context, sessionID, groupJID string) error
	OnGroupJoined(ctx context.Context, sessionID string, groupInfo *GroupInfo, joinMethod string) error

	// Eventos de solicitação
	OnJoinRequestReceived(ctx context.Context, sessionID, groupJID, requesterJID string) error
	OnJoinRequestApproved(ctx context.Context, sessionID, groupJID string, approvedJIDs []string, approvedBy string) error
	OnJoinRequestRejected(ctx context.Context, sessionID, groupJID string, rejectedJIDs []string, rejectedBy string) error
}

// QRGenerator define como QR codes são gerados para grupos
type QRGenerator interface {
	GenerateGroupInviteQR(inviteLink string) ([]byte, error)
	GenerateGroupInfoQR(groupInfo *GroupInfo) ([]byte, error)
}

// Validator define validações específicas para grupos
type Validator interface {
	ValidateGroupName(name string) error
	ValidateGroupDescription(description string) error
	ValidateParticipantJID(jid string) error
	ValidateInviteLink(link string) error
	ValidateGroupSettings(settings *GroupSettings) error
}

// ===== REQUEST/RESPONSE TYPES =====

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

// UpdateGroupSettingsRequest representa uma solicitação de atualização de configurações
type UpdateGroupSettingsRequest struct {
	GroupJID         string `json:"group_jid" validate:"required"`
	Announce         *bool  `json:"announce,omitempty"`
	Restrict         *bool  `json:"restrict,omitempty"`
	JoinApprovalMode string `json:"join_approval_mode,omitempty" validate:"omitempty,oneof=auto admin_approval"`
	MemberAddMode    string `json:"member_add_mode,omitempty" validate:"omitempty,oneof=all_members only_admins"`
	Locked           *bool  `json:"locked,omitempty"`
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

// GetInviteLinkRequest representa uma solicitação de link de convite
type GetInviteLinkRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}

// JoinGroupRequest representa uma solicitação para entrar em grupo
type JoinGroupRequest struct {
	InviteLink string `json:"invite_link" validate:"required"`
}

// LeaveGroupRequest representa uma solicitação para sair do grupo
type LeaveGroupRequest struct {
	GroupJID string `json:"group_jid" validate:"required"`
}

// GroupRequestAction representa uma ação em solicitação de grupo
type GroupRequestAction struct {
	GroupJID      string   `json:"group_jid" validate:"required"`
	RequesterJIDs []string `json:"requester_jids" validate:"required,min=1"`
	Action        string   `json:"action" validate:"required,oneof=approve reject"`
}
