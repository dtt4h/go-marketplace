# API

Base URL: `/api/v1`

Все ответы — JSON. Кодировка — UTF-8.

Swagger-документация доступна по адресу: `/swagger/index.html`

## Аутентификация

JWT Bearer token в заголовке:

```
Authorization: Bearer <access_token>
```

**Payload access token:**
```json
{
  "user_id": 1,
  "role": "buyer",
  "exp": 1784720989,
  "iat": 1784720089
}
```

- **Access token** — срок жизни 15 минут (настраивается через `JWT_ACCESS_TTL`), передаётся в JSON-ответе и заголовке `Authorization`
- **Refresh token** — срок жизни 7 дней (настраивается через `JWT_REFRESH_TTL`), передаётся в **httpOnly cookie** (`refresh_token`), не возвращается в JSON-теле ответа (`json:"-"`)
- Рефреш-токен хранится в БД (таблица `refresh_tokens`), при refresh — ротация (старый удаляется, создаётся новый)
- Cookie: `HttpOnly`, `SameSite=Strict`, `Path=/`, `MaxAge=7d`

Эндпоинты, требующие авторизации, помечены 🔒. Эндпоинты, требующие определённую роль, помечены 🔒 `[role]`.

## Формат ошибок

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required",
    "details": {}
  }
}
```

| Код | HTTP | Описание |
|-----|------|----------|
| `VALIDATION_ERROR` | 400 | Некорректные входные данные |
| `UNAUTHORIZED` | 401 | Не авторизован / неверный токен / неверный пароль |
| `FORBIDDEN` | 403 | Недостаточно прав |
| `NOT_FOUND` | 404 | Ресурс не найден |
| `CONFLICT` | 409 | Конфликт (email занят, магазин уже существует) |
| `INTERNAL_ERROR` | 500 | Внутренняя ошибка сервера |

---

## Auth

### POST `/auth/register`
Регистрация нового пользователя. Создаётся пользователь с ролью `buyer` (поле `role` в запросе игнорируется, всегда `buyer`).

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secret123",
  "username": "ivan",
  "role": "buyer"
}
```

**Правила валидации:**
- `email` — обязательное, уникальное
- `password` — минимум 8 символов
- `username` — обязательное, уникальное, до 100 символов

**Response 201:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "buyer",
    "username": "ivan"
  },
  "access_token": "eyJhbGci..."
}
```

> Refresh token устанавливается в httpOnly cookie `refresh_token`.

**Ошибки:**
- `VALIDATION_ERROR` (400) — слабый пароль, невалидный email
- `CONFLICT` (409) — email уже занят

### POST `/auth/login`
Вход. При успехе — новая пара токенов.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

**Response 200:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "buyer",
    "username": "ivan"
  },
  "access_token": "eyJhbGci..."
}
```

> Refresh token устанавливается в httpOnly cookie `refresh_token`.

**Ошибки:**
- `UNAUTHORIZED` (401) — неверный email или пароль

### POST `/auth/refresh`
Обновление токенов. Читает refresh-токен из cookie `refresh_token`. Старый refresh-токен удаляется, создаётся новая пара (ротация).

**Request:** тело пустое (токен берётся из cookie)

**Response 200:**
```json
{
  "access_token": "eyJhbGci..."
}
```

> Новый refresh token устанавливается в httpOnly cookie.

**Ошибки:**
- `UNAUTHORIZED` (401) — отсутствует cookie, невалидный или истёкший refresh-токен

### POST `/auth/logout` 🔒
Выход. Удаляет **все** refresh-токены текущего пользователя (завершает все сессии). Очищает cookie.

**Response 204:** No Content

**Ошибки:**
- `UNAUTHORIZED` (401) — отсутствует или невалидный access-токен

---

## Users

### GET `/users/me` 🔒
Текущий профиль пользователя.

**Response 200:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "role": "buyer",
  "username": "ivan",
  "avatar_url": "https://...",
  "phone": "+79991234567",
  "created_at": "2026-07-19T10:00:00Z"
}
```

> Поля `avatar_url` и `phone` присутствуют только если заданы (`omitempty`).

### PATCH `/users/me` 🔒
Обновление профиля. Только переданные поля обновляются.

**Request:**
```json
{
  "username": "ivan_updated",
  "phone": "+79997654321",
  "avatar_url": "https://..."
}
```

**Response 200:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "role": "buyer",
  "username": "ivan_updated",
  "avatar_url": "https://...",
  "phone": "+79997654321",
  "created_at": "2026-07-19T10:00:00Z"
}
```

