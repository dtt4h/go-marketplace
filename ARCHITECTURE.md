# Архитектура

## Обзор

Модульный монолит на Go с PostgreSQL. Модули разделены по бизнес-доменам, каждый имеет чёткие границы. Frontend — SPA на React.

```
┌─────────────┐     HTTPS      ┌──────────────────────────────────┐
│  Frontend   │ ────────────── │             Backend               │
│  React SPA  │                │          (Go, Chi)                │
└─────────────┘                │                                   │
                               │  ┌─────────┐  ┌─────────┐         │
                               │  │  Auth   │  │  Users  │         │
                               │  └─────────┘  └─────────┘         │
                               │  ┌─────────┐  ┌─────────┐         │
                               │  │Products │  │ Orders  │         │
                               │  └─────────┘  └─────────┘         │
                               │  ┌─────────┐                      │
                               │  │Payments │                      │
                               │  └─────────┘                      │
                               └───────────────┬───────────────────┘
                                               │
                                      ┌────────▼────────┐
                                      │   PostgreSQL    │
                                      └─────────────────┘
```

## Структура проекта

```
marketplace/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── config/                  # Конфигурация (env, flags)
│   ├── server/                  # HTTP-сервер, middleware, роутинг
│   ├── database/                # Подключение к БД, миграции
│   ├── logger/                  # Настройка slog
│   └── modules/
│       ├── auth/
│       │   ├── handler.go       # HTTP-хендлеры
│       │   ├── service.go       # Бизнес-логика
│       │   ├── repository.go    # Слой данных
│       │   ├── models.go        # Доменные модели
│       │   └── routes.go        # Регистрация роутов
│       ├── users/
│       │   └── ... (аналогично)
│       ├── products/
│       │   └── ...
│       ├── orders/
│       │   └── ...
│       └── payments/
│           └── ...
├── migrations/                  # SQL-миграции (golang-migrate)
├── pkg/                         # Переиспользуемые утилиты
│   ├── httputil/                # HTTP-хелперы (ответы, ошибки)
│   └── validator/               # Валидация
├── deployments/
│   ├── docker-compose.yml
│   ├── Dockerfile
│   └── nginx.conf
├── docs/                        # Документация
├── API.md
├── ARCHITECTURE.md
├── DATABASE.md
├── go.mod
└── go.sum
```

## Слои внутри модуля

Каждый модуль следует трёхслойной архитектуре:

| Слой | Файл | Ответственность |
|------|------|-----------------|
| **Handler** | `handler.go` | Приём HTTP-запросов, валидация входных данных, формирование HTTP-ответов |
| **Service** | `service.go` | Бизнес-логика, координация между модулями |
| **Repository** | `repository.go` | Доступ к данным (SQL-запросы через pgx) |

Зависимости направлены внутрь: Handler → Service → Repository. Модули не импортируют друг друга напрямую — взаимодействие через интерфейсы на уровне Service.

## Межмодульное взаимодействие

Модули общаются через публичные интерфейсы, определённые в `service.go` каждого модуля:

```go
// internal/modules/orders/service.go
type ProductsService interface {
    GetByID(ctx context.Context, id int64) (Product, error)
    DecrementStock(ctx context.Context, id int64, qty int) error
}
```

Orders-модуль зависит от интерфейса, а не от конкретной реализации Products. Реализация внедряется через DI при сборке приложения.

## Middleware

| Middleware | Назначение |
|------------|------------|
| `RequestID` | Генерация уникального ID запроса |
| `Logger` | Логирование запросов/ответов (slog) |
| `Recover` | Перехват паник |
| `CORS` | Разрешение кросс-доменных запросов |
| `JWTAuth` | Проверка JWT-токена, извлечение пользователя |
| `RoleGuard` | Проверка роли пользователя (`seller`, `admin`, ...) |

## Аутентификация

- **Access token** — JWT, срок жизни 15 минут, передаётся в `Authorization: Bearer <token>`.
- **Refresh token** — JWT, срок жизни 7 дней, хранится в http-only cookie.
- Пароли хешируются через bcrypt (cost = 12).
- При истечении access token — клиент использует refresh token для получения новой пары.

## Обработка ошибок

Единый формат ошибок API:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required",
    "details": {}
  }
}
```

Коды ошибок:

| Код | HTTP | Описание |
|-----|------|----------|
| `VALIDATION_ERROR` | 400 | Некорректные входные данные |
| `UNAUTHORIZED` | 401 | Не авторизован |
| `FORBIDDEN` | 403 | Недостаточно прав |
| `NOT_FOUND` | 404 | Ресурс не найден |
| `CONFLICT` | 409 | Конфликт (например, email занят) |
| `INTERNAL_ERROR` | 500 | Внутренняя ошибка сервера |

## Конфигурация

Все настройки через переменные окружения (`.env`):

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `APP_ENV` | `development` | Окружение: development/production |
| `HTTP_PORT` | `8080` | Порт HTTP-сервера |
| `DB_HOST` | `localhost` | Хост PostgreSQL |
| `DB_PORT` | `5432` | Порт PostgreSQL |
| `DB_USER` | — | Пользователь БД |
| `DB_PASSWORD` | — | Пароль БД |
| `DB_NAME` | `marketplace` | Имя БД |
| `DB_SSLMODE` | `disable` | SSL-режим |
| `JWT_SECRET` | — | Секрет для подписи JWT |
| `JWT_ACCESS_TTL` | `15m` | Время жизни access token |
| `JWT_REFRESH_TTL` | `168h` | Время жизни refresh token |
| `PAYMENT_GATEWAY_URL` | — | URL платёжного шлюза |
| `PAYMENT_GATEWAY_KEY` | — | API-ключ платёжного шлюза |

## Развертывание

Docker Compose поднимает 3 сервиса:

| Сервис | Образ | Назначение |
|--------|-------|------------|
| `app` | Сборка из Dockerfile | Backend (Go) |
| `db` | `postgres:16` | База данных |
| `nginx` | `nginx:alpine` | Reverse proxy, раздача статики |

## Будущее масштабирование

Модульный монолит спроектирован так, чтобы любой модуль можно было выделить в отдельный сервис:

1. Вынести БД модуля в отдельную схему/инстанс.
2. Заменить прямые вызовы Service на HTTP/gRPC.
3. Добавить message broker (NATS/Kafka) для асинхронных событий.
