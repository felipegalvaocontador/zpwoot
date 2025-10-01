package newsletter

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/newsletter"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type UseCase interface {
	CreateNewsletter(ctx context.Context, sessionID string, req *CreateNewsletterRequest) (*CreateNewsletterResponse, error)

	GetNewsletterInfo(ctx context.Context, sessionID string, req *GetNewsletterInfoRequest) (*NewsletterInfoResponse, error)

	GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, req *GetNewsletterInfoWithInviteRequest) (*NewsletterInfoResponse, error)

	FollowNewsletter(ctx context.Context, sessionID string, req *FollowNewsletterRequest) (*NewsletterActionResponse, error)

	UnfollowNewsletter(ctx context.Context, sessionID string, req *UnfollowNewsletterRequest) (*NewsletterActionResponse, error)

	GetSubscribedNewsletters(ctx context.Context, sessionID string) (*SubscribedNewslettersResponse, error)

	GetNewsletterMessages(ctx context.Context, sessionID string, req *GetNewsletterMessagesRequest) (*GetNewsletterMessagesResponse, error)

	GetNewsletterMessageUpdates(ctx context.Context, sessionID string, req *GetNewsletterMessageUpdatesRequest) (*GetNewsletterMessageUpdatesResponse, error)

	NewsletterMarkViewed(ctx context.Context, sessionID string, req *NewsletterMarkViewedRequest) (*NewsletterActionResponse, error)

	NewsletterSendReaction(ctx context.Context, sessionID string, req *NewsletterSendReactionRequest) (*NewsletterActionResponse, error)

	NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID string, req *NewsletterSubscribeLiveUpdatesRequest) (*NewsletterSubscribeLiveUpdatesResponse, error)

	NewsletterToggleMute(ctx context.Context, sessionID string, req *NewsletterToggleMuteRequest) (*NewsletterActionResponse, error)

	AcceptTOSNotice(ctx context.Context, sessionID string, req *AcceptTOSNoticeRequest) (*NewsletterActionResponse, error)

	UploadNewsletter(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error)

	UploadNewsletterReader(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error)
}

type useCaseImpl struct {
	newsletterManager ports.NewsletterManager
	newsletterService ports.NewsletterService
	sessionRepo       ports.SessionRepository
	logger            logger.Logger
}

func NewUseCase(
	newsletterManager ports.NewsletterManager,
	newsletterService ports.NewsletterService,
	sessionRepo ports.SessionRepository,
	logger logger.Logger,
) UseCase {
	return &useCaseImpl{
		newsletterManager: newsletterManager,
		newsletterService: newsletterService,
		sessionRepo:       sessionRepo,
		logger:            logger,
	}
}

func (uc *useCaseImpl) CreateNewsletter(ctx context.Context, sessionID string, req *CreateNewsletterRequest) (*CreateNewsletterResponse, error) {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid create newsletter request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !sess.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	name := uc.newsletterService.SanitizeNewsletterName(req.Name)
	description := uc.newsletterService.SanitizeNewsletterDescription(req.Description)

	uc.logger.InfoWithFields("Creating newsletter", map[string]interface{}{
		"session_id":  sessionID,
		"name":        name,
		"description": description,
	})

	newsletterInfo, err := uc.newsletterManager.CreateNewsletter(ctx, sessionID, name, description)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to create newsletter", map[string]interface{}{
			"session_id": sessionID,
			"name":       name,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to create newsletter: %w", err)
	}

	if err := uc.newsletterService.ProcessNewsletterInfo(newsletterInfo); err != nil {
		uc.logger.ErrorWithFields("Failed to process newsletter info", map[string]interface{}{
			"session_id":    sessionID,
			"newsletter_id": newsletterInfo.ID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to process newsletter info: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter created successfully", map[string]interface{}{
		"session_id":    sessionID,
		"newsletter_id": newsletterInfo.ID,
		"name":          newsletterInfo.Name,
	})

	return NewCreateNewsletterResponse(newsletterInfo), nil
}

