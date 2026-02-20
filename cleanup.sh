#!/bin/bash
# PhantomProxy - Очистка VPS от мусора
# Выполнить на сервере: bash cleanup.sh

echo "============================================================"
echo "ОЧИСТКА VPS ОТ МУСОРА"
echo "============================================================"
echo ""

# 1. Остановка процессов
echo "[1/6] Остановка процессов..."
pkill -9 -f phantom-proxy 2>/dev/null || true
pkill -9 -f orchestrator.py 2>/dev/null || true
pkill -9 -f 'python3.*main.py' 2>/dev/null || true
pkill -9 -f 'python3.*app.py' 2>/dev/null || true
pkill -9 -f node 2>/dev/null || true
echo "✅ Процессы остановлены"

# 2. Удаление phantom-proxy
echo ""
echo "[2/6] Удаление phantom-proxy..."
rm -rf ~/phantom-proxy
echo "✅ phantom-proxy удалён"

# 3. Очистка pip кэша
echo ""
echo "[3/6] Очистка pip кэша..."
rm -rf ~/.cache/pip
echo "✅ Pip кэш очищен"

# 4. Очистка go кэша
echo ""
echo "[4/6] Очистка go кэша..."
go clean -cache -modcache -i -r 2>/dev/null || true
echo "✅ Go кэш очищен"

# 5. Удаление временных файлов
echo ""
echo "[5/6] Удаление временных файлов..."
rm -f /tmp/*.log /tmp/*.tmp 2>/dev/null || true
rm -rf /tmp/phantom-* 2>/dev/null || true
echo "✅ Временные файлы удалены"

# 6. Проверка места
echo ""
echo "[6/6] Проверка места..."
echo ""
df -h / 
echo ""

# Дополнительные рекомендации
echo "============================================================"
echo "ДОПОЛНИТЕЛЬНЫЕ РЕКОМЕНДАЦИИ"
echo "============================================================"
echo ""

# Проверка больших файлов
echo "Большие файлы (>100MB):"
find /home/ubuntu -type f -size +100M 2>/dev/null | head -10

echo ""
echo "Старые логи:"
find /var/log -name "*.log.*" -mtime +7 2>/dev/null | head -10

echo ""
echo "============================================================"
echo "✅ ОЧИСТКА ЗАВЕРШЕНА"
echo "============================================================"
echo ""
echo "Теперь можешь установить PhantomProxy заново:"
echo "  cd ~"
echo "  mkdir phantom-proxy"
echo "  # Загрузить файлы"
echo "  bash final_install.sh"
echo ""
