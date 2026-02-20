# 🎯 УПРАВЛЕНИЕ КАМПАНИЯМИ

**Полное руководство по созданию и управлению Red Team кампаниями**

---

## 📖 ЧТО ТАКОЕ КАМПАНИЯ?

**Кампания** — это организованная симуляция атаки на клиента в рамках легального тестирования безопасности (Red Team engagement).

---

## 🚀 БЫСТРЫЙ СТАРТ

### 1. Создание кампании

```bash
# В главном меню:
5. 🎯 Create Campaign

# Введите:
Service: Microsoft 365
Subdomains: 10
```

### 2. Планирование кампании

```python
from v12_scheduler import Scheduler

scheduler = Scheduler()

# Запланировать на завтра
from datetime import datetime, timedelta

start = datetime.now() + timedelta(days=1, hours=2)
end = start + timedelta(hours=4)

scheduler.schedule_campaign(
    campaign_name='Q1 Phishing Campaign',
    service='Microsoft 365',
    start_time=start,
    end_time=end,
    created_by=1
)
```

### 3. Запуск кампании

```bash
# В главном меню:
1. 🚀 Start All Services

# Кампания запустится автоматически по расписанию
```

---

## 📋 ЭТАПЫ КАМПАНИИ

### Этап 1: Подготовка

#### 1.1 Получение RoE (Rules of Engagement)

```markdown
✅ Письменное разрешение от клиента
✅ Список разрешённых методов
✅ Список запрещённых действий
✅ Контакты для экстренной связи
✅ Даты проведения
```

#### 1.2 Настройка Scope

```python
from v12_compliance import ScopeEnforcement

scope = ScopeEnforcement()

# Настроить разрешённые домены
scope.save_scope({
    'allowed_domains': ['client.com', 'test.com'],
    'blacklisted_domains': ['gov', 'mil'],
    'max_emails_per_campaign': 1000,
    'allowed_hours': {'start': 9, 'end': 18}
})
```

#### 1.3 Создание команды

```python
from v12_team import TeamManager

manager = TeamManager()

# Создать операторов
manager.create_user('operator1', 'password123', role='operator')
manager.create_user('operator2', 'password456', role='operator')

# Назначить задачи
manager.assign_task(
    title='Setup Microsoft 365 Campaign',
    assigned_to=2,
    created_by=1,
    priority='high',
    due_date='2026-03-01'
)
```

---

### Этап 2: Настройка

#### 2.1 Выбор фишлета

```bash
# Доступные фишлеты:
https://212.233.93.147:8443/microsoft   # Microsoft 365
https://212.233.93.147:8443/google      # Google Workspace
https://212.233.93.147:8443/okta        # Okta SSO
https://212.233.93.147:8443/aws         # AWS Console
```

#### 2.2 Кастомизация фишлета

```bash
# Отредактируйте шаблон
nano ~/phantom-proxy/templates/microsoft_login.html

# Добавьте логотип клиента
# Измените цвета под бренд клиента
# Добавьте кастомный текст
```

#### 2.3 Настройка доменов

```python
# Сгенерировать поддомены
from v12_team import AutoAttack

attack = AutoAttack()
campaign = attack.create_campaign('Microsoft 365', count=10)

print(campaign['subdomains'])
# ['login-microsoft.verdebudget.ru',
#  'secure-microsoft.verdebudget.ru',
#  ...]
```

---

### Этап 3: Запуск

#### 3.1 Pre-launch проверка

```bash
# Проверьте что все сервисы работают
curl http://localhost:8080/health
curl http://localhost:3000/
curl -sk https://localhost:8443/microsoft
```

#### 3.2 Запуск кампании

```python
from v12_scheduler import Scheduler

scheduler = Scheduler()
scheduler.start_scheduler()  # Авто-запуск по расписанию
```

#### 3.3 Мониторинг

```bash
# В главном меню:
4. 📈 View Statistics

# Или через API:
curl http://localhost:8080/api/v1/stats
```

---

### Этап 4: Мониторинг

#### 4.1 Real-time мониторинг сессий

```python
from v12_notifications import NotificationManager

notif = NotificationManager()

# Настроить Telegram уведомления
notif.configure_telegram(
    bot_token='YOUR_BOT_TOKEN',
    chat_id='YOUR_CHAT_ID'
)

# Отправить алерт при новой сессии
notif.send_new_session_alert({
    'email': 'user@client.com',
    'service': 'Microsoft 365',
    'classification': 'EXCELLENT',
    'quality_score': 95
})
```

#### 4.2 SIEM интеграция

```python
from v12_siem import SIEMIntegration

siem = SIEMIntegration()

# Настроить Splunk
siem.configure_splunk(
    hec_url='https://splunk.client.com:8088',
    hec_token='YOUR_TOKEN'
)

# Отправить событие
siem.send_campaign_event({
    'id': 1,
    'name': 'Q1 Phishing Campaign',
    'status': 'running'
}, 'started')
```

#### 4.3 Dashboard

