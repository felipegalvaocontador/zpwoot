package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"zpwoot/internal/core/integrations/chatwoot"
	"zpwoot/internal/core/shared/errors"
)

// ChatwootRepository implementa a interface chatwoot.Repository para PostgreSQL
type ChatwootRepository struct {
	db *sqlx.DB
}

// NewChatwootRepository cria uma nova instância do repositório do Chatwoot
func NewChatwootRepository(db *sqlx.DB) chatwoot.Repository {
	return &ChatwootRepository{
		db: db,
	}
}

// chatwootModel representa o modelo de dados para PostgreSQL
type chatwootModel struct {
	ID               string         `db:"id"`
	SessionID        string         `db:"sessionId"`
	URL              string         `db:"url"`
	Token            string         `db:"token"`
	AccountID        string         `db:"accountId"`
	InboxID          sql.NullString `db:"inboxId"`
	Enabled          bool           `db:"enabled"`
	InboxName        sql.NullString `db:"inboxName"`
	AutoCreate       sql.NullBool   `db:"autoCreate"`
	SignMsg          sql.NullBool   `db:"signMsg"`
	SignDelimiter    sql.NullString `db:"signDelimiter"`
	ReopenConv       sql.NullBool   `db:"reopenConv"`
	ConvPending      sql.NullBool   `db:"convPending"`
	ImportContacts   sql.NullBool   `db:"importContacts"`
	ImportMessages   sql.NullBool   `db:"importMessages"`
	ImportDays       sql.NullInt32  `db:"importDays"`
	MergeBrazil      sql.NullBool   `db:"mergeBrazil"`
	Organization     sql.NullString `db:"organization"`
	Logo             sql.NullString `db:"logo"`
	Number           sql.NullString `db:"number"`
	IgnoreJids       pq.StringArray `db:"ignoreJids"`
	CreatedAt        time.Time      `db:"createdAt"`
	UpdatedAt        time.Time      `db:"updatedAt"`
}

// Create cria uma nova configuração do Chatwoot
func (r *ChatwootRepository) Create(ctx context.Context, config *chatwoot.Config) error {
	model := r.toModel(config)

	query := `
		INSERT INTO "zpChatwoot" (
			id, "sessionId", url, token, "accountId", "inboxId", enabled,
			"inboxName", "autoCreate", "signMsg", "signDelimiter", "reopenConv",
			"convPending", "importContacts", "importMessages", "importDays",
			"mergeBrazil", organization, logo, number, "ignoreJids",
			"createdAt", "updatedAt"
		) VALUES (
			:id, :sessionId, :url, :token, :accountId, :inboxId, :enabled,
			:inboxName, :autoCreate, :signMsg, :signDelimiter, :reopenConv,
			:convPending, :importContacts, :importMessages, :importDays,
			:mergeBrazil, :organization, :logo, :number, :ignoreJids,
			:createdAt, :updatedAt
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "idx_zp_chatwoot_unique_session" {
					return errors.ErrChatwootConfigAlreadyExists
				}
			}
		}
		return fmt.Errorf("failed to create chatwoot config: %w", err)
	}

	return nil
}

// GetByID busca uma configuração pelo ID
func (r *ChatwootRepository) GetByID(ctx context.Context, id uuid.UUID) (*chatwoot.Config, error) {
	var model chatwootModel
	query := `SELECT * FROM "zpChatwoot" WHERE id = $1`

	err := r.db.GetContext(ctx, &model, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrChatwootConfigNotFound
		}
		return nil, fmt.Errorf("failed to get chatwoot config by ID: %w", err)
	}

	return r.fromModel(&model), nil
}

// GetBySessionID busca uma configuração pela sessão
func (r *ChatwootRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*chatwoot.Config, error) {
	var model chatwootModel
	query := `SELECT * FROM "zpChatwoot" WHERE "sessionId" = $1`

	err := r.db.GetContext(ctx, &model, query, sessionID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrChatwootConfigNotFound
		}
		return nil, fmt.Errorf("failed to get chatwoot config by session ID: %w", err)
	}

	return r.fromModel(&model), nil
}

// Update atualiza uma configuração existente
func (r *ChatwootRepository) Update(ctx context.Context, config *chatwoot.Config) error {
	model := r.toModel(config)

	query := `
		UPDATE "zpChatwoot" SET
			url = :url,
			token = :token,
			"accountId" = :accountId,
			"inboxId" = :inboxId,
			enabled = :enabled,
			"inboxName" = :inboxName,
			"autoCreate" = :autoCreate,
			"signMsg" = :signMsg,
			"signDelimiter" = :signDelimiter,
			"reopenConv" = :reopenConv,
			"convPending" = :convPending,
			"importContacts" = :importContacts,
			"importMessages" = :importMessages,
			"importDays" = :importDays,
			"mergeBrazil" = :mergeBrazil,
			organization = :organization,
			logo = :logo,
			number = :number,
			"ignoreJids" = :ignoreJids,
			"updatedAt" = :updatedAt
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update chatwoot config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrChatwootConfigNotFound
	}

	return nil
}

