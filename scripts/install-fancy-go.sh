#!/bin/bash

# =============================================================================
# Fancy Login Go Installation Script
# =============================================================================

set -e

# Optional verbose logging
FANCY_VERBOSE=${FANCY_VERBOSE:-0}
fancy_log() {
  if [[ "$FANCY_VERBOSE" == "1" ]]; then
    echo "[fancy-install-go] $1"
  fi
}

echo "🔍 Checking for required tools..."

REQUIRED_TOOLS=("aws" "kubectl" "fzf" "k9s" "go")
MISSING_TOOLS=()

for tool in "${REQUIRED_TOOLS[@]}"; do
  if command -v "$tool" &> /dev/null; then
    echo "✅ $tool is installed."
  else
    echo "❌ $tool is missing."
    MISSING_TOOLS+=("$tool")
  fi
done

if [ ${#MISSING_TOOLS[@]} -ne 0 ]; then
  echo "❌ Missing required tools:"
  for tool in "${MISSING_TOOLS[@]}"; do
    echo "   - $tool"
  done
  echo -e "\nPlease install the missing tools before running the install script."
  echo "Note: Go is required to build the fancy-login binary."
  exit 1
fi

echo "✅ All required tools are installed."
echo "🔧 Installing fancy-login Go version..."

# Define paths
SCRIPT_DIR="${FANCY_SCRIPT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="${FANCY_BIN_DIR:-$HOME/.local/bin}"
AWS_DIR="${FANCY_AWS_DIR:-$HOME/.aws}"
KUBE_DIR="${FANCY_KUBE_DIR:-$HOME/.kube}"

# Create bin dir if needed
fancy_log "Creating bin dir at $BIN_DIR"
mkdir -p "$BIN_DIR"

# Build the Go binary
echo "🔨 Building Go binary..."
fancy_log "Building from $PROJECT_DIR"
cd "$PROJECT_DIR"

# Ensure go.mod exists and is correct
if [[ ! -f "go.mod" ]]; then
  fancy_log "Initializing go module"
  go mod init fancy-login
fi

# Build the binary
fancy_log "Building binary to $BIN_DIR/fancy-login-go"
go build -o "$BIN_DIR/fancy-login-go" ./cmd

# Make it executable
fancy_log "Making binary executable"
chmod +x "$BIN_DIR/fancy-login-go"

# Copy config files
fancy_log "Copying .fancy-namespaces.conf to $BIN_DIR/.fancy-namespaces.conf"
cp "$PROJECT_DIR/.fancy-namespaces.conf" "$BIN_DIR/.fancy-namespaces.conf"

if [ -f "$PROJECT_DIR/.fancy-contexts.conf" ]; then
  fancy_log "Copying .fancy-contexts.conf to $BIN_DIR/.fancy-contexts.conf"
  cp "$PROJECT_DIR/.fancy-contexts.conf" "$BIN_DIR/.fancy-contexts.conf"
fi

# Copy AWS config if it exists
if [ -f "$PROJECT_DIR/../aws/config" ]; then
  echo "📦 Copying AWS config..."
  fancy_log "Creating AWS dir at $AWS_DIR"
  mkdir -p "$AWS_DIR"
  fancy_log "Copying aws/config to $AWS_DIR/config"
  cp "$PROJECT_DIR/../aws/config" "$AWS_DIR/config"
fi

# Copy kube config if it exists
if [ -f "$PROJECT_DIR/../k8s/config" ]; then
  echo "📦 Copying Kubernetes config..."
  fancy_log "Creating kube dir at $KUBE_DIR"
  mkdir -p "$KUBE_DIR"
  fancy_log "Copying k8s/config to $KUBE_DIR/config"
  cp "$PROJECT_DIR/../k8s/config" "$KUBE_DIR/config"
fi

# Print installation complete message
echo "\n✅ Go version installation complete!"
echo "\n🔧 Add the following to your ~/.zshrc:"
echo "------------------------------------------------------------"
echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "# Fancy login function (Go version)"
echo "fancy-go() {"
echo "  if fancy-login-go \"\$@\"; then"
echo "    [[ -f /tmp/aws_profile.sh ]] && source /tmp/aws_profile.sh"
echo "  fi"
echo "}"
echo ""
echo "# Or create an alias for the Go version"
echo "alias fancy='fancy-go'"
echo "------------------------------------------------------------"
echo "\nThen run: source ~/.zshrc"
echo "\n🚀 You can now run the Go version using: fancy-login-go or fancy-go"
echo "   Test with: fancy-login-go --help"