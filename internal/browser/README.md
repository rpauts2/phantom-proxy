# 🌐 BROWSER POOL MODULE

Пул браузеров для эмуляции человеческого поведения

---

## 📋 ОПИСАНИЕ

Модуль Browser Pool который:

1. **Управляет пулом браузеров** — Playwright Chromium
2. **Эмулирует человека** — движения мыши, клики, скроллинг
3. **Скрывает автоматизацию** — stealth скрипты
4. **Балансирует нагрузку** — round-robin между браузерами

---

## 🚀 ВОЗМОЖНОСТИ

### ✅ Browser Management

- **Автомасштабирование** — от Min до Max браузеров
- **Health monitoring** — проверка активности
- **Auto-restart** — перезапуск при таймауте

### ✅ Human Emulation

- **Движения мыши** — случайные траектории
- **Клики** — рандомные координаты
- **Скроллинг** — имитация чтения
- **Random delays** — человеческие паузы

### ✅ Stealth

- **webdriver: false** — скрытие автоматизации
- **Random fingerprints** — UA, viewport, timezone
- **Anti-detection** — обход детектов ботов

---

## 📡 API ENDPOINTS

### POST /api/v1/browser/execute

Выполнение запроса через браузер.

**Request:**
```json
{
  "url": "https://login.microsoftonline.com",
  "method": "GET",
  "headers": {},
  "body": "",
  "screenshot": false
}
```

**Response:**
```json
{
  "success": true,
  "browser_id": "browser-0",
  "status": 200,
  "body": "...",
  "screenshot": "base64...",
  "execution_time_ms": 1234
}
```

### GET /api/v1/browser/stats

Статистика пула.

**Response:**
```json
{
  "total_browsers": 5,
  "active_browsers": 5,
  "total_requests": 100,
  "min_browsers": 2,
  "max_browsers": 10
}
```

### POST /api/v1/browser/screenshot/:id

Скриншот страницы.

---

## ⚙️ КОНФИГУРАЦИЯ

### config.yaml

```yaml
# Browser Pool
browser_pool:
  enabled: true
  
  # Размер пула
  min_browsers: 2
  max_browsers: 10
  
  # Таймауты
  browser_timeout: 1800  # секунд (30 минут)
  page_timeout: 60       # секунд
  
  # Поведение
  humanize_actions: true
  random_delays: true
  
  # Fingerprints
  random_user_agent: true
  random_viewport: true
  random_timezone: true
  
  # Playwright
  headless: true  # false для отладки
```

---

## 🔗 АРХИТЕКТУРА

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  PhantomProxy   │────▶│   Browser Pool   │────▶│   Playwright    │
│    (Go API)     │     │   (Load Balance) │     │   (Chromium)    │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌──────────────────┐
                        │  Stealth Scripts │
                        │  (Anti-detect)   │
                        └──────────────────┘
                               │
                               ▼
                        ┌──────────────────┐
                        │  Human Behavior  │
                        │  (Mouse, Clicks) │
                        └──────────────────┘
```

---

## 💡 ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ

### Пример 1: Выполнение запроса

```bash
curl -X POST http://localhost:8080/api/v1/browser/execute \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://login.microsoftonline.com",
    "method": "GET"
  }'
```

### Пример 2: Получение статистики

```bash
curl http://localhost:8080/api/v1/browser/stats \
  -H "Authorization: Bearer secret"
```

### Пример 3: Скриншот

```bash
curl -X POST http://localhost:8080/api/v1/browser/screenshot/browser-0 \
  -H "Authorization: Bearer secret"
```

---

## 🛡️ БЕЗОПАСНОСТЬ

### Anti-Detection

- **Stealth скрипты** — скрытие webdriver
- **Random fingerprints** — каждый браузер уникален
- **Human behavior** — эмуляция человека

### Resource Limits

- **Max browsers** — ограничение на количество
- **Timeouts** — авто-закрытие неактивных
- **Memory limits** — контроль потребления

---

## 📈 МОНИТОРИНГ

### Метрики

- Количество браузеров
- Запросы в секунду
- Среднее время выполнения
- Успешность/ошибки

### Логи

```bash
tail -f /var/log/phantom/browser-pool.log
```

---

## 🐛 TROUBLESHOOTING

### Ошибка: "No available browsers"

**Решение:** Увеличить `max_browsers` или уменьшить нагрузку.

### Ошибка: "Navigation failed"

**Решение:** Проверить доступность URL и увеличить `page_timeout`.

### Ошибка: "Playwright installation failed"

**Решение:** Запустить `playwright install` вручную.

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

1. **Distributed pool** — браузеры на разных серверах
2. **Mobile emulation** — iOS/Android браузеры
3. **Session persistence** — сохранение сессий между запросами
4. **Advanced fingerprints** — canvas, WebGL, fonts

---

**Версия:** 1.0.0  
**Автор:** PhantomProxy Team  
**Лицензия:** MIT
