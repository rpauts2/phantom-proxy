# PhantomProxy Pro - Project Structure (2026)

## Recommended Layout

```
phantomproxy/
├── cmd/
│   └── phantom-proxy/     # Go main (proxy + API)
├── internal/              # Go packages (private)
├── pkg/                   # Go packages (importable)
├── api/                   # FastAPI service (future split)
│   └── main.py
├── frontend/              # Next.js 15 (future)
├── configs/
├── migrations/
├── deploy/
├── scripts/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Services

| Service | Stack | Purpose |
|---------|-------|---------|
| phantom-proxy | Go 1.21 + Fiber | AiTM proxy, phishlets, C2, event bus |
| api (future) | FastAPI + Celery | Heavy AI tasks, reporting |
| frontend (future) | Next.js 15 + shadcn | Modern dashboard |

## Deprecated (consolidate)

- `phantomproxy_v*.py` → use `api/main.py` or Go API
- Shell install scripts → `Makefile` + `scripts/install-go-windows.ps1`
