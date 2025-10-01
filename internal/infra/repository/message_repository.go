package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type MessageRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func NewMessageRepository(db *sqlx.DB, logger *logger.Logger) ports.ChatwootMessageRepository {
	return &MessageRepository{
		db:     db,
		logger: logger,
	}
}

type zpMessageModel struct {
	CreatedAt        time.Time     `db:"createdAt"`
	ZpTimestamp      time.Time     `db:"zpTimestamp"`
	UpdatedAt        time.Time     `db:"updatedAt"`
	SyncedAt         sql.NullTime  `db:"syncedAt"`
	SessionID        string        `db:"sessionId"`
	ZpMessageID      string        `db:"zpMessageId"`
	ZpSender         string        `db:"zpSender"`
	ZpChat           string        `db:"zpChat"`
	ZpType           string        `db:"zpType"`
	Content          string        `db:"content"`
	ID               string        `db:"id"`
	SyncStatus       string        `db:"syncStatus"`
	CwMessageID      sql.NullInt64 `db:"cwMessageId"`
	CwConversationID sql.NullInt64 `db:"cwConversationId"`
	ZpFromMe         bool          `db:"zpFromMe"`
}

func (r *MessageRepository) CreateMessage(ctx context.Context, message *ports.ZpMessage) error {
	r.logger.InfoWithFields("Creating zpMessage mapping", map[string]interface{}{
		"session_id":    message.SessionID,
		"zp_message_id": message.ZpMessageID,
		"sync_status":   message.SyncStatus,
	})

	model := r.messageToModel(message)

	query := `
		INSERT INTO "zpMessage" (
			id, "sessionId", "zpMessageId", "zpSender", "zpChat", "zpTimestamp",
			"zpFromMe", "zpType", "content", "cwMessageId", "cwConversationId",
			"syncStatus", "createdAt", "updatedAt", "syncedAt"
		) VALUES (
			:id, :sessionId, :zpMessageId, :zpSender, :zpChat, :zpTimestamp,
			:zpFromMe, :zpType, :content, :cwMessageId, :cwConversationId,
			:syncStatus, :createdAt, :updatedAt, :syncedAt
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to create zpMessage", map[string]interface{}{
			"session_id":    message.SessionID,
			"zp_message_id": message.ZpMessageID,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to create zpMessage: %w", err)
	}

	return nil
}

func (r *MessageRepository) GetMessageByZpID(ctx context.Context, sessionID, zpMessageID string) (*ports.ZpMessage, error) {
	r.logger.DebugWithFields("Getting zpMessage by ZP ID", map[string]interface{}{
		"session_id":    sessionID,
		"zp_message_id": zpMessageID,
	})

	var model zpMessageModel
	query := `SELECT * FROM "zpMessage" WHERE "sessionId" = $1 AND "zpMessageId" = $2`

	err := r.db.GetContext(ctx, &model, query, sessionID, zpMessageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("zpMessage not found")
		}
		r.logger.ErrorWithFields("Failed to get zpMessage by ZP ID", map[string]interface{}{
			"session_id":    sessionID,
			"zp_message_id": zpMessageID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to get zpMessage: %w", err)
	}

	message := r.messageFromModel(&model)
	return message, nil
}

func (r *MessageRepository) GetMessageByCwID(ctx context.Context, cwMessageID int) (*ports.ZpMessage, error) {
	r.logger.DebugWithFields("Getting zpMessage by CW ID", map[string]interface{}{
		"cw_message_id": cwMessageID,
	})

	var model zpMessageModel
	query := `SELECT * FROM "zpMessage" WHERE "cwMessageId" = $1`

	err := r.db.GetContext(ctx, &model, query, cwMessageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("zpMessage not found")
		}
		r.logger.ErrorWithFields("Failed to get zpMessage by CW ID", map[string]interface{}{
			"cw_message_id": cwMessageID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to get zpMessage: %w", err)
	}

	message := r.messageFromModel(&model)
	return message, nil
}

func (r *MessageRepository) UpdateSyncStatus(ctx context.Context, id, status string, cwMessageID, cwConversationID *int) error {
	r.logger.InfoWithFields("Updating zpMessage sync status", map[string]interface{}{
		"id":                 id,
		"status":             status,
		"cw_message_id":      cwMessageID,
		"cw_conversation_id": cwConversationID,
	})

	query := `
		UPDATE "zpMessage" 
		SET "syncStatus" = $2, "updatedAt" = NOW()
	`
	args := []interface{}{id, status}
	argIndex := 3

	if cwMessageID != nil {
		query += fmt.Sprintf(`, "cwMessageId" = $%d`, argIndex)
		args = append(args, *cwMessageID)
		argIndex++
	}

	if cwConversationID != nil {
		query += fmt.Sprintf(`, "cwConversationId" = $%d`, argIndex)
		args = append(args, *cwConversationID)
		argIndex++
	}

	if status == "synced" {
		query += fmt.Sprintf(`, "syncedAt" = $%d`, argIndex)
		args = append(args, time.Now())
	}

	query += ` WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update zpMessage sync status", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to update zpMessage sync status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zpMessage not found")
	}

	return nil
}

