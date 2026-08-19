# База данных

## Обзор

PostgreSQL 16 (образ `postgres:16-alpine`). Доступ через pgx (пул соединений `pgxpool`). Миграции — golang-migrate (SQL-файлы в `migrations/`). SQL-запросы генерируются через sqlc (v2).

## Стек

| Компонент | Технология |
|-----------|------------|
| База данных | PostgreSQL 16 |
| Драйвер | pgx/v5 (через pgxpool) |
| Миграции | golang-migrate (file → postgres) |
| Генерация запросов | sqlc v2 (sql_package: pgx/v5) |
| Расширения | pg_trgm (нечёткий поиск по названию) |

## Миграции

Миграции — обычные SQL-файлы в `migrations/`. Выполняются при старте приложения (`main.go → runMigrations`).

### Правила работы с миграциями

- **Одна миграция = один файл `.up.sql` + `.down.sql`**
- Именование: `{version}_{description}.{up|down}.sql` (например, `000001_init.up.sql`)
- **Не редактировать** опубликованные миграции. Если нужно изменить схему — создавай новую миграцию.
- Если проект на стадии разработки и БД чистая — можно откатить и пересоздать: `docker compose down -v && docker compose up -d --build`

### Текущие миграции

**000001_init.up.sql** — создаёт:
- Расширение `pg_trgm`
- Типы: `user_role`, `product_status`, `order_status`, `payment_status`
- Функция `update_updated_at()` — триггер для автообновления `updated_at`
- Таблицы: `users`, `stores`, `categories`, `products`, `product_images`, `orders`, `order_items`, `payments`, `refresh_tokens`
- Колонки сброса пароля: `users.reset_token`, `users.reset_token_expires_at`
- Индексы для всех внешних ключей и частых фильтров
- GIN-индекс на `products.title` (триграммный поиск)

**000002_cart_items.up.sql** — таблица `cart_items` (корзина авторизованных пользователей, UNIQUE(user_id, product_id)).

**000003_guest_orders_seller_apps.up.sql**:
- `orders.user_id` становится nullable (гостевые заказы), FK → `ON DELETE SET NULL`
- Гостевые поля заказа: `public_token`, `customer_first_name`, `customer_last_name`, `customer_email`, `customer_phone`, `delivery_method`, `delivery_cost`, `tracking_number`
- Таблица `seller_applications` (заявки на статус продавца)

**000004_product_rejection_reason.up.sql** — колонка `products.rejection_reason` (причина отклонения при модерации).

**000005_order_expires_at.up.sql** — колонка `orders.expires_at` (NOT NULL, для автоматической отмены неоплаченных заказов через 24ч).

## ER-диаграмма

