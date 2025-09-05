package order

import (
	"errors"
	"fmt"
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

func validateDelivery(delivery *domain.Delivery, log logger.Logger) error {
	if delivery == nil {
		log.Warn("delivery is nil")
		return ErrInvalidDelivery
	}
	if delivery.Name == "" || delivery.Phone == "" || delivery.Zip == "" ||
		delivery.City == "" || delivery.Address == "" || delivery.Region == "" ||
		delivery.Email == "" {
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
	if payment.Transaction == "" || payment.Currency == "" || payment.Provider == "" ||
		payment.Bank == "" || payment.Amount < 0 || payment.PaymentDt <= 0 ||
		payment.DeliveryCost < 0 || payment.GoodsTotal < 0 || payment.CustomFee < 0 {
		log.Warn("invalid payment fields or values")
		return ErrInvalidPayment
	}
	return nil
}

func validateItems(orderUID string, items []domain.Item, log logger.Logger) error {
	for i, item := range items {
		if item.ChrtID <= 0 || item.Name == "" || item.Price < 0 || item.TotalPrice < 0 ||
			item.Sale < 0 || item.Sale > 100 || item.Status < 0 {
			log.Warn("invalid item", "order_uid", orderUID, "index", i, "chrt_id", item.ChrtID, "name", item.Name)
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
	if order.DateCreated.IsZero() || order.DateCreated.After(time.Now()) {
		log.Warn("invalid order: bad date_created", "order_uid", order.OrderUID, "date", order.DateCreated)
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

	log.Debug("order validated successfully", "order_uid", order.OrderUID)
	return nil
}
