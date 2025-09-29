package community

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Service defines the interface for community domain service
type Service interface {
	// Validation methods
	ValidateCommunityJID(jid string) error
	ValidateGroupJID(jid string) error
	ValidateLinkRequest(communityJID, groupJID string) error

	// Formatting methods
	FormatCommunityJID(jid string) string
	FormatGroupJID(jid string) string
	SanitizeCommunityName(name string) string
	SanitizeCommunityDescription(description string) string

	// Processing methods
	ProcessCommunityInfo(info *CommunityInfo) error
	ProcessLinkedGroups(groups []*LinkedGroup) error
	ProcessGroupLinkResult(result *GroupLinkInfo) error

	// Business logic methods
	CanLinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanUnlinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanViewCommunity(communityJID string, userJID string) error

	// Utility methods
	GenerateCommunityEvent(eventType CommunityEventType, communityJID, actorJID, targetJID string, data map[string]interface{}) *CommunityEvent
}

// serviceImpl implements the Service interface
type serviceImpl struct{}

// NewService creates a new community domain service
func NewService() Service {
	return &serviceImpl{}
}

// Validation methods

// ValidateCommunityJID validates a community JID format
func (s *serviceImpl) ValidateCommunityJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("community JID cannot be empty")
	}

	// Basic JID format validation
	if !strings.Contains(jid, "@") {
		return fmt.Errorf("invalid community JID format: missing @ symbol")
	}

	// Community JIDs should end with @g.us
	if !strings.HasSuffix(jid, "@g.us") {
		return fmt.Errorf("invalid community JID format: must end with @g.us")
	}

	// Extract the ID part (before @g.us)
	idPart := jid[:len(jid)-5]
	if len(idPart) == 0 {
		return fmt.Errorf("invalid community JID: empty ID part")
	}

	// Community IDs should be numeric and have a specific length
	matched, err := regexp.MatchString(`^\d{15,20}$`, idPart)
	if err != nil {
		return fmt.Errorf("error validating community JID: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid community JID format: ID must be 15-20 digits")
	}

	return nil
}

// ValidateGroupJID validates a group JID format
func (s *serviceImpl) ValidateGroupJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("group JID cannot be empty")
	}

	// Basic JID format validation
	if !strings.Contains(jid, "@") {
		return fmt.Errorf("invalid group JID format: missing @ symbol")
	}

	// Group JIDs should end with @g.us
	if !strings.HasSuffix(jid, "@g.us") {
		return fmt.Errorf("invalid group JID format: must end with @g.us")
	}

	// Extract the ID part (before @g.us)
	idPart := jid[:len(jid)-5]
	if len(idPart) == 0 {
		return fmt.Errorf("invalid group JID: empty ID part")
	}

	// Group IDs should be numeric and have a specific length
	matched, err := regexp.MatchString(`^\d{15,20}$`, idPart)
	if err != nil {
		return fmt.Errorf("error validating group JID: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid group JID format: ID must be 15-20 digits")
	}

	return nil
}

