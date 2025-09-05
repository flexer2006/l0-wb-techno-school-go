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

var (
	ErrInvalidJSON = errors.New("invalid JSON payload")
	ErrSaveFailed  = errors.New("failed to save order after retries")
)

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
	if len(payload) == 0 {
		s.log.Warn("received empty payload")
		return nil
	}

	order := domain.Order{}
	if err := json.Unmarshal(payload, &order); err != nil {
		s.log.Warn("failed to unmarshal JSON payload",
			"error", err,
			"payload_size", len(payload),
			"payload_preview", s.getPayloadPreview(payload))
		return nil
	}

	s.log.Info("unmarshaled order", "order_uid", order.OrderUID, "sm_id", order.SmID)

	if order.OrderUID == "" {
		s.log.Warn("unmarshaled order has empty order_uid, skipping")
		return nil
	}

	if err := ValidateOrder(&order, s.log); err != nil {
		s.log.Warn("order validation failed, skipping",
			"order_uid", order.OrderUID,
			"error", err)
		return nil
	}

	if err := s.saveOrderWithRetry(ctx, &order); err != nil {
		s.log.Error("failed to save order after retry",
			"order_uid", order.OrderUID,
			"error", err)
		return fmt.Errorf("save order to DB: %w", err)
	}

	orderFromDB, err := s.repo.GetOrder(ctx, order.OrderUID)
	if err != nil {
		s.log.Warn("failed to get order from DB after save, caching original",
			"order_uid", order.OrderUID,
			"error", err)
		s.cache.Set(&order)
	} else {
		s.cache.Set(orderFromDB)
	}

	s.log.Info("order processed successfully",
		"order_uid", order.OrderUID,
		"items_count", len(order.Items))

	return nil
}

func (s *OrderService) getPayloadPreview(payload []byte) string {
	const maxPreviewLen = 100
	if len(payload) <= maxPreviewLen {
		return string(payload)
	}
	return string(payload[:maxPreviewLen]) + "..."
}

func (s *OrderService) saveOrderWithRetry(ctx context.Context, order *domain.Order) error {
	const maxRetries = 2

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := s.repo.SaveOrderTx(ctx, order); err == nil {
			return nil
		}

		if attempt < maxRetries {
			s.log.Warn("save order attempt failed, retrying",
				"order_uid", order.OrderUID,
				"attempt", attempt+1)
		}
	}

	return fmt.Errorf("failed to save: %w", ErrSaveFailed)
}
