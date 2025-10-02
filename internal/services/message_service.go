package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/session"
	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/services/shared/validation"
	"zpwoot/platform/logger"
)

// MessageService implementa a camada de aplicação para mensagens
// Responsável por orquestrar entre core business logic e adapters externos
type MessageService struct {
	// Core business logic
	messagingCore *messaging.Service
	sessionCore   *session.Service

	// External dependencies (injected via interfaces)
	messageRepo messaging.Repository
	sessionRepo session.Repository
	whatsappGW  session.WhatsAppGateway

	// Platform dependencies
	logger    *logger.Logger
	validator *validation.Validator
}

// NewMessageService cria nova instância do serviço de aplicação
func NewMessageService(
	messagingCore *messaging.Service,
	sessionCore *session.Service,
	messageRepo messaging.Repository,
	sessionRepo session.Repository,
	whatsappGW session.WhatsAppGateway,
	logger *logger.Logger,
	validator *validation.Validator,
) *MessageService {
	return &MessageService{
		messagingCore: messagingCore,
		sessionCore:   sessionCore,
		messageRepo:   messageRepo,
		sessionRepo:   sessionRepo,
		whatsappGW:    whatsappGW,
		logger:        logger,
		validator:     validator,
	}
}

// CreateMessageRequest DTO para criação de mensagem
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

// CreateMessageResponse DTO para resposta de criação
type CreateMessageResponse struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	ZpMessageID string    `json:"zp_message_id"`
	SyncStatus  string    `json:"sync_status"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListMessagesRequest DTO para listagem de mensagens
type ListMessagesRequest struct {
	SessionID string `json:"session_id,omitempty" validate:"omitempty,uuid"`
	ChatJID   string `json:"chat_jid,omitempty"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}

