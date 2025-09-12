#!/bin/bash
# Simple build script for development

set -e

# Build for current platform
echo "🔨 Building fancy-login-go for $(go env GOOS)/$(go env GOARCH)..."

# Get version info
VERSION=${1:-"dev"}
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build with version info
go build -ldflags "
  -X main.version=$VERSION 
  -X main.buildTime=$BUILD_TIME
  -X main.gitCommit=$GIT_COMMIT
" -o fancy-login-go ./cmd

echo "✅ Build complete: fancy-login-go"
echo "📦 Version: $VERSION"
echo "⏰ Built: $BUILD_TIME"

# Test the binary
echo ""
echo "🧪 Testing binary:"
./fancy-login-go --help