func (uc *useCaseImpl) GetNewsletterInfo(ctx context.Context, sessionID string, req *GetNewsletterInfoRequest) (*NewsletterInfoResponse, error) {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid get newsletter info request", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	jid := uc.newsletterService.FormatNewsletterJID(req.NewsletterJID)

	return uc.getNewsletterInfoGeneric(ctx, sessionID, jid, "jid", uc.newsletterManager.GetNewsletterInfo)
}

func (uc *useCaseImpl) validateSessionAndConnection(ctx context.Context, sessionID string) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return fmt.Errorf("session is not connected")
	}

	return nil
}

func (uc *useCaseImpl) processNewsletterInfoCommon(sessionID string, newsletterInfo *newsletter.NewsletterInfo) error {
	if err := uc.newsletterService.ProcessNewsletterInfo(newsletterInfo); err != nil {
		uc.logger.ErrorWithFields("Failed to process newsletter info", map[string]interface{}{
			"session_id":    sessionID,
			"newsletter_id": newsletterInfo.ID,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to process newsletter info: %w", err)
	}
	return nil
}

func (uc *useCaseImpl) getNewsletterInfoGeneric(
	ctx context.Context,
	sessionID string,
	identifier string,
	identifierType string,
	getInfoFunc func(context.Context, string, string) (*newsletter.NewsletterInfo, error),
) (*NewsletterInfoResponse, error) {
	if err := uc.validateSessionAndConnection(ctx, sessionID); err != nil {
		return nil, err
	}

	uc.logger.InfoWithFields(fmt.Sprintf("Getting newsletter info via %s", identifierType), map[string]interface{}{
		"session_id":   sessionID,
		identifierType: identifier,
	})

	newsletterInfo, err := getInfoFunc(ctx, sessionID, identifier)
	if err != nil {
		uc.logger.ErrorWithFields(fmt.Sprintf("Failed to get newsletter info via %s", identifierType), map[string]interface{}{
			"session_id":   sessionID,
			identifierType: identifier,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter info via %s: %w", identifierType, err)
	}

	if err := uc.processNewsletterInfoCommon(sessionID, newsletterInfo); err != nil {
		return nil, err
	}

	uc.logger.InfoWithFields(fmt.Sprintf("Newsletter info retrieved via %s successfully", identifierType), map[string]interface{}{
		"session_id":    sessionID,
		"newsletter_id": newsletterInfo.ID,
		"name":          newsletterInfo.Name,
	})

	return NewNewsletterInfoResponse(newsletterInfo), nil
}

func (uc *useCaseImpl) validateNewsletterActionRequest(ctx context.Context, sessionID, jid string) (string, error) {
	if err := uc.validateSessionAndConnection(ctx, sessionID); err != nil {
		return "", err
	}

	formattedJID := uc.newsletterService.FormatNewsletterJID(jid)
	return formattedJID, nil
}

func (uc *useCaseImpl) newsletterActionGeneric(
	ctx context.Context,
	sessionID string,
	jid string,
	actionName string,
	actionFunc func(context.Context, string, string) error,
	responseFunc func(string) *NewsletterActionResponse,
) (*NewsletterActionResponse, error) {
	formattedJID, err := uc.validateNewsletterActionRequest(ctx, sessionID, jid)
	if err != nil {
		return nil, err
	}

	uc.logger.InfoWithFields(fmt.Sprintf("%s newsletter", actionName), map[string]interface{}{
		"session_id": sessionID,
		"jid":        formattedJID,
	})

	err = actionFunc(ctx, sessionID, formattedJID)
	if err != nil {
		uc.logger.ErrorWithFields(fmt.Sprintf("Failed to %s newsletter", actionName), map[string]interface{}{
			"session_id": sessionID,
			"jid":        formattedJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to %s newsletter: %w", actionName, err)
	}

	uc.logger.InfoWithFields(fmt.Sprintf("Newsletter %s successfully", actionName), map[string]interface{}{
		"session_id": sessionID,
		"jid":        formattedJID,
	})

	return responseFunc(formattedJID), nil
}

func (uc *useCaseImpl) uploadNewsletterGeneric(
	ctx context.Context,
	sessionID string,
	req *UploadNewsletterRequest,
	uploadType string,
	uploadFunc func(context.Context, string, []byte, string, string) (*newsletter.UploadNewsletterResponse, error),
) (*UploadNewsletterResponse, error) {
	uc.logger.InfoWithFields(fmt.Sprintf("Uploading newsletter media %s", uploadType), map[string]interface{}{
		"session_id": sessionID,
		"mime_type":  req.MimeType,
		"media_type": req.MediaType,
		"data_size":  len(req.Data),
	})

	if err := uc.validateUploadRequest(ctx, sessionID, req); err != nil {
		return nil, err
	}

	response, err := uploadFunc(ctx, sessionID, req.Data, req.MimeType, req.MediaType)
	if err != nil {
		uc.logger.ErrorWithFields(fmt.Sprintf("Failed to upload newsletter media %s", uploadType), map[string]interface{}{
			"session_id": sessionID,
			"mime_type":  req.MimeType,
			"media_type": req.MediaType,
			"data_size":  len(req.Data),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to upload newsletter media %s: %w", uploadType, err)
	}

	uc.logger.InfoWithFields(fmt.Sprintf("Newsletter media uploaded successfully %s", uploadType), map[string]interface{}{
		"session_id":  sessionID,
		"mime_type":   req.MimeType,
		"media_type":  req.MediaType,
		"data_size":   len(req.Data),
		"url":         response.URL,
		"handle":      response.Handle,
		"file_length": response.FileLength,
	})

	return &UploadNewsletterResponse{
		URL:        response.URL,
		DirectPath: response.DirectPath,
		Handle:     response.Handle,
		ObjectID:   response.ObjectID,
		FileSHA256: response.FileSHA256,
		FileLength: response.FileLength,
	}, nil
}

func (uc *useCaseImpl) validateUploadRequest(ctx context.Context, sessionID string, req *UploadNewsletterRequest) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}

func (uc *useCaseImpl) GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, req *GetNewsletterInfoWithInviteRequest) (*NewsletterInfoResponse, error) {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid get newsletter info with invite request", map[string]interface{}{
			"session_id": sessionID,
			"invite_key": req.InviteKey,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	inviteKey := uc.newsletterService.CleanInviteKey(req.InviteKey)

	return uc.getNewsletterInfoGeneric(ctx, sessionID, inviteKey, "invite_key", uc.newsletterManager.GetNewsletterInfoWithInvite)
}

func (uc *useCaseImpl) FollowNewsletter(ctx context.Context, sessionID string, req *FollowNewsletterRequest) (*NewsletterActionResponse, error) {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid follow newsletter request", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return uc.newsletterActionGeneric(ctx, sessionID, req.NewsletterJID, "follow", uc.newsletterManager.FollowNewsletter, NewSuccessFollowResponse)
}

func (uc *useCaseImpl) UnfollowNewsletter(ctx context.Context, sessionID string, req *UnfollowNewsletterRequest) (*NewsletterActionResponse, error) {
	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid unfollow newsletter request", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return uc.newsletterActionGeneric(ctx, sessionID, req.NewsletterJID, "unfollow", uc.newsletterManager.UnfollowNewsletter, NewSuccessUnfollowResponse)
}

func (uc *useCaseImpl) GetSubscribedNewsletters(ctx context.Context, sessionID string) (*SubscribedNewslettersResponse, error) {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !session.IsConnected {
		return nil, fmt.Errorf("session is not connected")
	}

	uc.logger.InfoWithFields("Getting subscribed newsletters", map[string]interface{}{
		"session_id": sessionID,
	})

	newsletters, err := uc.newsletterManager.GetSubscribedNewsletters(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get subscribed newsletters", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get subscribed newsletters: %w", err)
	}

	for _, newsletterInfo := range newsletters {
		if err := uc.newsletterService.ProcessNewsletterInfo(newsletterInfo); err != nil {
			uc.logger.WarnWithFields("Failed to process newsletter info", map[string]interface{}{
				"session_id":    sessionID,
				"newsletter_id": newsletterInfo.ID,
				"error":         err.Error(),
			})
		}
	}

	uc.logger.InfoWithFields("Subscribed newsletters retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"count":      len(newsletters),
	})

	return NewSubscribedNewslettersResponse(newsletters), nil
}

func (uc *useCaseImpl) GetNewsletterMessages(ctx context.Context, sessionID string, req *GetNewsletterMessagesRequest) (*GetNewsletterMessagesResponse, error) {
	uc.logger.InfoWithFields("Getting newsletter messages", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      req.Count,
		"before":     req.Before,
	})

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		uc.logger.ErrorWithFields("Session is nil", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil, fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	messages, err := uc.newsletterManager.GetNewsletterMessages(ctx, sessionID, req.NewsletterJID, req.Count, req.Before)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get newsletter messages", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter messages: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter messages retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(messages),
	})

	return NewGetNewsletterMessagesResponse(messages), nil
}

func (uc *useCaseImpl) GetNewsletterMessageUpdates(ctx context.Context, sessionID string, req *GetNewsletterMessageUpdatesRequest) (*GetNewsletterMessageUpdatesResponse, error) {
	uc.logger.InfoWithFields("Getting newsletter message updates", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      req.Count,
		"since":      req.Since,
		"after":      req.After,
	})

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		uc.logger.ErrorWithFields("Session is nil", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil, fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	updates, err := uc.newsletterManager.GetNewsletterMessageUpdates(ctx, sessionID, req.NewsletterJID, req.Count, req.Since, req.After)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get newsletter message updates", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter message updates: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter message updates retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(updates),
	})

	return NewGetNewsletterMessageUpdatesResponse(updates), nil
}

