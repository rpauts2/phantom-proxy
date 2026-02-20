#!/bin/bash
# PhantomProxy AI Orchestrator Installation Script

set -e

echo "🚀 Installing PhantomProxy AI Orchestrator..."

# Проверка Python
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 not found. Installing..."
    sudo apt update
    sudo apt install -y python3 python3-pip python3-venv
fi

# Проверка pip
if ! command -v pip3 &> /dev/null; then
    echo "❌ pip3 not found. Installing..."
    sudo apt install -y python3-pip
fi

# Создание виртуального окружения
echo "📦 Creating virtual environment..."
cd "$(dirname "$0")"
python3 -m venv venv
source venv/bin/activate

# Установка зависимостей
echo "📦 Installing dependencies..."
pip install --upgrade pip
pip install -r requirements.txt

# Установка Playwright браузеров
echo "🌐 Installing Playwright browsers..."
playwright install chromium

# Проверка Ollama
echo "🤖 Checking Ollama..."
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollama not found. Installing..."
    curl -fsSL https://ollama.com/install.sh | sudo sh
    echo "✅ Ollama installed. Please run: ollama pull llama3.2"
else
    echo "✅ Ollama found"
fi

# Pull LLM модели
echo "📥 Pulling Llama 3.2 model..."
ollama pull llama3.2

# Создание systemd сервиса
echo "📝 Creating systemd service..."
sudo tee /etc/systemd/system/phantom-ai.service > /dev/null <<EOF
[Unit]
Description=PhantomProxy AI Orchestrator
After=network.target ollama.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$(pwd)
Environment="PATH=$(pwd)/venv/bin"
ExecStart=$(pwd)/venv/bin/python orchestrator.py
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Перезапуск systemd
echo "🔄 Reloading systemd..."
sudo systemctl daemon-reload
sudo systemctl enable phantom-ai

echo ""
echo "✅ Installation complete!"
echo ""
echo "To start the service:"
echo "  sudo systemctl start phantom-ai"
echo ""
echo "To check status:"
echo "  sudo systemctl status phantom-ai"
echo ""
echo "AI Orchestrator will be available at: http://localhost:8081"
