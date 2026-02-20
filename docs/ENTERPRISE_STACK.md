# PhantomProxy v13 — Enterprise Killer Stack

## Архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                        PhantomProxy v13                          │
├─────────────────────────────────────────────────────────────────┤
│  Frontend (Next.js 15)  │  API (FastAPI)  │  Go Proxy (AiTM)     │
│  TanStack Query         │  Celery + Redis │  Fiber + TLS         │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL 16 + TimescaleDB  │  Redis 7  │  OpenTelemetry       │
└─────────────────────────────────────────────────────────────────┘
```

## Почему это сильнее конкурентов (2026)

| Критерий | PhantomProxy | Hoxhunt / KnowBe4 |
|----------|--------------|-------------------|
| **Скорость** | Go proxy, <5ms задержка | Часто облачный SaaS, выше латентность |
| **Self-hosted** | 100% on-prem | Ограничено или невозможно |
| **ФСТЭК** | Roadmap + GOST | Нет |
| **Цена** | 5–10× ниже | Высокая подписка |
| **AI** | Llama/LangGraph локально | AIDA в облаке |

## Запуск

```bash
# Минимум (proxy + API)
docker-compose up phantom-proxy postgres redis -d

# Полный стек
docker-compose up -d

# Helm (Kubernetes)
helm install phantomproxy ./helm/phantomproxy
```

## Порты

| Сервис | Порт |
|--------|------|
| Proxy HTTPS | 443 |
| Go API | 8080 |
| FastAPI | 8000 |
| Frontend | 3000 |
| Grafana | 3001 |
| Prometheus | 9090 |
| PostgreSQL | 5432 |
| Redis | 6379 |
