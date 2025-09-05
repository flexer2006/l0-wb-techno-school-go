package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/domain"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
)

var ErrInvalidJSON = errors.New("invalid JSON payload")

type OrderService struct {
	repo  ports.OrderRepository
	cache ports.Cache
	log   logger.Logger
}

func NewOrderService(repo ports.OrderRepository, cache ports.Cache, log logger.Logger) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
		log:   log,
	}
}

func (s *OrderService) ProcessMessage(ctx context.Context, payload []byte) error {
	var order domain.Order
	if err := json.Unmarshal(payload, &order); err != nil {
		s.log.Warn("failed to unmarshal JSON", "error", err)
		return fmt.Errorf("unmarshal JSON: %w", ErrInvalidJSON)
	}

	if err := ValidateOrder(&order, s.log); err != nil {
		s.log.Warn("order validation failed, skipping", "order_uid", order.OrderUID, "error", err)
		return nil
	}

	if err := s.repo.SaveOrderTx(ctx, &order); err != nil {
		s.log.Error("failed to save order to DB", "order_uid", order.OrderUID, "error", err)
		return fmt.Errorf("save order to DB: %w", err)
	}

	s.cache.Set(&order)
	s.log.Info("order processed successfully", "order_uid", order.OrderUID)
	return nil
}
