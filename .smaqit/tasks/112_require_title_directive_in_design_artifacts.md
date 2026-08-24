---
status: Completed
created: "2026-08-19"
mode: Assisted
started: "2026-08-20"
completed: "2026-08-24"
---

# Require Identifying Title Directive in Design Artifacts

## Description

Design PlantUML artifacts under `docs/designs/<layer>/` have no on-diagram identifier today, making it hard to tell which spec/concept a rendered PNG belongs to at a glance. Worse, `framework/ARTIFACTS.md` currently *forbids* a title outright ("MUST NOT contain a title, prose, table, reference section, second block, HTML, or embedded image"), and the Go validator (`installer/design.go`) actively strips any `title` directive as decorative before its structural scan — there is no enforcement of title presence anywhere.

This task reverses that prohibition and requires every design template to carry a `title` directive whose value is the design's own `id` frontmatter value (e.g. `title DSG-BUS-LOGIN-USE-CASE`). The design `id` was chosen deliberately over other candidate identifiers considered and rejected during planning:

- The Business spec's `UC[N]-[CONCEPT]` heading label (e.g. `UC1-LOGIN`) is purely cosmetic — it is never wired into frontmatter, requirement IDs, design IDs, or cross-layer references anywhere in the framework.
- A "linked spec identifier" prefix breaks down because spec-to-design cardinality is not guaranteed 1:1: `ARTIFACTS.md` explicitly allows one design to serve several related specs, and a Functional "Foundation" spec can `Enable` (serve) multiple Business use cases.

The design's own `id` is already unique, stable, and threaded through its frontmatter, filename, and requirement IDs — using it as the title requires no new cross-referencing logic and has a single unambiguous answer regardless of cardinality.

## Issue Triage Context

**Mode:** Skip
**Technologies:** None
**Platforms/Environments:** None
**Features/Integrations:** None
**Versions/Constraints:** None

## Design Decisions

- **Title content:** the design's own `id` frontmatter value, verbatim — not the Business `UC[N]` heading label (unwired elsewhere in the framework), not an attempted spec-identifier prefix (breaks under non-1:1 spec-to-design cardinality).
- **Scope:** applies uniformly across all 5 spec layers (business, functional, stack, infrastructure, coverage) plus `design-sequence`, not just `system-sequence` — the underlying problem (no on-diagram identifier) applies everywhere.
- **Enforcement strength:** content-match (title must exactly equal `id`), not mere presence, to catch drift if a design is ever renamed without updating its title.
- **Rule reversal is explicit and intentional**, not a silent exception — `framework/ARTIFACTS.md`'s prohibition is being replaced with a requirement, documented as such.

## Implementation Steps

1. Reverse the "MUST NOT contain a title" prohibition in `framework/ARTIFACTS.md` (~line 379) to require a `title` directive instead; update the adjoining Validation Gates / one-block-no-prose description (~line 409) accordingly.
2. Add `title <DESIGN_ID pattern>` as the first line inside `@startuml` in each of the 6 `templates/designs/*.template.md` files (business, functional, stack, infrastructure, coverage, design-sequence), matching each layer's own `id:` frontmatter pattern (e.g. `title DSG-BUS-[CONCEPT]-USE-CASE`).
3. In `installer/design.go`, add a title-extraction helper near `footboxHidden` (design.go:504) that scans the raw source — *before* `stripNonStructuralPlantUML` (design.go:448-495) discards it — for a `title <value>` line.
4. Wire the new check into `validateDesignMetadata` (design.go:321), alongside the existing `footboxHidden` invocation (~design.go:345-347): fail with `DESIGN-VISUAL-INVALID` if the directive is absent, and fail if present but its value doesn't exactly match the design's own `id` frontmatter field.
5. Update `skills/smaqit.design-validate/SKILL.md` step 4 wording so the now-required title directive is not flagged as "ceremonial content" during visual review.
6. Add new table-driven test cases to `installer/design_test.go` (missing title, title/id mismatch) alongside the existing `missingFootbox`/`footer` fixtures (~design_test.go:433-480), and update every existing passing fixture in the file to include a correct matching title line so none regress.
7. Run `go test ./...` and `go build ./...` in `installer/` to confirm.

## Known Issues Triage

**Triaged:** 2026-08-20
**Tools searched:** none
**Result:** Clear — Issue Triage Context Mode is Skip; this task is an internal change to smaqit's own templates, framework docs, and Go validator (`installer/design.go`), with no third-party dependency involved. Triage not applicable.

## Acceptance Criteria

