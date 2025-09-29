package chatwoot

import (
	"context"
	"fmt"
	"sync"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// Manager implements the ChatwootManager interface
type Manager struct {
	logger     *logger.Logger
	repository ports.ChatwootRepository
	clients    map[string]*Client
	configs    map[string]*ports.ChatwootConfig
	mu         sync.RWMutex
}

// NewManager creates a new Chatwoot manager
func NewManager(logger *logger.Logger, repository ports.ChatwootRepository) *Manager {
	return &Manager{
		logger:     logger,
		repository: repository,
		clients:    make(map[string]*Client),
		configs:    make(map[string]*ports.ChatwootConfig),
	}
}

// GetClient returns a Chatwoot client for the given session
func (m *Manager) GetClient(sessionID string) (ports.ChatwootClient, error) {
	m.mu.RLock()
	client, exists := m.clients[sessionID]
	m.mu.RUnlock()

	if exists {
		return client, nil
	}

	// Load config and create client
	config, err := m.GetConfig(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for session %s: %w", sessionID, err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("chatwoot integration is disabled for session %s", sessionID)
	}

	client = NewClient(config.URL, config.Token, config.AccountID, m.logger)

	m.mu.Lock()
	m.clients[sessionID] = client
	m.mu.Unlock()

	return client, nil
}

// IsEnabled checks if Chatwoot integration is enabled for a session
func (m *Manager) IsEnabled(sessionID string) bool {
	config, err := m.GetConfig(sessionID)
	if err != nil {
		// Only log as error if it's not a "not found" error
		if err.Error() != "failed to get config from repository: chatwoot config not found" {
			m.logger.ErrorWithFields("Failed to check if Chatwoot is enabled", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
		// Chatwoot is simply not configured, which is normal
		return false
	}

	return config.Enabled
}

// InitInstanceChatwoot initializes Chatwoot integration for a session
// Note: This method now contains minimal logic - business rules moved to domain service
func (m *Manager) InitInstanceChatwoot(sessionID, inboxName, webhookURL string, autoCreate bool) error {
	// Basic validation - technical concerns only
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	// Note: Complex inbox creation logic moved to domain service
	// This method should be refactored to receive the target inbox from domain layer
	// For now, we maintain basic functionality but remove business logic

	if autoCreate {
		m.logger.InfoWithFields("Auto-create inbox requested", map[string]interface{}{
			"session_id":  sessionID,
			"inbox_name":  inboxName,
			"webhook_url": webhookURL,
		})

		// TODO: This logic should be moved to application layer using domain service
		// The application layer should:
		// 1. Call domain service to determine if inbox should be created
		// 2. Call domain service to process inbox initialization
		// 3. Call this manager only for technical operations
	}

	return nil
}

// SetConfig sets the Chatwoot configuration for a session
func (m *Manager) SetConfig(sessionID string, config *ports.ChatwootConfig) error {
	// Store in repository
	ctx := context.Background()
	err := m.repository.UpdateConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to update config in repository: %w", err)
	}

	// Update cache
	m.mu.Lock()
	m.configs[sessionID] = config
	// Clear client cache to force recreation with new config
	delete(m.clients, sessionID)
	m.mu.Unlock()

	return nil
}

// GetConfig gets the Chatwoot configuration for a session
func (m *Manager) GetConfig(sessionID string) (*ports.ChatwootConfig, error) {
	m.mu.RLock()
	config, exists := m.configs[sessionID]
	m.mu.RUnlock()

	if exists {
		return config, nil
	}

	// Load from repository
	ctx := context.Background()
	config, err := m.repository.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from repository: %w", err)
	}

	// Cache it
	m.mu.Lock()
	m.configs[sessionID] = config
	m.mu.Unlock()

	return config, nil
}

// Cleanup cleans up resources for a session
func (m *Manager) Cleanup(sessionID string) error {
	m.mu.Lock()
	delete(m.clients, sessionID)
	delete(m.configs, sessionID)
	m.mu.Unlock()

	return nil
}

// Note: createBotContact method removed - business logic moved to domain service
// Bot contact creation should be handled by domain service and called from application layer

// GetStats returns statistics for Chatwoot integration
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"active_clients": len(m.clients),
		"cached_configs": len(m.configs),
	}
}
