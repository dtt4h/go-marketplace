# API

Base URL: `/api/v1`

Все ответы — JSON. Кодировка — UTF-8.

## Аутентификация

JWT Bearer token в заголовке:

```
Authorization: Bearer <access_token>
```

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

---

## Auth

### POST `/auth/register`
Регистрация нового пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secret123",
  "first_name": "Иван",
  "last_name": "Иванов"
}
```

**Response 201:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "buyer",
    "first_name": "Иван",
    "last_name": "Иванов"
  },
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

### POST `/auth/login`
Вход.

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
    "first_name": "Иван",
    "last_name": "Иванов"
  },
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

### POST `/auth/refresh`
Обновление токенов.

**Request:**
```json
{
  "refresh_token": "eyJhbGci..."
}
```

**Response 200:**
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

### POST `/auth/logout` 🔒
Выход. Отзыв refresh token.

**Response 204:** No Content

### POST `/auth/password-reset`
Запрос на сброс пароля. Отправляет email со ссылкой.

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Response 204:** No Content

### POST `/auth/password-reset/confirm`
Подтверждение сброса пароля.

**Request:**
```json
{
  "token": "reset-token-from-email",
  "password": "newsecret123"
}
```

**Response 204:** No Content

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
  "first_name": "Иван",
  "last_name": "Иванов",
  "avatar_url": "https://...",
  "phone": "+79991234567",
  "created_at": "2026-07-19T10:00:00Z"
}
```

### PATCH `/users/me` 🔒
Обновление профиля.

**Request:**
```json
{
  "first_name": "Пётр",
  "last_name": "Петров",
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
  "first_name": "Пётр",
  "last_name": "Петров",
  "avatar_url": "https://...",
  "phone": "+79997654321",
  "created_at": "2026-07-19T10:00:00Z"
}
```

### POST `/users/me/store` 🔒
Регистрация в качестве продавца (создание магазина).

**Request:**
```json
{
  "name": "Мой магазин",
  "description": "Описание магазина",
  "logo_url": "https://..."
}
```

**Response 201:**
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

### GET `/users/me/orders` 🔒
История заказов покупателя.

**Query params:** `page`, `limit`, `status`

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "status": "delivered",
      "total": "1500.00",
      "address": "г. Москва, ул. Пушкина, д. 10",
      "items_count": 3,
      "created_at": "2026-07-19T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

---

## Products

### GET `/products`
Список товаров с фильтрацией и пагинацией.

**Query params:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `page` | int | Страница (default 1) |
| `limit` | int | Лимит (default 20, max 100) |
| `category_id` | int | Фильтр по категории |
| `store_id` | int | Фильтр по магазину |
| `min_price` | float | Минимальная цена |
| `max_price` | float | Максимальная цена |
| `search` | string | Поиск по названию |
| `sort` | string | Сортировка: `price_asc`, `price_desc`, `created_desc` (default) |

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "title": "Ноутбук",
      "price": "50000.00",
      "stock": 10,
      "images": ["https://..."],
      "store": {
        "id": 1,
        "name": "ТехноМир"
      },
      "category": {
        "id": 5,
        "name": "Электроника"
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
  "price": "50000.00",
  "stock": 10,
  "status": "active",
  "images": [
    {"id": 1, "url": "https://...", "position": 0},
    {"id": 2, "url": "https://...", "position": 1}
  ],
  "store": {
    "id": 1,
    "name": "ТехноМир",
    "description": "Магазин электроники"
  },
  "category": {
    "id": 5,
    "name": "Электроника"
  },
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T10:00:00Z"
}
```

### POST `/products` 🔒 `[seller]`
Создание товара.

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

**Response 201:**
```json
{
  "id": 1,
  "title": "Ноутбук",
  "description": "Мощный ноутбук для работы",
  "price": "50000.00",
  "stock": 10,
  "status": "pending",
  "images": [
    {"id": 1, "url": "https://...", "position": 0}
  ],
  "category": {
    "id": 5,
    "name": "Электроника"
  },
  "created_at": "2026-07-19T10:00:00Z"
}
```

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

**Response 200:** (как GET `/products/{id}`)

### DELETE `/products/{id}` 🔒 `[seller]`
Удаление товара. Только владелец магазина.

**Response 204:** No Content

### GET `/products/categories`
Дерево категорий.

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
        "slug": "laptops",
        "children": []
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
  "status": "active",
  "reason": null
}
```

