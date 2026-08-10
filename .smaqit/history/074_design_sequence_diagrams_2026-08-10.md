# Design Sequence Diagrams

**Date:** 2026-08-10
**Session Focus:** Plan and implement a new post-implementation artifact category — the "Design Sequence Diagram" — that pairs with the black-box `system-sequence` diagram task 104 already validates, then consolidate the two once task 104 merged.
**Tasks Completed:** 102

---

## Actions Taken

### Session Start and Task Selection

- Loaded project context via `session.start`: README, most recent history (072, PlantUML runtime reliability), PLANNING.md, and the next unblocked task's source areas.
- User asked to plan "task 103," which was already Completed — clarified via `AskUserQuestion` that they meant task 102 (Post-Implementation Realization Diagrams), the next active item.

### Planning (`task.plan 102`)

- Assessed task 102 as Complex: Implementation Steps were entirely `TBD`, Design Decisions carried two unresolved TBDs (grounding mechanism, storage location).
- Ran three parallel Explore agents: how task 098 originally rolled out Design Artifacts (attestation model, `installer/design.go` mechanics), how `smaqit.development`/Phase 1 completion actually works today, and whether any existing precedent supports grounded file:line citation (none did).
- Presented a recommended lean, docs-only design; the user redirected significantly: full CLI parity mirroring Design Artifacts, storage under `docs/designs/functional/` (since these sequences are "truly the functional implementation"), and a deterministic CLI grounding check rather than self-attestation.
- Surfaced three concrete conflicts that direction created with the existing framework — Bounded Agents ownership, the Design Artifact reset-to-draft lifecycle rule, and the `system-sequence`-locked template shape — and proposed a `docs/designs/design-sequence/` sibling tree as the resolution that gets full technical parity without breaking any existing rule. User approved.
- Wrote the full execution plan and, on approval, updated the task file's Title, Description, Design Decisions, Implementation Steps, Acceptance Criteria, Files list, and Notes in place.

### Implementation (`task.start 102`, Assisted mode)

- Resolved as an owner task; created branch/worktree, ran issue triage (Clear — no relevant PlantUML issues), loaded the existing research map.
- `installer/design.go`: added `design-sequence` as a new pseudo-layer (prefix `DSD`) to `designProfiles`/`designLayerPrefix`/`designIDPattern`; added a conditionally-required `Realizes` frontmatter field without breaking the other five layers' schema; fixed `validateDesignReferences` to resolve design-sequence specs against `functional` (not a nonexistent `specs/design-sequence/`) and skip the lifecycle-rank coupling for this category; added `validateDesignSequenceGrounding` (citation existence check) and `validateDesignSequenceCompleteness` (operation-coverage check against the paired SSD), both wired into `attestDesign` so attestation is earned, not just procedurally ordered.
- Added a new sibling template (`templates/designs/design-sequence.template.md`) and `docs/designs/design-sequence/` scaffolding in `smaqit init`/`smaqit validate`.
- Updated `framework/ARTIFACTS.md` (new category section), `framework/PHASES.md` (Phase 1 table, completion criteria, Develop→Deploy prerequisites), and `agents/development.md` (new directive, Phase-Specific Rules step, Completion Criteria checkbox, and a precisely reworded — not exempted — Bounded Agents boundary line).
- Added 4 new tests; confirmed the full suite green.

### Consolidation with Task 104

- User reported task 104 (Strict Black-Box Validation for `system-sequence` Designs) had completed and merged; asked to consolidate.
- Discovered task 104's shipped implementation deliberately does *no* arrow/message parsing at all — its own commit message states this explicitly as a scope decision — invalidating task 102's original assumption that it would reuse task 104's arrow-parsing helpers.
- Merged `main` into task 102's branch (fast-forward, then reapplied stashed work), resolving one small conflict where both tasks added a new conditional check to `validateDesignMetadata`.
- Found real, previously-invisible duplication once task 104's actual code was visible: its inline note/title/legend-stripping state machine. Extracted it into a shared `stripNonStructuralPlantUML` helper, now used by both task 104's participant-classification scan and task 102's operation-label extraction — genuine DRY, confirmed behavior-preserving by task 104's own 11 pre-existing tests still passing unchanged.
- Corrected the task file's Design Decisions to record what actually happened rather than leave the earlier (pre-104-completion) assumption standing.

### Completion (`task.complete 102`)

