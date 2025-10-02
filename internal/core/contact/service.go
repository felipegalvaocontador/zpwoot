package contact

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/platform/logger"
)

type ProfilePictureInfo struct {
	JID        string     `json:"jid"`
	HasPicture bool       `json:"has_picture"`
	URL        string     `json:"url,omitempty"`
	ID         string     `json:"id,omitempty"`
	Type       string     `json:"type,omitempty"`
	DirectPath string     `json:"direct_path,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type UserInfo struct {
	JID          string     `json:"jid"`
	PhoneNumber  string     `json:"phone_number"`
	Name         string     `json:"name,omitempty"`
	Status       string     `json:"status,omitempty"`
	PictureID    string     `json:"picture_id,omitempty"`
	IsBusiness   bool       `json:"is_business"`
	VerifiedName string     `json:"verified_name,omitempty"`
	IsContact    bool       `json:"is_contact"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	IsOnline     bool       `json:"is_online"`
}

type ContactInfo struct {
	JID          string `json:"jid"`
	PhoneNumber  string `json:"phone_number"`
	Name         string `json:"name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBusiness   bool   `json:"is_business"`
	IsContact    bool   `json:"is_contact"`
}

type BusinessProfile struct {
	JID          string `json:"jid"`
	IsBusiness   bool   `json:"is_business"`
	BusinessName string `json:"business_name,omitempty"`
	Category     string `json:"category,omitempty"`
	Description  string `json:"description,omitempty"`
	Website      string `json:"website,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
}

type ContactGateway interface {
	IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]bool, error)

	GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (*ProfilePictureInfo, error)
	GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]*UserInfo, error)

	GetAllContacts(ctx context.Context, sessionID string) ([]*ContactInfo, error)

	GetBusinessProfile(ctx context.Context, sessionID, jid string) (*BusinessProfile, error)
}

type ContactRepository interface {
	Create(ctx context.Context, contact *Contact) error
	Update(ctx context.Context, contact *Contact) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Contact, error)
	GetByJID(ctx context.Context, sessionID uuid.UUID, jid string) (*Contact, error)
	List(ctx context.Context, req *ListContactsRequest) ([]*Contact, int64, error)
	GetStats(ctx context.Context, sessionID uuid.UUID) (*ContactStats, error)
	UpdateSyncStatus(ctx context.Context, req *UpdateSyncStatusRequest) error
}

type Service struct {
	gateway    ContactGateway
	repository ContactRepository
	logger     *logger.Logger
}

func NewService(gateway ContactGateway, repository ContactRepository, logger *logger.Logger) *Service {
	return &Service{
		gateway:    gateway,
		repository: repository,
		logger:     logger,
	}
}

