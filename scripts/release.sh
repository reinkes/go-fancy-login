#!/bin/bash
# Release automation script for fancy-login

set -e

# Configuration
BINARY_NAME="fancy-login-go"
BUILD_DIR="build"
RELEASE_DIR="$BUILD_DIR/release"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Helper functions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Parse arguments
VERSION=""
TAG_RELEASE=false
PUSH_DOCKER=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -t|--tag)
            TAG_RELEASE=true
            shift
            ;;
        -d|--docker)
            PUSH_DOCKER=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -v, --version VERSION   Set version (default: auto-detect)"
            echo "  -t, --tag              Create git tag"
            echo "  -d, --docker           Build and push Docker image"
            echo "  -h, --help             Show this help"
            echo ""
            echo "Examples:"
            echo "  $0 -v v1.0.0 -t        # Create tagged release v1.0.0"
            echo "  $0                     # Create development release"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Auto-detect version if not provided
if [ -z "$VERSION" ]; then
    if git describe --tags --exact-match HEAD 2>/dev/null; then
        VERSION=$(git describe --tags --exact-match HEAD)
        log_info "Using git tag: $VERSION"
    else
        VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
        log_info "Using auto-detected version: $VERSION"
    fi
fi

log_info "🚀 Starting release process for $BINARY_NAME v$VERSION"

# Validate git state
if [ "$TAG_RELEASE" = true ]; then
    if ! git diff-index --quiet HEAD --; then
        log_error "Working directory is dirty. Please commit or stash changes."
        exit 1
    fi
    
    if git rev-parse "$VERSION" >/dev/null 2>&1; then
        log_error "Tag $VERSION already exists!"
        exit 1
    fi
fi

# Step 1: Clean previous builds
log_info "🧹 Cleaning previous builds..."
make clean
log_success "Clean complete"

# Step 2: Run tests
log_info "🧪 Running tests..."
if ! make test; then
    log_error "Tests failed!"
    exit 1
fi
log_success "Tests passed"

# Step 3: Build all platforms
log_info "🔨 Building for all platforms..."
if ! make build-all VERSION="$VERSION"; then
    log_error "Build failed!"
    exit 1
fi
log_success "Build complete"

# Step 4: Create release archives
log_info "📦 Creating release archives..."
if ! make release; then
    log_error "Release archive creation failed!"
    exit 1
fi
log_success "Release archives created"

# Step 5: Validate binaries
log_info "🔍 Validating binaries..."
total_size=0
for binary in $BUILD_DIR/$BINARY_NAME-*; do
    if [ -f "$binary" ]; then
        size=$(stat -c%s "$binary" 2>/dev/null || stat -f%z "$binary")
        total_size=$((total_size + size))
        
        if [ $size -lt 1000000 ]; then
            log_warning "Binary $binary seems small ($size bytes)"
        fi
    fi
done

human_size=$(echo $total_size | numfmt --to=iec-i --suffix=B 2>/dev/null || echo "${total_size} bytes")
log_info "Total binary size: $human_size"

# Step 6: Create git tag (if requested)
if [ "$TAG_RELEASE" = true ]; then
    log_info "🏷️  Creating git tag $VERSION..."
    git tag -a "$VERSION" -m "Release $VERSION"
    log_success "Git tag created: $VERSION"
    
    log_info "Push tag with: git push origin $VERSION"
fi

# Step 7: Docker build (if requested)  
if [ "$PUSH_DOCKER" = true ]; then
    log_info "🐳 Building Docker image..."
    if ! make docker VERSION="$VERSION"; then
        log_error "Docker build failed!"
        exit 1
    fi
    log_success "Docker image built"
fi

# Step 8: Generate release notes
log_info "📝 Generating release information..."

cat > "$RELEASE_DIR/RELEASE_INFO.md" << EOF
# Release $VERSION

## Build Information
- **Version**: $VERSION
- **Build Time**: $(date -u +%Y-%m-%dT%H:%M:%SZ)
- **Git Commit**: $(git rev-parse --short HEAD)
- **Go Version**: $(go version)

## Supported Platforms
EOF

for binary in $BUILD_DIR/$BINARY_NAME-*; do
    if [ -f "$binary" ]; then
        basename_binary=$(basename "$binary")
        platform=$(echo "$basename_binary" | sed "s/$BINARY_NAME-//" | sed 's/\.exe$//')
        size=$(stat -c%s "$binary" 2>/dev/null || stat -f%z "$binary")
        human_size=$(echo $size | numfmt --to=iec-i --suffix=B 2>/dev/null || echo "${size} bytes")
        
        echo "- **$platform**: $human_size" >> "$RELEASE_DIR/RELEASE_INFO.md"
    fi
done

cat >> "$RELEASE_DIR/RELEASE_INFO.md" << EOF

## Installation
1. Download the appropriate binary for your platform
2. Make it executable: \`chmod +x $BINARY_NAME-*\`
3. Move to your PATH or use the installation scripts

## Verification
Verify downloads with SHA256 checksums:
\`\`\`bash
sha256sum -c checksums.txt
\`\`\`

## What's Changed
$(git log --oneline $(git describe --tags --abbrev=0 HEAD^)..HEAD 2>/dev/null | head -10 || echo "- Initial release")
EOF

# Summary
log_success "🎉 Release $VERSION complete!"
echo ""
echo "📁 Release artifacts:"
ls -la "$RELEASE_DIR"
echo ""
echo "🔗 Next steps:"
echo "   1. Review artifacts in $RELEASE_DIR/"
echo "   2. Upload to your preferred distribution method"
if [ "$TAG_RELEASE" = true ]; then
    echo "   3. Push git tag: git push origin $VERSION"
fi
echo "   4. Update documentation with new version"
echo ""
log_info "Release information saved to $RELEASE_DIR/RELEASE_INFO.md"