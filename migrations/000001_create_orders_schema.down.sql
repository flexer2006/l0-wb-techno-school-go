DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_items_chrt_id;
DROP INDEX IF EXISTS idx_items_nm_id;
DROP INDEX IF EXISTS idx_items_order_uid;
DROP INDEX IF EXISTS idx_orders_customer_id;
DROP INDEX IF EXISTS idx_orders_date_created;
DROP INDEX IF EXISTS ux_orders_track_number;

DROP TABLE IF EXISTS items CASCADE;
DROP TABLE IF EXISTS payment CASCADE;
DROP TABLE IF EXISTS delivery CASCADE;
DROP TABLE IF EXISTS orders CASCADE;