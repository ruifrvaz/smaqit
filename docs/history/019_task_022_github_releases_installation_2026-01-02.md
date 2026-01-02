# Session 019: Task 022 - GitHub Releases and Installation Workflow

**Date:** January 2, 2026  
**Focus:** Automated releases, changelog management, one-liner installation

## Session Overview

Completed Task 022 (GitHub Actions for automated releases) and extended it with a comprehensive installation workflow including one-liner script, version selection, and standard CLI flags. The session evolved from basic release automation to a complete end-to-end release and distribution system.

## Work Completed

### 1. GitHub Actions Release Workflow

**Objective:** Automate multi-platform builds and release creation

**Implementation:**
- Created `.github/workflows/release.yml` with triggers on `v*.*.*` tags and manual dispatch
- Builds for 4 platforms: Linux amd64, macOS amd64/arm64, Windows amd64
- Go 1.25 toolchain with cross-compilation via `GOOS`/`GOARCH`
- SHA256 checksums generation for all binaries
- Version embedding via git describe in build process

**Key Decision:** Use existing Makefile build system (reuse `prepare` and `build-all` targets) rather than implementing goreleaser or custom build scripts. Rationale: Keep build logic centralized, avoid duplication, leverage existing working system.

### 2. Changelog Management System

**Objective:** Provide informative release notes instead of generic messages

**Problem:** Initially used git tag annotations for release notes, which weren't informative enough for users.

**Solution:** AI-managed changelog system following Keep a Changelog format
- Initialized `CHANGELOG.md` with historical versions (0.0.1 through 0.4.0-beta)
- Created `/changelog.update` agent that reads session history files
- Created `.github/prompts/changelog.update.prompt.md` for version/date range parameters
- Release workflow extracts notes via sed from CHANGELOG.md with fallback chain:
  1. Version section (e.g., `## [0.4.1-beta]`)
  2. Unreleased section
  3. Tag annotation
  4. Default generic message

**Rationale:** Session history files already document all work. Agent transforms these into user-facing changelog entries, eliminating manual changelog maintenance.

### 3. One-Liner Installation Script

**Objective:** "Express way of getting smaqit up and running on new projects"

**Implementation:** `install.sh` with:
- Platform detection via `uname -s` / `uname -m` (Linux/Darwin/Windows, amd64/arm64)
- GitHub Releases API integration for version fetching
- `SMAQIT_VERSION` environment variable with three modes:
  - `latest` - stable releases only (falls back to prerelease if none exist)
  - `prerelease` - any release including pre-releases
  - `vX.Y.Z` - specific version
- Installation to `~/.local/bin` with automatic directory creation
- Verification via `--version` flag
- PATH checking with instructions if not configured
- Color-coded output (info/success/error) via ANSI escape codes

**Critical Bug Fixed:** Initial implementation contaminated stdout with info() output, causing `temp_file=$(download_binary)` to capture color codes in the filename. Solution: Changed from function return pattern to global `TEMP_FILE` variable, separating display output from data passing.

### 4. Standard CLI Flags

**Objective:** Support universal `--version` and `--help` flags

**Problem:** `smaqit` only recognized subcommands (`version`, `help`), not standard flags (`--version`, `-v`).

**Discussion:** Questioned whether all commands should have flag variants. Agreed on hybrid approach:
- Core subcommands: `init`, `status`, `validate`, `uninstall` (domain-specific actions)
- Universal flags: `--version`, `-v`, `--help`, `-h` (info queries)
- Rationale: Universal flags are standard across all CLI tools (go, python, git), improve UX for new users

**Implementation:** Updated `installer/main.go` switch case to accept multiple patterns:
```go
case "version", "--version", "-v":
case "help", "--help", "-h":
```

### 5. Repository Visibility

**Problem:** `install.sh` returned 404 when fetching from GitHub (repository was private).

**Solution:** Made repository public to enable `raw.githubusercontent.com` access and release downloads without authentication.

## Files Modified

### Created:
- `.github/workflows/release.yml` - Release automation workflow
- `CHANGELOG.md` - Project changelog (Keep a Changelog format)
- `.github/agents/changelog.update.agent.md` - Changelog update agent
- `.github/prompts/changelog.update.prompt.md` - Changelog update prompt
- `install.sh` - One-liner installation script

### Modified:
- `installer/main.go` - Added `--version`/`-v` and `--help`/`-h` flag support
- `README.md` - Added installation instructions and release process documentation
- `docs/tasks/022_github_release_action.md` - Task completion documentation

## Key Decisions

### 1. Makefile vs Goreleaser

**Decision:** Use existing Makefile

**Alternatives Considered:**
- Goreleaser: Popular Go release tool with built-in GitHub integration
- Custom build scripts: Full control but requires maintenance

