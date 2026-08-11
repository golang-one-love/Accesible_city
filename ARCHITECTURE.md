# Архитектура «Доступный путь»

Микросервисная система построения доступных маршрутов для маломобильных граждан. Пользователи сообщают о физических барьерах (бордюры, ступени, перекрытые тротуары), модераторы проверяют заявки, а сервис маршрутизации строит обходные пути с учётом доступности.

## 1. Обзор

```
                        ┌─────────────┐
   Браузер ───────────▶ │  Frontend   │  React + TS + Leaflet, :3000 (nginx внутри)
                        │  (React)    │
                        └──────┬──────┘
                               │ http
                        ┌──────▼──────┐
                        │    Nginx    │  обратный прокси, :80/:443
                        │ (proxy)     │  префикс /api/<svc> → /api/v1/<svc>
                        └──┬──┬──┬──┬─┘
                           │  │  │  │
        ┌──────────────────┘  │  │  └─────────────────────┐
        ▼                     ▼  ▼                       ▼
 ┌─────────────┐      ┌─────────────┐           ┌─────────────┐
 │  Auth Svc   │      │ Barrier Svc │           │   POI Svc   │
 │  Go :8081   │      │ Py  :8000   │           │  Py  :8002  │
 └──────┬──────┘      └──────┬──────┘           └─────────────┘
        │                    │ barrier.created
        │                    ▼
        │             ┌─────────────┐       Redis Streams: barrier.{created,approved,rejected,resolved}
        │             │ Moderation  │───────────────────────┐
        │             │  Svc        │   barrier.{approved,rejected,resolved}
        │             │  Py :8001   │                       │
        │             └─────────────┘                       ▼
        │                                          ┌─────────────┐
        │                                          │ Notification│
        │                                          │  Go :8083   │  WebSocket-раздача
        │                                          └─────────────┘
        │
        └──────────────────────────────────▶  Route Svc  (Go :8082)
```

Общие правила:

- **Каждый сервис изолирован**: своя БД, свои ключи доступа к БД, свой контейнер.
- **Межсервисная связь — только через Redis Streams** (события домена), HTTP-общения между сервисами бизнес-уровня нет (исключение: route-service обращается к barrier-service за списком барьеров для построения маршрута).
- **Все сервисы валидируют JWT** публичным ключом; подписывает токены только auth-service (приватный ключ).
- **Внешнее API фронтенда** — `/api/<сервис>/...` через nginx; внутри docker-сети сервисы слушают `/api/v1/<сервис>/...`.
- **Схема БД создаётся на старте сервиса** (Python — SQLAlchemy `create_all`, Go — `EnsureSchema`), миграционные инструменты (alembic, golang-migrate) в проекте не применяются.

## 2. Сервисы

| Сервис | Язык/Фреймворк | Внутренний порт | БД | Ответственность |
|--------|----------------|-----------------|----|-----------------|
| **auth-service** | Go 1.22 / Echo | 8081 | `auth_db` | регистрация, вход, refresh-токены, роли (user/moderator/admin), выпуск JWT (RS256) |
| **barrier-service** | Python 3.12 / FastAPI | 8000 | `barrier_db` | CRUD барьеров, фото (MinIO), подтверждения, жалобы; публикует `barrier.created`; потребляет approve/reject/resolved |
| **moderation-service** | Python 3.12 / FastAPI | 8001 | `moderation_db` | очередь модерации; потребляет `barrier.created`; публикует `barrier.approved` / `barrier.rejected` |
| **route-service** | Go 1.22 / Echo | 8082 | `route_db` | построение маршрута (A* по OSM-графу) с весами по типу барьера и профилю мобильности |
| **poi-service** | Python 3.12 / FastAPI | 8002 | `poi_db` | точки интереса (POI) и их «паспорт доступности» |
| **notification-service** | Go 1.22 / Echo | 8083 | `notification_db` | уведомления (persist + WebSocket push); потребляет `barrier.approved` / `barrier.resolved` |

Вспомогательные контейнеры: `postgres:16-alpine`, `redis:7-alpine`, `minio/minio`, `nginx:alpine`.

## 3. Доменные сущности

### barrier-service

- **Barrier** — `id (UUID)`, `type` (`high_curb, broken_elevator, closed_sidewalk, stairs, pothole, uneven_surface, parked_car`), `coordinates (lat/lon)`, `description`, `severity (1–5: low…blocking)`, `status` (`pending/approved/rejected/resolved`), `reporter_id`, `moderator_id`, `approved_at`, `resolved_at`, `created_at`, `updated_at`
- **BarrierPhoto** — `id`, `barrier_id`, `object_key` (MinIO), `uploaded_by`, `created_at`
- **BarrierConfirmation** — подтверждение барьера пользователем: `id`, `barrier_id`, `user_id`, `created_at`
- **BarrierComplaint** — жалоба на барьер: `id`, `barrier_id`, `user_id`, `reason`, `created_at`

