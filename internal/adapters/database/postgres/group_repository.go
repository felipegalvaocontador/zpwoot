package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/core/groups"
)

// GroupRepository implementa a interface groups.Repository para PostgreSQL
// Nota: Este repositório é um stub básico pois o legacy não possui tabela de grupos
// Os grupos são gerenciados via API do WhatsApp
type GroupRepository struct {
	db *sqlx.DB
}

// NewGroupRepository cria uma nova instância do repositório de grupos
func NewGroupRepository(db *sqlx.DB) groups.Repository {
	return &GroupRepository{
		db: db,
	}
}

// Create cria um novo grupo (stub - implementação futura)
func (r *GroupRepository) Create(ctx context.Context, group *groups.Group) error {
	return fmt.Errorf("group repository not implemented - groups are managed via WhatsApp API")
}

// GetByID busca um grupo pelo ID (stub - implementação futura)
func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*groups.Group, error) {
	return nil, fmt.Errorf("group repository not implemented - groups are managed via WhatsApp API")
}

// GetByJID busca um grupo pelo JID (stub - implementação futura)
func (r *GroupRepository) GetByJID(ctx context.Context, sessionID uuid.UUID, jid string) (*groups.Group, error) {
	return nil, fmt.Errorf("group repository not implemented - groups are managed via WhatsApp API")
}

// Update atualiza um grupo (stub - implementação futura)
func (r *GroupRepository) Update(ctx context.Context, group *groups.Group) error {
	return fmt.Errorf("group repository not implemented - groups are managed via WhatsApp API")
}

// Delete remove um grupo (stub - implementação futura)
func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("group repository not implemented - groups are managed via WhatsApp API")
}

// List retorna uma lista de grupos (stub - implementação futura)
func (r *GroupRepository) List(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*groups.Group, error) {
	return nil, fmt.Errorf("group repository not implemented - groups are managed via WhatsApp API")
}

// Count retorna o número total de grupos (stub - implementação futura)
func (r *GroupRepository) Count(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	return 0, fmt.Errorf("group repository not implemented - groups are managed via WhatsApp API")
}