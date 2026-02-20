# 👻 PHANTOMPROXY v12.2 PRO+ — COMPLETE EDITION

**Профессиональная Red Team Simulation Platform**

**© 2026 PhantomSec Labs. All rights reserved.**

---

## 🎯 НОВОЕ В v12.2 PRO+

### Добавленные модули:

| Модуль | Файл | Назначение |
|--------|------|------------|
| **Billing** | `v12_billing.py` | Счета и invoices |
| **Proposals** | `v12_proposals.py` | Коммерческие предложения |
| **Contracts** | `v12_proposals.py` | Контракты и RoE |
| **NDA Generator** | `v12_proposals.py` | Соглашения о конфиденциальности |

---

## 📁 ПОЛНАЯ СТРУКТУРА

```
~/phantom-proxy/
├── phantomproxy_v12_1_pro.py    # Главная программа
├── v12_billing.py               # ✅ Billing Module
├── v12_proposals.py             # ✅ Proposals & Contracts
├── v12_reporting.py             # Reporting Engine
├── v12_compliance.py            # Compliance Logging
├── phantom.db                   # База данных
├── branding/
│   ├── logo.png                 # Логотип
│   └── favicon.ico              # Favicon
├── invoices/                    # ✅ Счета (PDF)
├── proposals/                   # ✅ Предложения (PDF)
├── contracts/                   # ✅ Контракты (PDF)
├── reports/                     # Отчёты (PDF)
├── evidence/                    # Доказательства
└── compliance_logs/             # Audit логи
```

---

## 💳 BILLING MODULE

### Генерация счёта:

```python
from v12_billing import InvoiceGenerator

generator = InvoiceGenerator()

# Создать счёт
invoice = generator.create_invoice(
    client_id=1,
    campaign_id=1,
    rate=500,  # $500/hour
    due_days=30
)

print(f"Invoice: {invoice['invoice_number']}")
print(f"Total: ${invoice['total']:,.2f}")
print(f"PDF: {invoice['pdf_path']}")
```

### Функции:

- ✅ Авто-расчёт часов кампании
- ✅ Генерация PDF invoice
- ✅ Реквизиты компании
- ✅ Статусы оплаты
- ✅ Трекинг в БД

### Пример invoice:

```
════════════════════════════════════════
INVOICE
════════════════════════════════════════
Invoice Number: INV-0001
Issue Date: 2026-02-20
Due Date: 2026-03-22
Status: Pending
────────────────────────────────────────
Bill To:
Test Client
test@client.com
────────────────────────────────────────
Services Rendered:
Description              Hours   Rate    Amount
Red Team Testing         40      $500    $20,000
────────────────────────────────────────
Total: $20,000
────────────────────────────────────────
Payment Information:
Bank: ПАО Сбербанк
BIK: 044525225
Account: 40702810XXXXXXXXXXXXX
════════════════════════════════════════
```

---

## 📄 PROPOSAL GENERATOR

### Генерация предложения:

```python
from v12_proposals import ProposalGenerator

generator = ProposalGenerator()

proposal = generator.generate_proposal(
    client_name='ACME Corp',
    service_type='Red Team Assessment',
    duration='4 weeks',
    price=50000  # $50,000
)

print(f"Proposal: {proposal}")
```

### Включает:

- ✅ Executive Summary
- ✅ Scope of Work
- ✅ Timeline
- ✅ Pricing
- ✅ Terms & Conditions
- ✅ Contact Info

---

## 📋 CONTRACT GENERATOR

### Rules of Engagement (RoE):

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

### Non-Disclosure Agreement (NDA):

```python
nda = generator.generate_nda(client_name='ACME Corp')
print(f"NDA: {nda}")
```

---

## 🚀 БЫСТРЫЙ СТАРТ

### 1. Установка зависимостей:

```bash
cd ~/phantom-proxy
pip install reportlab cryptography
```

### 2. Запуск главной программы:

```bash
python3 phantomproxy_v12_1_pro.py
```

### 3. Меню:

```
======================================================================
  👻 PhantomSec Labs
  🚀 PhantomProxy Pro — Red Team Simulation Platform
======================================================================

  📌 MAIN MENU:
  1. 🚀 Start All Services
  2. 🛑 Stop
  3. 📊 View Status
  4. 📈 View Statistics
  5. 🎯 Create Campaign
  6. 📋 View Sessions
  7. 📄 Generate Report (PDF)
  8. ⚖️  Compliance Logs
  9. 👥 Client Portal
 10. 💳 Billing
 11. 📄 Generate Proposal
 12. 📋 Generate Contract/RoE
 13. 🚪 Exit
======================================================================
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

### API:
```
http://212.233.93.147:8080/health
http://212.233.93.147:8080/api/v1/stats
```

---

## 📊 WORKFLOW

### 1. Продажа проекта:

```
Client → Proposal (PDF) → Contract (PDF) → RoE (PDF) → NDA (PDF) → Deposit
```

### 2. Выполнение:

```
Campaign Setup → Phishing Attack → Credential Capture → Reporting
```

### 3. Завершение:

```
Final Report (PDF) → Invoice (PDF) → Payment → Retesting (optional)
```

---

## 🎨 BRANDING

### Настройки:

```python
# В phantomproxy_v12_1_pro.py
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

### В отчётах:

- ✅ Логотип в header
- ✅ Футер с контактами
- ✅ Копирайт компании
- ✅ Цветовая схема

---

## ⚖️ COMPLIANCE

### Логирование:

- ✅ Все действия
- ✅ Шифрование (Fernet)
- ✅ Hash verification
- ✅ Audit trail

### Scope Enforcement:

- ✅ Domain blacklist/whitelist
- ✅ Campaign limits
- ✅ Time restrictions
- ✅ Auto-kill switch

---

## 🛡️ БЕЗОПАСНОСТЬ

### Доступ:

- ✅ Hashed passwords (SHA-256)
- ✅ API keys
- ✅ 2FA (TOTP) — в разработке
- ✅ IP whitelisting — опционально

### Данные:

- ✅ Шифрование логов
- ✅ Auto-delete — опционально
- ✅ Data retention policies

---

## 📝 ШАБЛОНЫ

### Proposal Template:

```
1. Executive Summary
2. Scope of Work
3. Timeline
4. Investment
5. Terms & Conditions
6. Contact Information
```

### RoE Template:

```
1. Project Information
2. Authorized Activities
3. Prohibited Activities
4. Emergency Contacts
5. Signatures
```

### Invoice Template:

```
1. Invoice Number & Dates
2. Bill To
3. Services Rendered
4. Payment Information
5. Terms
```

---

## 🎯 ИТОГИ

**v12.2 PRO+ включает:**

✅ v12.1 PRO (Branding + White Label)  
✅ Billing & Invoicing  
✅ Proposal Generator  
✅ Contract Generator  
✅ RoE Generator  
✅ NDA Generator  
✅ Client Portal  
✅ Compliance Logging  
✅ Professional Reporting  

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

**v12.2 PRO+ — ГОТОВО К ПРОДАЖЕ!** 🚀

**© 2026 PhantomSec Labs. All rights reserved.**
