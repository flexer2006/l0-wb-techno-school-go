package ports

import (
	"context"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/domain"
	"github.com/gofiber/fiber/v3"
)

type Cache interface {
	Get(orderUID string) (*domain.Order, bool)
	Set(order *domain.Order)
	RestoreFromDB(ctx context.Context, repo OrderRepository) error
}

type HTTPServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	RegisterRoutes(orderHandler fiber.Handler)
}

type KafkaConsumer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type OrderService interface {
	ProcessMessage(ctx context.Context, payload []byte) error
}

type OrderRepository interface {
	SaveOrderTx(ctx context.Context, order *domain.Order) error
	GetOrder(ctx context.Context, orderUID string) (*domain.Order, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.Order, error)
}
