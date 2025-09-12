# Build Guide - Fancy Login Go

This guide covers building the fancy-login Go version for different platforms and environments.

## Quick Build

### Local Development Build
```bash
# From go/ directory
go build -o fancy-login-go ./cmd

# With version info
go build -ldflags "-X main.version=1.0.0" -o fancy-login-go ./cmd
```

### Cross-Platform Builds
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o fancy-login-go-linux-amd64 ./cmd

# Windows AMD64  
GOOS=windows GOARCH=amd64 go build -o fancy-login-go-windows-amd64.exe ./cmd

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o fancy-login-go-darwin-amd64 ./cmd

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o fancy-login-go-darwin-arm64 ./cmd
```

## Build Scripts

### All Platforms Script
```bash
#!/bin/bash
# scripts/build-all.sh

VERSION=${1:-"dev"}
BUILD_DIR="build"

echo "Building fancy-login v$VERSION for all platforms..."

# Clean build directory
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Build info
LDFLAGS="-X main.version=$VERSION -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Build for each platform
platforms=(
    "linux/amd64"
    "linux/arm64" 
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="fancy-login-go-$GOOS-$GOARCH"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo "Building $output_name..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/$output_name" ./cmd
    
    if [ $? -ne 0 ]; then
        echo "Failed to build for $platform"
        exit 1
    fi
done

echo "Build complete! Binaries in $BUILD_DIR/"
ls -la $BUILD_DIR/
```

## Development Setup

### Prerequisites
```bash
# Install Go 1.19+
go version

# Install dependencies for development
go mod download

# Install build tools (optional)
go install github.com/goreleaser/goreleaser@latest
```

### Development Workflow
```bash
# Format code
go fmt ./...

# Run tests (if any)
go test ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Build for current platform
go build ./cmd

# Quick test
./fancy-login-go --help
```

## Advanced Building

### With Build Info
```bash
# Set version and build metadata
VERSION=$(git describe --tags --always --dirty)
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse HEAD)

go build -ldflags "
  -X main.version=$VERSION 
  -X main.buildTime=$BUILD_TIME
  -X main.gitCommit=$GIT_COMMIT
" -o fancy-login-go ./cmd
```

### Optimized Release Build
```bash
# Smaller binary with optimizations
go build -ldflags "-s -w" -trimpath -o fancy-login-go ./cmd

# With UPX compression (if installed)
upx --best fancy-login-go
```

### Static Linking (Linux)
```bash
# For maximum compatibility
CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -o fancy-login-go ./cmd
```

## Docker Builds

### Multi-Stage Dockerfile
```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fancy-login-go ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates aws-cli kubectl
WORKDIR /root/

COPY --from=builder /app/fancy-login-go .
CMD ["./fancy-login-go"]
```

### Build Docker Image
```bash
# From project root
docker build -t fancy-login:latest -f go/Dockerfile go/

# Multi-architecture build
docker buildx build --platform linux/amd64,linux/arm64 -t fancy-login:latest go/
```

## Automated Builds

### GitHub Actions Example
```yaml
name: Build
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
        
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
        
    - name: Build
      env:
        GOOS: ${{ matrix.goos }}
        GOARCH: ${{ matrix.goarch }}
      run: |
        cd go
        go build -o fancy-login-go-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd
        
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: binaries
        path: go/fancy-login-go-*
```

## Release Process

### Manual Release
```bash
# Tag the release
git tag v1.0.0
git push origin v1.0.0

# Build all platforms
./scripts/build-all.sh v1.0.0

# Create release archives
cd build
for binary in fancy-login-go-*; do
    if [[ $binary == *.exe ]]; then
        zip "${binary%.exe}.zip" "$binary"
    else
        tar -czf "${binary}.tar.gz" "$binary"  
    fi
done
```

### Using GoReleaser
```yaml
# .goreleaser.yml
project_name: fancy-login

builds:
  - main: ./cmd
    binary: fancy-login-go
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

release:
  draft: true
```

```bash
# Build and release
goreleaser release --snapshot --rm-dist
```

## Troubleshooting

### Common Build Issues

**1. Module not found**
```bash
go mod tidy
go mod download
```

**2. Cross-compilation errors**
```bash
# Ensure no CGO dependencies
CGO_ENABLED=0 go build ./cmd
```

**3. Large binary size**
```bash
# Use build flags to reduce size
go build -ldflags "-s -w" -trimpath ./cmd
```

**4. Windows execution issues**
```bash
# Ensure .exe extension for Windows builds
GOOS=windows go build -o fancy-login-go.exe ./cmd
```

### Build Environment

**Required Environment Variables:**
- `GOOS`: Target operating system
- `GOARCH`: Target architecture  
- `CGO_ENABLED`: Enable/disable CGO (0 for static builds)

**Supported Platforms:**
- linux/amd64, linux/arm64
- darwin/amd64, darwin/arm64 
- windows/amd64, windows/arm64
- freebsd/amd64
- openbsd/amd64

## Performance Considerations

### Binary Size
- Go binaries are typically 3-5MB for this project
- Use `-ldflags "-s -w"` to strip debug info (~20% reduction)
- Use `-trimpath` to remove build path info
- UPX can compress further (~50% reduction)

### Build Speed
- Use `go build -i` for incremental builds
- Leverage build cache with `GOCACHE`
- Parallel builds with `GOMAXPROCS`

### Runtime Performance
- Cross-compiled binaries perform identically to native builds
- Static linking has minimal performance impact
- Go's garbage collector is well-tuned for CLI applications