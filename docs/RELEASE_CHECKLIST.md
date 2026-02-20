# PhantomProxy v13 — Release Checklist

## Готовый продукт

### Перед запуском

1. **Сертификаты**: `certs/cert.pem` и `certs/key.pem` (или `go run ./cmd/gendert`)
2. **config.yaml**: скопировать из `config.example.yaml`, задать `api_key`
3. **Frontend**: `NEXT_PUBLIC_API_KEY` = тот же `api_key` из config

### Команды запуска

| Режим | Команда |
|-------|---------|
| Go binary | `go build ./cmd/phantom-proxy && ./phantom-proxy --config config.yaml` |
| Docker minimal | `docker-compose -f docker-compose.minimal.yml up -d` |
| Docker full | `docker-compose up -d` |
| Frontend | `cd frontend && npm install && npm run dev` |

### Порты

- 443 — HTTPS proxy
- 8080 — Go API
- 3000 — Frontend (dev)
- 8000 — FastAPI (если включён)
- 9090 — Prometheus
- 3001 — Grafana
