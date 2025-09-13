# Makefile for fancy-login Go version

# Variables
BINARY_NAME := fancy-login-go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DIR := build

# Go build flags
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT) -s -w
BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

# Platform configurations
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# Default target
.PHONY: all
all: build

# Help target
.PHONY: help
help:
	@echo "🔨 Fancy Login Go Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build for current platform"
	@echo "  build-all     - Build for all supported platforms"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  install-templates - Install configuration templates to ~/.aws and ~/.kube"
	@echo "  release       - Create release archives"
	@echo "  docker        - Build Docker image"
	@echo "  version       - Show version info"
	@echo "  help          - Show this help"

# Build for current platform
.PHONY: build
build:
	@echo "🔨 Building $(BINARY_NAME) v$(VERSION) for $$(go env GOOS)/$$(go env GOARCH)..."
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd
	@echo "✅ Build complete: $(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "🔨 Building $(BINARY_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@$(foreach platform,$(PLATFORMS), \
		$(call build_platform,$(platform)) \
	)
	@echo "✅ All builds complete! Artifacts in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Platform build function
define build_platform
	$(eval GOOS := $(word 1,$(subst /, ,$1)))
	$(eval GOARCH := $(word 2,$(subst /, ,$1)))
	$(eval OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH))
	$(eval OUTPUT := $(if $(filter windows,$(GOOS)),$(OUTPUT).exe,$(OUTPUT)))
	@echo "  Building $(GOOS)/$(GOARCH)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(OUTPUT) ./cmd
endef

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(BINARY_NAME) $(BINARY_NAME).exe
	go clean -cache
	@echo "✅ Clean complete"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go fmt ./...
	go vet ./...
	go test -v ./...

# Run linter
.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run -v; \
	else \
		echo "⚠️ golangci-lint not found, skipping lint"; \
	fi

# Install binary
.PHONY: install
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	cp $(BINARY_NAME) $$GOPATH/bin/
	@echo "✅ Installed to $$GOPATH/bin/$(BINARY_NAME)"

# Install configuration templates
.PHONY: install-templates
install-templates:
	@echo "📋 Installing configuration templates..."
	@mkdir -p ~/.aws ~/.kube
	@if [ ! -f ~/.aws/config ]; then \
		echo "Installing AWS config template to ~/.aws/config"; \
		cp examples/aws-config.template ~/.aws/config; \
		echo "⚠️  Please edit ~/.aws/config with your actual AWS configuration"; \
	else \
		echo "~/.aws/config already exists, skipping AWS config installation"; \
		echo "💡 You can manually copy examples/aws-config.template if needed"; \
	fi
	@if [ ! -f ~/.kube/config ]; then \
		echo "Installing Kubernetes config template to ~/.kube/config"; \
		cp examples/kube-config.template ~/.kube/config; \
		echo "⚠️  Please edit ~/.kube/config with your actual cluster configuration"; \
	else \
		echo "~/.kube/config already exists, skipping Kubernetes config installation"; \
		echo "💡 You can manually copy examples/kube-config.template if needed"; \
	fi
	@echo "✅ Configuration template installation complete"
	@echo "📖 See examples/ directory for template documentation"

# Create release archives
.PHONY: release
release: build-all
	@echo "📦 Creating release archives..."
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			if echo "$$binary" | grep -q "windows"; then \
				zip "release/$${binary%.exe}.zip" "$$binary"; \
			else \
				tar -czf "release/$$binary.tar.gz" "$$binary"; \
			fi; \
		fi; \
	done
	@cd $(BUILD_DIR) && sha256sum $(BINARY_NAME)-* > release/checksums.txt
	@echo "✅ Release archives created in $(BUILD_DIR)/release/"
	@ls -la $(BUILD_DIR)/release/

# Build Docker image
.PHONY: docker
docker:
	@echo "🐳 Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "✅ Docker image built: $(BINARY_NAME):$(VERSION)"

# Show version info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Development shortcuts
.PHONY: dev
dev: build
	@echo "🚀 Running development build..."
	./$(BINARY_NAME) --help

.PHONY: run
run: build
	./$(BINARY_NAME) $(ARGS)

# Quick build and test
.PHONY: quick
quick:
	go build -o $(BINARY_NAME) ./cmd
	./$(BINARY_NAME) --help