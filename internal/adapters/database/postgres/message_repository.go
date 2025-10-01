package postgres

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
)

// MessageRepository implementa a interface messaging.Repository para PostgreSQL
type MessageRepository struct {
	db *sqlx.DB
}

// NewMessageRepository cria uma nova instância do repositório de mensagens
func NewMessageRepository(db *sqlx.DB) messaging.Repository {
	return &MessageRepository{
		db: db,
	}
}

// messageModel representa o modelo de dados para PostgreSQL
type messageModel struct {
	ID                string         `db:"id"`
	SessionID         string         `db:"sessionId"`
	ZpMessageID       string         `db:"zpMessageId"`
	ZpSender          string         `db:"zpSender"`
	ZpChat            string         `db:"zpChat"`
	ZpTimestamp       time.Time      `db:"zpTimestamp"`
	ZpFromMe          bool           `db:"zpFromMe"`
	ZpType            string         `db:"zpType"`
	Content           sql.NullString `db:"content"`
	CwMessageID       sql.NullInt32  `db:"cwMessageId"`
	CwConversationID  sql.NullInt32  `db:"cwConversationId"`
	SyncStatus        string         `db:"syncStatus"`
	CreatedAt         time.Time      `db:"createdAt"`
	UpdatedAt         time.Time      `db:"updatedAt"`
	SyncedAt          sql.NullTime   `db:"syncedAt"`
}

// Create cria uma nova mensagem no banco de dados
func (r *MessageRepository) Create(ctx context.Context, msg *messaging.Message) error {
	model := r.toModel(msg)

	query := `
		INSERT INTO "zpMessage" (
			id, "sessionId", "zpMessageId", "zpSender", "zpChat",
			"zpTimestamp", "zpFromMe", "zpType", content, "cwMessageId",
			"cwConversationId", "syncStatus", "createdAt", "updatedAt", "syncedAt"
		) VALUES (
			:id, :sessionId, :zpMessageId, :zpSender, :zpChat,
			:zpTimestamp, :zpFromMe, :zpType, :content, :cwMessageId,
			:cwConversationId, :syncStatus, :createdAt, :updatedAt, :syncedAt
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "idx_zp_message_unique_zp" {
					return errors.ErrMessageAlreadyExists
				}
			}
		}
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID busca uma mensagem pelo ID
func (r *MessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*messaging.Message, error) {
	var model messageModel
	query := `SELECT * FROM "zpMessage" WHERE id = $1`

	err := r.db.GetContext(ctx, &model, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrMessageNotFound
		}
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	return r.fromModel(&model), nil
}

// GetByWhatsAppID busca uma mensagem pelo ID do WhatsApp
func (r *MessageRepository) GetByWhatsAppID(ctx context.Context, sessionID uuid.UUID, whatsappID string) (*messaging.Message, error) {
	var model messageModel
	query := `SELECT * FROM "zpMessage" WHERE "sessionId" = $1 AND "zpMessageId" = $2`

	err := r.db.GetContext(ctx, &model, query, sessionID.String(), whatsappID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrMessageNotFound
		}
		return nil, fmt.Errorf("failed to get message by WhatsApp ID: %w", err)
	}

	return r.fromModel(&model), nil
}

// GetByChatwootID busca uma mensagem pelo ID do Chatwoot
func (r *MessageRepository) GetByChatwootID(ctx context.Context, chatwootID int32) (*messaging.Message, error) {
	var model messageModel
	query := `SELECT * FROM "zpMessage" WHERE "cwMessageId" = $1`

	err := r.db.GetContext(ctx, &model, query, chatwootID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrMessageNotFound
		}
		return nil, fmt.Errorf("failed to get message by Chatwoot ID: %w", err)
	}

	return r.fromModel(&model), nil
}

// Update atualiza uma mensagem existente
func (r *MessageRepository) Update(ctx context.Context, msg *messaging.Message) error {
	model := r.toModel(msg)

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
			"updatedAt" = :updatedAt,
			"syncedAt" = :syncedAt
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
		return errors.ErrMessageNotFound
	}

	return nil
}

// Delete remove uma mensagem do banco de dados
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
		return errors.ErrMessageNotFound
	}

	return nil
}

// ListBySession retorna mensagens de uma sessão específica
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
		messages[i] = r.fromModel(&model)
	}

	return messages, nil
}

// ListByChat retorna mensagens de um chat específico
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
		messages[i] = r.fromModel(&model)
	}

	return messages, nil
}

// ListBySyncStatus retorna mensagens com status de sincronização específico
func (r *MessageRepository) ListBySyncStatus(ctx context.Context, sessionID uuid.UUID, status messaging.SyncStatus, limit, offset int) ([]*messaging.Message, error) {
	var models []messageModel
	query := `
		SELECT * FROM "zpMessage"
		WHERE "sessionId" = $1 AND "syncStatus" = $2
		ORDER BY "createdAt" ASC
		LIMIT $3 OFFSET $4
	`

	err := r.db.SelectContext(ctx, &models, query, sessionID.String(), string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages by sync status: %w", err)
	}

	messages := make([]*messaging.Message, len(models))
	for i, model := range models {
		messages[i] = r.fromModel(&model)
	}

	return messages, nil
}

// UpdateSyncStatus atualiza o status de sincronização de uma mensagem
func (r *MessageRepository) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status messaging.SyncStatus) error {
	query := `
		UPDATE "zpMessage" SET
			"syncStatus" = $2,
			"syncedAt" = CASE WHEN $2 = 'synced' THEN NOW() ELSE "syncedAt" END,
			"updatedAt" = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id.String(), string(status))
	if err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrMessageNotFound
	}

	return nil
}

