package connect

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres/migration"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMaxConnsExceeded = errors.New("maxConns exceeds maximum allowed value")
	ErrMinConnsExceeded = errors.New("minConns exceeds maximum allowed value")
	ErrInvalidConnRange = errors.New("minConns cannot be greater than maxConns")
	ErrContextTimeout   = errors.New("database connection timeout exceeded")
)

type DB struct {
	pool *pgxpool.Pool
}

func BuildDSN(cfg config.DatabaseConfig) string {
	dsnURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:   cfg.Database,
	}

	q := dsnURL.Query()
	q.Set("sslmode", cfg.SSLMode)
	dsnURL.RawQuery = q.Encode()

	return dsnURL.String()
}

func NewPool(ctx context.Context, dsn string, maxConns, minConns int, connMaxLifetime, connMaxIdle, timeout time.Duration) (*DB, error) {
	connectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	if maxConns > 0 {
		if maxConns > math.MaxInt32 {
			return nil, fmt.Errorf("%w: %d", ErrMaxConnsExceeded, maxConns)
		}
		cfg.MaxConns = int32(maxConns)
	}

	if minConns > 0 {
		if minConns > math.MaxInt32 {
			return nil, fmt.Errorf("%w: %d", ErrMinConnsExceeded, minConns)
		}
		cfg.MinConns = int32(minConns)
	}

	if maxConns > 0 && minConns > 0 && minConns > maxConns {
		return nil, fmt.Errorf("%w: minConns=%d, maxConns=%d", ErrInvalidConnRange, minConns, maxConns)
	}

	if connMaxLifetime > 0 {
		cfg.MaxConnLifetime = connMaxLifetime
	}

	if connMaxIdle > 0 {
		cfg.MaxConnIdleTime = connMaxIdle
	}

	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		if errors.Is(connectCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: failed to create pool within %v", ErrContextTimeout, timeout)
		}
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		if errors.Is(connectCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: ping failed within %v", ErrContextTimeout, timeout)
		}
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return &DB{pool: pool}, nil
}

func NewPoolWithMigrations(ctx context.Context, cfg config.DatabaseConfig, log logger.Logger) (*DB, error) {
	dsn := BuildDSN(cfg)

	log.Info("connecting to database",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
	)

	database, err := NewPool(
		ctx,
		dsn,
		cfg.MaxOpenConns,
		cfg.MaxIdleConns,
		cfg.ConnMaxLifetime,
		cfg.ConnMaxIdleTime,
		cfg.Timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	log.Info("database connection established successfully")

	if err := database.Ping(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := runMigrations(dsn, cfg.MigrationsPath, log); err != nil {
		database.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	log.Info("database and migrations initialized successfully")
	return database, nil
}

func runMigrations(dsn, migrationsPath string, log logger.Logger) error {
	log.Info("starting database migrations", "path", migrationsPath)

	runner, err := migration.NewMigrationRunner(dsn, migrationsPath)
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}
	defer func() {
		if closeErr := runner.Close(); closeErr != nil {
			log.Error("failed to close migration runner", "error", closeErr)
		}
	}()

	if err := runner.Up(); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func (d *DB) Close() {
	if d.pool != nil {
		d.pool.Close()
	}
}

func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

func (d *DB) Ping(ctx context.Context) error {
	if err := d.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}
