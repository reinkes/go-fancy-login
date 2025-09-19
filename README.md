# Fancy Login Go

A streamlined AWS SSO login and Kubernetes context selection utility that provides a minimal, colorful, and interactive CLI experience for switching between cloud environments.

[![CI/CD Pipeline](https://github.com/reinkes/go-fancy-login/actions/workflows/ci.yml/badge.svg)](https://github.com/reinkes/go-fancy-login/actions/workflows/ci.yml)
[![Security Scan](https://github.com/reinkes/go-fancy-login/actions/workflows/security.yml/badge.svg)](https://github.com/reinkes/go-fancy-login/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/reinkes/go-fancy-login)](https://goreportcard.com/report/github.com/reinkes/go-fancy-login)

## ✨ Features

- **🔐 Interactive AWS SSO Login**: Profile selection with fzf, automatic session validation
- **⎈ Smart Kubernetes Context Switching**: Profile-based context mapping with interactive fallback
- **🐳 Configurable ECR Authentication**: Per-profile Docker login with regional support
- **🦄 k9s Integration**: Namespace-aware launches with auto-launch mode
- **🖥️ iTerm2 Visual Integration**: Tab titles and badges showing current namespace
- **⚡ Profile-Based Configuration**: Direct profile configuration via interactive wizard
- **🎨 Colorized Output**: Minimal, consistent UI with emoji icons and progress indicators
- **🔧 Cross-Platform Support**: Linux, macOS, and Windows binaries

## 🚀 Quick Start

### Installation

**Download from GitHub Releases (Recommended):**

1. Go to [Releases](https://github.com/reinkes/go-fancy-login/releases) and download the appropriate binary for your platform:
   - **Linux AMD64**: `fancy-login-go-v*.*.*-linux-amd64.tar.gz`
   - **Linux ARM64**: `fancy-login-go-v*.*.*-linux-arm64.tar.gz`
   - **macOS Intel**: `fancy-login-go-v*.*.*-darwin-amd64.tar.gz`
   - **macOS Apple Silicon**: `fancy-login-go-v*.*.*-darwin-arm64.tar.gz`
   - **Windows**: `fancy-login-go-v*.*.*-windows-amd64.exe.zip`

2. Extract and install:
   ```bash
   # Linux/macOS
   tar -xzf fancy-login-go-v*.*.*-[platform].tar.gz
   sudo mv fancy-login-go /usr/local/bin/

   # Or to user directory
   mkdir -p ~/.local/bin
   mv fancy-login-go ~/.local/bin/
   export PATH="$HOME/.local/bin:$PATH"
   ```

**Go Install (Cross-platform):**
```bash
# Install latest version
go install github.com/reinkes/go-fancy-login/cmd@latest

# Ensure $GOPATH/bin is in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

**Homebrew (macOS/Linux):**
```bash
# Add the custom tap and install
brew install reinkes/tap/fancy-login-go

# Or add tap first, then install
brew tap reinkes/tap
brew install fancy-login-go
```

**Build from Source:**
```bash
git clone https://github.com/reinkes/go-fancy-login.git
cd go-fancy-login
make build
sudo cp fancy-login-go /usr/local/bin/
```

### First Run - Configuration Wizard

```bash
# Run the interactive configuration wizard
fancy-login-go --config

# Or start using immediately - wizard runs automatically
fancy-login-go
```

The wizard will:
- Discover your AWS profiles automatically
- Configure ECR login, Kubernetes contexts, and k9s settings per profile
- Create your personalized configuration file

## 📖 Usage

### Basic Commands

```bash
# Interactive login with profile and context selection
fancy-login-go

# Verbose output showing all operations
fancy-login-go -v

# Auto-launch k9s for profiles configured with auto-launch
fancy-login-go -k

# Force AWS re-authentication
fancy-login-go -f

# Show version information
fancy-login-go --version

# Run configuration wizard
fancy-login-go --config
```

### Shell Integration

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
# Fancy login function
fancy() {
    if fancy-login-go "$@"; then
        # Source any environment changes
        [[ -f /tmp/aws_profile.sh ]] && source /tmp/aws_profile.sh
    fi
}
```

### Windows PowerShell

Add to your PowerShell profile (`$PROFILE`):

```powershell
function fancy {
    fancy-login-go.exe $args
    if ($LASTEXITCODE -eq 0) {
        if (Test-Path "$env:TEMP\aws_profile.ps1") {
            . "$env:TEMP\aws_profile.ps1"
        }
    }
}
```

## ⚙️ Configuration

### Profile-Based Configuration

Fancy Login uses a profile-based configuration system. Each AWS profile can be configured individually with:

- **ECR Login**: Whether to perform Docker login for this profile
- **ECR Region**: Which region to authenticate with for ECR
- **Kubernetes Context**: Which k8s context to switch to
- **K9s Auto-launch**: Whether to automatically launch k9s
- **Namespace Prefix**: For deriving namespaces from profile names

Configuration is stored in `~/.fancy-config.yaml`:

```yaml
settings:
  default_region: us-east-1
  config_wizard_run: true

profile_configs:
  company_DEV_developer:
    name: company_DEV_developer
    account_id: "123456789012"
    ecr_login: true
    ecr_region: us-east-1
    k8s_context: dev-cluster
    k9s_auto_launch: true
    namespace_prefix: dev

  company_PROD_admin:
    name: company_PROD_admin
    account_id: "987654321098"
    ecr_login: false
    ecr_region: us-east-1
    k8s_context: prod-cluster
    k9s_auto_launch: false
```

### Environment Variables

```bash
# Enable verbose output
export FANCY_VERBOSE=1

# Enable debug mode
export FANCY_DEBUG=1

# Override default AWS region
export FANCY_DEFAULT_REGION=eu-central-1

# Custom configuration paths
export FANCY_CONFIG_PATH="$HOME/.config/fancy-login.yaml"
```

## 🔧 Requirements

- **AWS CLI**: For SSO authentication and profile management
- **kubectl**: For Kubernetes context switching
- **fzf**: For interactive selection menus
- **docker**: For ECR authentication (optional)
- **k9s**: For Kubernetes cluster management (optional)

## 🏗️ Development & Contributing

### Building from Source

```bash
# Quick build
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint

# Create release archives
make release

# Show all available targets
make help
```

### Project Structure

```
.
├── cmd/                    # Main application entry point
├── internal/
│   ├── aws/               # AWS operations (SSO, ECR, profiles)
│   ├── config/            # Configuration handling and wizard
│   ├── k8s/               # Kubernetes operations
│   └── utils/             # Logging and utilities
├── tools/                 # Development tools and test utilities
├── examples/              # Configuration templates
└── .github/
    └── workflows/         # CI/CD pipeline definitions
```

### Running Tests

```bash
# Run all tests with coverage
make test

# Run specific test packages
go test ./internal/config/
go test ./internal/aws/

# Run with verbose output
go test -v ./...

# Generate coverage report
make test
open dist/test-results/coverage.html
```

## 🔄 CI/CD Pipeline

This project uses GitHub Actions for automated building, testing, and releasing:

### Automated Workflows

**On every push and PR:**
- ✅ **Lint & Format**: Code quality checks with golangci-lint
- ✅ **Test Suite**: Unit tests with coverage reporting
- ✅ **Multi-platform Build**: Linux, macOS, Windows (AMD64/ARM64)
- ✅ **Security Scanning**: Trivy, Gosec, and CodeQL analysis

**Daily security scans:**
- ✅ **Vulnerability Detection**: Automated dependency scanning
- ✅ **Code Analysis**: Static security analysis
- ✅ **Dependency Review**: Automated dependency updates

**On version tags:**
- ✅ **Automated Releases**: GitHub releases with changelog
- ✅ **Binary Distribution**: Multi-platform binaries with checksums
- ✅ **Package Manager Updates**: Homebrew formula updates

### Creating a Release

To create a new release:

```bash
# Tag your commit
git tag v1.2.3

# Push the tag
git push origin v1.2.3
```

GitHub Actions will automatically:
- Build binaries for all platforms
- Create a GitHub release with changelog
- Upload downloadable assets with SHA256 checksums
- Update package manager formulas

### Security Features

- **CodeQL Analysis**: Static code analysis for security vulnerabilities
- **Trivy Scanning**: Container and dependency vulnerability scanning
- **Gosec**: Go-specific security analysis
- **Dependency Review**: Automated security review for dependencies
- **SARIF Integration**: Security findings integrated with GitHub Security tab

## 🎯 Migration from Shell Version

The Go version maintains full compatibility with existing workflows while providing:

- **🚀 Performance**: 10x faster startup and execution
- **🛡️ Error Handling**: Robust error handling and recovery
- **🧪 Testing**: Comprehensive unit test coverage
- **🔧 Maintainability**: Clean, modular architecture
- **📦 Distribution**: Easy installation via package managers
- **🖥️ Cross-Platform**: Native Windows support

## 🐛 Troubleshooting

### Common Issues

**Binary not found:**
```bash
# Ensure binary is in PATH
echo $PATH
which fancy-login-go

# Add to PATH if needed
export PATH="$HOME/.local/bin:$PATH"
```

**Configuration issues:**
```bash
# Run configuration wizard
fancy-login-go --config

# Check configuration file
cat ~/.fancy-config.yaml

# Enable verbose mode for debugging
fancy-login-go -v
```

**AWS SSO issues:**
```bash
# Check AWS CLI configuration
aws configure list
aws sso login --profile YOUR_PROFILE

# Verify profile configuration
aws sts get-caller-identity --profile YOUR_PROFILE
```

**Kubernetes context issues:**
```bash
# Check available contexts
kubectl config get-contexts

# Verify current context
kubectl config current-context

# Test context switching manually
kubectl config use-context YOUR_CONTEXT
```

### Debug Mode

Enable debug mode for detailed troubleshooting:

```bash
export FANCY_DEBUG=1
fancy-login-go -v
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Run the test suite (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Contribution Guidelines

- Ensure all tests pass
- Add tests for new features
- Follow Go coding conventions
- Update documentation as needed
- Security: All contributions are automatically scanned for vulnerabilities

---

**Made with ❤️ for seamless cloud environment switching**