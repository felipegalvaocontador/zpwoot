package message

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/domain/message"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type UseCase interface {
	SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error)
	GetPollResults(ctx context.Context, req *GetPollResultsRequest) (*GetPollResultsResponse, error)
	RevokeMessage(ctx context.Context, req *RevokeMessageRequest) (*RevokeMessageResponse, error)
	EditMessage(ctx context.Context, req *EditMessageRequest) (*EditMessageResponse, error)
	MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error)
}

type useCaseImpl struct {
	sessionRepo    ports.SessionRepository
	wameowManager  ports.WameowManager
	mediaProcessor *message.MediaProcessor
	logger         *logger.Logger
}

func NewUseCase(
	sessionRepo ports.SessionRepository,
	wameowManager ports.WameowManager,
	logger *logger.Logger,
) UseCase {
	return &useCaseImpl{
		sessionRepo:    sessionRepo,
		wameowManager:  wameowManager,
		mediaProcessor: message.NewMediaProcessor(logger),
		logger:         logger,
	}
}

func (uc *useCaseImpl) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*SendMessageResponse, error) {
	uc.logger.InfoWithFields("Sending message", map[string]interface{}{
		"session_id": sessionID,
		"to":         req.RemoteJID,
		"type":       req.Type,
	})

	if err := uc.validateSession(ctx, sessionID); err != nil {
		return nil, err
	}

	domainReq := req.ToDomainRequest()
	if err := message.ValidateMessageRequest(domainReq); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	filePath, cleanup, err := uc.processMediaIfNeeded(ctx, domainReq)
	if err != nil {
		return nil, err
	}
	defer uc.cleanupMedia(cleanup, filePath)

	result, err := uc.sendMessageToWameow(sessionID, domainReq, filePath)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to send message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.RemoteJID,
			"type":       req.Type,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	uc.logger.InfoWithFields("Message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"to":         req.RemoteJID,
		"type":       req.Type,
		"message_id": result.MessageID,
	})

	return &SendMessageResponse{
		ID:        result.MessageID,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}, nil
}

func (uc *useCaseImpl) validateSession(ctx context.Context, sessionID string) error {
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if sess == nil {
		return fmt.Errorf("session not found")
	}

	if !sess.IsConnected {
		return fmt.Errorf("session is not connected")
	}

	return nil
}

func (uc *useCaseImpl) processMediaIfNeeded(ctx context.Context, domainReq *message.SendMessageRequest) (string, func() error, error) {
	if !domainReq.IsMediaMessage() || domainReq.File == "" {
		return "", nil, nil
	}

	processedMedia, err := uc.mediaProcessor.ProcessMediaForType(ctx, domainReq.File, domainReq.Type)
	if err != nil {
		return "", nil, fmt.Errorf("failed to process media: %w", err)
	}

	if domainReq.MimeType == "" {
		domainReq.MimeType = processedMedia.MimeType
	}

	if domainReq.Type == message.MessageTypeDocument && domainReq.Filename == "" {
		domainReq.Filename = "document"
	}

	return processedMedia.FilePath, processedMedia.Cleanup, nil
}

func (uc *useCaseImpl) cleanupMedia(cleanup func() error, filePath string) {
	if cleanup != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			uc.logger.WarnWithFields("Failed to cleanup temporary file", map[string]interface{}{
				"file_path": filePath,
				"error":     cleanupErr.Error(),
			})
		}
	}
}

func (uc *useCaseImpl) sendMessageToWameow(sessionID string, domainReq *message.SendMessageRequest, filePath string) (*message.SendResult, error) {
	var msgContextInfo *message.ContextInfo
	if domainReq.ContextInfo != nil {
		msgContextInfo = &message.ContextInfo{
			StanzaID:    domainReq.ContextInfo.StanzaID,
			Participant: domainReq.ContextInfo.Participant,
		}
	}

	return uc.wameowManager.SendMessage(
		sessionID,
		domainReq.To,
		string(domainReq.Type),
		domainReq.Body,
		domainReq.Caption,
		filePath,
		domainReq.Filename,
		domainReq.Latitude,
		domainReq.Longitude,
		domainReq.ContactName,
		domainReq.ContactPhone,
		msgContextInfo,
	)
}

func (uc *useCaseImpl) GetPollResults(ctx context.Context, req *GetPollResultsRequest) (*GetPollResultsResponse, error) {
	uc.logger.InfoWithFields("Getting poll results", map[string]interface{}{
		"to":              req.RemoteJID,
		"poll_message_id": req.PollMessageID,
	})


	return &GetPollResultsResponse{
		PollMessageID:         req.PollMessageID,
		PollName:              "Poll results not yet implemented",
		Options:               []PollOption{},
		TotalVotes:            0,
		SelectableOptionCount: 1,
		AllowMultipleAnswers:  false,
		RemoteJID:             req.RemoteJID,
	}, fmt.Errorf("poll results collection not yet implemented - requires event handling")
}

func (uc *useCaseImpl) RevokeMessage(ctx context.Context, req *RevokeMessageRequest) (*RevokeMessageResponse, error) {
	uc.logger.InfoWithFields("Revoking message", map[string]interface{}{
		"to":         req.RemoteJID,
		"message_id": req.MessageID,
	})

	result, err := uc.wameowManager.RevokeMessage(req.SessionID, req.RemoteJID, req.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke message: %w", err)
	}

	return &RevokeMessageResponse{
		ID:        result.MessageID,
		Status:    "revoked",
		Timestamp: result.Timestamp,
	}, nil
}

func (uc *useCaseImpl) EditMessage(ctx context.Context, req *EditMessageRequest) (*EditMessageResponse, error) {
	uc.logger.InfoWithFields("Editing message", map[string]interface{}{
		"to":         req.RemoteJID,
		"message_id": req.MessageID,
		"new_body":   req.NewBody,
	})

	err := uc.wameowManager.EditMessage(req.SessionID, req.RemoteJID, req.MessageID, req.NewBody)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	return &EditMessageResponse{
		ID:        req.MessageID,
		Status:    "edited",
		NewBody:   req.NewBody,
		Timestamp: time.Now(),
	}, nil
}

func (uc *useCaseImpl) MarkAsRead(ctx context.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error) {
	uc.logger.InfoWithFields("Marking messages as read", map[string]interface{}{
		"to":          req.RemoteJID,
		"message_ids": req.MessageIDs,
	})

	for _, messageID := range req.MessageIDs {
		err := uc.wameowManager.MarkRead(req.SessionID, req.RemoteJID, messageID)
		if err != nil {
			uc.logger.WarnWithFields("Failed to mark message as read", map[string]interface{}{
				"session_id": req.SessionID,
				"to":         req.RemoteJID,
				"message_id": messageID,
				"error":      err.Error(),
			})
		}
	}

	return &MarkAsReadResponse{
		MessageIDs: req.MessageIDs,
		Status:     "read",
		Timestamp:  time.Now(),
	}, nil
}
