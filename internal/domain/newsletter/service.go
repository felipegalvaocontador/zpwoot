package newsletter

import (
	"context"
	"strings"
)

// Repository defines the interface for newsletter data persistence
type Repository interface {
	// Note: For now, newsletters are managed entirely through WhatsApp
	// This interface is kept for future extensions if we need local storage
	GetNewsletter(ctx context.Context, id string) (*NewsletterInfo, error)
	SaveNewsletter(ctx context.Context, newsletter *NewsletterInfo) error
	DeleteNewsletter(ctx context.Context, id string) error
	ListNewsletters(ctx context.Context, userID string) ([]*NewsletterInfo, error)
}

// JIDValidator defines the interface for JID validation
type JIDValidator interface {
	IsValidJID(jid string) bool
	IsNewsletterJID(jid string) bool
	ParseJID(jid string) (string, error)
}

// Service provides newsletter domain business logic
type Service struct {
	jidValidator JIDValidator
}

// NewService creates a new newsletter service
func NewService(jidValidator JIDValidator) *Service {
	return &Service{
		jidValidator: jidValidator,
	}
}

// ValidateNewsletterCreation validates newsletter creation request
func (s *Service) ValidateNewsletterCreation(req *CreateNewsletterRequest) error {
	if req == nil {
		return ErrInvalidNewsletterName
	}

	return req.Validate()
}

// ValidateNewsletterName validates newsletter name
func (s *Service) ValidateNewsletterName(name string) error {
	if name == "" {
		return ErrEmptyNewsletterName
	}

	// Trim whitespace
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyNewsletterName
	}

	if len(name) > 25 {
		return ErrNewsletterNameTooLong
	}

	return nil
}

// ValidateNewsletterDescription validates newsletter description
func (s *Service) ValidateNewsletterDescription(description string) error {
	if len(description) > 512 {
		return ErrDescriptionTooLong
	}

	return nil
}

// ValidateNewsletterJID validates newsletter JID
func (s *Service) ValidateNewsletterJID(jid string) error {
	if jid == "" {
		return ErrInvalidNewsletterJID
	}

	if !IsValidNewsletterJID(jid) {
		return ErrInvalidNewsletterJID
	}

	if s.jidValidator != nil && !s.jidValidator.IsNewsletterJID(jid) {
		return ErrInvalidNewsletterJID
	}

	return nil
}

// ValidateInviteKey validates newsletter invite key
func (s *Service) ValidateInviteKey(inviteKey string) error {
	if inviteKey == "" {
		return ErrInvalidInviteKey
	}

	// Clean up the invite key by removing common prefixes
	cleanKey := strings.TrimSpace(inviteKey)
	cleanKey = strings.TrimPrefix(cleanKey, "https://whatsapp.com/channel/")
	cleanKey = strings.TrimPrefix(cleanKey, "whatsapp.com/channel/")
	cleanKey = strings.TrimPrefix(cleanKey, "channel/")

	if cleanKey == "" {
		return ErrInvalidInviteKey
	}

	return nil
}

// ValidateGetNewsletterInfoRequest validates get newsletter info request
func (s *Service) ValidateGetNewsletterInfoRequest(req *GetNewsletterInfoRequest) error {
	if req == nil {
		return ErrInvalidNewsletterJID
	}

	return s.ValidateNewsletterJID(req.JID)
}

// ValidateGetNewsletterInfoWithInviteRequest validates get newsletter info with invite request
func (s *Service) ValidateGetNewsletterInfoWithInviteRequest(req *GetNewsletterInfoWithInviteRequest) error {
	if req == nil {
		return ErrInvalidInviteKey
	}

	return s.ValidateInviteKey(req.InviteKey)
}

// ValidateFollowNewsletterRequest validates follow newsletter request
func (s *Service) ValidateFollowNewsletterRequest(req *FollowNewsletterRequest) error {
	if req == nil {
		return ErrInvalidNewsletterJID
	}

	return s.ValidateNewsletterJID(req.JID)
}