func (s *Service) CheckWhatsApp(ctx context.Context, sessionID string, req *contracts.CheckWhatsAppRequest) (*contracts.CheckWhatsAppResponse, error) {
	s.logger.InfoWithFields("Checking WhatsApp numbers", map[string]interface{}{
		"session_id":  sessionID,
		"phone_count": len(req.PhoneNumbers),
	})

	if len(req.PhoneNumbers) == 0 {
		return nil, fmt.Errorf("no phone numbers provided")
	}
	if len(req.PhoneNumbers) > 50 {
		return nil, fmt.Errorf("maximum 50 phone numbers allowed")
	}

	results, err := s.gateway.IsOnWhatsApp(ctx, sessionID, req.PhoneNumbers)
	if err != nil {
		s.logger.ErrorWithFields("Failed to check WhatsApp numbers", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	checkResults := make([]contracts.WhatsAppCheckResult, 0, len(req.PhoneNumbers))
	foundCount := 0

	for _, phone := range req.PhoneNumbers {
		isOnWhatsApp := results[phone]
		if isOnWhatsApp {
			foundCount++
		}

		result := contracts.WhatsAppCheckResult{
			PhoneNumber:  phone,
			IsOnWhatsApp: isOnWhatsApp,
		}

		if isOnWhatsApp {

			cleanPhone := s.cleanPhoneNumber(phone)
			result.JID = cleanPhone + "@s.whatsapp.net"
		}

		checkResults = append(checkResults, result)
	}

	response := &contracts.CheckWhatsAppResponse{
		Results: checkResults,
		Total:   len(req.PhoneNumbers),
		Found:   foundCount,
		Success: true,
		Message: fmt.Sprintf("Checked %d numbers, found %d on WhatsApp", len(req.PhoneNumbers), foundCount),
	}

	s.logger.InfoWithFields("WhatsApp numbers checked successfully", map[string]interface{}{
		"session_id": sessionID,
		"total":      response.Total,
		"found":      response.Found,
	})

	return response, nil
}

func (s *Service) GetProfilePictureInfo(ctx context.Context, sessionID string, req *contracts.GetProfilePictureInfoRequest) (*contracts.GetProfilePictureInfoResponse, error) {
	s.logger.InfoWithFields("Getting profile picture info", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.JID,
		"preview":    req.Preview,
	})

	info, err := s.gateway.GetProfilePictureInfo(ctx, sessionID, req.JID, req.Preview)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get profile picture info", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.JID,
			"error":      err.Error(),
		})
		return nil, err
	}

	response := &contracts.GetProfilePictureInfoResponse{
		JID:        info.JID,
		HasPicture: info.HasPicture,
		URL:        info.URL,
		ID:         info.ID,
		Type:       info.Type,
		DirectPath: info.DirectPath,
		UpdatedAt:  info.UpdatedAt,
		Success:    true,
		Message:    "Profile picture info retrieved successfully",
	}

	return response, nil
}

func (s *Service) GetUserInfo(ctx context.Context, sessionID string, req *contracts.GetUserInfoRequest) (*contracts.GetUserInfoResponse, error) {
	s.logger.InfoWithFields("Getting user info", map[string]interface{}{
		"session_id": sessionID,
		"jid_count":  len(req.JIDs),
	})

	if len(req.JIDs) == 0 {
		return nil, fmt.Errorf("no JIDs provided")
	}
	if len(req.JIDs) > 20 {
		return nil, fmt.Errorf("maximum 20 JIDs allowed")
	}

	users, err := s.gateway.GetUserInfo(ctx, sessionID, req.JIDs)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get user info", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	userInfos := make([]contracts.UserInfo, 0, len(users))
	for _, user := range users {
		userInfo := contracts.UserInfo{
			JID:          user.JID,
			PhoneNumber:  user.PhoneNumber,
			Name:         user.Name,
			Status:       user.Status,
			PictureID:    user.PictureID,
			IsBusiness:   user.IsBusiness,
			VerifiedName: user.VerifiedName,
			IsContact:    user.IsContact,
			LastSeen:     user.LastSeen,
			IsOnline:     user.IsOnline,
		}
		userInfos = append(userInfos, userInfo)
	}

	response := &contracts.GetUserInfoResponse{
		Users:   userInfos,
		Total:   len(req.JIDs),
		Found:   len(users),
		Success: true,
		Message: fmt.Sprintf("Retrieved info for %d users", len(users)),
	}

	return response, nil
}

