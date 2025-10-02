package db

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"zpwoot/internal/infra/db"
	"zpwoot/platform/logger"
)

type DB struct {
	*sqlx.DB
}

func New(databaseURL string) (*DB, error) {
	sqlxDB, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := sqlxDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: sqlxDB}, nil
}

func NewWithMigrations(databaseURL string, logger *logger.Logger) (*DB, error) {
	database, err := New(databaseURL)
	if err != nil {
		return nil, err
	}

	migrator := db.NewMigrator(database.DB.DB, logger)
	if err := migrator.RunMigrations(); err != nil {
		if closeErr := database.Close(); closeErr != nil {
			logger.Error("Failed to close database after migration error: " + closeErr.Error())
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

func (d *DB) Close() error {
	return d.DB.Close()
}

func (d *DB) Transaction(fn func(*sqlx.Tx) error) error {
	tx, err := d.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				_ = rollbackErr
			}
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				_ = rollbackErr
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

func (d *DB) Health() error {
	return d.Ping()
}

func (d *DB) GetDB() *sqlx.DB {
	return d.DB
}

func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(query, args...)
}

func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.Query(query, args...)
}

func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRow(query, args...)
}

func (d *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return d.DB.Get(dest, query, args...)
}

func (d *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return d.DB.Select(dest, query, args...)
}
