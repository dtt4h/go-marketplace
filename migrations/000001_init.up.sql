-- Расширения
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Типы
CREATE TYPE user_role AS ENUM ('buyer', 'seller', 'admin');
CREATE TYPE product_status AS ENUM ('pending', 'active', 'rejected', 'archived');
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'delivered', 'cancelled');
CREATE TYPE payment_status AS ENUM ('pending', 'succeeded', 'failed', 'refunded');

-- Триггерная функция для updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- users
CREATE TABLE users (
    id            BIGSERIAL    PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          user_role    NOT NULL DEFAULT 'buyer',
    username      VARCHAR(100) NOT NULL UNIQUE,
    avatar_url    TEXT,
    phone         VARCHAR(20),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- stores
CREATE TABLE stores (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    logo_url    TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_stores_updated_at BEFORE UPDATE ON stores FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- categories
CREATE TABLE categories (
    id        BIGSERIAL    PRIMARY KEY,
    parent_id BIGINT       REFERENCES categories(id) ON DELETE SET NULL,
    name      VARCHAR(100) NOT NULL,
    slug      VARCHAR(100) NOT NULL UNIQUE
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);

-- products
CREATE TABLE products (
    id          BIGSERIAL      PRIMARY KEY,
    store_id    BIGINT         NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id BIGINT         REFERENCES categories(id) ON DELETE SET NULL,
    title       VARCHAR(300)   NOT NULL,
    description TEXT,
    price       NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    stock       INTEGER        NOT NULL DEFAULT 0 CHECK (stock >= 0),
    status      product_status NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_store_id    ON products(store_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_status      ON products(status);
CREATE INDEX idx_products_price       ON products(price);
CREATE INDEX idx_products_created_at  ON products(created_at);
CREATE INDEX idx_products_title_trgm  ON products USING gin (title gin_trgm_ops);

CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- product_images
CREATE TABLE product_images (
    id         BIGSERIAL  PRIMARY KEY,
    product_id BIGINT     NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url        TEXT       NOT NULL,
    position   INTEGER    NOT NULL DEFAULT 0
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);

-- orders
CREATE TABLE orders (
    id         BIGSERIAL      PRIMARY KEY,
    user_id    BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     order_status   NOT NULL DEFAULT 'pending',
    total      NUMERIC(12, 2) NOT NULL,
    address    TEXT           NOT NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id    ON orders(user_id);
CREATE INDEX idx_orders_status     ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

CREATE TRIGGER trg_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- order_items
CREATE TABLE order_items (
    id         BIGSERIAL      PRIMARY KEY,
    order_id   BIGINT         NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT         NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INTEGER        NOT NULL CHECK (quantity > 0),
    price      NUMERIC(12, 2) NOT NULL
);

CREATE INDEX idx_order_items_order_id   ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- payments
CREATE TABLE payments (
    id                  BIGSERIAL      PRIMARY KEY,
    order_id            BIGINT         NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    amount              NUMERIC(12, 2) NOT NULL,
    currency            VARCHAR(3)     NOT NULL DEFAULT 'RUB',
    status              payment_status NOT NULL DEFAULT 'pending',
    provider            VARCHAR(50)    NOT NULL,
    provider_payment_id VARCHAR(255),
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_status ON payments(status);

CREATE TRIGGER trg_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- refresh_tokens
CREATE TABLE refresh_tokens (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(512) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token   ON refresh_tokens(token);