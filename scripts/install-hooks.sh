#!/bin/bash

# Install Git hooks for go-fancy-login
# This script sets up pre-commit hooks for linting and testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "🔧 Installing Git hooks..."

# Check if we're in a Git repository
if [ ! -d "$REPO_ROOT/.git" ]; then
    echo "❌ Error: Not in a Git repository"
    exit 1
fi

# Create pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash

# Pre-commit hook for go-fancy-login
# Runs linting and formatting checks before allowing commits

set -e

echo "🔍 Running pre-commit checks..."

# Check if golangci-lint is available
if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "⚠️  golangci-lint not found. Installing..."

    # Try to install golangci-lint
    if command -v go >/dev/null 2>&1; then
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        # Add GOPATH/bin to PATH if it's not already there
        export PATH="$PATH:$(go env GOPATH)/bin"
    else
        echo "❌ Go not found. Cannot install golangci-lint."
        echo "Please install golangci-lint manually: https://golangci-lint.run/usage/install/"
        exit 1
    fi
fi

# Run go fmt and check if files were modified
echo "📝 Running go fmt..."
UNFORMATTED=$(go fmt ./...)
if [ -n "$UNFORMATTED" ]; then
    echo "❌ The following files were not formatted:"
    echo "$UNFORMATTED"
    echo "Files have been automatically formatted. Please add them and commit again."
    exit 1
fi

# Run go vet
echo "🔧 Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet found issues. Please fix them before committing."
    exit 1
fi

# Run golangci-lint
echo "🔍 Running golangci-lint..."
if ! golangci-lint run --timeout=5m; then
    echo "❌ Linting failed. Please fix the issues before committing."
    exit 1
fi

# Run tests
echo "🧪 Running tests..."
if ! go test ./... -short; then
    echo "❌ Tests failed. Please fix them before committing."
    exit 1
fi

echo "✅ All pre-commit checks passed!"
EOF

# Make the hook executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Pre-commit hook installed successfully!"
echo ""
echo "The hook will now run automatically before each commit and will:"
echo "  📝 Format code with go fmt"
echo "  🔧 Check code with go vet"
echo "  🔍 Run golangci-lint"
echo "  🧪 Run tests"
echo ""
echo "To skip the hook for a specific commit, use: git commit --no-verify"
echo "To uninstall, delete: .git/hooks/pre-commit"