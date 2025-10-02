package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/shared/errors"
	"zpwoot/platform/logger"
)

type MessageRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func NewMessageRepository(db *sqlx.DB, logger *logger.Logger) messaging.Repository {
	return &MessageRepository{
		db:     db,
		logger: logger,
	}
}

type messageModel struct {
	ID               string         `db:"id"`
	SessionID        string         `db:"sessionId"`
	ZpMessageID      string         `db:"zpMessageId"`
	ZpSender         string         `db:"zpSender"`
	ZpChat           string         `db:"zpChat"`
	ZpTimestamp      time.Time      `db:"zpTimestamp"`
	ZpFromMe         bool           `db:"zpFromMe"`
	ZpType           string         `db:"zpType"`
	Content          sql.NullString `db:"content"`
	CwMessageID      sql.NullInt64  `db:"cwMessageId"`
	CwConversationID sql.NullInt64  `db:"cwConversationId"`
	SyncStatus       string         `db:"syncStatus"`
	SyncedAt         pq.NullTime    `db:"syncedAt"`
	CreatedAt        time.Time      `db:"createdAt"`
	UpdatedAt        time.Time      `db:"updatedAt"`
}

func (r *MessageRepository) Create(ctx context.Context, message *messaging.Message) error {
	r.logger.DebugWithFields("Creating message", map[string]interface{}{
		"message_id":    message.ID.String(),
		"session_id":    message.SessionID.String(),
		"zp_message_id": message.ZpMessageID,
	})

	model := r.messageToModel(message)

	query := `
		INSERT INTO "zpMessage" (
			id, "sessionId", "zpMessageId", "zpSender", "zpChat", "zpTimestamp",
			"zpFromMe", "zpType", content, "cwMessageId", "cwConversationId",
			"syncStatus", "syncedAt", "createdAt", "updatedAt"
		) VALUES (
			:id, :sessionId, :zpMessageId, :zpSender, :zpChat, :zpTimestamp,
			:zpFromMe, :zpType, :content, :cwMessageId, :cwConversationId,
			:syncStatus, :syncedAt, :createdAt, :updatedAt
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				if pqErr.Constraint == "idx_zp_message_unique_zp" {
					return errors.ErrAlreadyExists
				}
			case "23503":
				return fmt.Errorf("session not found")
			}
		}
		return fmt.Errorf("failed to create message: %w", err)
	}

	r.logger.InfoWithFields("Message created successfully", map[string]interface{}{
		"message_id":    message.ID.String(),
		"zp_message_id": message.ZpMessageID,
	})

	return nil
}

func (r *MessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*messaging.Message, error) {
	var model messageModel

	query := `SELECT * FROM "zpMessage" WHERE id = $1`
	err := r.db.GetContext(ctx, &model, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	return r.modelToMessage(&model)
}

func (r *MessageRepository) GetByZpMessageID(ctx context.Context, sessionID uuid.UUID, zpMessageID string) (*messaging.Message, error) {
	var model messageModel

	query := `SELECT * FROM "zpMessage" WHERE "sessionId" = $1 AND "zpMessageId" = $2`
	err := r.db.GetContext(ctx, &model, query, sessionID.String(), zpMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get message by zp message ID: %w", err)
	}

	return r.modelToMessage(&model)
}

func (r *MessageRepository) ExistsByZpMessageID(ctx context.Context, sessionID uuid.UUID, zpMessageID string) (bool, error) {
	var count int

	query := `SELECT COUNT(*) FROM "zpMessage" WHERE "sessionId" = $1 AND "zpMessageId" = $2`
	err := r.db.GetContext(ctx, &count, query, sessionID.String(), zpMessageID)
	if err != nil {
		return false, fmt.Errorf("failed to check message existence: %w", err)
	}

	return count > 0, nil
}

func (r *MessageRepository) Update(ctx context.Context, message *messaging.Message) error {
	message.UpdatedAt = time.Now()
	model := r.messageToModel(message)

	query := `
		UPDATE "zpMessage" SET
			"zpSender" = :zpSender,
			"zpChat" = :zpChat,
			"zpTimestamp" = :zpTimestamp,
			"zpFromMe" = :zpFromMe,
			"zpType" = :zpType,
			content = :content,
			"cwMessageId" = :cwMessageId,
			"cwConversationId" = :cwConversationId,
			"syncStatus" = :syncStatus,
			"syncedAt" = :syncedAt,
			"updatedAt" = :updatedAt
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *MessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM "zpMessage" WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *MessageRepository) List(ctx context.Context, limit, offset int) ([]*messaging.Message, error) {
	var models []messageModel

	query := `
		SELECT * FROM "zpMessage" 
		ORDER BY "zpTimestamp" DESC 
		LIMIT $1 OFFSET $2
	`
	err := r.db.SelectContext(ctx, &models, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		message, err := r.modelToMessage(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to message: %w", err)
		}
		messages[i] = message
	}

	return messages, nil
}

