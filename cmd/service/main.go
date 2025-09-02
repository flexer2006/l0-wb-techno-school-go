package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/di"
)

func main() {
	cfg := config.MustLoad()

	log := di.NewZapLogger(cfg.Logger)

	log.Info("starting application",
		"version", "1.0.0",
		"log_level", cfg.Logger.Level,
		"shutdown_timeout", cfg.Shutdown.Timeout,
	)

	gracefulShutdown := di.NewGracefulShutdown(cfg.Shutdown, log)

	ctx, cancel := context.WithCancel(context.Background())

	go gracefulShutdown(ctx)

	log.Info("application is running, press Ctrl+C to stop")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Info("received shutdown signal")

	cancel()

	log.Info("application stopped")
}