func (s *Service) ListContacts(ctx context.Context, sessionID string, req *contracts.ListContactsRequest) (*contracts.ListContactsResponse, error) {
	s.logger.InfoWithFields("Listing contacts", map[string]interface{}{
		"session_id": sessionID,
		"limit":      req.Limit,
		"offset":     req.Offset,
	})

	listReq := &ListContactsRequest{
		SessionID: sessionID,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	if listReq.Limit == 0 {
		listReq.Limit = 50
	}

	contacts, total, err := s.repository.List(ctx, listReq)
	if err != nil {
		s.logger.ErrorWithFields("Failed to list contacts", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	contactInfos := make([]contracts.ContactDetails, 0, len(contacts))
	for _, contact := range contacts {
		contactInfo := contracts.ContactDetails{
			JID:          contact.ZpJID,
			PhoneNumber:  contact.PhoneNumber,
			Name:         contact.GetDisplayName(),
			BusinessName: "",
			IsBusiness:   contact.IsBusiness,
			IsContact:    true,
			IsBlocked:    contact.IsBlocked,
		}
		contactInfos = append(contactInfos, contactInfo)
	}

	response := &contracts.ListContactsResponse{
		Contacts: contactInfos,
		Total:    int(total),
		Limit:    listReq.Limit,
		Offset:   listReq.Offset,
		Success:  true,
		Message:  fmt.Sprintf("Retrieved %d contacts", len(contactInfos)),
	}

	return response, nil
}

func (s *Service) SyncContacts(ctx context.Context, sessionID string, req *contracts.SyncContactsRequest) (*contracts.SyncContactsResponse, error) {
	s.logger.InfoWithFields("Syncing contacts", map[string]interface{}{
		"session_id": sessionID,
		"force":      req.Force,
	})

	whatsappContacts, err := s.gateway.GetAllContacts(ctx, sessionID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get WhatsApp contacts", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	syncedCount := 0
	newCount := 0
	updatedCount := 0

	for _, whatsappContact := range whatsappContacts {

		existingContact, err := s.repository.GetByJID(ctx, sessionUUID, whatsappContact.JID)
		if err != nil && err.Error() != "contact not found" {
			s.logger.WarnWithFields("Error checking existing contact", map[string]interface{}{
				"jid":   whatsappContact.JID,
				"error": err.Error(),
			})
			continue
		}

		if existingContact == nil {

			newContact := &Contact{
				ID:          uuid.New(),
				SessionID:   sessionUUID,
				ZpJID:       whatsappContact.JID,
				ZpName:      whatsappContact.Name,
				ZpPushName:  whatsappContact.Name,
				PhoneNumber: whatsappContact.PhoneNumber,
				IsBusiness:  whatsappContact.IsBusiness,
				SyncStatus:  string(SyncStatusPending),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			err = s.repository.Create(ctx, newContact)
			if err != nil {
				s.logger.WarnWithFields("Failed to create contact", map[string]interface{}{
					"jid":   whatsappContact.JID,
					"error": err.Error(),
				})
				continue
			}

			newCount++
		} else {

			existingContact.ZpName = whatsappContact.Name
			existingContact.ZpPushName = whatsappContact.Name
			existingContact.IsBusiness = whatsappContact.IsBusiness
			existingContact.UpdatedAt = time.Now()

			err = s.repository.Update(ctx, existingContact)
			if err != nil {
				s.logger.WarnWithFields("Failed to update contact", map[string]interface{}{
					"jid":   whatsappContact.JID,
					"error": err.Error(),
				})
				continue
			}

			updatedCount++
		}

		syncedCount++
	}

	response := &contracts.SyncContactsResponse{
		TotalContacts: len(whatsappContacts),
		SyncedCount:   syncedCount,
		NewCount:      newCount,
		UpdatedCount:  updatedCount,
		Success:       true,
		Message:       fmt.Sprintf("Synced %d contacts (%d new, %d updated)", syncedCount, newCount, updatedCount),
	}

	s.logger.InfoWithFields("Contacts synced successfully", map[string]interface{}{
		"session_id": sessionID,
		"total":      response.TotalContacts,
		"synced":     response.SyncedCount,
		"new":        response.NewCount,
		"updated":    response.UpdatedCount,
	})

	return response, nil
}

func (s *Service) GetBusinessProfile(ctx context.Context, sessionID string, req *contracts.GetBusinessProfileRequest) (*contracts.GetBusinessProfileResponse, error) {
	s.logger.InfoWithFields("Getting business profile", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.JID,
	})

	profile, err := s.gateway.GetBusinessProfile(ctx, sessionID, req.JID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get business profile", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.JID,
			"error":      err.Error(),
		})
		return nil, err
	}

	response := &contracts.GetBusinessProfileResponse{
		JID:          profile.JID,
		IsBusiness:   profile.IsBusiness,
		BusinessName: profile.BusinessName,
		Category:     profile.Category,
		Description:  profile.Description,
		Website:      profile.Website,
		Email:        profile.Email,
		Address:      profile.Address,
		Success:      true,
		Message:      "Business profile retrieved successfully",
	}

	return response, nil
}

func (s *Service) cleanPhoneNumber(phone string) string {
	cleaned := strings.ReplaceAll(phone, "+", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	return cleaned
}
