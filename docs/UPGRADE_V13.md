# PhantomProxy v13 — Upgrade & Architecture

## Что сделано

### 1. Go Installer
- `scripts/install-go-windows.ps1` — установка Go на Windows (winget или MSI)

### 2. Docker & Makefile
- `Dockerfile` — multi-stage build Go binary
- `docker-compose.yml` — proxy + PostgreSQL + Redis + Prometheus + Grafana
- `Makefile` — `make build`, `make run`, `make docker`, `make install`

### 3. Observability
- `/metrics` — Prometheus-совместимые метрики
- `deploy/prometheus.yml` — конфиг Prometheus
- Grafana на порту 3001

### 4. PostgreSQL
- `migrations/001_init.sql` — схема для будущей миграции
- TimescaleDB 16 в docker-compose

### 5. Unified API
- `api/main.py` — точка входа для Python API (замена разрозненных скриптов)

## Дальнейшие шаги (приоритеты)

1. **PostgreSQL driver** — добавить `jackc/pgx` и переключение SQLite/Postgres
2. **Redis** — кеш сессий, очереди событий
3. **Next.js frontend** — `npx create-next-app@latest frontend`
4. **Split API** — вынести тяжёлые AI-задачи в Celery worker
