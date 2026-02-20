# 👻 PHANTOMPROXY v12.4 PRO+++ — ULTIMATE BUSINESS EDITION

**Профессиональная Red Team Simulation Platform**

**© 2026 PhantomSec Labs. All rights reserved.**

---

## 🎯 НОВОЕ В v12.4 PRO+++

### Добавленные модули:

| Модуль | Файл | Назначение |
|--------|------|------------|
| **Notifications** | `v12_notifications.py` | Email, Telegram, Webhooks |
| **Security** | `v12_security.py` | 2FA TOTP, Session Management |
| **Team** | `v12_team.py` | Команда, роли, задачи |
| **Analytics** | `v12_analytics.py` | Дашборды, статистика |
| **Billing** | `v12_billing.py` | Invoices, трекинг |
| **Proposals** | `v12_proposals.py` | Предложения, контракты |
| **Reporting** | `v12_reporting.py` | PDF отчёты |
| **Compliance** | `v12_compliance.py` | Аудит, логирование |

---

## 📧 NOTIFICATION SYSTEM

### Настройка Email:

```python
from v12_notifications import NotificationManager

manager = NotificationManager()

# Gmail
manager.configure_email(
    smtp_server='smtp.gmail.com',
    smtp_port=587,
    username='your@gmail.com',
    password='your_password',
    from_email='your@gmail.com',
    from_name='PhantomSec Labs'
)
```

### Настройка Telegram:

```python
# Получить bot token: @BotFather
# Получить chat_id: @userinfobot

manager.configure_telegram(
    bot_token='123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11',
    chat_id='123456789'
)
```

### Добавление Webhook:

```python
manager.add_webhook('https://hooks.slack.com/services/YOUR/WEBHOOK/URL')
manager.add_webhook('https://your-siem.com/api/webhook')
```

### Отправка уведомлений:

```python
# Email с отчётом
manager.send_report_email(
    to_email='client@example.com',
    report_path='reports/Report_Client_20260220.pdf',
    campaign_name='Q1 Phishing Campaign'
)

# Telegram алерт
manager.send_telegram("""
🎯 <b>New Session Captured!</b>

<b>Email:</b> user@company.com
<b>Service:</b> Microsoft 365
<b>Quality:</b> EXCELLENT
<b>Score:</b> 95/100
""")

# Campaign alert
manager.send_campaign_alert(
    campaign_name='Q1 Campaign',
    event_type='started',
    details='Campaign launched successfully'
)

# Webhook (SIEM integration)
manager.send_webhook({
    'type': 'session_captured',
    'session': session_data,
    'timestamp': datetime.now().isoformat()
})
```

---

## 🔐 SECURITY MODULE

### 2FA TOTP Setup:

```python
from v12_security import SecurityManager

security = SecurityManager()

# Включить 2FA для пользователя
result = security.enable_2fa(user_id=1)

if result['success']:
    print(f"Secret: {result['secret']}")
    print(f"Backup Codes: {result['backup_codes']}")
    print(f"QR URI: {result['qr_provisioning_uri']}")
    
    # QR code для Google Authenticator
    import qrcode
    qr = qrcode.make(result['qr_provisioning_uri'])
    qr.save('2fa_qr.png')
```

### 2FA Verification:

```python
# Проверка TOTP кода
result = security.verify_2fa(user_id=1, code='123456')

if result['success']:
    print("✅ 2FA verified")
else:
    print(f"❌ {result['error']}")

# Проверка backup кода
result = security.verify_backup_code(user_id=1, code='backup_code_here')
```

### Session Management:

```python
# Создание сессии
result = security.create_session(
    user_id=1,
    ip_address='192.168.1.100',
    user_agent='Mozilla/5.0',
    expires_hours=24
)

token = result['token']

# Проверка сессии
validation = security.validate_session(token)

if validation['success']:
    print("✅ Session valid")
else:
    print(f"❌ {validation['error']}")

# Завершение сессии
security.invalidate_session(token)

# Завершение всех сессий пользователя
security.invalidate_all_sessions(user_id=1)
```

### Brute Force Protection:

```python
# Проверка на brute force
result = security.check_brute_force(
    username='admin',
    ip_address='192.168.1.100',
    max_attempts=5,
    window_minutes=15
)

if result['blocked']:
    print(f"⚠️ Blocked: {result['attempts']} failed attempts")
else:
    print(f"✅ Allowed: {result['attempts']} failed attempts")
```

### Login Attempt Logging:

```python
# Логирование попытки входа
security.log_login_attempt(
    username='admin',
    ip_address='192.168.1.100',
    success=True
)
```

### Active Sessions:

```python
# Получить активные сессии пользователя
sessions = security.get_user_sessions(user_id=1)

for session in sessions:
    print(f"Token: {session['token'][:20]}...")
    print(f"IP: {session['ip_address']}")
    print(f"Created: {session['created_at']}")
    print(f"Expires: {session['expires_at']}")
```

### Security Statistics:

```python
stats = security.get_security_stats()

print(f"Active Sessions: {stats['active_sessions']}")
print(f"2FA Enabled Users: {stats['twofa_enabled_users']}")
print(f"Failed Logins (24h): {stats['failed_logins_24h']}")
```

