# Архитектура

## Обзор

Модульный монолит на Go + PostgreSQL. Каждый модуль — бизнес-домен (Auth, Users, Products, Orders, Payments). Модули изолированы, взаимодействуют через интерфейсы.

```
┌──────────────┐     HTTPS      ┌──────────────────────────────────────────┐
│   Frontend   │ ────────────── │              Backend (Go, Chi)            │
│   React SPA  │                │                                          │
└──────────────┘                │  ┌───────────┐  ┌───────────┐            │
                               │  │ Middleware │  │  Router   │            │
                               │  │ chain     │  │ (chi/v5)  │            │
                               │  └─────┬─────┘  └─────┬─────┘            │
                               │        │               │                   │
                               │  ┌─────▼───────────────▼─────┐            │
                               │  │     Modules (handler →    │            │
                               │  │  service → repository)    │            │
                               │  │  ┌───┐ ┌───┐ ┌───┐ ┌───┐ │            │
                               │  │  │Au │ │Us │ │Pr │ │Or │ │            │
                               │  │  │th │ │rs │ │od │ │dr │ │            │
                               │  │  └───┘ └───┘ └───┘ └───┘ │            │
                               │  └──────────────┬────────────┘             │
                               │                  │                         │
                               │         ┌────────▼────────┐                │
                               │         │  PostgreSQL 16  │                │
                               │         │   (через pgx)   │                │
                               │         └─────────────────┘                │
                               └──────────────────────────────────────────┘
```

## Структура проекта

```
marketplace/
├── cmd/
│   └── server/
│       └── main.go               # Точка входа: config → DB → migrations → server
├── internal/
│   ├── config/                   # Загрузка конфига из .env / переменных окружения
│   │   └── config.go
│   ├── server/                   # HTTP-сервер, middleware, роутинг
│   │   ├── server.go             # Инициализация chi, middleware, запуск/остановка
│   │   ├── routes.go             # Регистрация всех роутов (/api/v1/...)
│   │   ├── dtos/                 # DTO для запросов/ответов (по модулям)
│   │   │   └── auth.go
│   │   └── middleware/           # Слой middleware
│   │       ├── auth.go           # JWTAuth, RoleGuard, UserIDFromCtx
│   │       ├── cors.go           # CORS
│   │       └── logger.go         # Логирование запросов
│   ├── database/                 # Подключение к БД, sqlc-сгенерированный код
│   │   ├── database.go           # Подключение pgxpool
│   │   └── sqlc/                 # Сгенерировано sqlc (не редактировать вручную!)
│   │       ├── db.go             # Конструктор Queries
│   │       ├── models.go         # User, RefreshToken, все enum
│   │       └── auth.sql.go       # Сгенерированные запросы из sql/queries/auth.sql
│   ├── logger/                   # Настройка slog
│   │   └── logger.go
│   └── modules/                  # Бизнес-модули
│       └── auth/                 # Модуль аутентификации (реализован)
│           ├── handler.go        # HTTP-хендлеры (Register, Login, Refresh, Logout)
│           ├── service.go        # Бизнес-логика (bcrypt, JWT, refresh-ротация)
│           └── repository.go     # Слой данных (обёртка над sqlc.Queries)
├── pkg/                          # Переиспользуемые утилиты
│   └── httputil/                 # HTTP-хелперы (JSON, DecodeJSON, Error, ...)
│       ├── pagination.go
│       └── response.go
├── sql/                          # SQL-файлы для sqlc
│   ├── schemas/                  # Копия схемы для sqlc (дублирует migrations)
│   │   └── 000001_init.up.sql
│   └── queries/                  # Именованные SQL-запросы
│       └── auth.sql
├── migrations/                   # SQL-миграции (golang-migrate)
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── docs/                         # Документация
│   ├── PRD.md
│   ├── use-cases.md
│   └── adr/
│       ├── adr-001.md
│       ├── adr-002.md
│       └── adr-003.md
├── docker-compose.yaml           # app + db (postgres:16-alpine)
├── Dockerfile                    # Многостадийная сборка: builder → alpine
├── sqlc.yaml                     # Конфигурация sqlc v2
├── Makefile
├── go.mod / go.sum
└── .env.example
```

## Слои внутри модуля

Каждый модуль следует трёхслойной архитектуре:

```
HTTP Request → Handler → Service → Repository → sqlc → PostgreSQL
```

