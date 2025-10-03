package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"zpwoot/internal/core/session"
)

// SessionResolverService implements session.SessionResolver interface
type SessionResolverService struct {
	repository session.Repository
}

// NewSessionResolver creates a new session resolver service
func NewSessionResolver(repository session.Repository) session.SessionResolver {
	return &SessionResolverService{
		repository: repository,
	}
}

// ResolveToID resolves a session name to its UUID for internal operations
func (r *SessionResolverService) ResolveToID(ctx context.Context, sessionName string) (uuid.UUID, error) {
	// First try to parse as UUID (for backward compatibility)
	if id, err := uuid.Parse(sessionName); err == nil {
		// Verify that this UUID exists
		_, err := r.repository.GetByID(ctx, id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("session with ID %s not found: %w", sessionName, err)
		}
		return id, nil
	}

	// If not a UUID, treat as session name
	sess, err := r.repository.GetByName(ctx, sessionName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("session with name '%s' not found: %w", sessionName, err)
	}

	return sess.ID, nil
}

// Resolve resolves a session name to complete session information
func (r *SessionResolverService) Resolve(ctx context.Context, sessionName string) (*session.ResolveResult, error) {
	// First try to parse as UUID (for backward compatibility)
	if id, err := uuid.Parse(sessionName); err == nil {
		// Get session by UUID
		sess, err := r.repository.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("session with ID %s not found: %w", sessionName, err)
		}
		
		return &session.ResolveResult{
			ID:      sess.ID,
			Name:    sess.Name,
			Session: sess,
		}, nil
	}

	// If not a UUID, treat as session name
	sess, err := r.repository.GetByName(ctx, sessionName)
	if err != nil {
		return nil, fmt.Errorf("session with name '%s' not found: %w", sessionName, err)
	}

	return &session.ResolveResult{
		ID:      sess.ID,
		Name:    sess.Name,
		Session: sess,
	}, nil
}
