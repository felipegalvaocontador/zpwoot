package ports

import (
	"context"

	"zpwoot/internal/domain/session"
)

type SessionRepository interface {
	Create(ctx context.Context, session *session.Session) error
	GetByID(ctx context.Context, id string) (*session.Session, error)
	GetByName(ctx context.Context, name string) (*session.Session, error)
	GetByDeviceJid(ctx context.Context, deviceJid string) (*session.Session, error)
	List(ctx context.Context, req *session.ListSessionsRequest) ([]*session.Session, int, error)
	Update(ctx context.Context, session *session.Session) error
	Delete(ctx context.Context, id string) error
	UpdateConnectionStatus(ctx context.Context, id string, isConnected bool) error
	UpdateLastSeen(ctx context.Context, id string) error
	GetActiveSessions(ctx context.Context) ([]*session.Session, error)
	CountByConnectionStatus(ctx context.Context, isConnected bool) (int, error)
}
