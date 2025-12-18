# smaqit installer build system

# Detect version from git tags, fallback to commit hash if no tags exist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build configuration
BINARY_NAME = smaqit
DIST_DIR = dist
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Platform targets
PLATFORMS = linux_amd64 darwin_amd64 darwin_arm64 windows_amd64

# Default target: build for current platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) version $(VERSION) for current platform..."
	@mkdir -p $(DIST_DIR)
	@go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME) .
	@echo "Built: $(DIST_DIR)/$(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: $(PLATFORMS)

# Platform-specific builds
.PHONY: linux_amd64
linux_amd64:
	@echo "Building for linux/amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_linux_amd64 .

.PHONY: darwin_amd64
darwin_amd64:
	@echo "Building for darwin/amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_amd64 .

.PHONY: darwin_arm64
darwin_arm64:
	@echo "Building for darwin/arm64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_arm64 .

.PHONY: windows_amd64
windows_amd64:
	@echo "Building for windows/amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_windows_amd64.exe .

# Show version that would be built
.PHONY: version
version:
	@echo "$(VERSION)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR)
	@echo "Done."

# Install to local bin (requires GOPATH or GOBIN set)
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $$GOPATH/bin..."
	@cp $(DIST_DIR)/$(BINARY_NAME) $$GOPATH/bin/$(BINARY_NAME)
	@echo "Installed: $$GOPATH/bin/$(BINARY_NAME)"

# Help
.PHONY: help
help:
	@echo "smaqit installer build targets:"
	@echo ""
	@echo "  make build          - Build for current platform"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make linux_amd64    - Build for Linux (amd64)"
	@echo "  make darwin_amd64   - Build for macOS Intel (amd64)"
	@echo "  make darwin_arm64   - Build for macOS Apple Silicon (arm64)"
	@echo "  make windows_amd64  - Build for Windows (amd64)"
	@echo "  make version        - Show version that would be built"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install        - Install to GOPATH/bin"
	@echo ""
	@echo "Current version: $(VERSION)"
