#!/bin/bash
# Build fancy-login-go for all supported platforms

set -e

VERSION=${1:-"dev"}
BUILD_DIR="build"

echo "🔨 Building fancy-login v$VERSION for all platforms..."

# Clean and create build directory
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Build metadata
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.gitCommit=$GIT_COMMIT -s -w"

echo "📦 Version: $VERSION"
echo "⏰ Build Time: $BUILD_TIME" 
echo "🔗 Git Commit: $GIT_COMMIT"
echo ""

# Platform configurations
platforms=(
    "linux/amd64"
    "linux/arm64" 
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Build for each platform
for platform in "${platforms[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="fancy-login-go-$GOOS-$GOARCH"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo "🔨 Building $output_name..."
    
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -trimpath \
        -o "$BUILD_DIR/$output_name" \
        ./cmd
    
    if [ $? -ne 0 ]; then
        echo "❌ Failed to build for $platform"
        exit 1
    fi
    
    # Show file size
    if [ $GOOS = "darwin" ]; then
        size=$(stat -f%z "$BUILD_DIR/$output_name" 2>/dev/null | numfmt --to=iec-i --suffix=B || echo "unknown")
    else
        size=$(stat -c%s "$BUILD_DIR/$output_name" 2>/dev/null | numfmt --to=iec-i --suffix=B || echo "unknown")
    fi
    echo "   📏 Size: $size"
done

echo ""
echo "✅ Build complete! Binaries in $BUILD_DIR/"
echo "📁 Contents:"
ls -la $BUILD_DIR/ | while read line; do
    echo "   $line"
done

# Calculate total size
total_size=0
for file in $BUILD_DIR/*; do
    if [ -f "$file" ]; then
        if [ "$(uname)" = "Darwin" ]; then
            size=$(stat -f%z "$file" 2>/dev/null || echo "0")
        else
            size=$(stat -c%s "$file" 2>/dev/null || echo "0")
        fi
        total_size=$((total_size + size))
    fi
done

total_human=$(echo $total_size | numfmt --to=iec-i --suffix=B 2>/dev/null || echo "${total_size} bytes")
echo ""
echo "📊 Total size: $total_human"
echo "🎉 Ready for distribution!"