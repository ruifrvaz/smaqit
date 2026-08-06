# PlantUML Runtime Reliability

**Date:** 2026-08-06
**Session Focus:** Make PlantUML visual artifacts dependable in consumer projects, then verify the published releases through fresh-install and update workflows.
**Tasks Completed:** 099, 100
**Releases Published:** v2.0.1, v2.0.2

---

## Actions Taken

### Visual Artifact Contract

- Completed the opaque PlantUML PNG work: rendered diagrams now have an opaque cream (`#FFF9F0`) canvas rather than a transparent or stark-white background. The renderer version marker, documentation, test cases, smoke coverage, and installer regressions were updated together.
- Validated a freshly installed consumer project visually: generated PlantUML diagrams rendered legibly with the new canvas and passed the design validation and attestation flow.
- Clarified the implementation boundary established earlier in the session: specification agents own design creation and visual attestation; implementation agents consume the specification Markdown and linked PlantUML source, not PNGs as implementation requirements.

### Task Lifecycle Reliability

- Corrected task-lifecycle resolution to recognize legacy parent task identifiers such as `B001`, preventing an unrelated historical parent format from blocking completion.
- Made task completion retry-safe after a task has already been marked completed but its first merge attempt was interrupted by unrelated dirty-worktree state.
- Completed and merged task 099 after verifying its acceptance criteria.

### Managed PlantUML Runtime

- Investigated the thousands of files created under `.smaqit/tools/` during installation and update. They are the deliberately vendored PlantUML MCP/runtime dependencies required for offline, reproducible rendering.
- Completed task 100 so `smaqit init` and `smaqit update` add exactly the narrow `.smaqit/tools/` rule to a consumer project's root `.gitignore`. Existing user rules and line endings are preserved; the installer does not untrack or alter already tracked files.
- Added unit and smoke coverage for fresh Git repositories, existing user `.gitignore` content, duplicate prevention, protected design artifacts, and an already tracked runtime file.
- Updated the README and visual-design wiki to explain that the managed runtime is generated and Git-ignored while PlantUML sources and PNG artifacts under `docs/designs/` remain versioned design deliverables.

### Releases and Consumer Verification

- Cleared the stale local `release/v2.0.0` worktree after confirming its PR and tag had already shipped.
- Prepared and published v2.0.1 for the opaque-rendering and lifecycle fixes, then verified its published Linux asset in an isolated consumer project.
- Prepared and published v2.0.2 for the managed-runtime Git-ignore behavior, then dry-ran the exact published binary in two isolated Git repositories: a fresh initialization and a real v2.0.1-to-v2.0.2 update. Both left the PlantUML runtime ignored, retained normal design paths as trackable, and preserved user-owned `.gitignore` content.

## Problems Solved

- **Transparent visual artifacts:** PlantUML raster output now has a predictable, readable background for visual QA across themes and viewers.
- **Lifecycle completion blocked by historical metadata:** old `B`-prefixed parent identifiers no longer break current task completion, and a safe retry is possible after a previously recorded completion.
- **Generated runtime polluting consumer Git status:** the approximately 3,875-file managed runtime is now ignored automatically without ignoring user-authored design sources or images.

## Decisions Made

- **PlantUML remains the canonical design source, with PNG as visual-QA projection.** This keeps diagrams structured and directly readable by agents while still enabling a rendered visual gate.
- **Use one strict install path, not fallbacks.** The embedded runtime is shipped with smaqit; consumers receive a clear failure when its prerequisites cannot run.
- **Ignore only `.smaqit/tools/`.** The rule is intentionally narrow so canonical `docs/designs/` artifacts are always available for source control and review.
- **Release worktrees are cleaned explicitly after release verification.** Publication does not delete the local release workspace automatically.

## Files Modified

- `tools/plantuml/render-png.mjs` — opaque cream PlantUML canvas rendering.
- `installer/design_tools.go`, `installer/design_test.go`, `scripts/prepare-design-tools.mjs` — renderer bundle revision and visual-artifact regression coverage.
- `installer/main.go`, `installer/main_test.go` — idempotent managed-runtime Git-ignore installation and tests.
- `scripts/smoke-test-installer.sh` — Git-ignore and tracked-runtime consumer smoke scenarios.
- `.agents/skills/smaqit.utils.worktree/scripts/9_resolve_task_lifecycle.sh` — historical parent-ID support and completion retry behavior.
- `README.md`, `docs/wiki/workflows/visual-designs.md`, `framework/ARTIFACTS.md`, `docs/test-cases/visual-design-artifacts.md` — PlantUML runtime and artifact-contract documentation.
- `CHANGELOG.md`, release metadata, `.smaqit/tasks/099_opaque_plantuml_png_rendering.md`, `.smaqit/tasks/100_ignore_managed_plantuml_runtime_files.md`, and `.smaqit/tasks/PLANNING.md` — task and release records.
- `smaqit.code-workspace` — workspace refreshes and released-worktree cleanup.

## Next Steps

- Remove the now-merged local `release/v2.0.2` worktree and refresh the workspace when release-worktree cleanup is next requested.
- Continue using both fresh-install and upgrade scenarios as release gates for changes that affect managed installation content.

## Session Metrics

- **Tasks completed:** 2
- **Patch releases published:** 2
- **Published consumer dry runs:** 3
- **Managed runtime footprint validated:** approximately 3,875 generated files, kept out of Git status by one exact ignore rule
