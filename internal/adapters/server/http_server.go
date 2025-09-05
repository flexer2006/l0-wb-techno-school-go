package server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
)

type httpServer struct {
	app *fiber.App
	log logger.Logger
	cfg config.ServerConfig
}

func NewHTTPServer(log logger.Logger, cfg config.ServerConfig) ports.HTTPServer {
	app := fiber.New()

	app.Use(LoggingMiddleware(log))

	app.Use("/static", static.New("./static"))

	return &httpServer{
		app: app,
		log: log,
		cfg: cfg,
	}
}

func (s *httpServer) Start(ctx context.Context) error {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
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
