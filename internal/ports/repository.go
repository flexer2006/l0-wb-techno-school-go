package ports

import (
	"context"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/domain"
)

type OrderRepository interface {
	SaveOrderTx(ctx context.Context, order *domain.Order) error
	GetOrder(ctx context.Context, orderUID string) (*domain.Order, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.Order, error)
}
