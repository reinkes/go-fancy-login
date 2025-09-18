# Go Migration Context - Fancy Login

## Project Overview
Successfully migrated the shell-based fancy-login utility to Go, maintaining full compatibility with existing configuration files and workflows while improving performance and maintainability.

**Fancy Login** is a streamlined AWS SSO login and Kubernetes context selection utility that provides a minimal, colorful, and interactive CLI experience for switching between cloud environments.

## Core Features & Use Cases

### Primary Use Cases
1. **Quick AWS SSO Authentication**
   - Interactive AWS profile selection using fzf
   - Automatic SSO login detection and handling
   - Force login option for expired sessions
   - Session validation and error handling

2. **Kubernetes Context Management**  
   - Automatic context switching based on AWS profile patterns
   - Interactive context selection when no mapping exists
   - Context mapping via `.fancy-contexts.conf` with wildcard support
   - Silent context switching with summary display

3. **ECR Authentication**
   - Configurable ECR login per profile via wizard configuration
   - Docker credential helper integration
   - Regional ECR endpoint support
   - Account ID detection and validation

4. **Namespace-Aware k9s Integration**
   - Automatic namespace derivation for `*_DEVENG` profiles  
   - Interactive k9s launch with namespace pre-selection
   - Auto-launch mode (`-k` flag) for scripted workflows
   - Project code to namespace mapping via `.fancy-namespaces.conf`

5. **iTerm2 Visual Integration**
   - Tab title updates showing current namespace (`ns:namespace-name`)
   - Badge display with namespace information (`🟢 ns:namespace-name`)
   - Instant visual identification of active environment
   - Base64-encoded badge format for compatibility

### Advanced Features

**Intelligent Profile Handling:**
- Direct profile-based configuration via interactive wizard
- Per-profile ECR login, Kubernetes context, and K9s settings
- Profile validation and session checking
- Smart configuration updates for new profiles

**User Experience Enhancements:**
- Colorized output with consistent emoji icons (☁️ AWS, ⎈ k8s, 🐳 ECR, 🦄 Summary)
- Minimal output by default, detailed logs in verbose mode (`-v`)
- Progress spinners for slow operations (AWS SSO, ECR login)
- Labeled summary output before k9s prompts
- Context switching output suppression except summary line

**Configuration-Driven Behavior:**
- Environment variable customization for all paths and settings
- Configurable default AWS region
- Custom config file locations
- Installation path flexibility

### Workflow Examples

**Standard Daily Workflow:**
1. Run `fancy-go` in terminal
2. Select AWS profile from fzf list
3. Automatic SSO authentication (if needed)  
4. Auto-select k8s context based on profile mapping
5. ECR login (if configured for profile)
6. See colorized summary with namespace
7. Optional k9s launch in correct namespace (if configured)

**DEVENG Profile Workflow:**
```bash
fancy-go -k  # Auto-launch k9s
# → Selects OV_TEST_DEVENG
# → Maps to test-cluster context  
# → Derives test-myapp-overviews namespace
# → Launches k9s directly in that namespace
```

**Production Profile Workflow:**
```bash
fancy-go
# → Selects OV_PROD_MONITORING  
# → Maps to prod-cluster context
# → No ECR login (not _DEV_)
# → Shows summary, prompts for k9s (not DEVENG)
```

### Feature Details & Behavior

**AWS Profile Selection:**
- Lists all profiles from `~/.aws/config` using regex parsing
- Uses fzf for interactive selection with fuzzy matching
- Handles both `[profile name]` and `[default]` formats
- Validates profile existence and SSO configuration
- Exports profile to `/tmp/aws_profile.sh` for shell integration

**SSO Authentication Logic:**
- Checks existing session validity with `aws sts get-caller-identity`
- Detects SSO profiles by scanning config for `sso_*` parameters
- Shows progress spinner during authentication unless verbose mode
- Handles authentication failures gracefully with user prompts
- Force login option bypasses session validity checks

**Kubernetes Context Mapping:**
- Processes `.fancy-contexts.conf` line by line
- Supports shell-style wildcards (`*`, `?`) converted to regex
- Maps AWS profiles to k8s contexts automatically
- Falls back to fzf selection if no mapping found
- Switches contexts silently unless verbose mode enabled

