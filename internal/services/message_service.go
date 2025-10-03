package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/session"
	"zpwoot/internal/services/shared/validation"
	"zpwoot/platform/logger"
)

type MessageService struct {
	messagingCore *messaging.Service
	sessionCore   *session.Service
	resolver      session.SessionResolver

	messageRepo messaging.Repository
	sessionRepo session.Repository
	whatsappGW  session.WhatsAppGateway

	logger    *logger.Logger
	validator *validation.Validator

	sessionService *SessionService
}

func NewMessageService(
	messagingCore *messaging.Service,
	sessionCore *session.Service,
	resolver session.SessionResolver,
	messageRepo messaging.Repository,
	sessionRepo session.Repository,
	whatsappGW session.WhatsAppGateway,
	logger *logger.Logger,
	validator *validation.Validator,
	sessionService *SessionService,
) *MessageService {
	return &MessageService{
		messagingCore:  messagingCore,
		sessionCore:    sessionCore,
		resolver:       resolver,
		messageRepo:    messageRepo,
		sessionRepo:    sessionRepo,
		whatsappGW:     whatsappGW,
		logger:         logger,
		validator:      validator,
		sessionService: sessionService,
	}
}

func (s *MessageService) validateSession(ctx context.Context, sessionName string) (*session.Session, error) {
	sessionInfo, err := s.sessionCore.GetSessionByName(ctx, sessionName)
	if err != nil {
		return nil, fmt.Errorf("session %s not found: %w", sessionName, err)
	}

	if !sessionInfo.IsConnected {
		return nil, fmt.Errorf("session %s is not connected", sessionName)
	}

	return sessionInfo, nil
}

// resolveSessionID resolve idOrName e retorna ID, nome e sessão
func (s *MessageService) resolveSessionID(ctx context.Context, idOrName string) (uuid.UUID, string, *session.Session, error) {
	resolved, err := s.resolver.Resolve(ctx, idOrName)
	if err != nil {
		return uuid.Nil, "", nil, err
	}

	return resolved.ID, resolved.Name, resolved.Session, nil
}

type CreateMessageRequest struct {
	SessionID   string `json:"session_id" validate:"required,uuid"`
	ZpMessageID string `json:"zp_message_id" validate:"required"`
	ZpSender    string `json:"zp_sender" validate:"required"`
	ZpChat      string `json:"zp_chat" validate:"required"`
	ZpTimestamp string `json:"zp_timestamp" validate:"required"`
	ZpFromMe    bool   `json:"zp_from_me"`
	ZpType      string `json:"zp_type" validate:"required"`
	Content     string `json:"content,omitempty"`
}

type CreateMessageResponse struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	ZpMessageID string    `json:"zp_message_id"`
	SyncStatus  string    `json:"sync_status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListMessagesRequest struct {
	SessionID string `json:"session_id,omitempty" validate:"omitempty,uuid"`
	ChatJID   string `json:"chat_jid,omitempty"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}

type ListMessagesResponse struct {
	Messages []*contracts.MessageDTO `json:"messages"`
	Total    int64                   `json:"total"`
	Limit    int                     `json:"limit"`
	Offset   int                     `json:"offset"`
}

type UpdateSyncStatusRequest struct {
	MessageID        string `json:"message_id" validate:"required,uuid"`
	SyncStatus       string `json:"sync_status" validate:"required,oneof=pending synced failed"`
	CwMessageID      *int   `json:"cw_message_id,omitempty"`
	CwConversationID *int   `json:"cw_conversation_id,omitempty"`
}

