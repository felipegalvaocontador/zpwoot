package group

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrInvalidGroupJID     = errors.New("invalid group JID")
	ErrInvalidGroupName    = errors.New("invalid group name")
	ErrGroupNameTooLong    = errors.New("group name too long (max 25 characters)")
	ErrDescriptionTooLong  = errors.New("description too long (max 512 characters)")
	ErrNoParticipants      = errors.New("no participants provided")
	ErrInvalidAction       = errors.New("invalid action (must be add, remove, promote, or demote)")
	ErrGroupNotFound       = errors.New("group not found")
	ErrNotGroupAdmin       = errors.New("user is not a group admin")
	ErrCannotLeaveAsOwner  = errors.New("group owner cannot leave group")
	ErrInvalidInviteLink   = errors.New("invalid invite link")
	ErrParticipantNotFound = errors.New("participant not found in group")
	ErrAlreadyParticipant  = errors.New("user is already a participant")
	ErrCannotRemoveOwner   = errors.New("cannot remove group owner")
)

// GroupInfo represents a WhatsApp group
type GroupInfo struct {
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
	GroupJID     string             `json:"groupJid"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Owner        string             `json:"owner"`
	Participants []GroupParticipant `json:"participants"`
	Settings     GroupSettings      `json:"settings"`
}

// GroupParticipant represents a participant in a group
type GroupParticipant struct {
	JID          string `json:"jid"`
	IsAdmin      bool   `json:"isAdmin"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

// GroupSettings represents group configuration settings
type GroupSettings struct {
	Announce bool `json:"announce"` // Only admins can send messages
	Locked   bool `json:"locked"`   // Only admins can edit group info
}

// CreateGroupRequest represents the data needed to create a group
type CreateGroupRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Participants []string `json:"participants"`
}

// UpdateParticipantsRequest represents the data needed to update group participants
type UpdateParticipantsRequest struct {
	GroupJID     string   `json:"groupJid"`
	Action       string   `json:"action"`
	Participants []string `json:"participants"`
}

// UpdateParticipantsResult represents the result of updating participants
type UpdateParticipantsResult struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
}

// SetGroupNameRequest represents the data needed to set group name
type SetGroupNameRequest struct {
	GroupJID string `json:"groupJid"`
	Name     string `json:"name"`
}

// SetGroupDescriptionRequest represents the data needed to set group description
type SetGroupDescriptionRequest struct {
	GroupJID    string `json:"groupJid"`
	Description string `json:"description"`
}

// SetGroupPhotoRequest represents the data needed to set group photo
type SetGroupPhotoRequest struct {
	GroupJID string `json:"groupJid"`
	Photo    []byte `json:"photo"`
}

// UpdateGroupSettingsRequest represents the data needed to update group settings
type UpdateGroupSettingsRequest struct {
	Announce *bool  `json:"announce,omitempty"`
	Locked   *bool  `json:"locked,omitempty"`
	GroupJID string `json:"groupJid"`
}

// JoinGroupRequest represents the data needed to join a group
type JoinGroupRequest struct {
	InviteLink string `json:"inviteLink"`
}

// LeaveGroupRequest represents the data needed to leave a group
type LeaveGroupRequest struct {
	GroupJID string `json:"groupJid"`
}

// GetInviteLinkRequest represents the data needed to get group invite link
type GetInviteLinkRequest struct {
	GroupJID string `json:"groupJid"`
	Reset    bool   `json:"reset"`
}

// InviteLinkResponse represents the group invite link
type InviteLinkResponse struct {
	InviteLink string `json:"inviteLink"`
}

// Business logic methods

// IsCurrentUserAdmin checks if the current user is an admin of the group
func (g *GroupInfo) IsCurrentUserAdmin() bool {
	// This would need the current user's JID to check
	// For now, we'll return false as a placeholder
	// In a real implementation, this would check against the current session's JID
	return false
}

// GetAdmins returns all admin participants
func (g *GroupInfo) GetAdmins() []GroupParticipant {
	var admins []GroupParticipant
	for _, participant := range g.Participants {
		if participant.IsAdmin || participant.IsSuperAdmin {
			admins = append(admins, participant)
		}
	}
	return admins
}

