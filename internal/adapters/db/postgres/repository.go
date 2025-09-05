package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/db/postgres/connect"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/domain"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/logger"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEmptyOrderUID = errors.New("order_uid cannot be empty")
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrder  = errors.New("invalid order data")
)

const (
	insertOrderSQL = `
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature, 
            customer_id, delivery_service, shardkey, sm_id, date_created, 
            oof_shard, raw
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (order_uid) 
        DO UPDATE SET 
            track_number = EXCLUDED.track_number,
            entry = EXCLUDED.entry,
            locale = EXCLUDED.locale,
            internal_signature = EXCLUDED.internal_signature,
            customer_id = EXCLUDED.customer_id,
            delivery_service = EXCLUDED.delivery_service,
            shardkey = EXCLUDED.shardkey,
            sm_id = EXCLUDED.sm_id,
            date_created = EXCLUDED.date_created,
            oof_shard = EXCLUDED.oof_shard,
            raw = EXCLUDED.raw,
            updated_at = now()`

	insertDeliverySQL = `
        INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid)
        DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            zip = EXCLUDED.zip,
            city = EXCLUDED.city,
            address = EXCLUDED.address,
            region = EXCLUDED.region,
            email = EXCLUDED.email`

	insertPaymentSQL = `
        INSERT INTO payment (
            order_uid, transaction, request_id, currency, provider, amount, 
            payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid)
        DO UPDATE SET
            transaction = EXCLUDED.transaction,
            request_id = EXCLUDED.request_id,
            currency = EXCLUDED.currency,
            provider = EXCLUDED.provider,
            amount = EXCLUDED.amount,
            payment_dt = EXCLUDED.payment_dt,
            bank = EXCLUDED.bank,
            delivery_cost = EXCLUDED.delivery_cost,
            goods_total = EXCLUDED.goods_total,
            custom_fee = EXCLUDED.custom_fee`

	deleteItemsSQL = `DELETE FROM items WHERE order_uid = $1`

	selectOrderSQL = `
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, 
               delivery_service, shardkey, sm_id, date_created, oof_shard, raw, created_at, updated_at
        FROM orders 
        WHERE order_uid = $1`

	selectDeliverySQL = `
        SELECT name, phone, zip, city, address, region, email
        FROM delivery 
        WHERE order_uid = $1`

	selectPaymentSQL = `
        SELECT transaction, request_id, currency, provider, amount, payment_dt, 
               bank, delivery_cost, goods_total, custom_fee, payment_ts
        FROM payment 
        WHERE order_uid = $1`

	selectItemsSQL = `
        SELECT chrt_id, track_number, price, rid, name, sale, size, 
               total_price, nm_id, brand, status
        FROM items 
        WHERE order_uid = $1
        ORDER BY id`

	selectRecentOrderUIDsSQL = `
        SELECT order_uid
        FROM orders 
        ORDER BY created_at DESC 
        LIMIT $1`
)

type Queryable interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type orderRepository struct {
	db  *connect.DB
	log logger.Logger
}

func NewOrderRepository(db *connect.DB, log logger.Logger) ports.OrderRepository {
	return &orderRepository{
		db:  db,
		log: log,
	}
}

func validateOrder(order *domain.Order) error {
	if order.OrderUID == "" {
		return ErrEmptyOrderUID
	}
	if order.DateCreated.IsZero() {
		return fmt.Errorf("%w: missing date_created", ErrInvalidOrder)
	}
	for _, item := range order.Items {
		if item.ChrtID == 0 || item.Name == "" {
			return fmt.Errorf("%w: invalid item data (chrt_id or name)", ErrInvalidOrder)
		}
		if item.Price < 0 || item.TotalPrice < 0 {
			return fmt.Errorf("%w: invalid item data (price or total_price)", ErrInvalidOrder)
		}
	}
	return nil
}

