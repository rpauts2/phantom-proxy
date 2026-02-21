#!/bin/bash
# PhantomProxy v14.0 - Auto Installer for Linux
# One-command installation script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="${INSTALL_DIR:-/opt/phantomproxy}"
PYTHON_VERSION="${PYTHON_VERSION:-3.11}"
GO_VERSION="${GO_VERSION:-1.21}"
DOCKER_COMPOSE_VERSION="${DOCKER_COMPOSE_VERSION:-v2.24.0}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
██████╗  ██████╗ ██╗     ██╗     ██╗███╗   ██╗ ██████╗
██╔══██╗██╔═══██╗██║     ██║     ██║████╗  ██║██╔════╝
██████╔╝██║   ██║██║     ██║     ██║██╔██╗ ██║██║  ███╗
██╔═══╝ ██║   ██║██║     ██║     ██║██║╚██╗██║██║   ██║
██║     ╚██████╔╝███████╗███████╗██║██║ ╚████║╚██████╔╝
╚═╝      ╚═════╝ ╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝ ╚═════╝

        Enterprise Red Team Platform v14.0.0
        © 2026 PhantomSec Labs
EOF
    echo -e "${NC}\n"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "Please run as root (sudo ./install.sh)"
        exit 1
    fi
}

check_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
        log_info "Detected OS: $OS"
    else
        log_error "Cannot detect OS"
        exit 1
    fi
}

install_dependencies() {
    log_info "Installing system dependencies..."

    case $OS in
        *"Ubuntu"*|*"Debian"*)
            apt-get update
            apt-get install -y \
                curl wget git vim nano \
                build-essential libssl-dev \
                software-properties-common \
                apt-transport-https ca-certificates gnupg
            ;;
        *"CentOS"*|*"Fedora"*|*"RHEL"*)
            yum install -y \
                curl wget git vim nano \
                gcc gcc-c++ make openssl-devel \
                epel-release
            ;;
        *)
            log_warning "Unknown OS, trying basic installation..."
            ;;
    esac

    log_success "System dependencies installed"
}

install_docker() {
    if command -v docker &> /dev/null; then
        log_info "Docker already installed"
        return
    fi

    log_info "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh

    systemctl enable docker
    systemctl start docker

    log_success "Docker installed"
}

install_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        log_info "Docker Compose already installed"
        return
    fi

    log_info "Installing Docker Compose..."
    curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" \
        -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose

    log_success "Docker Compose installed"
}

install_go() {
    if command -v go &> /dev/null; then
        log_info "Go already installed ($(go version))"
        return
    fi

    log_info "Installing Go ${GO_VERSION}..."
    cd /tmp
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"

    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    echo 'export PATH=$PATH:$HOME/go/bin' >> /etc/profile
    source /etc/profile

    log_success "Go installed"
}

install_python() {
    if command -v python3 &> /dev/null; then
        log_info "Python already installed ($(python3 --version))"
        return
    fi

    log_info "Installing Python ${PYTHON_VERSION}..."

    case $OS in
        *"Ubuntu"*|*"Debian"*)
            add-apt-repository ppa:deadsnakes/ppa
            apt-get update
            apt-get install -y python${PYTHON_VERSION} python${PYTHON_VERSION}-venv python${PYTHON_VERSION}-dev
            ;;
        *)
            log_warning "Please install Python ${PYTHON_VERSION} manually"
            ;;
    esac

    log_success "Python installed"
}

create_directories() {
    log_info "Creating directory structure..."

    mkdir -p "$INSTALL_DIR"/{core,internal,cmd,ai_service,api,frontend,configs/phishlets,deploy,certs,logs}

    log_success "Directories created"
}

setup_python_env() {
    log_info "Setting up Python environment..."

    cd "$INSTALL_DIR"
    python3 -m venv venv
    source venv/bin/activate

    # Install Python dependencies
    if [ -f "requirements.txt" ]; then
        pip install -r requirements.txt
    fi

    log_success "Python environment setup"
}

setup_go_env() {
    log_info "Setting up Go environment..."

    cd "$INSTALL_DIR"
    go mod download

    log_success "Go environment setup"
}

generate_certs() {
    log_info "Generating SSL certificates..."

    cd "$INSTALL_DIR/certs"
    openssl req -x509 -newkey rsa:4096 \
        -keyout key.pem -out cert.pem \
        -days 365 -nodes \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=phantom.local"

    chmod 600 key.pem
    chmod 644 cert.pem

    log_success "SSL certificates generated"
}

create_config() {
    log_info "Creating configuration file..."

    cat > "$INSTALL_DIR/config.yaml" << 'EOF'
# PhantomProxy v14.0 Configuration
bind_ip: "0.0.0.0"
https_port: 8443
domain: "phantom.local"
cert_path: "./certs/cert.pem"
key_path: "./certs/key.pem"
database_path: "./phantom.db"
api_enabled: true
api_port: 8080
api_key: "change-me-to-secure-random-string"
debug: false
EOF

    log_success "Configuration created"
}

create_systemd_service() {
    log_info "Creating systemd service..."

    cat > /etc/systemd/system/phantomproxy.service << EOF
[Unit]
Description=PhantomProxy v14.0 - Enterprise Red Team Platform
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/phantom-proxy --config $INSTALL_DIR/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable phantomproxy

    log_success "Systemd service created"
}

build_binary() {
    log_info "Building PhantomProxy binary..."

    cd "$INSTALL_DIR"
    go build -ldflags="-s -w" -o phantom-proxy ./cmd/phantom-proxy-v14

    log_success "Binary built"
}

set_permissions() {
    log_info "Setting permissions..."

    chmod +x "$INSTALL_DIR/phantom-proxy"
    chmod 755 "$INSTALL_DIR"
    chmod -R 755 "$INSTALL_DIR/core"
    chmod -R 755 "$INSTALL_DIR/internal"

    log_success "Permissions set"
}

print_summary() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}          ${CYAN}INSTALLATION COMPLETED SUCCESSFULLY${NC}                  ${GREEN}║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Installation Directory:${NC} $INSTALL_DIR"
    echo -e "${BLUE}Version:${NC} 14.0.0"
    echo ""
    echo -e "${YELLOW}Quick Start:${NC}"
    echo "  Start:     sudo systemctl start phantomproxy"
    echo "  Status:    sudo systemctl status phantomproxy"
    echo "  Logs:      sudo journalctl -u phantomproxy -f"
    echo "  Stop:      sudo systemctl stop phantomproxy"
    echo ""
    echo -e "${YELLOW}Docker Compose:${NC}"
    echo "  Start:     cd $INSTALL_DIR && docker-compose up -d"
    echo "  Stop:      cd $INSTALL_DIR && docker-compose down"
    echo ""
    echo -e "${YELLOW}Web Interface:${NC}"
    echo "  Frontend:  http://localhost:3000"
    echo "  API:       http://localhost:8080"
    echo "  Proxy:     https://localhost:8443"
    echo ""
    echo -e "${CYAN}Documentation: $INSTALL_DIR/docs/README.md${NC}"
    echo ""
}

# Main
main() {
    print_banner

    log_info "Starting installation..."
    echo ""

    check_root
    check_os
    install_dependencies
    install_docker
    install_docker_compose
    install_go
    install_python
    create_directories
    setup_python_env
    setup_go_env
    generate_certs
    create_config
    build_binary
    create_systemd_service
    set_permissions

    print_summary

    log_success "Installation complete!"
}

# Run
main "$@"
