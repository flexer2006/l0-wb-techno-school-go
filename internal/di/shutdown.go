package di

import (
	"context"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/shutdown"
)

func NewGracefulShutdown(cfg config.ShutdownConfig, log logger.Logger) func(context.Context, ...func(context.Context) error) {
	timeout := cfg.Timeout

	return func(ctx context.Context, hooks ...func(context.Context) error) {
		shutdown.Wait(ctx, timeout, log, hooks...)
	}
}
