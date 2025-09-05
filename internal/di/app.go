package di

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/cache"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres/connect"
	consumer "github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/kafka"
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

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	components, err := initComponents(ctx, cfg, zapLogger)
	if err != nil {
		zapLogger.Error("failed to initialize components", "error", err)
		return fmt.Errorf("%w: %w", ErrApplicationStartup, err)
	}

	startServices(ctx, components, zapLogger)

	zapLogger.Info("application is running, press Ctrl+C to stop")

	return handleShutdown(ctx, cancel, cfg, components, zapLogger)
}

func initComponents(ctx context.Context, cfg *config.Config, log logger.Logger) (*serviceComponents, error) {
	database, err := NewDatabase(ctx, cfg.Database, log)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	log.Info("database initialized successfully")

	repo := NewRepository(database, log)
	caches := NewCache(log)

	if err := caches.RestoreFromDB(ctx, repo); err != nil {
		log.Warn("failed to restore caches from DB", "error", err)
	} else {
		log.Info("caches restored from DB")
	}

	service := NewService(repo, caches, log)
	kafkaConsumer := NewKafkaConsumer(cfg.Kafka, service, log)
	httpServer := NewHTTPServer(caches, repo, log, cfg.Server)
	gracefulShutdown := NewGracefulShutdown(cfg.Shutdown, log)

	return &serviceComponents{
		database:         database,
		repo:             repo,
		cache:            caches,
		service:          service,
		kafkaConsumer:    kafkaConsumer,
		httpServer:       httpServer,
		gracefulShutdown: gracefulShutdown,
	}, nil
}

type serviceComponents struct {
	database         *connect.DB
	repo             ports.OrderRepository
	cache            ports.Cache
	service          *order.OrderService
	kafkaConsumer    ports.KafkaConsumer
	httpServer       ports.HTTPServer
	gracefulShutdown func(context.Context, ...func(context.Context) error)
}

func startServices(ctx context.Context, comp *serviceComponents, log logger.Logger) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	waitGroup.Go(func() {
		if err := comp.kafkaConsumer.Start(ctx); err != nil {
			log.Error("failed to start Kafka consumer", "error", err)
		}
	})

	waitGroup.Go(func() {
		if err := comp.httpServer.Start(ctx); err != nil {
			log.Error("failed to start HTTP server", "error", err)
		}
	})

	waitGroup.Wait()
}

func handleShutdown(ctx context.Context, cancel context.CancelCauseFunc, cfg *config.Config, comp *serviceComponents, log logger.Logger) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	<-sigCh
	log.Info("received shutdown signal, starting graceful shutdown")

	cancel(ErrForcedShutdown)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.Shutdown.Timeout)
	defer shutdownCancel()

	comp.gracefulShutdown(shutdownCtx,
		func(hookCtx context.Context) error {
			log.Info("closing database connection")
			comp.database.Close()
			return nil
		},
		func(hookCtx context.Context) error {
			log.Info("stopping HTTP server")
			return comp.httpServer.Stop(hookCtx)
		},
		func(hookCtx context.Context) error {
			log.Info("stopping Kafka consumer")
			return comp.kafkaConsumer.Stop(hookCtx)
		},
	)

	log.Info("shutdown completed successfully")
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
	return consumer.NewKafkaConsumerWithConfig(cfg, service, log)
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
