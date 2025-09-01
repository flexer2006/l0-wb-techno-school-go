BEGIN;

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY CHECK (LENGTH(order_uid) > 0),
    track_number TEXT,
    entry TEXT,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INTEGER,
    date_created TIMESTAMPTZ,
    oof_shard TEXT,
    raw JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_orders_track_number ON orders (track_number) WHERE track_number IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders (date_created);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders (customer_id);

CREATE TABLE IF NOT EXISTS delivery (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_uid TEXT NOT NULL UNIQUE REFERENCES orders(order_uid) ON DELETE CASCADE ON UPDATE NO ACTION,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

CREATE TABLE IF NOT EXISTS payment (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_uid TEXT NOT NULL UNIQUE REFERENCES orders(order_uid) ON DELETE CASCADE ON UPDATE NO ACTION,
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount NUMERIC(14,2) CHECK (amount >= 0),
    payment_dt BIGINT,
    payment_ts TIMESTAMPTZ GENERATED ALWAYS AS (to_timestamp(payment_dt)::timestamptz) STORED,
    bank TEXT,
    delivery_cost NUMERIC(14,2) CHECK (delivery_cost >= 0),
    goods_total NUMERIC(14,2) CHECK (goods_total >= 0),
    custom_fee NUMERIC(14,2) CHECK (custom_fee >= 0)
);

CREATE TABLE IF NOT EXISTS items (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_uid TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE ON UPDATE NO ACTION,
    chrt_id BIGINT,
    track_number TEXT,
    price NUMERIC(14,2) CHECK (price >= 0),
    rid TEXT,
    name TEXT,
    sale INTEGER CHECK (sale >= 0 AND sale <= 100),
    size TEXT,
    total_price NUMERIC(14,2) CHECK (total_price >= 0),
    nm_id BIGINT,
    brand TEXT,
    status INTEGER
);

CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items (order_uid);
CREATE INDEX IF NOT EXISTS idx_items_nm_id ON items (nm_id);
CREATE INDEX IF NOT EXISTS idx_items_chrt_id ON items (chrt_id);

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;