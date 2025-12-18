# Task: Create Cross-Platform Go Installer Build System

**ID**: 012
**Status**: new

## Context

Create a complete Go application build system for the smaqit installer. The installer CLI (`installer/main.go`) needs proper build scripts to compile into executables for multiple target platforms (Windows, macOS, Linux) with different architectures.

## Acceptance Criteria

- [x] Create `Makefile` with targets for building all platforms
- [x] Create `build.sh` (bash script) for Unix-like systems
- [x] Create `build.bat` (batch script) for Windows systems
- [x] Support target platforms:
  - Linux (amd64)
  - macOS (amd64, arm64)
  - Windows (amd64)
- [x] Output binaries to `dist/` folder with platform-specific naming
- [x] Include `clean` target to remove build artifacts
- [x] Include `install` target for local installation
- [x] Document build instructions in README or separate BUILD.md

## Notes

- Go supports cross-compilation via `GOOS` and `GOARCH` environment variables
- Consider using `go build -ldflags` to embed version info
- Installer version should sync with SMAQIT.md version (per copilot-instructions)
- Binary naming convention: `smaqit-{os}-{arch}[.exe]`
