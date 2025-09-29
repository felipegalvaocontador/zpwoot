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

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Transaction(fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				_ = rollbackErr // Explicitly ignore rollback error
			}
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				_ = rollbackErr // Explicitly ignore rollback error
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

func (db *DB) Health() error {
	return db.Ping()
}

func (db *DB) GetDB() *sqlx.DB {
	return db.DB
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(query, args...)
}

func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Get(dest, query, args...)
}

func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Select(dest, query, args...)
}
