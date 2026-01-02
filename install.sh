#!/bin/bash
# smaqit installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="ruifrvaz/smaqit"
INSTALL_DIR="${HOME}/.local/bin"

# Helper functions
info() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported operating system: $os"
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
    
    info "Detected platform: ${OS}/${ARCH}"
}

# Get latest release version from GitHub API
get_latest_version() {
    info "Fetching latest release..."
    
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    VERSION=$(curl -fsSL "$api_url" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version"
    fi
    
    info "Latest version: ${VERSION}"
}

# Download binary
download_binary() {
    local binary_name="smaqit_${OS}_${ARCH}"
    
    if [ "$OS" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${binary_name}"
    local temp_file="/tmp/smaqit_${VERSION}"
    
    info "Downloading from ${download_url}..."
    
    if ! curl -fsSL -o "$temp_file" "$download_url"; then
        error "Failed to download binary"
    fi
    
    info "Download complete"
    echo "$temp_file"
}

# Install binary
install_binary() {
    local temp_file=$1
    local target="${INSTALL_DIR}/smaqit"
    
    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
    
    # Make executable
    chmod +x "$temp_file"
    
    # Move to install directory
    mv "$temp_file" "$target"
    
    info "Installed to ${target}"
}

# Verify installation
verify_installation() {
    local target="${INSTALL_DIR}/smaqit"
    
    if ! "$target" --version &>/dev/null; then
        error "Installation verification failed"
    fi
    
    local installed_version=$("$target" --version 2>&1 || echo "unknown")
    info "Verified installation: ${installed_version}"
}

# Check if install directory is in PATH
check_path() {
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        warn "${INSTALL_DIR} is not in your PATH"
        echo ""
        echo "Add to your shell config (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        echo "Then reload your shell:"
        echo "  source ~/.bashrc  # or ~/.zshrc"
        echo ""
    fi
}

# Main installation flow
main() {
    echo "smaqit installer"
    echo "================"
    echo ""
    
    detect_platform
    get_latest_version
    
    local temp_file=$(download_binary)
    install_binary "$temp_file"
    verify_installation
    check_path
    
    echo ""
    info "Installation complete!"
    echo ""
    echo "Get started:"
    echo "  smaqit init       # Initialize in your project"
    echo "  smaqit --help     # View available commands"
    echo ""
}

main
