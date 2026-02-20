# 🚀 PHANTOMPROXY - БЫСТРАЯ УСТАНОВКА

**ОДНА КОМАНДА для установки:**

```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147 "bash ~/install.sh"
```

**Или по шагам:**

```bash
# 1. Подключись
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147

# 2. Запусти установку
bash install.sh

# 3. Проверь
curl http://localhost:8080/health
curl -k https://localhost:8443/
```

---

## ✅ ОЖИДАЕМЫЙ РЕЗУЛЬТАТ

```
============================================================
PHANTOMPROXY v1.7.0 - ULTIMATE AUTO INSTALL
============================================================

[1/5] Очистка...
✅ Очищено

[2/5] Установка Python сервисов...
✅ Python сервисы созданы

[3/5] Генерация SSL...
✅ SSL готов

[4/5] Запуск сервисов...
  🚀 API Server
  🚀 HTTPS Proxy
✅ Сервисы запущены

[5/5] Тестирование...

✅ API Server (8080)
✅ API Stats
✅ HTTPS Proxy (8443)

============================================================
🎉 УСТАНОВКА ЗАВЕРШЕНА УСПЕШНО!
============================================================

PhantomProxy v1.7.0 работает:
  API:       http://212.233.93.147:8080
  HTTPS:     https://212.233.93.147:8443
```

---

## 📊 ЧТО УСТАНОВЛЕНО

| Сервис | Порт | Статус |
|--------|------|--------|
| API Server | 8080 | ✅ |
| HTTPS Proxy | 8443 | ✅ |

**Файлы на сервере:**
- `~/install.sh` — скрипт установки
- `~/phantom-proxy/api.py` — API сервер
- `~/phantom-proxy/https.py` — HTTPS прокси
- `~/phantom-proxy/cert.pem` — SSL сертификат
- `~/phantom-proxy/key.pem` — SSL ключ

---

## 🧪 ТЕСТЫ

```bash
# API Health
curl http://212.233.93.147:8080/health

# API Stats
curl http://212.233.93.147:8080/api/v1/stats

# HTTPS Proxy
curl -k https://212.233.93.147:8443/
```

---

**Выполни команду и отправь скриншот результата!** 🚀
