package di

import (
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres/connect"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

func NewRepository(db *connect.DB, log logger.Logger) ports.OrderRepository {
	return postgres.NewOrderRepository(db, log)
}
