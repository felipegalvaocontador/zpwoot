package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type webhookRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func NewWebhookRepository(db *sqlx.DB, logger *logger.Logger) ports.WebhookRepository {
	return &webhookRepository{
		db:     db,
		logger: logger,
	}
}

type webhookModel struct {
	CreatedAt time.Time      `db:"createdAt"`
	UpdatedAt time.Time      `db:"updatedAt"`
	ID        string         `db:"id"`
	URL       string         `db:"url"`
	Events    string         `db:"events"`
	SessionID sql.NullString `db:"sessionId"`
	Secret    sql.NullString `db:"secret"`
	Enabled   bool           `db:"enabled"`
}

func (r *webhookRepository) Create(ctx context.Context, wh *webhook.WebhookConfig) error {
	r.logger.InfoWithFields("Creating webhook", map[string]interface{}{
		"webhook_id": wh.ID.String(),
		"url":        wh.URL,
		"session_id": wh.SessionID,
	})

	model := r.toModel(wh)

	query := `
		INSERT INTO "zpWebhooks" (id, "sessionId", url, secret, events, enabled, "createdAt", "updatedAt")
		VALUES (:id, :sessionId, :url, :secret, :events, :enabled, :createdAt, :updatedAt)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to create webhook", map[string]interface{}{
			"webhook_id": wh.ID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	r.logger.InfoWithFields("Webhook created successfully", map[string]interface{}{
		"webhook_id": wh.ID.String(),
	})

	return nil
}

func (r *webhookRepository) GetByID(ctx context.Context, id string) (*webhook.WebhookConfig, error) {
	r.logger.InfoWithFields("Getting webhook by ID", map[string]interface{}{
		"webhook_id": id,
	})

	var model webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE id = $1`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, webhook.ErrWebhookNotFound
		}
		r.logger.ErrorWithFields("Failed to get webhook by ID", map[string]interface{}{
			"webhook_id": id,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	wh, err := r.fromModel(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return wh, nil
}

func (r *webhookRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*webhook.WebhookConfig, error) {
	r.logger.DebugWithFields("Getting webhooks by session ID", map[string]interface{}{
		"session_id": sessionID,
	})

	var models []webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE "sessionId" = $1 ORDER BY "createdAt" DESC`

	err := r.db.SelectContext(ctx, &models, query, sessionID)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get webhooks by session ID", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get webhooks: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

func (r *webhookRepository) GetGlobalWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error) {
	r.logger.Debug("Getting global webhooks")

	var models []webhookModel
	query := `SELECT * FROM "zpWebhooks" WHERE "sessionId" IS NULL ORDER BY "createdAt" DESC`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get global webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get global webhooks: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

func (r *webhookRepository) List(ctx context.Context, req *webhook.ListWebhooksRequest) ([]*webhook.WebhookConfig, int, error) {
	r.logger.InfoWithFields("Listing webhooks", map[string]interface{}{
		"session_id": req.SessionID,
		"enabled":    req.Enabled,
		"limit":      req.Limit,
		"offset":     req.Offset,
	})

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if req.SessionID != nil {
		whereClause += fmt.Sprintf(" AND \"sessionId\" = $%d", argIndex)
		args = append(args, *req.SessionID)
		argIndex++
	}

	if req.Enabled != nil {
		whereClause += fmt.Sprintf(" AND enabled = $%d", argIndex)
		args = append(args, *req.Enabled)
		argIndex++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"zpWebhooks\" %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.ErrorWithFields("Failed to count webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to count webhooks: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT * FROM "zpWebhooks" %s
		ORDER BY "createdAt" DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, req.Offset)

	var models []webhookModel
	err = r.db.SelectContext(ctx, &models, query, args...)
	if err != nil {
		r.logger.ErrorWithFields("Failed to list webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to list webhooks: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, total, nil
}

func (r *webhookRepository) Update(ctx context.Context, wh *webhook.WebhookConfig) error {
	r.logger.InfoWithFields("Updating webhook", map[string]interface{}{
		"webhook_id": wh.ID.String(),
	})

	model := r.toModel(wh)
	model.UpdatedAt = time.Now()

	query := `
		UPDATE "zpWebhooks"
		SET "sessionId" = :sessionId, url = :url, secret = :secret,
		    events = :events, enabled = :enabled, "updatedAt" = :updatedAt
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update webhook", map[string]interface{}{
			"webhook_id": wh.ID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return webhook.ErrWebhookNotFound
	}

	return nil
}

func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	r.logger.InfoWithFields("Deleting webhook", map[string]interface{}{
		"webhook_id": id,
	})

	query := `DELETE FROM "zpWebhooks" WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.ErrorWithFields("Failed to delete webhook", map[string]interface{}{
			"webhook_id": id,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return webhook.ErrWebhookNotFound
	}

	return nil
}

func (r *webhookRepository) UpdateStatus(ctx context.Context, id string, enabled bool) error {
	r.logger.InfoWithFields("Updating webhook status", map[string]interface{}{
		"webhook_id": id,
		"enabled":    enabled,
	})

	query := `UPDATE "zpWebhooks" SET enabled = $1, "updatedAt" = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, enabled, time.Now(), id)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update webhook status", map[string]interface{}{
			"webhook_id": id,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to update webhook status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return webhook.ErrWebhookNotFound
	}

	return nil
}