func (r *MessageRepository) ListBySession(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*messaging.Message, error) {
	var models []messageModel

	query := `
		SELECT * FROM "zpMessage" 
		WHERE "sessionId" = $1 
		ORDER BY "zpTimestamp" DESC 
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &models, query, sessionID.String(), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages by session: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		message, err := r.modelToMessage(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to message: %w", err)
		}
		messages[i] = message
	}

	return messages, nil
}

func (r *MessageRepository) ListByChat(ctx context.Context, sessionID uuid.UUID, chatJID string, limit, offset int) ([]*messaging.Message, error) {
	var models []messageModel

	query := `
		SELECT * FROM "zpMessage" 
		WHERE "sessionId" = $1 AND "zpChat" = $2 
		ORDER BY "zpTimestamp" DESC 
		LIMIT $3 OFFSET $4
	`
	err := r.db.SelectContext(ctx, &models, query, sessionID.String(), chatJID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages by chat: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		message, err := r.modelToMessage(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to message: %w", err)
		}
		messages[i] = message
	}

	return messages, nil
}

func (r *MessageRepository) GetByCwMessageID(ctx context.Context, cwMessageID int) (*messaging.Message, error) {
	var model messageModel

	query := `SELECT * FROM "zpMessage" WHERE "cwMessageId" = $1`
	err := r.db.GetContext(ctx, &model, query, cwMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get message by cw message ID: %w", err)
	}

	return r.modelToMessage(&model)
}

func (r *MessageRepository) GetByCwConversationID(ctx context.Context, cwConversationID int, limit, offset int) ([]*messaging.Message, error) {
	var models []messageModel

	query := `
		SELECT * FROM "zpMessage"
		WHERE "cwConversationId" = $1
		ORDER BY "zpTimestamp" DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &models, query, cwConversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by cw conversation ID: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		message, err := r.modelToMessage(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to message: %w", err)
		}
		messages[i] = message
	}

	return messages, nil
}

func (r *MessageRepository) ListBySyncStatus(ctx context.Context, status messaging.SyncStatus, limit, offset int) ([]*messaging.Message, error) {
	var models []messageModel

	query := `
		SELECT * FROM "zpMessage"
		WHERE "syncStatus" = $1
		ORDER BY "createdAt" DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &models, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages by sync status: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		message, err := r.modelToMessage(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to message: %w", err)
		}
		messages[i] = message
	}

	return messages, nil
}

func (r *MessageRepository) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status messaging.SyncStatus, cwMessageID, cwConversationID *int) error {
	now := time.Now()

	query := `
		UPDATE "zpMessage" SET
			"syncStatus" = $2,
			"cwMessageId" = $3,
			"cwConversationId" = $4,
			"syncedAt" = $5,
			"updatedAt" = $6
		WHERE id = $1
	`

	var syncedAt *time.Time
	if status == messaging.SyncStatusSynced {
		syncedAt = &now
	}

	result, err := r.db.ExecContext(ctx, query, id.String(), string(status), cwMessageID, cwConversationID, syncedAt, now)
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *MessageRepository) GetPendingSyncMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*messaging.Message, error) {
	var models []messageModel

	query := `
		SELECT * FROM "zpMessage"
		WHERE "sessionId" = $1 AND "syncStatus" = 'pending'
		ORDER BY "createdAt" ASC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &models, query, sessionID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sync messages: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		message, err := r.modelToMessage(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to message: %w", err)
		}
		messages[i] = message
	}

	return messages, nil
}

