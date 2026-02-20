# 👻 PHANTOMPROXY v12.3 PRO++ — COMPLETE BUSINESS EDITION

**Профессиональная Red Team Simulation Platform**

**© 2026 PhantomSec Labs. All rights reserved.**

---

## 🎯 НОВОЕ В v12.3 PRO++

### Добавленные модули:

| Модуль | Файл | Назначение |
|--------|------|------------|
| **Team Management** | `v12_team.py` | Команда, роли, задачи |
| **Analytics** | `v12_analytics.py` | Дашборды, графики, статистика |
| **Billing** | `v12_billing.py` | Счета и invoices |
| **Proposals** | `v12_proposals.py` | Предложения и контракты |
| **Reporting** | `v12_reporting.py` | PDF отчёты |
| **Compliance** | `v12_compliance.py` | Аудит и логирование |

---

## 📁 ПОЛНАЯ СТРУКТУРА ПРОЕКТА

```
~/phantom-proxy/
├── phantomproxy_v12_1_pro.py    # Главная программа (Branded UI)
│
├── v12_team.py                  # ✅ Team Management
├── v12_analytics.py             # ✅ Advanced Analytics
├── v12_billing.py               # ✅ Billing & Invoices
├── v12_proposals.py             # ✅ Proposals & Contracts
├── v12_reporting.py             # ✅ Professional Reporting
├── v12_compliance.py            # ✅ Compliance Logging
│
├── phantom.db                   # База данных
│
├── branding/                    # Branding assets
├── invoices/                    # Invoices (PDF)
├── proposals/                   # Proposals (PDF)
├── contracts/                   # Contracts (PDF)
├── reports/                     # Reports (PDF)
├── evidence/                    # Evidence files
└── compliance_logs/             # Audit logs
```

---

## 👥 TEAM MANAGEMENT

### Роли и права:

| Роль | Описание | Права |
|------|----------|-------|
| **admin** | Полный доступ | Все функции |
| **operator** | Оператор кампаний | Campaigns, Reports, Clients |
| **viewer** | Наблюдатель | View only |
| **client** | Клиент | Portal access, Reports download |

### Создание пользователя:

```python
from v12_team import TeamManager

manager = TeamManager()

# Создать пользователя
result = manager.create_user(
    username='john.doe',
    password='SecurePass123!',
    role='operator',
    email='john@company.com'
)

if result['success']:
    print(f"User ID: {result['user_id']}")
    print(f"API Key: {result['api_key']}")
```

### Аутентификация:

```python
# Логин
result = manager.authenticate('john.doe', 'SecurePass123!')

if result['success']:
    print(f"Token: {result['user']['token']}")
    print(f"Role: {result['user']['role']}")
```

### Назначение задачи:

```python
# Назначить задачу
result = manager.assign_task(
    title='Setup Q1 Phishing Campaign',
    assigned_to=2,  # User ID
    created_by=1,
    campaign_id=5,
    description='Configure Microsoft 365 phishing campaign',
    priority='high',
    due_date='2026-03-01'
)
```

### Проверка прав:

```python
# Проверить право доступа
has_access = manager.check_permission(user_id=2, permission='campaigns.view')
print(f"Can view campaigns: {has_access}")
```

---

## 📊 ADVANCED ANALYTICS

### Общая статистика:

```python
from v12_analytics import AnalyticsDashboard

analytics = AnalyticsDashboard()

# Получить статистику за 30 дней
overview = analytics.get_overview_stats(days=30)

print(f"Sessions: {overview['total_sessions']}")
print(f"Campaigns: {overview['total_campaigns']}")
print(f"Clients: {overview['total_clients']}")
print(f"Avg Quality: {overview['avg_quality']}")
```

### Тренды по дням:

```python
# Дневной тренд
trends = analytics.get_daily_trend(days=30)

for day in trends:
    print(f"{day['date']}: {day['sessions']} sessions, quality {day['avg_quality']}")
```

### Разбивка по сервисам:

```python
# Сервисы
services = analytics.get_service_breakdown(days=30)

for svc in services:
    print(f"{svc['service']}: {svc['count']} sessions, {svc['success_rate']}% success")
```

### Воронка конверсии:

```python
# Конверсия
funnel = analytics.get_conversion_funnel()

print(f"Total: {funnel['total_sessions']}")
print(f"With Credentials: {funnel['with_credentials']} ({funnel['credential_rate']}%)")
print(f"High Quality: {funnel['high_quality']} ({funnel['quality_rate']}%)")
```

### Полный Dashboard JSON:

```python
# Генерация полного дашборда
dashboard = analytics.generate_dashboard_json(days=30)

# Сохранить в файл
import json
with open('dashboard.json', 'w') as f:
    json.dump(dashboard, f, indent=2)
```

---

## 💳 BILLING

### Создать счёт:

```python
from v12_billing import InvoiceGenerator

generator = InvoiceGenerator()

invoice = generator.create_invoice(
    client_id=1,
    campaign_id=1,
    rate=500,  # $500/hour
    due_days=30
)

print(f"Invoice #{invoice['invoice_number']}")
print(f"Total: ${invoice['total']:,.2f}")
print(f"PDF: {invoice['pdf_path']}")
```

