package chatwoot

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"github.com/google/uuid"
)

type MessageMapper struct {
	logger     *logger.Logger
	repository ports.ChatwootMessageRepository
}

func NewMessageMapper(logger *logger.Logger, repository ports.ChatwootMessageRepository) *MessageMapper {
	return &MessageMapper{
		logger:     logger,
		repository: repository,
	}
}

func (mm *MessageMapper) CreateMapping(ctx context.Context, sessionID, zpMessageID, zpSender, zpChat, zpType, content string, zpTimestamp time.Time, zpFromMe bool) (*ports.ZpMessage, error) {
	mapping := &ports.ZpMessage{
		ID:          uuid.New().String(),
		SessionID:   sessionID,
		ZpMessageID: zpMessageID,
		ZpSender:    zpSender,
		ZpChat:      zpChat,
		ZpTimestamp: zpTimestamp,
		ZpFromMe:    zpFromMe,
		ZpType:      zpType,
		Content:     content,
		SyncStatus:  "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := mm.repository.CreateMessage(ctx, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create message mapping: %w", err)
	}

	return mapping, nil
}

func (mm *MessageMapper) UpdateMapping(ctx context.Context, sessionID, zpMessageID string, cwMessageID, cwConversationID int) error {
	mm.logger.InfoWithFields("Updating message mapping", map[string]interface{}{
		"session_id":         sessionID,
		"zp_message_id":      zpMessageID,
		"cw_message_id":      cwMessageID,
		"cw_conversation_id": cwConversationID,
	})

	mapping, err := mm.repository.GetMessageByZpID(ctx, sessionID, zpMessageID)
	if err != nil {
		return fmt.Errorf("failed to get existing mapping: %w", err)
	}

	err = mm.repository.UpdateSyncStatus(ctx, mapping.ID, "synced", &cwMessageID, &cwConversationID)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}

	return nil
}

func (mm *MessageMapper) GetMappingByZpID(ctx context.Context, sessionID, zpMessageID string) (*ports.ZpMessage, error) {
	mapping, err := mm.repository.GetMessageByZpID(ctx, sessionID, zpMessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping by ZP ID: %w", err)
	}

	return mapping, nil
}

func (mm *MessageMapper) GetMappingByCwID(ctx context.Context, cwMessageID int) (*ports.ZpMessage, error) {
	mm.logger.DebugWithFields("Getting mapping by CW ID", map[string]interface{}{
		"cw_message_id": cwMessageID,
	})

	mapping, err := mm.repository.GetMessageByCwID(ctx, cwMessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping by CW ID: %w", err)
	}

	return mapping, nil
}

func (mm *MessageMapper) MarkAsFailed(ctx context.Context, sessionID, zpMessageID string) error {
	mapping, err := mm.repository.GetMessageByZpID(ctx, sessionID, zpMessageID)
	if err != nil {
		return fmt.Errorf("failed to get existing mapping: %w", err)
	}

	err = mm.repository.UpdateSyncStatus(ctx, mapping.ID, "failed", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to mark mapping as failed: %w", err)
	}

	return nil
}

func (mm *MessageMapper) GetPendingMappings(ctx context.Context, sessionID string, limit int) ([]*ports.ZpMessage, error) {
	mappings, err := mm.repository.GetPendingSyncMessages(ctx, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending mappings: %w", err)
	}

	return mappings, nil
}

func (mm *MessageMapper) DeleteMapping(ctx context.Context, sessionID, zpMessageID string) error {
	mm.logger.InfoWithFields("Deleting mapping", map[string]interface{}{
		"session_id":    sessionID,
		"zp_message_id": zpMessageID,
	})

	mapping, err := mm.repository.GetMessageByZpID(ctx, sessionID, zpMessageID)
	if err != nil {
		return fmt.Errorf("failed to get existing mapping: %w", err)
	}

	err = mm.repository.DeleteMessage(ctx, mapping.ID)
	if err != nil {
		return fmt.Errorf("failed to delete mapping: %w", err)
	}

	return nil
}

func (mm *MessageMapper) GetMappingStats(ctx context.Context, sessionID string) (*MappingStats, error) {
	mappings, err := mm.repository.GetMessagesBySession(ctx, sessionID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get mappings for stats: %w", err)
	}

	stats := &MappingStats{
		SessionID: sessionID,
		Total:     len(mappings),
	}

	for _, mapping := range mappings {
		switch mapping.SyncStatus {
		case "pending":
			stats.Pending++
		case "synced":
			stats.Synced++
		case "failed":
			stats.Failed++
		}
	}

	return stats, nil
}

type MappingStats struct {
	SessionID string `json:"session_id"`
	Total     int    `json:"total"`
	Pending   int    `json:"pending"`
	Synced    int    `json:"synced"`
	Failed    int    `json:"failed"`
}

func (mm *MessageMapper) IsMessageMapped(ctx context.Context, sessionID, zpMessageID string) bool {
	mapping, err := mm.GetMappingByZpID(ctx, sessionID, zpMessageID)
	if err != nil {
		return false
	}

	return mapping.CwMessageID != nil && *mapping.CwMessageID > 0
}

func (mm *MessageMapper) GetChatwootMessageID(ctx context.Context, sessionID, zpMessageID string) (int, error) {
	mapping, err := mm.GetMappingByZpID(ctx, sessionID, zpMessageID)
	if err != nil {
		return 0, fmt.Errorf("mapping not found: %w", err)
	}

	if mapping.CwMessageID == nil {
		return 0, fmt.Errorf("chatwoot message ID not set")
	}

	return *mapping.CwMessageID, nil
}

func (mm *MessageMapper) GetWhatsAppMessageID(ctx context.Context, cwMessageID int) (string, error) {
	mapping, err := mm.GetMappingByCwID(ctx, cwMessageID)
	if err != nil {
		return "", fmt.Errorf("mapping not found: %w", err)
	}

	return mapping.ZpMessageID, nil
}

func (mm *MessageMapper) CleanupOldMappings(ctx context.Context, sessionID string, olderThanDays int) (int, error) {
	return 0, nil
}