func (uc *useCaseImpl) NewsletterMarkViewed(ctx context.Context, sessionID string, req *NewsletterMarkViewedRequest) (*NewsletterActionResponse, error) {
	uc.logger.InfoWithFields("Marking newsletter messages as viewed", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(req.ServerIDs),
	})

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	err = uc.newsletterManager.NewsletterMarkViewed(ctx, sessionID, req.NewsletterJID, req.ServerIDs)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to mark newsletter messages as viewed", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to mark newsletter messages as viewed: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter messages marked as viewed successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"count":      len(req.ServerIDs),
	})

	return NewNewsletterActionResponse(req.NewsletterJID, "success", "Messages marked as viewed successfully"), nil
}

func (uc *useCaseImpl) NewsletterSendReaction(ctx context.Context, sessionID string, req *NewsletterSendReactionRequest) (*NewsletterActionResponse, error) {
	uc.logReactionRequest(sessionID, req)

	if err := uc.validateReactionRequest(ctx, sessionID, req); err != nil {
		return nil, err
	}

	serverID, err := uc.resolveServerID(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	return uc.sendReactionAndRespond(ctx, sessionID, req, serverID)
}

func (uc *useCaseImpl) logReactionRequest(sessionID string, req *NewsletterSendReactionRequest) {
	uc.logger.InfoWithFields("Sending newsletter reaction", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"server_id":  req.ServerID,
		"reaction":   req.Reaction,
	})
}

