package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

// Database wrapper para sqlx.DB com funcionalidades específicas
type Database struct {
	*sqlx.DB
	config config.DatabaseConfig
	logger *logger.Logger
}

// New cria nova conexão com banco de dados
func New(cfg config.DatabaseConfig, log *logger.Logger) (*Database, error) {
	// Conectar ao banco
	db, err := sqlx.Connect("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configurar pool de conexões
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Testar conexão
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		DB:     db,
		config: cfg,
		logger: log,
	}

	// Database connected successfully

	return database, nil
}

// NewFromAppConfig cria database a partir da configuração da aplicação
func NewFromAppConfig(appConfig *config.Config, log *logger.Logger) (*Database, error) {
	return New(appConfig.Database, log)
}

// Close fecha conexão com banco de dados
func (d *Database) Close() error {
	d.logger.InfoWithFields("Closing database connection", map[string]interface{}{
		"module": "database",
	})
	return d.DB.Close()
}

// Health verifica saúde da conexão
func (d *Database) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return d.PingContext(ctx)
}

// Transaction executa função dentro de uma transação
func (d *Database) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := d.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				d.logger.ErrorWithFields("Failed to rollback transaction after panic", map[string]interface{}{
					"error": rollbackErr.Error(),
					"panic": p,
				})
			}
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				d.logger.ErrorWithFields("Failed to rollback transaction", map[string]interface{}{
					"error":          rollbackErr.Error(),
					"original_error": err.Error(),
				})
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(tx)
	return err
}

// Stats retorna estatísticas do pool de conexões
func (d *Database) Stats() sql.DBStats {
	return d.DB.Stats()
}

// GetConfig retorna configuração do banco
func (d *Database) GetConfig() config.DatabaseConfig {
	return d.config
}

// ===== MÉTODOS DE CONVENIÊNCIA =====

// ExecContext executa query com contexto
func (d *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := d.DB.ExecContext(ctx, query, args...)
	
	d.logQuery("EXEC", query, time.Since(start), err)
	return result, err
}

// QueryContext executa query de seleção com contexto
func (d *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.DB.QueryContext(ctx, query, args...)
	
	d.logQuery("QUERY", query, time.Since(start), err)
	return rows, err
}

// QueryRowContext executa query que retorna uma linha com contexto
func (d *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := d.DB.QueryRowContext(ctx, query, args...)
	
	d.logQuery("QUERY_ROW", query, time.Since(start), nil)
	return row
}

// GetContext executa query e escaneia resultado em dest
func (d *Database) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := d.DB.GetContext(ctx, dest, query, args...)
	
	d.logQuery("GET", query, time.Since(start), err)
	return err
}

// SelectContext executa query e escaneia múltiplos resultados em dest
func (d *Database) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := d.DB.SelectContext(ctx, dest, query, args...)
	
	d.logQuery("SELECT", query, time.Since(start), err)
	return err
}

// NamedExecContext executa query nomeada com contexto
func (d *Database) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := d.DB.NamedExecContext(ctx, query, arg)
	
	d.logQuery("NAMED_EXEC", query, time.Since(start), err)
	return result, err
}

// NamedQueryContext executa query nomeada de seleção com contexto
func (d *Database) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := d.DB.NamedQueryContext(ctx, query, arg)
	
	d.logQuery("NAMED_QUERY", query, time.Since(start), err)
	return rows, err
}

// ===== MÉTODOS PRIVADOS =====

// logQuery registra informações da query executada
func (d *Database) logQuery(operation, query string, duration time.Duration, err error) {
	if !d.logger.IsDebugEnabled() {
		return
	}

	fields := map[string]interface{}{
		"operation":    operation,
		"duration_ms":  duration.Milliseconds(),
		"query_length": len(query),
	}

	if err != nil {
		fields["error"] = err.Error()
		d.logger.ErrorWithFields("Database query failed", fields)
	} else {
		if duration > 100*time.Millisecond {
			d.logger.WarnWithFields("Slow database query", fields)
		} else {
			d.logger.DebugWithFields("Database query executed", fields)
		}
	}
}

// ===== HEALTH CHECK =====

// HealthCheck estrutura para verificação de saúde
type HealthCheck struct {
	Status      string        `json:"status"`
	Latency     time.Duration `json:"latency"`
	Connections DBStats       `json:"connections"`
	Error       string        `json:"error,omitempty"`
}

// DBStats estatísticas do banco
type DBStats struct {
	OpenConnections     int `json:"open_connections"`
	InUse              int `json:"in_use"`
	Idle               int `json:"idle"`
	WaitCount          int64 `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64 `json:"max_idle_closed"`
	MaxIdleTimeClosed  int64 `json:"max_idle_time_closed"`
	MaxLifetimeClosed  int64 `json:"max_lifetime_closed"`
}

// PerformHealthCheck executa verificação completa de saúde
func (d *Database) PerformHealthCheck(ctx context.Context) HealthCheck {
	start := time.Now()
	
	err := d.Health(ctx)
	latency := time.Since(start)
	
	stats := d.Stats()
	dbStats := DBStats{
		OpenConnections:    stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}

	healthCheck := HealthCheck{
		Latency:     latency,
		Connections: dbStats,
	}

	if err != nil {
		healthCheck.Status = "unhealthy"
		healthCheck.Error = err.Error()
	} else {
		healthCheck.Status = "healthy"
	}

	return healthCheck
}
