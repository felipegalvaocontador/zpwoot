package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/core/newsletter"
)

// NewsletterRepository implementa a interface newsletter.Repository para PostgreSQL
// Nota: Este repositório é um stub básico pois o legacy não possui funcionalidade de newsletter
type NewsletterRepository struct {
	db *sqlx.DB
}

// NewNewsletterRepository cria uma nova instância do repositório de newsletter
func NewNewsletterRepository(db *sqlx.DB) newsletter.Repository {
	return &NewsletterRepository{
		db: db,
	}
}

// Create cria um novo newsletter (stub - implementação futura)
func (r *NewsletterRepository) Create(ctx context.Context, newsletter *newsletter.Newsletter) error {
	return fmt.Errorf("newsletter repository not implemented - feature not available in legacy")
}

// GetByID busca um newsletter pelo ID (stub - implementação futura)
func (r *NewsletterRepository) GetByID(ctx context.Context, id uuid.UUID) (*newsletter.Newsletter, error) {
	return nil, fmt.Errorf("newsletter repository not implemented - feature not available in legacy")
}

// Update atualiza um newsletter (stub - implementação futura)
func (r *NewsletterRepository) Update(ctx context.Context, newsletter *newsletter.Newsletter) error {
	return fmt.Errorf("newsletter repository not implemented - feature not available in legacy")
}

// Delete remove um newsletter (stub - implementação futura)
func (r *NewsletterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("newsletter repository not implemented - feature not available in legacy")
}

// List retorna uma lista de newsletters (stub - implementação futura)
func (r *NewsletterRepository) List(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*newsletter.Newsletter, error) {
	return nil, fmt.Errorf("newsletter repository not implemented - feature not available in legacy")
}

// Count retorna o número total de newsletters (stub - implementação futura)
func (r *NewsletterRepository) Count(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	return 0, fmt.Errorf("newsletter repository not implemented - feature not available in legacy")
}