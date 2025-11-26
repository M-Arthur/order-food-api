-- db/migrations/001_init.sql

-- Drop tables if they exist (dev convenience)
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;

-- Orders table
CREATE TABLE orders (
    id          VARCHAR(64) PRIMARY KEY,
    coupon_code TEXT NULL
);

-- Order items table
CREATE TABLE order_items (
    order_id   VARCHAR(64) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(64) NOT NULL,
    quantity   INT NOT NULL,

    PRIMARY KEY (order_id, product_id)
);

-- Optional: some indexes to speed up queries later
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items (product_id);

-- Optional: seed data example (remove if you don't want sample rows)
-- INSERT INTO orders (id, coupon_code) VALUES ('order-seed-1', NULL);
-- INSERT INTO order_items (order_id, product_id, quantity)
-- VALUES ('order-seed-1', '10', 2);