// Delete remove uma configuração do Chatwoot
func (r *ChatwootRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM "zpChatwoot" WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete chatwoot config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrChatwootConfigNotFound
	}

	return nil
}

// List retorna uma lista paginada de configurações
func (r *ChatwootRepository) List(ctx context.Context, limit, offset int) ([]*chatwoot.Config, error) {
	var models []chatwootModel
	query := `
		SELECT * FROM "zpChatwoot"
		ORDER BY "createdAt" DESC
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &models, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chatwoot configs: %w", err)
	}

	configs := make([]*chatwoot.Config, len(models))
	for i, model := range models {
		configs[i] = r.fromModel(&model)
	}

	return configs, nil
}

// ListEnabled retorna configurações habilitadas
func (r *ChatwootRepository) ListEnabled(ctx context.Context) ([]*chatwoot.Config, error) {
	var models []chatwootModel
	query := `SELECT * FROM "zpChatwoot" WHERE enabled = true ORDER BY "createdAt" DESC`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled chatwoot configs: %w", err)
	}

	configs := make([]*chatwoot.Config, len(models))
	for i, model := range models {
		configs[i] = r.fromModel(&model)
	}

	return configs, nil
}

// UpdateStatus atualiza apenas o status de habilitação
func (r *ChatwootRepository) UpdateStatus(ctx context.Context, id uuid.UUID, enabled bool) error {
	query := `UPDATE "zpChatwoot" SET enabled = $2, "updatedAt" = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String(), enabled)
	if err != nil {
		return fmt.Errorf("failed to update chatwoot config status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrChatwootConfigNotFound
	}

	return nil
}

