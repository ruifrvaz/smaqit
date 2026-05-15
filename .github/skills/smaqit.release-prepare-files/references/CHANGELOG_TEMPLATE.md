# CHANGELOG Section Template

Use this template for every version section when preparing a release.

## Section Structure

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- (new features, new files, new behaviors)

### Changed
- (changes to existing functionality)

### Deprecated
- (features soon to be removed)

### Removed
- (features, files, or behaviors deleted in this release)

### Fixed
- (bug fixes, corrections, stale reference cleanup)

### Security
- (vulnerability patches)

### Chore
- (internal housekeeping: CI, tooling, config, non-functional changes)
```

## Rules

- **All 7 sections MUST be present** in every version entry.
- **If a section has no entries**, write `- Nothing to add.` — do NOT omit the section header.
- **One item per line**, starting with `-`.
- **Reference PRs or sessions** at the end of each item in parentheses: `(PR #42)` or `(Session 054)`.
- **Do NOT use sub-bullets** unless grouping closely related items under a parent entry.

## Unreleased Section

The `[Unreleased]` section at the top follows the same 7-section structure. Content accumulates here between releases and is moved to the new version section at release time.

## Scratch (empty) Unreleased template

```markdown
## [Unreleased]

### Added
- Nothing to add.

### Changed
- Nothing to add.

### Deprecated
- Nothing to add.

### Removed
- Nothing to add.

### Fixed
- Nothing to add.

### Security
- Nothing to add.

### Chore
- Nothing to add.
```
