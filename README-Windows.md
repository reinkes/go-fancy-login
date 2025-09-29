# Fancy Login - Windows Setup Guide

Complete setup guide for running fancy-login Go version on Windows.

## Prerequisites

Fancy Login requires several command-line tools to function properly. Choose your preferred package manager:

### Required Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| **AWS CLI** | AWS authentication | `winget install Amazon.AWSCLI` |
| **kubectl** | Kubernetes cluster management | `winget install Kubernetes.kubectl` |
| **fzf** | Interactive fuzzy finder | `winget install junegunn.fzf` |

### Optional Tools (Recommended)

| Tool | Purpose | Installation |
|------|---------|--------------|
| **k9s** | Kubernetes cluster visualization | `winget install Kubernetes.k9s` |
| **Docker Desktop** | Container runtime for ECR | `winget install Docker.DockerDesktop` |

### Package Manager Options

- **Winget**: Built into Windows 10/11 (recommended)
- **Scoop**: Lightweight command-line installer
- **Chocolatey**: Full-featured package manager
- **Direct Downloads**: From official websites

## Installation Steps

### 1. Install Dependencies

**Using Winget (Recommended):**
```powershell
# Required tools
winget install Amazon.AWSCLI
winget install Kubernetes.kubectl
winget install junegunn.fzf

# Optional tools
winget install Kubernetes.k9s
winget install Docker.DockerDesktop
```

**Using Scoop:**
```powershell
# Install Scoop first if not already installed
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# Required tools
scoop install aws kubectl fzf

# Optional tools
scoop install k9s docker
```

**Using Chocolatey:**
```powershell
# Required tools
choco install awscli kubernetes-cli fzf

# Optional tools
choco install k9s docker-desktop
```

**Verify Installation:**
```powershell
# Check required tools
aws --version
kubectl version --client
fzf --version

# Check optional tools
k9s version      # (optional)
docker --version # (optional)
```

### 2. Install fancy-login-go

**Option A: Scoop (Recommended)**
```powershell
# Install directly from GitHub
scoop install https://raw.githubusercontent.com/reinkes/homebrew-tap/main/bucket/fancy-login-go.json

# Scoop will suggest any missing dependencies
```

**Option B: Download from Releases**
```powershell
# Download from GitHub Releases and extract to PATH
# See main README.md for download links
```

**Option C: Build from Source**
```powershell
# Clone and build
git clone https://github.com/reinkes/go-fancy-login.git
cd go-fancy-login
.\scripts\install-fancy-go.ps1

# Or with verbose output
.\scripts\install-fancy-go.ps1 -Verbose
```

### 3. PowerShell Setup

```powershell
# Check if profile exists
Test-Path $PROFILE

# Create profile if it doesn't exist  
if (!(Test-Path $PROFILE)) { New-Item -Type File -Path $PROFILE -Force }

# Edit your profile
notepad $PROFILE

# Add this line to your profile:
. "$env:USERPROFILE\AppData\Local\fancy-login\fancy-go.ps1"

# Reload profile
. $PROFILE
```

### 4. Command Prompt Setup (Alternative)

```batch
# Create a batch file in your PATH, e.g., C:\Windows\System32\fancy.bat
@echo off
"%USERPROFILE%\AppData\Local\fancy-login\fancy-go.bat" %*
```

## Usage

### PowerShell (Recommended)
```powershell
# Basic usage
fancy-go

# With options
fancy-go -v          # Verbose output
fancy-go -k          # Auto-launch k9s
fancy-go --help      # Show help

# Using alias (if configured)
fancy
```

### Command Prompt
```batch
# Direct batch usage
fancy-go.bat
fancy-go.bat -v
fancy-go.bat -k

# If you created the system batch file
fancy
fancy -v
```

### Direct Binary Usage
```powershell
# Run the Go binary directly
fancy-login-go.exe --help
fancy-login-go.exe -v
```

## File Locations

### Installation Directory
- **Binary**: `%USERPROFILE%\AppData\Local\fancy-login\fancy-login-go.exe`
- **Config Files**: `%USERPROFILE%\AppData\Local\fancy-login\`
- **Shell Scripts**: `%USERPROFILE%\AppData\Local\fancy-login\fancy-go.ps1`

### Configuration Files
- **AWS Config**: `%USERPROFILE%\.aws\config`  
- **Kubernetes Config**: `%USERPROFILE%\.kube\config`
- **Context Mapping**: `%USERPROFILE%\AppData\Local\fancy-login\.fancy-contexts.conf`
- **Namespace Mapping**: `%USERPROFILE%\AppData\Local\fancy-login\.fancy-namespaces.conf`

### Temporary Files
- **PowerShell**: `%TEMP%\aws_profile.ps1`
- **Batch**: `%TEMP%\aws_profile.bat`

## Terminal Integration

### Windows Terminal
- Supports tab title updates
- Shows current namespace in tab title (`ns:namespace-name`)
- Automatically detected via `$env:WT_SESSION`

### PowerShell ISE
- Basic functionality supported
- No tab title integration

### Command Prompt
- Basic functionality supported  
- No tab title integration

## Troubleshooting

### Common Issues

**1. "fancy-go not recognized"**
- Ensure PowerShell profile is loaded: `. $PROFILE`
- Check if installation completed successfully
- Restart PowerShell

**2. "fancy-login-go.exe not found"**  
- Check PATH includes `%USERPROFILE%\AppData\Local\fancy-login`
- Verify binary was built successfully
- Try absolute path to binary

**3. "AWS credentials not found"**
- Ensure AWS CLI is configured: `aws configure`
- Check AWS config file exists: `%USERPROFILE%\.aws\config`
- Verify SSO configuration

**4. "kubectl not found"**
- Install kubectl via winget/chocolatey
- Add kubectl to PATH manually if needed
- Test: `kubectl version --client`

**5. Environment variable not persisting**
- Check PowerShell profile sources `fancy-go.ps1`
- Verify temp files are created: `%TEMP%\aws_profile.ps1`
- Try restarting PowerShell

### Debugging

```powershell
# Enable verbose output
fancy-go -v

# Check installation
Test-Path "$env:USERPROFILE\AppData\Local\fancy-login\fancy-login-go.exe"

# Test direct binary
& "$env:USERPROFILE\AppData\Local\fancy-login\fancy-login-go.exe" --help

# Check profile function
Get-Command fancy-go
```

## Performance Notes

- Go binary starts faster than PowerShell scripts
- fzf selection is very responsive
- AWS CLI calls are the slowest part (network-dependent)
- k9s launch time depends on cluster size

## Differences from Unix Version

1. **Paths**: Uses `AppData\Local` instead of `.local/bin`
2. **Temp Files**: Creates both `.ps1` and `.bat` profile scripts  
3. **Terminal**: Windows Terminal integration instead of iTerm2
4. **Shell**: PowerShell function instead of bash function
5. **Permissions**: No executable bit needed on Windows

The Go version provides the same functionality across all platforms with these platform-specific adaptations.