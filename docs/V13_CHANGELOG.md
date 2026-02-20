# PhantomProxy v13 — Changelog

## Обзор

Версия v13 добавляет модульную архитектуру и расширенный функционал для симуляции целевых атак в рамках authorized security assessments.

## Новые модули

### 1. Event Bus и модульная архитектура
- `internal/events/bus.go` — шина событий для связи модулей
- События: `credential.captured`, `session.captured`, `cookie.captured`, `victim.landed`, `exfil.*`, `payload.generated`, `c2.beacon`

### 2. C2 Integration
- **Sliver** (`internal/c2/sliver.go`) — интеграция с Sliver C2 (REST/gRPC)
- **Cobalt Strike** (`internal/c2/cobaltstrike.go`) — External C2 / Beacon format
- **Empire** (`internal/c2/empire.go`) — REST API
- **HTTP Callback** (`internal/c2/http_callback.go`) — generic HTTP/S callback
- **DNS Tunnel** (`internal/c2/dns_tunnel.go`) — эксфильтрация через DNS

При перехвате кредов и сессий данные автоматически отправляются во все включённые C2.

### 3. Credential Stuffing
- `internal/credentialstuffing/stuffing.go` — проверка кредов на других сервисах
- Rate limiting, настраиваемые целевые сервисы

### 4. HIBP (Have I Been Pwned)
- `internal/credentialstuffing/hibp.go` — k-anonymity API для проверки паролей в утечках
- Password spraying automation

### 5. Payload Generator
- `internal/payload/generator.go` — оркестратор msfvenom
- Поддержка: Windows EXE/DLL, shellcode, PowerShell
- Vuln scanner (делегирует nmap)

### 6. Evasion Config
- `internal/evasion/config.go` — параметры для внешних инструментов (Sliver, Havoc)
- Sleep obfuscation, sandbox evasion, AMSI/ETW, process injection (передача в конфиг импланта)

### 7. Exfiltration Simulation
- `internal/exfiltration/simulator.go` — симуляция эксфильтрации для DLP-тестов
- Metadata-only режим, поддержка облачных провайдеров

### 8. Social Engineering
- `internal/social/automation.go` — шаблоны писем, профилирование целей, rate-limited рассылка

## Конфигурация

Секция `v13` в `config.yaml`:

```yaml
v13:
  c2:
    sliver: { enabled, server_url, operator_token, callback_host }
    http_callback: { enabled, callback_url, headers }
    dns_tunnel: { enabled, domain, chunk_size }
  credential_stuffing: { enabled, targets, rate_limit }
  hibp: { enabled, api_key }
  payload: { enabled, msfvenom_path, output_dir }
  evasion: { sleep_obfuscation, sandbox_evasion, amsi_bypass, etw_patch, process_injection }
  exfiltration: { enabled, target_types, max_size_mb, cloud_provider }
  social_engineering: { enabled, smtp_host, smtp_port, rate_limit }
```

## Новые идеи (дополнительно)

- **Modular implant config** — передача evasion-параметров в Sliver/Havoc
- **Campaign analytics** — метрики по фишинговым кампаниям
- **Credential waterfall** — автоматическая проверка кредов по цепочке сервисов
- **DLP trigger testing** — генерация fake payload для проверки DLP