func (r *orderRepository) SaveOrderTx(ctx context.Context, order *domain.Order) error {
	if err := validateOrder(order); err != nil {
		r.log.Warn("invalid order data, skipping", "order_uid", order.OrderUID, "error", err)
		return err
	}

	transaction, err := r.db.Pool().Begin(ctx)
	if err != nil {
		r.log.Error("failed to begin transaction", "order_uid", order.OrderUID, "error", err)
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := transaction.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			r.log.Error("failed to rollback transaction", "order_uid", order.OrderUID, "error", rollbackErr)
		}
	}()

	if err := r.saveOrderInTx(ctx, transaction, order); err != nil {
		return err
	}

	if err = transaction.Commit(ctx); err != nil {
		r.log.Error("failed to commit transaction", "order_uid", order.OrderUID, "error", err)
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.log.Info("order saved successfully", "order_uid", order.OrderUID, "items_count", len(order.Items))
	return nil
}

func (r *orderRepository) saveOrderInTx(ctx context.Context, transaction Queryable, order *domain.Order) error {
	rawData, err := json.Marshal(order)
	if err != nil {
		r.log.Error("failed to marshal order to JSON", "order_uid", order.OrderUID, "error", err)
		return fmt.Errorf("marshal order to JSON: %w", err)
	}

	_, err = transaction.Exec(ctx, insertOrderSQL,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated,
		order.OofShard, rawData,
	)
	if err != nil {
		r.log.Error("failed to upsert order", "order_uid", order.OrderUID, "error", err)
		return fmt.Errorf("upsert order: %w", err)
	}

	if order.Delivery != nil {
		_, err = transaction.Exec(ctx, insertDeliverySQL,
			order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
			order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
		)
		if err != nil {
			r.log.Error("failed to upsert delivery", "order_uid", order.OrderUID, "error", err)
			return fmt.Errorf("upsert delivery: %w", err)
		}
	}

	if order.Payment != nil {
		_, err = transaction.Exec(ctx, insertPaymentSQL,
			order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
			order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
			order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
		)
		if err != nil {
			r.log.Error("failed to upsert payment", "order_uid", order.OrderUID, "error", err)
			return fmt.Errorf("upsert payment: %w", err)
		}
	}

	_, err = transaction.Exec(ctx, deleteItemsSQL, order.OrderUID)
	if err != nil {
		r.log.Error("failed to delete old items", "order_uid", order.OrderUID, "error", err)
		return fmt.Errorf("delete old items: %w", err)
	}

	if len(order.Items) > 0 {
		if err := r.insertItemsBatch(ctx, transaction, order.OrderUID, order.Items); err != nil {
			return err
		}
	}

	return nil
}

func (r *orderRepository) insertItemsBatch(ctx context.Context, transaction Queryable, orderUID string, items []domain.Item) error {
	columns := []string{
		"order_uid", "chrt_id", "track_number", "price", "rid",
		"name", "sale", "size", "total_price", "nm_id",
		"brand", "status",
	}

	rows := make([][]any, len(items))
	for i, item := range items {
		rows[i] = []any{
			orderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID,
			item.Brand, item.Status,
		}
	}

	if _, err := transaction.CopyFrom(ctx, pgx.Identifier{"items"}, columns, pgx.CopyFromRows(rows)); err == nil {
		return nil
	} else {
		r.log.Debug("CopyFrom failed, using fallback INSERT", "order_uid", orderUID, "error", err)
	}

	batchSQL, valueArgs := r.buildBatchInsertSQL(orderUID, items)
	if _, err := transaction.Exec(ctx, batchSQL, valueArgs...); err != nil {
		r.log.Error("failed to batch insert items (fallback)", "order_uid", orderUID, "items_count", len(items), "error", err)
		return fmt.Errorf("batch insert items: %w", err)
	}

	return nil
}

func (r *orderRepository) buildBatchInsertSQL(orderUID string, items []domain.Item) (string, []any) {
	valueStrings := make([]string, 0, len(items))
	valueArgs := make([]any, 0, len(items)*12)

	for i, item := range items {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*12+1, i*12+2, i*12+3, i*12+4, i*12+5, i*12+6, i*12+7, i*12+8, i*12+9, i*12+10, i*12+11, i*12+12))

		valueArgs = append(valueArgs,
			orderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID,
			item.Brand, item.Status,
		)
	}

	batchSQL := fmt.Sprintf(`
        INSERT INTO items (
            order_uid, chrt_id, track_number, price, rid, name, sale, 
            size, total_price, nm_id, brand, status
        ) VALUES %s`, strings.Join(valueStrings, ","))

	return batchSQL, valueArgs
}

