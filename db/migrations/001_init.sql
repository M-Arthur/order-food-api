-- db/migrations/001_init.sql

-- Drop tables if they exist (dev convenience)
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;

-- Orders table
CREATE TABLE orders (
    id          VARCHAR(64) PRIMARY KEY,
    coupon_code TEXT NULL
);

-- Products table
CREATE TABLE products (
    id          VARCHAR(64) PRIMARY KEY,
    name        TEXT NOT NULL,
    price_cents BIGINT NOT NULL,
    category    TEXT NOT NULL
);

-- Order items table
CREATE TABLE order_items (
    order_id   VARCHAR(64) NOT NULL,
    product_id VARCHAR(64) NOT NULL,
    quantity   INT NOT NULL,

    PRIMARY KEY (order_id, product_id),

    -- Do NOT allow deleting an order if items exist
    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders (id)
        ON DELETE RESTRICT,

    -- Do NOT allow deleting a product if items exist
    CONSTRAINT fk_order_items_product
        FOREIGN KEY (product_id)
        REFERENCES products (id)
        ON DELETE RESTRICT
);

-- Optional: some indexes to speed up queries later
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items (product_id);

INSERT INTO products (id, name, price_cents, category) VALUES
('1', 'Waffle with Berries', 650, 'Waffle'),
('2', 'Vanilla Bean Crème Brûlée', 700, 'Crème Brûlée'),
('3', 'Macaron Mix of Five', 800, 'Macaron'),
('4', 'Classic Tiramisu', 550, 'Tiramisu'),
('5', 'Pistachio Baklava', 400, 'Baklava'),
('6', 'Lemon Meringue Pie', 500, 'Pie'),
('7', 'Red Velvet Cake', 450, 'Cake'),
('8', 'Salted Caramel Brownie', 450, 'Brownie'),
('9', 'Vanilla Panna Cotta', 650, 'Panna Cotta');
-- Optional: seed data example (remove if you don't want sample rows)
-- INSERT INTO orders (id, coupon_code) VALUES ('order-seed-1', NULL);
-- INSERT INTO order_items (order_id, product_id, quantity)
-- VALUES ('order-seed-1', '10', 2);