- Verified all 10 acceptance criteria against the implemented worktree; wrote Findings (implementation approach, decisions, blockers, follow-up).
- Committed implementation in the worktree, merged into `main` with `--no-ff`, verified clean build/tests on primary post-merge, updated task status and `PLANNING.md`, removed the worktree, deleted the branch, and refreshed the workspace file.

## Problems Solved

- **No mechanism existed for a post-implementation "as-built" collaboration diagram.** `system-sequence` (task 104) captures only the external, pre-implementation contract; nothing previously documented how internal objects actually realize it, or verified that documentation against real code.
- **A plausible storage choice would have silently broken two existing framework rules.** Placing the new diagram inside `docs/designs/functional/` looked natural ("these sequences are truly the functional implementation") but would have violated Bounded Agents (spec-agent-owned directory) and triggered the Design Artifact reset-to-draft rule on every post-implementation regeneration. Resolved by construction via a sibling tree, not by carving exceptions into existing principles.
- **A real, only-visible-after-the-fact code duplication.** Task 104's actual implementation diverged from its own plan in a way that only became apparent once merged; the note/title/legend-stripping logic it ended up needing was independently re-derivable by task 102 but was better shared.

## Decisions Made

- **New sibling artifact category, not a Design Artifact sidecar** — `docs/designs/design-sequence/`, prefix `DSD`, owned procedurally by `smaqit.development`, with full CLI parity (same hash/attestation/render/validate machinery) but its own lifecycle text carrying no reset-to-draft clause.
- **Grounding via PlantUML line comments** (`' impl: <path>:<line>`), read directly rather than through the shared structural-line filter, since citations are exactly the `'`-prefixed lines that filter discards.
- **Attestation gates on two deterministic heuristic checks**, not subjective visual review — grounding (citation existence) and completeness (operation-label coverage against the paired SSD) — both explicitly documented as lint-style heuristics, not semantic verification.
- **Consolidate opportunistically, not by force.** Rather than assuming code reuse a priori, verified what task 104 actually shipped before claiming a dependency; extracted a shared helper only where genuine duplication existed (note/title/legend stripping), leaving `extractOperationLabels` independent since no arrow-parsing precedent existed to share.

## Files Modified

- `installer/design.go` — new `design-sequence` layer/profile/prefix entries, `Realizes` field, `validateDesignSequenceGrounding`, `validateDesignSequenceCompleteness`, `attestDesign` wiring, `validateDesignReferences` spec-layer/rank-tracking branch, shared `stripNonStructuralPlantUML` helper (post-merge consolidation).
- `installer/design_test.go` — 4 new tests (grounding, completeness, attestation gating, full render→attest→validate round trip).
- `installer/main.go` — `docs/designs/design-sequence` scaffolding in `smaqit init` and `smaqit validate`.
- `templates/designs/design-sequence.template.md` — new sibling template.
- `framework/ARTIFACTS.md`, `framework/PHASES.md`, `agents/development.md` — new artifact category documentation, phase-table/completion-criteria updates, new agent directives and a precisely reworded Bounded Agents boundary line.
- `.smaqit/tasks/102_post_implementation_realization_diagrams.md`, `.smaqit/tasks/PLANNING.md` — task lifecycle state through plan/start/complete.
- `smaqit.code-workspace` — worktree registration and cleanup.

## Next Steps

- No follow-up required for task 102 itself.
- Remaining Active backlog is unchanged by this session: 094 (feature-new browser/E2E gate), 077 (retroactive specifications), 074 (extensible-through-templates context), 071 (Q&A agent/skill for wiki docs) — all still status `new`.
- Noted in passing, not addressed: task 083 shows `Status: In Progress` on disk but is absent from `PLANNING.md`'s Active table — pre-existing stale bookkeeping, unrelated to this session.

## Session Metrics

- **Tasks completed:** 1 (102)
- **Tasks consolidated with:** 1 (104, merged mid-session)
- **New Go functions:** 5 (`validateDesignSequenceGrounding`, `validateDesignSequenceCompleteness`, `resolveRootRelativePath`, `extractOperationLabels`/`normalizeOperationLabel`, `stripNonStructuralPlantUML`)
- **New tests:** 4 (task 102) — full suite (15 pre-existing + 11 from task 104 + 4 new) green throughout
- **Merge conflicts resolved:** 1 (`validateDesignMetadata`, two independently-added conditional checks)
