-- БД
CREATE DATABASE IF NOT EXISTS orders;
-- Юзер + права под БД
CREATE USER orders_user WITH PASSWORD 'orders_password';
GRANT ALL PRIVILEGES ON DATABASE orders TO orders_user;
-- Таблица
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    data JSONB NOT NULL
);
-- Индекс
CREATE INDEX idx_orders_data ON orders USING GIN (data);
-- Права на таблицу
GRANT ALL PRIVILEGES ON TABLE orders TO orders_user;