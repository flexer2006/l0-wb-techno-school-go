package di

import (
	"context"
	"errors"
	"fmt"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres/connect"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres/migration"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

var (
	ErrDatabasePoolCreation  = errors.New("failed to create database pool")
	ErrDatabaseHealthCheck   = errors.New("database health check failed")
	ErrMigrationRunnerCreate = errors.New("failed to create migration runner")
	ErrMigrationVersion      = errors.New("failed to get migration version")
	ErrDatabaseDirtyState    = errors.New("database is in dirty state")
	ErrMigrationApply        = errors.New("failed to apply migrations")
)

func NewDatabase(ctx context.Context, cfg config.DatabaseConfig, log logger.Logger) (*connect.DB, error) {
	dsn := connect.BuildDSN(cfg)

	log.Info("connecting to database",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
	)

	database, err := connect.NewPool(
		ctx,
		dsn,
		cfg.MaxOpenConns,
		cfg.MaxIdleConns,
		cfg.ConnMaxLifetime,
		cfg.ConnMaxIdleTime,
		cfg.Timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDatabasePoolCreation, err)
	}

	log.Info("database connection established successfully")

	if err := database.Ping(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("%w: %w", ErrDatabaseHealthCheck, err)
	}

	currentVersion, err := getDatabaseVersion(dsn, cfg.MigrationsPath, log)
	if err != nil {
		log.Warn("failed to get current database version", "error", err)
	} else {
		log.Info("current database migration version", "version", currentVersion)
	}

	if err := runMigrations(dsn, cfg.MigrationsPath, log); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	finalVersion, err := getDatabaseVersion(dsn, cfg.MigrationsPath, log)
	if err != nil {
		log.Warn("failed to get final database version", "error", err)
	} else {
		if finalVersion != currentVersion {
			log.Info("database migrations applied successfully",
				"from_version", currentVersion,
				"to_version", finalVersion)
		} else {
			log.Info("database schema is up to date", "version", finalVersion)
		}
	}

	log.Info("database migrations completed successfully")
	return database, nil
}

func getDatabaseVersion(dsn, migrationsPath string, log logger.Logger) (uint, error) {
	runner, err := migration.NewMigrationRunner(dsn, migrationsPath)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrMigrationRunnerCreate, err)
	}
	defer func() {
		if closeErr := runner.Close(); closeErr != nil {
			log.Error("failed to close migration runner", "error", closeErr)
		}
	}()

	version, dirty, err := runner.Version()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrMigrationVersion, err)
	}

	if dirty {
		log.Error("database is in dirty state, manual intervention required", "version", version)
		return 0, fmt.Errorf("%w at version %d", ErrDatabaseDirtyState, version)
	}

	return version, nil
}

func runMigrations(dsn, migrationsPath string, log logger.Logger) error {
	log.Info("starting database migrations", "path", migrationsPath)

	runner, err := migration.NewMigrationRunner(dsn, migrationsPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMigrationRunnerCreate, err)
	}
	defer func() {
		if closeErr := runner.Close(); closeErr != nil {
			log.Error("failed to close migration runner", "error", closeErr)
		}
	}()

	version, dirty, err := runner.Version()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMigrationVersion, err)
	}

	if dirty {
		log.Warn("database is in dirty state during migration", "version", version)
	}

	if err := runner.Up(); err != nil {
		return fmt.Errorf("%w: %w", ErrMigrationApply, err)
	}

	return nil
}
