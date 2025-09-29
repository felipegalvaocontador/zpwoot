package contact

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/contact"
	"zpwoot/platform/logger"
)

// UseCase defines the interface for contact use cases
type UseCase interface {
	CheckWhatsApp(ctx context.Context, req *CheckWhatsAppRequest) (*CheckWhatsAppResponse, error)
	GetProfilePicture(ctx context.Context, req *GetProfilePictureRequest) (*ProfilePictureResponse, error)
	GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error)
	ListContacts(ctx context.Context, req *ListContactsRequest) (*ListContactsResponse, error)
	SyncContacts(ctx context.Context, req *SyncContactsRequest) (*SyncContactsResponse, error)
	GetBusinessProfile(ctx context.Context, req *GetBusinessProfileRequest) (*BusinessProfileResponse, error)
	GetContactStats(ctx context.Context, req *GetContactStatsRequest) (*GetContactStatsResponse, error)
}

type useCaseImpl struct {
	contactService contact.Service
	logger         *logger.Logger
}

// NewUseCase creates a new contact use case
func NewUseCase(contactService contact.Service, logger *logger.Logger) UseCase {
	return &useCaseImpl{
		contactService: contactService,
		logger:         logger,
	}
}

// CheckWhatsApp checks if phone numbers are registered on WhatsApp
func (uc *useCaseImpl) CheckWhatsApp(ctx context.Context, req *CheckWhatsAppRequest) (*CheckWhatsAppResponse, error) {
	uc.logger.InfoWithFields("Checking WhatsApp numbers", map[string]interface{}{
		"session_id":  req.SessionID,
		"phone_count": len(req.PhoneNumbers),
	})

	domainReq := &contact.CheckWhatsAppRequest{
		SessionID:    req.SessionID,
		PhoneNumbers: req.PhoneNumbers,
	}

	result, err := uc.contactService.CheckWhatsApp(ctx, domainReq)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to check WhatsApp numbers", map[string]interface{}{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Convert domain results to DTO
	dtoResults := make([]WhatsAppStatus, len(result.Results))
	for i, domainResult := range result.Results {
		dtoResults[i] = WhatsAppStatus{
			PhoneNumber:  domainResult.PhoneNumber,
			IsOnWhatsApp: domainResult.IsOnWhatsApp,
			JID:          domainResult.JID,
			IsBusiness:   domainResult.IsBusiness,
			VerifiedName: domainResult.VerifiedName,
		}
	}

	return &CheckWhatsAppResponse{
		Results: dtoResults,
		Total:   result.Total,
		Checked: result.Checked,
	}, nil
}

