# 👻 PHANTOMPROXY v14.0

**Enterprise Red Team Simulation Platform**

[![Version](https://img.shields.io/badge/version-14.0.0-blue)]()
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.21+-blue)]()
[![Python](https://img.shields.io/badge/python-3.11+-blue)]()
[![CI/CD](https://github.com/rpauts2/phantom-proxy/actions/workflows/ci-cd.yml/badge.svg)]()

---

## ⚡ QUICK START

### Linux (1 command):
```bash
sudo ./install.sh
```

### Windows (1 command):
```powershell
.\install.ps1
```

### Docker:
```bash
docker-compose up -d
```

---

## 🎯 FEATURES

### Core
- ✅ **AiTM Reverse Proxy** - HTTP/HTTPS/HTTP3 with TLS 1.3
- ✅ **Session Management** - Redis-backed with cookie capture
- ✅ **Phishlet Engine** - 10+ pre-configured templates
- ✅ **2FA/MFA Bypass** - Token interception and replay

### AI/ML
- ✅ **LangGraph Agents** - 4 autonomous AI agents
- ✅ **RAG System** - ChromaDB vector store
- ✅ **Smart Scoring** - 8-factor behavioral analysis
- ✅ **Auto-Phishlet** - AI-generated phishing templates

### Enterprise
- ✅ **Multi-Tenant** - Full isolation with quotas
- ✅ **Zero-Trust mTLS** - Client certificate authentication
- ✅ **Auth Integration** - Keycloak/Zitadel support
- ✅ **FSTEC Compliance** - GOST encryption for logs

### Attack Simulation
- ✅ **Vishing** - Voice phishing with Twilio/SMS.ru
- ✅ **Smishing** - SMS phishing campaigns
- ✅ **C2 Integration** - Sliver, Empire, Cobalt Strike

---

## 📦 INSTALLATION

### Automated Install:

**Linux:**
```bash
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy
sudo ./install.sh
```

**Windows:**
```powershell
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy
.\install.ps1
```

### Manual Install:

**1. Dependencies:**
```bash
# Go 1.21+
go mod download

# Python 3.11+
pip install -r requirements.txt

# Node.js 20+
cd frontend && npm install
```

**2. Generate Certificates:**
```bash
mkdir -p certs
openssl req -x509 -newkey rsa:4096 \
  -keyout certs/key.pem -out certs/cert.pem \
  -days 365 -nodes
```

**3. Build:**
```bash
go build -o phantom-proxy ./cmd/phantom-proxy-v14
```

**4. Run:**
```bash
./phantom-proxy --config config.yaml
```

---

## 🚀 USAGE

### Console UI:
```bash
python console.py
```

**Commands:**
```
help          - Show help
status        - System status
dashboard     - Main dashboard
sessions      - Active sessions
phishlets     - Loaded phishlets
logs          - System logs
quit          - Exit
```

### Web Interface:
- **Frontend:** http://localhost:3000
- **API:** http://localhost:8080
- **Proxy:** https://localhost:8443

### Docker:
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Makefile Commands:
```bash
make help         # Show all commands
make build        # Build binary
make test         # Run tests
make docker       # Start Docker
make health       # Health check
make backup       # Create backup
```

---

## 📁 PROJECT STRUCTURE

```
phantom-proxy/
├── cmd/phantom-proxy-v14/    # Go entry point
├── core/proxy/               # AiTM engine
├── internal/                 # Services
│   ├── tenant/              # Multi-tenant
│   ├── risk/                # Risk scoring
│   ├── vishing/             # Voice phishing
│   ├── c2/                  # C2 integration
│   ├── mtls/                # Zero-trust
│   ├── auth/                # Authentication
│   └── fstec/                # FSTEC compliance
├── ai_service/              # AI service (LangGraph)
├── api/                     # FastAPI backend
├── frontend/                # Next.js dashboard
├── configs/phishlets/       # Phishlet configs
├── deploy/                  # DevOps configs
├── install.sh               # Linux installer
├── install.ps1              # Windows installer
└── docker-compose.yml       # Docker stack
```

---

## 📊 API ENDPOINTS

### Core (59 endpoints):
- `GET /health` - Health check
- `GET /api/v1/stats` - Statistics
- `GET/POST /api/v1/sessions` - Session management
- `GET/POST /api/v1/phishlets` - Phishlet management
- `POST /api/v1/risk/events` - Risk scoring
- `GET/POST /api/v1/fstec/audit` - Audit logs

### AI Service:
- `POST /v1/generate/email` - Generate phishing email
- `POST /v1/generate/phishlet` - Generate phishlet
- `POST /v1/rag/search` - RAG search
- `POST /api/v1/agents/run-campaign` - AI campaign

---

## 🧪 TESTING

```bash
# All tests
make test

# Go tests
make test-go

# Python tests
make test-python

# Health check
make health

# Coverage
go test -v -cover ./...
```

---

## 🔧 MAINTENANCE

### Backup:
```bash
# Create backup
make backup

# Or manually
python backup.py backup

# List backups
python backup.py list

# Restore
python backup.py restore backup_20260220.tar.gz
```

### Health Check:
```bash
# Script
./healthcheck.sh

# Or make
make health
```

### Update:
```bash
git pull
make build
sudo systemctl restart phantomproxy
```

---

## 📖 DOCUMENTATION

- **[API Documentation](docs/API.md)** - Full API reference
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment
- **[Phishlet Guide](docs/PHISHLETS.md)** - Creating phishlets
- **[Architecture](docs/ARCHITECTURE.md)** - System architecture
- **[Security](SECURITY.md)** - Security policy
- **[Contributing](CONTRIBUTING.md)** - Contribution guidelines

---

## ⚠️ LEGAL DISCLAIMER

This software is for **authorized security testing only**.

**Required:**
- ✅ Written permission from system owners
- ✅ Signed Rules of Engagement (RoE)
- ✅ Proper legal authorization

**Prohibited:**
- ❌ Unauthorized access
- ❌ Credential theft
- ❌ Fraud or identity theft

By using this software, you agree to comply with all applicable laws.

---

## 🏆 VERSIONS

| Version | Status | Features |
|---------|--------|----------|
| v14.0.0 | ✅ Current | Full enterprise stack |
| v13.0.0 | ⚠️ Legacy | Basic multi-tenant |
| v12.0.0 | ❌ EOL | Initial release |

---

## 📞 SUPPORT

- **GitHub Issues:** https://github.com/rpauts2/phantom-proxy/issues
- **Security:** security@phantomseclabs.com
- **Documentation:** https://docs.phantomproxy.io

---

## 📄 LICENSE

MIT License - see [LICENSE](LICENSE) for details.

---

**© 2026 PhantomSec Labs. All rights reserved.**
