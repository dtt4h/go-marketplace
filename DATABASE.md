# База данных

## Обзор

PostgreSQL 16. Доступ через pgx. Миграции — golang-migrate (SQL-файлы в `migrations/`).

## ER-диаграмма

```
┌──────────────┐       ┌──────────────┐
│    users     │       │    stores    │
├──────────────┤       ├──────────────┤
│ id (PK)      │◄──────│ id (PK)      │
│ email        │  1:1  │ user_id (FK) │
│ password_hash│       │ name         │
│ role         │       │ description  │
│ first_name   │       │ logo_url     │
│ last_name    │       │ created_at   │
│ avatar_url   │       │ updated_at   │
│ phone        │       └──────┬───────┘
│ created_at   │              │
│ updated_at   │              │ 1:N
└──────┬───────┘              │
       │                      ▼
       │ 1:N          ┌──────────────┐
       │              │   products   │
       │              ├──────────────┤
       │              │ id (PK)      │
       │              │ store_id (FK)│
       │              │ category_id  │
       │   ┌──────────│ title        │
       │   │          │ description  │
       │   │          │ price        │
       │   │          │ stock        │
       │   │          │ status       │
       │   │          │ created_at   │
       │   │          │ updated_at   │
       │   │          └──────┬───────┘
       │   │                 │
       │   │    ┌────────────┘
       │   │    │ 1:N
       │   │    ▼
       │   │  ┌──────────────────┐
       │   │  │ product_images   │
       │   │  ├──────────────────┤
       │   │  │ id (PK)          │
       │   │  │ product_id (FK)  │
       │   │  │ url              │
       │   │  │ position         │
       │   │  └──────────────────┘
       │   │
       │   │          ┌──────────────────┐
       │   └──────────│   categories     │
       │              ├──────────────────┤
       │              │ id (PK)          │
       │              │ parent_id (FK)   │
       │              │ name             │
       │              │ slug             │
       │              └──────────────────┘
       │
       │ 1:N
       ▼
┌──────────────┐       ┌──────────────────┐       ┌──────────────────┐
│    orders    │──────►│   order_items    │       │    payments      │
├──────────────┤  1:N  ├──────────────────┤       ├──────────────────┤
│ id (PK)      │       │ id (PK)          │       │ id (PK)          │
│ user_id (FK) │       │ order_id (FK)    │       │ order_id (FK)    │
│ status       │       │ product_id (FK)  │       │ amount           │
│ total        │       │ quantity         │       │ currency         │
│ address      │       │ price            │       │ status           │
│ created_at   │       │ created_at       │       │ provider         │
│ updated_at   │       └──────────────────┘       │ provider_payment │
└──────────────┘                                  │ created_at       │
                                                  │ updated_at       │
                                                  └──────────────────┘
```

## Таблицы

### users

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | Идентификатор |
| `email` | `VARCHAR(255)` | UNIQUE, NOT NULL | Email |
| `password_hash` | `VARCHAR(255)` | NOT NULL | Хеш пароля (bcrypt) |
| `role` | `user_role` | NOT NULL, DEFAULT `'buyer'` | Роль: buyer/seller/admin |
| `first_name` | `VARCHAR(100)` | | Имя |
| `last_name` | `VARCHAR(100)` | | Фамилия |
| `avatar_url` | `TEXT` | | URL аватара |
| `phone` | `VARCHAR(20)` | | Телефон |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | Дата создания |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | Дата обновления |

