package group

import (
	"time"

	"github.com/google/uuid"
)


type Group struct {

	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`


	GroupJID    string `json:"group_jid"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Owner       string `json:"owner"`


	Settings GroupSettings `json:"settings"`


	Participants []Participant `json:"participants"`


	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


type GroupSettings struct {

	Announce bool `json:"announce"`


	Restrict bool `json:"restrict"`


	JoinApprovalMode string `json:"join_approval_mode"`


	MemberAddMode string `json:"member_add_mode"`


	Locked bool `json:"locked"`
}


type Participant struct {
	JID      string           `json:"jid"`
	Role     ParticipantRole  `json:"role"`
	JoinedAt time.Time        `json:"joined_at"`
	AddedBy  string           `json:"added_by,omitempty"`
	Status   ParticipantStatus `json:"status"`
}


type ParticipantRole string

const (
	ParticipantRoleOwner      ParticipantRole = "owner"
	ParticipantRoleAdmin      ParticipantRole = "admin"
	ParticipantRoleMember     ParticipantRole = "member"
	ParticipantRolePending    ParticipantRole = "pending"
	ParticipantRoleRequesting ParticipantRole = "requesting"
)


type ParticipantStatus string

const (
	ParticipantStatusActive   ParticipantStatus = "active"
	ParticipantStatusLeft     ParticipantStatus = "left"
	ParticipantStatusRemoved  ParticipantStatus = "removed"
	ParticipantStatusBanned   ParticipantStatus = "banned"
	ParticipantStatusPending  ParticipantStatus = "pending"
	ParticipantStatusRequesting ParticipantStatus = "requesting"
)


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


type InviteLink struct {
	GroupJID  string    `json:"group_jid"`
	Link      string    `json:"link"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool      `json:"is_active"`
}


type GroupRequest struct {
	GroupJID    string    `json:"group_jid"`
	RequesterJID string   `json:"requester_jid"`
	RequestedAt time.Time `json:"requested_at"`
	Status      string    `json:"status"`
	ReviewedBy  string    `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
}




func (g *Group) HasParticipant(jid string) bool {
	for _, participant := range g.Participants {
		if participant.JID == jid && participant.Status == ParticipantStatusActive {
			return true
		}
	}
	return false
}


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


func (g *Group) IsParticipantOwner(jid string) bool {
	return g.Owner == jid
}


func (g *Group) GetParticipant(jid string) *Participant {
	for i, participant := range g.Participants {
		if participant.JID == jid {
			return &g.Participants[i]
		}
	}
	return nil
}


func (g *Group) GetActiveParticipants() []Participant {
	var active []Participant
	for _, participant := range g.Participants {
		if participant.Status == ParticipantStatusActive {
			active = append(active, participant)
		}
	}
	return active
}


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

		return userJID != g.Owner
	case GroupActionJoin:
		return !g.HasParticipant(userJID)
	default:
		return false
	}
}


func (gi *GroupInfo) HasParticipant(jid string) bool {
	for _, participant := range gi.Participants {
		if participant.JID == jid && participant.Status == ParticipantStatusActive {
			return true
		}
	}
	return false
}


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
