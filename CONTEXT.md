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

3. **ECR Authentication for Development**
   - Automatic ECR login for profiles containing `_DEV_`
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
- Wildcard pattern matching for AWS profiles (`*_PROD_*`, `*_TEST_*`)
- DEVENG profile special handling with namespace derivation
- Development profile detection for ECR login
- Profile validation and session checking

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
5. ECR login (if DEV profile)
6. See colorized summary with derived namespace
7. Optional k9s launch in correct namespace

**DEVENG Profile Workflow:**
```bash
fancy-go -k  # Auto-launch k9s
# → Selects OV_TEST_DEVENG
# → Maps to test-cluster context  
# → Derives test-mykn-track-overviews namespace
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
- Example: `OV_TEST_DEVENG` → `test-mykn-track-overviews`
- Uses `.fancy-namespaces.conf` for project code → name mapping
- Case conversion: ENV (uppercase) → env (lowercase)
- Only applies to profiles ending in `DEVENG`

**ECR Login Behavior:**
- Triggered only for profiles containing `_DEV_`
- Gets AWS account ID from `aws sts get-caller-identity`
- Uses configured region or `FANCY_DEFAULT_REGION`
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

## What Was Accomplished
- ✅ Complete Go port of all shell functionality
- ✅ Preserved exact behavior and user experience  
- ✅ Maintained compatibility with existing config files
- ✅ Added proper error handling and logging
- ✅ Created installation script and documentation

## Architecture

### Directory Structure
```
go/
├── cmd/main.go                    # Main application entry point
├── internal/
│   ├── aws/aws.go                 # AWS operations (SSO, ECR, profiles)  
│   ├── config/config.go           # Configuration handling
│   ├── k8s/k8s.go                 # Kubernetes operations
│   └── utils/logger.go            # Logging and utilities
├── scripts/install-fancy-go.sh    # Installation script
├── .fancy-contexts.conf           # Context mappings (copied from parent)
├── .fancy-namespaces.conf         # Namespace mappings (copied from parent)
└── README.md                      # Documentation
```

### Key Components

**config/config.go:**
- Handles environment variables and configuration loading
- Parses `.fancy-contexts.conf` and `.fancy-namespaces.conf` files
- Provides wildcard pattern matching for AWS profiles
- Derives namespaces from DEVENG profile patterns

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
- Context mapping based on AWS profiles

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
IMP=mykn-track-importer
DET=mykn-track-details
MD=mykn-masterdata
OV=mykn-track-overviews
```

## Usage Patterns

**Installation:**
```bash
cd go && ./scripts/install-fancy-go.sh
```

**Shell Setup:** Add to ~/.zshrc and reload
**Daily Usage:** `fancy-go` or `fancy` (with alias)

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