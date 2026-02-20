# 🧬 GAN OBFUSCATION MODULE

Динамическая обфускация кода через нейросеть

---

## 📋 ОПИСАНИЕ

GAN Obfuscation модуль который:

1. **Генерирует полиморфный код** — уникальная обфускация для каждой сессии
2. **Использует шаблоны мутаций** — variable rename, string transform, dead code
3. **Контролирует качество** — через discriminator
4. **Экспортирует в ONNX** — для быстрого инференса в Go

---

## 🚀 ВОЗМОЖНОСТИ

### ✅ Dynamic Obfuscation

- **Variable renaming** — `_0x{hex}` формат
- **String transformation** — fromCharCode, base64, hex
- **Dead code injection** — мёртвый код
- **Control flow** — обфускация потока управления

### ✅ Quality Control

- **Discriminator** — оценка качества обфускации
- **Confidence scoring** — уверенность в результате
- **Mutation tracking** — какие мутации применены

### ✅ Fast Inference

- **ONNX экспорт** — быстрая инференция
- **Go интеграция** — через onnxruntime_go
- **Session-based** — разные seed для разных сессий

---

## 📡 API ENDPOINTS

### POST /api/v1/gan/obfuscate

Обфускация кода.

**Request:**
```json
{
  "code": "var email = document.querySelector('#email').value;",
  "level": "high",
  "session_id": "abc123"
}
```

**Response:**
```json
{
  "success": true,
  "original_hash": "a1b2c3...",
  "obfuscated_code": "var _0x5a2b = String.fromCharCode(100,111,99)...",
  "mutations_applied": ["variable_rename", "string_transform", "dead_code"],
  "seed": 123456,
  "confidence": 0.95
}
```

### POST /api/v1/gan/train

Дообучение модели.

### GET /api/v1/gan/stats

Статистика GAN модуля.

---

## ⚙️ КОНФИГУРАЦИЯ

### config.yaml

```yaml
# GAN Obfuscation
gan_obfuscation:
  enabled: true
  default_level: "high"
  model_path: "./gan_models/obfuscator.onnx"
  auto_train: true
  train_interval: 86400  # раз в сутки
```

---

## 💡 ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ

### Обфускация JavaScript

```bash
curl -X POST http://localhost:8084/api/v1/gan/obfuscate \
  -H "Content-Type: application/json" \
  -d '{
    "code": "var password = document.querySelector(\"#pass\").value;",
    "level": "high",
    "session_id": "session-123"
  }'
```

### Результат:

**До:**
```javascript
var password = document.querySelector("#pass").value;
```

**После:**
```javascript
var _0x5a2b = String.fromCharCode(100,111,99,117,109,101,110,116)+".querySelector(\"#pass\").value";void 0;
if(true) { var _var7f = _0x5a2b; }
```

---

## 📈 МОНИТОРИНГ

### Метрики

- Количество обфускаций
- Среднее количество мутаций
- Качество обфускации
- Время обфускации

---

**Версия:** 1.0.0  
**Автор:** PhantomProxy Team  
**Лицензия:** MIT