```
Откройте: http://212.233.93.147:3000

Разделы:
📊 Dashboard      - Общая статистика
📋 Sessions       - Все сессии
🎯 Campaigns      - Управление кампаниями
📄 Reports        - Отчёты
👥 Clients        - Клиентский портал
💳 Billing        - Счета
⚖️  Compliance     - Аудит
```

---

### Этап 5: Завершение

#### 5.1 Остановка кампании

```python
from v12_scheduler import Scheduler

scheduler = Scheduler()
scheduler.update_campaign_status(campaign_id=1, status='completed')
```

#### 5.2 Генерация отчёта

```python
from v12_reporting import ReportGenerator

report_gen = ReportGenerator()
report_path = report_gen.generate_pdf_report(
    campaign_id=1,
    client_name='Client Name'
)

print(f"Report: {report_path}")
```

#### 5.3 Отправка отчёта

```python
from v12_notifications import NotificationManager

notif = NotificationManager()
notif.configure_email(
    smtp_server='smtp.gmail.com',
    smtp_port=587,
    username='your@gmail.com',
    password='password',
    from_email='your@gmail.com'
)

notif.send_report_email(
    to_email='client@client.com',
    report_path=report_path,
    campaign_name='Q1 Phishing Campaign'
)
```

#### 5.4 Создание счёта

```python
from v12_billing import InvoiceGenerator

generator = InvoiceGenerator()
invoice = generator.create_invoice(
    client_id=1,
    campaign_id=1,
    rate=500,  # $500/hour
    due_days=30
)

print(f"Invoice: ${invoice['total']:,.2f}")
```

---

## 📊 МЕТРИКИ КАМПАНИИ

### Основные метрики

```python
from v12_analytics import AnalyticsDashboard

analytics = AnalyticsDashboard()

# Общая статистика
overview = analytics.get_overview_stats(days=30)
print(f"Total Sessions: {overview['total_sessions']}")
print(f"Campaigns: {overview['total_campaigns']}")
print(f"Clients: {overview['total_clients']}")
print(f"Avg Quality: {overview['avg_quality']}")

# Конверсия
funnel = analytics.get_conversion_funnel()
print(f"Credential Rate: {funnel['credential_rate']}%")
print(f"Quality Rate: {funnel['quality_rate']}%")
```

### Метрики по сервисам

```python
services = analytics.get_service_breakdown(days=30)

for service in services:
    print(f"{service['service']}:")
    print(f"  Sessions: {service['count']}")
    print(f"  Success Rate: {service['success_rate']}%")
    print(f"  Avg Quality: {service['avg_quality']}")
```

---

## 🎯 BEST PRACTICES

### Планирование

✅ Получите письменное разрешение (RoE)  
✅ Определите четкие границы (scope)  
✅ Настройте emergency контакты  
✅ Задокументируйте все действия  

### Выполнение

✅ Начните с малой группы пользователей  
✅ Мониторьте в реальном времени  
✅ Будьте готовы к экстренной остановке  
✅ Логируйте все действия  

### Отчётность

✅ Включите executive summary  
✅ Добавьте технические детали  
✅ Предоставьте рекомендации  
✅ Включите метрики успеха  

### Безопасность

✅ Используйте шифрование  
✅ Удаляйте данные после отчёта  
✅ Ограничьте доступ к результатам  
✅ Проводите debrief с клиентом  

---

## 🐛 TROUBLESHOOTING

### Кампания не запускается

```bash
# Проверьте планировщик
python3 -c "from v12_scheduler import Scheduler; s = Scheduler(); print(s.get_scheduler_stats())"

# Проверьте логи
tail -f ~/phantom-proxy/scheduler_logs/*.log

# Перезапустите планировщик
pkill -f scheduler
cd ~/phantom-proxy
python3 phantomproxy_v12_1_pro.py
```

### Данные не собираются

```bash
# Проверьте API
curl http://localhost:8080/api/v1/stats

# Проверьте базу данных
sqlite3 ~/phantom-proxy/phantom.db "SELECT COUNT(*) FROM sessions;"

# Проверьте логи
tail -f ~/phantom-proxy/api.log
```

### Клиент не получил отчёт

```bash
# Проверьте email логи
tail -f ~/phantom-proxy/notifications/*.log

# Проверьте что файл существует
ls -lh ~/phantom-proxy/reports/*.pdf

# Отправьте вручную
python3 -c "
from v12_notifications import NotificationManager
n = NotificationManager()
n.configure_email(...)
n.send_report_email('client@client.com', 'reports/Report.pdf', 'Campaign')
"
```

---

## 📚 ДОПОЛНИТЕЛЬНЫЕ РЕСУРСЫ

- [Phishlets Guide](./Phishlets-Guide.md)
- [Sessions Guide](./Sessions-Guide.md)
- [Reports Guide](./Reports-Guide.md)
- [SIEM Integration](./SIEM-Integration.md)
- [Team Management](./Team-Management.md)

---

**© 2026 PhantomSec Labs. All rights reserved.**

**Last Updated:** February 20, 2026
