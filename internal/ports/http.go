package ports

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type HTTPServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	RegisterRoutes(orderHandler fiber.Handler)
}
