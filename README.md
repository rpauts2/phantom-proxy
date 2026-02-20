# 👻 PHANTOMPROXY PRO

**Enterprise Red Team Simulation Platform**

![Version](https://img.shields.io/badge/version-12.5%20PRO%2B%2B%2B%2B-blue)
![License](https://img.shields.io/badge/license-Proprietary-green)
![Python](https://img.shields.io/badge/python-3.8%2B-blue)

---

## 📖 ОПИСАНИЕ

**PhantomProxy Pro** — профессиональная платформа для симуляции Red Team атак, тестирования на проникновение и оценки безопасности организации.

**Легальное использование:** Только для аккредитованных организаций с письменными разрешениями (RoE).

---

## ⚠️ LEGAL DISCLAIMER

> **ВАЖНО:** Этот инструмент предназначен ТОЛЬКО для легального тестирования безопасности.
>
> **Разрешено:**
> - ✅ Тестирование с письменного разрешения владельца
> - ✅ Red Team операции по договору
> - ✅ Обучение по кибербезопасности
> - ✅ Исследовательские цели
>
> **Запрещено:**
> - ❌ Несанкционированный доступ
> - ❌ Кража данных
> - ❌ Мошенничество
> - ❌ Любое использование без письменного разрешения
>
> **Используя этот инструмент, вы подтверждаете наличие письменного разрешения (RoE) от владельца тестируемых систем.**

---

## 🎯 ВОЗМОЖНОСТИ

### 🔐 Security Features
- [x] 2FA TOTP Authentication
- [x] Session Management
- [x] Brute Force Protection
- [x] Encrypted Audit Logging
- [x] Role-Based Access Control

### 📊 Analytics & Reporting
- [x] Real-Time Dashboard
- [x] Advanced Analytics
- [x] Custom Report Generator
- [x] PDF Report Export
- [x] SIEM Integration (Splunk, ELK, QRadar)

### 💼 Business Features
- [x] Client Management
- [x] Campaign Scheduling
- [x] Auto-Reports
- [x] Billing & Invoicing
- [x] Proposal Generator
- [x] Contract Templates (RoE, NDA)

### 📧 Notifications
- [x] Email Notifications
- [x] Telegram Bot
- [x] Webhooks (Slack, SIEM)
- [x] Campaign Alerts
- [x] Session Alerts

### 👥 Team Management
- [x] Multi-User Support
- [x] Roles & Permissions
- [x] Task Assignment
- [x] Activity Logging
- [x] Team Statistics

### 🎨 Branding
- [x] White-Label UI
- [x] Custom Colors
- [x] Company Logo
- [x] Custom Texts
- [x] Multi-Language Support (RU/EN)

---

## 📁 СТРУКТУРА ПРОЕКТА

```
phantom-proxy/
├── README.md                    # Документация
├── LICENSE                      # Лицензия
├── .gitignore                   # Git ignore
├── config.example.yaml          # Пример конфигурации
├── requirements.txt             # Python зависимости
│
├── phantomproxy_v12_1_pro.py    # Главная программа (Branded UI)
│
├── modules/
│   ├── v12_siem.py              # SIEM Integration
│   ├── v12_scheduler.py         # Campaign Scheduler
│   ├── v12_notifications.py     # Notifications (Email, Telegram)
│   ├── v12_security.py          # Security (2FA, Sessions)
│   ├── v12_team.py              # Team Management
│   ├── v12_analytics.py         # Analytics Dashboard
│   ├── v12_billing.py           # Billing & Invoices
│   ├── v12_proposals.py         # Proposals & Contracts
│   ├── v12_reporting.py         # PDF Reports
│   └── v12_compliance.py        # Compliance Logging
│
├── templates/                   # Фишлеты
│   ├── microsoft_login.html
│   ├── google_login.html
│   └── ...
│
├── branding/                    # Branding assets
├── reports/                     # PDF отчёты
├── invoices/                    # Invoices
├── proposals/                   # Proposals
├── contracts/                   # Contracts
└── ...
```

---

## 🚀 БЫСТРЫЙ СТАРТ

### 1. Установка

```bash
# Клонирование репозитория
git clone https://github.com/YOUR_USERNAME/phantom-proxy.git
cd phantom-proxy

# Установка зависимостей
pip install -r requirements.txt

# Копирование конфига
cp config.example.yaml config.yaml

# Настройка (отредактируйте config.yaml)
nano config.yaml
```

### 2. Инициализация

```bash
# Запуск главной программы
python3 phantomproxy_v12_1_pro.py

# В меню выберите: 1 (Start All Services)
```

### 3. Доступ

```
Panel:       http://localhost:3000
API:         http://localhost:8080/health
Client:      http://localhost:3000/clients
Billing:     http://localhost:3000/billing
```

---

## 📖 ДОКУМЕНТАЦИЯ

### Полная документация:

| Документ | Описание |
|----------|----------|
| [V12_5_PRO++++_ENTERPRISE.md](./docs/V12_5_PRO++++_ENTERPRISE.md) | Полное руководство v12.5 |
| [API.md](./docs/API.md) | API документация |
| [DEPLOYMENT.md](./docs/DEPLOYMENT.md) | Руководство по развёртыванию |
| [SECURITY.md](./docs/SECURITY.md) | Security policy |
| [CONTRIBUTING.md](./docs/CONTRIBUTING.md) | Contribution guidelines |

### Quick Start Guides:

- [Installation Guide](./docs/guides/installation.md)
- [Configuration Guide](./docs/guides/configuration.md)
- [First Campaign Guide](./docs/guides/first-campaign.md)
- [SIEM Integration Guide](./docs/guides/siem-integration.md)

---

## 🔧 КОНФИГУРАЦИЯ

### Базовая настройка (config.yaml):

```yaml
# Company Information
company:
  name: "PhantomSec Labs"
  email: "info@phantomseclabs.com"
  phone: "+7 (XXX) XXX-XX-XX"
  website: "https://phantomseclabs.com"

# Database
database:
  path: "./phantom.db"
  backup_enabled: true
  backup_interval: "daily"

# Server
server:
  api_port: 8080
  panel_port: 3000
  https_port: 8443

# Email Notifications
email:
  enabled: false
  smtp_server: "smtp.gmail.com"
  smtp_port: 587
  username: ""
  password: ""
  from_email: ""

# Telegram Notifications
telegram:
  enabled: false
  bot_token: ""
  chat_id: ""

# SIEM Integration
siem:
  splunk:
    enabled: false
    hec_url: ""
    hec_token: ""
  elk:
    enabled: false
    es_url: ""
    index: ""
  syslog:
    enabled: false
    server: ""
    port: 514
```

---

## 📊 API ENDPOINTS

### Authentication:
```
POST /api/v1/login          # Login
POST /api/v1/logout         # Logout
POST /api/v1/2fa/enable     # Enable 2FA
POST /api/v1/2fa/verify     # Verify 2FA
```

### Campaigns:
```
GET  /api/v1/campaigns      # List campaigns
POST /api/v1/campaigns      # Create campaign
GET  /api/v1/campaigns/:id  # Get campaign
PUT  /api/v1/campaigns/:id  # Update campaign
DELETE /api/v1/campaigns/:id # Delete campaign
```

### Sessions:
```
GET  /api/v1/sessions       # List sessions
GET  /api/v1/sessions/:id   # Get session
DELETE /api/v1/sessions/:id # Delete session
POST /api/v1/credentials    # Capture credentials
```

### Analytics:
```
GET  /api/v1/analytics/overview      # Overview stats
GET  /api/v1/analytics/trends        # Daily trends
GET  /api/v1/analytics/services      # Service breakdown
GET  /api/v1/analytics/dashboard      # Full dashboard JSON
```

### Billing:
```
GET  /api/v1/invoices      # List invoices
POST /api/v1/invoices      # Create invoice
GET  /api/v1/invoices/:id  # Get invoice
PUT  /api/v1/invoices/:id/pay # Mark as paid
```

### Reports:
```
POST /api/v1/reports/generate  # Generate PDF report
GET  /api/v1/reports/:id        # Get report
DELETE /api/v1/reports/:id      # Delete report
```

---

## 🧪 ТЕСТИРОВАНИЕ

### Запуск тестов:

```bash
# Unit tests
pytest tests/unit/

# Integration tests
pytest tests/integration/

# Coverage
pytest --cov=modules tests/
```

### Тестирование модулей:

```bash
# Test SIEM module
python3 modules/v12_siem.py

# Test Scheduler
python3 modules/v12_scheduler.py

# Test Security
python3 modules/v12_security.py

# Test Notifications
python3 modules/v12_notifications.py
```

---

## 🤝 CONTRIBUTING

### Как внести вклад:

1. Fork репозиторий
2. Создайте feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit изменения (`git commit -m 'Add some AmazingFeature'`)
4. Push в branch (`git push origin feature/AmazingFeature`)
5. Откройте Pull Request

### Code Style:

```bash
# Format code
black modules/

# Lint code
flake8 modules/

# Type checking
mypy modules/
```

---

## 📝 CHANGELOG

### v12.5 PRO++++ (2026-02-20)
- ✅ SIEM Integration (Splunk, ELK, QRadar, Syslog)
- ✅ Campaign Scheduler
- ✅ Auto-Reports
- ✅ Automated Tasks

### v12.4 PRO+++ (2026-02-20)
- ✅ Email Notifications
- ✅ Telegram Bot
- ✅ Webhooks
- ✅ 2FA TOTP
- ✅ Session Management

### v12.3 PRO++ (2026-02-20)
- ✅ Team Management
- ✅ Advanced Analytics
- ✅ Dashboard

### v12.2 PRO+ (2026-02-20)
- ✅ Billing & Invoices
- ✅ Proposal Generator
- ✅ Contract Templates

### v12.1 PRO (2026-02-20)
- ✅ White-Label Branding
- ✅ Client Portal
- ✅ Custom Colors

### v12.0 (2026-02-20)
- ✅ Professional Reporting
- ✅ Compliance Logging
- ✅ Evidence Collection

[Full Changelog](./CHANGELOG.md)

---

## 👥 AUTHORS

- **Lead Developer:** PhantomSec Labs
- **Contributors:** [List of contributors]

---

## 📞 SUPPORT

- **Email:** support@phantomseclabs.com
- **Documentation:** https://docs.phantomseclabs.com
- **Issues:** https://github.com/phantom-proxy/issues

---

## 📜 LICENSE

**Proprietary License**

Этот инструмент предназначен ТОЛЬКО для легального использования с письменного разрешения владельца тестируемых систем.

Полный текст лицензии: [LICENSE](./LICENSE)

---

## ⚠️ SECURITY

Для сообщения об уязвимостях: security@phantomseclabs.com

Политика безопасности: [SECURITY.md](./docs/SECURITY.md)

---

**© 2026 PhantomSec Labs. All rights reserved.**
