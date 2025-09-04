package di

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
)

const AppVersion = "1.0.0"

var (
	ErrShutdownTimeout    = errors.New("shutdown timeout exceeded")
	ErrForcedShutdown     = errors.New("forced shutdown by second signal")
	ErrApplicationStartup = errors.New("application startup failed")
)

func RunService() error {
	cfg := config.MustLoad()

	zapLogger := NewZapLogger(cfg.Logger)

	zapLogger.Info("starting application",
		"version", AppVersion,
		"log_level", cfg.Logger.Level,
		"shutdown_timeout", cfg.Shutdown.Timeout,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := NewDatabase(ctx, cfg.Database, zapLogger)
	if err != nil {
		zapLogger.Error("failed to initialize database", "error", err)
		return fmt.Errorf("%w: database initialization failed: %w", ErrApplicationStartup, err)
	}

	zapLogger.Info("database initialized successfully")

	gracefulShutdown := NewGracefulShutdown(cfg.Shutdown, zapLogger)

	go gracefulShutdown(ctx, func(hookCtx context.Context) error {
		zapLogger.Info("closing database connection")
		database.Close()
		return nil
	})

	zapLogger.Info("application is running, press Ctrl+C to stop")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	<-sigCh
	zapLogger.Info("received shutdown signal, starting graceful shutdown")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Shutdown.Timeout)
	defer shutdownCancel()

	doneCh := make(chan struct{})
	go func() {
		gracefulShutdown(shutdownCtx, func(hookCtx context.Context) error {
			zapLogger.Info("closing database connection")
			database.Close()
			return nil
		})
		close(doneCh)
	}()

	select {
	case <-doneCh:
		zapLogger.Info("shutdown completed successfully")
	case <-shutdownCtx.Done():
		if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
			zapLogger.Error("shutdown timeout exceeded")
			return fmt.Errorf("%w after %v: %w", ErrShutdownTimeout, cfg.Shutdown.Timeout, shutdownCtx.Err())
		}
	case <-sigCh:
		zapLogger.Warn("second interrupt received, forcing immediate exit")
		return ErrForcedShutdown
	}

	zapLogger.Info("application stopped")
	return nil
}
