# 🔄 DOMAIN ROTATION MODULE

Автоматическая ротация доменов для PhantomProxy

---

## 📋 ОПИСАНИЕ

Модуль автоматической ротации доменов который:

1. **Регистрирует новые домены** через Namecheap/GoDaddy API
2. **Настраивает DNS** записи автоматически
3. **Получает SSL** сертификаты через Let's Encrypt
4. **Ротирует домены** по расписанию или при блокировке

---

## 🚀 ВОЗМОЖНОСТИ

### ✅ Автоматическая регистрация

- Поддержка Namecheap API
- Поддержка GoDaddy API (TODO)
- Генерация случайных поддоменов
- Проверка доступности домена

### ✅ DNS Автоматизация

- Интеграция с Cloudflare DNS
- Интеграция с Route53 (TODO)
- Автоматическая настройка A/CNAME записей

### ✅ SSL Сертификаты

- Let's Encrypt через lego (ACME)
- Автоматическое продление
- Хранение в SQLite

### ✅ Ротация

- По расписанию (каждые N часов)
- При детекте блокировки
- Рандомизация времени ротации

---

## 📡 API ENDPOINTS

### POST /api/v1/domains/register

Регистрация нового домена.

**Request:**
```json
{
  "base_domain": "verdebudget.ru",
  "years": 1
}
```

**Response:**
```json
{
  "success": true,
  "domain": "app-a1b2c3d4e5f6.verdebudget.ru",
  "status": "registered",
  "ssl_status": "valid"
}
```

### POST /api/v1/domains/rotate

Принудительная ротация домена.

**Response:**
```json
{
  "success": true,
  "old_domain": "app-old.verdebudget.ru",
  "new_domain": "web-new.verdebudget.ru",
  "message": "Domain rotated successfully"
}
```

### GET /api/v1/domains

Список всех доменов.

**Response:**
```json
{
  "domains": [
    {
      "domain": "app-abc123.verdebudget.ru",
      "status": "active",
      "ssl_status": "valid",
      "dns_status": "configured",
      "created_at": "2026-02-18T12:00:00Z",
      "expires_at": "2027-02-18T12:00:00Z"
    }
  ],
  "total": 1,
  "current_domain": "app-abc123.verdebudget.ru"
}
```

---

## ⚙️ КОНФИГУРАЦИЯ

### config.yaml

```yaml
# Domain Rotation
domain_rotation:
  enabled: true
  registrar: namecheap  # namecheap, godaddy
  min_domain_age: 60    # минут
  max_domain_age: 1440  # минут (24 часа)
  auto_renew_before: 24 # часов до истечения SSL
  max_domains: 10
  
  # Namecheap
  namecheap:
    api_key: "YOUR_API_KEY"
    api_user: "YOUR_API_USER"
    client_ip: "YOUR_SERVER_IP"
    use_sandbox: true
  
  # SSL
  ssl:
    provider: letsencrypt
    email: "admin@verdebudget.ru"
    storage_path: "./certs"
```

---

## 🔗 ИНТЕГРАЦИЯ

### Через API

```bash
# Регистрация домена
curl -X POST http://localhost:8080/api/v1/domains/register \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{"base_domain": "verdebudget.ru"}'

# Ротация
curl -X POST http://localhost:8080/api/v1/domains/rotate \
  -H "Authorization: Bearer verdebudget-secret-2026"

# Список доменов
curl http://localhost:8080/api/v1/domains \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

### Через Go код

```go
import "github.com/phantom-proxy/phantom-proxy/internal/domain"

// Создание ротатора
rotator := domain.NewDomainRotator(&domain.RotatorConfig{
    RegistrarName:     "namecheap",
    RegistrarAPIKey:   "YOUR_KEY",
    MinDomainAge:      60,
    MaxDomainAge:      1440,
    AutoRenewBefore:   24,
    MaxDomains:        10,
}, logger)

// Запуск
rotator.Start(ctx)

// Регистрация домена
newDomain, err := rotator.RegisterDomain(ctx, "verdebudget.ru")

// Ротация
newDomain, err := rotator.RotateDomain(ctx, "verdebudget.ru")

// Получение текущего домена
current := rotator.GetCurrentDomain()
```

---

## 🛡️ БЕЗОПАСНОСТЬ

### API Key Protection

- Namecheap API ключи шифруются
- Хранение в encrypted vault
- Доступ только из whitelisted IP

### Rate Limiting

- Максимум 10 доменов в час
- Максимум 100 доменов всего

### DNS Protection

- Использование Cloudflare Proxy
- Скрытие реального IP сервера

---

## 📈 МОНИТОРИНГ

### Метрики

- Количество активных доменов
- Время до следующей ротации
- Статус SSL сертификатов
- Количество заблокированных доменов

### Логи

```bash
tail -f /var/log/phantom/domain-rotation.log
```

---

## 🐛 TROUBLESHOOTING

### Ошибка: "Max domains limit reached"

**Решение:** Увеличить `max_domains` в конфиге или удалить старые домены.

### Ошибка: "Failed to obtain SSL"

**Решение:** Проверить что порт 80 открыт для HTTP challenge.

### Ошибка: "Namecheap API authentication failed"

**Решение:** Проверить API ключ и Client IP в настройках Namecheap.

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

1. **GoDaddy интеграция** — поддержка второго регистратора
2. **Cloudflare DNS** — автоматическая настройка DNS
3. **Auto-block detection** — ротация при детекте блокировки
4. **Domain health check** — проверка доступности доменов

---

**Версия:** 1.0.0  
**Автор:** PhantomProxy Team  
**Лицензия:** MIT
