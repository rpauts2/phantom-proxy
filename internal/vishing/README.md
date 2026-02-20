# 🎤 VISHING 2.0 MODULE

Голосовые дипфейки для обхода 2FA

---

## 📋 ОПИСАНИЕ

Vishing 2.0 — модуль голосовых дипфейков который:

1. **Клонирует голос** — Coqui TTS
2. **Совершает звонки** — Twilio API
3. **Генерирует сценарии** — LLM (Llama 3.2)
4. **Записывает разговор** — для анализа

---

## 🚀 ВОЗМОЖНОСТИ

### ✅ Voice Cloning

- **Coqui TTS** — качественная синтезация
- **Клонирование** — из образца голоса
- **Эмоции** — neutral, happy, angry, sad
- **Мультиязычность** — en, es, fr, de, etc.

### ✅ Automated Calls

- **Twilio Integration** — звонки по всему миру
- **DTMF сбор** — ввод цифр жертвой
- **Speech recognition** — распознавание речи
- **Запись звонков** — сохранение в MP3

### ✅ Dynamic Scenarios

- **LLM генерация** — сценарии через Llama 3.2
- **Шаблоны** — Microsoft Support, Bank Security, etc.
- **Адаптивность** — изменение сценария в реальном времени

---

## 📡 API ENDPOINTS

### POST /api/v1/vishing/call

Совершение звонка.

**Request:**
```json
{
  "phone_number": "+1234567890",
  "voice_profile": "support_agent",
  "scenario": "microsoft_support",
  "custom_data": {}
}
```

**Response:**
```json
{
  "success": true,
  "call_id": "CA123...",
  "status": "initiated",
  "message": "Call initiated successfully",
  "recording_url": "https://api.twilio.com/..."
}
```

### GET /api/v1/vishing/call/:id

Статус звонка.

**Response:**
```json
{
  "call_id": "CA123...",
  "status": "completed",
  "recording_url": "https://...",
  "duration": 120
}
```

### POST /api/v1/vishing/generate-scenario

Генерация сценария через LLM.

**Request:**
```json
{
  "target_service": "Microsoft 365",
  "goal": "Get verification code"
}
```

**Response:**
```json
{
  "success": true,
  "scenario": {
    "name": "microsoft_365_support",
    "script": "Hello, this is Microsoft Support...",
    "target_prompt": "Please enter your verification code",
    "max_duration": 300
  }
}
```

### POST /api/v1/vishing/voice

Регистрация голосового профиля.

---

## 🔗 АРХИТЕКТУРА

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  PhantomProxy   │────▶│  Vishing Engine  │────▶│   Coqui TTS     │
│    (Go API)     │     │   (Python)       │     │  (Voice Clone)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌──────────────────┐
                        │     Twilio       │
                        │   (Voice Calls)  │
                        └──────────────────┘
                               │
                               ▼
                        ┌──────────────────┐
                        │    Victim's      │
                        │     Phone        │
                        └──────────────────┘
```

---

## ⚙️ КОНФИГУРАЦИЯ

### config.yaml

```yaml
# Vishing 2.0
vishing:
  enabled: true
  
  # Twilio
  twilio:
    account_sid: "AC..."
    auth_token: "..."
    phone_number: "+1234567890"
  
  # TTS
  tts:
    model: "tts_models/en/ljspeech/tacotron2-DDC"
    language: "en"
  
  # LLM
  llm:
    model: "llama3.2"
  
  # Voices directory
  voices_dir: "./voices"
  
  # Recordings directory
  recordings_dir: "./recordings"
```

---

## 🛡️ БЕЗОПАСНОСТЬ

### Anti-Detection

- **Caller ID spoofing** — подмена номера
- **Voice modulation** — изменение тембра
- **Background noise** — добавление шумов офиса

### Legal Compliance

⚠️ **WARNING:** Использование только для легального тестирования!

---

## 📈 МОНИТОРИНГ

### Метрики

- Количество звонков
- Средняя длительность
- Успешность (получение кода)
- Расходы на Twilio

### Логи

```bash
tail -f /var/log/phantom/vishing.log
```

---

## 🎯 ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ

### Пример 1: Звонок с готовым сценарием

```bash
curl -X POST http://localhost:8080/api/v1/vishing/call \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890",
    "voice_profile": "support_agent",
    "scenario": "microsoft_support"
  }'
```

### Пример 2: Генерация сценария

```bash
curl -X POST http://localhost:8080/api/v1/vishing/generate-scenario \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{
    "target_service": "Bank of America",
    "goal": "Get card PIN"
  }'
```

### Пример 3: Регистрация голоса

```bash
curl -X POST http://localhost:8082/api/v1/vishing/voice \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ceo_voice",
    "reference_audio": "./voices/ceo_sample.wav",
    "language": "en"
  }'
```

---

## 🐛 TROUBLESHOOTING

### Ошибка: "TTS model not found"

**Решение:** Запустить `pip install -r requirements.txt` и проверить наличие модели.

### Ошибка: "Twilio authentication failed"

**Решение:** Проверить TWILIO_ACCOUNT_SID и TWILIO_AUTH_TOKEN в .env.

### Ошибка: "No GPU found"

**Решение:** TTS работает и на CPU, но медленно. Для GPU установить CUDA.

---

## 📝 ЗАВИСИМОСТИ

### Python

```txt
TTS==0.21.0
twilio==8.11.0
ollama==0.1.7
fastapi==0.110.0
```

### Go

```go
github.com/playwright-community/playwright-go  // для browser pool
```

---

## 🎯 СЛЕДУЮЩИЕ ШАГИ

1. **Real-time voice conversion** — изменение голоса в реальном времени
2. **Multi-language support** — поддержка всех языков
3. **Emotion control** — управление эмоциями голоса
4. **Conference calls** — звонки с несколькими участниками

---

**Версия:** 1.0.0  
**Автор:** PhantomProxy Team  
**Лицензия:** MIT

⚠️ **Использовать только для легального тестирования!**