```
┌──────────────────┐       ┌──────────────┐
│      users       │       │    stores    │
├──────────────────┤       ├──────────────┤
│ id (PK)          │◄──────│ id (PK)      │
│ email            │  1:1  │ user_id (FK) │
│ password_hash    │       │ name         │
│ role             │       │ description  │
│ username         │       │ logo_url     │
│ avatar_url       │       │ created_at   │
│ phone            │       │ updated_at   │
│ reset_token      │       └──────┬───────┘
│ reset_token_exp  │              │
│ created_at       │              │ 1:N
│ updated_at       │              ▼
└──────┬───────────┘       ┌──────────────┐
       │                   │   products   │
       │ 1:N               ├──────────────┤
       │                   │ id (PK)      │
       │                   │ store_id (FK)│
       │           ┌───────│ category_id  │
       │           │       │ title        │
       │           │       │ description  │
       │           │       │ price        │
       │           │       │ stock        │
       │           │       │ status       │
       │           │       │ rejection_   │
       │           │       │   reason     │
       │           │       │ created_at   │
       │           │       │ updated_at   │
       │           │       └──────┬───────┘
       │           │              │
       │           │   ┌──────────┘
       │           │   │ 1:N
       │           │   ▼
       │           │ ┌──────────────────┐
       │           │ │ product_images   │
       │           │ ├──────────────────┤
       │           │ │ id (PK)          │
       │           │ │ product_id (FK)  │
       │           │ │ url              │
       │           │ │ position         │
       │           │ │ object_key       │
       │           │ └──────────────────┘
       │           │
       │           │          ┌──────────────────┐
       │           └──────────│   categories     │
       │                      ├──────────────────┤
       │                      │ id (PK)          │
       │                      │ parent_id (FK)   │
       │                      │ name             │
       │                      │ slug             │
       │                      └──────────────────┘
       │
       │ 1:N        ┌──────────────────┐       ┌──────────────────┐
       ├───────────►│   cart_items     │       │ seller_applications│
       │            ├──────────────────┤       ├──────────────────┤
       │            │ id (PK)          │       │ id (PK)          │
       │            │ user_id (FK)     │       │ user_id (FK)     │
       │            │ product_id (FK)  │       │ store_name       │
       │            │ quantity         │       │ description      │
       │            │ created_at       │       │ status           │
       │            │ updated_at       │       │ created_at       │
       │            └──────────────────┘       │ updated_at       │
       │                                        └──────────────────┘
       │ 1:N (nullable)
       ▼
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│     orders       │──────►│   order_items    │       │    payments      │
├──────────────────┤  1:N  ├──────────────────┤       ├──────────────────┤
│ id (PK)          │       │ id (PK)          │       │ id (PK)          │
│ user_id (FK,NULL)│       │ order_id (FK)    │       │ order_id (FK)    │
│ status           │       │ product_id (FK)  │       │ amount           │
│ total            │       │ quantity         │       │ currency         │
│ address          │       │ price            │       │ status           │
│ public_token     │       │ created_at       │       │ provider         │
│ customer_*       │       └──────────────────┘       │ provider_payment │
│ delivery_method  │                                  │ created_at       │
│ delivery_cost    │                                  │ updated_at       │
│ tracking_number  │                                  └──────────────────┘
│ expires_at       │
│ created_at       │
│ updated_at       │
└──────────────────┘
┌──────────────────┐
│  refresh_tokens  │
├──────────────────┤
│ id (PK)          │
│ user_id (FK)     │──→ users(id)
│ token (UNIQUE)   │
│ expires_at       │
│ created_at       │
└──────────────────┘
```

## Таблицы

### users

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | Идентификатор |
| `email` | `VARCHAR(255)` | UNIQUE, NOT NULL | Email |
| `password_hash` | `VARCHAR(255)` | NOT NULL | Хеш пароля (bcrypt, cost=12) |
| `role` | `user_role` | NOT NULL, DEFAULT `'buyer'` | Роль: buyer/seller/admin |
| `username` | `VARCHAR(100)` | UNIQUE, NOT NULL | Никнейм |
| `avatar_url` | `TEXT` | | URL аватара |
| `phone` | `VARCHAR(20)` | | Телефон |
| `reset_token` | `VARCHAR(64)` | | Токен сброса пароля |
| `reset_token_expires_at` | `TIMESTAMPTZ` | | Срок действия токена сброса |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | Дата создания |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | Дата обновления |

```sql
CREATE TYPE user_role AS ENUM ('buyer', 'seller', 'admin');

CREATE TABLE users (
    id                    BIGSERIAL    PRIMARY KEY,
    email                 VARCHAR(255) NOT NULL UNIQUE,
    password_hash         VARCHAR(255) NOT NULL,
    role                  user_role    NOT NULL DEFAULT 'buyer',
    username              VARCHAR(100) NOT NULL UNIQUE,
    avatar_url            TEXT,
    phone                 VARCHAR(20),
    reset_token           VARCHAR(64),
    reset_token_expires_at TIMESTAMPTZ,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_reset_token ON users(reset_token);
```

### refresh_tokens

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `user_id` | `BIGINT` | FK → users(id), NOT NULL | Пользователь |
| `token` | `VARCHAR(512)` | UNIQUE, NOT NULL | Refresh token (JWT) |
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

### stores

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
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
| `rejection_reason` | `TEXT` | | Причина отклонения (миграция 000004) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TYPE product_status AS ENUM ('pending', 'active', 'rejected', 'archived');

