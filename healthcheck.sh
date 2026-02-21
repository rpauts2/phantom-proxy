#!/bin/bash
# PhantomProxy v14.0 - Health Check Script
# Use in production monitoring

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
PROXY_URL="${PROXY_URL:-https://localhost:8443}"
API_URL="${API_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
TIMEOUT=5

# Counters
PASS=0
FAIL=0

# Functions
check_service() {
    local name=$1
    local url=$2
    
    if curl -sf --max-time $TIMEOUT "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $name is UP"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} $name is DOWN"
        ((FAIL++))
    fi
}

check_port() {
    local name=$1
    local port=$2
    
    if nc -z localhost $port 2>/dev/null; then
        echo -e "${GREEN}✓${NC} $port ($name) is listening"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} $port ($name) is NOT listening"
        ((FAIL++))
    fi
}

check_process() {
    local name=$1
    
    if pgrep -f "$name" > /dev/null; then
        echo -e "${GREEN}✓${NC} $name is running"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} $name is NOT running"
        ((FAIL++))
    fi
}

# Main
echo "╔══════════════════════════════════════════════════════════╗"
echo "║     PhantomProxy v14.0 - Health Check                   ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

echo "Checking Processes..."
check_process "phantom-proxy"
check_process "docker-compose"

echo ""
echo "Checking Ports..."
check_port "HTTPS Proxy" "8443"
check_port "API" "8080"
check_port "Frontend" "3000"
check_port "Redis" "6379"
check_port "PostgreSQL" "5432"

echo ""
echo "Checking Services..."
check_service "API Health" "$API_URL/health"
check_service "Frontend" "$FRONTEND_URL"

echo ""
echo "Checking Docker Containers..."
if command -v docker &> /dev/null; then
    docker ps --format "table {{.Names}}\t{{.Status}}" 2>/dev/null || true
fi

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo -e "║  ${GREEN}PASSED:${NC} $PASS  ${RED}FAILED:${NC} $FAIL                            ║"
echo "╚══════════════════════════════════════════════════════════╝"

if [ $FAIL -gt 0 ]; then
    exit 1
fi

exit 0
