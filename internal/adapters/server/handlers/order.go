package handlers

import (
	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
	"github.com/gofiber/fiber/v3"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
)

func OrderHandler(cache ports.Cache, repo ports.OrderRepository, log logger.Logger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		orderUID := ctx.Params("id")
		if orderUID == "" {
			log.WithContext(ctx).Warn("empty order UID in request")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order UID is required"})
		}

		if order, found := cache.Get(orderUID); found {
			log.WithContext(ctx).Debug("order found in cache", "order_uid", orderUID)
			return ctx.JSON(order)
		}

		order, err := repo.GetOrder(ctx, orderUID)
		if err != nil {
			log.WithContext(ctx).Error("failed to get order from DB", "order_uid", orderUID, "error", err)
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
		}

		cache.Set(order)
		log.WithContext(ctx).Info("order retrieved from DB and cached", "order_uid", orderUID)

		return ctx.JSON(order)
	}
}
