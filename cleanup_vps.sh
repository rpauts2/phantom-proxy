#!/bin/bash
# Очистка VPS - Удалить всё кроме PhantomProxy

echo "⚠️ ВНИМАНИЕ! Удаление всех файлов кроме PhantomProxy..."
read -p "Продолжить? (y/n): " confirm

if [ "$confirm" != "y" ]; then
    echo "Отменено"
    exit 0
fi

echo ""
echo "Очистка..."

# Остановка всех процессов
pkill -f python3 2>/dev/null || true
pkill -f node 2>/dev/null || true
pkill -f npm 2>/dev/null || true

# Удаление мусора
rm -rf ~/node_modules 2>/dev/null || true
rm -rf ~/npm-debug.log 2>/dev/null || true
rm -rf /tmp/* 2>/dev/null || true
rm -rf ~/.cache/pip 2>/dev/null || true

# Очистка apt кэша
sudo apt clean 2>/dev/null || true
sudo apt autoremove -y 2>/dev/null || true

# Проверка места
echo ""
echo "Очистка завершена!"
echo ""
echo "Место на диске:"
df -h /

echo ""
echo "PhantomProxy файлы:"
ls -la ~/phantom-proxy/ 2>/dev/null || echo "PhantomProxy не найден"

echo ""
echo "✅ ГОТОВО!"
