# PowerShell installation script for fancy-login Go version
# Run this script from the go/ directory

param(
    [switch]$Verbose = $false
)

function Write-FancyLog {
    param([string]$Message)
    if ($Verbose) {
        Write-Host "[fancy-install-go] $Message" -ForegroundColor Cyan
    }
}

Write-Host "🔍 Checking for required tools..." -ForegroundColor Yellow

$RequiredTools = @("aws", "kubectl", "fzf", "k9s", "go")
$MissingTools = @()

foreach ($tool in $RequiredTools) {
    try {
        $null = Get-Command $tool -ErrorAction Stop
        Write-Host "✅ $tool is installed." -ForegroundColor Green
    }
    catch {
        Write-Host "❌ $tool is missing." -ForegroundColor Red
        $MissingTools += $tool
    }
}

if ($MissingTools.Count -gt 0) {
    Write-Host "❌ Missing required tools:" -ForegroundColor Red
    foreach ($tool in $MissingTools) {
        Write-Host "   - $tool" -ForegroundColor Red
    }
    Write-Host "`nPlease install the missing tools before running the install script." -ForegroundColor Yellow
    Write-Host "Note: Go is required to build the fancy-login binary." -ForegroundColor Yellow
    Write-Host "`nInstallation suggestions:"
    Write-Host "- Install Go from: https://golang.org/dl/"
    Write-Host "- Install other tools via Chocolatey, Scoop, or direct downloads"
    exit 1
}

Write-Host "✅ All required tools are installed." -ForegroundColor Green
Write-Host "🔧 Installing fancy-login Go version..." -ForegroundColor Yellow

# Define paths
$ScriptDir = $PSScriptRoot
$ProjectDir = Split-Path $ScriptDir -Parent

# Determine user home directory based on OS
if ($IsWindows -or ($null -eq $IsWindows -and $env:OS -eq "Windows_NT")) {
    $UserHome = $env:USERPROFILE
    $BinDir = Join-Path $UserHome "AppData\Local\fancy-login"
} else {
    # Non-Windows (macOS, Linux)
    $UserHome = $env:HOME
    $BinDir = Join-Path $UserHome ".local/bin/fancy-login"
}

$AWSDir = Join-Path $UserHome ".aws"
$KubeDir = Join-Path $UserHome ".kube"

Write-FancyLog "Script Dir: $ScriptDir"
Write-FancyLog "Project Dir: $ProjectDir"
Write-FancyLog "User Home: $UserHome"
Write-FancyLog "Bin Dir: $BinDir"

# Create bin dir if needed
Write-FancyLog "Creating bin dir at $BinDir"
if (-not (Test-Path $BinDir)) {
    try {
        $result = New-Item -ItemType Directory -Path $BinDir -Force
        Write-FancyLog "Created directory: $($result.FullName)"
    } catch {
        Write-Error "Failed to create directory $BinDir`: $_"
        exit 1
    }
} else {
    Write-FancyLog "Directory already exists: $BinDir"
}

# Verify directory exists before proceeding
if (-not (Test-Path $BinDir)) {
    Write-Error "Directory $BinDir does not exist after creation attempt"
    exit 1
}
Write-FancyLog "Confirmed directory exists: $BinDir"

# Build the Go binary
Write-Host "🔨 Building Go binary..." -ForegroundColor Yellow
Write-FancyLog "Building from $ProjectDir"
Set-Location $ProjectDir

# Ensure go.mod exists and is correct
if (-not (Test-Path "go.mod")) {
    Write-FancyLog "Initializing go module"
    go mod init fancy-login
}

# Build the binary
Write-FancyLog "Building binary to $BinDir\fancy-login-go.exe"
$buildResult = go build -o "$BinDir\fancy-login-go.exe" .\cmd
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build Go binary"
    exit 1
}

# Copy config files
Write-FancyLog "Copying .fancy-namespaces.conf to $BinDir\.fancy-namespaces.conf"
if (Test-Path "$ProjectDir\.fancy-namespaces.conf") {
    Copy-Item "$ProjectDir\.fancy-namespaces.conf" "$BinDir\.fancy-namespaces.conf" -Force
} else {
    Write-FancyLog "Warning: .fancy-namespaces.conf not found, skipping"
}

if (Test-Path "$ProjectDir\.fancy-contexts.conf") {
    Write-FancyLog "Copying .fancy-contexts.conf to $BinDir\.fancy-contexts.conf"
    Copy-Item "$ProjectDir\.fancy-contexts.conf" "$BinDir\.fancy-contexts.conf" -Force
} else {
    Write-FancyLog "Warning: .fancy-contexts.conf not found, skipping"
}

# Copy shell integration files
Write-FancyLog "Copying shell integration files"
if (Test-Path "$ScriptDir\fancy-go.ps1") {
    Copy-Item "$ScriptDir\fancy-go.ps1" "$BinDir\fancy-go.ps1" -Force
} else {
    Write-FancyLog "Warning: fancy-go.ps1 not found, skipping"
}

if (Test-Path "$ScriptDir\fancy-go.bat") {
    Copy-Item "$ScriptDir\fancy-go.bat" "$BinDir\fancy-go.bat" -Force
} else {
    Write-FancyLog "Warning: fancy-go.bat not found, skipping"
}

# Add to PATH if not already there
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$BinDir*") {
    Write-Host "📦 Adding to PATH..." -ForegroundColor Yellow
    $newPath = "$currentPath;$BinDir"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Host "✅ Added $BinDir to user PATH" -ForegroundColor Green
}

# Print installation complete message
Write-Host "`n✅ Go version installation complete!" -ForegroundColor Green
Write-Host "`n🔧 PowerShell Setup:" -ForegroundColor Yellow
Write-Host "------------------------------------------------------------" -ForegroundColor Gray
Write-Host "# Add the following to your PowerShell profile:"
Write-Host "# To edit profile: notepad `$PROFILE"
Write-Host ""
Write-Host ". `"$BinDir\fancy-go.ps1`""
Write-Host ""
Write-Host "# Then restart PowerShell or run:"
Write-Host ". `$PROFILE"
Write-Host "------------------------------------------------------------" -ForegroundColor Gray

Write-Host "`n🔧 Command Prompt Setup:" -ForegroundColor Yellow
Write-Host "------------------------------------------------------------" -ForegroundColor Gray
Write-Host "# Add the following to a batch file in your PATH:"
Write-Host "`"$BinDir\fancy-go.bat`" %*"
Write-Host "------------------------------------------------------------" -ForegroundColor Gray

Write-Host "`n🚀 Usage:" -ForegroundColor Green
Write-Host "   PowerShell: fancy-go or fancy"
Write-Host "   Command Prompt: fancy-go.bat"
Write-Host "   Direct: fancy-login-go.exe --help"
Write-Host "`n   Test with: fancy-login-go.exe --help" -ForegroundColor Cyan