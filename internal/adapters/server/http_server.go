package server

import (
	"context"
	"fmt"
	"net"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
)

const (
	ReadBufferSize  = 8192
	WriteBufferSize = 8192
	Concurrency     = 512 * 1024
)

type httpServer struct {
	app *fiber.App
	log logger.Logger
	cfg config.ServerConfig
}

func NewHTTPServer(log logger.Logger, cfg config.ServerConfig) ports.HTTPServer {
	fiberCfg := fiber.Config{
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.Timeout,
		IdleTimeout:     cfg.IdleTimeout,
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		Concurrency:     Concurrency,
	}

	app := fiber.New(fiberCfg)

	app.Use(LoggingMiddleware(log))

	staticCfg := static.Config{
		CacheDuration: 60 * cfg.Timeout,
	}
	app.Use("/static", static.New("./static", staticCfg))

	return &httpServer{
		app: app,
		log: log,
		cfg: cfg,
	}
}

func (s *httpServer) Start(ctx context.Context) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))
	s.log.Info("starting HTTP server", "addr", addr)

	go func() {
		if err := s.app.Listen(addr); err != nil {
			s.log.Error("HTTP server failed", "error", err)
		}
	}()

	<-ctx.Done()
	s.log.Info("shutting down HTTP server")

	return fmt.Errorf("shutdown: %w", s.app.Shutdown())
}

func (s *httpServer) Stop(ctx context.Context) error {
	s.log.Info("stopping HTTP server")
	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()

	return fmt.Errorf("shutdown with context: %w", s.app.ShutdownWithContext(shutdownCtx))
}

func (s *httpServer) RegisterRoutes(orderHandler fiber.Handler) {
	s.app.Get("/order/:id", orderHandler)
	s.app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
