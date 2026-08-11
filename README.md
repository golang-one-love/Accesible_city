# Доступный путь — Навигатор для маломобильных граждан

Микросервисная система построения доступных маршрутов в обход физических барьеров (высокие бордюры, сломанные лифты, перекрытые тротуары). Пользователи отмечают барьеры на карте, модераторы проверяют их, а система строит маршрут с учётом доступности и профиля мобильности.

## Стек

| Слой | Технологии |
|------|------------|
| Frontend | React + TypeScript + Vite, Leaflet |
| Backend | Go 1.22 (auth, route, notification) + Python 3.12 / FastAPI (barrier, moderation, poi) |
| Хранилища | PostgreSQL 16 (отдельная БД на сервис), Redis 7 (Streams), MinIO (S3, фото) |
| Инфраструктура | Docker Compose, Nginx (reverse proxy) |

## Быстрый старт

Все команды выполняются в **PowerShell / bash**. Нужен только Docker (Go, Node и `make` на машине не требуются — ключи и образы готовятся контейнерами).

```powershell
# 1. Переменные окружения (скопировать из примера)
Copy-Item .env.example .env

# 2. Сгенерировать ключи JWT (RS256) — Docker вместо локального Go
docker run --rm -v "${PWD}:/src" -w /src golang:1.22-alpine go run generate_keys.go

# 3. Поднять весь стек
docker compose -p accessible-path up -d --build

# фронтенд: http://localhost
# MinIO console: http://localhost:9001
# Swagger (Python-сервисы): http://localhost/api/{barriers|moderation|poi}/docs
```

> Приложение запускается из меню Docker Desktop (или `docker compose up -d`). Убедитесь, что Docker Desktop запущен заранее.

> **Windows:** проект «-p accessible-path» обязателен, если путь к репозиторию содержит кириллицу (`D:\Андрей\...`) — иначе имя проекта по умолчанию генерируется из пути и расходится с `docker compose ps`/`docker logs`.

`Makefile` и команды `make up`/`make logs` и т.п. — опциональны, требуют GNU Make (доступен в Git Bash/Unix). Эквиваленты: `docker compose -p accessible-path ps`, `docker compose -p accessible-path logs -f <service>`.

## Первый запуск / полезные операции

- **Схема БД** создаётся автоматически при старте сервиса (`SQLAlchemy create_all` в Python-сервисах, `EnsureSchema` в Go-сервисах). `infra/postgres/init.sql` создаёт только БД и роли. Отдельные миграции не нужны.
- **Перезапуск бэкенд-контейнера:** после `docker compose ... restart <svc>` nginx может отдавать 502 (stale keepalive-соединения). Решение: `docker restart accessible-path-nginx`.
- **Ключи JWT:** только auth-service держит приватный ключ; остальные сервисы проверяют подпись публичным ключом, смонтированным в `/keys` (см. `.env` → `JWT_PUBLIC_KEY_PATH`). Генерация: команда из «Быстрого старта» (шаг 2).

## API

Внешние пути идут через nginx с префиксом `/api/<сервис>/...` и проксируются на `/api/v1/<сервис>/...` внутри docker-сети:

| Путь | Сервис |
|------|--------|
| `/api/auth/*` | регистрация, логин, refresh, validate |
| `/api/barriers/*` | барьеры, фото, подтверждения, жалобы |
| `/api/moderation/*` | очередь модерации, approve/reject |
| `/api/routes/*` | построение маршрута |
| `/api/poi/*` | точки интереса (POI) |
| `/api/notifications/*` | уведомления, + WebSocket `/api/notifications/ws?token=...` |

## Документация

- **[ARCHITECTURE.md](ARCHITECTURE.md)** — подробно: микросервисы, доменные сущности, событийная модель, потоки данных, зависимости и эксплуатационные заметки.

## Лицензия

MIT