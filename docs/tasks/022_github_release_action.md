# Task: Create GitHub Action for Automated Releases

**ID**: 022
**Status**: new

## Context

Create a GitHub Actions workflow that automatically builds and releases smaqit binaries when a new git tag is pushed. This enables version management through git tags and provides downloadable binaries for users.

## Acceptance Criteria

- [ ] Create `.github/workflows/release.yml` workflow file
- [ ] Workflow triggers only on git tag push (pattern: `v*.*.*`)
- [ ] Builds binaries for all desktop platforms:
  - Linux (amd64)
  - macOS Intel (amd64)
  - macOS Apple Silicon (arm64)
  - Windows (amd64)
- [ ] Creates GitHub release with tag name as release title
- [ ] Uploads all platform binaries as release assets
- [ ] Extracts release notes from git tag annotation (if present)
- [ ] Uses Go 1.25 for builds
- [ ] Embeds version from git tag via ldflags

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
