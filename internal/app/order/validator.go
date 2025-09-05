package order

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/domain"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
)

var (
	ErrInvalidOrderUID = errors.New("order_uid cannot be empty")
	ErrInvalidDate     = errors.New("date_created cannot be zero")
	ErrInvalidDelivery = errors.New("delivery fields cannot be empty")
	ErrInvalidPayment  = errors.New("payment fields invalid")
	ErrInvalidItem     = errors.New("item fields invalid")
	ErrOrderNil        = errors.New("order is nil")
)

const (
	maxSalePercent   = 100
	minPositiveValue = 0
	timeTolerance    = 1 * time.Minute
)

func validateDelivery(delivery *domain.Delivery, log logger.Logger) error {
	if delivery == nil {
		log.Warn("delivery is nil")
		return ErrInvalidDelivery
	}

	requiredFields := []string{
		delivery.Name, delivery.Phone, delivery.Zip,
		delivery.City, delivery.Address, delivery.Region, delivery.Email,
	}

	if slices.Contains(requiredFields, "") {
		log.Warn("incomplete delivery fields")
		return ErrInvalidDelivery
	}
	return nil
}

func validatePayment(payment *domain.Payment, log logger.Logger) error {
	if payment == nil {
		log.Warn("payment is nil")
		return ErrInvalidPayment
	}

	if payment.Transaction == "" || payment.Currency == "" ||
		payment.Provider == "" || payment.Bank == "" {
		log.Warn("invalid payment string fields")
		return ErrInvalidPayment
	}

	if payment.PaymentDt <= 0 {
		log.Warn("invalid payment timestamp")
		return ErrInvalidPayment
	}

	amounts := []float64{
		payment.Amount, payment.DeliveryCost,
		payment.GoodsTotal, payment.CustomFee,
	}
	for _, amount := range amounts {
		if amount < minPositiveValue {
			log.Warn("invalid payment amount")
			return ErrInvalidPayment
		}
	}
	return nil
}

func validateItems(orderUID string, items []domain.Item, log logger.Logger) error {
	for itemIndex, item := range items {
		if item.ChrtID <= 0 {
			log.Warn("invalid item chrt_id", "order_uid", orderUID, "index", itemIndex, "chrt_id", item.ChrtID)
			return ErrInvalidItem
		}

		if item.Name == "" {
			log.Warn("invalid item name", "order_uid", orderUID, "index", itemIndex)
			return ErrInvalidItem
		}

		if item.Price < minPositiveValue || item.TotalPrice < minPositiveValue {
			log.Warn("invalid item price", "order_uid", orderUID, "index", itemIndex, "price", item.Price, "total_price", item.TotalPrice)
			return ErrInvalidItem
		}

		if item.Sale < minPositiveValue || item.Sale > maxSalePercent || item.Status < minPositiveValue {
			log.Warn("invalid item sale or status", "order_uid", orderUID, "index", itemIndex, "sale", item.Sale, "status", item.Status)
			return ErrInvalidItem
		}
	}
	return nil
}

func ValidateOrder(order *domain.Order, log logger.Logger) error {
	if order == nil {
		log.Warn("attempt to validate nil order")
		return fmt.Errorf("order nil: %w", ErrOrderNil)
	}

	if order.OrderUID == "" {
		log.Warn("invalid order: empty order_uid")
		return ErrInvalidOrderUID
	}

	now := time.Now().UTC()
	orderTime := order.DateCreated.UTC()
	futureLimit := now.Add(timeTolerance)

	if orderTime.IsZero() {
		log.Warn("invalid order: zero date_created", "order_uid", order.OrderUID)
		return ErrInvalidDate
	}

	if orderTime.After(futureLimit) {
		log.Warn("invalid order: date_created too far in future",
			"order_uid", order.OrderUID,
			"date", order.DateCreated,
			"server_time", now,
			"tolerance_minutes", timeTolerance.Minutes())
		return ErrInvalidDate
	}

	if err := validateDelivery(order.Delivery, log); err != nil {
		return err
	}
	if err := validatePayment(order.Payment, log); err != nil {
		return err
	}
	if err := validateItems(order.OrderUID, order.Items, log); err != nil {
		return err
	}

	log.Debug("order validated successfully", "order_uid", order.OrderUID, "order_time", orderTime)
	return nil
}