### moderation-service

- **ModerationRequest** — `id`, `barrier_id`, `reporter_id`, `status` (`pending/approved/rejected`), `moderator_id`, `moderator_comment`, `created_at`, `updated_at`, `reviewed_at`
- Бизнес-правило: модератор не может обработать собственную заявку (`Cannot moderate own request`).

### poi-service

- **PointOfInterest** — `id`, `name`, `category` (restaurant, cafe, shop, pharmacy, hospital, clinic, bank, post_office, government, park, museum, theater, library, school, university, hotel, transport, other), `latitude/longitude`, `address`, `accessibility` (список `AccessibilityFeature`: ramp, elevator, wide_door, accessible_toilet, tactile_paving, braille_signs, audio_guide, low_counter, parking, induction_loop), `rating`, `created_by`, `created_at`

### auth-service

- **User** — `id (UUID)`, `email` (value object, валидация), `password_hash` (bcrypt, value object), `role` (`user/moderator/admin`), `is_active`, `created_at`, `updated_at`

### route-service

- **GraphNode / GraphEdge** — узлы и рёбра OSM-графа с координатами
- **MobilityProfile** — профиль мобильности: `wheelchair`, `stroller`, `elderly`, `default`; каждому профилю соответствует вес/штраф для каждого типа барьера
- **Route** — результат построения: `nodes`, `total_distance`, `max_severity`

### notification-service

- **Notification** — `id`, `user_id`, `type` (`barrier_approved, barrier_resolved, barrier_nearby, new_poi`), `title`, `message`, `payload (JSONB)`, `is_read`, `created_at`, `read_at`

## 4. Событийная модель (Redis Streams)

Публикация через `XADD`, потребление — группы потребителей с `XREADGROUP` + `XACK` (consumer groups, а не Pub/Sub).

| Поток | Издатель | Потребители | Содержимое payload |
|-------|----------|-------------|--------------------|
| `barrier.created` | barrier-service | moderation-service | `type, coordinates, description, severity, reporter_id` |
| `barrier.approved` | moderation-service (по результату модерации) и barrier-service (прямой approve) | barrier-service, notification-service | `type, barrier_id, reporter_id, moderator_id, comment` |
| `barrier.rejected` | moderation-service | barrier-service | `type, barrier_id, reporter_id, moderator_id, comment` |
| `barrier.resolved` | barrier-service | notification-service | `type, barrier_id, reporter_id, coordinates` |

Формат события (единый для всех):

```json
{
  "event_type": "barrier.approved",
  "aggregate_id": "4306c807-...",
  "timestamp": "2026-08-11T12:04:08Z",
  "payload": { "...": "..." }
}
```

Полный цикл: **создание барьера → модерация → уведомление**:

```
POST /api/barriers (user)
   │ barrier-service: status=pending, XADD barrier.created
   ▼
moderation-service: XREADGROUP barrier.created → очередь модерации
   │ POST /api/moderation/queue/:id/approve (moderator, не автор)
   ▼
moderation-service: статус запроса=approved, XADD barrier.approved
   ├─▶ barrier-service: XREADGROUP → barrier.status=approved, approved_at
   └─▶ notification-service: XREADGROUP → notifications(row) + WS broadcast
```

## 5. Аутентификация и авторизация

- **JWT RS256.** auth-service генерирует пару ключей (см. `generate_keys.go`), приватный ключ хранит только у себя (`services/auth-service/keys/private.pem`).
- Публичный ключ монтируется в контейнеры barrier / moderation / poi / notification (`./services/auth-service/keys:/keys:ro`) и используется для проверки подписи (`JWT_PUBLIC_KEY_PATH=/keys/public.pem`).
- Токены: `access` (TTL по умолчанию 15 мин), `refresh` (168 ч). В claims: `sub` (UUID пользователя), `email`, `role`, `type` (`access`/`refresh`), `exp`.
- Python-сервисы: единая зависимость `get_current_user_id` (заголовок `Authorization: Bearer`), проверка `type == "access"` (см. `src/adapters/inbound/http/security.py`).
- Go-сервисы: middleware на группе `/api/v1` (notification-service — `JWTUserIDMiddleware`; входные данные — Bearer-заголовок или query `?token=` для WebSocket).
- Роли: `user` (создание барьеров, подтверждения, маршруты), `moderator` (очередь модерации), `admin` (управление).

