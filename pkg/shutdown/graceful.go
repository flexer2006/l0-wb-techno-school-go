package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

func Wait(ctx context.Context, timeout time.Duration, log logger.Logger, hooks ...func(context.Context) error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	log.Info("waiting for shutdown signal or context cancellation")

	select {
	case sig := <-sigCh:
		log.Info("received signal, starting graceful shutdown", "signal", sig)
	case <-ctx.Done():
		log.Info("context cancelled, starting graceful shutdown")
	}

	if ctx.Err() != nil {
		log.Warn("context already cancelled, skipping hooks")
		return
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var waitGroup sync.WaitGroup
	errCh := make(chan error, len(hooks))

	for _, hook := range hooks {
		waitGroup.Add(1)
		go func(shutdownFunc func(context.Context) error) {
			defer waitGroup.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Error("panic in shutdown hook", "panic", r)
				}
			}()
			if err := shutdownFunc(shutdownCtx); err != nil {
				errCh <- err
			}
		}(hook)
	}

	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(errCh)
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Warn("shutdown timeout exceeded, forcing exit")
	case <-done:
		log.Info("all shutdown hooks completed")
	}

	var errorCount int
	for err := range errCh {
		errorCount++
		log.Error("shutdown hook failed", "error", err.Error())
	}

	if errorCount > 0 {
		log.Error("shutdown completed with errors", "count", errorCount)
	} else {
		log.Info("shutdown completed successfully")
	}
}
