# Деплой «Доступного пути» на Beget VPS (1 CPU / 2 GB RAM)

Продакшн-стек использует **управляемый PostgreSQL и S3 от Beget**, на VPS остаются только контейнеры приложений + локальный Redis. Наружу открыты только порты 80/443 (nginx).

```
Интернет ──► nginx (80/443, HTTPS, Let's Encrypt)
               ├── /api/auth/...        -> auth-service
               ├── /api/barriers/...    -> barrier-service
               ├── /api/moderation/...  -> moderation-service
               ├── /api/routes/...      -> route-service
               ├── /api/poi/...         -> poi-service
               ├── /api/notifications/... + /ws -> notification-service
               └── /                    -> frontend (SPA)
Все сервисы  ──► Redis (локальный контейнер, стримы событий)
Все сервисы  ──► Managed PostgreSQL (внешний, 6 БД)
barrier-service ──► Managed S3 (фото барьеров)
```

## 0. Что понадобится

- VPS Beget: 1 CPU / 2 GB, публичный IP (ниже `1.2.3.4`)
- Управляемый PostgreSQL от Beget (6 БД)
- S3-совместимое хранилище Beget (1 бакет)
- Домен (или поддомен), например `accessible.my.ru`
- SSH-доступ к VPS

## 1. Панель Beget

### DNS
Добавьте A-запись: `accessible.my.ru → 1.2.3.4` (IP вашего VPS).

### Управляемый PostgreSQL
1. Создайте сервер БД в панели Beget («БД → PostgreSQL»), дождитесь готовности.
2. Откройте SQL-консоль сервера и выполните скрипт [`deploy/beget/databases.sql`](deploy/beget/databases.sql) — он создаст 6 БД и 6 пользователей. **Замените пароли** на свои (те же, что пропишете в `.env.prod`).
3. В настройках сервера БД откройте доступ для IP вашего VPS (whitelist).
4. Запишите хост сервера (вида `pg1234567-beget.app`) и порт.
5. Если панель требует SSL — ставьте в `.env.prod` `POSTGRES_SSLMODE=require`.

> Таблицы создадут сами сервисы при первом старте (`EnsureSchema` / `metadata.create_all`). PostGIS не требуется.

### S3
1. Создайте бакет `barrier-photos`.
2. Сгенерируйте access/secret ключи в панели S3.
3. Хост обычно `s3.ru1.storage.beget.cloud`, регион `ru-1` — уточните в панели.

## 2. Подготовка VPS (один раз)

```bash
ssh root@1.2.3.4
apt update && apt install -y git
git clone <ваш репозиторий> /opt/accessible-path
cd /opt/accessible-path
sudo ./deploy/deploy.sh setup   # ставит docker и своп 2G
```

Скрипт `setup` добавляет swap-файл 2G (защита от OOM при сборке образов на 2GB ОЗУ) и ставит Docker.

## 3. Конфигурация

```bash
cd /opt/accessible-path
cp .env.prod.example .env.prod
nano .env.prod
```

Заполните обязательные поля:
- `DOMAIN`, `CERTBOT_EMAIL` — домен и почта для Let's Encrypt (можно оставить пустым, HTTPS появится позже)
- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_SSLMODE` + пароли 6 пользователей
- `REDIS_PASSWORD` — придумайте сильный пароль
- `MINIO_*`, `S3_REGION`, `S3_ADDRESSING_STYLE` — доступы к S3
- `SEED_MODERATOR_EMAIL/PASSWORD` — модератор создастся автоматически при первом старте auth-service

JWT-ключи (RSA) скрипт сгенерирует сам в `services/auth-service/keys/`. **Не коммитьте этот каталог.**

## 4. Запуск

```bash
./deploy/deploy.sh up          # сборка образов (10-30 мин на 1 ядре) + старт
./deploy/deploy.sh status      # проверить статусы
```

Первая сборка долгая: 3 Go-сервиса, 3 Python-сервиса, фронтенд. При обновлении репозитория:

```bash
git pull
./deploy/deploy.sh restart
```

## 5. HTTPS (Let's Encrypt)

```bash
./deploy/deploy.sh ssl-init    # выпуск сертификата (домен должен смотреть на VPS)
```

Сертификаты хранятся в `infra/nginx/ssl/` (вне git), автообновление — cron:

```bash
crontab -e
# раз в месяц
0 3 1 * * cd /opt/accessible-path && ./deploy/deploy.sh ssl-renew >> /var/log/certbot-renew.log 2>&1
```

Если сначала запустить без `DOMAIN`, nginx всё равно поднимется (443 с временным ключом) — после `ssl-init` произойдёт reload.

## 6. Проверка после деплоя

```bash
curl -s https://accessible.my.ru/health            # healthy
curl -s https://accessible.my.ru/api/barriers       # JSON барьеров
# регистрация пользователя в UI, создание барьера, одобрение модератором
# -> уведомление barrier_approved, при барьере на сохранённом маршруте -> route_updated
```

## 7. Операции

| Действие | Команда |
|---|---|
| Логи | `./deploy/deploy.sh logs` (или `logs route-service`) |
| Стоп | `./deploy/deploy.sh down` |
| Пересборка после `git pull` | `./deploy/deploy.sh restart` |
| Статусы | `./deploy/deploy.sh status` |
| БД-бэкап (внешняя PG, один сервис) | `pg_dump "postgres://<user>:<pass>@<host>/<db>" > backup.sql` |
| Бэкап S3 | экспорт бакета инструментами Beget |

Память (лимиты в `docker-compose.prod.yml`, сумма ≈ 1.7 ГБ): auth 128M, barrier 384M, moderation 384M, route 256M, poi 256M, notification 128M, redis 256M, frontend/nginx 64M. На 2 GB ОЗУ работает с запасом; при переполнении Docker просто перезапустит контейнер (режим `unless-stopped`).

## 8. Траблшутинг

- **Сервис не стартует, в логах `database ... refused`** — проверьте `POSTGRES_HOST`/порт и whitelist IP в панели Beget; если панель требует TLS — `POSTGRES_SSLMODE=require`.
- **`barrier-service` падает в логах с `head_bucket`/`create_bucket`** — создайте бакет вручную в панели S3 (сервис теперь не падает, но фото загружаться не будут, пока бакет не существует).
- **Фото открываются как «битая подпись»** — проверьте `MINIO_PUBLIC_URL` (публичный адрес S3) и `S3_ADDRESSING_STYLE=path`.
- **Долгий старт route-service** — это импорт OSM-графа (Overpass); healthcheck рассчитан на 10 минут, дальше он поднимается по кэшу из БД.
- **nginx не поднимается** — `docker compose --env-file .env.prod -f docker-compose.prod.yml ps`; сервисы ждут внешнюю БД и сами перезапускаются.

## 9. Обновление стека

```bash
cd /opt/accessible-path && git pull && ./deploy/deploy.sh restart
```

Схема БД создаётся идемпотентно (`CREATE TABLE IF NOT EXISTS` / `ALTER ... ADD COLUMN IF NOT EXISTS`) — обновление безопасно.
