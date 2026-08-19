---
status: In Progress
created: "2026-08-19"
mode: Assisted
started: "2026-08-20"
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

- [ ] `framework/ARTIFACTS.md`'s Design Artifacts section requires (not prohibits) a `title` directive matching the design's `id` frontmatter value
- [ ] All 6 `templates/designs/*.template.md` files include a `title <DESIGN_ID pattern>` line as the first line inside `@startuml`
- [ ] `installer/design.go` enforces both presence of a `title` directive and that its value exactly matches the design's own `id` field, across all diagram types, failing `DESIGN-VISUAL-INVALID` on violation
- [ ] `skills/smaqit.design-validate/SKILL.md`'s guidance no longer flags the required title as ceremonial content
- [ ] New `installer/design_test.go` cases cover missing-title and title/id-mismatch failures; all existing passing fixtures updated to include a matching title
- [ ] `go test ./...` passes in `installer/`

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

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
