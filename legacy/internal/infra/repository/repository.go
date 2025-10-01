package repository

import (
	"github.com/jmoiron/sqlx"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type Repositories struct {
	Session         ports.SessionRepository
	Webhook         ports.WebhookRepository
	Chatwoot        ports.ChatwootRepository
	ChatwootMessage ports.ChatwootMessageRepository
}

func NewRepositories(db *sqlx.DB, logger *logger.Logger) *Repositories {
	return &Repositories{
		Session:         NewSessionRepository(db, logger),
		Webhook:         NewWebhookRepository(db, logger),
		Chatwoot:        NewChatwootRepository(db, logger),
		ChatwootMessage: NewMessageRepository(db, logger),
	}
}

func (r *Repositories) GetSessionRepository() ports.SessionRepository {
	return r.Session
}

func (r *Repositories) GetWebhookRepository() ports.WebhookRepository {
	return r.Webhook
}

func (r *Repositories) GetChatwootRepository() ports.ChatwootRepository {
	return r.Chatwoot
}

func (r *Repositories) GetChatwootMessageRepository() ports.ChatwootMessageRepository {
	return r.ChatwootMessage
}