func (r *orderRepository) GetOrder(ctx context.Context, orderUID string) (*domain.Order, error) {
	if orderUID == "" {
		return nil, ErrEmptyOrderUID
	}

	var order domain.Order
	var rawData []byte

	err := r.db.Pool().QueryRow(ctx, selectOrderSQL, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated,
		&order.OofShard, &rawData, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Debug("order not found", "order_uid", orderUID)
			return nil, fmt.Errorf("%w: %s", ErrOrderNotFound, orderUID)
		}
		r.log.Error("failed to get order", "order_uid", orderUID, "error", err)
		return nil, fmt.Errorf("get order: %w", err)
	}

	order.Raw = rawData

	if delivery, err := r.getDelivery(ctx, orderUID); err != nil {
		return nil, err
	} else if delivery != nil {
		order.Delivery = delivery
	}

	if payment, err := r.getPayment(ctx, orderUID); err != nil {
		return nil, err
	} else if payment != nil {
		order.Payment = payment
	}

	items, err := r.getItems(ctx, orderUID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	r.log.Debug("order retrieved successfully", "order_uid", orderUID, "items_count", len(items))
	return &order, nil
}

func (r *orderRepository) getDelivery(ctx context.Context, orderUID string) (*domain.Delivery, error) {
	var delivery domain.Delivery
	err := r.db.Pool().QueryRow(ctx, selectDeliverySQL, orderUID).Scan(
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
		&delivery.Address, &delivery.Region, &delivery.Email,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.log.Error("failed to get delivery", "order_uid", orderUID, "error", err)
		return nil, fmt.Errorf("get delivery: %w", err)
	}

	delivery.OrderUID = orderUID
	return &delivery, nil
}

func (r *orderRepository) getPayment(ctx context.Context, orderUID string) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.Pool().QueryRow(ctx, selectPaymentSQL, orderUID).Scan(
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider,
		&payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost,
		&payment.GoodsTotal, &payment.CustomFee, &payment.PaymentTs,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.log.Error("failed to get payment", "order_uid", orderUID, "error", err)
		return nil, fmt.Errorf("get payment: %w", err)
	}

	payment.OrderUID = orderUID
	return &payment, nil
}

func (r *orderRepository) getItems(ctx context.Context, orderUID string) ([]domain.Item, error) {
	rows, err := r.db.Pool().Query(ctx, selectItemsSQL, orderUID)
	if err != nil {
		r.log.Error("failed to get items", "order_uid", orderUID, "error", err)
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		var item domain.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			r.log.Error("failed to scan item", "order_uid", orderUID, "error", err)
			return nil, fmt.Errorf("scan item: %w", err)
		}
		item.OrderUID = orderUID
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("failed to iterate items", "order_uid", orderUID, "error", err)
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (r *orderRepository) ListRecent(ctx context.Context, limit int) ([]*domain.Order, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.Pool().Query(ctx, selectRecentOrderUIDsSQL, limit)
	if err != nil {
		r.log.Error("failed to get recent order UIDs", "limit", limit, "error", err)
		return nil, fmt.Errorf("get recent order UIDs: %w", err)
	}
	defer rows.Close()

	var orderUIDs []string
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			r.log.Error("failed to scan order UID", "error", err)
			return nil, fmt.Errorf("scan order UID: %w", err)
		}
		orderUIDs = append(orderUIDs, orderUID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("failed to iterate order UIDs", "error", err)
		return nil, fmt.Errorf("iterate order UIDs: %w", err)
	}

	var orders []*domain.Order
	for _, orderUID := range orderUIDs {
		order, err := r.GetOrder(ctx, orderUID)
		if err != nil {
			r.log.Warn("failed to get order during list recent", "order_uid", orderUID, "error", err)
			continue
		}
		orders = append(orders, order)
	}

	r.log.Info("recent orders retrieved", "requested", limit, "found", len(orderUIDs), "returned", len(orders))
	return orders, nil
}
