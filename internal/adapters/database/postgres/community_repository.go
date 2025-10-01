package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/core/community"
)

// CommunityRepository implementa a interface community.Repository para PostgreSQL
// Nota: Este repositório é um stub básico pois o legacy não possui funcionalidade de community
type CommunityRepository struct {
	db *sqlx.DB
}

// NewCommunityRepository cria uma nova instância do repositório de community
func NewCommunityRepository(db *sqlx.DB) community.Repository {
	return &CommunityRepository{
		db: db,
	}
}

// Create cria uma nova community (stub - implementação futura)
func (r *CommunityRepository) Create(ctx context.Context, community *community.Community) error {
	return fmt.Errorf("community repository not implemented - feature not available in legacy")
}

// GetByID busca uma community pelo ID (stub - implementação futura)
func (r *CommunityRepository) GetByID(ctx context.Context, id uuid.UUID) (*community.Community, error) {
	return nil, fmt.Errorf("community repository not implemented - feature not available in legacy")
}

// Update atualiza uma community (stub - implementação futura)
func (r *CommunityRepository) Update(ctx context.Context, community *community.Community) error {
	return fmt.Errorf("community repository not implemented - feature not available in legacy")
}

// Delete remove uma community (stub - implementação futura)
func (r *CommunityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("community repository not implemented - feature not available in legacy")
}

// List retorna uma lista de communities (stub - implementação futura)
func (r *CommunityRepository) List(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*community.Community, error) {
	return nil, fmt.Errorf("community repository not implemented - feature not available in legacy")
}

// Count retorna o número total de communities (stub - implementação futura)
func (r *CommunityRepository) Count(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	return 0, fmt.Errorf("community repository not implemented - feature not available in legacy")
}