## 6. БД

`infra/postgres/init.sql` создаёт на старте 6 изолированных БД с отдельными ролями (пароли в `.env`/`.env.example` → `*_DB_USER` / `*_DB_PASSWORD` / `*_DB_NAME`):

| БД | Владелец роли | Таблицы (создаются сервисом при старте) |
|----|---------------|------------------------------------------|
| `auth_db` | `auth_user` | `users` |
| `barrier_db` | `barrier_user` | `barriers`, `barrier_photos`, `barrier_confirmations`, `barrier_complaints` |
| `moderation_db` | `moderation_user` | `moderation_queue` |
| `route_db` | `route_user` | `osm_graph_cache` |
| `poi_db` | `poi_user` | `points_of_interest` |
| `notification_db` | `notification_user` | `notifications` |

Расширения: `uuid-ossp`/`pgcrypto` везде, `postgis` в barrier/route/poi БД.

## 7. Внешние зависимости и данные

- **OpenStreetMap:** route-service тянет граф дорог через `OSM_API_URL` / `OSM_OVERPASS_URL`, кэширует в `osm_graph_cache`. Веса рёбер зависят от барьеров (barrier-service) и профиля мобильности; доступные барьеры исключаются/штрафуются.
- **MinIO (S3):** хранение фото барьеров (`barrier-photos` bucket). `S3Storage` — `minio-py`; публичный доступ к фото через `MINIO_PUBLIC_URL`/`MINIO_PUBLIC_ENDPOINT`.

## 8. Структура репозитория

```
├── services/
│   ├── auth-service/            Go, гексагональная архитектура
│   │   └── internal/
│   │       ├── domain/{entity,valueobject,service,errors.go}
│   │       ├── ports/{in,out}             # use case-интерфейсы, репозитории
│   │       ├── application/usecase/       # реализация use cases
│   │       └── adapters/{in/http, out/{jwt,postgres}}
│   ├── barrier-service/         Python (src/)
│   ├── moderation-service/      Python (src/)
│   ├── poi-service/             Python (src/)
│   ├── route-service/           Go
│   └── notification-service/    Go
├── frontend/                    React + TS (features/pages/widgets/shared/entities)
├── infra/
│   ├── nginx/nginx.conf         reverse proxy, маршрутизация /api/<svc>
│   └── postgres/init.sql        создание БД и ролей
├── .github/workflows/           CI/CD
├── docker-compose.yml           dev-окружение
├── docker-compose.prod.yml      prod-оверлей
├── generate_keys.go             генерация RSA-пары для JWT
├── .env.example
├── Makefile
├── README.md                    краткая инструкция
└── ARCHITECTURE.md              этот документ
```

Правило зависимостей (гексагональная архитектура для всех сервисов):

```
adapters → application/usecases → domain   (домен ничего не знает об инфраструктуре)
```

### Python-сервис (шаблон)

```
src/
├── main.py                        # wiring, lifespan (инициализация БД, подписки на события)
├── config.py                      # настройки из env
├── domain/
│   ├── entities/                  # сущности и перечисления
│   ├── value_objects/             # Coordinates и т.п.
│   ├── events.py                  # DomainEvent, BarrierCreated/Approved/Rejected/Resolved
│   ├── repositories.py            # порты (интерфейсы репозиториев, EventBus)
│   └── services/                  # бизнес-правила (BarrierService, ModerationService, POIService)
├── application/
│   ├── use_cases/                 # CreateBarrierUseCase, ApproveModerationUseCase, ...
│   └── dto/schemas.py             # Pydantic-схемы
└── adapters/
    ├── inbound/http/              # роутеры FastAPI + security.py (JWT-зависимости)
    └── out/
        ├── persistence/           # модели SQLAlchemy, репозитории
        ├── eventbus/redis_streams.py  # XADD/XREADGROUP
        ├── cache/                 # (зарезервировано)
        └── storage/s3_client.py   # MinIO
```

### Go-сервис (шаблон)

```
cmd/server/main.go                 # wiring, конфиг из env, запуск
internal/
├── domain/                        # entity, valueobject, service, errors
├── ports/in, ports/out            # интерфейсы use cases и выходных адаптеров
├── application/usecase/           # реализация сценариев
└── adapters/
    ├── in/http/                   # Echo-хендлеры, DTO, middleware (JWT)
    └── out/                       # postgres (schema.go + репозитории), eventbus (Redis), jwt, cache
```

