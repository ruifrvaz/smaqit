#!/usr/bin/env bash
# smaqit installer build script for Unix-like systems

set -e

# Detect version from git tags, fallback to commit hash if no tags exist
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}

BINARY_NAME="smaqit"
DIST_DIR="dist"
LDFLAGS="-ldflags -X main.Version=$VERSION"

# Help message
show_help() {
    echo "smaqit installer build script"
    echo ""
    echo "Usage: ./build.sh [command]"
    echo ""
    echo "Commands:"
    echo "  build          - Build for current platform"
    echo "  build-all      - Build for all platforms"
    echo "  linux          - Build for Linux (amd64)"
    echo "  darwin-intel   - Build for macOS Intel (amd64)"
    echo "  darwin-arm     - Build for macOS Apple Silicon (arm64)"
    echo "  windows        - Build for Windows (amd64)"
    echo "  version        - Show version that would be built"
    echo "  clean          - Remove build artifacts"
    echo "  help           - Show this help message"
    echo ""
    echo "Current version: $VERSION"
}

# Build for current platform
build_current() {
    echo "Building $BINARY_NAME version $VERSION for current platform..."
    mkdir -p "$DIST_DIR"
    go build $LDFLAGS -o "$DIST_DIR/$BINARY_NAME" .
    echo "Built: $DIST_DIR/$BINARY_NAME"
}

# Build for all platforms
build_all() {
    echo "Building $BINARY_NAME version $VERSION for all platforms..."
    build_linux
    build_darwin_intel
    build_darwin_arm
    build_windows
    echo "All builds complete."
}

# Platform-specific builds
build_linux() {
    echo "Building for linux/amd64..."
    mkdir -p "$DIST_DIR"
    GOOS=linux GOARCH=amd64 go build $LDFLAGS -o "$DIST_DIR/${BINARY_NAME}_linux_amd64" .
}

build_darwin_intel() {
    echo "Building for darwin/amd64..."
    mkdir -p "$DIST_DIR"
    GOOS=darwin GOARCH=amd64 go build $LDFLAGS -o "$DIST_DIR/${BINARY_NAME}_darwin_amd64" .
}

build_darwin_arm() {
    echo "Building for darwin/arm64..."
    mkdir -p "$DIST_DIR"
    GOOS=darwin GOARCH=arm64 go build $LDFLAGS -o "$DIST_DIR/${BINARY_NAME}_darwin_arm64" .
}

build_windows() {
    echo "Building for windows/amd64..."
    mkdir -p "$DIST_DIR"
    GOOS=windows GOARCH=amd64 go build $LDFLAGS -o "$DIST_DIR/${BINARY_NAME}_windows_amd64.exe" .
}

# Show version
show_version() {
    echo "$VERSION"
}

# Clean build artifacts
clean() {
    echo "Cleaning build artifacts..."
    rm -rf "$DIST_DIR"
    echo "Done."
}

# Main command dispatcher
case "${1:-build}" in
    build)
        build_current
        ;;
    build-all)
        build_all
        ;;
    linux)
        build_linux
        ;;
    darwin-intel)
        build_darwin_intel
        ;;
    darwin-arm)
        build_darwin_arm
        ;;
    windows)
        build_windows
        ;;
    version)
        show_version
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
