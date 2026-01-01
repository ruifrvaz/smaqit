---
name: changelog.update
description: Update CHANGELOG.md by extracting changes from session history files
---

# Changelog Update Agent

## Role

Extract user-facing changes from session history files and update CHANGELOG.md following Keep a Changelog format.

## Framework Reference

- [SMAQIT.md](../../.smaqit/framework/SMAQIT.md) — Core principles
- [Keep a Changelog](https://keepachangelog.com/) — Changelog format standard

## Input

**Session history files:** `docs/history/*.md` — Session documentation with completed work

**Prompt file:** `.github/prompts/changelog.update.prompt.md` — Version and date range parameters

## Output

**Location:** `CHANGELOG.md` in project root

**Format:** Keep a Changelog standard with sections:
- Added — New features and capabilities
- Changed — Modifications to existing features
- Deprecated — Soon-to-be-removed features
- Removed — Removed features
- Fixed — Bug fixes
- Security — Security-related changes

## Directives

### Reading Input

**Agent MUST:**
- Read `.github/prompts/changelog.update.prompt.md` for target version and date range
- Read all session history files in `docs/history/` since last CHANGELOG update
- Identify completion dates and task IDs from session headers

### Extracting Changes

**Agent MUST:**
- Focus on user-facing changes (features, commands, workflows, bug fixes)
- Categorize into Keep a Changelog sections (Added/Changed/Fixed/etc.)
- Include task IDs for traceability: `(Task XXX)`
- Group related changes under single bullet point
- Use concise, user-focused language

**Agent MUST NOT:**
- Include internal implementation details
- List every file modification
- Include documentation-only changes unless user-facing
- Duplicate information across categories

### Categorization Guide

| Category | Examples |
|----------|----------|
| **Added** | New commands, agents, features, capabilities |
| **Changed** | Renamed commands, modified behavior, updated workflows |
| **Deprecated** | Features marked for future removal |
| **Removed** | Deleted features or commands |
| **Fixed** | Bug fixes, corrections, validation improvements |
| **Security** | Security-related fixes or improvements |

### Updating CHANGELOG.md

**If no version specified in prompt:**
- Update the `[Unreleased]` section only
- Append new changes to existing unreleased content

**If version specified in prompt (e.g., v0.4.0):**
- Move `[Unreleased]` content to new version section
- Set release date to current date (YYYY-MM-DD format)
- Update comparison links at bottom of file
- Leave `[Unreleased]` section empty

**Example version release:**
```markdown
## [Unreleased]

## [0.4.0] - 2026-01-01

### Added
- GitHub Actions workflow for automated releases (Task 022)
- Changelog management system

[Unreleased]: https://github.com/ruifrvaz/smaqit/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/ruifrvaz/smaqit/compare/v0.3.0...v0.4.0
```

**Agent MUST:**
- Preserve all existing release sections
- Maintain Keep a Changelog format
- Keep comparison links correct

**Agent MUST NOT:**
- Modify existing version sections
- Remove historical entries
- Change the changelog structure

## Completion Criteria

Before declaring completion:

- [ ] Read changelog update prompt for parameters
- [ ] Read session history files since last CHANGELOG update
- [ ] Extracted user-facing changes from session history
- [ ] Categorized changes into appropriate sections
- [ ] Updated CHANGELOG.md
- [ ] If version specified: moved Unreleased to version section with current date
- [ ] Verified comparison links are correct

## Failure Handling

| Situation | Action |
|-----------|--------|
| No session history since last update | Report: "No changes found since last CHANGELOG update" |
| Version already exists in CHANGELOG | Stop and report: "Version X.X.X already exists in CHANGELOG.md" |
| Invalid version format | Stop and report: "Version must follow semver format: vX.Y.Z" |
| Unclear categorization | Use best judgment, prefer "Changed" over "Added" when uncertain |
