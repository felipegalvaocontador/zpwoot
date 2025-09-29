package group

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"zpwoot/pkg/uuid"
)

// JIDValidator interface for validating WhatsApp JIDs
type JIDValidator interface {
	IsValid(jid string) bool
	Normalize(jid string) string
}

type Service struct {
	generator    *uuid.Generator
	jidValidator JIDValidator
}

type Repository interface {
	CreateGroup(ctx context.Context, sessionID string, req *CreateGroupRequest) (*GroupInfo, error)
	GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*GroupInfo, error)
	ListJoinedGroups(ctx context.Context, sessionID string) ([]*GroupInfo, error)
	UpdateGroupParticipants(ctx context.Context, sessionID string, req *UpdateParticipantsRequest) (*UpdateParticipantsResult, error)
	SetGroupName(ctx context.Context, sessionID string, req *SetGroupNameRequest) error
	SetGroupDescription(ctx context.Context, sessionID string, req *SetGroupDescriptionRequest) error
	SetGroupPhoto(ctx context.Context, sessionID string, req *SetGroupPhotoRequest) error
	GetGroupInviteLink(ctx context.Context, sessionID string, req *GetInviteLinkRequest) (*InviteLinkResponse, error)
	JoinGroupViaLink(ctx context.Context, sessionID string, req *JoinGroupRequest) (*GroupInfo, error)
	LeaveGroup(ctx context.Context, sessionID string, req *LeaveGroupRequest) error
	UpdateGroupSettings(ctx context.Context, sessionID string, req *UpdateGroupSettingsRequest) error
}

func NewService(repo Repository, wameow interface{}, jidValidator JIDValidator) *Service {
	return &Service{
		generator:    uuid.New(),
		jidValidator: jidValidator,
	}
}

// ValidateGroupCreation validates group creation parameters
func (s *Service) ValidateGroupCreation(req *CreateGroupRequest) error {
	if req == nil {
		return ErrInvalidGroupName
	}

	if err := s.ValidateGroupName(req.Name); err != nil {
		return err
	}

	if len(req.Participants) == 0 {
		return ErrNoParticipants
	}

	if len(req.Participants) > 256 {
		return fmt.Errorf("too many participants (max 256)")
	}

	// Validate participant JIDs
	for _, participant := range req.Participants {
		if err := s.validateJID(participant); err != nil {
			return fmt.Errorf("invalid participant %s: %w", participant, err)
		}
	}

	if err := s.ValidateGroupDescription(req.Description); err != nil {
		return err
	}

	return nil
}

