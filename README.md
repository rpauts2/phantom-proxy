# PhantomProxy v14.0

**Продвинутый фишинговый прокси-сервер для Red Team операций и пентеста**

[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-14.0-red.svg)](CHANGELOG.md)

## 🚀 Особенности

### 🔒 Безопасность и Анонимность
- **TLS/SSL терминация** с автоматической генерацией сертификатов
- **HTTP/3 поддержка** для современных браузеров
- **DNS маскировка** и ротация DNS-провайдеров
- **JA3/JA4 спуфинг** для обхода детектирования
- **ГОСТ 28147-89 шифрование** для российских стандартов

### 🎯 Фишинговые Векторы
- **Phishlets** - модули для популярных сервисов (Microsoft 365, Google Workspace, VK, Telegram и др.)
- **Smishing** - SMS фишинг с интеграцией Twilio
- **Vishing** - голосовой фишинг с TTS и ASR
- **Email фишинг** с поддержкой SMTP и шаблонами
- **ClickFix** - автоматическое исправление ссылок в письмах

### 🤖 AI и ML
- **LLM интеграция** (Ollama, OpenAI, Gemini)
- **AI-оркестратор** для автоматизации атак
- **ML Risk Score** - оценка рисков на основе машинного обучения
- **AI-анализ** захваченных данных

### 📊 Мониторинг и Аналитика
- **Grafana Dashboard** для визуализации атак
- **Prometheus** метрики и мониторинг
- **Loki** логирование и анализ
- **Реалтайм статистика** по сессиям и кампаниям

### 🔧 Интеграции
- **C2 Frameworks**: Cobalt Strike, Sliver, Empire
- **Gophish** - интеграция с платформой фишинга
- **Telegram Bot** для уведомлений
- **API** для внешних систем

## 📦 Установка

### Требования
- Go 1.21+
- Docker & Docker Compose
- Node.js 20+ (для frontend)

### Быстрый старт

```bash
# Клонирование репозитория
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy

# Сборка
go build -o phantom-proxy ./cmd/phantom-proxy-v14

# Запуск сервисов
docker-compose up -d

# Запуск основного сервиса
./phantom-proxy --config config.yaml
```

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps
```

## 🎛️ Конфигурация

### Основной конфиг (`config.yaml`)

```yaml
# Сеть
bind_ip: "0.0.0.0"
https_port: 443
domain: "your-domain.com"

# SSL сертификаты
auto_cert: true
cert_path: "./certs/cert.pem"
key_path: "./certs/key.pem"

# API
api_port: 8080
api_key: "your-secure-api-key"

# База данных
database_path: "./phantom.db"

# Модули
modules:
  ai: true
  ml: true
  risk: true
  campaign: true
  email: true
  phishing: true
  smishing: true
  vishing: true
```

### Phishlets

Поддерживаются фишлеты для:
- Microsoft 365 / Office 365
- Google Workspace
- VK, Telegram, Instagram
- Yandex, Mail.ru
- Сбербанк, Тинькофф
- Wildberries, Ozon

## 🚀 Использование

### Запуск с веб-панелью

```bash
# Запуск frontend
cd frontend
npm install
npm run dev

# Веб-панель доступна на http://localhost:3000
```

### API Примеры

```bash
# Health check
curl http://localhost:8080/health

# Статистика
curl -H "Authorization: Bearer your-api-key" \
  http://localhost:8080/api/v1/stats

# Включение phishlet
curl -X POST http://localhost:8080/api/v1/phishlets/microsoft_365/enable \
  -H "Authorization: Bearer your-api-key"
```

### Docker Compose сервисы

| Сервис | Порт | URL |
|--------|------|-----|
| PhantomProxy | 443 | https://your-domain.com |
| Grafana | 3001 | http://localhost:3001 |
| Prometheus | 9090 | http://localhost:9090 |
| Loki | 3100 | http://localhost:3100 |

## 📊 Веб-панель

### Dashboard (Grafana)
- Мониторинг активности системы
- Статистика по сессиям и атакам
- Анализ рисков и угроз
- Просмотр логов в реальном времени

### API Endpoints
- `/api/v1/stats` - Статистика системы
- `/api/v1/sessions` - Список сессий
- `/api/v1/phishlets` - Управление phishlets
- `/api/v1/campaigns` - Управление кампаниями

## 🔧 Разработка

### Сборка

```bash
# Сборка основного бинарника
go build -o phantom-proxy ./cmd/phantom-proxy-v14

# Сборка с тегами
go build -tags=debug -o phantom-proxy ./cmd/phantom-proxy-v14
```

### Тестирование

```bash
# Запуск тестов
go test ./...

# Запуск конкретного пакета
go test ./internal/proxy/...
```

### Frontend разработка

```bash
cd frontend
npm install
npm run dev
```

## 📚 Документация

- [Полное руководство](FULL_MANUAL.md)
- [API Документация](docs/API.md)
- [Phishlets](configs/phishlets/README.md)
- [Docker Compose](docker-compose.yml)

## ⚠️ Предупреждение

**Этот инструмент предназначен ТОЛЬКО для:**
- Авторизованного тестирования безопасности
- Red Team операций с письменным разрешением
- Обучения и исследований в контролируемой среде

**Запрещено использование:**
- Для несанкционированного доступа к системам
- В целях мошенничества или кражи данных
- Без письменного разрешения владельца систем

## 🤝 Вклад

Приветствуются pull requests! Перед отправкой убедитесь, что:
1. Код проходит все тесты
2. Добавлена документация для новых функций
3. Соблюдены стандарты кодирования

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE) файл

## 🙏 Благодарности

- [Go Team](https://golang.org/) за отличный язык
- [Grafana Labs](https://grafana.com/) за мониторинг
- [Ollama](https://ollama.ai/) за LLM интеграцию

---

**⚠️ Используйте ответственно и только в законных целях!**