### Получить все счета:

```python
invoices = generator.get_all_invoices()

for inv in invoices:
    print(f"{inv['invoice_number']}: ${inv['total']:,.2f} - {inv['status']}")
```

### Отметить оплаченным:

```python
generator.mark_as_paid(invoice_id=1)
```

---

## 📄 PROPOSALS & CONTRACTS

### Коммерческое предложение:

```python
from v12_proposals import ProposalGenerator

generator = ProposalGenerator()

proposal = generator.generate_proposal(
    client_name='ACME Corp',
    service_type='Red Team Assessment',
    duration='4 weeks',
    price=50000
)

print(f"Proposal: {proposal}")
```

### Rules of Engagement:

```python
from v12_proposals import ContractGenerator

generator = ContractGenerator()

roe = generator.generate_roe(
    client_name='ACME Corp',
    campaign_name='Q1 Phishing Campaign',
    start_date='2026-03-01',
    end_date='2026-03-31',
    authorized_ips=['192.168.1.100', '10.0.0.50']
)

print(f"RoE: {roe}")
```

### NDA:

```python
nda = generator.generate_nda(client_name='ACME Corp')
print(f"NDA: {nda}")
```

---

## 🚀 БЫСТРЫЙ СТАРТ

### 1. Установка:

```bash
cd ~/phantom-proxy
pip install reportlab cryptography
```

### 2. Тестирование модулей:

```bash
# Team Management
python3 v12_team.py

# Analytics
python3 v12_analytics.py

# Billing
python3 v12_billing.py

# Proposals
python3 v12_proposals.py
```

### 3. Запуск главной программы:

```bash
python3 phantomproxy_v12_1_pro.py
```

---

## 📊 WORKFLOW (ПОЛНЫЙ ЦИКЛ)

### 1. Продажа:

```
Client Contact
    ↓
Proposal Generator → PDF → Send to Client
    ↓
Contract + RoE → PDF → Sign
    ↓
NDA → PDF → Sign
    ↓
Deposit Invoice → PDF → Payment
```

### 2. Планирование:

```
Team Manager → Create Users
    ↓
Assign Roles (admin, operator, viewer)
    ↓
Create Tasks → Assign to Team
    ↓
Setup Campaign
```

### 3. Выполнение:

```
Launch Campaign
    ↓
Monitor Sessions (Real-time)
    ↓
Analytics Dashboard → Track Progress
    ↓
Compliance Logs → Audit Trail
```

### 4. Завершение:

```
Generate Report (PDF)
    ↓
Generate Invoice (PDF)
    ↓
Send to Client
    ↓
Payment → Mark as Paid
    ↓
Retesting (Optional)
```

---

## 🎨 BRANDING

### Настройки:

```python
# phantomproxy_v12_1_pro.py
COMPANY_NAME = "PhantomSec Labs"
COMPANY_LOGO_TEXT = "👻"
BRAND_COLORS = {
    'primary': '#1E3A8A',
    'secondary': '#3B82F6',
    'accent': '#EF4444',
}
CONTACT_INFO = {
    'email': 'info@phantomseclabs.com',
    'phone': '+7 (XXX) XXX-XX-XX',
    'website': 'https://phantomseclabs.com',
}
```

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

## ⚖️ COMPLIANCE

### Логирование:

- ✅ Все действия пользователей
- ✅ Шифрование (Fernet)
- ✅ Hash verification
- ✅ Audit trail
- ✅ Activity log

### Scope Enforcement:

- ✅ Domain blacklist/whitelist
- ✅ Campaign limits
- ✅ Time restrictions
- ✅ Auto-kill switch

---

## 🛡️ БЕЗОПАСНОСТЬ

### Аутентификация:

- ✅ Hashed passwords (SHA-256)
- ✅ API keys
- ✅ Session tokens
- ✅ 2FA (TOTP) — в разработке

### Доступ:

- ✅ Role-based permissions
- ✅ IP whitelisting — опционально
- ✅ Session recording

### Данные:

- ✅ Шифрование логов
- ✅ Auto-delete policies
- ✅ Data retention

---

## 📈 ANALYTICS FEATURES

### Dashboard:

- ✅ Overview statistics
- ✅ Daily trends
- ✅ Service breakdown
- ✅ Quality distribution
- ✅ Geographic distribution
- ✅ Hourly distribution
- ✅ Conversion funnel
- ✅ Revenue stats

### Export:

- ✅ JSON export
- ✅ CSV export
- ✅ PDF reports

---

## 🎯 ИТОГИ

**v12.3 PRO++ включает:**

✅ **Branding + White Label**  
✅ **Team Management** (Roles, Tasks, Permissions)  
✅ **Advanced Analytics** (Dashboards, Trends, Funnels)  
✅ **Billing & Invoicing** (PDF, Tracking)  
✅ **Proposals & Contracts** (PDF Templates)  
✅ **Professional Reporting** (PDF Reports)  
✅ **Compliance Logging** (Encrypted Audit)  
✅ **Client Portal** (Login, Reports Download)  

**Готово для профессионального использования!**

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

**v12.3 PRO++ — ГОТОВО К ПРОДАЖЕ!** 🚀

**© 2026 PhantomSec Labs. All rights reserved.**