func (s *MessageService) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*CreateMessageResponse, error) {

	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	_, err = s.sessionCore.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	zpTimestamp, err := time.Parse(time.RFC3339, req.ZpTimestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	coreReq := &messaging.CreateMessageRequest{
		SessionID:   sessionID,
		ZpMessageID: req.ZpMessageID,
		ZpSender:    req.ZpSender,
		ZpChat:      req.ZpChat,
		ZpTimestamp: zpTimestamp,
		ZpFromMe:    req.ZpFromMe,
		ZpType:      messaging.MessageType(req.ZpType),
		Content:     req.Content,
	}

	message, err := s.messagingCore.CreateMessage(ctx, coreReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	s.logger.InfoWithFields("Message created via application service", map[string]interface{}{
		"message_id":    message.ID.String(),
		"session_id":    message.SessionID.String(),
		"zp_message_id": message.ZpMessageID,
		"type":          message.ZpType,
	})

	return &CreateMessageResponse{
		ID:          message.ID.String(),
		SessionID:   message.SessionID.String(),
		ZpMessageID: message.ZpMessageID,
		SyncStatus:  message.SyncStatus,
		CreatedAt:   message.CreatedAt,
	}, nil
}

func (s *MessageService) GetMessage(ctx context.Context, messageID string) (*contracts.MessageDTO, error) {

	id, err := uuid.Parse(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	message, err := s.messagingCore.GetMessage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return s.messageToDTO(message), nil
}

func (s *MessageService) ListMessages(ctx context.Context, req *ListMessagesRequest) (*ListMessagesResponse, error) {

	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	coreReq := &messaging.ListMessagesRequest{
		SessionID: req.SessionID,
		ChatJID:   req.ChatJID,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	messages, total, err := s.messagingCore.ListMessages(ctx, coreReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	messageDTOs := make([]*contracts.MessageDTO, len(messages))
	for i, message := range messages {
		messageDTOs[i] = s.messageToDTO(message)
	}

	return &ListMessagesResponse{
		Messages: messageDTOs,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

func (s *MessageService) UpdateSyncStatus(ctx context.Context, req *UpdateSyncStatusRequest) error {

	if err := s.validator.ValidateStruct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	messageID, err := uuid.Parse(req.MessageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	status := messaging.SyncStatus(req.SyncStatus)
	err = s.messagingCore.UpdateSyncStatus(ctx, messageID, status, req.CwMessageID, req.CwConversationID)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	s.logger.InfoWithFields("Message sync status updated", map[string]interface{}{
		"message_id":         req.MessageID,
		"sync_status":        req.SyncStatus,
		"cw_message_id":      req.CwMessageID,
		"cw_conversation_id": req.CwConversationID,
	})

	return nil
}

func (s *MessageService) GetPendingSyncMessages(ctx context.Context, sessionID string, limit int) ([]*contracts.MessageDTO, error) {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	messages, err := s.messagingCore.GetPendingSyncMessages(ctx, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sync messages: %w", err)
	}

	messageDTOs := make([]*contracts.MessageDTO, len(messages))
	for i, message := range messages {
		messageDTOs[i] = s.messageToDTO(message)
	}

	return messageDTOs, nil
}

func (s *MessageService) GetMessageStats(ctx context.Context, sessionID *string) (*messaging.MessageStats, error) {
	if sessionID != nil {

		id, err := uuid.Parse(*sessionID)
		if err != nil {
			return nil, fmt.Errorf("invalid session ID: %w", err)
		}

		return s.messagingCore.GetStatsBySession(ctx, id)
	}

	return s.messagingCore.GetStats(ctx)
}

func (s *MessageService) SendTextMessage(ctx context.Context, sessionName, to, content string) (*contracts.SendMessageResponse, error) {

	if sessionName == "" || to == "" || content == "" {
		return nil, fmt.Errorf("sessionName, to, and content are required")
	}

	_, err := s.validateSession(ctx, sessionName)
	if err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Sending text message via WhatsApp", map[string]interface{}{
		"session_name": sessionName,
		"to":           to,
		"content_len":  len(content),
	})

	result, err := s.whatsappGW.SendTextMessage(ctx, sessionName, to, content)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message via WhatsApp Gateway: %w", err)
	}

	response := &contracts.SendMessageResponse{
		MessageID: result.MessageID,
		To:        result.To,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	s.logger.InfoWithFields("Text message sent successfully", map[string]interface{}{
		"session_name": sessionName,
		"message_id":   result.MessageID,
		"to":           result.To,
	})

	return response, nil
}

func (s *MessageService) SendMediaMessage(ctx context.Context, sessionName, to, mediaURL, caption, mediaType string) (*contracts.SendMessageResponse, error) {

	if sessionName == "" || to == "" || mediaURL == "" {
		return nil, fmt.Errorf("sessionName, to, and mediaURL are required")
	}

	_, err := s.validateSession(ctx, sessionName)
	if err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Sending media message via WhatsApp", map[string]interface{}{
		"session_name": sessionName,
		"to":           to,
		"media_url":    mediaURL,
		"media_type":   mediaType,
		"has_caption":  caption != "",
	})

	result, err := s.whatsappGW.SendMediaMessage(ctx, sessionName, to, mediaURL, caption, mediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to send media message via WhatsApp Gateway: %w", err)
	}

	response := &contracts.SendMessageResponse{
		MessageID: result.MessageID,
		To:        result.To,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	s.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"session_name": sessionName,
		"message_id":   result.MessageID,
		"to":           result.To,
		"media_type":   mediaType,
	})

	return response, nil
}

func (s *MessageService) SendImageMessage(ctx context.Context, sessionID, to, file, caption, filename string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "image")
}

func (s *MessageService) SendAudioMessage(ctx context.Context, sessionID, to, file, caption string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "audio")
}

func (s *MessageService) SendVideoMessage(ctx context.Context, sessionID, to, file, caption, filename string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "video")
}

