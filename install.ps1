# PhantomProxy v14.0 - Auto Installer for Windows
# PowerShell one-command installation script

param(
    [string]$InstallDir = "C:\PhantomProxy",
    [switch]$NoDocker,
    [switch]$SkipPython
)

# Colors
function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[✓] $args" -ForegroundColor Green }
function Write-Warning { Write-Host "[!] $args" -ForegroundColor Yellow }
function Write-Error { Write-Host "[✗] $args" -ForegroundColor Red }

# Banner
function Print-Banner {
    Write-Host @"

██████╗  ██████╗ ██╗     ██╗     ██╗███╗   ██╗ ██████╗
██╔══██╗██╔═══██╗██║     ██║     ██║████╗  ██║██╔════╝
██████╔╝██║   ██║██║     ██║     ██║██╔██╗ ██║██║  ███╗
██╔═══╝ ██║   ██║██║     ██║     ██║██║╚██╗██║██║   ██║
██║     ╚██████╔╝███████╗███████╗██║██║ ╚████║╚██████╔╝
╚═╝      ╚═════╝ ╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝ ╚═════╝

        Enterprise Red Team Platform v14.0.0
        © 2026 PhantomSec Labs

"@ -ForegroundColor Cyan
    Write-Host ""
}

# Check Admin
function Check-Admin {
    $isAdmin = ([Security.Principal.WindowsPrincipal] `
        [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(`
        [Security.Principal.WindowsBuiltInRole]::Administrator)

    if (-not $isAdmin) {
        Write-Error "Please run as Administrator"
        Write-Info "Right-click PowerShell and select 'Run as Administrator'"
        exit 1
    }
}

# Install Chocolatey
function Install-Chocolatey {
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        Write-Info "Chocolatey already installed"
        return
    }

    Write-Info "Installing Chocolatey..."
    Set-ExecutionPolicy Bypass -Scope Process -Force
    [System.Net.ServicePointManager]::SecurityProtocol = `
        [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
    iex ((New-Object System.Net.WebClient).DownloadString(`
        'https://community.chocolatey.org/install.ps1'))
    Write-Success "Chocolatey installed"
}

# Install Dependencies
function Install-Dependencies {
    Write-Info "Installing dependencies..."

    choco install -y git
    choco install -y openssl
    choco install -y docker-desktop
    choco install -y golang
    choco install -y python3
    choco install -y nodejs-lts
    choco install -y visualstudio2022buildtools

    Write-Success "Dependencies installed"
}

# Install Docker
function Install-Docker {
    if ($NoDocker) {
        Write-Info "Skipping Docker installation"
        return
    }

    if (Get-Command docker -ErrorAction SilentlyContinue) {
        Write-Info "Docker already installed"
        return
    }

    Write-Info "Installing Docker Desktop..."
    choco install -y docker-desktop
    Write-Success "Docker Desktop installed"

    # Start Docker
    Start-Service "Docker Desktop Service"
}

# Create Directories
function Create-Directories {
    Write-Info "Creating directory structure..."

    $dirs = @(
        "core", "internal", "cmd", "ai_service", "api",
        "frontend", "configs\phishlets", "deploy", "certs", "logs"
    )

    foreach ($dir in $dirs) {
        New-Item -ItemType Directory -Force -Path "$InstallDir\$dir" | Out-Null
    }

    Write-Success "Directories created"
}

# Setup Python
function Setup-Python {
    if ($SkipPython) {
        Write-Info "Skipping Python setup"
        return
    }

    Write-Info "Setting up Python environment..."

    Set-Location "$InstallDir"
    python -m venv venv
    .\venv\Scripts\Activate.ps1

    if (Test-Path "requirements.txt") {
        pip install -r requirements.txt
    }

    Write-Success "Python environment setup"
}

# Setup Go
function Setup-Go {
    Write-Info "Setting up Go environment..."

    Set-Location "$InstallDir"
    go mod download

    Write-Success "Go environment setup"
}

# Generate Certificates
function Generate-Certs {
    Write-Info "Generating SSL certificates..."

    Set-Location "$InstallDir\certs"

    openssl req -x509 -newkey rsa:4096 `
        -keyout key.pem -out cert.pem `
        -days 365 -nodes `
        -subj "/C=US/ST=State/L=City/O=Organization/CN=phantom.local"

    icacls key.pem /grant "$($env:USERNAME):R" | Out-Null
    icacls cert.pem /grant "$($env:USERNAME):R" | Out-Null

    Write-Success "SSL certificates generated"
}

# Create Config
function Create-Config {
    Write-Info "Creating configuration file..."

    $config = @"
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
"@

    $config | Out-File -FilePath "$InstallDir\config.yaml" -Encoding UTF8

    Write-Success "Configuration created"
}

# Build Binary
function Build-Binary {
    Write-Info "Building PhantomProxy binary..."

    Set-Location "$InstallDir"
    go build -ldflags="-s -w" -o phantom-proxy.exe ./cmd/phantom-proxy-v14

    Write-Success "Binary built"
}

# Create Shortcut
function Create-Shortcut {
    Write-Info "Creating desktop shortcut..."

    $WshShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WshShell.CreateShortcut("$Home\Desktop\PhantomProxy.lnk")
    $Shortcut.TargetPath = "$InstallDir\phantom-proxy.exe"
    $Shortcut.WorkingDirectory = "$InstallDir"
    $Shortcut.Description = "PhantomProxy v14.0 - Enterprise Red Team Platform"
    $Shortcut.Save()

    Write-Success "Desktop shortcut created"
}

# Add to PATH
function Add-ToPath {
    Write-Info "Adding to PATH..."

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$currentPath;$InstallDir",
            "User"
        )
    }

    Write-Success "Added to PATH"
}

# Print Summary
function Print-Summary {
    Write-Host ""
    Write-Host "╔══════════════════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║          INSTALLATION COMPLETED SUCCESSFULLY             ║" -ForegroundColor Green
    Write-Host "╚══════════════════════════════════════════════════════════╝" -ForegroundColor Green
    Write-Host ""
    Write-Host "Installation Directory: $InstallDir" -ForegroundColor Cyan
    Write-Host "Version: 14.0.0" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Quick Start:" -ForegroundColor Yellow
    Write-Host "  Start:     .\phantom-proxy.exe --config config.yaml"
    Write-Host "  Console:   python console.py"
    Write-Host "  Docker:    docker-compose up -d"
    Write-Host ""
    Write-Host "Web Interface:" -ForegroundColor Yellow
    Write-Host "  Frontend:  http://localhost:3000"
    Write-Host "  API:       http://localhost:8080"
    Write-Host "  Proxy:     https://localhost:8443"
    Write-Host ""
    Write-Host "Documentation: $InstallDir\docs\README.md" -ForegroundColor Cyan
    Write-Host ""
}

# Main
function Main {
    Print-Banner

    Write-Info "Starting installation..."
    Write-Host ""

    Check-Admin
    Install-Chocolatey
    Install-Dependencies
    Install-Docker
    Create-Directories
    Setup-Python
    Setup-Go
    Generate-Certs
    Create-Config
    Build-Binary
    Create-Shortcut
    Add-ToPath

    Print-Summary

    Write-Success "Installation complete!"
    Write-Info "Please restart your computer for all changes to take effect"
}

# Run
Main