// ValidateLinkRequest validates a request to link a group to a community
func (s *serviceImpl) ValidateLinkRequest(communityJID, groupJID string) error {
	if err := s.ValidateCommunityJID(communityJID); err != nil {
		return fmt.Errorf("invalid community JID: %w", err)
	}

	if err := s.ValidateGroupJID(groupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	// Check if trying to link community to itself
	if communityJID == groupJID {
		return fmt.Errorf("cannot link community to itself")
	}

	return nil
}

// Formatting methods

// FormatCommunityJID ensures a community JID is properly formatted
func (s *serviceImpl) FormatCommunityJID(jid string) string {
	if jid == "" {
		return ""
	}

	// Remove any whitespace
	jid = strings.TrimSpace(jid)

	// If already properly formatted, return as is
	if strings.HasSuffix(jid, "@g.us") {
		return jid
	}

	// Add @g.us suffix if missing
	return jid + "@g.us"
}

// FormatGroupJID ensures a group JID is properly formatted
func (s *serviceImpl) FormatGroupJID(jid string) string {
	if jid == "" {
		return ""
	}

	// Remove any whitespace
	jid = strings.TrimSpace(jid)

	// If already properly formatted, return as is
	if strings.HasSuffix(jid, "@g.us") {
		return jid
	}

	// Add @g.us suffix if missing
	return jid + "@g.us"
}

// SanitizeCommunityName sanitizes a community name
func (s *serviceImpl) SanitizeCommunityName(name string) string {
	// Remove leading/trailing whitespace
	name = strings.TrimSpace(name)

	// Remove excessive whitespace
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	// Limit length (WhatsApp community names have limits)
	if len(name) > 25 {
		name = name[:25]
	}

	return name
}

// SanitizeCommunityDescription sanitizes a community description
func (s *serviceImpl) SanitizeCommunityDescription(description string) string {
	// Remove leading/trailing whitespace
	description = strings.TrimSpace(description)

	// Remove excessive whitespace
	description = regexp.MustCompile(`\s+`).ReplaceAllString(description, " ")

	// Limit length (WhatsApp community descriptions have limits)
	if len(description) > 512 {
		description = description[:512]
	}

	return description
}

// Processing methods

// ProcessCommunityInfo processes and validates community information
func (s *serviceImpl) ProcessCommunityInfo(info *CommunityInfo) error {
	if info == nil {
		return fmt.Errorf("community info cannot be nil")
	}

	// Validate and format JID
	if err := s.ValidateCommunityJID(info.JID); err != nil {
		return fmt.Errorf("invalid community JID: %w", err)
	}
	info.JID = s.FormatCommunityJID(info.JID)

	// Sanitize name and description
	info.Name = s.SanitizeCommunityName(info.Name)
	info.Description = s.SanitizeCommunityDescription(info.Description)

	// Validate counts
	if info.ParticipantCount < 0 {
		info.ParticipantCount = 0
	}
	if info.GroupCount < 0 {
		info.GroupCount = 0
	}

	return nil
}

// ProcessLinkedGroups processes and validates linked groups
func (s *serviceImpl) ProcessLinkedGroups(groups []*LinkedGroup) error {
	if groups == nil {
		return nil
	}

	for i, group := range groups {
		if group == nil {
			continue
		}

		// Validate and format group JID
		if err := s.ValidateGroupJID(group.JID); err != nil {
			return fmt.Errorf("invalid group JID at index %d: %w", i, err)
		}
		group.JID = s.FormatGroupJID(group.JID)

		// Sanitize name and description
		group.Name = s.SanitizeCommunityName(group.Name)
		group.Description = s.SanitizeCommunityDescription(group.Description)

		// Validate participant count
		if group.ParticipantCount < 0 {
			group.ParticipantCount = 0
		}
	}

	return nil
}

// ProcessGroupLinkResult processes the result of a group link operation
func (s *serviceImpl) ProcessGroupLinkResult(result *GroupLinkInfo) error {
	if result == nil {
		return fmt.Errorf("group link result cannot be nil")
	}

	// Validate and format JIDs
	if err := s.ValidateCommunityJID(result.CommunityJID); err != nil {
		return fmt.Errorf("invalid community JID: %w", err)
	}
	result.CommunityJID = s.FormatCommunityJID(result.CommunityJID)

	if err := s.ValidateGroupJID(result.GroupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}
	result.GroupJID = s.FormatGroupJID(result.GroupJID)

	// Set timestamp if not set
	if result.LinkedAt.IsZero() {
		result.LinkedAt = time.Now()
	}

	return nil
}

// Business logic methods

// CanLinkGroup checks if a user can link a group to a community
func (s *serviceImpl) CanLinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error {
	// Validate JIDs
	if err := s.ValidateLinkRequest(communityJID, groupJID); err != nil {
		return err
	}

	// Check user permissions
	if !isOwner && !isAdmin {
		return fmt.Errorf("insufficient permissions: only community owners and admins can link groups")
	}

	return nil
}

// CanUnlinkGroup checks if a user can unlink a group from a community
func (s *serviceImpl) CanUnlinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error {
	// Validate JIDs
	if err := s.ValidateLinkRequest(communityJID, groupJID); err != nil {
		return err
	}

	// Check user permissions
	if !isOwner && !isAdmin {
		return fmt.Errorf("insufficient permissions: only community owners and admins can unlink groups")
	}

	return nil
}

// CanViewCommunity checks if a user can view community information
func (s *serviceImpl) CanViewCommunity(communityJID string, userJID string) error {
	// Validate community JID
	if err := s.ValidateCommunityJID(communityJID); err != nil {
		return err
	}

	// For now, allow all users to view community info
	// This can be enhanced with actual permission checking
	return nil
}

// Utility methods

// GenerateCommunityEvent generates a community event
func (s *serviceImpl) GenerateCommunityEvent(eventType CommunityEventType, communityJID, actorJID, targetJID string, data map[string]interface{}) *CommunityEvent {
	return &CommunityEvent{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:         eventType,
		CommunityJID: communityJID,
		ActorJID:     actorJID,
		TargetJID:    targetJID,
		Data:         data,
		Timestamp:    time.Now(),
	}
}
