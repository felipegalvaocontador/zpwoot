// Refactored: separated responsibilities; improved builder pattern; standardized error handling
package wameow

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WameowLogger adapts our logger to whatsmeow's logger interface
type WameowLogger struct {
	logger *logger.Logger
	module string
}

// NewWameowLogger creates a new whatsmeow logger adapter
func NewWameowLogger(logger *logger.Logger) waLog.Logger {
	return &WameowLogger{
		logger: logger,
		module: "whatsmeow",
	}
}

func (w *WameowLogger) Errorf(msg string, args ...interface{}) {
	w.logger.ErrorWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": w.module,
	})
}

func (w *WameowLogger) Warnf(msg string, args ...interface{}) {
	w.logger.WarnWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": w.module,
	})
}

func (w *WameowLogger) Infof(msg string, args ...interface{}) {
	w.logger.InfoWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": w.module,
	})
}

func (w *WameowLogger) Debugf(msg string, args ...interface{}) {
	w.logger.DebugWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": w.module,
	})
}

func (w *WameowLogger) Sub(module string) waLog.Logger {
	return &WameowLogger{
		logger: w.logger,
		module: fmt.Sprintf("%s.%s", w.module, module),
	}
}

// Factory creates and configures wameow components
type Factory struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

// NewFactory creates a new factory instance
func NewFactory(logger *logger.Logger, sessionRepo ports.SessionRepository) (*Factory, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if sessionRepo == nil {
		return nil, fmt.Errorf("session repository cannot be nil")
	}

	return &Factory{
		logger:      logger,
		sessionRepo: sessionRepo,
	}, nil
}

// CreateManager creates a new manager with the given database connection
func (f *Factory) CreateManager(db *sql.DB) (*Manager, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	container, err := f.createSQLStoreContainer(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL store container: %w", err)
	}

	manager := NewManager(container, f.sessionRepo, f.logger)
	return manager, nil
}

func (f *Factory) createSQLStoreContainer(db *sql.DB) (*sqlstore.Container, error) {
	waLogger := NewWameowLogger(f.logger)

	container := sqlstore.NewWithDB(db, "postgres", waLogger)
	if container == nil {
		return nil, fmt.Errorf("sqlstore.NewWithDB returned nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := container.Upgrade(ctx); err != nil {
		return nil, fmt.Errorf("failed to upgrade database schema: %w", err)
	}

	return container, nil
}

// ManagerBuilder provides a fluent interface for building Manager instances
type ManagerBuilder struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
	db          *sql.DB
}

// NewManagerBuilder creates a new manager builder
func NewManagerBuilder() *ManagerBuilder {
	return &ManagerBuilder{}
}

// WithLogger sets the logger for the manager
func (b *ManagerBuilder) WithLogger(logger *logger.Logger) *ManagerBuilder {
	b.logger = logger
	return b
}

// WithSessionRepository sets the session repository for the manager
func (b *ManagerBuilder) WithSessionRepository(repo ports.SessionRepository) *ManagerBuilder {
	b.sessionRepo = repo
	return b
}

// WithDatabase sets the database connection for the manager
func (b *ManagerBuilder) WithDatabase(db *sql.DB) *ManagerBuilder {
	b.db = db
	return b
}

// Build creates and returns a new Manager instance
func (b *ManagerBuilder) Build() (*Manager, error) {
	if err := b.validate(); err != nil {
		return nil, fmt.Errorf("builder validation failed: %w", err)
	}

	factory, err := NewFactory(b.logger, b.sessionRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create factory: %w", err)
	}

	return factory.CreateManager(b.db)
}

// validate checks that all required fields are set
func (b *ManagerBuilder) validate() error {
	if b.logger == nil {
		return fmt.Errorf("logger is required")
	}
	if b.sessionRepo == nil {
		return fmt.Errorf("session repository is required")
	}
	if b.db == nil {
		return fmt.Errorf("database connection is required")
	}
	return nil
}

// HealthCheck returns the health status of the manager and all sessions
func (m *Manager) HealthCheck() map[string]interface{} {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	stats := m.calculateSessionStats()

	m.logger.DebugWithFields("Health check performed", map[string]interface{}{
		"total_sessions":     stats.Total,
		"connected_sessions": stats.Connected,
		"logged_in_sessions": stats.LoggedIn,
	})

	return map[string]interface{}{
		"total_sessions":     stats.Total,
		"connected_sessions": stats.Connected,
		"logged_in_sessions": stats.LoggedIn,
		"healthy":            stats.Total == 0 || stats.Connected > 0,
		"timestamp":          time.Now().Unix(),
		"uptime_seconds":     time.Since(time.Now()).Seconds(), // This would need to be tracked properly
	}
}

// GetStats returns detailed statistics about the manager
func (m *Manager) GetStats() map[string]interface{} {
	healthData := m.HealthCheck()

	// Add additional stats
	healthData["version"] = "1.0.0"   // This should come from build info
	healthData["go_version"] = "1.21" // This should come from runtime

	return healthData
}

// sessionStats holds session statistics
type sessionStats struct {
	Total     int
	Connected int
	LoggedIn  int
}

// calculateSessionStats calculates session statistics
func (m *Manager) calculateSessionStats() sessionStats {
	stats := sessionStats{
		Total: len(m.clients),
	}

	for _, client := range m.clients {
		if client.IsConnected() {
			stats.Connected++
		}
		if client.IsLoggedIn() {
			stats.LoggedIn++
		}
	}

	return stats
}

// LogLevelToWALevel converts our log levels to whatsmeow log levels
func LogLevelToWALevel(level string) string {
	levelMap := map[string]string{
		"ERROR": "error",
		"WARN":  "warn",
		"INFO":  "info",
		"DEBUG": "debug",
	}

	if waLevel, exists := levelMap[level]; exists {
		return waLevel
	}

	return "info" // default level
}