CREATE TABLE products (
    id               BIGSERIAL      PRIMARY KEY,
    store_id         BIGINT         NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category_id      BIGINT         REFERENCES categories(id) ON DELETE SET NULL,
    title            VARCHAR(300)   NOT NULL,
    description      TEXT,
    price            NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    stock            INTEGER        NOT NULL DEFAULT 0 CHECK (stock >= 0),
    status           product_status NOT NULL DEFAULT 'pending',
    rejection_reason TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_store_id    ON products(store_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_status      ON products(status);
CREATE INDEX idx_products_price       ON products(price);
CREATE INDEX idx_products_created_at  ON products(created_at);
CREATE INDEX idx_products_title_trgm  ON products USING gin (title gin_trgm_ops);
```

> `gin_trgm_ops` требует расширения `pg_trgm` для полнотекстового поиска (нечёткий поиск по части названия).

### product_images

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `product_id` | `BIGINT` | FK → products(id), NOT NULL | Товар |
| `url` | `TEXT` | NOT NULL | URL/ключ изображения |
| `position` | `INTEGER` | NOT NULL, DEFAULT 0 | Порядок отображения |
| `object_key` | `TEXT` | | S3/MinIO object key |

```sql
CREATE TABLE product_images (
    id         BIGSERIAL  PRIMARY KEY,
    product_id BIGINT     NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url        TEXT       NOT NULL,
    position   INTEGER    NOT NULL DEFAULT 0,
    object_key TEXT
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);
```

> `url` хранит object key (путь в S3). Для получения публичного URL используется presigned URL через API.

### orders

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `user_id` | `BIGINT` | FK → users(id), **NULL** (гость) | Покупатель |
| `status` | `order_status` | NOT NULL, DEFAULT `'pending'` | Статус |
| `total` | `NUMERIC(12,2)` | NOT NULL | Сумма заказа |
| `address` | `TEXT` | NOT NULL | Адрес доставки |
| `public_token` | `VARCHAR(64)` | UNIQUE | Токен гостевого заказа (миграция 000003) |
| `customer_first_name` | `VARCHAR(100)` | | Имя гостя |
| `customer_last_name` | `VARCHAR(100)` | | Фамилия гостя |
| `customer_email` | `VARCHAR(255)` | | Email гостя |
| `customer_phone` | `VARCHAR(20)` | | Телефон гостя |
| `delivery_method` | `VARCHAR(50)` | | Способ доставки |
| `delivery_cost` | `NUMERIC(12,2)` | NOT NULL, DEFAULT 0 | Стоимость доставки (серверная, всегда 0) |
| `tracking_number` | `VARCHAR(100)` | | Трек-номер доставки |
| `expires_at` | `TIMESTAMPTZ` | NOT NULL (миграция 000005) | Срок оплаты (24ч) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'delivered', 'cancelled');

CREATE TABLE orders (
    id                  BIGSERIAL      PRIMARY KEY,
    user_id             BIGINT         REFERENCES users(id) ON DELETE SET NULL,
    status              order_status   NOT NULL DEFAULT 'pending',
    total               NUMERIC(12, 2) NOT NULL,
    address             TEXT           NOT NULL,
    public_token        VARCHAR(64)    UNIQUE,
    customer_first_name VARCHAR(100),
    customer_last_name  VARCHAR(100),
    customer_email      VARCHAR(255),
    customer_phone      VARCHAR(20),
    delivery_method     VARCHAR(50),
    delivery_cost       NUMERIC(12, 2) NOT NULL DEFAULT 0,
    tracking_number     VARCHAR(100),
    expires_at          TIMESTAMPTZ    NOT NULL,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id    ON orders(user_id);
CREATE INDEX idx_orders_status     ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
```

> `expires_at` выставляется при создании заказа (`now() + 24h`). Неоплаченные заказы, у которых `expires_at < now()`, автоматически переводятся в `cancelled` фоновым процессом `OrderExpirer`, остатки возвращаются на склад.

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

### cart_items

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `user_id` | `BIGINT` | FK → users(id), NOT NULL | Пользователь |
| `product_id` | `BIGINT` | FK → products(id), NOT NULL | Товар |
| `quantity` | `INTEGER` | NOT NULL, CHECK > 0 | Количество |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TABLE cart_items (
    id         BIGSERIAL  PRIMARY KEY,
    user_id    BIGINT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT     NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INTEGER    NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, product_id)
);

CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);
CREATE INDEX idx_cart_items_product_id ON cart_items(product_id);
```

### seller_applications

| Колонка | Тип | Ограничения | Описание |
|---------|-----|-------------|----------|
| `id` | `BIGSERIAL` | PK | |
| `user_id` | `BIGINT` | FK → users(id), NOT NULL | Заявитель |
| `store_name` | `VARCHAR(200)` | NOT NULL | Название будущего магазина |
| `description` | `TEXT` | | Описание |
| `status` | `VARCHAR(20)` | NOT NULL, DEFAULT `'pending'` | pending/approved/rejected |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `now()` | |

```sql
CREATE TABLE seller_applications (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_name  VARCHAR(200) NOT NULL,
    description TEXT,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_seller_applications_user_id ON seller_applications(user_id);
CREATE INDEX idx_seller_applications_status  ON seller_applications(status);
```

> Магазин создаётся **только** при одобрении заявки (`status = 'approved'`): роль пользователя меняется на `seller` и создаётся запись в `stores` в одной транзакции.

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

## Индексы

| Индекс | Таблица | Колонка | Назначение |
|--------|---------|---------|------------|
| `idx_categories_parent_id` | categories | parent_id | Поиск подкатегорий |
| `idx_products_store_id` | products | store_id | Товары магазина |
| `idx_products_category_id` | products | category_id | Фильтр по категории |
| `idx_products_status` | products | status | Модерация |
| `idx_products_price` | products | price | Сортировка/фильтр по цене |
| `idx_products_created_at` | products | created_at | Сортировка по новизне |
| `idx_products_title_trgm` | products | title (GIN) | Нечёткий поиск по названию |
| `idx_product_images_product_id` | product_images | product_id | Поиск изображений товара |
| `idx_orders_user_id` | orders | user_id | Заказы пользователя |
| `idx_orders_status` | orders | status | Фильтр по статусу |
| `idx_orders_created_at` | orders | created_at | Сортировка по дате |
| `idx_order_items_order_id` | order_items | order_id | Позиции заказа |
| `idx_order_items_product_id` | order_items | product_id | Поиск по товару |
| `idx_cart_items_user_id` | cart_items | user_id | Корзина пользователя |
| `idx_cart_items_product_id` | cart_items | product_id | Поиск по товару в корзине |
| `idx_seller_applications_user_id` | seller_applications | user_id | Заявки пользователя |
| `idx_seller_applications_status` | seller_applications | status | Фильтр pending-заявок |
| `idx_payments_status` | payments | status | Фильтр по статусу |
| `idx_users_reset_token` | users | reset_token | Поиск по токену сброса пароля |
| `idx_refresh_tokens_user_id` | refresh_tokens | user_id | Поиск токенов пользователя |
| `idx_refresh_tokens_token` | refresh_tokens | token | Поиск по токену |

## Расширения

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;  -- Нечёткий поиск по названию товара
```

## Триггеры

Автоматическое обновление `updated_at` при UPDATE:

```sql
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Применяется к таблицам: users, stores, products, orders, payments
```

## sqlc

SQL-запросы пишутся в `sql/queries/`. Схема берётся из `migrations/`. Для генерации используется `sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "migrations"
    queries: "sql/queries"
    gen:
      go:
        package: "database"
        out: "internal/database/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        json_tags_case_style: "camel"
```

**Важно:** `sqlc.yaml` использует `migrations/` как источник схемы — синхронизация с БД автоматическая.

### Типы nullable в sqlc

sqlc v2 с `sql_package: "pgx/v5"` генерирует для nullable полей тип `pgtype.Text` (не `*string`, не `NullText`).

**Конвертация:**
```go
// *string → pgtype.Text
pgtype.Text{String: *s, Valid: true}

// pgtype.Text → *string
if t.Valid { return &t.String }
return nil
```