package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/domain"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

var (
	ErrRepoNil         = errors.New("repo cannot be nil")
	ErrContextCanceled = errors.New("context canceled")
)

type inMemoryCache struct {
	cache sync.Map
	log   logger.Logger
}

func NewInMemoryCache(log logger.Logger) ports.Cache {
	return &inMemoryCache{
		cache: sync.Map{},
		log:   log,
	}
}

func (c *inMemoryCache) Get(orderUID string) (*domain.Order, bool) {
	if orderUID == "" {
		c.log.Warn("attempt to get order with empty orderUID")
		return nil, false
	}

	val, found := c.cache.Load(orderUID)
	if !found {
		c.log.Debug("order not found in cache", "order_uid", orderUID)
		return nil, false
	}

	order, isOrder := val.(*domain.Order)
	if !isOrder {
		c.log.Error("invalid type in cache", "order_uid", orderUID)
		return nil, false
	}

	c.log.Debug("order retrieved from cache", "order_uid", orderUID)
	return order, true
}

func (c *inMemoryCache) Set(order *domain.Order) {
	if order == nil || order.OrderUID == "" || order.DateCreated.IsZero() {
		c.log.Warn("attempt to save invalid order (nil, empty orderUID or date_created)")
		return
	}

	c.cache.Store(order.OrderUID, order)
	c.log.Debug("order saved in cache", "order_uid", order.OrderUID)
}

func (c *inMemoryCache) RestoreFromDB(ctx context.Context, repo ports.OrderRepository) error {
	if repo == nil {
		return fmt.Errorf("repo nil: %w", ErrRepoNil)
	}

	const limit = 1000

	c.cache.Range(func(key, value any) bool {
		c.cache.Delete(key)
		return true
	})
	c.log.Debug("cache cleared before restore")

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled before query: %w", ErrContextCanceled)
	default:
	}

	orders, err := repo.ListRecent(ctx, limit)
	if err != nil {
		c.log.Error("failed to restore cache from DB", "error", err, "limit", limit)
		return fmt.Errorf("restore cache from DB: %w", err)
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			c.log.Warn("context canceled during cache restore", "processed_orders", len(orders))
			return fmt.Errorf("context canceled during restore: %w", ErrContextCanceled)
		default:
		}

		if order != nil && order.OrderUID != "" {
			c.Set(order)
		}
	}

	c.log.Info("cache restored from DB", "orders_count", len(orders), "limit", limit, "cleared", true)
	return nil
}