// ValidateParticipantUpdate validates participant update parameters
func (s *Service) ValidateParticipantUpdate(req *UpdateParticipantsRequest) error {
	if req == nil {
		return ErrInvalidGroupJID
	}

	if err := s.validateJID(req.GroupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if len(req.Participants) == 0 {
		return ErrNoParticipants
	}

	if len(req.Participants) > 50 {
		return fmt.Errorf("too many participants in single operation (max 50)")
	}

	// Validate participant JIDs
	for _, participant := range req.Participants {
		if err := s.validateJID(participant); err != nil {
			return fmt.Errorf("invalid participant %s: %w", participant, err)
		}
	}

	// Validate action
	validActions := []string{"add", "remove", "promote", "demote"}
	isValidAction := false
	for _, action := range validActions {
		if req.Action == action {
			isValidAction = true
			break
		}
	}
	if !isValidAction {
		return ErrInvalidAction
	}

	return nil
}

// ValidateGroupName validates group name
func (s *Service) ValidateGroupName(name string) error {
	if name == "" {
		return ErrInvalidGroupName
	}

	if len(name) > 25 {
		return ErrGroupNameTooLong
	}

	// Check for invalid characters (basic validation)
	if strings.TrimSpace(name) == "" {
		return ErrInvalidGroupName
	}

	return nil
}

// ValidateGroupDescription validates group description
func (s *Service) ValidateGroupDescription(description string) error {
	if len(description) > 512 {
		return ErrDescriptionTooLong
	}

	return nil
}

// ValidateInviteLink validates invite link format
func (s *Service) ValidateInviteLink(link string) error {
	if link == "" {
		return ErrInvalidInviteLink
	}

	// WhatsApp invite links follow the pattern: https://chat.whatsapp.com/XXXXXX
	inviteLinkPattern := `^https://chat\.whatsapp\.com/[A-Za-z0-9]+$`
	matched, err := regexp.MatchString(inviteLinkPattern, link)
	if err != nil {
		return fmt.Errorf("error validating invite link: %w", err)
	}

	if !matched {
		return ErrInvalidInviteLink
	}

	return nil
}

// CanPerformAction checks if user can perform a specific action on the group
func (s *Service) CanPerformAction(userJID, groupJID, action string, groupInfo *GroupInfo) error {
	if groupInfo == nil {
		return ErrGroupNotFound
	}

	// Check if user is a participant
	if !groupInfo.HasParticipant(userJID) {
		return ErrParticipantNotFound
	}

	// Actions that require admin privileges
	adminActions := []string{"remove", "promote", "demote", "set_name", "set_description", "set_photo", "set_settings"}
	requiresAdmin := false
	for _, adminAction := range adminActions {
		if action == adminAction {
			requiresAdmin = true
			break
		}
	}

	if requiresAdmin && !groupInfo.IsParticipantAdmin(userJID) {
		return ErrNotGroupAdmin
	}

	// Special case: cannot remove group owner
	if action == "remove" {
		for _, participant := range groupInfo.Participants {
			if participant.JID == userJID && participant.JID == groupInfo.Owner {
				return ErrCannotRemoveOwner
			}
		}
	}

	// Special case: group owner cannot leave
	if action == "leave" && userJID == groupInfo.Owner {
		return ErrCannotLeaveAsOwner
	}

	return nil
}

// ProcessParticipantChanges processes and validates participant changes
func (s *Service) ProcessParticipantChanges(req *UpdateParticipantsRequest, currentGroup *GroupInfo) error {
	if req == nil || currentGroup == nil {
		return fmt.Errorf("invalid request or group info")
	}

	switch req.Action {
	case "add":
		// Check if participants are already in the group
		for _, participant := range req.Participants {
			if currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is already in the group", participant)
			}
		}

	case "remove":
		// Check if participants are in the group
		for _, participant := range req.Participants {
			if !currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is not in the group", participant)
			}
			// Cannot remove group owner
			if participant == currentGroup.Owner {
				return ErrCannotRemoveOwner
			}
		}

	case "promote":
		// Check if participants are in the group and not already admins
		for _, participant := range req.Participants {
			if !currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is not in the group", participant)
			}
			if currentGroup.IsParticipantAdmin(participant) {
				return fmt.Errorf("participant %s is already an admin", participant)
			}
		}

	case "demote":
		// Check if participants are admins
		for _, participant := range req.Participants {
			if !currentGroup.HasParticipant(participant) {
				return fmt.Errorf("participant %s is not in the group", participant)
			}
			if !currentGroup.IsParticipantAdmin(participant) {
				return fmt.Errorf("participant %s is not an admin", participant)
			}
			// Cannot demote group owner
			if participant == currentGroup.Owner {
				return fmt.Errorf("cannot demote group owner")
			}
		}
	}

	return nil
}

// validateJID validates WhatsApp JID format using our improved validator
func (s *Service) validateJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("JID cannot be empty")
	}

	// Use the injected JID validator that handles multiple formats
	if s.jidValidator != nil && !s.jidValidator.IsValid(jid) {
		return fmt.Errorf("invalid JID format")
	}

	// Fallback to basic validation if no validator is provided
	if s.jidValidator == nil {
		jidPattern := `^[0-9]+@(s\.whatsapp\.net|g\.us)$`
		matched, err := regexp.MatchString(jidPattern, jid)
		if err != nil {
			return fmt.Errorf("error validating JID: %w", err)
		}
		if !matched {
			return fmt.Errorf("invalid JID format")
		}
	}

	return nil
}

// Note: Group operations are handled directly by the UseCase via WameowManager
// This service only provides validation functions
