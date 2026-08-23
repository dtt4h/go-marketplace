# Go Marketplace

Маркетплейс: продавцы создают витрины и управляют товарами, покупатели ищут, заказывают и оплачивают, админ модерирует площадку. Бэкенд — модульный монолит на Go, фронтенд — React SPA, данные — PostgreSQL, картинки — S3/MinIO.

Проект в активной разработке: часть фронтенда пока работает на моках, платёжный шлюз и Redis подключены наполовину (см. [Статус](#статус)).

## Содержание

- [Стек](#стек)
- [Роли](#роли)
- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Структура репозитория](#структура-репозитория)
- [Быстрый старт](#быстрый-старт)
- [Конфигурация](#конфигурация)
- [Команды](#команды)
- [API](#api)
- [Telegram-бот для админа](#telegram-бот-для-админа)
- [База данных](#база-данных)
- [Тесты](#тесты)
- [Статус](#статус)
- [Документация](#документация)

---

## Стек

| Слой | Технология |
|------|------------|
| Backend | Go 1.26, chi/v5, pgx/v5, golang-migrate, sqlc v2, slog |
| Frontend | React 19, TypeScript, Vite, React Router 7, Zustand, react-hook-form + zod, Axios |
| База данных | PostgreSQL 16 (`pg_trgm` для поиска) |
| Хранилище файлов | S3-совместимое (MinIO в Docker / S3 в prod) |
| Инфраструктура | Docker + Docker Compose, многостадийный Dockerfile (non-root) |
| Админ-панель | Telegram-бот (long polling, `gopkg.in/telebot.v3`) |
| Документация API | Swagger (swaggo) |

Почему именно этот стек — ADR-002 (`docs/adr/adr-002.md`). Модульный монолит vs микросервисы — ADR-001.

## Роли

| Роль | Что умеет |
|------|-----------|
| Гость | Каталог, карточка товара, корзина, гостевой заказ с `public_token` |
| Buyer | Всё, что гость, плюс история заказов, корзина на сервере, заявка «стать продавцом» |
| Seller | Каталог товаров своего магазина, обработка заказов, трек-номер, настройки магазина |
| Admin | Модерация товаров, одобрение заявок продавцов, все заказы, статистика (HTTP + Telegram-бот) |

Пользователь может быть одновременно покупателем и продавцом.

## Возможности

### Аутентификация
- Регистрация/вход: bcrypt (cost 12), JWT HS256
- Access token (15 мин) — в памяти клиента, refresh (7 дней) — в httpOnly cookie `SameSite=Strict`
- Ротация refresh-токена: при `/auth/refresh` старый удаляется, создаётся новый
- Logout закрывает **все** сессии пользователя
- Восстановление пароля (`forgot-password` / `reset-password`)
- Rate limiting на `register`/`login`/`forgot-password` (10 запросов/мин на IP)

### Каталог
- Пагинация (`page`, `limit`, max 100)
- Фильтры: категория, магазин, `min_price`/`max_price`, поиск по названию (ILIKE + триграммный GIN-индекс)
- Сортировка: `price_asc`, `price_desc`, `created_desc`
- Дерево категорий (рекурсивный CTE)
- Только товары со статусом `active` видны в публичном каталоге

### Продавцы
- Заявка на статус продавца → модерация админом
- При одобрении: роль→`seller`, магазин создаётся в той же транзакции. Прямого эндпоинта создания магазина нет — защита от обхода модерации
- CRUD товаров (новые — со статусом `pending`)
- Загрузка изображений: через бэкенд (multipart, ≤10 МБ, jpeg/png/webp/gif/svg) или напрямую в S3 по presigned URL
- Модерация с причиной отклонения (`rejection_reason` подставляется в email)

### Заказы
- Авторизованные и гостевые (гостю выдаётся `public_token`)
- `delivery.cost` игнорируется сервером (всегда 0) — защита от подмены цены доставки
- Списывание остатков в атомарной транзакции, `expires_at = now() + 24h`
- Статус-машина: `pending → paid → shipped → delivered`, в `cancelled` из `pending`/`paid`
- `paid` выставляется **только** платёжным webhook'ом
- Фоновая отмена просроченных неоплаченных заказов (`OrderExpirer`, тикер 5 мин) с возвратом остатков

### Платежи
- Мок-провайдер, `confirmation_url` на оплату
- Webhook: `succeeded` → заказ `paid` + письмо с чеком; `failed`; `refunded`
- Возврат средств (`refund`)
- Отдельный эндпоинт для оплаты гостевых заказов по `public_token`

### Уведомления
- SMTP-письма: подтверждение заказа, смена статуса, новый заказ продавцу, чек (HTML с таблицей товаров), одобрение/отклонение заявки, уведомление админу о новой заявке
- Если SMTP не настроен — письма молча пропускаются (лог `WARN`)

### Админ-панель
- HTTP: `/admin/stats`, `/admin/orders*`, `/admin/products?status=`, `PATCH /products/{id}/moderate`, `PATCH /admin/seller-applications/{id}`
- Telegram-бот: статистика, модерация товаров и заявок одной кнопкой (см. ниже)

## Архитектура

### Слои внутри модуля

```
HTTP Request → Handler → Service → Repository → sqlc → PostgreSQL
```

| Слой | Ответственность |
|------|-----------------|
| Handler | Декодинг JSON, извлечение `user_id` из JWT, HTTP-ответ (корректный статус/код ошибки) |
| Service | Бизнес-логика, валидация, транзакции, генерация токенов. Не импортирует `net/http` |
| Repository | Обёртка над `sqlc.Queries`, скрытие `pgtype.*`, работа через транзакции |

Handler ↔ Service общаются через DTO (`internal/server/dtos/`). Service ↔ Repository — на sqlc-моделях (`db.User`, `db.Product` и т.д.).

### Один процесс, два канала входа

```
┌──────────────────── Go процесс (cmd/server) ────────────────────┐
│                                                                  │
│  HTTP :8080 (chi)          Telegram-бот (long polling)          │
│  ┌─────────────────┐       ┌────────────────────────┐           │
│  │ middleware chain │       │  allowlist по user ID  │           │
│  │ (logger, cors,  │       │  /stats /moderate      │           │
│  │  JWT, role)     │       │  /applications /orders │           │
│  └────────┬────────┘       └───────────┬────────────┘           │
│           │                            │                         │
│           └──────────┬─────────────────┘                         │
│                      ▼                                           │
│              Service Layer (один экземпляр)                     │
│   adminStatsSvc · productSvc · sellerAppSvc · adminOrderSvc    │
│                      │                                           │
│                      ▼                                           │
│              PostgreSQL (pgxpool) + S3/MinIO                    │
└─────────────────────────────────────────────────────────────────┘
```

Бот и HTTP-сервер делят одни и те же сервисы — бизнес-логика не дублируется. Frontend не трогается: административная часть живёт только в боте и в HTTP-эндпоинтах.

### Middleware (порядок)

`RequestID → RealIP → Logger → Recoverer → CORS → RateLimit → JWTAuth → RoleGuard`

Rate limit, JWT и роли применяются точечно к роутам через `r.With(...)`, а не глобально.

## Структура репозитория

```
go-marketplace/
├── cmd/server/main.go          # Точка входа: config → DB → миграции → HTTP + бот + OrderExpirer
├── internal/
│   ├── bot/                    # Telegram-бот (handlers, callbacks, inline-кнопки)
│   ├── config/                 # Чтение env
│   ├── server/
│   │   ├── server.go           # chi, middleware, старт/останов
│   │   ├── routes.go           # Все роуты /api/v1
│   │   ├── middleware/         # auth (JWT+role), cors, logger, ratelimit
│   │   └── dtos/               # Все DTO по модулям
│   ├── database/               # pgxpool + sqlc-сгенерированный код (не редактировать)
│   ├── logger/                 # slog: JSON (prod) / text (dev)
│   ├── storage/                # S3/MinIO: upload, delete, presigned
│   └── modules/
│       ├── auth/               # register, login, refresh, logout, forgot/reset
│       ├── users/              # профиль, магазин
│       ├── products/           # каталог, категории, изображения, модерация
│       ├── cart/               # корзина (для авторизованных)
│       ├── orders/             # заказы, статус-машина, expirer, админка/статистика
│       ├── payments/           # mock-платежи, webhook, refund
│       ├── sellerapplications/ # заявки продавцов
│       └── notifications/      # SMTP-письма, HTML-рецепты
├── pkg/httputil/               # JSON, ошибки, пагинация
├── pkg/pgutil/                 # помощники pgtype
├── migrations/                 # SQL-миграции (golang-migrate)
├── sql/queries/                # Запросы для sqlc
├── docs/                       # PRD, use-cases, ADR, Swagger
├── web/                        # Frontend (React + Vite)
├── docker-compose.yaml
├── Dockerfile
├── sqlc.yaml
├── Makefile
└── .env.example
```

Полное описание слоёв, модулей и соглашений — `ARCHITECTURE.md`.

## Быстрый старт

### Через Docker (рекомендую)

```bash
cp .env.example .env          # затем заполни S3 и Telegram-настройки
docker compose up -d --build
```

Поднимутся: `db` (PostgreSQL 16), `minio` (+ инициализация бакета), `app`. Миграции накатятся автоматически при старте.

- API: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html
- Health: http://localhost:8080/health
- MinIO console: http://localhost:9001 (`minioadmin/minioadmin`)

### Локально (без Docker)

Требуется: Go 1.26+, PostgreSQL 16, по желанию MinIO и SMTP.

```bash
# 1. Поднять БД и MinIO, если не хочется вручную:
docker compose up -d db minio

# 2. Собрать и запустить
go run ./cmd/server

# 3. Фронтенд (в отдельном терминале)
cd web
npm install
npm run dev            # http://localhost:5173, прокси /api → :8080
```

`go mod download` сработает и без прокси, но если столкнёшься с
`unrecognized import path "gopkg.in"` — включи прокси:

```bash
GOPROXY=https://proxy.golang.org go mod download
```

## Конфигурация

Все настройки — через переменные окружения (`.env`). Ключевые:

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `APP_ENV` | `development` | `production` включает JSON-логи |
| `HTTP_PORT` | `8080` | Порт API |
| `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME` | `localhost:5432/postgres/postgres/marketplace` | PostgreSQL |
| `JWT_SECRET` | `dev-secret-change-me` | **В prod обязателен** |
| `JWT_ACCESS_TTL` / `JWT_REFRESH_TTL` | `15m` / `168h` | Сроки жизне токенов |
| `S3_ENDPOINT/REGION/ACCESS_KEY/SECRET_KEY/BUCKET` | — | Картинки товаров |
| `SMTP_*` | — | Email-уведомления |
| `TELEGRAM_BOT_TOKEN` | — | Токен бота (см. ниже) |
| `TELEGRAM_ADMIN_IDS` | — | Comma-separated Telegram user ID |
| `REDIS_*` | — | Reserved (см. [Статус](#статус)) |
| `PAYMENT_GATEWAY_URL/KEY` | — | Reserved |
| `PAYMENTS_WEBHOOK_SECRET` | — | Секрет для платеж. webhook |

## Команды

```bash
make run          # go run ./cmd/server
make build        # go build -o bin/server
make test         # go test ./...
make fmt          # go fmt ./...
make vet          # go vet ./...
make sqlc         # sqlc generate
make docker-up    # docker compose up -d --build
make docker-down  # docker compose down
make docker-reset # полный сброс с удалением volume (-v)
```

## API

Базовый URL: `/api/v1`. Все ответы JSON. Документация: `/swagger/index.html`.

| Метод и путь | Описание |
|---------------|----------|
| `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout` | Аутентификация |
| `POST /auth/forgot-password`, `/auth/reset-password` | Сброс пароля |
| `GET/PATCH /users/me` | Профиль |
| `GET/PATCH /users/me/store` | Магазин продавца |
| `GET /products`, `GET /products/categories`, `GET /products/{id}` | Публичный каталог |
| `POST/PATCH/DELETE /products`… | Управление товарами (`[seller]`) |
| `POST /products/{id}/images`, `DELETE /products/images/{id}` | Изображения (`[seller]`) |
| `GET /products/{id}/images/presigned` | Presigned URL для прямого upload в S3 (`[seller]`) |
| `PATCH /products/{id}/moderate`, `GET /admin/products` | Модерация (`[admin]`) |
| `GET/POST/PATCH/DELETE /cart*` | Корзина авторизованных (`[buyer]`) |
| `POST /orders`, `GET /orders/public/{token}` | Гостевой заказ |
| `GET /orders/me`, `/orders/me/{id}` | Заказы покупателя |
| `GET /orders/seller`, `PATCH /orders/{id}/status`, `/orders/{id}/tracking` | Заказы продавца |
| `POST /payments`, `GET /payments/me`, `/payments/{id}`, `POST /payments/webhook`, `POST /payments/{id}/refund` | Платежи |
| `POST /seller-applications`, `GET /seller-applications/me` | Заявки продавцов |
| `GET /admin/seller-applications`, `PATCH /admin/seller-applications/{id}` | Модерация заявок (`[admin]`) |
| `GET /admin/orders*`, `GET /admin/stats` | Админка заказов/статистики (`[admin]`) |

Ошибки в едином формате:

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "...", "details": {} } }
```

Коды: `400 VALIDATION_ERROR`, `401 UNAUTHORIZED`, `403 FORBIDDEN`, `404 NOT_FOUND`, `409 CONFLICT`, `429 RATE_LIMITED`, `500 INTERNAL_ERROR` (детали логируются, клиенту — generic message).

Полное описание всех эндпоинтов: `API.md`.

## Telegram-бот для админа

Админ-панель в Telegram — внутренний инструмент, отдельный от фронта. Работает в том же процессе, что и HTTP-сервер, и переспользует те же сервисы.

### Настройка

1. Создай бота через [@BotFather](https://t.me/BotFather), получи токен.
2. Узнай свой Telegram user ID у [@userinfobot](https://t.me/userinfobot).
3. В `.env`:

```
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
TELEGRAM_ADMIN_IDS=123456789,987654321
```

4. Перезапусти сервис. В логах появится `telegram bot starting`.

Без токена бот не запускается (просто лог). Доступ имеют только ID из `TELEGRAM_ADMIN_IDS`; остальные получают отказ.

### Команды

| Команда | Действие |
|---------|----------|
| `/stats` | Статистика: пользователи, продавцы, товары, заказы по статусам, выручка |
| `/moderate` | Товары на модерации с кнопками ✅/❌ (в один клик) |
| `/applications` | Заявки продавцов, кнопки одобрить/отклонить |
| `/orders` | Последние 10 заказов |
| `/help` | Справка |

## База данных

PostgreSQL 16. Миграции — `golang-migrate` (`migrations/`), применяются при старте приложения. Схема: `users`, `stores`, `categories`, `products` (+`rejection_reason`), `product_images`, `orders` (+гостевые поля, `expires_at`), `order_items`, `payments`, `refresh_tokens`, `cart_items`, `seller_applications`.

Код доступа к БД генерируется `sqlc` из `sql/queries/` — файлы в `internal/database/sqlc/` редактировать вручную нельзя.

```bash
make sqlc          # перегенерировать
make migrate-up    # относительно localhost
make migrate-down
```

Есть индексы на все FK и частые фильтры, GIN-индекс для триграммного поиска, триггер автообновления `updated_at`, enum-типы для статусов и ролей. Детально: `DATABASE.md`.

## Тесты

Go:

```bash
make test          # go test ./... (unit-тесты, 7 модулей)
go test -tags=integration -timeout 180s ./internal/integration/...  # integration (Docker required)
```

Покрыты:
- **Service-слои**: `auth`, `cart`, `orders`, `payments`, `products`, `sellerapplications`, `users` (mock из `pgxmock`)
- **Handler-слои**: `auth` (68 тестов), `products` (38 тестов) — chi router + mock middleware
- **Integration**: `auth`, `products`, `cart` — testcontainers-go + реальная PostgreSQL 16

JS: тестов нет (в планах).

## Статус

| Что | Статус |
|-----|--------|
| Backend API | ✅ Реализовано полностью (auth, users, products, cart, orders, payments, seller-apps, admin, stats) |
| Миграции | ✅ 5 миграций покрывают схему |
| Swagger | ✅ Генерируется через swaggo |
| Telegram-бот | ✅ `/stats`, `/moderate`, `/applications`, `/orders` |
| Frontend (web/) | 🟡 В работе — каталог/карточка/корзина/профиль есть, но хуки используют **моки** вместо HTTP API; кабинет продавца и админка на моках |
| Платежи | 🟡 Mock-провайдер + webhook; реальный шлюз требует `PAYMENT_GATEWAY_URL/KEY` |
| Redis | ✅ Кэш каталога + rate limiting (Redis с in-memory fallback) |
| Rate limiting | ✅ Redis-backed, in-memory fallback для dev |
| CI/CD | ✅ GitHub Actions: build, test, integration tests, Trivy, frontend build, Docker push |
| Веб-сборка веб-фронта | ✅ Dockerfile.web (multi-stage, nginx) |
| Security headers | ✅ Backend middleware + nginx |
| Prometheus metrics | ✅ `/metrics` (counter, histogram, gauge) |
| Health checks | ✅ `/health/live` (liveness), `/health/ready` (readiness: DB + Redis ping) |
| S3 storage | ✅ Graceful degradation (fallback если S3 не настроен) |
| Integration tests | ✅ testcontainers-go + PostgreSQL (build tag `integration`) |
| Production config | ✅ `nginx.prod.conf` (TLS, HSTS), config validation, `stop_grace_period` |
| Backup automation | ✅ `db-backup-cron` сервис (cron, retention) |

---

## Деплой в продакшен

### 1. Подготовка

```bash
# Клонировать репозиторий на сервер
git clone <repo-url> && cd go-marketplace

# Создать .env из примера и заполнить продакшен-значениями
cp .env.example .env
```

**Обязательные переменные для production:**

| Переменная | Требование |
|------------|------------|
| `APP_ENV` | `production` |
| `JWT_SECRET` | Минимум 32 символа, случайный |
| `DB_PASSWORD` | Не `postgres` |
| `REDIS_PASSWORD` | Обязательно |
| `PAYMENTS_WEBHOOK_SECRET` | Обязательно |
| `S3_*` | Реальный S3 (AWS, Yandex Object Storage) |
| `SMTP_*` | Для email-уведомлений |
| `TELEGRAM_BOT_TOKEN` + `TELEGRAM_ADMIN_IDS` | Для админ-бота |

### 2. TLS сертификаты

```bash
# Получить сертификаты через certbot
mkdir -p nginx-ssl
certbot certonly --standalone -d yourdomain.com
cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx-ssl/
cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx-ssl/
```

### 3. Запуск с production-профилем

```bash
# Запуск с автоматическими бэкапами (cron, ежедневно в 03:00)
docker compose --profile production up -d --build
```

Это поднимет:
- `db`, `redis`, `minio` — инфраструктура
- `app` — бэкенд (миграции применяются автоматически)
- `web` — nginx + frontend
- `db-backup-cron` — ежедневный бэкап с retention (по умолчанию 7 дней)

### 4. Nginx с TLS

Для продакшена смонтируйте `nginx.prod.conf` и сертификаты:

```yaml
# docker-compose.override.yml (не коммитить)
services:
  web:
    volumes:
      - ./web/nginx.prod.conf:/etc/nginx/conf.d/default.conf
      - ./nginx-ssl:/etc/nginx/ssl:ro
```

### 5. Мониторинг

```bash
# Prometheus и Grafana (отдельно)
# Импортировать дашборд: docs/grafana/dashboard.json
```

Метрики доступны на `/metrics` (без авторизации, но в prod nginx ограничивает доступ по IP).

### 6. Бэкапы

```bash
# Ручной бэкап
docker compose run --rm db-backup

# Восстановление
./scripts/backup.sh --restore backups/marketplace_20260101_030000.dump

# Автоматические бэкапы — через db-backup-cron (profile: production)
# Retention: BACKUP_RETENTION_DAYS (по умолчанию 7)
```

### 7. Обновление

```bash
git pull
docker compose --profile production up -d --build
# Миграции применятся автоматически при старте app
```

## Документация

| Файл | О чём |
|------|-------|
| `ARCHITECTURE.md` | Слои, модульные правила, middleware, потоки auth/orders/payments |
| `DATABASE.md` | Схема, таблицы, индексы, sqlc |
| `API.md` | Описание всех эндпоинтов с примерами |
| `docs/PRD.md` | Требования к продукту (MVP) |
| `docs/use-cases.md` | Use-case'ы по ролям |
| `docs/adr/adr-001..003` | Решения: модульный монолит, стек, фронтенд |