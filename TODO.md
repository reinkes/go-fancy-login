# Fancy Login Go - Optimization TODO

## 🚀 High Impact, Low Effort (Immediate Actions)

### Workflow Cleanup
- [ ] **Remove GitLab CI** - Delete `.gitlab-ci.yml` if not actively using GitLab
- [ ] **Merge GitHub workflows** - Consolidate `ci.yml`, `security.yml`, `homebrew.yml` into single pipeline
- [ ] **Fix gosec installation** - ✅ Already fixed with `GOPROXY=direct`
- [ ] **Add ARM64 to CI** - Include `linux/arm64` and `darwin/arm64` in test matrix
- [ ] **Add Go module caching** - Reduce CI build times with proper Go cache configuration
- [ ] **Parallel job optimization** - Run lint, test, security scans in parallel instead of sequential

### Installation Script Cleanup
- [ ] **Remove legacy scripts** - Delete outdated files in `/scripts/`:
  - [ ] `build-all.sh`
  - [ ] `build.sh`
  - [ ] `fancy-go.bat`
  - [ ] `fancy-go.ps1`
  - [ ] `release.sh`
- [ ] **Simplify install scripts** - Keep only `install-fancy-go.sh` and `install-fancy-go.ps1`
- [ ] **Remove config file copying** - ✅ Already removed from install scripts
- [ ] **Use dynamic paths** - Replace hardcoded paths with `go env GOPATH` detection

## 🎯 Medium Impact, Medium Effort

### Cross-Platform Improvements
- [ ] **Windows PowerShell Core support** - Update scripts for cross-platform PowerShell
- [ ] **Container support** - Add lightweight Docker image with multi-stage build
- [ ] **Package manager integration**:
  - [ ] Complete Homebrew formula (already started)
  - [ ] Add Chocolatey package for Windows
  - [ ] Create AUR package for Arch Linux
  - [ ] Add snap package for Ubuntu

### Code Simplification
- [ ] **CLI framework migration** - Replace `flag` package with `cobra` or `cli` for better UX
- [ ] **Config management refactor**:
  - [ ] Merge similar config structs
  - [ ] Reduce parser complexity
  - [ ] Standardize YAML handling
- [ ] **Error handling standardization** - Create consistent error patterns across codebase
- [ ] **Dependency audit** - Remove unused imports and minimize external dependencies

### Installation Method Modernization
- [ ] **Go install support** - Enable `go install github.com/user/repo/cmd@latest` installation
- [ ] **Auto-detection improvements** - Better binary location and PATH management
- [ ] **Version management** - Support multiple versions and easy switching

## 🔄 Long-term Refactoring

### Architecture Improvements
- [ ] **Plugin system** - Modular architecture for different cloud providers
- [ ] **Configuration validation** - Add schema validation for YAML configs
- [ ] **Logging standardization** - Implement structured logging with levels
- [ ] **Testing improvements**:
  - [ ] Add integration tests
  - [ ] Increase unit test coverage
  - [ ] Add benchmark tests

### Performance Optimizations
- [ ] **Binary size reduction** - Optimize build flags and remove unused code
- [ ] **Startup time optimization** - Lazy loading of heavy dependencies
- [ ] **Memory usage optimization** - Profile and optimize memory allocations

### User Experience Enhancements
- [ ] **Interactive configuration wizard** - Guided setup for first-time users
- [ ] **Shell completion** - Add bash/zsh/fish completion scripts
- [ ] **Better error messages** - More descriptive and actionable error output
- [ ] **Progress indicators** - Add progress bars for long-running operations

## 📋 Workflow Consolidation Plan

### Proposed Single Pipeline Structure
```yaml
name: CI/CD Pipeline
on: [push, pull_request, release]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.21, 1.22, 1.23]
    # Run tests in parallel across platforms

  security:
    needs: test
    # Run security scans (gosec, trivy, codeql)

  build:
    needs: [test, security]
    strategy:
      matrix:
        include:
          - os: linux, arch: amd64
          - os: linux, arch: arm64
          - os: darwin, arch: amd64
          - os: darwin, arch: arm64
          - os: windows, arch: amd64
    # Build all platform binaries

  release:
    needs: build
    if: startsWith(github.ref, 'refs/tags/v')
    # Create GitHub release with assets

  homebrew:
    needs: release
    if: startsWith(github.ref, 'refs/tags/v')
    # Update Homebrew formula
```

## 🗂️ File Structure Cleanup

### Files to Remove
- [ ] `.gitlab-ci.yml` (if not using GitLab)
- [ ] `scripts/build-all.sh`
- [ ] `scripts/build.sh`
- [ ] `scripts/fancy-go.bat`
- [ ] `scripts/fancy-go.ps1`
- [ ] `scripts/release.sh`

### Files to Consolidate
- [ ] Merge `README.md` and `README-Windows.md` into single cross-platform README
- [ ] Consolidate example configs into `examples/` directory with better documentation

## 🎯 Priority Order

1. **Week 1**: High impact, low effort items (workflow cleanup, script removal)
2. **Week 2**: Installation modernization and cross-platform improvements
3. **Week 3**: Code simplification and CLI framework migration
4. **Week 4**: Long-term architecture improvements

## 📊 Success Metrics

- [ ] **Build time reduction**: Target 50% faster CI builds
- [ ] **Installation simplicity**: Single command installation on all platforms
- [ ] **Binary size**: Reduce by 20% through optimization
- [ ] **Code coverage**: Achieve 80%+ test coverage
- [ ] **User experience**: Zero-config setup for common use cases

## 🔗 Dependencies

Some tasks have dependencies on others:
- CLI framework migration should happen before UX enhancements
- Package manager integration requires stable binary builds
- Plugin system requires architecture refactoring first

---

*This TODO represents a comprehensive optimization roadmap. Items can be tackled individually or in groups based on priority and available time.*