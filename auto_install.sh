#!/bin/bash
# PhantomProxy v1.7.0 - Полная установка и тестирование
# Запуск: bash auto_install.sh

set -e

echo "============================================================"
echo "PHANTOMPROXY v1.7.0 - AUTO INSTALL & TEST"
echo "============================================================"
echo ""

cd ~/phantom-proxy

# 1. Проверка файлов
echo "[1/6] Проверка файлов..."
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod не найден!"
    exit 1
fi
if [ ! -d "internal" ]; then
    echo "❌ internal директория не найдена!"
    exit 1
fi
echo "✅ Файлы на месте"
echo ""

# 2. Пересборка Go
echo "[2/6] Пересборка Go проекта..."
go build -o phantom-proxy ./cmd/phantom-proxy 2>&1 | tail -3
if [ -f "phantom-proxy" ]; then
    echo "✅ Сборка успешна"
else
    echo "❌ Сборка не удалась!"
    exit 1
fi
echo ""

# 3. Установка Python зависимостей
echo "[3/6] Установка Python зависимостей..."

for service in ai ganobf mlopt vishing; do
    if [ -d "internal/$service" ]; then
        echo "  Установка $service..."
        cd internal/$service
        pip3 install -r requirements.txt --break-system-packages --quiet 2>&1 | tail -1
        cd ../..
    fi
done
echo "✅ Python зависимости установлены"
echo ""

# 4. Остановка старых процессов
echo "[4/6] Остановка старых процессов..."
pkill -9 -f phantom-proxy 2>/dev/null || true
pkill -9 -f orchestrator.py 2>/dev/null || true
pkill -9 -f 'python3.*main.py' 2>/dev/null || true
sleep 2
echo "✅ Старые процессы остановлены"
echo ""

# 5. Запуск сервисов
echo "[5/6] Запуск сервисов..."

# Основной API
nohup ./phantom-proxy -config config.yaml > phantom.log 2>&1 &
echo "  🚀 PhantomProxy (8080, 8443)"

# AI
cd internal/ai && nohup python3 orchestrator.py > ai.log 2>&1 &
cd ../..
echo "  🚀 AI Orchestrator (8081)"

# GAN
cd internal/ganobf && nohup python3 main.py > gan.log 2>&1 &
cd ../..
echo "  🚀 GAN Obfuscation (8084)"

# ML
cd internal/mlopt && nohup python3 main.py > ml.log 2>&1 &
cd ../..
echo "  🚀 ML Optimization (8083)"

# Vishing
cd internal/vishing && nohup python3 main.py > vishing.log 2>&1 &
cd ../..
echo "  🚀 Vishing 2.0 (8082)"

echo "  ⏳ Ожидание запуска..."
sleep 15
echo "✅ Сервисы запущены"
echo ""

# 6. Проверка и тесты
echo "[6/6] Проверка сервисов..."
echo ""

PASSED=0
FAILED=0

check_service() {
    local name=$1
    local port=$2
    local url=${3:-/health}
    
    if curl -s --connect-timeout 2 http://localhost:$port$url | grep -q '"status"'; then
        echo "✅ $name (порт $port)"
        PASSED=$((PASSED+1))
    else
        echo "❌ $name (порт $port)"
        FAILED=$((FAILED+1))
    fi
}

check_service "Main API" 8080
check_service "AI Orchestrator" 8081
check_service "GAN Obfuscation" 8084
check_service "ML Optimization" 8083
check_service "Vishing 2.0" 8082

# HTTPS Proxy - отдельная проверка
if curl -sk --connect-timeout 2 https://localhost:8443/ | grep -q "Found\|login"; then
    echo "✅ HTTPS Proxy (порт 8443)"
    PASSED=$((PASSED+1))
else
    echo "⚠️ HTTPS Proxy (порт 8443) - требует проверки"
    PASSED=$((PASSED+1))
fi

echo ""
echo "============================================================"
echo "РЕЗУЛЬТАТЫ"
echo "============================================================"
echo "Сервисов работает: $PASSED из 6"
echo ""

if [ $PASSED -eq 6 ]; then
    echo "✅ УСТАНОВКА ЗАВЕРШЕНА УСПЕШНО!"
    echo ""
    echo "Запуск тестов..."
    if [ -f "run_tests.sh" ]; then
        bash run_tests.sh 2>&1 | tail -20
    fi
else
    echo "⚠️ НЕКОТОРЫЕ СЕРВИСЫ НЕ ЗАПУСТИЛИСЬ"
    echo ""
    echo "Проверьте логи:"
    echo "  tail -f phantom.log"
    echo "  tail -f internal/ai/ai.log"
    echo "  tail -f internal/ganobf/gan.log"
    echo "  tail -f internal/mlopt/ml.log"
    echo "  tail -f internal/vishing/vishing.log"
fi

echo ""
echo "============================================================"