func (uc *useCaseImpl) validateReactionRequest(ctx context.Context, sessionID string, req *NewsletterSendReactionRequest) error {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}

func (uc *useCaseImpl) resolveServerID(ctx context.Context, sessionID string, req *NewsletterSendReactionRequest) (string, error) {
	serverID := req.ServerID
	if serverID != "" {
		return serverID, nil
	}

	uc.logger.InfoWithFields("ServerID not provided, looking up from MessageID", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"message_id": req.MessageID,
	})

	messages, err := uc.newsletterManager.GetNewsletterMessages(ctx, sessionID, req.NewsletterJID, 50, "")
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get newsletter messages for ServerID lookup", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("failed to lookup ServerID: %w", err)
	}

	for _, msg := range messages {
		if msg.ID == req.MessageID {
			serverID = msg.ServerID
			uc.logger.InfoWithFields("Found ServerID for MessageID", map[string]interface{}{
				"session_id": sessionID,
				"message_id": req.MessageID,
				"server_id":  serverID,
			})
			break
		}
	}

	if serverID == "" {
		return "", fmt.Errorf("could not find ServerID for MessageID %s in newsletter %s", req.MessageID, req.NewsletterJID)
	}

	return serverID, nil
}

func (uc *useCaseImpl) sendReactionAndRespond(ctx context.Context, sessionID string, req *NewsletterSendReactionRequest, serverID string) (*NewsletterActionResponse, error) {
	err := uc.newsletterManager.NewsletterSendReaction(ctx, sessionID, req.NewsletterJID, serverID, req.Reaction, req.MessageID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to send newsletter reaction", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"server_id":  serverID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to send newsletter reaction: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter reaction sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"server_id":  serverID,
		"message_id": req.MessageID,
		"reaction":   req.Reaction,
	})

	message := "Reaction sent successfully"
	if req.Reaction == "" {
		message = "Reaction removed successfully"
	}

	return NewNewsletterActionResponse(req.NewsletterJID, "success", message), nil
}

