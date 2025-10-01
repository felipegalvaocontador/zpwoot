package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/session"
	"zpwoot/internal/services/shared/dto"
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
	logger *logger.Logger,
	validator *validation.Validator,
) *MessageService {
	return &MessageService{
		messagingCore: messagingCore,
		sessionCore:   sessionCore,
		messageRepo:   messageRepo,
		sessionRepo:   sessionRepo,
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
	Messages []*dto.MessageDTO `json:"messages"`
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
func (s *MessageService) GetMessage(ctx context.Context, messageID string) (*dto.MessageDTO, error) {
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
	messageDTOs := make([]*dto.MessageDTO, len(messages))
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
func (s *MessageService) GetPendingSyncMessages(ctx context.Context, sessionID string, limit int) ([]*dto.MessageDTO, error) {
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
	messageDTOs := make([]*dto.MessageDTO, len(messages))
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

// messageToDTO converte uma mensagem do domínio para DTO
func (s *MessageService) messageToDTO(message *messaging.Message) *dto.MessageDTO {
	return &dto.MessageDTO{
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