| Слой | Файл | Ответственность |
|------|------|-----------------|
| **Handler** | `handler.go` | Декодинг JSON, вызов Service, формирование HTTP-ответа (201/200/204/ошибки) |
| **Service** | `service.go` | Бизнес-логика, валидация, генерация токенов, хеширование |
| **Repository** | `repository.go` | Обёртка над sqlc.Queries, скрытие pgtype-типов |

**Правила:**
- Handler и Service общаются через DTO (из `internal/server/dtos/`)
- Repository возвращает `db.User` (sqlc-модель) — конвертация в DTO происходит в Service/Handler
- Service не импортирует `net/http`
- Handler не импортирует `database/sqlc` напрямую

## Роутинг

Все роуты регистрируются в `internal/server/routes.go`. Принцип:

```go
func (s *Server) registerAuthRoutes(r chi.Router) {
    queries := db.New(s.db)          // sqlc-конструктор
    repo := auth.NewAuthRepository(queries)
    svc := auth.NewAuthService(repo, s.cfg)
    handler := auth.NewAuthHandler(svc)

    r.Post("/auth/register", handler.Register)                    // без JWT
    r.Post("/auth/login", handler.Login)
    r.Post("/auth/refresh", handler.Refresh)
    r.With(mw.JWTAuth(s.cfg)).Post("/auth/logout", handler.Logout) // с JWT
}
```

Middleware применяется **к конкретному эндпоинту** через `r.With(mw.JWTAuth(cfg)).Post(...)`, а не глобально.

## Middleware (порядок применения)

1. `RequestID` — генерация уникального ID запроса
2. `RealIP` — определение реального IP
3. `Logger` — логирование запросов/ответов (slog)
4. `Recoverer` — перехват паник
5. `CORS` — разрешение кросс-доменных запросов
6. `JWTAuth` — проверка JWT (только на защищённые роуты)
7. `RoleGuard` — проверка роли (только на роуты с ролевой моделью)

## Аутентификация (Auth Module)

### Поток регистрации
1. Handler принимает `{email, password, username}`
2. Service проверяет длину пароля (≥ 8 символов)
3. Хеширует пароль bcrypt (cost=12)
4. Создаёт пользователя через Repository (role= buyer)
5. Генерирует access + refresh JWT
6. Сохраняет refresh-токен в БД
7. Возвращает `{user, access_token, refresh_token}`

### Поток логина
1. Handler принимает `{email, password}`
2. Service ищет пользователя по email
3. Сравнивает пароль через bcrypt
4. Генерирует токены, сохраняет refresh
5. Возвращает `{user, access_token, refresh_token}`

### Поток refresh (ротация)
1. Handler принимает `{refresh_token}`
2. Service ищет токен в БД
3. Проверяет срок годности (`expires_at`)
4. Если истёк — удаляет и возвращает ошибку
5. Получает пользователя по user_id из токена
6. Удаляет старый refresh-токен
7. Создаёт новый refresh-токен
8. Возвращает новую пару токенов

### Поток logout
1. Middleware JWTAuth проверяет access-токен, извлекает user_id
2. Handler получает user_id из контекста (`mw.UserIDFromCtx`)
3. Service удаляет **все** refresh-токены пользователя (завершение всех сессий)

### JWT токены

| Параметр | Access Token | Refresh Token |
|----------|-------------|---------------|
| Срок жизни | 15 минут | 7 дней |
| Хранение | В памяти клиента | В БД + у клиента |
| Подпись | HS256 | HS256 |
| Payload | `user_id`, `role`, `exp`, `iat` | `user_id`, `role`, `exp`, `iat` |

## Обработка ошибок

Единый формат через `pkg/httputil`:

```go
httputil.JSON(w, http.StatusCreated, resp)          // 201
httputil.NoContent(w)                                // 204
httputil.ValidationError(w, "message", details)     // 400
httputil.Unauthorized(w, "message")                 // 401
httputil.Forbidden(w, "message")                    // 403
httputil.NotFound(w, "message")                     // 404
httputil.Conflict(w, "message")                     // 409
httputil.InternalError(w, "error details")          // 500 (логирует детали, клиенту — generic message)
```

**Важно:** `InternalError` логирует переданный текст на сервере, но клиенту всегда возвращает `"internal server error"` (без раскрытия деталей).

## Конфигурация

