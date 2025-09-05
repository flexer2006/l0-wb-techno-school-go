package migration

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	ErrMigrationInit   = errors.New("failed to initialize migration")
	ErrMigrationFailed = errors.New("migration execution failed")
	ErrMigrationClose  = errors.New("failed to close migrator")
)

type MigrationRunner struct {
	migrator *migrate.Migrate
}

func NewMigrationRunner(dsn, migrationsPath string) (*MigrationRunner, error) {
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMigrationInit, err)
	}

	return &MigrationRunner{migrator: m}, nil
}

func (mr *MigrationRunner) Up() error {
	if err := mr.migrator.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("%w: %w", ErrMigrationFailed, err)
		}
	}
	return nil
}

func (mr *MigrationRunner) Down() error {
	if err := mr.migrator.Down(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("%w: %w", ErrMigrationFailed, err)
		}
	}
	return nil
}

func (mr *MigrationRunner) Version() (uint, bool, error) {
	version, dirty, err := mr.migrator.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

func (mr *MigrationRunner) Close() error {
	sourceErr, databaseErr := mr.migrator.Close()

	if sourceErr != nil && databaseErr != nil {
		return fmt.Errorf("%w: source error: %w, database error: %w", ErrMigrationClose, sourceErr, databaseErr)
	}
	if sourceErr != nil {
		return fmt.Errorf("%w: source error: %w", ErrMigrationClose, sourceErr)
	}
	if databaseErr != nil {
		return fmt.Errorf("%w: database error: %w", ErrMigrationClose, databaseErr)
	}
	return nil
}