func (r *webhookRepository) GetEnabledWebhooks(ctx context.Context) ([]*webhook.WebhookConfig, error) {
	r.logger.Info("Getting enabled webhooks")

	query := `SELECT * FROM "zpWebhooks" WHERE enabled = true ORDER BY "createdAt" DESC`

	var models []webhookModel
	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get enabled webhooks", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get enabled webhooks: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

func (r *webhookRepository) GetWebhooksByEvent(ctx context.Context, eventType string) ([]*webhook.WebhookConfig, error) {
	r.logger.InfoWithFields("Getting webhooks by event", map[string]interface{}{
		"event_type": eventType,
	})

	query := `SELECT * FROM "zpWebhooks" WHERE enabled = true AND $1 = ANY(events) ORDER BY "createdAt" DESC`

	var models []webhookModel
	err := r.db.SelectContext(ctx, &models, query, eventType)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get webhooks by event", map[string]interface{}{
			"event_type": eventType,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get webhooks by event: %w", err)
	}

	webhooks := make([]*webhook.WebhookConfig, len(models))
	for i, model := range models {
		wh, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert webhook model", map[string]interface{}{
				"webhook_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		webhooks[i] = wh
	}

	return webhooks, nil
}

func (r *webhookRepository) CountByStatus(ctx context.Context, enabled bool) (int, error) {
	query := `SELECT COUNT(*) FROM "zpWebhooks" WHERE enabled = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, enabled)
	if err != nil {
		return 0, fmt.Errorf("failed to count webhooks by status: %w", err)
	}

	return count, nil
}

func (r *webhookRepository) GetWebhookStats(ctx context.Context, webhookID string) (*ports.WebhookStats, error) {
	return &ports.WebhookStats{
		WebhookID: webhookID,
	}, nil
}

func (r *webhookRepository) UpdateWebhookStats(ctx context.Context, webhookID string, stats *ports.WebhookStats) error {
	r.logger.InfoWithFields("Updating webhook stats", map[string]interface{}{
		"webhook_id": webhookID,
		"stats":      stats,
	})
	return nil
}

func (r *webhookRepository) toModel(wh *webhook.WebhookConfig) *webhookModel {
	model := &webhookModel{
		ID:        wh.ID.String(),
		URL:       wh.URL,
		Enabled:   wh.Enabled,
		CreatedAt: wh.CreatedAt,
		UpdatedAt: wh.UpdatedAt,
	}

	if wh.SessionID != nil {
		model.SessionID = sql.NullString{String: *wh.SessionID, Valid: true}
	}

	if wh.Secret != "" {
		model.Secret = sql.NullString{String: wh.Secret, Valid: true}
	}

	if len(wh.Events) > 0 {
		eventsJSON, err := json.Marshal(wh.Events)
		if err == nil {
			model.Events = string(eventsJSON)
		}
	} else {
		model.Events = "[]"
	}

	return model
}

func (r *webhookRepository) fromModel(model *webhookModel) (*webhook.WebhookConfig, error) {
	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook ID: %w", err)
	}

	wh := &webhook.WebhookConfig{
		ID:        id,
		URL:       model.URL,
		Enabled:   model.Enabled,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	if model.SessionID.Valid {
		wh.SessionID = &model.SessionID.String
	}

	if model.Secret.Valid {
		wh.Secret = model.Secret.String
	}

	if model.Events != "" {
		var events []string
		if err := json.Unmarshal([]byte(model.Events), &events); err == nil {
			wh.Events = events
		} else {
			wh.Events = []string{}
		}
	} else {
		wh.Events = []string{}
	}

	return wh, nil
}
