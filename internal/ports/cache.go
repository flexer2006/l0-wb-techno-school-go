package ports

import (
	"context"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/domain"
)

type Cache interface {
	Get(orderUID string) (*domain.Order, bool)
	Set(order *domain.Order)
	RestoreFromDB(ctx context.Context, repo OrderRepository) error
}
