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

type WameowLogger struct {
	logger *logger.Logger
	module string
}

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

type Factory struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

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

type ManagerBuilder struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
	db          *sql.DB
}

func NewManagerBuilder() *ManagerBuilder {
	return &ManagerBuilder{}
}

func (b *ManagerBuilder) WithLogger(logger *logger.Logger) *ManagerBuilder {
	b.logger = logger
	return b
}

func (b *ManagerBuilder) WithSessionRepository(repo ports.SessionRepository) *ManagerBuilder {
	b.sessionRepo = repo
	return b
}

func (b *ManagerBuilder) WithDatabase(db *sql.DB) *ManagerBuilder {
	b.db = db
	return b
}

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

func (m *Manager) GetStats() map[string]interface{} {
	healthData := m.HealthCheck()

	healthData["version"] = "1.0.0"   // This should come from build info
	healthData["go_version"] = "1.21" // This should come from runtime

	return healthData
}

type sessionStats struct {
	Total     int
	Connected int
	LoggedIn  int
}

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
