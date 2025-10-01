package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"zpwoot/platform/logger"
)

// Service implementa a lógica de negócio pura para mensagens
// Esta é a camada Core da Clean Architecture - sem dependências externas
type Service struct {
	repository Repository
	logger     *logger.Logger
}

// NewService cria uma nova instância do serviço de mensagens
func NewService(repo Repository, logger *logger.Logger) *Service {
	return &Service{
		repository: repo,
		logger:     logger,
	}
}

// CreateMessage cria uma nova mensagem no sistema
func (s *Service) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*Message, error) {
	// Validação de entrada
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid create request: %w", err)
	}

	// Verificar se mensagem já existe
	exists, err := s.repository.ExistsByZpMessageID(ctx, req.SessionID, req.ZpMessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to check message existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("message with zpMessageID %s already exists", req.ZpMessageID)
	}

	// Criar nova mensagem
	now := time.Now()
	message := &Message{
		ID:          uuid.New(),
		SessionID:   req.SessionID,
		ZpMessageID: req.ZpMessageID,
		ZpSender:    req.ZpSender,
		ZpChat:      req.ZpChat,
		ZpTimestamp: req.ZpTimestamp,
		ZpFromMe:    req.ZpFromMe,
		ZpType:      string(req.ZpType),
		Content:     req.Content,
		SyncStatus:  string(SyncStatusPending),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Persistir mensagem
	if err := s.repository.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	s.logger.InfoWithFields("Message created successfully", map[string]interface{}{
		"message_id":    message.ID.String(),
		"session_id":    message.SessionID.String(),
		"zp_message_id": message.ZpMessageID,
		"type":          message.ZpType,
		"from_me":       message.ZpFromMe,
	})

	return message, nil
}

// GetMessage busca uma mensagem por ID
func (s *Service) GetMessage(ctx context.Context, id uuid.UUID) (*Message, error) {
	message, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return message, nil
}

// GetMessageByZpID busca uma mensagem pelo ID do WhatsApp
func (s *Service) GetMessageByZpID(ctx context.Context, sessionID uuid.UUID, zpMessageID string) (*Message, error) {
	message, err := s.repository.GetByZpMessageID(ctx, sessionID, zpMessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message by zp id: %w", err)
	}

	return message, nil
}

// UpdateSyncStatus atualiza o status de sincronização de uma mensagem
func (s *Service) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status SyncStatus, cwMessageID, cwConversationID *int) error {
	// Validar status
	if !IsValidSyncStatus(string(status)) {
		return fmt.Errorf("invalid sync status: %s", status)
	}

	// Atualizar status
	if err := s.repository.UpdateSyncStatus(ctx, id, status, cwMessageID, cwConversationID); err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	s.logger.InfoWithFields("Message sync status updated", map[string]interface{}{
		"message_id":         id.String(),
		"sync_status":        string(status),
		"cw_message_id":      cwMessageID,
		"cw_conversation_id": cwConversationID,
	})

	return nil
}

// MarkAsSynced marca uma mensagem como sincronizada com sucesso
func (s *Service) MarkAsSynced(ctx context.Context, id uuid.UUID, cwMessageID, cwConversationID int) error {
	if err := s.repository.MarkAsSynced(ctx, id, cwMessageID, cwConversationID); err != nil {
		return fmt.Errorf("failed to mark message as synced: %w", err)
	}

	s.logger.InfoWithFields("Message marked as synced", map[string]interface{}{
		"message_id":         id.String(),
		"cw_message_id":      cwMessageID,
		"cw_conversation_id": cwConversationID,
	})

	return nil
}

// MarkAsFailed marca uma mensagem como falha na sincronização
func (s *Service) MarkAsFailed(ctx context.Context, id uuid.UUID, errorReason string) error {
	if err := s.repository.MarkAsFailed(ctx, id, errorReason); err != nil {
		return fmt.Errorf("failed to mark message as failed: %w", err)
	}

	s.logger.ErrorWithFields("Message marked as failed", map[string]interface{}{
		"message_id":   id.String(),
		"error_reason": errorReason,
	})

	return nil
}

// ListMessages lista mensagens com paginação
func (s *Service) ListMessages(ctx context.Context, req *ListMessagesRequest) ([]*Message, int64, error) {
	// Validar request
	if err := s.validateListRequest(req); err != nil {
		return nil, 0, fmt.Errorf("invalid list request: %w", err)
	}

	var messages []*Message
	var err error

	// Buscar mensagens baseado nos filtros
	if req.SessionID != "" {
		sessionID, err := uuid.Parse(req.SessionID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid session ID: %w", err)
		}

		if req.ChatJID != "" {
			messages, err = s.repository.ListByChat(ctx, sessionID, req.ChatJID, req.Limit, req.Offset)
		} else {
			messages, err = s.repository.ListBySession(ctx, sessionID, req.Limit, req.Offset)
		}
	} else {
		messages, err = s.repository.List(ctx, req.Limit, req.Offset)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list messages: %w", err)
	}

	// Contar total
	var total int64
	if req.SessionID != "" {
		sessionID, _ := uuid.Parse(req.SessionID)
		if req.ChatJID != "" {
			total, err = s.repository.CountByChat(ctx, sessionID, req.ChatJID)
		} else {
			total, err = s.repository.CountBySession(ctx, sessionID)
		}
	} else {
		total, err = s.repository.Count(ctx)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return messages, total, nil
}

// GetPendingSyncMessages busca mensagens pendentes de sincronização
func (s *Service) GetPendingSyncMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*Message, error) {
	messages, err := s.repository.GetPendingSyncMessages(ctx, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sync messages: %w", err)
	}

	return messages, nil
}

// GetStats retorna estatísticas de mensagens
func (s *Service) GetStats(ctx context.Context) (*MessageStats, error) {
	stats, err := s.repository.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get message stats: %w", err)
	}

	return stats, nil
}

// GetStatsBySession retorna estatísticas de mensagens por sessão
func (s *Service) GetStatsBySession(ctx context.Context, sessionID uuid.UUID) (*MessageStats, error) {
	stats, err := s.repository.GetStatsBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message stats by session: %w", err)
	}

	return stats, nil
}

// validateCreateRequest valida os dados de criação de mensagem
func (s *Service) validateCreateRequest(req *CreateMessageRequest) error {
	if req.SessionID == uuid.Nil {
		return fmt.Errorf("session ID is required")
	}
	if req.ZpMessageID == "" {
		return fmt.Errorf("zp message ID is required")
	}
	if req.ZpSender == "" {
		return fmt.Errorf("zp sender is required")
	}
	if req.ZpChat == "" {
		return fmt.Errorf("zp chat is required")
	}
	if req.ZpTimestamp.IsZero() {
		return fmt.Errorf("zp timestamp is required")
	}
	if !IsValidMessageType(string(req.ZpType)) {
		return fmt.Errorf("invalid message type: %s", req.ZpType)
	}

	return nil
}

// validateListRequest valida os dados de listagem de mensagens
func (s *Service) validateListRequest(req *ListMessagesRequest) error {
	if req.Limit <= 0 {
		req.Limit = 50 // default
	}
	if req.Limit > 100 {
		req.Limit = 100 // max
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	return nil
}