### GET `/users/me/store` 🔒
Текущий магазин пользователя (для продавца).

**Response 200:**
```json
{
  "id": 1,
  "user_id": 1,
  "name": "Мой магазин",
  "description": "Описание магазина",
  "logo_url": "https://...",
  "created_at": "2026-07-19T10:00:00Z"
}
```

**Ошибки:**
- `NOT_FOUND` (404) — магазин не найден (пользователь не продавец)

### PATCH `/users/me/store` 🔒
Обновление магазина (название, описание, логотип). Только владелец магазина.

**Request:**
```json
{
  "name": "Новое название",
  "description": "Новое описание",
  "logo_url": "https://..."
}
```

> Все поля необязательные — обновляются только переданные.

**Response 200:** (как GET `/users/me/store`)

**Ошибки:**
- `NOT_FOUND` (404) — магазин не найден

> Магазин создаётся **только** через одобрение заявки продавца (`PATCH /admin/seller-applications/{id}` со статусом `approved`). Прямого эндпоинта создания магазина нет — это защита от обхода модерации.

---

## Seller Applications

### POST `/seller-applications` 🔒
Подача заявки на статус продавца. После одобрения админом пользователь получает роль `seller` и магазин.

**Request:**
```json
{
  "store_name": "Мой магазин",
  "description": "Описание магазина"
}
```

**Правила валидации:**
- `store_name` — обязательное
- `description` — необязательное

**Response 201:**
```json
{
  "id": 1,
  "user_id": 1,
  "store_name": "Мой магазин",
  "description": "Описание магазина",
  "status": "pending",
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T10:00:00Z"
}
```

**Ошибки:**
- `CONFLICT` (409) — уже есть pending-заявка

### GET `/seller-applications/me` 🔒
Список заявок текущего пользователя.

