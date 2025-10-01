package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/core/media"
)

// MediaRepository implementa a interface media.Repository para PostgreSQL
// Nota: Este repositório é um stub básico pois o legacy não possui tabela de mídia
// A mídia é gerenciada via cache em arquivo e API do WhatsApp
type MediaRepository struct {
	db *sqlx.DB
}

// NewMediaRepository cria uma nova instância do repositório de mídia
func NewMediaRepository(db *sqlx.DB) media.Repository {
	return &MediaRepository{
		db: db,
	}
}

// Create cria um novo item de mídia (stub - implementação futura)
func (r *MediaRepository) Create(ctx context.Context, mediaItem *media.MediaItem) error {
	return fmt.Errorf("media repository not implemented - media is managed via file cache and WhatsApp API")
}

// GetByID busca um item de mídia pelo ID (stub - implementação futura)
func (r *MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*media.MediaItem, error) {
	return nil, fmt.Errorf("media repository not implemented - media is managed via file cache and WhatsApp API")
}

// GetByMessageID busca mídia pelo ID da mensagem (stub - implementação futura)
func (r *MediaRepository) GetByMessageID(ctx context.Context, sessionID uuid.UUID, messageID string) (*media.MediaItem, error) {
	return nil, fmt.Errorf("media repository not implemented - media is managed via file cache and WhatsApp API")
}

// Update atualiza um item de mídia (stub - implementação futura)
func (r *MediaRepository) Update(ctx context.Context, mediaItem *media.MediaItem) error {
	return fmt.Errorf("media repository not implemented - media is managed via file cache and WhatsApp API")
}

// Delete remove um item de mídia (stub - implementação futura)
func (r *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("media repository not implemented - media is managed via file cache and WhatsApp API")
}

// List retorna uma lista de itens de mídia (stub - implementação futura)
func (r *MediaRepository) List(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*media.MediaItem, error) {
	return nil, fmt.Errorf("media repository not implemented - media is managed via file cache and WhatsApp API")
}

// Count retorna o número total de itens de mídia (stub - implementação futura)
func (r *MediaRepository) Count(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	return 0, fmt.Errorf("media repository not implemented - media is managed via file cache and WhatsApp API")
}