**Response 200:** (как GET `/products/{id}`)

---

## Orders

### POST `/orders` 🔒
Создание заказа из корзины.

**Request:**
```json
{
  "items": [
    {"product_id": 1, "quantity": 2},
    {"product_id": 5, "quantity": 1}
  ],
  "address": "г. Москва, ул. Пушкина, д. 10, кв. 5"
}
```

**Response 201:**
```json
{
  "id": 1,
  "status": "pending",
  "total": "105000.00",
  "address": "г. Москва, ул. Пушкина, д. 10, кв. 5",
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "quantity": 2,
      "price": "50000.00",
      "product_title": "Ноутбук"
    },
    {
      "id": 2,
      "product_id": 5,
      "quantity": 1,
      "price": "5000.00",
      "product_title": "Мышка"
    }
  ],
  "created_at": "2026-07-19T10:00:00Z"
}
```

### GET `/orders/{id}` 🔒
Детали заказа. Доступен покупателю (владельцу) и продавцам, чьи товары в заказе.

**Response 200:** (как в POST `/orders`)

### POST `/orders/{id}/cancel` 🔒
Отмена заказа. Доступна покупателю (если статус `pending` или `paid`) и продавцу.

**Response 200:**
```json
{
  "id": 1,
  "status": "cancelled",
  "total": "105000.00",
  "created_at": "2026-07-19T10:00:00Z"
}
```

### GET `/orders/seller` 🔒 `[seller]`
Заказы, содержащие товары текущего продавца.

**Query params:** `page`, `limit`, `status`

**Response 200:**
```json
{
  "items": [
    {
      "id": 1,
      "status": "paid",
      "total": "105000.00",
      "buyer": {
        "id": 10,
        "first_name": "Иван",
        "last_name": "Иванов"
      },
      "items": [
        {
          "product_id": 1,
          "quantity": 2,
          "price": "50000.00",
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

### PATCH `/orders/{id}/status` 🔒 `[seller]`
Обновление статуса заказа продавцом.

**Request:**
```json
{
  "status": "shipped"
}
```

**Response 200:** (как GET `/orders/{id}`)

---

## Payments

### POST `/payments` 🔒
Инициализация платежа по заказу.

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
  "amount": "105000.00",
  "currency": "RUB",
  "status": "pending",
  "provider": "yookassa",
  "confirmation_url": "https://pay.yookassa.ru/..."
}
```

### GET `/payments/{id}` 🔒
Статус платежа.

**Response 200:**
```json
{
  "id": 1,
  "order_id": 1,
  "amount": "105000.00",
  "currency": "RUB",
  "status": "succeeded",
  "provider": "yookassa",
  "provider_payment_id": "pm-123456",
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T10:05:00Z"
}
```

### POST `/payments/webhook`
Webhook от платёжного провайдера. Не требует JWT, авторизация через подпись провайдера.

**Request:** (формат зависит от провайдера)

**Response 200:**
```json
{
  "status": "ok"
}
```

### POST `/payments/{id}/refund` 🔒
Возврат средств. Доступен администратору или при отмене заказа.

**Response 200:**
```json
{
  "id": 1,
  "order_id": 1,
  "amount": "105000.00",
  "currency": "RUB",
  "status": "refunded",
  "provider": "yookassa",
  "created_at": "2026-07-19T10:00:00Z",
  "updated_at": "2026-07-19T11:00:00Z"
}
```
