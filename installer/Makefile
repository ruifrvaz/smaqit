# smaqit installer build system

# Detect version from git tags, fallback to commit hash if no tags exist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build configuration
BINARY_NAME = smaqit
DEV_BINARY_NAME = smaqit-dev
DIST_DIR = dist
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Copy embedded files (required for go:embed at compile time)
.PHONY: prepare
prepare:
	@echo "Copying embedded files..."
	@cp -r ../framework .
	@mkdir -p templates
	@cp -r ../templates/specs templates/
	@cp -r ../agents .
	@cp -r ../prompts .

# Default target: build for current platform
.PHONY: build
build: prepare
	@echo "Building $(DEV_BINARY_NAME) version $(VERSION) for current platform..."
	@mkdir -p $(DIST_DIR)
	@go build $(LDFLAGS) -o $(DIST_DIR)/$(DEV_BINARY_NAME) .
	@echo "Built: $(DIST_DIR)/$(DEV_BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: prepare
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_linux_amd64 .
	@echo "Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_amd64 .
	@echo "Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_darwin_arm64 .
	@echo "Building for windows/amd64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)_windows_amd64.exe .
	@echo "✓ Built all platforms"

# Show version that would be built
.PHONY: version
version:
	@echo "$(VERSION)"

# Install to local bin
.PHONY: install
install: build
	@if [ -n "$$GOPATH" ]; then \
		echo "Installing $(DEV_BINARY_NAME) to $$GOPATH/bin..."; \
		mkdir -p $$GOPATH/bin; \
		cp $(DIST_DIR)/$(DEV_BINARY_NAME) $$GOPATH/bin/$(DEV_BINARY_NAME); \
		echo "✓ Installed: $$GOPATH/bin/$(DEV_BINARY_NAME)"; \
	else \
		echo "Installing $(DEV_BINARY_NAME) to ~/.local/bin..."; \
		mkdir -p $$HOME/.local/bin; \
		cp $(DIST_DIR)/$(DEV_BINARY_NAME) $$HOME/.local/bin/$(DEV_BINARY_NAME); \
		echo "✓ Installed: ~/.local/bin/$(DEV_BINARY_NAME)"; \
		if ! echo $$PATH | grep -q "$$HOME/.local/bin"; then \
			echo ""; \
			echo "⚠ ~/.local/bin is not in your PATH"; \
			echo "  Add to your shell config:"; \
			echo "  echo 'export PATH=\"\$$HOME/.local/bin:\$$PATH\"' >> ~/.bashrc"; \
			echo ""; \
			echo "Then reload:"; \
			echo "  source ~/.bashrc"; \
		fi; \
	fi

# Uninstall from local bin
.PHONY: uninstall
uninstall:
	@if [ -n "$$GOPATH" ] && [ -f "$$GOPATH/bin/$(DEV_BINARY_NAME)" ]; then \
		rm $$GOPATH/bin/$(DEV_BINARY_NAME); \
		echo "✓ Removed: $$GOPATH/bin/$(DEV_BINARY_NAME)"; \
	elif [ -f "$$HOME/.local/bin/$(DEV_BINARY_NAME)" ]; then \
		rm $$HOME/.local/bin/$(DEV_BINARY_NAME); \
		echo "✓ Removed: ~/.local/bin/$(DEV_BINARY_NAME)"; \
	else \
		echo "$(DEV_BINARY_NAME) not found in PATH locations"; \
	fi; \
	if grep -q 'HOME/.local/bin.*PATH' $$HOME/.bashrc 2>/dev/null; then \
		printf "\nAlso remove PATH export from ~/.bashrc? [y/N]: "; \
		read response; \
		case "$$response" in \
			[yY]|[yY][eE][sS]) \
				grep -v 'HOME/.local/bin.*PATH' $$HOME/.bashrc > $$HOME/.bashrc.tmp && \
				mv $$HOME/.bashrc.tmp $$HOME/.bashrc && \
				echo "✓ Removed PATH export from ~/.bashrc" && \
				echo "ℹ Run 'source ~/.bashrc' to apply changes"; \
				;; \
			*) \
				echo "✓ Kept PATH export in ~/.bashrc"; \
				;; \
		esac; \
	fi; \
	if [ -d "$(DIST_DIR)" ] || [ -d "framework" ] || [ -d "templates" ] || [ -d "agents" ] || [ -d "prompts" ] || [ -d "test" ]; then \
		printf "\nAlso clean build artifacts (dist/, embedded files, test/)? [y/N]: "; \
		read clean_response; \
		case "$$clean_response" in \
			[yY]|[yY][eE][sS]) \
				rm -rf $(DIST_DIR) framework/ templates/ agents/ prompts/ test/; \
				echo "✓ Cleaned build artifacts"; \
				;; \
			*) \
				echo "✓ Kept build artifacts"; \
				;; \
		esac; \
	fi

# Help
help:
	@echo "smaqit installer build targets:"
	@echo ""
	@echo "  make build      - Build for current platform"
	@echo "  make build-all  - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make install    - Build and install to ~/.local/bin (adds to PATH if needed)"
	@echo "  make uninstall  - Remove binary (prompts for PATH and artifact cleanup)"
	@echo "  make version    - Show version that will be built"
	@echo ""
	@echo "Current version: $(VERSION)"