func (r *MessageRepository) GetMessagesBySession(ctx context.Context, sessionID string, limit, offset int) ([]*ports.ZpMessage, error) {
	r.logger.DebugWithFields("Getting zpMessages by session", map[string]interface{}{
		"session_id": sessionID,
		"limit":      limit,
		"offset":     offset,
	})

	var models []zpMessageModel
	query := `
		SELECT * FROM "zpMessage" 
		WHERE "sessionId" = $1 
		ORDER BY "zpTimestamp" DESC 
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &models, query, sessionID, limit, offset)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get zpMessages by session", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get zpMessages: %w", err)
	}

	messages := make([]*ports.ZpMessage, 0, len(models))
	for _, model := range models {
		message := r.messageFromModel(&model)
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *MessageRepository) GetMessagesByChat(ctx context.Context, sessionID, chatJID string, limit, offset int) ([]*ports.ZpMessage, error) {
	r.logger.DebugWithFields("Getting zpMessages by chat", map[string]interface{}{
		"session_id": sessionID,
		"chat_jid":   chatJID,
		"limit":      limit,
		"offset":     offset,
	})

	var models []zpMessageModel
	query := `
		SELECT * FROM "zpMessage" 
		WHERE "sessionId" = $1 AND "zpChat" = $2 
		ORDER BY "zpTimestamp" DESC 
		LIMIT $3 OFFSET $4
	`

	err := r.db.SelectContext(ctx, &models, query, sessionID, chatJID, limit, offset)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get zpMessages by chat", map[string]interface{}{
			"session_id": sessionID,
			"chat_jid":   chatJID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get zpMessages: %w", err)
	}

	messages := make([]*ports.ZpMessage, 0, len(models))
	for _, model := range models {
		message := r.messageFromModel(&model)
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *MessageRepository) GetPendingSyncMessages(ctx context.Context, sessionID string, limit int) ([]*ports.ZpMessage, error) {
	r.logger.DebugWithFields("Getting pending sync zpMessages", map[string]interface{}{
		"session_id": sessionID,
		"limit":      limit,
	})

	var models []zpMessageModel
	query := `
		SELECT * FROM "zpMessage" 
		WHERE "sessionId" = $1 AND "syncStatus" = 'pending' 
		ORDER BY "zpTimestamp" ASC 
		LIMIT $2
	`

	err := r.db.SelectContext(ctx, &models, query, sessionID, limit)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get pending sync zpMessages", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get pending sync zpMessages: %w", err)
	}

	messages := make([]*ports.ZpMessage, 0, len(models))
	for _, model := range models {
		message := r.messageFromModel(&model)
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, id string) error {
	r.logger.InfoWithFields("Deleting zpMessage", map[string]interface{}{
		"id": id,
	})

	query := `DELETE FROM "zpMessage" WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.ErrorWithFields("Failed to delete zpMessage", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to delete zpMessage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("zpMessage not found")
	}

	return nil
}

func (r *MessageRepository) messageToModel(message *ports.ZpMessage) *zpMessageModel {
	model := &zpMessageModel{
		ID:          message.ID,
		SessionID:   message.SessionID,
		ZpMessageID: message.ZpMessageID,
		ZpSender:    message.ZpSender,
		ZpChat:      message.ZpChat,
		ZpTimestamp: message.ZpTimestamp,
		ZpFromMe:    message.ZpFromMe,
		ZpType:      message.ZpType,
		Content:     message.Content,
		SyncStatus:  message.SyncStatus,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}

	if message.CwMessageID != nil {
		model.CwMessageID = sql.NullInt64{Int64: int64(*message.CwMessageID), Valid: true}
	}

	if message.CwConversationID != nil {
		model.CwConversationID = sql.NullInt64{Int64: int64(*message.CwConversationID), Valid: true}
	}

	if message.SyncedAt != nil {
		model.SyncedAt = sql.NullTime{Time: *message.SyncedAt, Valid: true}
	}

	return model
}

func (r *MessageRepository) messageFromModel(model *zpMessageModel) *ports.ZpMessage {
	message := &ports.ZpMessage{
		ID:          model.ID,
		SessionID:   model.SessionID,
		ZpMessageID: model.ZpMessageID,
		ZpSender:    model.ZpSender,
		ZpChat:      model.ZpChat,
		ZpTimestamp: model.ZpTimestamp,
		ZpFromMe:    model.ZpFromMe,
		ZpType:      model.ZpType,
		Content:     model.Content,
		SyncStatus:  model.SyncStatus,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	if model.CwMessageID.Valid {
		cwMessageID := int(model.CwMessageID.Int64)
		message.CwMessageID = &cwMessageID
	}

	if model.CwConversationID.Valid {
		cwConversationID := int(model.CwConversationID.Int64)
		message.CwConversationID = &cwConversationID
	}

	if model.SyncedAt.Valid {
		message.SyncedAt = &model.SyncedAt.Time
	}

	return message
}