**Response 200:**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "store_name": "Мой магазин",
    "description": "Описание магазина",
    "status": "pending",
    "created_at": "2026-07-19T10:00:00Z",
    "updated_at": "2026-07-19T10:00:00Z"
  }
]
```

### GET `/admin/seller-applications` 🔒 `[admin]`
Список pending-заявок продавцов.

**Query params:** `page` (default 1), `limit` (default 20, max 100)

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "user_id": 1,
      "store_name": "Мой магазин",
      "description": "Описание магазина",
      "status": "pending",
      "created_at": "2026-07-19T10:00:00Z",
      "updated_at": "2026-07-19T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

### PATCH `/admin/seller-applications/{id}` 🔒 `[admin]`
Одобрение или отклонение заявки продавца.

**Request:**
```json
{
  "status": "approved",
  "reason": "Текст причины отклонения"
}
```

> Допустимые статусы: `approved`, `rejected`. При `approved` пользователь получает роль `seller` и магазин создаётся в той же транзакции. `reason` используется при отклонении (подставляется в email).

**Response 200:** (как POST `/seller-applications`)

**Ошибки:**
- `NOT_FOUND` (404) — заявка не найдена
- `CONFLICT` (409) — заявка уже обработана
- `VALIDATION_ERROR` (400) — невалидный статус

---

## Products

### GET `/products`
Список товаров с фильтрацией и пагинацией. Только товары со статусом `active`.

**Query params:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `page` | int | Страница (default 1, min 1) |
| `limit` | int | Лимит (default 20, max 100) |
| `category_id` | int | Фильтр по категории |
| `store_id` | int | Фильтр по магазину |
| `min_price` | float | Минимальная цена |
| `max_price` | float | Максимальная цена |
| `search` | string | Поиск по названию (ILIKE) |
| `sort` | string | Сортировка: `price_asc`, `price_desc`, `created_desc` (default) |

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "title": "Ноутбук",
      "price": "50000",
      "stock": 10,
      "store": {
        "id": 1,
        "name": "ТехноМир"
      },
      "created_at": "2026-07-19T10:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

### GET `/products/{id}`
Карточка товара.

**Response 200:**
```json
{
  "id": 1,
  "title": "Ноутбук",
  "description": "Мощный ноутбук для работы",
  "price": "50000",
  "stock": 10,
  "status": "active",
  "images": [
    {"id": 1, "url": "https://...", "position": 0}
  ],
  "store": {
    "id": 1,
    "name": "ТехноМир",
    "description": "Магазин электроники"
  },
  "category": {
    "id": 5,
    "name": "Электроника",
    "slug": "electronics"
  },
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T10:00:00Z"
}
```

**Ошибки:**
- `NOT_FOUND` (404) — товар не найден

### POST `/products` 🔒 `[seller]`
Создание товара. Новый товар получает статус `pending` (требует модерации).

**Request:**
```json
{
  "category_id": 5,
  "title": "Ноутбук",
  "description": "Мощный ноутбук для работы",
  "price": "50000.00",
  "stock": 10,
  "images": ["https://..."]
}
```

**Правила валидации:**
- `title` — обязательное
- `price` — обязательное, положительное число
- `stock` — неотрицательное
- `category_id`, `description`, `images` — необязательные

**Response 201:**
```json
{
  "id": 1,
  "title": "Ноутбук",
  "description": "Мощный ноутбук для работы",
  "price": "50000",
  "stock": 10,
  "status": "pending",
  "images": [
    {"id": 1, "url": "https://...", "position": 0}
  ],
  "store": {
    "id": 1,
    "name": "ТехноМир",
    "description": "Магазин электроники"
  },
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T10:00:00Z"
}
```

**Ошибки:**
- `VALIDATION_ERROR` (400) — пустой title, невалидная цена
- `FORBIDDEN` (403) — пользователь не продавец

### PATCH `/products/{id}` 🔒 `[seller]`
Обновление товара. Только владелец магазина.

**Request:**
```json
{
  "title": "Ноутбук Pro",
  "price": "55000.00",
  "stock": 8
}
```

> Все поля необязательные — обновляются только переданные.

**Response 200:** (как GET `/products/{id}`)

**Ошибки:**
- `FORBIDDEN` (403) — не владелец
- `NOT_FOUND` (404) — товар не найден

### DELETE `/products/{id}` 🔒 `[seller]`
Удаление товара. Только владелец магазина.

**Response 204:** No Content

### POST `/products/{id}/images` 🔒 `[seller]`
Загрузка изображения для товара. Файл отправляется как `multipart/form-data`.

**Query params:** нет

**Request:** `multipart/form-data` с полем `file`

**Поддерживаемые типы:** `image/jpeg`, `image/png`, `image/webp`, `image/gif`, `image/svg+xml`

**Максимальный размер:** 10 МБ

**Response 201:**
```json
{
  "id": 1,
  "url": "https://...",
  "position": 0
}
```

**Ошибки:**
- `VALIDATION_ERROR` (400) — файл не указан, неподдерживаемый тип, пустой файл
- `NOT_FOUND` (404) — товар не найден
- `FORBIDDEN` (403) — не владелец магазина

### DELETE `/products/images/{id}` 🔒 `[seller]`
Удаление изображения товара. Только владелец магазина.

**Response 204:** No Content

**Ошибки:**
- `NOT_FOUND` (404) — изображение не найдено
- `FORBIDDEN` (403) — не владелец магазина

### GET `/products/{id}/images/presigned` 🔒 `[seller]`
Получение presigned URL для прямой загрузки файла в S3/MinIO (обход бэкенда).

**Query params:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `content_type` | string | MIME-тип файла (по умолчанию `application/octet-stream`) |

**Response 200:**
```json
{
  "upload_url": "https://minio:9000/bucket/products/images/uuid.jpg?X-Amz-...",
  "object_key": "products/images/uuid.jpg",
  "content_type": "image/jpeg"
}
```

> `upload_url` — прямая ссылка для `PUT` запроса с телом файла. Срок действия — 15 минут.

**Ошибки:**
- `NOT_FOUND` (404) — товар не найден
- `FORBIDDEN` (403) — не владелец магазина

### GET `/products/categories`
Дерево категорий (рекурсивный CTE).

**Response 200:**
```json
[
  {
    "id": 1,
    "name": "Электроника",
    "slug": "electronics",
    "children": [
      {
        "id": 5,
        "name": "Ноутбуки",
        "slug": "laptops"
      }
    ]
  }
]
```

### PATCH `/products/{id}/moderate` 🔒 `[admin]`
Модерация товара.

**Request:**
```json
{
  "status": "rejected",
  "reason": "Фото не соответствует описанию"
}
```

> Допустимые статусы: `active`, `rejected`, `archived`. Поле `reason` обязательно при отклонении (`rejected`) — сохраняется в `rejection_reason` и подставляется в email продавцу.

**Response 200:** (как GET `/products/{id}`, включая `rejection_reason` при отклонении)

**Ошибки:**
- `FORBIDDEN` (403) — не админ
- `VALIDATION_ERROR` (400) — невалидный статус

### GET `/admin/products` 🔒 `[admin]`
Список товаров по статусу модерации (для админ-панели).

**Query params:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `status` | string | Статус: `pending`, `active`, `rejected`, `archived` (обязательный) |
| `page` | int | Страница (default 1) |
| `limit` | int | Лимит (default 20, max 100) |

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "title": "Ноутбук",
      "price": "50000",
      "stock": 10,
      "status": "pending",
      "store": {
        "id": 1,
        "name": "ТехноМир"
      },
      "created_at": "2026-07-19T10:00:00Z"
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 20
}
```