- [x] `framework/ARTIFACTS.md`'s Design Artifacts section requires (not prohibits) a `title` directive matching the design's `id` frontmatter value
- [x] All 6 `templates/designs/*.template.md` files include a `title <DESIGN_ID pattern>` line as the first line inside `@startuml`
- [x] `installer/design.go` enforces both presence of a `title` directive and that its value exactly matches the design's own `id` field, across all diagram types, failing `DESIGN-VISUAL-INVALID` on violation
- [x] `skills/smaqit.design-validate/SKILL.md`'s guidance no longer flags the required title as ceremonial content
- [x] New `installer/design_test.go` cases cover missing-title and title/id-mismatch failures; all existing passing fixtures updated to include a matching title
- [x] `go test ./...` passes in `installer/`

## Findings

**Implementation approach:**
- Reversed the title prohibition in `framework/ARTIFACTS.md` (line 379, plus the Validation Gates summary at line 409) to a requirement, and added a matching `title` line to all 6 `templates/designs/*.template.md` files as the first line inside `@startuml`.
- Added `titleDirective()` in `installer/design.go`, mirroring `footboxHidden`'s note/legend-aware raw-source scan, then wired presence + exact id-match enforcement into `validateDesignMetadata` right after the layer/profile checks, applying uniformly across all diagram types (not scoped to `system-sequence` only).
- Updated `smaqit.design-validate/SKILL.md` step 4 so the required title is not flagged as ceremonial content during visual review.
- Added `TestDesignRequiresMatchingTitleDirective` covering missing-title and title/id-mismatch rejections, and updated every existing fixture across `installer/design_test.go` (both inline `@startuml` literals and the shared fixture-generator helpers) to carry a matching title so none regressed.

**Decisions made:**
- Title content is the design's own `id` frontmatter value verbatim — not the Business `UC[N]` heading label (never wired anywhere else in the framework) and not an attempted spec-identifier prefix (breaks under non-1:1 spec-to-design cardinality); this matches the plan's Design Decisions exactly.
- `titleDirective()` scans the raw source directly (before `stripNonStructuralPlantUML` discards title lines) and only supports the single-line `title <value>` form; the multi-line `title`/`end title` block form is deliberately not read as content.

**Blockers encountered:**
- `installer/`'s `go:embed` targets (agents/skills/tools) are gitignored generated artifacts not present in a fresh worktree; copied them from the primary checkout's already-`make prepare`d installer directory instead of re-running the full (network-dependent) prepare pipeline, then regenerated `installer/templates/{specs,designs}` from the worktree's own edited canonical templates so the embedded copies reflected this task's changes.
- The `TestSystemSequenceProfileSupportsMultipleDesignsPerSpec` test built its two design variants via `strings.Replace(source, "DSG-FUN-TEST-FLOW-SYSTEM-SEQUENCE", "...-REGISTRATION-...", 1)` — replacing only the first (frontmatter `id`) occurrence. Adding a title line matching the same base id introduced a second occurrence that needed swapping too, so the replace count changed from `1` to `-1` (replace all) for both call sites.
- One fixture generator (`validDesignSource`) is shared by a test that intentionally never reaches YAML parsing (`TestParseDesignRejectsProseAndUnsafeIncludes`, which short-circuits on earlier prose/include checks) using YAML-invalid bracket placeholder hashes (`[SOURCE_SHA256]`); the new title-directive test does reach YAML parsing, which surfaced that those placeholders parse as YAML flow-sequences, not strings. Standardized all `validDesignSource` call sites in the file on the already-established quoted-empty-string (`""`) convention used elsewhere.

**Follow-up identified:**
- None — no real design artifacts exist yet anywhere in this repo (`docs/designs/` is empty), so there is no migration burden left for a future task.

## Files to Create / Modify

| File | Action |
|------|--------|
| `framework/ARTIFACTS.md` | Modify — reverse title prohibition to a requirement |
| `templates/designs/business.template.md` | Modify — add `title` line |
| `templates/designs/functional.template.md` | Modify — add `title` line |
| `templates/designs/stack.template.md` | Modify — add `title` line |
| `templates/designs/infrastructure.template.md` | Modify — add `title` line |
| `templates/designs/coverage.template.md` | Modify — add `title` line |
| `templates/designs/design-sequence.template.md` | Modify — add `title` line |
| `installer/design.go` | Modify — title-extraction helper + validation wiring |
| `installer/design_test.go` | Modify — new fixtures/table rows, update existing fixtures |
| `skills/smaqit.design-validate/SKILL.md` | Modify — step 4 wording |

## Notes

Planned via `smaqit.task-plan` (2026-08-19) using 3 parallel Explore subagents covering template conventions, validator mechanics, and use-case/spec ID conventions. No real design artifacts exist yet anywhere in this repo (`docs/designs/` is empty), so this task has no migration burden — only templates, framework docs, the validator, and its tests are affected.
