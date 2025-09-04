package di

import (
	"github.com/flexer2006/l0-wb-techno-school-go/internal/app/order"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

func NewService(repo ports.OrderRepository, cache ports.Cache, log logger.Logger) *order.OrderService {
	return order.NewOrderService(repo, cache, log)
}