Все настройки через переменные окружения (`.env`). Загрузка — `config.Load()`:

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `APP_ENV` | `development` | Окружение: development/production |
| `HTTP_PORT` | `8080` | Порт HTTP-сервера |
| `HTTP_READ_TIMEOUT` | `10s` | Таймаут чтения |
| `HTTP_WRITE_TIMEOUT` | `10s` | Таймаут записи |
| `HTTP_IDLE_TIMEOUT` | `60s` | Таймаут idle |
| `DB_HOST` | `localhost` | Хост PostgreSQL |
| `DB_PORT` | `5432` | Порт PostgreSQL |
| `DB_USER` | `postgres` | Пользователь БД |
| `DB_PASSWORD` | `postgres` | Пароль БД |
| `DB_NAME` | `marketplace` | Имя БД |
| `DB_SSLMODE` | `disable` | SSL-режим |
| `JWT_SECRET` | **обязателен** | Секрет для подписи JWT |
| `JWT_ACCESS_TTL` | `15m` | Время жизни access token |
| `JWT_REFRESH_TTL` | `168h` | Время жизни refresh token (7 дней) |
| `PAYMENT_GATEWAY_URL` | — | URL платёжного шлюза |
| `PAYMENT_GATEWAY_KEY` | — | API-ключ платёжного шлюза |
| `REDIS_HOST` | `localhost` | Хост Redis |
| `REDIS_PORT` | `6379` | Порт Redis |

## sqlc

SQL-запросы пишутся в `sql/queries/`, схема — в `sql/schemas/`. Генерация:

```bash
sqlc generate
```

Генерирует код в `internal/database/sqlc/`:
- `db.go` — конструктор `New(db DBTX) *Queries`
- `models.go` — Go-структуры, соответствующие таблицам
- `*.sql.go` — методы с типами параметров и результатами

**Особенности:**
- sqlc v2, `sql_package: "pgx/v5"`
- `emit_json_tags: true`, `json_tags_case_style: "camel"`
- Nullable поля → `pgtype.Text` (не `*string`, не `NullText`)
- Enums → Go-типы (напр. `UserRoleBuyer`, `UserRoleSeller`, `UserRoleAdmin`)

## Развёртывание

```bash
# Первый запуск
docker compose up -d --build

# Пересборка после изменений
docker compose up -d --build app

# Просмотр логов
docker compose logs -f app

# Остановка
docker compose down
```

Docker Compose:

| Сервис | Образ | Назначение |
|--------|-------|------------|
| `app` | Сборка из Dockerfile (многостадийная) | Backend (Go) |
| `db` | `postgres:16-alpine` | База данных |

Приложение подключается к БД через имя сервиса `db` (не `localhost`).

## Будущее масштабирование

Модульный монолит спроектирован так, чтобы любой модуль можно было выделить в отдельный сервис:

1. Вынести БД модуля в отдельную схему/инстанс
2. Заменить прямые вызовы Service на HTTP/gRPC
3. Добавить message broker (NATS/Kafka) для асинхронных событий
4. Добавить Redis для кэша каталога и rate-limiting

## Соглашения для разработчиков

### Именование
- Файлы: `snake_case.go`
- Пакеты: `oneword` (без подчёркиваний, без верблюжьей нотации)
- Экспортируемые функции/типы: `PascalCase`
- Приватные функции/поля: `camelCase`
- Константы ошибок: `ErrDescriptiveName`

### Структура модуля
- `handler.go` — только HTTP
- `service.go` — только бизнес-логика
- `repository.go` — только работа с БД
- DTO — в `internal/server/dtos/<module>.go`

### Импорт sqlc
```go
import db "github.com/.../internal/database/sqlc"
// Использование: db.User, db.UserRoleBuyer, db.New(conn)
```

### Обработка nullable полей
```go
// *string → pgtype.Text (в service.go)
func nullText(s *string) pgtype.Text {
    if s == nil {
        return pgtype.Text{Valid: false}
    }
    return pgtype.Text{String: *s, Valid: true}
}

// pgtype.Text → *string (в DTO)
func textToPtr(t pgtype.Text) *string {
    if t.Valid {
        return &t.String
    }
    return nil
}
```

### Graceful shutdown
Приложение обрабатывает `SIGINT`/`SIGTERM`: даётся 30 секунд на завершение, после чего контекст отменяется и сервер принудительно закрывается.