**Rationale:** 
- Makefile already works and is tested
- Avoids introducing new dependencies
- Centralizes build logic in one place
- Team already familiar with make commands

### 2. Changelog Automation Approach

**Decision:** AI agent reads session history files

**Alternatives Considered:**
- Conventional commits parsing (e.g., semantic-release)
- Manual changelog maintenance
- Git log formatting

**Rationale:**
- Session history files already document all work with context
- AI can extract user-facing changes and categorize them
- No additional commit message conventions needed
- Captures rationale and decisions, not just code changes

### 3. Installation Method

**Decision:** Shell script with curl piping

**Alternatives Considered:**
- Package managers (apt, brew, chocolatey) - too heavy for small CLI tool
- Go install (requires Go on user machine)
- Binary downloads with manual PATH setup (poor UX)

**Rationale:**
- Single command for users: `curl ... | bash`
- No prerequisites except curl
- Standard pattern used by rustup, nvm, etc.
- Supports version selection via environment variable

### 4. Version Selection Design

**Decision:** Environment variable with three modes (latest/prerelease/vX.Y.Z)

**Alternatives Considered:**
- CLI flags: `install.sh --version=X.Y.Z` (requires saving script first)
- Separate scripts: `install-latest.sh`, `install-prerelease.sh` (duplication)
- Always install latest regardless of stability (poor for production)

**Rationale:**
- Environment variable works with curl piping: `SMAQIT_VERSION=X bash`
- Default behavior (latest stable) is safe for production
- Explicit prerelease opt-in for testing
- Specific version for reproducibility

## Testing and Validation

### Release Workflow Testing:
- Created v0.4.0-beta release successfully
- Verified all 4 platform binaries built and uploaded
- Confirmed SHA256 checksums generated
- Validated release notes extracted from CHANGELOG.md

### Installation Script Testing:
- Tested platform detection on Linux amd64
- Verified version selection with prerelease flag
- Fixed stdout contamination bug (global variable pattern)
- Confirmed `--version` flag works after CLI update
- Validated v0.4.1-beta installation end-to-end

### Go Installation:
- User machine didn't have Go installed
- Recommended `sudo snap install go --classic` over manual tarball approach
- Addressed classic confinement security warning (expected for dev tools)

## Release Created

**v0.4.1-beta** (January 2, 2026)

**Added:**
- One-liner installation script with platform detection
- Version selection via SMAQIT_VERSION environment variable
- Standard CLI flag support (--version, -v, --help, -h)

**Fixed:**
- Install script stdout contamination from info messages
- Repository visibility for public access

## Next Steps

1. **Monitor Release Adoption:** Track downloads/issues from v0.4.1-beta users
2. **Consider Stable Release:** Move to v0.4.1 (non-beta) once installation workflow proven stable
3. **Package Managers:** Evaluate adding Homebrew tap or other package manager support if demand exists
4. **Installation Documentation:** Add troubleshooting section to README based on user feedback

## Session Metrics

**Duration:** ~3 hours (including Go installation)  
**Tasks Completed:** 1 (Task 022 + extended scope)  
**Files Created:** 5 (workflow, changelog, agents, prompts, install script)  
**Files Modified:** 3 (main.go, README.md, task file)  
**Releases Created:** 2 (v0.4.0-beta, v0.4.1-beta)  
**Key Outcomes:**
- Fully automated release workflow operational
- One-line installation command working
- Standard CLI conventions adopted
- Repository now public for easy distribution

## Technical Notes

### GitHub Actions Workflow Pattern:
```yaml
on:
  push:
    tags: ['v*.*.*']
  workflow_dispatch:
    inputs:
      version:
        description: 'Version tag'
```

### Install Command Pattern:
```bash
# Latest stable (with prerelease fallback)
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | bash

# Latest prerelease
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | SMAQIT_VERSION=prerelease bash

# Specific version
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | SMAQIT_VERSION=v0.4.1-beta bash
```

### Changelog Update Workflow:
1. Fill in `.github/prompts/changelog.update.prompt.md` with target version
2. Run changelog update agent (reads session history)
3. Agent updates CHANGELOG.md Unreleased section or creates versioned section
4. Commit changelog
5. Create and push git tag
6. GitHub Actions builds and releases with notes from CHANGELOG.md

## Lessons Learned

1. **Check Prerequisites Early:** Should have verified Go installation at session start, not after needing to build
2. **Start Simple:** Suggesting snap first would have been faster than manual tarball installation
3. **Test Return Values:** Shell function return via echo is fragile when other echo statements exist; use variables instead
4. **CDN Caching:** GitHub's raw content CDN can cache; cache-busting with query params (`?$(date +%s)`) bypasses
5. **Standard Conventions Matter:** Users expect `--version` to work; implementing standard patterns improves adoption