---

## 🚀 ПОЛНЫЙ WORKFLOW

### 1. Продажа проекта:

```python
from v12_proposals import ProposalGenerator, ContractGenerator
from v12_billing import InvoiceGenerator
from v12_notifications import NotificationManager

# Proposal
proposal_gen = ProposalGenerator()
proposal = proposal_gen.generate_proposal(
    client_name='ACME Corp',
    service_type='Red Team Assessment',
    price=50000
)

# Отправка Email
notif = NotificationManager()
notif.configure_email(...)  # Настроить SMTP

notif.send_email(
    to_email='client@acme.com',
    subject='Red Team Proposal',
    body='<h1>Proposal Attached</h1>',
    attachments=[proposal]
)
```

### 2. Создание команды:

```python
from v12_team import TeamManager

manager = TeamManager()

# Создать пользователей
manager.create_user('operator1', 'password123', role='operator', email='op1@test.com')
manager.create_user('viewer1', 'password456', role='viewer', email='view@test.com')

# Назначить задачи
manager.assign_task(
    title='Setup Q1 Campaign',
    assigned_to=2,
    created_by=1,
    priority='high',
    due_date='2026-03-01'
)
```

### 3. Запуск кампании с уведомлениями:

```python
# Включить 2FA для оператора
security.enable_2fa(user_id=2)

# Создать сессию
session = security.create_session(
    user_id=2,
    ip_address='192.168.1.100',
    user_agent='Mozilla/5.0'
)

# Отправить Telegram алерт о запуске
notif.send_campaign_alert(
    campaign_name='Q1 Phishing Campaign',
    event_type='started'
)
```

### 4. Мониторинг с аналитикой:

```python
from v12_analytics import AnalyticsDashboard

analytics = AnalyticsDashboard()

# Получить дашборд
dashboard = analytics.generate_dashboard_json(days=30)

# Отправить webhook в SIEM
notif.send_webhook({
    'type': 'campaign_stats',
    'stats': dashboard,
    'timestamp': datetime.now().isoformat()
})
```

### 5. Завершение и отчёт:

```python
from v12_reporting import ReportGenerator

report_gen = ReportGenerator()
report_path = report_gen.generate_pdf_report(client_name='ACME Corp')

# Отправить отчёт клиенту
notif.send_report_email(
    to_email='client@acme.com',
    report_path=report_path,
    campaign_name='Q1 Phishing Campaign'
)

# Создать счёт
invoice_gen = InvoiceGenerator()
invoice = invoice_gen.create_invoice(
    client_id=1,
    campaign_id=1,
    rate=500,
    due_days=30
)

# Отправить счёт
notif.send_email(
    to_email='billing@acme.com',
    subject=f"Invoice {invoice['invoice_number']}",
    body='<h1>Invoice Attached</h1>',
    attachments=[invoice['pdf_path']]
)
```

---

## 📊 ВСЕ ФУНКЦИИ v12.4 PRO+++

### Branding:
- ✅ White-label UI
- ✅ Custom colors
- ✅ Logo & favicon
- ✅ Custom texts

### Team:
- ✅ User management
- ✅ Roles & permissions
- ✅ Task assignment
- ✅ Activity logging

### Security:
- ✅ 2FA TOTP
- ✅ Session management
- ✅ Brute force protection
- ✅ Login attempt logging

### Notifications:
- ✅ Email (SMTP)
- ✅ Telegram bot
- ✅ Webhooks (Slack, SIEM)
- ✅ Campaign alerts
- ✅ Session alerts

### Analytics:
- ✅ Overview statistics
- ✅ Daily trends
- ✅ Service breakdown
- ✅ Conversion funnel
- ✅ Geographic distribution
- ✅ JSON export

### Billing:
- ✅ Invoice generation (PDF)
- ✅ Time tracking
- ✅ Payment tracking
- ✅ Revenue stats

### Proposals:
- ✅ Proposal generator (PDF)
- ✅ Contract generator (PDF)
- ✅ RoE generator (PDF)
- ✅ NDA generator (PDF)

### Reporting:
- ✅ Professional reports (PDF)
- ✅ Evidence collection
- ✅ Auto-email reports

### Compliance:
- ✅ Encrypted logging
- ✅ Audit trail
- ✅ Scope enforcement
- ✅ Kill switch

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

---

## 📁 ПОЛНАЯ СТРУКТУРА

```
~/phantom-proxy/
├── phantomproxy_v12_1_pro.py    # Главная программа
│
├── v12_notifications.py         # ✅ Notifications
├── v12_security.py              # ✅ Security (2FA)
├── v12_team.py                  # ✅ Team Management
├── v12_analytics.py             # ✅ Analytics
├── v12_billing.py               # ✅ Billing
├── v12_proposals.py             # ✅ Proposals
├── v12_reporting.py             # ✅ Reporting
├── v12_compliance.py            # ✅ Compliance
│
├── phantom.db                   # База данных
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

**v12.4 PRO+++ — ГОТОВО К ПРОДАЖЕ!** 🚀

**© 2026 PhantomSec Labs. All rights reserved.**