// Count retorna o número total de configurações
func (r *ChatwootRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM "zpChatwoot"`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count chatwoot configs: %w", err)
	}

	return count, nil
}

// toModel converte uma entidade Config para o modelo de banco de dados
func (r *ChatwootRepository) toModel(config *chatwoot.Config) *chatwootModel {
	model := &chatwootModel{
		ID:        config.ID.String(),
		SessionID: config.SessionID.String(),
		URL:       config.URL,
		Token:     config.Token,
		AccountID: config.AccountID,
		Enabled:   config.Enabled,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}

	// InboxID
	if config.InboxID != nil {
		model.InboxID = sql.NullString{String: *config.InboxID, Valid: true}
	}

	// InboxName
	if config.InboxName != nil {
		model.InboxName = sql.NullString{String: *config.InboxName, Valid: true}
	}

	// AutoCreate
	if config.AutoCreate != nil {
		model.AutoCreate = sql.NullBool{Bool: *config.AutoCreate, Valid: true}
	}

	// SignMsg
	if config.SignMsg != nil {
		model.SignMsg = sql.NullBool{Bool: *config.SignMsg, Valid: true}
	}

	// SignDelimiter
	if config.SignDelimiter != nil {
		model.SignDelimiter = sql.NullString{String: *config.SignDelimiter, Valid: true}
	}

	// ReopenConv
	if config.ReopenConv != nil {
		model.ReopenConv = sql.NullBool{Bool: *config.ReopenConv, Valid: true}
	}

	// ConvPending
	if config.ConvPending != nil {
		model.ConvPending = sql.NullBool{Bool: *config.ConvPending, Valid: true}
	}

	// ImportContacts
	if config.ImportContacts != nil {
		model.ImportContacts = sql.NullBool{Bool: *config.ImportContacts, Valid: true}
	}

	// ImportMessages
	if config.ImportMessages != nil {
		model.ImportMessages = sql.NullBool{Bool: *config.ImportMessages, Valid: true}
	}

	// ImportDays
	if config.ImportDays != nil {
		model.ImportDays = sql.NullInt32{Int32: *config.ImportDays, Valid: true}
	}

	// MergeBrazil
	if config.MergeBrazil != nil {
		model.MergeBrazil = sql.NullBool{Bool: *config.MergeBrazil, Valid: true}
	}

	// Organization
	if config.Organization != nil {
		model.Organization = sql.NullString{String: *config.Organization, Valid: true}
	}

	// Logo
	if config.Logo != nil {
		model.Logo = sql.NullString{String: *config.Logo, Valid: true}
	}

	// Number
	if config.Number != nil {
		model.Number = sql.NullString{String: *config.Number, Valid: true}
	}

	// IgnoreJids
	if config.IgnoreJids != nil {
		model.IgnoreJids = pq.StringArray(config.IgnoreJids)
	}

	return model
}

// fromModel converte um modelo de banco de dados para uma entidade Config
func (r *ChatwootRepository) fromModel(model *chatwootModel) *chatwoot.Config {
	id, _ := uuid.Parse(model.ID)
	sessionID, _ := uuid.Parse(model.SessionID)

	config := &chatwoot.Config{
		ID:        id,
		SessionID: sessionID,
		URL:       model.URL,
		Token:     model.Token,
		AccountID: model.AccountID,
		Enabled:   model.Enabled,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	// InboxID
	if model.InboxID.Valid {
		config.InboxID = &model.InboxID.String
	}

	// InboxName
	if model.InboxName.Valid {
		config.InboxName = &model.InboxName.String
	}

	// AutoCreate
	if model.AutoCreate.Valid {
		config.AutoCreate = &model.AutoCreate.Bool
	}

	// SignMsg
	if model.SignMsg.Valid {
		config.SignMsg = &model.SignMsg.Bool
	}

	// SignDelimiter
	if model.SignDelimiter.Valid {
		config.SignDelimiter = &model.SignDelimiter.String
	}

	// ReopenConv
	if model.ReopenConv.Valid {
		config.ReopenConv = &model.ReopenConv.Bool
	}

	// ConvPending
	if model.ConvPending.Valid {
		config.ConvPending = &model.ConvPending.Bool
	}

	// ImportContacts
	if model.ImportContacts.Valid {
		config.ImportContacts = &model.ImportContacts.Bool
	}

	// ImportMessages
	if model.ImportMessages.Valid {
		config.ImportMessages = &model.ImportMessages.Bool
	}

	// ImportDays
	if model.ImportDays.Valid {
		config.ImportDays = &model.ImportDays.Int32
	}

	// MergeBrazil
	if model.MergeBrazil.Valid {
		config.MergeBrazil = &model.MergeBrazil.Bool
	}

	// Organization
	if model.Organization.Valid {
		config.Organization = &model.Organization.String
	}

	// Logo
	if model.Logo.Valid {
		config.Logo = &model.Logo.String
	}

	// Number
	if model.Number.Valid {
		config.Number = &model.Number.String
	}

	// IgnoreJids
	if len(model.IgnoreJids) > 0 {
		config.IgnoreJids = []string(model.IgnoreJids)
	}

	return config
}