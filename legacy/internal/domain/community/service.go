package community

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Service interface {
	ValidateCommunityJID(jid string) error
	ValidateGroupJID(jid string) error
	ValidateLinkRequest(communityJID, groupJID string) error

	FormatCommunityJID(jid string) string
	FormatGroupJID(jid string) string
	SanitizeCommunityName(name string) string
	SanitizeCommunityDescription(description string) string

	ProcessCommunityInfo(info *CommunityInfo) error
	ProcessLinkedGroups(groups []*LinkedGroup) error
	ProcessGroupLinkResult(result *GroupLinkInfo) error

	CanLinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanUnlinkGroup(communityJID, groupJID string, userJID string, isOwner, isAdmin bool) error
	CanViewCommunity(communityJID string, userJID string) error

	GenerateCommunityEvent(eventType CommunityEventType, communityJID, actorJID, targetJID string, data map[string]interface{}) *CommunityEvent
}

type serviceImpl struct{}

func NewService() Service {
	return &serviceImpl{}
}

func (s *serviceImpl) ValidateCommunityJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("community JID cannot be empty")
	}

	if !strings.Contains(jid, "@") {
		return fmt.Errorf("invalid community JID format: missing @ symbol")
	}

	if !strings.HasSuffix(jid, "@g.us") {
		return fmt.Errorf("invalid community JID format: must end with @g.us")
	}

	idPart := jid[:len(jid)-5]
	if idPart == "" {
		return fmt.Errorf("invalid community JID: empty ID part")
	}

	matched, err := regexp.MatchString(`^\d{15,20}$`, idPart)
	if err != nil {
		return fmt.Errorf("error validating community JID: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid community JID format: ID must be 15-20 digits")
	}

	return nil
}

func (s *serviceImpl) ValidateGroupJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("group JID cannot be empty")
	}

	if !strings.Contains(jid, "@") {
		return fmt.Errorf("invalid group JID format: missing @ symbol")
	}

	if !strings.HasSuffix(jid, "@g.us") {
		return fmt.Errorf("invalid group JID format: must end with @g.us")
	}

	idPart := jid[:len(jid)-5]
	if idPart == "" {
		return fmt.Errorf("invalid group JID: empty ID part")
	}

	matched, err := regexp.MatchString(`^\d{15,20}$`, idPart)
	if err != nil {
		return fmt.Errorf("error validating group JID: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid group JID format: ID must be 15-20 digits")
	}

	return nil
}

func (s *serviceImpl) ValidateLinkRequest(communityJID, groupJID string) error {
	if err := s.ValidateCommunityJID(communityJID); err != nil {
		return fmt.Errorf("invalid community JID: %w", err)
	}

	if err := s.ValidateGroupJID(groupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if communityJID == groupJID {
		return fmt.Errorf("cannot link community to itself")
	}

	return nil
}

func (s *serviceImpl) FormatCommunityJID(jid string) string {
	if jid == "" {
		return ""
	}

	jid = strings.TrimSpace(jid)

	if strings.HasSuffix(jid, "@g.us") {
		return jid
	}

	return jid + "@g.us"
}

func (s *serviceImpl) FormatGroupJID(jid string) string {
	if jid == "" {
		return ""
	}

	jid = strings.TrimSpace(jid)

	if strings.HasSuffix(jid, "@g.us") {
		return jid
	}

	return jid + "@g.us"
}

func (s *serviceImpl) SanitizeCommunityName(name string) string {
	name = strings.TrimSpace(name)

	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	if len(name) > 25 {
		name = name[:25]
	}

	return name
}

func (s *serviceImpl) SanitizeCommunityDescription(description string) string {
	description = strings.TrimSpace(description)

	description = regexp.MustCompile(`\s+`).ReplaceAllString(description, " ")

	if len(description) > 512 {
		description = description[:512]
	}

	return description
}

func (s *serviceImpl) ProcessCommunityInfo(info *CommunityInfo) error {
	if info == nil {
		return fmt.Errorf("community info cannot be nil")
	}

	if err := s.ValidateCommunityJID(info.JID); err != nil {
		return fmt.Errorf("invalid community JID: %w", err)
	}
	info.JID = s.FormatCommunityJID(info.JID)

	info.Name = s.SanitizeCommunityName(info.Name)
	info.Description = s.SanitizeCommunityDescription(info.Description)

	if info.ParticipantCount < 0 {
		info.ParticipantCount = 0
	}
	if info.GroupCount < 0 {
		info.GroupCount = 0
	}

	return nil
}

func (s *serviceImpl) ProcessLinkedGroups(groups []*LinkedGroup) error {
	if groups == nil {
		return nil
	}

	for i, group := range groups {
		if group == nil {
			continue
		}

		if err := s.ValidateGroupJID(group.JID); err != nil {
			return fmt.Errorf("invalid group JID at index %d: %w", i, err)
		}
		group.JID = s.FormatGroupJID(group.JID)

		group.Name = s.SanitizeCommunityName(group.Name)
		group.Description = s.SanitizeCommunityDescription(group.Description)

		if group.ParticipantCount < 0 {
			group.ParticipantCount = 0
		}
	}

	return nil
}

func (s *serviceImpl) ProcessGroupLinkResult(result *GroupLinkInfo) error {
	if result == nil {
		return fmt.Errorf("group link result cannot be nil")
	}

	if err := s.ValidateCommunityJID(result.CommunityJID); err != nil {
		return fmt.Errorf("invalid community JID: %w", err)
	}
	result.CommunityJID = s.FormatCommunityJID(result.CommunityJID)

	if err := s.ValidateGroupJID(result.GroupJID); err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}
	result.GroupJID = s.FormatGroupJID(result.GroupJID)

	if result.LinkedAt.IsZero() {
		result.LinkedAt = time.Now()
	}

	return nil
}

func (s *serviceImpl) CanLinkGroup(communityJID, groupJID, userJID string, isOwner, isAdmin bool) error {
	if err := s.ValidateLinkRequest(communityJID, groupJID); err != nil {
		return err
	}

	if !isOwner && !isAdmin {
		return fmt.Errorf("insufficient permissions: only community owners and admins can link groups")
	}

	return nil
}

func (s *serviceImpl) CanUnlinkGroup(communityJID, groupJID, userJID string, isOwner, isAdmin bool) error {
	if err := s.ValidateLinkRequest(communityJID, groupJID); err != nil {
		return err
	}

	if !isOwner && !isAdmin {
		return fmt.Errorf("insufficient permissions: only community owners and admins can unlink groups")
	}

	return nil
}

func (s *serviceImpl) CanViewCommunity(communityJID, userJID string) error {
	if err := s.ValidateCommunityJID(communityJID); err != nil {
		return err
	}

	return nil
}

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
