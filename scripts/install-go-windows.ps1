# PhantomProxy - Go Installation Script for Windows
# Run as Administrator: .\install-go-windows.ps1

$ErrorActionPreference = "Stop"
$GoVersion = "1.23.4"
$GoUrl = "https://go.dev/dl/go$GoVersion.windows-amd64.msi"
$InstallerPath = "$env:TEMP\go-$GoVersion-windows-amd64.msi"

Write-Host "PhantomProxy: Installing Go $GoVersion for Windows..." -ForegroundColor Cyan

# Check if Go already installed
$existing = Get-Command go -ErrorAction SilentlyContinue
if ($existing) {
    $ver = go version 2>$null
    Write-Host "Go already installed: $ver" -ForegroundColor Green
    Write-Host "To reinstall, uninstall first from Control Panel."
    exit 0
}

# Try winget first (Windows 11 / Win10 1809+)
$winget = Get-Command winget -ErrorAction SilentlyContinue
if ($winget) {
    Write-Host "Using winget to install Go..."
    winget install GoLang.Go --accept-package-agreements --accept-source-agreements
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Go installed via winget. Restart your terminal to use 'go' command." -ForegroundColor Green
        exit 0
    }
}

# Fallback: download MSI
Write-Host "Downloading Go from $GoUrl..."
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
Invoke-WebRequest -Uri $GoUrl -OutFile $InstallerPath -UseBasicParsing

Write-Host "Running installer (silent)..."
Start-Process msiexec.exe -ArgumentList "/i", $InstallerPath, "/quiet" -Wait

# Add to PATH for current session
$goPath = "C:\Program Files\Go\bin"
if (Test-Path $goPath) {
    $env:Path = "$goPath;$env:Path"
    Write-Host "Go installed to C:\Program Files\Go" -ForegroundColor Green
    Write-Host "RESTART YOUR TERMINAL or run: `$env:Path = 'C:\Program Files\Go\bin;' + `$env:Path" -ForegroundColor Yellow
    go version
} else {
    Write-Host "Installation completed. Restart terminal and run: go version" -ForegroundColor Yellow
}
