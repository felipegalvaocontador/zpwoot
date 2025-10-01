package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/core/contacts"
)

// ContactRepository implementa a interface contacts.Repository para PostgreSQL
// Nota: Este repositório é um stub básico pois o legacy não possui tabela de contatos
// Os contatos são gerenciados via API do WhatsApp e Chatwoot
type ContactRepository struct {
	db *sqlx.DB
}

// NewContactRepository cria uma nova instância do repositório de contatos
func NewContactRepository(db *sqlx.DB) contacts.Repository {
	return &ContactRepository{
		db: db,
	}
}

// Create cria um novo contato (stub - implementação futura)
func (r *ContactRepository) Create(ctx context.Context, contact *contacts.Contact) error {
	return fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}

// GetByID busca um contato pelo ID (stub - implementação futura)
func (r *ContactRepository) GetByID(ctx context.Context, id uuid.UUID) (*contacts.Contact, error) {
	return nil, fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}

// GetByJID busca um contato pelo JID (stub - implementação futura)
func (r *ContactRepository) GetByJID(ctx context.Context, sessionID uuid.UUID, jid string) (*contacts.Contact, error) {
	return nil, fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}

// Update atualiza um contato (stub - implementação futura)
func (r *ContactRepository) Update(ctx context.Context, contact *contacts.Contact) error {
	return fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}

// Delete remove um contato (stub - implementação futura)
func (r *ContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}

// List retorna uma lista de contatos (stub - implementação futura)
func (r *ContactRepository) List(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*contacts.Contact, error) {
	return nil, fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}

// Search busca contatos por nome/telefone (stub - implementação futura)
func (r *ContactRepository) Search(ctx context.Context, sessionID uuid.UUID, query string, limit, offset int) ([]*contacts.Contact, error) {
	return nil, fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}

// Count retorna o número total de contatos (stub - implementação futura)
func (r *ContactRepository) Count(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	return 0, fmt.Errorf("contact repository not implemented - contacts are managed via WhatsApp/Chatwoot APIs")
}