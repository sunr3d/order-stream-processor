-- Таблица
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    data JSONB NOT NULL
);
-- Индекс
CREATE INDEX idx_orders_data ON orders USING GIN (data);
-- Права пользователю
GRANT ALL PRIVILEGES ON TABLE orders TO orders_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO orders_user;