BEGIN;

CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
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
    raw JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_orders_track_number ON orders (track_number) WHERE track_number IS NOT NULL;

CREATE TABLE IF NOT EXISTS delivery (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_uid TEXT NOT NULL UNIQUE REFERENCES orders(order_uid) ON DELETE CASCADE,
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
    order_uid TEXT NOT NULL UNIQUE REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction_id TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount NUMERIC(14,2),
    payment_dt BIGINT,
    payment_ts TIMESTAMPTZ GENERATED ALWAYS AS (to_timestamp(payment_dt)::timestamptz) STORED,
    bank TEXT,
    delivery_cost NUMERIC(14,2),
    goods_total NUMERIC(14,2),
    custom_fee NUMERIC(14,2)
);

CREATE TABLE IF NOT EXISTS items (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_uid TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number TEXT,
    price NUMERIC(14,2),
    rid TEXT,
    name TEXT,
    sale INTEGER,
    size TEXT,
    total_price NUMERIC(14,2),
    nm_id BIGINT,
    brand TEXT,
    status INTEGER
);

CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders (date_created);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders (customer_id);
CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items (order_uid);
CREATE INDEX IF NOT EXISTS idx_items_nm_id ON items (nm_id);
CREATE INDEX IF NOT EXISTS idx_items_chrt_id ON items (chrt_id);

COMMIT;