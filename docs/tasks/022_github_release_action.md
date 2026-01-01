# Task: Create GitHub Action for Automated Releases

**ID**: 022
**Status**: Completed (2026-01-01)

## Context

Create a GitHub Actions workflow that automatically builds and releases smaqit binaries when a new git tag is pushed. This enables version management through git tags and provides downloadable binaries for users.

## Acceptance Criteria

- [x] Create `.github/workflows/release.yml` workflow file
- [x] Workflow triggers only on git tag push (pattern: `v*.*.*`)
- [x] Builds binaries for all desktop platforms:
  - Linux (amd64)
  - macOS Intel (amd64)
  - macOS Apple Silicon (arm64)
  - Windows (amd64)
- [x] Creates GitHub release with tag name as release title
- [x] Uploads all platform binaries as release assets
- [x] Extracts release notes from git tag annotation (if present)
- [x] Uses Go 1.25 for builds
- [x] Embeds version from git tag via ldflags

## Completion Summary

Created `.github/workflows/release.yml` that:
- Triggers on `v*.*.*` tag push
- Uses existing Makefile `build-all` target to build all platforms
- Generates SHA256 checksums for all binaries
- Extracts release notes from annotated tag message
- Creates GitHub release using `softprops/action-gh-release@v2`
- Marks versions with `-` (e.g., `v0.3.0-beta`) as prereleases

**Files created:**
- `.github/workflows/release.yml`

## Notes

- Workflow should use `actions/checkout@v4` for repo checkout
- Use `actions/setup-go@v5` for Go installation
- Consider `softprops/action-gh-release@v1` or `ncipollo/release-action@v1` for release creation
- Binary naming: `smaqit_{os}_{arch}[.exe]` (consistent with Makefile)
- Checksums file (SHA256) should be generated for all binaries
- Workflow should fail if any platform build fails

## Example Release Flow

```bash
# Developer creates and pushes tag
git tag -a v0.2.0 -m "Release v0.2.0: Add installer build system"
git push origin v0.2.0

# GitHub Actions automatically:
# 1. Detects tag push
# 2. Builds all platform binaries
# 3. Creates GitHub release
# 4. Uploads binaries as assets
```
