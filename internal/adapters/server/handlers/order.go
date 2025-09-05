package handlers

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

func OrderHandler(cache ports.Cache, repo ports.OrderRepository, log logger.Logger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		orderUID := ctx.Params("id")
		if orderUID == "" {
			log.Warn("empty order UID in request")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order UID is required"})
		}

		if order, found := cache.Get(orderUID); found {
			log.Debug("order found in cache", "order_uid", orderUID)
			return ctx.JSON(order)
		}

		order, err := repo.GetOrder(context.Background(), orderUID)
		if err != nil {
			log.Error("failed to get order from DB", "order_uid", orderUID, "error", err)
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
		}

		cache.Set(order)
		log.Info("order retrieved from DB and cached", "order_uid", orderUID)

		return ctx.JSON(order)
	}
}