// ListMessagesResponse DTO para resposta de listagem
type ListMessagesResponse struct {
	Messages []*contracts.MessageDTO `json:"messages"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// UpdateSyncStatusRequest DTO para atualização de status de sync
type UpdateSyncStatusRequest struct {
	MessageID        string `json:"message_id" validate:"required,uuid"`
	SyncStatus       string `json:"sync_status" validate:"required,oneof=pending synced failed"`
	CwMessageID      *int   `json:"cw_message_id,omitempty"`
	CwConversationID *int   `json:"cw_conversation_id,omitempty"`
}

// CreateMessage cria uma nova mensagem
func (s *MessageService) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*CreateMessageResponse, error) {
	// Validar entrada
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse session ID
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Verificar se sessão existe
	_, err = s.sessionCore.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Parse timestamp
	zpTimestamp, err := time.Parse(time.RFC3339, req.ZpTimestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Criar request para o core
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

	// Criar mensagem via core service
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

// GetMessage busca uma mensagem por ID
func (s *MessageService) GetMessage(ctx context.Context, messageID string) (*contracts.MessageDTO, error) {
	// Parse message ID
	id, err := uuid.Parse(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	// Buscar mensagem via core service
	message, err := s.messagingCore.GetMessage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return s.messageToDTO(message), nil
}

// ListMessages lista mensagens com filtros e paginação
func (s *MessageService) ListMessages(ctx context.Context, req *ListMessagesRequest) (*ListMessagesResponse, error) {
	// Validar entrada
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Aplicar defaults
	if req.Limit == 0 {
		req.Limit = 50
	}

	// Criar request para o core
	coreReq := &messaging.ListMessagesRequest{
		SessionID: req.SessionID,
		ChatJID:   req.ChatJID,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	// Listar mensagens via core service
	messages, total, err := s.messagingCore.ListMessages(ctx, coreReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	// Converter para DTOs
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

// UpdateSyncStatus atualiza o status de sincronização de uma mensagem
func (s *MessageService) UpdateSyncStatus(ctx context.Context, req *UpdateSyncStatusRequest) error {
	// Validar entrada
	if err := s.validator.ValidateStruct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Parse message ID
	messageID, err := uuid.Parse(req.MessageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	// Atualizar status via core service
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

// GetPendingSyncMessages busca mensagens pendentes de sincronização
func (s *MessageService) GetPendingSyncMessages(ctx context.Context, sessionID string, limit int) ([]*contracts.MessageDTO, error) {
	// Parse session ID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Buscar mensagens pendentes via core service
	messages, err := s.messagingCore.GetPendingSyncMessages(ctx, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sync messages: %w", err)
	}

	// Converter para DTOs
	messageDTOs := make([]*contracts.MessageDTO, len(messages))
	for i, message := range messages {
		messageDTOs[i] = s.messageToDTO(message)
	}

	return messageDTOs, nil
}

// GetMessageStats retorna estatísticas de mensagens
func (s *MessageService) GetMessageStats(ctx context.Context, sessionID *string) (*messaging.MessageStats, error) {
	if sessionID != nil {
		// Parse session ID
		id, err := uuid.Parse(*sessionID)
		if err != nil {
			return nil, fmt.Errorf("invalid session ID: %w", err)
		}

		return s.messagingCore.GetStatsBySession(ctx, id)
	}

	return s.messagingCore.GetStats(ctx)
}

// ===== WHATSAPP MESSAGE SENDING METHODS =====

// SendTextMessage envia uma mensagem de texto via WhatsApp
func (s *MessageService) SendTextMessage(ctx context.Context, sessionID, to, content string) (*contracts.SendMessageResponse, error) {
	// Validar parâmetros
	if sessionID == "" || to == "" || content == "" {
		return nil, fmt.Errorf("sessionID, to, and content are required")
	}

	// Parse session ID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Verificar se sessão existe e está conectada
	sessionInfo, err := s.sessionCore.GetSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !sessionInfo.IsConnected {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	s.logger.InfoWithFields("Sending text message via WhatsApp", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"content_len": len(content),
	})

	// Enviar mensagem via WhatsApp Gateway
	result, err := s.whatsappGW.SendTextMessage(ctx, sessionID, to, content)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message via WhatsApp Gateway: %w", err)
	}

	// Criar resposta
	response := &contracts.SendMessageResponse{
		MessageID: result.MessageID,
		To:        result.To,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	s.logger.InfoWithFields("Text message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"message_id": result.MessageID,
		"to":         result.To,
	})

	return response, nil
}

// SendMediaMessage envia uma mensagem de mídia via WhatsApp
func (s *MessageService) SendMediaMessage(ctx context.Context, sessionID, to, mediaURL, caption, mediaType string) (*contracts.SendMessageResponse, error) {
	// Validar parâmetros
	if sessionID == "" || to == "" || mediaURL == "" {
		return nil, fmt.Errorf("sessionID, to, and mediaURL are required")
	}

	// Parse session ID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Verificar se sessão existe e está conectada
	sessionInfo, err := s.sessionCore.GetSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !sessionInfo.IsConnected {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	s.logger.InfoWithFields("Sending media message via WhatsApp", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"media_url":  mediaURL,
		"media_type": mediaType,
		"has_caption": caption != "",
	})

	// Enviar mensagem via WhatsApp Gateway
	result, err := s.whatsappGW.SendMediaMessage(ctx, sessionID, to, mediaURL, caption, mediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to send media message via WhatsApp Gateway: %w", err)
	}

	// Criar resposta
	response := &contracts.SendMessageResponse{
		MessageID: result.MessageID,
		To:        result.To,
		Status:    result.Status,
		Timestamp: result.Timestamp,
	}

	s.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"message_id": result.MessageID,
		"to":         result.To,
		"media_type": mediaType,
	})

	return response, nil
}

// SendImageMessage envia uma mensagem de imagem via WhatsApp
func (s *MessageService) SendImageMessage(ctx context.Context, sessionID, to, file, caption, filename string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "image")
}

// SendAudioMessage envia uma mensagem de áudio via WhatsApp
func (s *MessageService) SendAudioMessage(ctx context.Context, sessionID, to, file, caption string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "audio")
}

// SendVideoMessage envia uma mensagem de vídeo via WhatsApp
func (s *MessageService) SendVideoMessage(ctx context.Context, sessionID, to, file, caption, filename string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "video")
}

// SendDocumentMessage envia uma mensagem de documento via WhatsApp
func (s *MessageService) SendDocumentMessage(ctx context.Context, sessionID, to, file, caption, filename string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, caption, "document")
}

// SendStickerMessage envia uma mensagem de sticker via WhatsApp
func (s *MessageService) SendStickerMessage(ctx context.Context, sessionID, to, file string) (*contracts.SendMessageResponse, error) {
	return s.SendMediaMessage(ctx, sessionID, to, file, "", "sticker")
}

// SendLocationMessage envia uma mensagem de localização via WhatsApp
func (s *MessageService) SendLocationMessage(ctx context.Context, sessionID, to string, latitude, longitude float64, address string) (*contracts.SendMessageResponse, error) {
	// Validar parâmetros
	if sessionID == "" || to == "" {
		return nil, fmt.Errorf("sessionID and to are required")
	}

	// Parse session ID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Verificar se sessão existe e está conectada
	sessionInfo, err := s.sessionCore.GetSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !sessionInfo.IsConnected {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	s.logger.InfoWithFields("Sending location message via WhatsApp", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"latitude":   latitude,
		"longitude":  longitude,
		"address":    address,
	})

	// Enviar mensagem via WhatsApp Gateway
	result, err := s.whatsappGW.SendLocationMessage(ctx, sessionID, to, latitude, longitude, address)
	if err != nil {
		return nil, fmt.Errorf("failed to send location message via WhatsApp Gateway: %w", err)
	}

	// Criar resposta
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

// SendContactMessage envia uma mensagem de contato via WhatsApp
func (s *MessageService) SendContactMessage(ctx context.Context, sessionID, to, contactName, contactPhone string) (*contracts.SendMessageResponse, error) {
	// Validar parâmetros
	if sessionID == "" || to == "" || contactName == "" || contactPhone == "" {
		return nil, fmt.Errorf("sessionID, to, contactName, and contactPhone are required")
	}

	// Parse session ID
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	// Verificar se sessão existe e está conectada
	sessionInfo, err := s.sessionCore.GetSession(ctx, sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if !sessionInfo.IsConnected {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	s.logger.InfoWithFields("Sending contact message via WhatsApp", map[string]interface{}{
		"session_id":    sessionID,
		"to":            to,
		"contact_name":  contactName,
		"contact_phone": contactPhone,
	})

	// Enviar mensagem via WhatsApp Gateway
	result, err := s.whatsappGW.SendContactMessage(ctx, sessionID, to, contactName, contactPhone)
	if err != nil {
		return nil, fmt.Errorf("failed to send contact message via WhatsApp Gateway: %w", err)
	}

	// Criar resposta
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

// messageToDTO converte uma mensagem do domínio para DTO
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