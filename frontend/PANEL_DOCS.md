# 🎛️ PhantomProxy Control Panel

## 📖 Обзор

Веб-панель управления PhantomProxy v14.0 - централизованный интерфейс для управления всеми компонентами системы.

---

## 🚀 Быстрый старт

### 1. Запуск панели

```bash
# Запустите PhantomProxy
./phantom-proxy.exe --config config.yaml

# Откройте браузер
http://localhost:8080
```

### 2. Логин

Панель использует API Key аутентификацию:
- **API Key по умолчанию**: `change-me-to-secure-random-string`
- Измените в настройках панели!

---

## 📱 Разделы панели

### 1. Dashboard 📊

**Статистика в реальном времени:**
- Активные phishlets
- Активные сессии
- Перехваченные credentials
- Запросов сегодня

**Статус сервисов:**
- Proxy Server
- API Server
- Database
- Redis
- Event Bus

**График активности:**
- Requests per minute
- Real-time обновления

---

### 2. Phishlets 🔐

**Управление phishlets:**
- ✅ Просмотр всех доступных phishlets
- ✅ Включение/выключение одним кликом
- ✅ Просмотр деталей phishlet
- ✅ Статус каждого phishlet

**Доступные phishlets:**
| Phishlet | Статус |
|----------|--------|
| Microsoft 365 | ✅ Active |
| Google Workspace | ✅ Active |
| Yandex | ✅ Active |
| Mail.ru | ✅ Active |
| SberBank | ✅ Active |
| VK | ✅ Active |
| Ozon | ✅ Active |
| Wildberries | ✅ Active |
| TikTok | ✅ Active |
| Instagram | ✅ Active |
| Facebook | ✅ Active |
| Telegram | ✅ Active |

---

### 3. Сессии 👥

**Список активных сессий:**
- ID сессии
- IP жертвы
- Целевой хост
- Время создания
- Статус

**Действия:**
- 🗑️ Удалить сессию
- 👁️ Просмотр деталей
- 📋 Копировать ID

---

### 4. Credentials 🔑

**Перехваченные данные:**
- Username/Email
- Password
- Целевой сервис
- Время перехвата

**Действия:**
- 📋 Копировать credentials
- 📥 Экспорт в CSV/JSON
- 🔍 Поиск по credentials

---

### 5. Логи 📝

**Live логи в реальном времени:**
- Запросы к прокси
- Перехваченные credentials
- Создание сессий
- Ошибки системы

**Функции:**
- ⏸️ Пауза/Продолжить
- 🗑️ Очистка логов
- 🎨 Цветовая кодировка:
  - 🔵 Info
  - 🟢 Success
  - 🟡 Warning
  - 🔴 Error

---

### 6. Настройки ⚙️

**Основные настройки:**
- API Key
- HTTPS Port
- API Port
- Domain
- Debug Mode

**Сохранение:**
- Настройки сохраняются в localStorage
- Применяются немедленно

---

## 🎨 Особенности UI

### Тёмная тема
- Снижает нагрузку на глаза
- Профессиональный вид
- Modern design

### Адаптивность
- Работает на desktop
- Работает на tablet
- Работает на mobile

### Real-time обновления
- SSE (Server-Sent Events)
- Автообновление каждые 5 секунд
- Мгновенные уведомления

---

## 🔌 API Integration

### SSE Events

Панель подключается к `/api/v1/events` и получает:

```javascript
{
  "type": "credential.captured",
  "session_id": "abc123",
  "username": "user@example.com",
  "timestamp": "2026-02-22T12:00:00Z"
}
```

### REST API

Все действия выполняются через REST API:

```bash
# Получить список phishlets
GET /api/v1/phishlets

# Включить phishlet
POST /api/v1/phishlets/microsoft365/enable

# Получить сессии
GET /api/v1/sessions

# Получить credentials
GET /api/v1/credentials
```

---

## 🛠️ Кастомизация

### Изменение темы

Откройте `frontend/js/panel.js` и измените CSS variables:

```css
:root {
    --primary-color: #6366f1;
    --secondary-color: #8b5cf6;
    --success-color: #10b981;
    --danger-color: #ef4444;
}
```

### Добавление виджетов

Добавьте новый виджет в Dashboard:

```html
<div class="col-md-3">
    <div class="card stat-card">
        <i class="bi bi-new-icon display-4 mb-3"></i>
        <div class="stat-number" id="newStat">0</div>
        <div class="text-muted">New Stat</div>
    </div>
</div>
```

---

## 📊 Статистика

### Dashboard Metrics

| Метрика | Описание | Обновление |
|---------|----------|------------|
| Active Phishlets | Количество активных phishlets | 5 сек |
| Active Sessions | Количество активных сессий | 5 сек |
| Captured Credentials | Всего перехвачено credentials | 5 сек |
| Requests Today | Запросов сегодня | 5 сек |

### Activity Chart

- Показывает requests per minute
- Последние 20 точек
- Автообновление

---

## 🔐 Безопасность

### API Key

- Хранится в localStorage
- Используется для всех запросов
- Измените default key!

### Рекомендации

1. Смените API Key по умолчанию
2. Используйте HTTPS
3. Ограничьте доступ по IP
4. Включите 2FA для панели

---

## 🐛 Troubleshooting

### Панель не загружается

```bash
# Проверьте что PhantomProxy запущен
./phantom-proxy.exe --version

# Проверьте порт API
curl http://localhost:8080/health
```

### Нет данных в панели

```bash
# Проверьте API Key в настройках
# Проверьте логи PhantomProxy
tail -f logs/phantom.log
```

### SSE не подключается

```bash
# Проверьте endpoint
curl http://localhost:8080/api/v1/events
```

---

## 📱 Mobile Version

Панель адаптирована для мобильных устройств:

- ✅ Responsive design
- ✅ Touch-friendly кнопки
- ✅ Mobile navigation
- ✅ Optimized charts

---

## 🎯 Горячие клавиши

| Клавиша | Действие |
|---------|----------|
| `R` | Refresh all |
| `D` | Dashboard |
| `P` | Phishlets |
| `S` | Sessions |
| `C` | Credentials |
| `L` | Logs |
| `N` | Settings |

---

## 📞 Поддержка

- **Документация**: ./docs/
- **API Docs**: http://localhost:8080/api/docs
- **Issues**: https://github.com/phantom-proxy/phantom-proxy/issues

---

**Версия**: 1.0.0  
**Последнее обновление**: Февраль 2026