// ValidateUnfollowNewsletterRequest validates unfollow newsletter request
func (s *Service) ValidateUnfollowNewsletterRequest(req *UnfollowNewsletterRequest) error {
	if req == nil {
		return ErrInvalidNewsletterJID
	}

	return s.ValidateNewsletterJID(req.JID)
}

// ProcessNewsletterInfo processes and validates newsletter information
func (s *Service) ProcessNewsletterInfo(info *NewsletterInfo) error {
	if info == nil {
		return ErrNewsletterNotFound
	}

	// Validate the newsletter info
	if err := info.Validate(); err != nil {
		return err
	}

	// Additional business logic can be added here
	// For example: checking permissions, applying business rules, etc.

	return nil
}

// CanUserManageNewsletter checks if a user can manage a newsletter
func (s *Service) CanUserManageNewsletter(info *NewsletterInfo, userRole NewsletterRole) bool {
	if info == nil {
		return false
	}

	// Only admins and owners can manage newsletters
	return userRole == NewsletterRoleAdmin || userRole == NewsletterRoleOwner
}

// CanUserFollowNewsletter checks if a user can follow a newsletter
func (s *Service) CanUserFollowNewsletter(info *NewsletterInfo) bool {
	if info == nil {
		return false
	}

	// Users can follow active newsletters
	return info.IsActive()
}

// SanitizeNewsletterName sanitizes newsletter name
func (s *Service) SanitizeNewsletterName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Remove excessive whitespace
	name = strings.Join(strings.Fields(name), " ")

	return name
}

// SanitizeNewsletterDescription sanitizes newsletter description
func (s *Service) SanitizeNewsletterDescription(description string) string {
	// Trim whitespace
	description = strings.TrimSpace(description)

	return description
}

// CleanInviteKey cleans and normalizes invite key
func (s *Service) CleanInviteKey(inviteKey string) string {
	// Trim whitespace
	inviteKey = strings.TrimSpace(inviteKey)

	// Remove common prefixes
	inviteKey = strings.TrimPrefix(inviteKey, "https://whatsapp.com/channel/")
	inviteKey = strings.TrimPrefix(inviteKey, "whatsapp.com/channel/")
	inviteKey = strings.TrimPrefix(inviteKey, "channel/")

	return inviteKey
}

// FormatNewsletterJID ensures JID has correct newsletter format
func (s *Service) FormatNewsletterJID(jid string) string {
	if strings.Contains(jid, "@newsletter") {
		return jid
	}

	// If it's just the ID part, add the newsletter suffix
	if !strings.Contains(jid, "@") {
		return jid + "@newsletter"
	}

	return jid
}

// ExtractNewsletterID extracts the ID part from a newsletter JID
func (s *Service) ExtractNewsletterID(jid string) string {
	if strings.Contains(jid, "@newsletter") {
		return strings.Split(jid, "@")[0]
	}

	return jid
}

// IsNewsletterOwner checks if the user is the owner of the newsletter
func (s *Service) IsNewsletterOwner(info *NewsletterInfo) bool {
	if info == nil {
		return false
	}

	return info.IsOwner()
}

// IsNewsletterAdmin checks if the user is an admin of the newsletter
func (s *Service) IsNewsletterAdmin(info *NewsletterInfo) bool {
	if info == nil {
		return false
	}

	return info.IsAdmin()
}

// GetNewsletterPermissions returns the permissions for a newsletter
func (s *Service) GetNewsletterPermissions(info *NewsletterInfo) map[string]bool {
	if info == nil {
		return map[string]bool{
			"canManage":   false,
			"canFollow":   false,
			"canUnfollow": false,
			"canView":     false,
		}
	}

	return map[string]bool{
		"canManage":   info.CanManage(),
		"canFollow":   info.IsActive() && info.Role == NewsletterRoleGuest,
		"canUnfollow": info.Role == NewsletterRoleSubscriber,
		"canView":     true,
	}
}