```sql
CREATE TYPE user_role AS ENUM ('buyer', 'seller', 'admin');

CREATE TABLE users (
    id            BIGSERIAL    PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          user_role    NOT NULL DEFAULT 'buyer',
    first_name    VARCHAR(100),
    last_name     VARCHAR(100),
    avatar_url    TEXT,
    phone         VARCHAR(20),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

### stores

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | Идентификатор |
| `user_id` | `BIGINT` | FK → users(id), UNIQUE, NOT NULL | Владелец-продавец |
| `name` | `VARCHAR(200)` | NOT NULL | Название магазина |
| `description` | `TEXT` | | Описание |
| `logo_url` | `TEXT` | | Логотип |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TABLE stores (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    logo_url    TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

### categories

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `parent_id` | `BIGINT` | FK → categories(id), NULL | Родительская категория |
| `name` | `VARCHAR(100)` | NOT NULL | Название |
| `slug` | `VARCHAR(100)` | UNIQUE, NOT NULL | URL-слаг |

```sql
CREATE TABLE categories (
    id        BIGSERIAL    PRIMARY KEY,
    parent_id BIGINT       REFERENCES categories(id) ON DELETE SET NULL,
    name      VARCHAR(100) NOT NULL,
    slug      VARCHAR(100) NOT NULL UNIQUE
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
```

### products

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `store_id` | `BIGINT` | FK → stores(id), NOT NULL | Магазин |
| `category_id` | `BIGINT` | FK → categories(id) | Категория |
| `title` | `VARCHAR(300)` | NOT NULL | Название |
| `description` | `TEXT` | | Описание |
| `price` | `NUMERIC(12,2)` | NOT NULL, CHECK > 0 | Цена |
| `stock` | `INTEGER` | NOT NULL, DEFAULT 0, CHECK >= 0 | Остаток |
| `status` | `product_status` | NOT NULL, DEFAULT `'pending'` | Статус модерации |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TYPE product_status AS ENUM ('pending', 'active', 'rejected', 'archived');

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
CREATE INDEX idx_products_title_trgm  ON products USING gin (title gin_trgm_ops);
```

> `gin_trgm_ops` требует расширения `pg_trgm` для полнотекстового поиска.

### product_images

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `product_id` | `BIGINT` | FK → products(id), NOT NULL | Товар |
| `url` | `TEXT` | NOT NULL | URL изображения |
| `position` | `INTEGER` | NOT NULL, DEFAULT 0 | Порядок отображения |

```sql
CREATE TABLE product_images (
    id         BIGSERIAL  PRIMARY KEY,
    product_id BIGINT     NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url        TEXT       NOT NULL,
    position   INTEGER    NOT NULL DEFAULT 0
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);
```

### orders

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `user_id` | `BIGINT` | FK → users(id), NOT NULL | Покупатель |
| `status` | `order_status` | NOT NULL, DEFAULT `'pending'` | Статус |
| `total` | `NUMERIC(12,2)` | NOT NULL | Сумма заказа |
| `address` | `TEXT` | NOT NULL | Адрес доставки |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'delivered', 'cancelled');

CREATE TABLE orders (
    id         BIGSERIAL      PRIMARY KEY,
    user_id    BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     order_status   NOT NULL DEFAULT 'pending',
    total      NUMERIC(12, 2) NOT NULL,
    address    TEXT           NOT NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status  ON orders(status);
```

### order_items

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `order_id` | `BIGINT` | FK → orders(id), NOT NULL | Заказ |
| `product_id` | `BIGINT` | FK → products(id), NOT NULL | Товар |
| `quantity` | `INTEGER` | NOT NULL, CHECK > 0 | Количество |
| `price` | `NUMERIC(12,2)` | NOT NULL | Цена на момент заказа |

```sql
CREATE TABLE order_items (
    id         BIGSERIAL      PRIMARY KEY,
    order_id   BIGINT         NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT         NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INTEGER        NOT NULL CHECK (quantity > 0),
    price      NUMERIC(12, 2) NOT NULL
);

CREATE INDEX idx_order_items_order_id   ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
```

### payments

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `order_id` | `BIGINT` | FK → orders(id), UNIQUE, NOT NULL | Заказ |
| `amount` | `NUMERIC(12,2)` | NOT NULL | Сумма |
| `currency` | `VARCHAR(3)` | NOT NULL, DEFAULT `'RUB'` | Валюта |
| `status` | `payment_status` | NOT NULL, DEFAULT `'pending'` | Статус |
| `provider` | `VARCHAR(50)` | NOT NULL | Платёжный провайдер |
| `provider_payment_id` | `VARCHAR(255)` | | ID платежа у провайдера |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TYPE payment_status AS ENUM ('pending', 'succeeded', 'failed', 'refunded');

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
```

### refresh_tokens

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `user_id` | `BIGINT` | FK → users(id), NOT NULL | Пользователь |
| `token` | `VARCHAR(512)` | UNIQUE, NOT NULL | Хеш refresh token |
| `expires_at` | `TIMESTAMPTZ` | NOT NULL | Срок действия |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TABLE refresh_tokens (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(512) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token   ON refresh_tokens(token);
```

## Индексы

Помимо указанных выше, рекомендуется:

| Индекс | Таблица | Назначение |
|--------|---------|------------|
| `idx_products_price` | products | Фильтрация по цене |
| `idx_products_created_at` | products | Сортировка по новизне |
| `idx_orders_created_at` | orders | Сортировка по дате |

## Расширения

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;  -- Нечёткий поиск по названию товара
```

## Триггеры

Триггер на автоматическое обновление `updated_at`:

```sql
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Применить ко всем таблицам с updated_at:
CREATE TRIGGER trg_users_updated_at    BEFORE UPDATE ON users    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_stores_updated_at   BEFORE UPDATE ON stores   FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_orders_updated_at   BEFORE UPDATE ON orders   FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```
