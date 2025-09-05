package server

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

func LoggingMiddleware(log logger.Logger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()
		err := ctx.Next()
		duration := time.Since(start)

		log.Info("HTTP request",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"status", ctx.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
			"ip", ctx.IP(),
		)
		if err != nil {
			return fmt.Errorf("middleware next: %w", err)
		}
		return nil
	}
}