func (r *MessageRepository) GetFailedSyncMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*messaging.Message, error) {
	var models []messageModel

	query := `
		SELECT * FROM "zpMessage"
		WHERE "sessionId" = $1 AND "syncStatus" = 'failed'
		ORDER BY "updatedAt" DESC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &models, query, sessionID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed sync messages: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		message, err := r.modelToMessage(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to message: %w", err)
		}
		messages[i] = message
	}

	return messages, nil
}

func (r *MessageRepository) MarkAsSynced(ctx context.Context, id uuid.UUID, cwMessageID, cwConversationID int) error {
	now := time.Now()

	query := `
		UPDATE "zpMessage" SET
			"syncStatus" = 'synced',
			"cwMessageId" = $2,
			"cwConversationId" = $3,
			"syncedAt" = $4,
			"updatedAt" = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id.String(), cwMessageID, cwConversationID, now, now)
	if err != nil {
		return fmt.Errorf("failed to mark message as synced: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *MessageRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorReason string) error {
	now := time.Now()

	query := `
		UPDATE "zpMessage" SET
			"syncStatus" = 'failed',
			"updatedAt" = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id.String(), now)
	if err != nil {
		return fmt.Errorf("failed to mark message as failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *MessageRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM "zpMessage"`
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

func (r *MessageRepository) CountBySession(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM "zpMessage" WHERE "sessionId" = $1`
	err := r.db.GetContext(ctx, &count, query, sessionID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to count messages by session: %w", err)
	}

	return count, nil
}

func (r *MessageRepository) CountByChat(ctx context.Context, sessionID uuid.UUID, chatJID string) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM "zpMessage" WHERE "sessionId" = $1 AND "zpChat" = $2`
	err := r.db.GetContext(ctx, &count, query, sessionID.String(), chatJID)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages by chat: %w", err)
	}

	return count, nil
}

func (r *MessageRepository) CountBySyncStatus(ctx context.Context, status messaging.SyncStatus) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM "zpMessage" WHERE "syncStatus" = $1`
	err := r.db.GetContext(ctx, &count, query, string(status))
	if err != nil {
		return 0, fmt.Errorf("failed to count messages by sync status: %w", err)
	}

	return count, nil
}

func (r *MessageRepository) CountByType(ctx context.Context, messageType messaging.MessageType) (int64, error) {
	var count int64

	query := `SELECT COUNT(*) FROM "zpMessage" WHERE "zpType" = $1`
	err := r.db.GetContext(ctx, &count, query, string(messageType))
	if err != nil {
		return 0, fmt.Errorf("failed to count messages by type: %w", err)
	}

	return count, nil
}

func (r *MessageRepository) GetStats(ctx context.Context) (*messaging.MessageStats, error) {
	stats := &messaging.MessageStats{
		MessagesByType:   make(map[string]int64),
		MessagesByStatus: make(map[string]int64),
	}

	totalCount, err := r.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats.TotalMessages = totalCount

	typeQuery := `
		SELECT "zpType", COUNT(*) as count
		FROM "zpMessage"
		GROUP BY "zpType"
	`
	typeRows, err := r.db.QueryContext(ctx, typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by type: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var msgType string
		var count int64
		if err := typeRows.Scan(&msgType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}
		stats.MessagesByType[msgType] = count
	}

	statusQuery := `
		SELECT "syncStatus", COUNT(*) as count
		FROM "zpMessage"
		GROUP BY "syncStatus"
	`
	statusRows, err := r.db.QueryContext(ctx, statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by status: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int64
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		stats.MessagesByStatus[status] = count

		switch status {
		case "synced":
			stats.SyncedMessages = count
		case "pending":
			stats.PendingMessages = count
		case "failed":
			stats.FailedMessages = count
		}
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var todayCount int64
	todayQuery := `SELECT COUNT(*) FROM "zpMessage" WHERE "createdAt" >= $1`
	err = r.db.GetContext(ctx, &todayCount, todayQuery, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get today count: %w", err)
	}
	stats.MessagesToday = todayCount

	var weekCount int64
	weekQuery := `SELECT COUNT(*) FROM "zpMessage" WHERE "createdAt" >= $1`
	err = r.db.GetContext(ctx, &weekCount, weekQuery, weekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get week count: %w", err)
	}
	stats.MessagesThisWeek = weekCount

	var monthCount int64
	monthQuery := `SELECT COUNT(*) FROM "zpMessage" WHERE "createdAt" >= $1`
	err = r.db.GetContext(ctx, &monthCount, monthQuery, monthStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get month count: %w", err)
	}
	stats.MessagesThisMonth = monthCount

	return stats, nil
}

func (r *MessageRepository) GetStatsBySession(ctx context.Context, sessionID uuid.UUID) (*messaging.MessageStats, error) {
	stats := &messaging.MessageStats{
		MessagesByType:   make(map[string]int64),
		MessagesByStatus: make(map[string]int64),
	}

	sessionIDStr := sessionID.String()

	totalCount, err := r.CountBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count for session: %w", err)
	}
	stats.TotalMessages = totalCount

	typeQuery := `
		SELECT "zpType", COUNT(*) as count
		FROM "zpMessage"
		WHERE "sessionId" = $1
		GROUP BY "zpType"
	`
	typeRows, err := r.db.QueryContext(ctx, typeQuery, sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by type for session: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var msgType string
		var count int64
		if err := typeRows.Scan(&msgType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}
		stats.MessagesByType[msgType] = count
	}

	statusQuery := `
		SELECT "syncStatus", COUNT(*) as count
		FROM "zpMessage"
		WHERE "sessionId" = $1
		GROUP BY "syncStatus"
	`
	statusRows, err := r.db.QueryContext(ctx, statusQuery, sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by status for session: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int64
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		stats.MessagesByStatus[status] = count

		switch status {
		case "synced":
			stats.SyncedMessages = count
		case "pending":
			stats.PendingMessages = count
		case "failed":
			stats.FailedMessages = count
		}
	}

	return stats, nil
}

func (r *MessageRepository) GetStatsForPeriod(ctx context.Context, sessionID uuid.UUID, from, to int64) (*messaging.MessageStats, error) {
	stats := &messaging.MessageStats{
		MessagesByType:   make(map[string]int64),
		MessagesByStatus: make(map[string]int64),
	}

	fromTime := time.Unix(from, 0)
	toTime := time.Unix(to, 0)
	sessionIDStr := sessionID.String()

	var totalCount int64
	totalQuery := `
		SELECT COUNT(*) FROM "zpMessage"
		WHERE "sessionId" = $1 AND "createdAt" BETWEEN $2 AND $3
	`
	err := r.db.GetContext(ctx, &totalCount, totalQuery, sessionIDStr, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count for period: %w", err)
	}
	stats.TotalMessages = totalCount

	typeQuery := `
		SELECT "zpType", COUNT(*) as count
		FROM "zpMessage"
		WHERE "sessionId" = $1 AND "createdAt" BETWEEN $2 AND $3
		GROUP BY "zpType"
	`
	typeRows, err := r.db.QueryContext(ctx, typeQuery, sessionIDStr, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by type for period: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var msgType string
		var count int64
		if err := typeRows.Scan(&msgType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}
		stats.MessagesByType[msgType] = count
	}

	statusQuery := `
		SELECT "syncStatus", COUNT(*) as count
		FROM "zpMessage"
		WHERE "sessionId" = $1 AND "createdAt" BETWEEN $2 AND $3
		GROUP BY "syncStatus"
	`
	statusRows, err := r.db.QueryContext(ctx, statusQuery, sessionIDStr, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by status for period: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int64
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		stats.MessagesByStatus[status] = count

		switch status {
		case "synced":
			stats.SyncedMessages = count
		case "pending":
			stats.PendingMessages = count
		case "failed":
			stats.FailedMessages = count
		}
	}

	return stats, nil
}

func (r *MessageRepository) DeleteOldMessages(ctx context.Context, olderThanDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	query := `DELETE FROM "zpMessage" WHERE "createdAt" < $1`
	result, err := r.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old messages: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *MessageRepository) DeleteBySession(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	query := `DELETE FROM "zpMessage" WHERE "sessionId" = $1`
	result, err := r.db.ExecContext(ctx, query, sessionID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to delete messages by session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *MessageRepository) CleanupFailedMessages(ctx context.Context, olderThanHours int) (int64, error) {
	cutoffDate := time.Now().Add(-time.Duration(olderThanHours) * time.Hour)

	query := `DELETE FROM "zpMessage" WHERE "syncStatus" = 'failed' AND "updatedAt" < $1`
	result, err := r.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup failed messages: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *MessageRepository) messageToModel(message *messaging.Message) *messageModel {
	model := &messageModel{
		ID:          message.ID.String(),
		SessionID:   message.SessionID.String(),
		ZpMessageID: message.ZpMessageID,
		ZpSender:    message.ZpSender,
		ZpChat:      message.ZpChat,
		ZpTimestamp: message.ZpTimestamp,
		ZpFromMe:    message.ZpFromMe,
		ZpType:      message.ZpType,
		SyncStatus:  message.SyncStatus,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}

	if message.Content != "" {
		model.Content = sql.NullString{String: message.Content, Valid: true}
	}

	if message.CwMessageID != nil {
		model.CwMessageID = sql.NullInt64{Int64: int64(*message.CwMessageID), Valid: true}
	}

	if message.CwConversationID != nil {
		model.CwConversationID = sql.NullInt64{Int64: int64(*message.CwConversationID), Valid: true}
	}

	if message.SyncedAt != nil {
		model.SyncedAt = pq.NullTime{Time: *message.SyncedAt, Valid: true}
	}

	return model
}

func (r *MessageRepository) modelToMessage(model *messageModel) (*messaging.Message, error) {

	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message ID: %w", err)
	}

	sessionID, err := uuid.Parse(model.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session ID: %w", err)
	}

	message := &messaging.Message{
		ID:          id,
		SessionID:   sessionID,
		ZpMessageID: model.ZpMessageID,
		ZpSender:    model.ZpSender,
		ZpChat:      model.ZpChat,
		ZpTimestamp: model.ZpTimestamp,
		ZpFromMe:    model.ZpFromMe,
		ZpType:      model.ZpType,
		SyncStatus:  model.SyncStatus,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	if model.Content.Valid {
		message.Content = model.Content.String
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

	return message, nil
}