**Ошибки:**
- `VALIDATION_ERROR` (400) — невалидный статус
- `FORBIDDEN` (403) — не админ

---

## Orders

### POST `/orders`
Создание заказа. Списывает товар со склада в транзакции. Не требует авторизации: для авторизованных пользователей `user_id` берётся из JWT, для гостей — заполняются `customer` и `delivery`.

**Request (авторизованный пользователь):**
```json
{
  "items": [
    {"product_id": 1, "quantity": 2},
    {"product_id": 5, "quantity": 1}
  ],
  "delivery": {
    "address": "г. Москва, ул. Пушкина, д. 10, кв. 5",
    "method": "courier"
  }
}
```

**Request (гость):**
```json
{
  "customer": {
    "first_name": "Иван",
    "last_name": "Петров",
    "email": "ivan@example.com",
    "phone": "+79991234567"
  },
  "items": [
    {"product_id": 1, "quantity": 2}
  ],
  "delivery": {
    "address": "г. Москва, ул. Пушкина, д. 10, кв. 5",
    "method": "courier"
  }
}
```

**Правила валидации:**
- `items` — минимум 1 элемент
- `quantity` — больше 0
- `delivery.address` — обязательное
- Для гостя: `customer.first_name`, `customer.email`, `customer.phone` — обязательные
- Товар должен существовать и иметь достаточный остаток
- `delivery.cost` игнорируется сервером (всегда 0) — защита от подмены цены доставки

**Response 201:**
```json
{
  "id": 1,
  "status": "pending",
  "total": "105000",
  "address": "г. Москва, ул. Пушкина, д. 10, кв. 5",
  "public_token": "a1b2c3...",
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "quantity": 2,
      "price": "50000",
      "product_title": "Ноутбук"
    }
  ],
  "created_at": "2026-07-19T10:00:00Z"
}
```

> `public_token` возвращается только для гостевых заказов (по нему доступны статус и оплата без авторизации). Для авторизованных заказов поле отсутствует.

**Ошибки:**
- `VALIDATION_ERROR` (400) — пустой заказ, `quantity <= 0`, не указан адрес, не заполнены данные гостя
- `NOT_FOUND` (404) — товар не найден
- `VALIDATION_ERROR` (400) — недостаточный остаток

### GET `/orders/public/{public_token}`
Статус и детали гостевого заказа по публичному токену. Не требует авторизации.

**Response 200:** (как POST `/orders`, включая `customer` и `delivery_method`/`delivery_cost`)

**Ошибки:**
- `NOT_FOUND` (404) — заказ не найден

### POST `/orders/public/{public_token}/payment`
Инициализация платежа для гостевого заказа. Не требует авторизации.

**Response 201:** (как POST `/payments`)

**Ошибки:**
- `NOT_FOUND` (404) — заказ не найден
- `CONFLICT` (409) — платёж уже существует
- `VALIDATION_ERROR` (400) — заказ уже оплачен

### GET `/orders/me` 🔒
Список заказов текущего пользователя (покупателя).

**Query params:** `page` (default 1), `limit` (default 20, max 100)

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "status": "delivered",
      "total": "1500",
      "address": "г. Москва, ул. Пушкина, д. 10",
      "items_count": 3,
      "tracking_number": "RU123456789",
      "created_at": "2026-07-19T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

> `tracking_number` присутствует только если задан.

### GET `/orders/me/{id}` 🔒
Детали заказа. Доступен покупателю (владельцу) и продавцам, чьи товары в заказе.

**Response 200:** (как в POST `/orders`)

**Ошибки:**
- `NOT_FOUND` (404) — заказ не найден
- `FORBIDDEN` (403) — доступ запрещён (не покупатель и не продавец товаров в заказе)

### GET `/orders/seller` 🔒 `[seller]`
Заказы, содержащие товары текущего продавца.

