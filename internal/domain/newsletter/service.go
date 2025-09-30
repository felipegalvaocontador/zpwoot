package newsletter

import (
	"context"
	"strings"
)

type Repository interface {
	GetNewsletter(ctx context.Context, id string) (*NewsletterInfo, error)
	SaveNewsletter(ctx context.Context, newsletter *NewsletterInfo) error
	DeleteNewsletter(ctx context.Context, id string) error
	ListNewsletters(ctx context.Context, userID string) ([]*NewsletterInfo, error)
}

type JIDValidator interface {
	IsValidJID(jid string) bool
	IsNewsletterJID(jid string) bool
	ParseJID(jid string) (string, error)
}

type Service struct {
	jidValidator JIDValidator
}

func NewService(jidValidator JIDValidator) *Service {
	return &Service{
		jidValidator: jidValidator,
	}
}

func (s *Service) ValidateNewsletterCreation(req *CreateNewsletterRequest) error {
	if req == nil {
		return ErrInvalidNewsletterName
	}

	return req.Validate()
}

func (s *Service) ValidateNewsletterName(name string) error {
	if name == "" {
		return ErrEmptyNewsletterName
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyNewsletterName
	}

	if len(name) > 25 {
		return ErrNewsletterNameTooLong
	}

	return nil
}

func (s *Service) ValidateNewsletterDescription(description string) error {
	if len(description) > 512 {
		return ErrDescriptionTooLong
	}

	return nil
}

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

func (s *Service) ValidateInviteKey(inviteKey string) error {
	if inviteKey == "" {
		return ErrInvalidInviteKey
	}

	cleanKey := strings.TrimSpace(inviteKey)
	cleanKey = strings.TrimPrefix(cleanKey, "https://whatsapp.com/channel/")
	cleanKey = strings.TrimPrefix(cleanKey, "whatsapp.com/channel/")
	cleanKey = strings.TrimPrefix(cleanKey, "channel/")

	if cleanKey == "" {
		return ErrInvalidInviteKey
	}

	return nil
}

func (s *Service) ValidateGetNewsletterInfoRequest(req *GetNewsletterInfoRequest) error {
	if req == nil {
		return ErrInvalidNewsletterJID
	}

	return s.ValidateNewsletterJID(req.JID)
}

func (s *Service) ValidateGetNewsletterInfoWithInviteRequest(req *GetNewsletterInfoWithInviteRequest) error {
	if req == nil {
		return ErrInvalidInviteKey
	}

	return s.ValidateInviteKey(req.InviteKey)
}

func (s *Service) ValidateFollowNewsletterRequest(req *FollowNewsletterRequest) error {
	if req == nil {
		return ErrInvalidNewsletterJID
	}

	return s.ValidateNewsletterJID(req.JID)
}

func (s *Service) ValidateUnfollowNewsletterRequest(req *UnfollowNewsletterRequest) error {
	if req == nil {
		return ErrInvalidNewsletterJID
	}

	return s.ValidateNewsletterJID(req.JID)
}

func (s *Service) ProcessNewsletterInfo(info *NewsletterInfo) error {
	if info == nil {
		return ErrNewsletterNotFound
	}

	if err := info.Validate(); err != nil {
		return err
	}

	return nil
}

func (s *Service) CanUserManageNewsletter(info *NewsletterInfo, userRole NewsletterRole) bool {
	if info == nil {
		return false
	}

	return userRole == NewsletterRoleAdmin || userRole == NewsletterRoleOwner
}

func (s *Service) CanUserFollowNewsletter(info *NewsletterInfo) bool {
	if info == nil {
		return false
	}

	return info.IsActive()
}

func (s *Service) SanitizeNewsletterName(name string) string {
	name = strings.TrimSpace(name)

	name = strings.Join(strings.Fields(name), " ")

	return name
}

func (s *Service) SanitizeNewsletterDescription(description string) string {
	description = strings.TrimSpace(description)

	return description
}

func (s *Service) CleanInviteKey(inviteKey string) string {
	inviteKey = strings.TrimSpace(inviteKey)

	inviteKey = strings.TrimPrefix(inviteKey, "https://whatsapp.com/channel/")
	inviteKey = strings.TrimPrefix(inviteKey, "whatsapp.com/channel/")
	inviteKey = strings.TrimPrefix(inviteKey, "channel/")

	return inviteKey
}

func (s *Service) FormatNewsletterJID(jid string) string {
	if strings.Contains(jid, "@newsletter") {
		return jid
	}

	if !strings.Contains(jid, "@") {
		return jid + "@newsletter"
	}

	return jid
}

func (s *Service) ExtractNewsletterID(jid string) string {
	if strings.Contains(jid, "@newsletter") {
		return strings.Split(jid, "@")[0]
	}

	return jid
}

func (s *Service) IsNewsletterOwner(info *NewsletterInfo) bool {
	if info == nil {
		return false
	}

	return info.IsOwner()
}

func (s *Service) IsNewsletterAdmin(info *NewsletterInfo) bool {
	if info == nil {
		return false
	}

	return info.IsAdmin()
}

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
