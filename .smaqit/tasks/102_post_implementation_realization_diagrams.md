# Design Sequence Diagrams for the Development Phase

**Status:** Completed
**Mode:** Assisted
**Created:** 2026-08-06
**Started:** 2026-08-07
**Completed:** 2026-08-10

## Description

`system-sequence` diagrams (Functional's default profile) are black-box UML System Sequence Diagrams — one actor, one opaque system participant, external contract only (see task 104, which adds strict validation for exactly this shape). Nothing in the current artifact model captures the next step in the classic analysis-design progression (Larman, *Applying UML and Patterns*): use-case → system sequence diagram (external contract) → **design sequence diagram** (internal realization, showing the actual objects/handlers that collaborate to fulfill the contract). Phase 1's Implementation Artifacts (`framework/PHASES.md` line 11) are just "Code, README, Development report" — no diagram at all.

Found via a downstream project's own task (2026-08-06): a pilot of the new PlantUML design-pair workflow on a brownfield project revealed that 5 of that project's 12 hand-authored Mermaid diagrams (its equivalents of `sequence-order-lifecycle.md`, `sequence-admin-login.md`, `sequence-customer-registration-login.md`, `sequence-reminder-escalation.md`, `sequence-error-recovery.md`) are exactly this artifact type — and that keeping them hand-authored has already required three separate "fold drift back in" tasks to stay accurate against real implementation. For a greenfield project there is no such hand-authored set to fall back on at all — this class of diagram simply doesn't get produced anywhere in the current model.

**Approach:** `smaqit.development` (Phase 1's implementation agent) generates one Design Sequence Diagram per implemented Functional spec, after code/tests pass and before phase completion — a new artifact category, "Design Sequence Diagrams," stored in its own sibling tree (`docs/designs/design-sequence/`, prefix `DSD`) with full CLI parity to the existing Design Artifact machinery (schema, hashing, `smaqit design render/attest/validate`), but owned procedurally by Development rather than a specification agent. Two deterministic checks give it real teeth, wired into `smaqit design attest` itself: **grounding** (every message must cite a real `file:line`, checked to exist) and **completeness** (every operation the paired `system-sequence` SSD promises must appear in the diagram, via a `realizes:` frontmatter link). Value: (1) a validation gate closing a completeness check smaqit currently has no mechanism for; (2) an easier human-approval surface than a raw diff before Deploy begins; (3) living documentation that can't hand-drift, because it's regenerated and re-validated rather than hand-maintained.

## Design Decisions

- **New sibling artifact category, not a Design Artifact sidecar.** "Design Sequence Diagrams" live in their own top-level tree, `docs/designs/design-sequence/` — a sibling to `business/`, `functional/`, `stack/`, `infrastructure/`, `coverage/`, not nested inside any of them. Owned by `smaqit.development`, not any specification agent.
- **Why a sibling tree, not `docs/designs/functional/`:** `docs/designs/<layer>/` is producer-owned by that layer's specification agent (Bounded Agents principle, `framework/SMAQIT.md:65-68`) and its Design Lifecycle rule resets a linked active spec to `draft` on every semantic edit (`framework/ARTIFACTS.md`'s Design Lifecycle section) — appropriate for pre-implementation spec-input designs, actively wrong for a diagram regenerated after implementation. Confirmed this reset is a procedural instruction followed only by spec agents (no Go code anywhere auto-resets spec status) — so a new, separately-documented category simply never inherits it; no exception to an existing rule is needed.
- **Bounded Agents stays untouched.** `agents/development.md`'s MUST NOT ("Author, render, visually review, attest, or repair designs...") gets reworded to scope explicitly to the Design Artifacts category — it keeps meaning exactly what it already says for the five spec layers; Development is authorized only for the new `design-sequence` layer, which was never a Design Artifact.
- **Technical parity, real not superficial.** Same YAML+PlantUML pair structure, same hash/attestation fields, same `smaqit design render/attest/validate` CLI verbs — extending `designProfiles`, `designLayerPrefix` (prefix `DSD`), and `designIDPattern` exactly as every existing layer already does. `smaqit design validate`'s directory walk is already generic, so structural validation applies unmodified.
- **Cardinality:** one design-sequence diagram per implemented Functional spec, pairing 1:1 with that spec's `system-sequence` SSD via a new `realizes:` frontmatter field.
- **Grounding mechanism:** self-citation + deterministic CLI check. Each PlantUML message carries a `' impl: <path>:<line>` line comment, extracted by its own direct regex over the raw source rather than through `stripNonStructuralPlantUML` (below) — that shared filter discards `'`-prefixed lines as comments, which is exactly what a citation line is, so grounding extraction deliberately reads the raw source instead. A new `validateDesignSequenceGrounding` check resolves each citation safely against the project root (reusing `validateDesignReferences`'s existing traversal-safe path resolution) and fails if the file or line doesn't exist. Existence-only — not semantic verification that the cited code does what the diagram claims, mirroring task 104's own accepted "lint-style heuristic, not semantic verifier" limitation.
- **Completeness mechanism:** a new `validateDesignSequenceCompleteness` check extracts the paired SSD's promised operation labels and this diagram's own labels via `extractOperationLabels`, and fails, naming the gap, if any SSD-promised operation has no match here. **Correction after task 104 shipped:** the original plan assumed reuse of task 104's arrow-parsing helpers, but task 104's actual implementation deliberately does *no* arrow/message parsing at all — its commit message states this explicitly, "to avoid the bug surface a fuller arrow-token heuristic would carry." There was nothing to reuse for label extraction itself; `extractOperationLabels`/`plantUMLArrow` remain task 102's own, necessarily independent code. What *did* consolidate once task 104 merged: its inline note/title/legend-stripping state machine (originally private to `validateSystemSequenceProfile`) was extracted into a shared `stripNonStructuralPlantUML` helper, now used by both `validateSystemSequenceProfile` (task 104) and `extractOperationLabels` (task 102) — real DRY, not just documentation cleanup, verified behavior-preserving by task 104's own 11 pre-existing test cases still passing unchanged after the extraction.
- **Attestation is earned, not just ordered.** Both checks run inside `attestDesign`, gated on `f.Layer == "design-sequence"`, *before* stamping `Status: "passed"` — attestation fails outright (reusing `DESIGN-VISUAL-INVALID`) if either check fails, rather than relying on Development calling checks in the right order.
- **Complement, not substitute, for code review.** Especially for security-sensitive or edge-case-heavy code, a diagram is a lossy abstraction. Documentation and agent instructions must say this explicitly rather than imply the diagram replaces review.
- **Sequencing:** builds on task 101 (Reliable Design Toolchain, completed) for the rendering pipeline. Task 104 (Strict Black-Box Validation, completed and merged as of this task's implementation) landed after task 102's initial implementation; its branch was merged into task 102's, the one resulting conflict (both branches added a new conditional check to `validateDesignMetadata` in the same spot) was resolved by keeping both checks sequentially, and the `stripNonStructuralPlantUML` consolidation above was done as a follow-up refactor, confirmed behavior-preserving by the full test suite (task 104's 11 tests + task 102's 4, all passing).

## Implementation Steps

**Phase A — Schema & directory plumbing**
1. In `installer/design.go`: add `"design-sequence"` to `designProfiles`, `designLayerPrefix["design-sequence"] = "DSD"`, and `DSD` to the `designIDPattern` regex alternation (`design.go:25,31-41`).
2. Add a `Realizes string` field (`yaml:"realizes"`) to the design frontmatter struct, pointing to the paired `system-sequence` design's ID.
3. In `installer/main.go`, add `docs/designs/design-sequence` to the `smaqit init` directory-scaffolding list (`main.go:513-517`).
4. New template `templates/designs/design-sequence.template.md` — ID `DSG-DSD-[CONCEPT]-DESIGN-SEQUENCE`, `layer: design-sequence`, `diagram_type: design-sequence`, `realizes: DSG-FUN-[CONCEPT]-SYSTEM-SEQUENCE`, PlantUML skeleton with 2+ internal collaborators and `' impl: <path>:<line>` citations.

**Phase B — Grounding & completeness checks** (depends on Phase A; completeness also depends on task 104 landing)
5. Add `validateDesignSequenceGrounding(d *designArtifact, root string) error`: extract `' impl: <path>:<line>` comments, resolve safely against `root` (reuse `validateDesignReferences`'s traversal-safe resolution, `design.go:396-452`), fail on missing file or out-of-range line.
6. Add `validateDesignSequenceCompleteness(d *designArtifact, root string) error`: resolve `d.Front.Realizes`, extract the paired SSD's promised operation labels by extending task 104's arrow-parsing helpers, extract this diagram's own labels, fail naming any unmatched SSD operation.
7. Wire both into `attestDesign` (`design.go:722-746`), gated on `f.Layer == "design-sequence"`, before stamping `Status: "passed"`. Reuse `DESIGN-VISUAL-INVALID` on failure.

**Phase C — Framework & agent documentation**
8. `framework/ARTIFACTS.md`: new `## Design Sequence Diagrams` section (ownership, storage, grounding/completeness contract, own Lifecycle paragraph with no reset-to-draft clause), and add to the Develop-phase Implementation Artifacts bullet list (`:492-495`).
9. `framework/PHASES.md`: Phase 1 Overview row (`:9-11`), Phase Completion paragraph (`:79-81`), and Develop→Deploy prerequisites (`:217-229`) each gain a new criterion.
10. `agents/development.md`: Output section (`:34`), a new MUST directive, Phase-Specific Rules gains a new step 8 after "Verify" (`:257-267`), Completion Criteria gains a new checkbox (`:284-300`), and the existing MUST NOT design-authoring line (`:69-70`) is reworded to scope explicitly to Design Artifacts.
11. Confirm `.smaqit/definitions/agents/development.frontmatter.yaml` tool grants already cover this (likely no change needed).

**Phase D — Tests**
12. `installer/design_test.go`: tests mirroring `TestParseDesignRejectsProseAndUnsafeIncludes` — grounding accepts real citations / rejects missing-file and out-of-range-line citations; completeness accepts full coverage / rejects a missing operation, naming it; attestation refusal when either check fails; full render→attest→validate round trip on a hand-built fixture pair.
13. Run `go test ./installer/...` and confirm no regressions, especially against task 104's own new tests.

## Known Issues Triage
**Triaged:** 2026-08-07
**Tools searched:** PlantUML, @plantuml/mcp-js
**Result:** Clear

### Blocking Issues
- None

### Advisory Issues
- None — searched `plantuml/plantuml` open issues for "sequence diagram" (no platform keyword applicable); 20 results returned, none labeled `bug`/`regression`, and all concern plantuml.jar's own rendering engine (lifeline alignment, deactivate arrows, note spacing). None touch this task's actual mechanism: smaqit's own regex-based source-text scanning (participant/arrow-label extraction, `' impl: path:line` citation parsing), which operates on raw PlantUML text independently of plantuml.jar's parser/renderer and reuses the already-validated render pipeline from tasks 099/100/101 unmodified.

### Historical (Closed)
- None searched — no closed-issue workaround relevant to this task's mechanism.

### Unresolvable Tools
- @plantuml/mcp-js — no matching GitHub repository found via search

## Acceptance Criteria

- [x] `framework/ARTIFACTS.md` documents the new "Design Sequence Diagrams" category: ownership (Development, Phase 1), storage (`docs/designs/design-sequence/`), content boundary, and its own lifecycle (no reset-to-draft clause)
- [x] `smaqit.development` generates one Design Sequence Diagram per implemented Functional spec, grounded in `file:line` citations to the code it just wrote
- [x] `smaqit design attest` refuses to stamp a passing attestation when a cited `file:line` reference doesn't resolve to a real location in the codebase
- [x] `smaqit design attest` refuses to stamp a passing attestation when the diagram omits an operation its paired `system-sequence` SSD promises, naming the missing operation
- [x] A complete design-sequence diagram (all citations resolve, all SSD operations represented) attests and validates cleanly end-to-end via `smaqit design render` → `attest` → `validate`
- [x] `framework/PHASES.md`'s Phase 1 table, completion criteria, and Develop→Deploy prerequisites include the design-sequence diagram as a required output
- [x] `agents/development.md` directives, Phase-Specific Rules, and Completion Criteria require generating and validating design-sequence diagrams; its design-authoring MUST NOT line is reworded to scope explicitly to Design Artifacts
- [x] `smaqit init` scaffolds `docs/designs/design-sequence/`
- [x] Documentation and agent instructions frame the design-sequence diagram as a complement to code review, never a substitute
- [x] Full installer test suite passes with no unrelated regressions

## Findings

**Implementation approach:**
- New sibling artifact category "Design Sequence Diagrams" in `docs/designs/design-sequence/` (prefix `DSD`), extending `designProfiles`/`designLayerPrefix`/`designIDPattern` the same way every existing layer already works, so `smaqit design render/attest/validate` apply unmodified.
- Two new heuristic checks — `validateDesignSequenceGrounding` (every `' impl: <path>:<line>` citation must resolve to a real, in-range location) and `validateDesignSequenceCompleteness` (every operation the paired `system-sequence` design promises must have a matching label) — wired into `attestDesign` so attestation is earned, not just procedurally ordered.
- `validateDesignReferences` extended with a `specLayer`/`trackRank` branch so design-sequence designs resolve their `specifications` link against `functional` (not a nonexistent `specs/design-sequence/`) and are exempt from the Design Artifact lifecycle-rank coupling.
- `agents/development.md`'s design-authorship MUST NOT reworded to scope explicitly to Design Artifacts rather than exempted, keeping Bounded Agents intact for the five spec layers.
- 4 new tests covering grounding rejection/acceptance, completeness rejection/acceptance, attestation refusal on each failure mode, and a full Node/MCP-backed render→attest→validate round trip.

**Decisions made:**
- Storage as a `docs/designs/`-sibling tree, not nested in `docs/designs/functional/` — resolves three concrete conflicts (Bounded Agents ownership, Design Artifact reset-to-draft lifecycle rule, `system-sequence`-locked template shape) by construction rather than by exception. See task file Design Decisions for full reasoning.
- Grounding citations use PlantUML line comments (`' impl: path:line`), read directly by their own regex rather than through the shared `stripNonStructuralPlantUML` filter (which discards `'`-prefixed lines as comments) — deliberate, since citations need exactly the lines that filter removes.
- After task 104 merged: consolidated its inline note/title/legend-stripping logic into a shared `stripNonStructuralPlantUML` helper, now used by both `validateSystemSequenceProfile` (task 104) and `extractOperationLabels` (task 102) — confirmed behavior-preserving via task 104's own 11 pre-existing tests still passing unchanged.
- Corrected an inaccurate assumption from initial planning: task 104's shipped implementation does no arrow/message parsing at all (explicit scope decision in its own commit message), so there were no arrow-parsing helpers to reuse for the completeness check — `extractOperationLabels` remains task 102's own, necessarily independent code.

**Blockers encountered:**
- None. One git merge conflict (both task 102 and task 104 added a new conditional check to `validateDesignMetadata` in the same location) — resolved by keeping both checks sequentially, verified by the full test suite.

**Follow-up identified:**
- None required for this task. Optional future consideration: if a sixth diagram_type is ever added to the `design-sequence` profile, the current single-purpose `designProfiles["design-sequence"]` map entry would need to grow — not needed today (cardinality is deliberately one profile per this category).

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/design.go` | Modify — new layer/prefix/profile entries, `Realizes` field, `validateDesignSequenceGrounding`, `validateDesignSequenceCompleteness`, `attestDesign` wiring |
| `installer/design_test.go` | Modify — regression tests per Implementation Step 12 |
| `installer/main.go` | Modify — `docs/designs/design-sequence` scaffolding |
| `templates/designs/design-sequence.template.md` | Create — new sibling template |
| `framework/ARTIFACTS.md` | Modify — new Design Sequence Diagrams category section |
| `framework/PHASES.md` | Modify — Phase 1 table, completion criteria, Develop→Deploy prerequisites |
| `agents/development.md` | Modify — Output, new directive, Phase-Specific Rules, Completion Criteria, reworded MUST NOT |
| `.smaqit/definitions/agents/development.frontmatter.yaml` | Modify if needed — confirm during implementation |

## Notes

Motivated directly by a downstream project's task 052 pilot (2026-08-06), which produced 5 hand-authored diagrams of exactly this type that had already drifted from implementation three times. Refined via `smaqit.task-plan 102` (2026-08-07): the original "Realization Artifacts" framing was replaced with "Design Sequence Diagrams" per direct user naming, and storage moved to a `docs/designs/`-sibling category after resolving three concrete conflicts with placing it inside `docs/designs/functional/` — Bounded Agents ownership, the Design Artifact reset-to-draft lifecycle rule, and the `system-sequence`-locked template shape. Distinct from task 077 (retroactive specifications, not diagrams) and task 101 (toolchain reliability, a completed soft dependency). Soft-depends on task 104 (in progress) for shared PlantUML arrow-parsing helpers.
