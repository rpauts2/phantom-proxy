# 🏠 PHANTOMPROXY PRO WIKI

**Добро пожаловать в официальную документацию PhantomProxy Pro v12.5 PRO++++**

---

## 📖 БЫСТРЫЙ СТАРТ

### 1. Установка

```bash
# Клонирование репозитория
git clone https://github.com/rpauts2/phantom-proxy.git
cd phantom-proxy

# Установка зависимостей
pip install -r requirements.txt

# Копирование конфига
cp config.example.yaml config.yaml

# Настройка (отредактируйте config.yaml)
nano config.yaml
```

### 2. Запуск

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
Billing:     http://212.233.93.147:3000/billing
```

---

## 📚 СОДЕРЖАНИЕ

### Основное

| Документ | Описание |
|----------|----------|
| [📋 Что такое PhantomProxy?](./What-is-PhantomProxy.md) | Введение в платформу |
| [🚀 Быстрый старт](./Quick-Start.md) | 5 минут до первого запуска |
| [📦 Установка](./Installation.md) | Подробная инструкция по установке |
| [⚙️ Настройка](./Configuration.md) | Конфигурация платформы |
| [🎯 Первое использование](./First-Use.md) | Первая кампания |

### Модули

| Модуль | Документация |
|--------|-------------|
| **Фишлеты** | [Фишлеты](./Phishlets.md) |
| **Кампании** | [Кампании](./Campaigns.md) |
| **Сессии** | [Сессии](./Sessions.md) |
| **Отчёты** | [Отчёты](./Reports.md) |
| **Billing** | [Billing](./Billing.md) |
| **SIEM** | [SIEM Integration](./SIEM-Integration.md) |
| **Scheduler** | [Планировщик](./Scheduler.md) |

### Продвинутые темы

| Тема | Документ |
|------|----------|
| **API** | [API Documentation](./API.md) |
| **Безопасность** | [Security](./Security.md) |
| **2FA** | [2FA Setup](./2FA-Setup.md) |
| **SIEM** | [SIEM Guide](./SIEM-Guide.md) |
| **Команда** | [Team Management](./Team-Management.md) |
| **Бэкапы** | [Backup & Restore](./Backup-Restore.md) |

### Интеграции

| Интеграция | Гид |
|------------|-----|
| **Splunk** | [Splunk Integration](./integrations/Splunk.md) |
| **ELK** | [ELK Integration](./integrations/ELK.md) |
| **QRadar** | [QRadar Integration](./integrations/QRadar.md) |
| **Telegram** | [Telegram Bot](./integrations/Telegram.md) |
| **Email** | [Email Notifications](./integrations/Email.md) |
| **Slack** | [Slack Integration](./integrations/Slack.md) |

---

## 🎯 ЧТО ТАКОЕ PHANTOMPROXY?

**PhantomProxy Pro** — профессиональная платформа для симуляции Red Team атак и тестирования безопасности организации.

### Возможности

✅ **150+ функций**  
✅ **14 профессиональных модулей**  
✅ **SIEM интеграция** (Splunk, ELK, QRadar)  
✅ **2FA аутентификация**  
✅ **Командная работа**  
✅ **Billing и инвойсы**  
✅ **Автоматизация** (Scheduler, Auto-reports)  
✅ **White-label branding**  

### Для кого

- 🔹 Red Team операторы
- 🔹 Pentest компании
- 🔹 MSSP провайдеры
- 🔹 Enterprise security команды
- 🔹 Training организации

---

## 📊 АРХИТЕКТУРА

```
┌─────────────────────────────────────────────────────────────┐
│                    PHANTOMPROXY v12.5                       │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐   │
│  │  SIEM        │  │  Scheduler   │  │  Notifications  │   │
│  │  Integration │  │              │  │  (Email/TG)     │   │
│  └──────────────┘  └──────────────┘  └─────────────────┘   │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐   │
│  │  Security    │  │  Team        │  │  Analytics      │   │
│  │  (2FA)       │  │  Management  │  │  Dashboard      │   │
│  └──────────────┘  └──────────────┘  └─────────────────┘   │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐   │
│  │  Billing     │  │  Proposals   │  │  Reporting      │   │
│  │  & Invoices  │  │  & Contracts │  │  (PDF)          │   │
│  └──────────────┘  └──────────────┘  └─────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           Web Panel (Port 3000)                      │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           REST API (Port 8080)                       │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎓 ОБУЧЕНИЕ

### Уровень 1: Новичок

1. [Что такое PhantomProxy?](./What-is-PhantomProxy.md)
2. [Быстрый старт](./Quick-Start.md)
3. [Первая кампания](./First-Campaign.md)
4. [Просмотр сессий](./Viewing-Sessions.md)
5. [Генерация отчёта](./Generating-Reports.md)

### Уровень 2: Продвинутый

1. [Настройка кампаний](./Campaign-Setup.md)
2. [Работа с фишлетами](./Working-with-Phishlets.md)
3. [Командная работа](./Team-Management.md)
4. [Billing и инвойсы](./Billing-Invoicing.md)
5. [Настройка уведомлений](./Notifications-Setup.md)

### Уровень 3: Эксперт

1. [SIEM интеграция](./SIEM-Integration.md)
2. [API использование](./API-Usage.md)
3. [Автоматизация](./Automation.md)
4. [Кастомизация](./Customization.md)
5. [Production деплой](./Production-Deployment.md)

---

## ⚠️ ЮРИДИЧЕСКАЯ ИНФОРМАЦИЯ

**ВАЖНО:** Этот инструмент предназначен ТОЛЬКО для легального тестирования безопасности.

### Разрешено

✅ Тестирование с письменного разрешения (RoE)  
✅ Red Team операции по договору  
✅ Обучение по кибербезопасности  
✅ Исследовательские цели  

### Запрещено

❌ Несанкционированный доступ  
❌ Кража данных  
❌ Мошенничество  
❌ Любое использование без письменного разрешения  

**Подробнее:** [Ethical Boundaries](../docs/ETHICAL_BOUNDARIES.md)

---

## 📞 ПОДДЕРЖКА

### Контакты

- **Email:** support@phantomseclabs.com
- **GitHub Issues:** https://github.com/rpauts2/phantom-proxy/issues
- **Documentation:** https://github.com/rpauts2/phantom-proxy/wiki

### Ресурсы

- [Project Overview](../docs/PROJECT_OVERVIEW.md)
- [Security Policy](../SECURITY.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Code of Conduct](../CODE_OF_CONDUCT.md)

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

[Full Changelog](../CHANGELOG.md)

---

**© 2026 PhantomSec Labs. All rights reserved.**

**Last Updated:** February 20, 2026