func (uc *useCaseImpl) NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID string, req *NewsletterSubscribeLiveUpdatesRequest) (*NewsletterSubscribeLiveUpdatesResponse, error) {
	uc.logger.InfoWithFields("Subscribing to newsletter live updates", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
	})

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	duration, err := uc.newsletterManager.NewsletterSubscribeLiveUpdates(ctx, sessionID, req.NewsletterJID)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to subscribe to newsletter live updates", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to subscribe to newsletter live updates: %w", err)
	}

	uc.logger.InfoWithFields("Subscribed to newsletter live updates successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"duration":   duration,
	})

	return &NewsletterSubscribeLiveUpdatesResponse{Duration: duration}, nil
}

func (uc *useCaseImpl) NewsletterToggleMute(ctx context.Context, sessionID string, req *NewsletterToggleMuteRequest) (*NewsletterActionResponse, error) {
	uc.logger.InfoWithFields("Toggling newsletter mute status", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"mute":       req.Mute,
	})

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	err = uc.newsletterManager.NewsletterToggleMute(ctx, sessionID, req.NewsletterJID, req.Mute)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to toggle newsletter mute status", map[string]interface{}{
			"session_id": sessionID,
			"jid":        req.NewsletterJID,
			"mute":       req.Mute,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to toggle newsletter mute status: %w", err)
	}

	uc.logger.InfoWithFields("Newsletter mute status toggled successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid":        req.NewsletterJID,
		"mute":       req.Mute,
	})

	message := "Newsletter muted successfully"
	if !req.Mute {
		message = "Newsletter unmuted successfully"
	}

	return NewNewsletterActionResponse(req.NewsletterJID, "success", message), nil
}

func (uc *useCaseImpl) AcceptTOSNotice(ctx context.Context, sessionID string, req *AcceptTOSNoticeRequest) (*NewsletterActionResponse, error) {
	uc.logger.InfoWithFields("Accepting TOS notice", map[string]interface{}{
		"session_id": sessionID,
		"notice_id":  req.NoticeID,
		"stage":      req.Stage,
	})

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.ErrorWithFields("Session not found", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	if err := req.Validate(); err != nil {
		uc.logger.ErrorWithFields("Invalid request", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	err = uc.newsletterManager.AcceptTOSNotice(ctx, sessionID, req.NoticeID, req.Stage)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to accept TOS notice", map[string]interface{}{
			"session_id": sessionID,
			"notice_id":  req.NoticeID,
			"stage":      req.Stage,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to accept TOS notice: %w", err)
	}

	uc.logger.InfoWithFields("TOS notice accepted successfully", map[string]interface{}{
		"session_id": sessionID,
		"notice_id":  req.NoticeID,
		"stage":      req.Stage,
	})

	return NewNewsletterActionResponse("", "success", "TOS notice accepted successfully"), nil
}

func (uc *useCaseImpl) UploadNewsletter(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error) {
	return uc.uploadNewsletterGeneric(ctx, sessionID, req, "", uc.newsletterManager.UploadNewsletter)
}

func (uc *useCaseImpl) UploadNewsletterReader(ctx context.Context, sessionID string, req *UploadNewsletterRequest) (*UploadNewsletterResponse, error) {
	return uc.uploadNewsletterGeneric(ctx, sessionID, req, "with reader", uc.newsletterManager.UploadNewsletterReader)
}
