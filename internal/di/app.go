package di

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/cache"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres/connect"
	kafkalocal "github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/kafka"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/server"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/server/handlers"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/app/order"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/shutdown"
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

	repo := NewRepository(database, zapLogger)

	cache := NewCache(zapLogger)

	if err := cache.RestoreFromDB(ctx, repo); err != nil {
		zapLogger.Warn("failed to restore cache from DB", "error", err)
	} else {
		zapLogger.Info("cache restored from DB")
	}

	service := NewService(repo, cache, zapLogger)

	kafkaConsumer := NewKafkaConsumer(cfg.Kafka, service, zapLogger)

	httpServer := NewHTTPServer(cache, repo, zapLogger, cfg.Server)

	gracefulShutdown := NewGracefulShutdown(cfg.Shutdown, zapLogger)

	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			zapLogger.Error("failed to start Kafka consumer", "error", err)
		}
	}()

	go func() {
		if err := httpServer.Start(ctx); err != nil {
			zapLogger.Error("failed to start HTTP server", "error", err)
		}
	}()

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
		gracefulShutdown(shutdownCtx,
			func(hookCtx context.Context) error {
				zapLogger.Info("closing database connection")
				database.Close()
				return nil
			},
			func(hookCtx context.Context) error {
				zapLogger.Info("stopping HTTP server")
				return httpServer.Stop(hookCtx)
			},
			func(hookCtx context.Context) error {
				zapLogger.Info("stopping Kafka consumer")
				return kafkaConsumer.Stop(hookCtx)
			},
		)
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

func NewZapLogger(cfg config.LoggerConfig) logger.Logger {
	return logger.NewZapLoggerFromConfig(cfg)
}
func NewDatabase(ctx context.Context, cfg config.DatabaseConfig, log logger.Logger) (*connect.DB, error) {
	db, err := connect.NewPoolWithMigrations(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("new database: %w", err)
	}
	return db, nil
}

func NewRepository(db *connect.DB, log logger.Logger) ports.OrderRepository {
	return postgres.NewOrderRepository(db, log)
}

func NewCache(log logger.Logger) ports.Cache {
	return cache.NewInMemoryCache(log)
}

func NewService(repo ports.OrderRepository, cache ports.Cache, log logger.Logger) *order.OrderService {
	return order.NewOrderService(repo, cache, log)
}

func NewKafkaConsumer(cfg config.KafkaConfig, service ports.OrderService, log logger.Logger) ports.KafkaConsumer {
	return kafkalocal.NewKafkaConsumerWithConfig(cfg, service, log)
}

func NewHTTPServer(cache ports.Cache, repo ports.OrderRepository, log logger.Logger, cfg config.ServerConfig) ports.HTTPServer {
	httpSrv := server.NewHTTPServer(log, cfg)
	orderHandler := handlers.OrderHandler(cache, repo, log)
	httpSrv.RegisterRoutes(orderHandler)
	return httpSrv
}

func NewGracefulShutdown(cfg config.ShutdownConfig, log logger.Logger) func(context.Context, ...func(context.Context) error) {
	timeout := cfg.Timeout
	return func(ctx context.Context, hooks ...func(context.Context) error) {
		shutdown.Wait(ctx, timeout, log, hooks...)
	}
}