**Namespace Derivation Rules:**
- Pattern: `{PROJECT}_{ENV}_DEVENG` → `{env}-{project-name}`
- Example: `OV_TEST_DEVENG` → `test-myapp-overviews`
- Uses `.fancy-namespaces.conf` for project code → name mapping
- Case conversion: ENV (uppercase) → env (lowercase)
- Only applies to profiles ending in `DEVENG`

**ECR Login Behavior:**
- Configurable per profile via configuration wizard
- Gets AWS account ID from `aws sts get-caller-identity`
- Uses configured region for the profile
- Authenticates with ECR using `aws ecr get-login-password`
- Pipes credentials to `docker login` command
- Shows success/failure in summary output

**k9s Integration Modes:**
1. **Interactive Mode** (default): Prompts user before launching
2. **Auto Mode** (`-k` flag): Launches immediately for DEVENG profiles
3. **Namespace Mode**: Always launches with derived namespace
4. **Environment Inheritance**: Passes AWS_PROFILE to k9s process

**Output Modes & Verbosity:**
- **Normal Mode**: Shows only summary block and prompts
- **Verbose Mode** (`-v`): Shows all intermediate steps and commands
- **Color Coding**: Consistent emoji and color scheme throughout
- **Progress Indicators**: Spinners for slow AWS/Docker operations
- **Error Handling**: Clear error messages with suggested actions

### Edge Cases Handled

**Missing Dependencies:**
- Graceful failure if fzf, k9s, kubectl, aws, docker not found
- Clear error messages indicating which tool is missing
- Installation script validates all requirements upfront

**Configuration Issues:**
- Missing or malformed config files handled gracefully  
- Default fallbacks for missing environment variables
- Validation of AWS config file format and profile structure

**Authentication Edge Cases:**
- Expired SSO sessions detected and handled automatically
- Non-SSO profiles supported with user confirmation prompts
- Network failures during authentication show appropriate errors
- Invalid profiles or credentials handled with clear messaging

**Context & Namespace Issues:**
- Missing k8s contexts handled with fallback to current context
- Invalid namespace patterns logged but don't break execution
- k9s launch failures don't affect overall workflow success

## Configuration Management System

### Profile-Based Configuration
The application now uses a modern profile-based configuration system that replaces pattern matching with direct profile configuration:

**Configuration File Location:**
- Primary: `~/.fancy-config.yaml` (user's home directory)
- Development: `./.fancy-config.yaml` (local development override)

**Profile Configuration Structure:**
```yaml
profile_configs:
  MD_DEV_ADMIN:
    name: MD_DEV_ADMIN
    account_id: "774305606488"
    ecr_login: true
    ecr_region: eu-central-1
    k8s_context: ""
    k9s_auto_launch: false
  MD_PROD_DEVENG:
    name: MD_PROD_DEVENG
    account_id: "108782075333"
    ecr_login: false
    ecr_region: ""
    k8s_context: "prod-cluster"
    k9s_auto_launch: true
settings:
  default_region: eu-central-1
  config_wizard_run: true
  prefer_local_configs: true
```

### Interactive Configuration Wizard

**Initial Setup:**
```bash
fancy-login-go --config
```

**Wizard Features:**
- **Profile Discovery**: Automatically detects AWS profiles and Kubernetes contexts
- **Smart Configuration**: Shows existing configurations and allows selective updates
- **Profile Status**: Visual indicators show which profiles are already configured
- **Context Selection**: Interactive selection of Kubernetes contexts per profile
- **ECR Configuration**: Per-profile ECR login settings with region selection
- **K9s Settings**: Auto-launch configuration for each profile

**Wizard Modes:**
1. **First Run**: Configure all discovered profiles
2. **Add New Profiles**: Only configure newly discovered profiles (default)
3. **Override All**: Reconfigure all profiles (with confirmation)

**Rerun Behavior:**
When the wizard detects existing configuration:
```
📋 Found existing configuration with 20 profiles
Configuration mode:
  1. Override all (reconfigure all profiles)
  2. Add new profiles only (keep existing, add new ones)
Choice [2]:
```

**Profile Configuration Flow:**
1. **Profile Selection**: Choose which profiles to configure
2. **ECR Login**: Enable/disable ECR authentication per profile
3. **ECR Region**: Set region for ECR login (defaults to profile region)
4. **Kubernetes Context**: Select from available contexts or "None"
5. **K9s Auto-launch**: Enable automatic K9s launch for profiles with contexts

**Configuration Benefits:**
- **Granular Control**: Each profile has individual settings
- **Easy Updates**: Add new profiles without affecting existing ones
- **Visual Feedback**: Clear status indicators for configured profiles
- **Consistent Prompts**: Standardized `[Y/n]` and `[y/N]` defaults
- **Safe Operation**: Confirmation prompts for destructive operations
- **Reliability**: No hanging on empty contexts or failed interactive selections
- **Timeout Protection**: 60-second safety timeout prevents indefinite waiting

### Migration from Pattern-Based System
- **Backward Compatibility**: Old `.fancy-contexts.conf` still supported as fallback
- **DEV Profile Logic**: Removed automatic DEV profile handling
- **Direct Mapping**: Profile names directly map to configurations
- **Simplified Logic**: No more complex pattern matching or special cases
- **Empty Context Handling**: Profiles with empty `k8s_context` skip context switching entirely
- **Timeout Protection**: All interactive selections timeout after 60 seconds to prevent hanging

### Reliability and Error Handling

**Interactive Selection Improvements:**
- **Hanging Prevention**: 60-second timeout on all `fzf` interactive selections
- **Empty Context Logic**: Profiles with no configured context skip selection entirely
- **Clear Error Messages**: Descriptive timeout and failure messages
- **Graceful Degradation**: Fallback behaviors when selections fail

**Profile Execution Flow:**
1. **Configured Context**: Direct context switch without user interaction
2. **Empty Context**: Skip context switching, show "(not configured for this profile)"
3. **Legacy Fallback**: Interactive selection with 60-second timeout protection
4. **Timeout Handling**: Clear error message and graceful exit after timeout

**Common Scenarios:**
- **OV_DEV_ADMIN** (empty `k8s_context`): Shows "not configured" and continues
- **MD_PROD_DEVENG** (configured context): Automatically switches to `prod-cluster`
- **Unconfigured Profiles**: Interactive selection with timeout protection
- **Terminal Issues**: Timeout prevents indefinite hanging in non-interactive environments

## What Was Accomplished
- ✅ Complete Go port of all shell functionality
- ✅ Preserved exact behavior and user experience
- ✅ Maintained compatibility with existing config files
- ✅ Added proper error handling and logging
- ✅ Created installation script and documentation
- ✅ Implemented comprehensive GitHub Actions CI/CD pipeline with security scanning
- ✅ Added version flag functionality for build information retrieval
- ✅ Created secure configuration template system for AWS and Kubernetes configs
- ✅ **NEW**: Implemented profile-based configuration system with interactive wizard
- ✅ **NEW**: Added smart configuration update modes (add new vs override all)
- ✅ **NEW**: Created per-profile ECR login, Kubernetes context, and K9s settings
- ✅ **NEW**: Removed legacy DEV profile special handling for cleaner logic
- ✅ **NEW**: Added intelligent empty context handling (skips interactive selection)
- ✅ **NEW**: Implemented 60-second timeouts for all interactive selections to prevent hanging
- ✅ **NEW**: Added package manager support (Homebrew) for easy installation
- ✅ **NEW**: Implemented automated release management with downloadable archives

## Architecture

### Directory Structure
```
go-fancy-login/
├── cmd/main.go                    # Main application entry point
├── internal/
│   ├── aws/aws.go                 # AWS operations (SSO, ECR, profiles)  
│   ├── config/config.go           # Configuration handling
│   ├── k8s/k8s.go                 # Kubernetes operations
│   └── utils/logger.go            # Logging and utilities
├── examples/
│   ├── aws-config.template        # AWS configuration template
│   ├── kube-config.template       # Kubernetes configuration template
│   └── README.md                  # Template documentation
├── build/                         # Build artifacts (created by make)
├── scripts/                       # Build and utility scripts
├── .gitlab-ci.yml                 # GitLab CI/CD pipeline
├── .fancy-contexts.conf           # Context mappings
├── .fancy-namespaces.conf         # Namespace mappings  
├── Makefile                       # Build system
├── go.mod                         # Go module definition
└── README.md                      # Project documentation
```

### Key Components

**config/fancy_config.go:**
- Modern YAML-based profile configuration system
- Direct profile-to-settings mapping without pattern matching
- Per-profile ECR login, Kubernetes context, and K9s settings
- Configuration file management and path resolution

**config/wizard.go:**
- Interactive configuration wizard with profile discovery
- Smart update modes (add new vs override all)
- Visual status indicators for existing configurations
- Consistent prompt handling with proper defaults

**config/parsers.go:**
- AWS profile parsing from `~/.aws/config`
- Kubernetes context discovery from `~/.kube/config`
- Backward compatibility with legacy configuration files

**aws/aws.go:**
- AWS profile selection using fzf
- SSO authentication handling with spinner UI
- ECR login for _DEV_ profiles
- Account ID retrieval
- Temp file creation for shell integration

**k8s/k8s.go:**
- Kubernetes context selection and switching
- iTerm2 integration (tab titles and badges)
- k9s launching with proper namespace and AWS profile inheritance
- Direct context lookup from profile configuration
- Smart handling of empty contexts (skips selection entirely)
- Timeout-protected interactive selection for legacy configurations

**utils/logger.go:**
- Colorized logging with verbose mode support
- Spinner animations for long-running operations
- Consistent error handling and user feedback

## Critical Implementation Details

### Shell Integration Fix
**Issue:** AWS_PROFILE environment variable doesn't persist after Go binary exits.

**Solution:** Uses the same pattern as original shell version:
1. Go binary writes `export AWS_PROFILE=<profile>` to `/tmp/aws_profile.sh`
2. Shell wrapper function sources this file after successful execution
3. Environment variable persists in current shell session

**Required Shell Setup:**
```bash
fancy-go() {
  if fancy-login-go "$@"; then
    [[ -f /tmp/aws_profile.sh ]] && source /tmp/aws_profile.sh
  fi
}
```

### AWS Profile Inheritance for k9s
**Issue:** k9s subprocess couldn't access AWS credentials.

**Solution:** 
- Set `os.Setenv("AWS_PROFILE", awsProfile)` in main process
- Pass full environment + explicit AWS_PROFILE to k9s command:
```go
cmd.Env = os.Environ()
cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_PROFILE=%s", awsProfile))
```

## Configuration Files (Unchanged)

**.fancy-contexts.conf:** Maps AWS profiles to k8s contexts
```
*_PROD_* = prod-cluster
*_TEST_* = test-cluster  
```

**.fancy-namespaces.conf:** Maps project codes to namespace prefixes
```
IMP=myapp-importer
DET=myapp-details
MD=myapp-masterdata
OV=myapp-overviews
```

## CI/CD Pipeline

### GitHub Actions Workflows
The project includes comprehensive GitHub Actions workflows for CI/CD automation:

**Main CI/CD Workflow (`.github/workflows/ci.yml`):**

**Lint Stage:**
- Uses latest `golangci-lint` for code quality analysis
- 10-minute timeout for comprehensive analysis
- Automatic caching for faster runs

**Test Stage:**
- Runs tests with race detection and coverage
- Generates HTML coverage reports
- Uploads coverage artifacts for review

**Build Stage:**
- Multi-platform builds (Linux, macOS, Windows)
- Supports AMD64 and ARM64 architectures
- Version injection from Git tags or commit hashes
- Creates compressed archives (tar.gz for Unix, zip for Windows)

**Release Stage:**
- Automated release creation on tag push
- Uploads all platform binaries as release assets
- Generates release notes automatically

**Package Management:**
- Automated Homebrew formula updates
- Direct binary downloads with checksums
- Multi-platform release artifacts

**Additional Workflows:**

**Release Workflow (`.github/workflows/release.yml`):**
- Triggered on version tags (`v*.*.*`)
- Generates changelog from Git commits
- Creates release with downloadable assets
- Includes SHA256 checksums for verification

**Security Workflow (`.github/workflows/security.yml`):**
- Daily security scans with Trivy and Gosec
- CodeQL static analysis
- Dependency review for pull requests
- SARIF uploads to GitHub Security tab

**Pipeline Features:**
- Go module caching for faster builds
- Parallel job execution
- Multi-platform binary releases
- Automated security scanning
- Release artifact management
- Package manager integration (Homebrew)

### Version Management

**Version Flag Functionality:**
- `--version` flag displays build-time information
- Shows version, build timestamp, and Git commit hash
- Version information injected via Makefile ldflags during build
- Enables easy identification of deployed versions

```bash
./fancy-login-go --version
# Output:
# fancy-login-go version v1.2.3
# Build time: 2025-09-12T13:25:54Z  
# Git commit: daae12c
```

**Build-time Variables:**
- `version`: Git tag or commit hash with dirty flag
- `buildTime`: ISO 8601 timestamp of build
- `gitCommit`: Short Git commit hash
- Variables automatically populated by Makefile during compilation

## Configuration Template System

### Secure Configuration Distribution
Instead of including sensitive configuration files in the repository, the project provides a template-based approach:

**Template Files:**
- `examples/aws-config.template`: AWS configuration with placeholders
- `examples/kube-config.template`: Kubernetes configuration for EKS
- `examples/README.md`: Comprehensive setup documentation

**Template Installation:**
```bash
make install-templates  # Safe installation (won't overwrite existing configs)
```

**Template Features:**
- **Security**: No sensitive data in repository
- **Placeholder Values**: Clear indicators for customization points  
- **Multi-Environment**: Support for dev/staging/prod configurations
- **EKS Integration**: Pre-configured for AWS EKS clusters
- **Profile Consistency**: Maintains AWS/Kubernetes profile alignment

**Installation Safety:**
- Only installs templates if target config files don't exist
- Provides clear warnings about required customization
- Offers manual installation instructions as alternative
- Includes validation guidance for proper setup

### Configuration Workflow
1. **Initial Setup**: Run `make install-templates` after installation
2. **Customization**: Replace placeholder values with actual configuration
3. **Validation**: Test configuration with AWS and kubectl commands
4. **Profile Alignment**: Ensure AWS profiles match Kubernetes config AWS_PROFILE values

## Usage Patterns

**Installation:**
```bash
make build && make install           # Build and install binary
make install-templates              # Install configuration templates
```

**Development:**
```bash
make build                          # Quick build for testing
make test                           # Run test suite
make lint                           # Code quality analysis
make release                        # Multi-platform release build
```

**Package Manager Installation:**
```bash
# Homebrew (macOS/Linux)
brew install fancy-login-go

# Manual installation from GitHub Releases
# Download the appropriate archive for your platform
curl -L https://github.com/[username]/go-fancy-login/releases/latest/download/fancy-login-go-[version]-[platform].tar.gz
tar -xzf fancy-login-go-[version]-[platform].tar.gz
sudo mv fancy-login-go /usr/local/bin/
```

**Release Downloads:**
- Download platform-specific archives from GitHub Releases
- Verify checksums using provided SHA256 files
- Extract and add to PATH

**Shell Setup:** Add to ~/.zshrc and reload
**Daily Usage:** `fancy-go` or `fancy` (with alias)
**Version Check:** `fancy-go --version`

## Testing Status
- ✅ Help output working correctly
- ✅ AWS profile selection working
- ✅ k8s context switching working  
- ✅ Namespace derivation working
- ✅ k9s launch with proper credentials working
- ✅ Shell environment variable persistence working

## Key Success Factors
1. **Exact Compatibility:** Go version behaves identically to shell version
2. **Environment Integration:** Proper shell wrapper for persistent environment variables
3. **Subprocess Management:** Correct environment inheritance for k9s and other tools
4. **Configuration Reuse:** Same config files work for both versions
5. **Error Handling:** Improved error messages and user feedback

## Migration Benefits Realized
- **Performance:** Faster startup and execution than shell scripts
- **Maintainability:** Clean, testable Go code with proper separation of concerns  
- **Reliability:** Better error handling and edge case management
- **Cross-platform:** Can be compiled for different operating systems
- **Development:** Easier to extend and modify functionality

The Go version is production-ready and can fully replace the shell version while maintaining all existing workflows and muscle memory.