# PhantomProxy — Исправления и улучшения

## Выполнено

### 1. Go API — отсутствующие обработчики
- **generatePhishlet** — вызывает AI оркестратор (`ai.GeneratePhishlet`)
- **analyzeSite** — вызывает `ai.AnalyzeSite`
- **registerDomain**, **rotateDomain**, **listDomains** — заглушки (ожидают подключения domain.Rotator)

### 2. deletePhishlet
- Добавлен `database.DeletePhishlet(id)` — устанавливает `is_active = FALSE`
- API-обработчик вызывает метод БД

### 3. ListPhishlets
- Фильтрация только по активным: `WHERE is_active = TRUE`

### 4. Celery Redis
- Broker/backend берутся из `REDIS_URL` (для Docker: `redis://redis:6379/0`)

### 5. Frontend
- Добавлен `frontend/public/robots.txt` — чтобы Dockerfile копировал папку `public`
- Dockerfile корректно собирает Next.js

### 6. Удалено
- `api/main.py` — дублировал точку входа, используется `api/app/main.py`

### 7. API run
- `api/run.py` — скрипт запуска FastAPI

## Команды сборки

```bash
# Go (из корня проекта)
go build -o phantom-proxy.exe ./cmd/phantom-proxy

# Тесты
go test ./...

# Python API
cd api && pip install -r requirements.txt && python -c "from app.main import app; print('OK')"
```

## Примечание по domain-пакету

`internal/domain/rotator.go` использует `github.com/go-acme/lego` — пакет не добавлен в go.mod, так как domain не импортируется в основном бинарнике. Для использования domain-ротации выполните:
```bash
go get github.com/go-acme/lego/v4
```
