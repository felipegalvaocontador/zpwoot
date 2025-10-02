package group

import (
	"time"

	"github.com/google/uuid"
)

// Group representa um grupo WhatsApp no sistema
type Group struct {
	// Identificadores únicos
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`

	// WhatsApp Group Identifiers
	GroupJID    string `json:"group_jid"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Owner       string `json:"owner"`

	// Group Settings
	Settings GroupSettings `json:"settings"`

	// Participants
	Participants []Participant `json:"participants"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupSettings representa as configurações de um grupo
type GroupSettings struct {
	// Who can send messages
	Announce bool `json:"announce"` // true = only admins, false = all members

	// Who can edit group info
	Restrict bool `json:"restrict"` // true = only admins, false = all members

	// Join approval mode
	JoinApprovalMode string `json:"join_approval_mode"` // "auto", "admin_approval"

	// Member add mode
	MemberAddMode string `json:"member_add_mode"` // "all_members", "only_admins"

	// Group locked (can't be modified)
	Locked bool `json:"locked"`
}

// Participant representa um participante do grupo
type Participant struct {
	JID      string           `json:"jid"`
	Role     ParticipantRole  `json:"role"`
	JoinedAt time.Time        `json:"joined_at"`
	AddedBy  string           `json:"added_by,omitempty"`
	Status   ParticipantStatus `json:"status"`
}

// ParticipantRole define os papéis dos participantes
type ParticipantRole string

const (
	ParticipantRoleOwner      ParticipantRole = "owner"
	ParticipantRoleAdmin      ParticipantRole = "admin"
	ParticipantRoleMember     ParticipantRole = "member"
	ParticipantRolePending    ParticipantRole = "pending"
	ParticipantRoleRequesting ParticipantRole = "requesting"
)

// ParticipantStatus define o status dos participantes
type ParticipantStatus string

const (
	ParticipantStatusActive   ParticipantStatus = "active"
	ParticipantStatusLeft     ParticipantStatus = "left"
	ParticipantStatusRemoved  ParticipantStatus = "removed"
	ParticipantStatusBanned   ParticipantStatus = "banned"
	ParticipantStatusPending  ParticipantStatus = "pending"
	ParticipantStatusRequesting ParticipantStatus = "requesting"
)

// GroupAction representa ações que podem ser realizadas em grupos
type GroupAction string

const (
	GroupActionCreate              GroupAction = "create"
	GroupActionAddParticipant      GroupAction = "add_participant"
	GroupActionRemoveParticipant   GroupAction = "remove_participant"
	GroupActionPromoteParticipant  GroupAction = "promote_participant"
	GroupActionDemoteParticipant   GroupAction = "demote_participant"
	GroupActionSetName             GroupAction = "set_name"
	GroupActionSetDescription      GroupAction = "set_description"
	GroupActionSetPhoto            GroupAction = "set_photo"
	GroupActionSetSettings         GroupAction = "set_settings"
	GroupActionLeave               GroupAction = "leave"
	GroupActionJoin                GroupAction = "join"
	GroupActionGetInviteLink       GroupAction = "get_invite_link"
	GroupActionRevokeInviteLink    GroupAction = "revoke_invite_link"
)

// GroupInfo representa informações básicas de um grupo
type GroupInfo struct {
	GroupJID     string        `json:"group_jid"`
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	Owner        string        `json:"owner"`
	Participants []Participant `json:"participants"`
	Settings     GroupSettings `json:"settings"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// InviteLink representa um link de convite do grupo
type InviteLink struct {
	GroupJID  string    `json:"group_jid"`
	Link      string    `json:"link"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool      `json:"is_active"`
}

// GroupRequest representa uma solicitação para entrar no grupo
type GroupRequest struct {
	GroupJID    string    `json:"group_jid"`
	RequesterJID string   `json:"requester_jid"`
	RequestedAt time.Time `json:"requested_at"`
	Status      string    `json:"status"` // "pending", "approved", "rejected"
	ReviewedBy  string    `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
}

// ===== MÉTODOS DE HELPER =====

// HasParticipant verifica se um JID é participante do grupo
func (g *Group) HasParticipant(jid string) bool {
	for _, participant := range g.Participants {
		if participant.JID == jid && participant.Status == ParticipantStatusActive {
			return true
		}
	}
	return false
}

// IsParticipantAdmin verifica se um participante é admin
func (g *Group) IsParticipantAdmin(jid string) bool {
	for _, participant := range g.Participants {
		if participant.JID == jid && 
		   (participant.Role == ParticipantRoleAdmin || participant.Role == ParticipantRoleOwner) &&
		   participant.Status == ParticipantStatusActive {
			return true
		}
	}
	return false
}

// IsParticipantOwner verifica se um participante é o dono
func (g *Group) IsParticipantOwner(jid string) bool {
	return g.Owner == jid
}

// GetParticipant retorna um participante específico
func (g *Group) GetParticipant(jid string) *Participant {
	for i, participant := range g.Participants {
		if participant.JID == jid {
			return &g.Participants[i]
		}
	}
	return nil
}

// GetActiveParticipants retorna apenas participantes ativos
func (g *Group) GetActiveParticipants() []Participant {
	var active []Participant
	for _, participant := range g.Participants {
		if participant.Status == ParticipantStatusActive {
			active = append(active, participant)
		}
	}
	return active
}

// GetAdmins retorna apenas administradores ativos
func (g *Group) GetAdmins() []Participant {
	var admins []Participant
	for _, participant := range g.Participants {
		if (participant.Role == ParticipantRoleAdmin || participant.Role == ParticipantRoleOwner) &&
		   participant.Status == ParticipantStatusActive {
			admins = append(admins, participant)
		}
	}
	return admins
}

// CanPerformAction verifica se um usuário pode realizar uma ação
func (g *Group) CanPerformAction(userJID string, action GroupAction) bool {
	participant := g.GetParticipant(userJID)
	if participant == nil || participant.Status != ParticipantStatusActive {
		return false
	}

	switch action {
	case GroupActionAddParticipant, GroupActionRemoveParticipant, 
		 GroupActionPromoteParticipant, GroupActionDemoteParticipant,
		 GroupActionSetName, GroupActionSetDescription, GroupActionSetPhoto,
		 GroupActionSetSettings, GroupActionGetInviteLink, GroupActionRevokeInviteLink:
		return g.IsParticipantAdmin(userJID)
	case GroupActionLeave:
		// Owner cannot leave unless transferring ownership
		return userJID != g.Owner
	case GroupActionJoin:
		return !g.HasParticipant(userJID)
	default:
		return false
	}
}

// HasParticipant verifica se um JID é participante do grupo (método para GroupInfo)
func (gi *GroupInfo) HasParticipant(jid string) bool {
	for _, participant := range gi.Participants {
		if participant.JID == jid && participant.Status == ParticipantStatusActive {
			return true
		}
	}
	return false
}

// IsParticipantAdmin verifica se um participante é admin (método para GroupInfo)
func (gi *GroupInfo) IsParticipantAdmin(jid string) bool {
	for _, participant := range gi.Participants {
		if participant.JID == jid && 
		   (participant.Role == ParticipantRoleAdmin || participant.Role == ParticipantRoleOwner) &&
		   participant.Status == ParticipantStatusActive {
			return true
		}
	}
	return false
}