**Query params:** `page` (default 1), `limit` (default 20, max 100)

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "status": "paid",
      "total": "105000",
      "buyer": {
        "id": 10,
        "username": "ivan"
      },
      "customer": {
        "first_name": "Иван",
        "last_name": "Петров",
        "email": "ivan@example.com",
        "phone": "+79991234567"
      },
      "items": [
        {
          "id": 1,
          "product_id": 1,
          "quantity": 2,
          "price": "50000",
          "product_title": "Ноутбук"
        }
      ],
      "created_at": "2026-07-19T10:00:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

> `customer` присутствует только для гостевых заказов. `buyer` — для авторизованных.

### PATCH `/orders/{id}/status` 🔒 `[seller]`
Обновление статуса заказа. Доступна продавцу (для товаров в заказе) и покупателю (владельцу заказа).

**Request:**
```json
{
  "status": "shipped"
}
```

**Допустимые переходы:**

| Текущий статус | Доступные переходы | Кто может |
|----------------|-------------------|-----------|
| `pending` | `cancelled` | покупатель, продавец |
| `paid` | `shipped`, `cancelled` | продавец |
| `paid` | `cancelled` | покупатель |
| `shipped` | `delivered` | продавец |

> Статус `paid` устанавливается **только** через платёжный webhook, недоступен вручную.

**Response 200:** (как GET `/orders/me/{id}`)

**Ошибки:**
- `VALIDATION_ERROR` (400) — невалидный статус или недопустимый переход
- `FORBIDDEN` (403) — не продавец и не покупатель
- `NOT_FOUND` (404) — заказ не найден

### PATCH `/orders/{id}/tracking` 🔒 `[seller]`
Установка трек-номера доставки. Доступна продавцу (для товаров в заказе) и админу.

**Request:**
```json
{
  "tracking_number": "RU123456789"
}
```

**Response 200:** (как GET `/orders/me/{id}`)

**Ошибки:**
- `FORBIDDEN` (403) — не продавец товаров в заказе
- `NOT_FOUND` (404) — заказ не найден

### Автоматическая отмена просроченных заказов

Неоплаченные заказы в статусе `pending` автоматически отменяются через **24 часа** после создания (`expires_at`). Фоновый процесс `OrderExpirer` (тикер 5 минут) находит просроченные заказы и переводит их в `cancelled`, возвращая остатки товаров на склад в одной транзакции.

---

## Payments

### POST `/payments` 🔒
Инициализация платежа по заказу. Заказ должен быть в статусе `pending`.

**Request:**
```json
{
  "order_id": 1
}
```

**Response 201:**
```json
{
  "id": 1,
  "order_id": 1,
  "amount": "105000",
  "currency": "RUB",
  "status": "pending",
  "provider": "mock",
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T10:00:00Z"
}
```

**Ошибки:**
- `NOT_FOUND` (404) — заказ не найден
- `CONFLICT` (409) — платёж уже существует
- `VALIDATION_ERROR` (400) — заказ уже оплачен

### GET `/payments/{id}` 🔒
Статус платежа. Доступен только владельцу заказа.

**Response 200:**
```json
{
  "id": 1,
  "order_id": 1,
  "amount": "105000",
  "currency": "RUB",
  "status": "succeeded",
  "provider": "mock",
  "provider_payment_id": "pm-123456",
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T10:05:00Z"
}
```

> Поле `provider_payment_id` присутствует только если задано (`omitempty`).

**Ошибки:**
- `NOT_FOUND` (404) — платёж не найден

### POST `/payments/webhook`
Webhook от платёжного провайдера. Не требует JWT. При статусе `succeeded` — переводит заказ в `paid`.

**Request:**
```json
{
  "order_id": 1,
  "provider_payment_id": "pm-123456",
  "status": "succeeded"
}
```

> Допустимые статусы: `succeeded`, `failed`.

**Response 200:**
```json
{
  "status": "ok"
}
```

**Ошибки:**
- `NOT_FOUND` (404) — платёж не найден
- `VALIDATION_ERROR` (400) — невалидный статус

### POST `/payments/{id}/refund` 🔒
Возврат средств. Доступен владельцу заказа. Платёж должен быть в статусе `succeeded`.

**Response 200:**
```json
{
  "id": 1,
  "order_id": 1,
  "amount": "105000",
  "currency": "RUB",
  "status": "refunded",
  "provider": "mock",
  "provider_payment_id": "pm-123456",
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T11:00:00Z"
}
```

**Ошибки:**
- `NOT_FOUND` (404) — платёж не найден
- `VALIDATION_ERROR` (400) — возврат невозможен (не `succeeded`)