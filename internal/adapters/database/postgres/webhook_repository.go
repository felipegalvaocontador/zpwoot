package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"zpwoot/internal/core/integrations/webhook"
	"zpwoot/internal/core/shared/errors"
)

// WebhookRepository implementa a interface webhook.Repository para PostgreSQL
type WebhookRepository struct {
	db *sqlx.DB
}

// NewWebhookRepository cria uma nova instância do repositório de webhooks
func NewWebhookRepository(db *sqlx.DB) webhook.Repository {
	return &WebhookRepository{
		db: db,
	}
}

// webhookModel representa o modelo de dados para PostgreSQL
type webhookModel struct {
	ID        string         `db:"id"`
	SessionID sql.NullString `db:"sessionId"`
	URL       string         `db:"url"`
	Secret    sql.NullString `db:"secret"`
	Events    string         `db:"events"` // JSON array
	Enabled   bool           `db:"enabled"`
	CreatedAt time.Time      `db:"createdAt"`
	UpdatedAt time.Time      `db:"updatedAt"`
}

// Create cria um novo webhook
func (r *WebhookRepository) Create(ctx context.Context, wh *webhook.Config) error {
	model, err := r.toModel(wh)
	if err != nil {
		return fmt.Errorf("failed to convert webhook to model: %w", err)
	}

	query := `
		INSERT INTO "zpWebhooks" (id, "sessionId", url, secret, events, enabled, "createdAt", "updatedAt")
		VALUES (:id, :sessionId, :url, :secret, :events, :enabled, :createdAt, :updatedAt)
	`

	_, err = r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	return nil
}

// GetByID busca um webhook pelo ID
func (r *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*webhook.Config, error) {
	var model webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE id = $1`

	err := r.db.GetContext(ctx, &model, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrWebhookNotFound
		}
		return nil, fmt.Errorf("failed to get webhook by ID: %w", err)
	}

	return r.fromModel(&model)
}

// GetBySessionID busca webhooks de uma sessão
func (r *WebhookRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*webhook.Config, error) {
	var models []webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE "sessionId" = $1 ORDER BY "createdAt" DESC`

	err := r.db.SelectContext(ctx, &models, query, sessionID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get webhooks by session ID: %w", err)
	}

	webhooks := make([]*webhook.Config, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to webhook: %w", err)
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

// Update atualiza um webhook existente
func (r *WebhookRepository) Update(ctx context.Context, wh *webhook.Config) error {
	model, err := r.toModel(wh)
	if err != nil {
		return fmt.Errorf("failed to convert webhook to model: %w", err)
	}

	query := `
		UPDATE "zpWebhooks" SET
			"sessionId" = :sessionId,
			url = :url,
			secret = :secret,
			events = :events,
			enabled = :enabled,
			"updatedAt" = :updatedAt
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrWebhookNotFound
	}

	return nil
}

// Delete remove um webhook
func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM "zpWebhooks" WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrWebhookNotFound
	}

	return nil
}

// List retorna uma lista paginada de webhooks
func (r *WebhookRepository) List(ctx context.Context, limit, offset int) ([]*webhook.Config, error) {
	var models []webhookModel
	query := `
		SELECT * FROM "zpWebhooks"
		ORDER BY "createdAt" DESC
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &models, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	webhooks := make([]*webhook.Config, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to webhook: %w", err)
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

// toModel converte uma entidade Config para o modelo de banco de dados
func (r *WebhookRepository) toModel(wh *webhook.Config) (*webhookModel, error) {
	model := &webhookModel{
		ID:        wh.ID.String(),
		URL:       wh.URL,
		Enabled:   wh.Enabled,
		CreatedAt: wh.CreatedAt,
		UpdatedAt: wh.UpdatedAt,
	}

	// SessionID
	if wh.SessionID != nil {
		model.SessionID = sql.NullString{String: wh.SessionID.String(), Valid: true}
	}

	// Secret
	if wh.Secret != nil {
		model.Secret = sql.NullString{String: *wh.Secret, Valid: true}
	}

	// Events
	eventsJSON, err := json.Marshal(wh.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}
	model.Events = string(eventsJSON)

	return model, nil
}

// fromModel converte um modelo de banco de dados para uma entidade Config
func (r *WebhookRepository) fromModel(model *webhookModel) (*webhook.Config, error) {
	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook ID: %w", err)
	}

	wh := &webhook.Config{
		ID:        id,
		URL:       model.URL,
		Enabled:   model.Enabled,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	// SessionID
	if model.SessionID.Valid {
		sessionID, err := uuid.Parse(model.SessionID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse session ID: %w", err)
		}
		wh.SessionID = &sessionID
	}

	// Secret
	if model.Secret.Valid {
		wh.Secret = &model.Secret.String
	}

	// Events
	var events []string
	if err := json.Unmarshal([]byte(model.Events), &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}
	wh.Events = events

	return wh, nil
}