func (s *MessageService) SendDocumentMessage(ctx context.Context, sessionID, to, file, caption, filename string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "document")
}

func (s *MessageService) SendStickerMessage(ctx context.Context, sessionID, to, file string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, "", "sticker")
}

func (s *MessageService) SendLocationMessage(ctx context.Context, sessionID, to string, latitude, longitude float64, address string) (*contracts.SendMessageResponse, error) {

	if sessionID == "" || to == "" {
		return nil, fmt.Errorf("sessionID and to are required")
	}

	_, sessionName, _, err := s.resolveSessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Sending location message via WhatsApp", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"latitude":   latitude,
		"longitude":  longitude,
		"address":    address,
	})

	result, err := s.whatsappGW.SendLocationMessage(ctx, sessionName, to, latitude, longitude, address)
	if err != nil {
		return nil, fmt.Errorf("failed to send location message via WhatsApp Gateway: %w", err)
	}

	response := &contracts.SendMessageResponse{
		MessageID: result.MessageID,
		To:        result.To,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	s.logger.InfoWithFields("Location message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"message_id": result.MessageID,
		"to":         result.To,
	})

	return response, nil
}

func (s *MessageService) SendContactMessage(ctx context.Context, sessionID, to, contactName, contactPhone string) (*contracts.SendMessageResponse, error) {

	if sessionID == "" || to == "" || contactName == "" || contactPhone == "" {
		return nil, fmt.Errorf("sessionID, to, contactName, and contactPhone are required")
	}

	_, _, _, err := s.resolveSessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Sending contact message via WhatsApp", map[string]interface{}{
		"session_id":    sessionID,
		"to":            to,
		"contact_name":  contactName,
		"contact_phone": contactPhone,
	})

	result, err := s.whatsappGW.SendContactMessage(ctx, sessionID, to, contactName, contactPhone)
	if err != nil {
		return nil, fmt.Errorf("failed to send contact message via WhatsApp Gateway: %w", err)
	}

	response := &contracts.SendMessageResponse{
		MessageID: result.MessageID,
		To:        result.To,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	s.logger.InfoWithFields("Contact message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"message_id": result.MessageID,
		"to":         result.To,
	})

	return response, nil
}

func (s *MessageService) messageToDTO(message *messaging.Message) *contracts.MessageDTO {
	return &contracts.MessageDTO{
		ID:               message.ID.String(),
		SessionID:        message.SessionID.String(),
		ZpMessageID:      message.ZpMessageID,
		ZpSender:         message.ZpSender,
		ZpChat:           message.ZpChat,
		ZpTimestamp:      message.ZpTimestamp,
		ZpFromMe:         message.ZpFromMe,
		ZpType:           message.ZpType,
		Content:          message.Content,
		CwMessageID:      message.CwMessageID,
		CwConversationID: message.CwConversationID,
		SyncStatus:       message.SyncStatus,
		SyncedAt:         message.SyncedAt,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
	}
}
