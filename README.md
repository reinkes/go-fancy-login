# Fancy Login - Go Version

This is a Go port of the original shell-based fancy-login utility. It provides the same functionality with improved performance and easier maintenance.

## Features

- Interactive AWS SSO login and profile selection using fzf
- Automatic Kubernetes context switching based on AWS profiles
- ECR login for development profiles
- k9s integration with namespace-aware launches
- iTerm2 integration (tab titles and badges)
- Colorized, minimal output with verbose mode option
- Configuration-driven context and namespace mapping

## Requirements

- Go 1.19+ (for building)
- aws CLI
- kubectl
- fzf
- k9s
- docker (for ECR login)

## Installation

### Linux/macOS

1. **Build and install:**
   ```bash
   cd go
   ./scripts/install-fancy-go.sh
   ```

2. **Add to your shell configuration (~/.zshrc):**
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   
   # Fancy login function (Go version)
   fancy-go() {
     if fancy-login-go "$@"; then
       [[ -f /tmp/aws_profile.sh ]] && source /tmp/aws_profile.sh
     fi
   }
   
   # Or create an alias for convenience
   alias fancy='fancy-go'
   ```

3. **Reload your shell:**
   ```bash
   source ~/.zshrc
   ```

### Windows

1. **Build and install (PowerShell):**
   ```powershell
   cd go
   .\scripts\install-fancy-go.ps1
   ```

2. **Add to PowerShell profile ($PROFILE):**
   ```powershell
   # Edit your profile
   notepad $PROFILE
   
   # Add this line:
   . "$env:USERPROFILE\AppData\Local\fancy-login\fancy-go.ps1"
   ```

3. **Reload PowerShell:**
   ```powershell
   . $PROFILE
   ```

**Alternative for Command Prompt:**
- Use `fancy-go.bat` directly or add to PATH
- Binary installs to: `%USERPROFILE%\AppData\Local\fancy-login\`

## Usage

### Linux/macOS
```bash
# Basic usage
fancy-go  # or fancy with alias

# With verbose output  
fancy-go -v

# Auto-launch k9s for DEVENG profiles
fancy-go -k

# Direct binary usage
fancy-login-go --help
```

### Windows
```powershell
# PowerShell (after setup)
fancy-go
fancy-go -v
fancy-go -k

# Command Prompt
fancy-go.bat
fancy-go.bat -v

# Direct binary usage
fancy-login-go.exe --help
```

## Configuration Files

### `.fancy-contexts.conf`
Maps AWS profiles to Kubernetes contexts using shell-style wildcards:
```
*_PROD_* = prod-cluster
*_TEST_* = test-cluster
```

### `.fancy-namespaces.conf`
Maps project codes to namespace prefixes:
```
IMP=mykn-track-importer
DET=mykn-track-details
MD=mykn-masterdata
OV=mykn-track-overviews
```

## Environment Variables

- `FANCY_VERBOSE`: Enable verbose output (0/1)
- `FANCY_DEBUG`: Enable debug mode (0/1)
- `FANCY_NAMESPACE_CONFIG`: Path to namespace config file
- `FANCY_PROFILE_TEMP`: Path to AWS profile temp file
- `FANCY_DEFAULT_REGION`: Default AWS region
- `FANCY_BIN_DIR`: Installation directory
- `FANCY_AWS_DIR`: AWS config directory
- `FANCY_KUBE_DIR`: Kubernetes config directory

## Project Structure

```
go/
├── cmd/
│   └── main.go              # Main application entry point
├── internal/
│   ├── aws/
│   │   └── aws.go           # AWS operations (SSO, ECR, profiles)
│   ├── config/
│   │   └── config.go        # Configuration handling
│   ├── k8s/
│   │   └── k8s.go           # Kubernetes operations
│   └── utils/
│       └── logger.go        # Logging and utilities
├── scripts/
│   └── install-fancy-go.sh  # Installation script
├── .fancy-contexts.conf     # Context mappings
├── .fancy-namespaces.conf   # Namespace mappings
└── README.md                # This file
```

## Differences from Shell Version

- **Performance**: Faster startup and execution
- **Error Handling**: More robust error handling and reporting
- **Maintainability**: Cleaner code structure with proper separation of concerns
- **Cross-platform**: Can be compiled for different operating systems
- **Testing**: Easier to unit test individual components

## Building from Source

### Quick Build
```bash
cd go
go mod tidy
go build -o fancy-login-go ./cmd
```

### Using Make
```bash
cd go
make help              # Show all available targets
make build             # Build for current platform
make build-all         # Build for all platforms
make release           # Create release archives
make version           # Show version info
```

### Using Build Scripts
```bash
cd go
./scripts/build.sh                    # Simple build
./scripts/build-all.sh v1.0.0         # All platforms
./scripts/release.sh -v v1.0.0 -t     # Full release
```

### Cross-Platform Builds
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o fancy-login-go-linux-amd64 ./cmd

# Windows  
GOOS=windows GOARCH=amd64 go build -o fancy-login-go-windows-amd64.exe ./cmd

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o fancy-login-go-darwin-amd64 ./cmd

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o fancy-login-go-darwin-arm64 ./cmd
```

See [BUILD.md](BUILD.md) for detailed build instructions and CI/CD setup.

## Development

To modify the Go version:

1. Make changes to the source code
2. Test locally: `go run ./cmd [flags]`
3. Build: `go build -o fancy-login-go ./cmd`
4. Install: `cp fancy-login-go $HOME/.local/bin/`

## Troubleshooting

- **Module issues**: Run `go mod tidy` to fix dependency issues
- **Binary not found**: Ensure `$HOME/.local/bin` is in your PATH
- **Config not found**: Check that config files are in the correct location
- **Verbose mode**: Use `-v` flag to see detailed logging

## Migration from Shell Version

The Go version maintains full compatibility with the shell version's configuration files and behavior. You can run both versions side by side or replace the shell version entirely.