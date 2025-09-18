# GitHub Migration Guide

## Quick Start for GitHub

### 🚀 Releases and Downloads

**Latest Release:** Check the [Releases page](../../releases) for the latest version.

**Platform Downloads:**
- **Linux AMD64**: `fancy-login-go-v*.*.***-linux-amd64.tar.gz`
- **Linux ARM64**: `fancy-login-go-v*.*.***-linux-arm64.tar.gz`
- **macOS Intel**: `fancy-login-go-v*.*.***-darwin-amd64.tar.gz`
- **macOS Apple Silicon**: `fancy-login-go-v*.*.***-darwin-arm64.tar.gz`
- **Windows**: `fancy-login-go-v*.*.***-windows-amd64.exe.zip`

### 📦 Installation

1. **Download** the appropriate archive for your platform
2. **Verify** the checksum using the provided `.sha256` file
3. **Extract** the archive and move the binary to your PATH
4. **Configure** using `fancy-login-go --config`

### 📦 Package Managers

**Homebrew (Recommended):**
```bash
# Install via Homebrew (automatically updated)
brew install fancy-login-go

# Run the tool
fancy-login-go --config
```

**Manual Installation:**
```bash
# Download and install manually
curl -L https://github.com/[username]/go-fancy-login/releases/latest/download/fancy-login-go-[version]-[platform].tar.gz
tar -xzf fancy-login-go-[version]-[platform].tar.gz
sudo mv fancy-login-go /usr/local/bin/
```

### 🔄 CI/CD Pipeline

**Automated Workflows:**
- **Lint & Test**: On every push and PR
- **Build**: Multi-platform binaries for all commits
- **Security**: Daily vulnerability scans
- **Release**: Automatic on version tags

**Supported Platforms:**
- Linux (AMD64, ARM64)
- macOS (Intel, Apple Silicon)
- Windows (AMD64)
- Package Managers (Homebrew)

### 📋 Migration from GitLab

**Key Changes:**
- GitHub Actions instead of GitLab CI
- GitHub Releases for binary distribution
- Homebrew integration for easy installation
- Enhanced security scanning with CodeQL

**Pipeline Features:**
- ✅ Automated testing and linting
- ✅ Multi-platform builds
- ✅ Security scanning (Trivy, Gosec, CodeQL)
- ✅ Dependency review
- ✅ Automated releases
- ✅ Package manager integration
- ✅ Artifact management

### 🔒 Security

**Integrated Security:**
- CodeQL static analysis
- Trivy vulnerability scanning
- Gosec security scanning
- Dependency review
- SARIF integration with GitHub Security

**Scheduled Scans:**
- Daily security checks
- Automatic vulnerability alerts
- Dependency update notifications

### 📖 Documentation

For complete documentation, see [CONTEXT.md](../CONTEXT.md).

### 🏷️ Creating Releases

To create a new release:

1. **Tag** your commit: `git tag v1.2.3`
2. **Push** the tag: `git push origin v1.2.3`
3. **GitHub Actions** automatically:
   - Builds all platform binaries
   - Creates release with changelog
   - Uploads downloadable assets
   - Calculates checksums
   - Publishes Docker images

### 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Submit a pull request

**Automated Checks:**
- Linting with golangci-lint
- Security scanning
- Test coverage
- Dependency review