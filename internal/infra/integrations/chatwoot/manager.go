package chatwoot

import (
	"context"
	"fmt"
	"sync"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type Manager struct {
	logger     *logger.Logger
	repository ports.ChatwootRepository
	clients    map[string]*Client
	configs    map[string]*ports.ChatwootConfig
	mu         sync.RWMutex
}

func NewManager(logger *logger.Logger, repository ports.ChatwootRepository) *Manager {
	return &Manager{
		logger:     logger,
		repository: repository,
		clients:    make(map[string]*Client),
		configs:    make(map[string]*ports.ChatwootConfig),
	}
}

func (m *Manager) GetClient(sessionID string) (ports.ChatwootClient, error) {
	m.mu.RLock()
	client, exists := m.clients[sessionID]
	m.mu.RUnlock()

	if exists {
		return client, nil
	}

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

func (m *Manager) IsEnabled(sessionID string) bool {
	config, err := m.GetConfig(sessionID)
	if err != nil {
		if err.Error() != "failed to get config from repository: chatwoot config not found" {
			m.logger.ErrorWithFields("Failed to check if Chatwoot is enabled", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
		return false
	}

	return config.Enabled
}

func (m *Manager) InitInstanceChatwoot(sessionID, inboxName, webhookURL string, autoCreate bool) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	if autoCreate {
		m.logger.InfoWithFields("Auto-create inbox requested", map[string]interface{}{
			"session_id":  sessionID,
			"inbox_name":  inboxName,
			"webhook_url": webhookURL,
		})
	}

	return nil
}

func (m *Manager) SetConfig(sessionID string, config *ports.ChatwootConfig) error {
	ctx := context.Background()
	err := m.repository.UpdateConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to update config in repository: %w", err)
	}

	m.mu.Lock()
	m.configs[sessionID] = config
	delete(m.clients, sessionID)
	m.mu.Unlock()

	return nil
}

func (m *Manager) GetConfig(sessionID string) (*ports.ChatwootConfig, error) {
	m.mu.RLock()
	config, exists := m.configs[sessionID]
	m.mu.RUnlock()

	if exists {
		return config, nil
	}

	ctx := context.Background()
	config, err := m.repository.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from repository: %w", err)
	}

	m.mu.Lock()
	m.configs[sessionID] = config
	m.mu.Unlock()

	return config, nil
}

func (m *Manager) Cleanup(sessionID string) error {
	m.mu.Lock()
	delete(m.clients, sessionID)
	delete(m.configs, sessionID)
	m.mu.Unlock()

	return nil
}

func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"active_clients": len(m.clients),
		"cached_configs": len(m.configs),
	}
}
