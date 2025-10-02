package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"zpwoot/platform/logger"
)

//go:embed migrations
var migrationsFS embed.FS

type Migration struct {
	AppliedAt *time.Time
	Name      string
	UpSQL     string
	DownSQL   string
	Version   int
}

type Migrator struct {
	db     *Database
	logger *logger.Logger
}

func NewMigrator(db *Database, logger *logger.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

func (m *Migrator) RunMigrations() error {
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	pendingCount := 0
	for _, migration := range migrations {
		if !m.isMigrationApplied(migration.Version, appliedMigrations) {
			if err := m.executeMigration(migration); err != nil {
				return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
			}
			pendingCount++
		}
	}

	return nil
}

func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS "zpMigrations" (
			"version" INTEGER PRIMARY KEY,
			"name" VARCHAR(255) NOT NULL,
			"appliedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS "idx_zp_migrations_applied_at" ON "zpMigrations" ("appliedAt");
		
		COMMENT ON TABLE "zpMigrations" IS 'Database migrations tracking table';
		COMMENT ON COLUMN "zpMigrations"."version" IS 'Migration version number';
		COMMENT ON COLUMN "zpMigrations"."name" IS 'Migration name';
		COMMENT ON COLUMN "zpMigrations"."appliedAt" IS 'When migration was applied';
	`

	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

func (m *Migrator) loadMigrations() ([]*Migration, error) {
	entries, err := m.readMigrationDirectory()
	if err != nil {
		return nil, err
	}

	migrationFiles, err := m.processMigrationFiles(entries)
	if err != nil {
		return nil, err
	}

	migrations := m.buildMigrationObjects(migrationFiles)

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) readMigrationDirectory() ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}
	return entries, nil
}

func (m *Migrator) processMigrationFiles(entries []fs.DirEntry) (map[int]map[string]string, error) {
	migrationFiles := make(map[int]map[string]string)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version, err := m.extractVersionFromFilename(entry.Name())
		if err != nil {
			m.logger.WarnWithFields("Skipping invalid migration file", map[string]interface{}{
				"filename": entry.Name(),
				"error":    err.Error(),
			})
			continue
		}

		content, err := m.readMigrationFile(entry.Name())
		if err != nil {
			return nil, err
		}

		if migrationFiles[version] == nil {
			migrationFiles[version] = make(map[string]string)
		}

		m.categorizeMigrationFile(entry.Name(), content, migrationFiles[version])
	}

	return migrationFiles, nil
}

func (m *Migrator) extractVersionFromFilename(filename string) (int, error) {
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid filename format")
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid version number: %w", err)
	}

	return version, nil
}

func (m *Migrator) readMigrationFile(filename string) (string, error) {
	content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", filename))
	if err != nil {
		return "", fmt.Errorf("failed to read migration file %s: %w", filename, err)
	}
	return string(content), nil
}

func (m *Migrator) categorizeMigrationFile(filename, content string, files map[string]string) {
	if strings.Contains(filename, ".up.sql") {
		files["up"] = content
		nameParts := strings.Split(filename, "_")
		if len(nameParts) > 1 {
			name := strings.Join(nameParts[1:], "_")
			name = strings.TrimSuffix(name, ".up.sql")
			files["name"] = name
		}
	} else if strings.Contains(filename, ".down.sql") {
		files["down"] = content
	}
}

func (m *Migrator) buildMigrationObjects(migrationFiles map[int]map[string]string) []*Migration {
	migrations := make([]*Migration, 0, len(migrationFiles))

	for version, files := range migrationFiles {
		migration := &Migration{
			Version: version,
			Name:    files["name"],
			UpSQL:   files["up"],
			DownSQL: files["down"],
		}

		if migration.UpSQL == "" {
			m.logger.WarnWithFields("Migration missing up.sql file", map[string]interface{}{
				"version": version,
			})
			continue
		}

		migrations = append(migrations, migration)
	}

	return migrations
}

func (m *Migrator) getAppliedMigrations() (map[int]bool, error) {
	query := `SELECT "version" FROM "zpMigrations" ORDER BY "version"`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			m.logger.Error("Failed to close rows: " + err.Error())
		}
	}()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

func (m *Migrator) isMigrationApplied(version int, appliedMigrations map[int]bool) bool {
	return appliedMigrations[version]
}

func (m *Migrator) executeMigration(migration *Migration) error {
	m.logger.InfoWithFields("Applying migration", map[string]interface{}{
		"version": migration.Version,
		"name":    migration.Name,
	})

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var committed bool
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				m.logger.Error("Failed to rollback transaction: " + rollbackErr.Error())
			}
		}
	}()

	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	insertQuery := `
		INSERT INTO "zpMigrations" ("version", "name", "appliedAt")
		VALUES ($1, $2, NOW())
	`
	if _, err := tx.Exec(insertQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}
	committed = true

	m.logger.InfoWithFields("Migration applied successfully", map[string]interface{}{
		"version": migration.Version,
		"name":    migration.Name,
	})

	return nil
}

func (m *Migrator) Rollback() error {
	m.logger.Info("Rolling back last migration...")

	version, name, err := m.getLastMigration()
	if err != nil {
		return err
	}

	if version == 0 {
		m.logger.Info("No migrations to rollback")
		return nil
	}

	targetMigration, err := m.findTargetMigration(version)
	if err != nil {
		return err
	}

	return m.executeRollback(targetMigration, version, name)
}

func (m *Migrator) getLastMigration() (int, string, error) {
	query := `SELECT "version", "name" FROM "zpMigrations" ORDER BY "version" DESC LIMIT 1`

	var version int
	var name string
	err := m.db.QueryRow(query).Scan(&version, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", nil
		}
		return 0, "", fmt.Errorf("failed to get last migration: %w", err)
	}

	return version, name, nil
}

func (m *Migrator) findTargetMigration(version int) (*Migration, error) {
	migrations, err := m.loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, migration := range migrations {
		if migration.Version == version {
			if migration.DownSQL == "" {
				return nil, fmt.Errorf("migration %d has no down SQL", version)
			}
			return migration, nil
		}
	}

	return nil, fmt.Errorf("migration %d not found in files", version)
}

func (m *Migrator) executeRollback(targetMigration *Migration, version int, name string) error {
	m.logger.InfoWithFields("Rolling back migration", map[string]interface{}{
		"version": version,
		"name":    name,
	})

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var committed bool
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				m.logger.Error("Failed to rollback transaction: " + rollbackErr.Error())
			}
		}
	}()

	if _, err := tx.Exec(targetMigration.DownSQL); err != nil {
		return fmt.Errorf("failed to execute rollback SQL: %w", err)
	}

	deleteQuery := `DELETE FROM "zpMigrations" WHERE "version" = $1`
	if _, err := tx.Exec(deleteQuery, version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}
	committed = true

	m.logger.InfoWithFields("Migration rolled back successfully", map[string]interface{}{
		"version": version,
		"name":    name,
	})

	return nil
}

func (m *Migrator) GetMigrationStatus() ([]*Migration, error) {
	migrations, err := m.loadMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if appliedMigrations[migration.Version] {
			now := time.Now()
			migration.AppliedAt = &now
		}
	}

	return migrations, nil
}
