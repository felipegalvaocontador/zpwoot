package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	pkgErrors "zpwoot/pkg/errors"
	"zpwoot/platform/logger"
)

type sessionRepository struct {
	db     *sqlx.DB
	logger *logger.Logger
}

func NewSessionRepository(db *sqlx.DB, logger *logger.Logger) ports.SessionRepository {
	return &sessionRepository{
		db:     db,
		logger: logger,
	}
}

type sessionModel struct {
	CreatedAt       time.Time      `db:"createdAt"`
	UpdatedAt       time.Time      `db:"updatedAt"`
	QRCodeExpiresAt sql.NullTime   `db:"qrCodeExpiresAt"`
	ConnectedAt     sql.NullTime   `db:"connectedAt"`
	LastSeen        sql.NullTime   `db:"lastSeen"`
	ID              string         `db:"id"`
	Name            string         `db:"name"`
	DeviceJid       sql.NullString `db:"deviceJid"`
	ConnectionError sql.NullString `db:"connectionError"`
	QRCode          sql.NullString `db:"qrCode"`
	ProxyConfig     sql.NullString `db:"proxyConfig"`
	IsConnected     bool           `db:"isConnected"`
}

func (r *sessionRepository) Create(ctx context.Context, sess *session.Session) error {
	r.logger.InfoWithFields("Creating session", map[string]interface{}{
		"session_id": sess.ID.String(),
		"name":       sess.Name,
	})

	model := r.toModel(sess)

	query := `
		INSERT INTO "zpSessions" (id, name, "deviceJid", "isConnected", "connectionError", "qrCode", "qrCodeExpiresAt", "proxyConfig", "createdAt", "updatedAt", "connectedAt", "lastSeen")
		VALUES (:id, :name, :deviceJid, :isConnected, :connectionError, :qrCode, :qrCodeExpiresAt, :proxyConfig, :createdAt, :updatedAt, :connectedAt, :lastSeen)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to create session", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(err.Error(), "zpSessions_name_key") {
			return pkgErrors.NewWithDetails(409, "Session already exists", fmt.Sprintf("A session with the name '%s' already exists", sess.Name))
		}

		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.InfoWithFields("Session created successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, id string) (*session.Session, error) {
	var model sessionModel
	query := `SELECT * FROM "zpSessions" WHERE id = $1`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, session.ErrSessionNotFound
		}
		r.logger.ErrorWithFields("Failed to get session by ID", map[string]interface{}{
			"session_id": id,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	sess, err := r.fromModel(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return sess, nil
}

func (r *sessionRepository) GetByIDOrName(ctx context.Context, idOrName string) (*session.Session, error) {
	r.logger.InfoWithFields("Getting session by ID or name", map[string]interface{}{
		"identifier": idOrName,
	})

	// Primeiro tenta por ID (UUID)
	if isValidUUID(idOrName) {
		return r.GetByID(ctx, idOrName)
	}

	// Se não é UUID, tenta por nome
	return r.GetByName(ctx, idOrName)
}

func (r *sessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	r.logger.InfoWithFields("Getting session by name", map[string]interface{}{
		"session_name": name,
	})

	var model sessionModel
	query := `SELECT * FROM "zpSessions" WHERE name = $1`

	err := r.db.GetContext(ctx, &model, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, session.ErrSessionNotFound
		}
		r.logger.ErrorWithFields("Failed to get session by name", map[string]interface{}{
			"session_name": name,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	sess, err := r.fromModel(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return sess, nil
}

func (r *sessionRepository) GetByDeviceJid(ctx context.Context, deviceJid string) (*session.Session, error) {
	r.logger.InfoWithFields("Getting session by device JID", map[string]interface{}{
		"device_jid": deviceJid,
	})

	var model sessionModel
	query := `SELECT * FROM "zpSessions" WHERE "deviceJid" = $1`

	err := r.db.GetContext(ctx, &model, query, deviceJid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, session.ErrSessionNotFound
		}
		r.logger.ErrorWithFields("Failed to get session by device JID", map[string]interface{}{
			"device_jid": deviceJid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	sess, err := r.fromModel(&model)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}

	return sess, nil
}

func (r *sessionRepository) List(ctx context.Context, req *session.ListSessionsRequest) ([]*session.Session, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if req.IsConnected != nil {
		whereClause += fmt.Sprintf(" AND \"isConnected\" = $%d", argIndex)
		args = append(args, *req.IsConnected)
		argIndex++
	}

	if req.DeviceJid != nil {
		whereClause += fmt.Sprintf(" AND \"deviceJid\" = $%d", argIndex)
		args = append(args, *req.DeviceJid)
		argIndex++
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "zpSessions" %s`, whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.ErrorWithFields("Failed to count sessions", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT * FROM "zpSessions" %s
		ORDER BY "createdAt" DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, req.Offset)

	var models []sessionModel
	err = r.db.SelectContext(ctx, &models, query, args...)
	if err != nil {
		r.logger.ErrorWithFields("Failed to list sessions", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := make([]*session.Session, len(models))
	for i, model := range models {
		sess, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert session model", map[string]interface{}{
				"session_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		sessions[i] = sess
	}

	return sessions, total, nil
}

func (r *sessionRepository) Update(ctx context.Context, sess *session.Session) error {
	model := r.toModel(sess)
	model.UpdatedAt = time.Now()

	query := `
		UPDATE "zpSessions"
		SET name = :name, "deviceJid" = :deviceJid, "isConnected" = :isConnected,
		    "connectionError" = :connectionError, "qrCode" = :qrCode, "qrCodeExpiresAt" = :qrCodeExpiresAt,
		    "proxyConfig" = :proxyConfig, "connectedAt" = :connectedAt,
		    "lastSeen" = :lastSeen, "updatedAt" = :updatedAt
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update session", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return session.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	r.logger.InfoWithFields("Deleting session", map[string]interface{}{
		"session_id": id,
	})

	query := `DELETE FROM "zpSessions" WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.ErrorWithFields("Failed to delete session", map[string]interface{}{
			"session_id": id,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return session.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) UpdateConnectionStatus(ctx context.Context, id string, isConnected bool) error {
	query := `UPDATE "zpSessions" SET "isConnected" = $1, "updatedAt" = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, isConnected, time.Now(), id)
	if err != nil {
		r.logger.ErrorWithFields("Failed to update session connection status", map[string]interface{}{
			"session_id": id,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to update session connection status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return session.ErrSessionNotFound
	}

	r.logger.InfoWithFields("Session connection status updated successfully", map[string]interface{}{
		"session_id": id,
	})

	return nil
}

func (r *sessionRepository) UpdateLastSeen(ctx context.Context, id string) error {
	query := `UPDATE "zpSessions" SET "lastSeen" = $1, "updatedAt" = $2 WHERE id = $3`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return session.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) GetActiveSessions(ctx context.Context) ([]*session.Session, error) {
	r.logger.Info("Getting active sessions")

	query := `SELECT * FROM "zpSessions" WHERE "isConnected" = true ORDER BY "createdAt" DESC`

	var models []sessionModel
	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		r.logger.ErrorWithFields("Failed to get active sessions", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	sessions := make([]*session.Session, len(models))
	for i, model := range models {
		sess, err := r.fromModel(&model)
		if err != nil {
			r.logger.ErrorWithFields("Failed to convert session model", map[string]interface{}{
				"session_id": model.ID,
				"error":      err.Error(),
			})
			continue
		}
		sessions[i] = sess
	}

	return sessions, nil
}

func (r *sessionRepository) CountByConnectionStatus(ctx context.Context, isConnected bool) (int, error) {
	query := `SELECT COUNT(*) FROM "zpSessions" WHERE "isConnected" = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, isConnected)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions by connection status: %w", err)
	}

	return count, nil
}

func (r *sessionRepository) toModel(sess *session.Session) *sessionModel {
	model := &sessionModel{
		ID:          sess.ID.String(),
		Name:        sess.Name,
		IsConnected: sess.IsConnected,
		CreatedAt:   sess.CreatedAt,
		UpdatedAt:   sess.UpdatedAt,
	}

	if sess.DeviceJid != "" {
		model.DeviceJid = sql.NullString{String: sess.DeviceJid, Valid: true}
	}

	if sess.ProxyConfig != nil {
		proxyJSON, err := json.Marshal(sess.ProxyConfig)
		if err == nil {
			model.ProxyConfig = sql.NullString{String: string(proxyJSON), Valid: true}
		}
	}

	if sess.ConnectionError != nil && *sess.ConnectionError != "" {
		model.ConnectionError = sql.NullString{String: *sess.ConnectionError, Valid: true}
	}

	if sess.QRCode != "" {
		model.QRCode = sql.NullString{String: sess.QRCode, Valid: true}
	}

	if sess.QRCodeExpiresAt != nil && !sess.QRCodeExpiresAt.IsZero() {
		model.QRCodeExpiresAt = sql.NullTime{Time: *sess.QRCodeExpiresAt, Valid: true}
	}

	if sess.LastSeen != nil && !sess.LastSeen.IsZero() {
		model.LastSeen = sql.NullTime{Time: *sess.LastSeen, Valid: true}
	}

	if sess.ConnectedAt != nil && !sess.ConnectedAt.IsZero() {
		model.ConnectedAt = sql.NullTime{Time: *sess.ConnectedAt, Valid: true}
	}

	return model
}

func (r *sessionRepository) fromModel(model *sessionModel) (*session.Session, error) {
	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	sess := &session.Session{
		ID:          id,
		Name:        model.Name,
		IsConnected: model.IsConnected,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	if model.DeviceJid.Valid {
		sess.DeviceJid = model.DeviceJid.String
	}

	if model.ConnectionError.Valid {
		sess.ConnectionError = &model.ConnectionError.String
	}

	if model.QRCode.Valid {
		sess.QRCode = model.QRCode.String
	}

	if model.QRCodeExpiresAt.Valid {
		sess.QRCodeExpiresAt = &model.QRCodeExpiresAt.Time
	}

	if model.ProxyConfig.Valid {
		var proxyConfig session.ProxyConfig
		if err := json.Unmarshal([]byte(model.ProxyConfig.String), &proxyConfig); err == nil {
			sess.ProxyConfig = &proxyConfig
		}
	}

	if model.LastSeen.Valid {
		sess.LastSeen = &model.LastSeen.Time
	}

	if model.ConnectedAt.Valid {
		sess.ConnectedAt = &model.ConnectedAt.Time
	}

	return sess, nil
}

// isValidUUID verifica se a string é um UUID válido
func isValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}