// UpdateChatwootIDs atualiza os IDs do Chatwoot para uma mensagem
func (r *MessageRepository) UpdateChatwootIDs(ctx context.Context, id uuid.UUID, messageID, conversationID int32) error {
	query := `
		UPDATE "zpMessage" SET
			"cwMessageId" = $2,
			"cwConversationId" = $3,
			"updatedAt" = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id.String(), messageID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to update Chatwoot IDs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrMessageNotFound
	}

	return nil
}

// Count retorna o número total de mensagens
func (r *MessageRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM "zpMessage"`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// CountBySession retorna o número de mensagens de uma sessão
func (r *MessageRepository) CountBySession(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM "zpMessage" WHERE "sessionId" = $1`

	err := r.db.GetContext(ctx, &count, query, sessionID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to count messages by session: %w", err)
	}

	return count, nil
}

// toModel converte uma entidade Message para o modelo de banco de dados
func (r *MessageRepository) toModel(msg *messaging.Message) *messageModel {
	model := &messageModel{
		ID:          msg.ID.String(),
		SessionID:   msg.SessionID.String(),
		ZpMessageID: msg.WhatsAppID,
		ZpSender:    msg.Sender,
		ZpChat:      msg.Chat,
		ZpTimestamp: msg.Timestamp,
		ZpFromMe:    msg.FromMe,
		ZpType:      string(msg.Type),
		SyncStatus:  string(msg.SyncStatus),
		CreatedAt:   msg.CreatedAt,
		UpdatedAt:   msg.UpdatedAt,
	}

	// Content
	if msg.Content != nil {
		model.Content = sql.NullString{String: *msg.Content, Valid: true}
	}

	// ChatwootMessageID
	if msg.ChatwootMessageID != nil {
		model.CwMessageID = sql.NullInt32{Int32: *msg.ChatwootMessageID, Valid: true}
	}

	// ChatwootConversationID
	if msg.ChatwootConversationID != nil {
		model.CwConversationID = sql.NullInt32{Int32: *msg.ChatwootConversationID, Valid: true}
	}

	// SyncedAt
	if msg.SyncedAt != nil {
		model.SyncedAt = sql.NullTime{Time: *msg.SyncedAt, Valid: true}
	}

	return model
}

// fromModel converte um modelo de banco de dados para uma entidade Message
func (r *MessageRepository) fromModel(model *messageModel) *messaging.Message {
	id, _ := uuid.Parse(model.ID)
	sessionID, _ := uuid.Parse(model.SessionID)

	msg := &messaging.Message{
		ID:          id,
		SessionID:   sessionID,
		WhatsAppID:  model.ZpMessageID,
		Sender:      model.ZpSender,
		Chat:        model.ZpChat,
		Timestamp:   model.ZpTimestamp,
		FromMe:      model.ZpFromMe,
		Type:        messaging.MessageType(model.ZpType),
		SyncStatus:  messaging.SyncStatus(model.SyncStatus),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	// Content
	if model.Content.Valid {
		msg.Content = &model.Content.String
	}

	// ChatwootMessageID
	if model.CwMessageID.Valid {
		msg.ChatwootMessageID = &model.CwMessageID.Int32
	}

	// ChatwootConversationID
	if model.CwConversationID.Valid {
		msg.ChatwootConversationID = &model.CwConversationID.Int32
	}

	// SyncedAt
	if model.SyncedAt.Valid {
		msg.SyncedAt = &model.SyncedAt.Time
	}

	return msg
}