// GetRegularParticipants returns all non-admin participants
func (g *GroupInfo) GetRegularParticipants() []GroupParticipant {
	var regular []GroupParticipant
	for _, participant := range g.Participants {
		if !participant.IsAdmin && !participant.IsSuperAdmin {
			regular = append(regular, participant)
		}
	}
	return regular
}

// HasParticipant checks if a JID is a participant in the group
func (g *GroupInfo) HasParticipant(jid string) bool {
	for _, participant := range g.Participants {
		if participant.JID == jid {
			return true
		}
	}
	return false
}

// IsParticipantAdmin checks if a specific participant is an admin
func (g *GroupInfo) IsParticipantAdmin(jid string) bool {
	for _, participant := range g.Participants {
		if participant.JID == jid {
			return participant.IsAdmin || participant.IsSuperAdmin
		}
	}
	return false
}

// AddParticipant adds a new participant to the group
func (g *GroupInfo) AddParticipant(jid string) {
	if !g.HasParticipant(jid) {
		g.Participants = append(g.Participants, GroupParticipant{
			JID:          jid,
			IsAdmin:      false,
			IsSuperAdmin: false,
		})
		g.UpdatedAt = time.Now()
	}
}

// RemoveParticipant removes a participant from the group
func (g *GroupInfo) RemoveParticipant(jid string) {
	for i, participant := range g.Participants {
		if participant.JID == jid {
			g.Participants = append(g.Participants[:i], g.Participants[i+1:]...)
			g.UpdatedAt = time.Now()
			break
		}
	}
}

// PromoteParticipant promotes a participant to admin
func (g *GroupInfo) PromoteParticipant(jid string) {
	for i, participant := range g.Participants {
		if participant.JID == jid {
			g.Participants[i].IsAdmin = true
			g.UpdatedAt = time.Now()
			break
		}
	}
}

// DemoteParticipant demotes an admin to regular participant
func (g *GroupInfo) DemoteParticipant(jid string) {
	for i, participant := range g.Participants {
		if participant.JID == jid {
			g.Participants[i].IsAdmin = false
			g.Participants[i].IsSuperAdmin = false
			g.UpdatedAt = time.Now()
			break
		}
	}
}

// UpdateName updates the group name
func (g *GroupInfo) UpdateName(name string) {
	g.Name = name
	g.UpdatedAt = time.Now()
}

// UpdateDescription updates the group description
func (g *GroupInfo) UpdateDescription(description string) {
	g.Description = description
	g.UpdatedAt = time.Now()
}

// UpdateSettings updates the group settings
func (g *GroupInfo) UpdateSettings(announce, locked *bool) {
	if announce != nil {
		g.Settings.Announce = *announce
	}
	if locked != nil {
		g.Settings.Locked = *locked
	}
	g.UpdatedAt = time.Now()
}

// Validate validates the group information
func (g *GroupInfo) Validate() error {
	if g.GroupJID == "" {
		return ErrInvalidGroupJID
	}
	if g.Name == "" {
		return ErrInvalidGroupName
	}
	if len(g.Participants) == 0 {
		return ErrNoParticipants
	}
	return nil
}

// ValidateCreateRequest validates a create group request
func (r *CreateGroupRequest) Validate() error {
	if r.Name == "" {
		return ErrInvalidGroupName
	}
	if len(r.Name) > 25 {
		return ErrGroupNameTooLong
	}
	if len(r.Participants) == 0 {
		return ErrNoParticipants
	}
	if len(r.Description) > 512 {
		return ErrDescriptionTooLong
	}
	return nil
}

// ValidateUpdateParticipantsRequest validates an update participants request
func (r *UpdateParticipantsRequest) Validate() error {
	if r.GroupJID == "" {
		return ErrInvalidGroupJID
	}
	if len(r.Participants) == 0 {
		return ErrNoParticipants
	}
	if r.Action != "add" && r.Action != "remove" && r.Action != "promote" && r.Action != "demote" {
		return ErrInvalidAction
	}
	return nil
}
