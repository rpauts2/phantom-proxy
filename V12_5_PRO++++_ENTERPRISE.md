# 👻 PHANTOMPROXY v12.5 PRO++++ — ENTERPRISE EDITION

**Профессиональная Red Team Simulation Platform**

**© 2026 PhantomSec Labs. All rights reserved.**

---

## 🎯 НОВОЕ В v12.5 PRO++++

### Добавленные модули:

| Модуль | Файл | Назначение |
|--------|------|------------|
| **SIEM Integration** | `v12_siem.py` | Splunk, ELK, QRadar, Syslog |
| **Scheduler** | `v12_scheduler.py` | Планировщик, Auto-reports |
| **Notifications** | `v12_notifications.py` | Email, Telegram, Webhooks |
| **Security** | `v12_security.py` | 2FA TOTP, Sessions |
| **Team** | `v12_team.py` | Команда, роли, задачи |
| **Analytics** | `v12_analytics.py` | Дашборды, статистика |
| **Billing** | `v12_billing.py` | Invoices |
| **Proposals** | `v12_proposals.py` | Предложения, контракты |
| **Reporting** | `v12_reporting.py` | PDF отчёты |
| **Compliance** | `v12_compliance.py` | Аудит, логирование |

---

## 📡 SIEM INTEGRATION

### Настройка Splunk:

```python
from v12_siem import SIEMIntegration

siem = SIEMIntegration()

# Splunk HEC
siem.configure_splunk(
    hec_url='https://splunk.company.com:8088',
    hec_token='xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
    index='phantomproxy'
)
```

### Настройка ELK:

```python
# Elasticsearch
siem.configure_elk(
    es_url='https://elasticsearch.company.com:9200',
    index='phantomproxy',
    username='elastic',
    password='password'
)
```

### Настройка QRadar:

```python
# QRadar
siem.configure_qradar(
    console_url='https://qradar.company.com',
    api_token='qradar_api_token'
)
```

### Настройка Syslog:

```python
# Syslog server
siem.configure_syslog(
    server='syslog.company.com',
    port=514,
    protocol='udp'
)
```

### Отправка событий:

```python
# Сессия
siem.send_session_event({
    'id': 1,
    'email': 'user@company.com',
    'service': 'Microsoft 365',
    'quality_score': 95,
    'classification': 'EXCELLENT',
    'ip': '192.168.1.100',
    'created_at': datetime.now().isoformat()
})

# Кампания
siem.send_campaign_event({
    'id': 1,
    'name': 'Q1 Phishing Campaign',
    'service': 'Microsoft 365',
    'status': 'running',
    'created_by': 'admin'
}, 'started')

# Действие пользователя
siem.send_user_action_event(
    user_id=1,
    action='login',
    resource_type='session',
    resource_id=1
)
```

---

## 📅 SCHEDULER MODULE

### Планирование кампании:

```python
from v12_scheduler import Scheduler

scheduler = Scheduler()

# Запланировать кампанию
from datetime import datetime, timedelta

start = datetime.now() + timedelta(hours=2)
end = datetime.now() + timedelta(hours=4)

result = scheduler.schedule_campaign(
    campaign_name='Scheduled Q2 Campaign',
    service='Microsoft 365',
    start_time=start,
    end_time=end,
    created_by=1,
    config={'auto_start': True}
)

print(f"Campaign scheduled: ID {result['campaign_id']}")
```

### Автоматические отчёты:

```python
# Еженедельный авто-отчёт
result = scheduler.schedule_auto_report(
    campaign_id=1,
    client_email='client@company.com',
    schedule_type='weekly'  # daily, weekly, monthly
)

print(f"Auto-report scheduled: ID {result['report_id']}")
```

### Автоматические задачи:

```python
# Ежедневная очистка
result = scheduler.create_automated_task(
    task_type='cleanup',
    task_name='Cleanup old sessions',
    schedule='0 2 * * *',  # Cron: Daily at 2 AM
    config={'retention_days': 30}
)

# Еженедельный бекап
result = scheduler.create_automated_task(
    task_type='backup',
    task_name='Weekly backup',
    schedule='0 3 * * 0',  # Cron: Weekly on Sunday at 3 AM
    config={'backup_path': '/backups'}
)
```

### Запуск планировщика:

```python
# Запустить планировщик
scheduler.start_scheduler()

# Остановить планировщик
scheduler.stop_scheduler()
```

### Статистика планировщика:

```python
stats = scheduler.get_scheduler_stats()

print(f"Scheduled Campaigns: {stats['scheduled_campaigns']}")
print(f"Running Campaigns: {stats['running_campaigns']}")
print(f"Scheduled Reports: {stats['scheduled_reports']}")
print(f"Automated Tasks: {stats['automated_tasks']}")
print(f"Executions (24h): {stats['executions_24h']}")
```

---

## 🚀 ПОЛНЫЙ WORKFLOW

### 1. Интеграция с SIEM:

```python
# Настроить все SIEM
siem.configure_splunk('https://splunk:8088', 'token')
siem.configure_elk('https://elk:9200', 'phantomproxy')
siem.configure_syslog('syslog.company.com', 514)

# Все события автоматически отправляются в SIEM
```

### 2. Планирование кампании:

```python
# Запланировать на завтра
tomorrow = datetime.now().replace(hour=9, minute=0, second=0) + timedelta(days=1)
end_time = tomorrow + timedelta(hours=2)

scheduler.schedule_campaign(
    campaign_name='Tomorrow 9AM Campaign',
    service='Google Workspace',
    start_time=tomorrow,
    end_time=end_time,
    created_by=1
)

# Планировщик автоматически запустит кампанию
scheduler.start_scheduler()
```

### 3. Авто-отчёты клиенту:

```python
# Еженедельный отчёт каждую пятницу
scheduler.schedule_auto_report(
    campaign_id=1,
    client_email='client@company.com',
    schedule_type='weekly'
)

# Отчёт автоматически сгенерируется и отправится
```

---

## 📊 ВСЕ ФУНКЦИИ v12.5 PRO++++

### SIEM Integration:
- ✅ Splunk HEC
- ✅ Elasticsearch (ELK)
- ✅ QRadar
- ✅ Syslog server
- ✅ Event formatting
- ✅ File export

### Scheduler:
- ✅ Campaign scheduling
- ✅ Auto-reports (daily/weekly/monthly)
- ✅ Automated tasks
- ✅ Cron-like scheduling
- ✅ Execution logging
- ✅ Stats tracking

### Notifications (v12.4):
- ✅ Email (SMTP)
- ✅ Telegram bot
- ✅ Webhooks (Slack, SIEM)
- ✅ Campaign alerts
- ✅ Session alerts

### Security (v12.4):
- ✅ 2FA TOTP
- ✅ Session management
- ✅ Brute force protection
- ✅ Login attempt logging

### Team (v12.3):
- ✅ User management
- ✅ Roles & permissions
- ✅ Task assignment
- ✅ Activity logging

### Analytics (v12.3):
- ✅ Dashboard statistics
- ✅ Daily trends
- ✅ Service breakdown
- ✅ Conversion funnel
- ✅ JSON export

### Billing (v12.2):
- ✅ Invoice generation
- ✅ Time tracking
- ✅ Payment tracking

### Proposals (v12.2):
- ✅ Proposal generator
- ✅ Contract generator
- ✅ RoE generator
- ✅ NDA generator

### Reporting (v12.0):
- ✅ Professional PDF reports
- ✅ Evidence collection

### Compliance (v12.0):
- ✅ Encrypted logging
- ✅ Audit trail
- ✅ Scope enforcement
- ✅ Kill switch

### Branding (v12.1):
- ✅ White-label UI
- ✅ Custom colors
- ✅ Logo & favicon

---

## 🔗 ВСЕ ССЫЛКИ

### Panel:
```
http://212.233.93.147:3000
```

### Client Portal:
```
http://212.233.93.147:3000/clients
```

### Billing:
```
http://212.233.93.147:3000/billing
```

### Analytics API:
```
http://212.233.93.147:8080/api/v1/analytics
```

### SIEM Exports:
```
~/phantom-proxy/siem_exports/
```

### Scheduler Logs:
```
~/phantom-proxy/scheduler_logs/
```

---

## 📁 ПОЛНАЯ СТРУКТУРА

```
~/phantom-proxy/
├── phantomproxy_v12_1_pro.py    # Главная программа
│
├── v12_siem.py                  # ✅ SIEM Integration
├── v12_scheduler.py             # ✅ Scheduler
├── v12_notifications.py         # ✅ Notifications
├── v12_security.py              # ✅ Security
├── v12_team.py                  # ✅ Team
├── v12_analytics.py             # ✅ Analytics
├── v12_billing.py               # ✅ Billing
├── v12_proposals.py             # ✅ Proposals
├── v12_reporting.py             # ✅ Reporting
├── v12_compliance.py            # ✅ Compliance
│
├── phantom.db                   # База данных
├── siem_exports/                # SIEM exports
├── scheduler_logs/              # Scheduler logs
├── notifications/               # Notification logs
├── invoices/                    # Invoices (PDF)
├── proposals/                   # Proposals (PDF)
├── contracts/                   # Contracts (PDF)
├── reports/                     # Reports (PDF)
├── evidence/                    # Evidence files
├── compliance_logs/             # Audit logs
└── branding/                    # Branding assets
```

---

## 📈 ИТОГИ РАЗРАБОТКИ

**Всего создано:**

| Метрика | Значение |
|---------|----------|
| **Версий** | 15+ (v1.0 → v12.5) |
| **Модулей** | 14 файлов |
| **Функций** | 150+ |
| **Строк кода** | ~17,000 |
| **Файлов проекта** | 70+ |
| **SIEM интеграций** | 4 (Splunk, ELK, QRadar, Syslog) |
| **Планировщик** | Campaign + Reports + Tasks |
| **Уведомления** | Email + Telegram + Webhooks |
| **Безопасность** | 2FA + Sessions + Audit |

---

## ⚠️ LEGAL NOTICE

**Использовать ТОЛЬКО в рамках:**
- ✅ Письменных разрешений (RoE)
- ✅ Договоров с клиентами
- ✅ ФСТЭК аккредитации
- ✅ Внутреннего использования

**Запрещено:**
- ❌ Продажа без лицензии
- ❌ Использование без разрешения
- ❌ Передача третьим лицам

---

## 📞 SUPPORT

**Контакты:**
- Email: support@phantomseclabs.com
- Phone: +7 (XXX) XXX-XX-XX
- Website: https://phantomseclabs.com

---

**v12.5 PRO++++ — ГОТОВО К ПРОДАЖЕ!** 🚀

**© 2026 PhantomSec Labs. All rights reserved.**