// GetProfilePicture gets profile picture information for a contact
func (uc *useCaseImpl) GetProfilePicture(ctx context.Context, req *GetProfilePictureRequest) (*ProfilePictureResponse, error) {
	domainReq := &contact.GetProfilePictureRequest{
		SessionID: req.SessionID,
		JID:       req.JID,
		Preview:   req.Preview,
	}

	result, err := uc.contactService.GetProfilePicture(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	return &ProfilePictureResponse{
		JID:        result.JID,
		URL:        result.URL,
		ID:         result.ID,
		Type:       result.Type,
		DirectPath: result.DirectPath,
		UpdatedAt:  result.UpdatedAt,
		HasPicture: result.HasPicture,
	}, nil
}

// GetUserInfo gets detailed information about WhatsApp users
func (uc *useCaseImpl) GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	domainReq := &contact.GetUserInfoRequest{
		SessionID: req.SessionID,
		JIDs:      req.JIDs,
	}

	result, err := uc.contactService.GetUserInfo(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain results to DTO
	dtoUsers := make([]UserInfo, len(result.Users))
	for i, domainUser := range result.Users {
		dtoUsers[i] = UserInfo{
			JID:          domainUser.JID,
			PhoneNumber:  domainUser.PhoneNumber,
			Name:         domainUser.Name,
			Status:       domainUser.Status,
			PictureID:    domainUser.PictureID,
			IsBusiness:   domainUser.IsBusiness,
			VerifiedName: domainUser.VerifiedName,
			IsContact:    domainUser.IsContact,
			LastSeen:     domainUser.LastSeen,
			IsOnline:     domainUser.IsOnline,
		}
	}

	return &GetUserInfoResponse{
		Users: dtoUsers,
		Total: result.Total,
		Found: result.Found,
	}, nil
}

// ListContacts lists contacts from the WhatsApp account
func (uc *useCaseImpl) ListContacts(ctx context.Context, req *ListContactsRequest) (*ListContactsResponse, error) {
	domainReq := &contact.ListContactsRequest{
		SessionID: req.SessionID,
		Limit:     req.Limit,
		Offset:    req.Offset,
		Search:    req.Search,
	}

	result, err := uc.contactService.ListContacts(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain results to DTO
	dtoContacts := make([]Contact, len(result.Contacts))
	for i, domainContact := range result.Contacts {
		dtoContacts[i] = Contact{
			JID:         domainContact.JID,
			PhoneNumber: domainContact.PhoneNumber,
			Name:        domainContact.Name,
			ShortName:   domainContact.ShortName,
			PushName:    domainContact.PushName,
			IsBusiness:  domainContact.IsBusiness,
			IsContact:   domainContact.IsContact,
			IsBlocked:   domainContact.IsBlocked,
			AddedAt:     domainContact.AddedAt,
			UpdatedAt:   domainContact.UpdatedAt,
		}
	}

	return &ListContactsResponse{
		Contacts: dtoContacts,
		Total:    result.Total,
		Limit:    result.Limit,
		Offset:   result.Offset,
		HasMore:  result.HasMore,
	}, nil
}

// SyncContacts synchronizes contacts from the device with WhatsApp
func (uc *useCaseImpl) SyncContacts(ctx context.Context, req *SyncContactsRequest) (*SyncContactsResponse, error) {
	domainReq := &contact.SyncContactsRequest{
		SessionID: req.SessionID,
		Force:     req.Force,
	}

	result, err := uc.contactService.SyncContacts(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("Contacts synchronized successfully: %d synced, %d added, %d updated, %d removed",
		result.Synced, result.Added, result.Updated, result.Removed)

	return &SyncContactsResponse{
		Synced:   result.Synced,
		Added:    result.Added,
		Updated:  result.Updated,
		Removed:  result.Removed,
		Total:    result.Total,
		SyncedAt: result.SyncedAt,
		Message:  message,
	}, nil
}

// GetBusinessProfile gets business profile information
func (uc *useCaseImpl) GetBusinessProfile(ctx context.Context, req *GetBusinessProfileRequest) (*BusinessProfileResponse, error) {
	domainReq := &contact.GetBusinessProfileRequest{
		SessionID: req.SessionID,
		JID:       req.JID,
	}

	result, err := uc.contactService.GetBusinessProfile(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	profile := BusinessProfile{
		JID:         result.Profile.JID,
		Name:        result.Profile.Name,
		Category:    result.Profile.Category,
		Description: result.Profile.Description,
		Website:     result.Profile.Website,
		Email:       result.Profile.Email,
		Address:     result.Profile.Address,
		Verified:    result.Profile.Verified,
	}

	return &BusinessProfileResponse{
		Profile:   profile,
		Found:     result.Found,
		UpdatedAt: result.UpdatedAt,
	}, nil
}

// GetContactStats gets statistics about contacts
func (uc *useCaseImpl) GetContactStats(ctx context.Context, req *GetContactStatsRequest) (*GetContactStatsResponse, error) {
	domainReq := &contact.GetContactStatsRequest{
		SessionID: req.SessionID,
	}

	result, err := uc.contactService.GetContactStats(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	stats := ContactStats{
		TotalContacts:    result.Stats.TotalContacts,
		WhatsAppContacts: result.Stats.WhatsAppContacts,
		BusinessContacts: result.Stats.BusinessContacts,
		BlockedContacts:  result.Stats.BlockedContacts,
		SyncRate:         result.Stats.SyncRate,
		LastSyncAt:       result.Stats.LastSyncAt,
	}

	return &GetContactStatsResponse{
		SessionID: req.SessionID,
		Stats:     stats,
		UpdatedAt: result.UpdatedAt,
	}, nil
}