## 9. Frontend

- **Стек:** React 18 + TypeScript + Vite, Leaflet (карта), Zustand (сторы), Axios (API).
- **Слои:** `entities` (типы: user, barrier, moderation, route, poi, notification), `features` (сторы/формы по фичам), `widgets` (шапка, маркеры барьеров, панель маршрута), `pages` (Login, Register, Map, BarrierForm, ModerationQueue, Notifications, PoiDetail, RouteBuilder), `shared` (api-клиент, geo-утилиты, UI-кит, тесты).
- **API-клиент** (`shared/api/axios.ts`) добавляет Bearer-токен из стора, обрабатывает refresh и редирект на /login при 401.
- Тесты: vitest + testing-library (`src/shared/test/utils.test.tsx`), запуск `npm test` из `frontend/`.

## 10. API (через nginx)

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/auth/register` | регистрация (email, пароль, роль) |
| POST | `/api/auth/login` | вход → access+refresh |
| POST | `/api/auth/refresh` | обновление access-токена |
| GET  | `/api/auth/validate` | проверка access-токена |
| GET/POST | `/api/barriers` (+`/:id`) | список / создание / получение барьеров |
| POST | `/api/barriers/:id/photo` | загрузка фото (MinIO) |
| POST | `/api/barriers/:id/confirm` · `/complaint` | подтверждения и жалобы |
| GET | `/api/moderation/queue` (+`?status=&limit=&offset=`) | очередь модерации |
| GET | `/api/moderation/queue/:id` | запись модерации |
| POST | `/api/moderation/queue/:id/approve` · `/reject` | решение модератора |
| POST | `/api/routes/build` | построение маршрута (`start/finish` lat/lon, `mobility_profile`) |
| GET/POST | `/api/poi` (+`/:id`) | точки интереса |
| POST | `/api/poi/search/nearby` | поиск POI рядом (body: координаты) |
| PATCH/DELETE | `/api/poi/:id` | обновление / удаление POI |
| GET | `/api/notifications` (+`?limit=&offset=`) | уведомления пользователя |
| GET | `/api/notifications/unread-count` | счётчик непрочитанных |
| POST | `/api/notifications/read` · `/read-all` | отметка прочитанным |
| GET | `/api/notifications/ws?token=...` | WebSocket-поток уведомлений (Upgrade-заголовки) |

Health-эндпоинты сервисов — `/health` только внутри docker-сети (используются healthcheck-ами, в nginx наружу не проксируются). Swagger у Python-сервисов: `/api/<svc>/docs`.

## 11. Запуск и эксплуатация

Полная инструкция — в [README.md](README.md). Ключевые моменты:

```bash
cp .env.example .env                # переменные окружения (пароли БД, TTL токенов и т.д.)
go run generate_keys.go             # RSA-ключи для JWT (перезаписать при необходимости)
docker compose -p accessible-path up -d --build
```

- **302/404 на health:** health-эндпоинты живут только по `/health` внутри сети; наружу ходят только `/api/...` маршруты.
- **502 после рестарта бэкенда:** сбросить upstream-соединения nginx: `docker restart accessible-path-nginx`.
- **Смена кода:** исходники не монтируются в контейнеры — после изменений нужен `docker compose -p accessible-path build <service>`.
- **Токены:** access TTL 15 мин по умолчанию; при длительной работе тестов снова логин. Средняя цепочка проверки: `login.json` → `token.txt`.
- **Логи:** `docker compose logs -f <service>` (контейнеры именуются `accessible-path-<сервис>`).
- **Тесты:** Go — сборка/`go vet` внутри `golang:1.22-alpine` (сеть сборки к proxy.golang.org может не работать в контейнере — модули кэшируются в некоторые образы); Python — pytest; Frontend — `npm test` (vitest). Написать новые тесты можно в любом из сервисов.

## 12. Известные особенности и отличия от исходного ТЗ

- Go-сервисы используют **Echo** вместо net/http (зафиксированное решение по ходу разработки).
- Вместо golang-migrate/alembic — создание схемы на старте (`create_all` / `EnsureSchema`); это делает развёртывание проще, но изменение схемы требует правки `models.py`/`schema.go` и пересборки.
- `docker-compose.prod.yml` — оверлей для прода (TLS, ресурсы); базовые переменные те же, healthcheck через `wget` (доступен в Alpine-образах Go-сервисов и nginx, но **отсутствует** в `python:3.12-slim` — для Python-сервисов healthcheck через `python -c "import httpx; ..."`).