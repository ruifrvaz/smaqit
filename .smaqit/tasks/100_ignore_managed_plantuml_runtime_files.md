# Ignore Managed PlantUML Runtime Files

**Status:** Completed
**Mode:** Assisted
**Created:** 2026-08-06
**Started:** 2026-08-06
**Completed:** 2026-08-06

## Description

Ensure `smaqit init` and `smaqit update` prevent the embedded PlantUML runtime under `.smaqit/tools/` from appearing as thousands of untracked project files. The runtime is a generated, versioned installation artifact containing pinned MCP, Resvg/WASM, font, and Node dependencies; it must remain local to each consumer project and never be committed.

The forward fix must add and preserve a narrow Git ignore rule for the managed tools directory without ignoring canonical design artifacts. `docs/designs/**/*.md` and `docs/designs/**/*.png` remain project-owned and versioned.

## Design Decisions

- **Ignore boundary:** Add `.smaqit/tools/` to the consumer project's root `.gitignore`; do not ignore `.smaqit/` broadly.
- **Preservation:** Merge the managed rule idempotently into an existing `.gitignore` without overwriting unrelated user rules.
- **Forward-only:** Do not delete, migrate, or automatically untrack runtime files already added to consumer Git indexes.

## Implementation Steps

1. Locate the initialization/update path that materializes the bundled design runtime.
2. Add an idempotent managed `.gitignore` update for `.smaqit/tools/`, preserving existing file content and line endings as appropriate.
3. Cover fresh initialization, reinitialization with an existing `.gitignore`, and update behavior in installer tests and smoke coverage.
4. Confirm Git ignores the runtime while still tracking canonical `docs/designs/` Markdown and PNG artifacts.
5. Document the managed-runtime ignore behavior for consumers.

## Known Issues Triage

**Triaged:** 2026-08-06
**Tools searched:** PlantUML
**Result:** Clear

### Blocking Issues
- None.

### Advisory Issues
- None.

### Historical (Closed)
- None relevant to managed Git-ignore behavior.

### Unresolvable Tools
- None.

## Acceptance Criteria

- [x] A fresh `smaqit init` results in `.smaqit/tools/` being ignored by Git.
- [x] `smaqit update` adds the managed ignore rule to an existing project without overwriting unrelated `.gitignore` content.
- [x] Repeated initialization or update does not duplicate the ignore rule.
- [x] The ignore rule does not exclude `docs/designs/**/*.md` or `docs/designs/**/*.png`.
- [x] Installer unit and smoke tests cover the managed runtime ignore behavior.
- [x] Existing projects with already-tracked runtime files are not modified or untracked automatically.

## Findings

**Implementation approach:**
- Added an idempotent installer helper that appends only `.smaqit/tools/` after successful runtime materialization.
- Covered helper behavior in Go tests and consumer Git behavior in the installer smoke test.

**Decisions made:**
- Preserve existing ignore-file content and its LF or CRLF convention.
- Leave already-tracked runtime files unchanged; the rule affects only future untracked files.

**Blockers encountered:**
- The initial task registration had to be committed before creating an isolated worktree so the branch inherited its task file.
- No upstream PlantUML issue affected this installer-owned change.

**Follow-up identified:**
- Include the forward fix in the next patch release.

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/` initialization/update implementation | Modify |
| `installer/*_test.go` | Modify |
| `scripts/smoke-test-installer.sh` | Modify |
| Consumer installation documentation | Modify |

## Notes

v2.0.1 materializes approximately 3,875 files (about 67 MB) under `.smaqit/tools/plantuml/` for the pinned offline PlantUML toolchain. This is expected runtime content, but it should not appear as untracked consumer source files.
