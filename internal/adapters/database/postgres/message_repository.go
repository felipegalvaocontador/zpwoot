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