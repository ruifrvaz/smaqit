# Strict Black-Box Validation for `system-sequence` Designs

**Status:** Completed
**Created:** 2026-08-07
**Started:** 2026-08-07
**Completed:** 2026-08-10
**Mode:** Assisted

## Description

`system-sequence` is the Functional layer's default design profile (`framework/LAYERS.md:271-275`), modeled on the classic UML System Sequence Diagram: one actor (or several), one opaque system participant, no internal collaborators. Nothing in the current validation pipeline enforces that shape. `installer/design.go`'s `validateDesignMetadata` (line 299) checks that `diagram_type` is a *controlled string* for the layer (line 315-317) but never inspects the PlantUML body's actual participant structure.

Found via a downstream project's retry of its PlantUML design-pair pilot: the same spec produced two different diagrams across two runs, both tagged `diagram_type: system-sequence` — the first a true black box (`Customer` + one `IODIS CRM` system participant), the second naming three internal collaborators (`CreateHandler`, `ActivateHandler`, `ActSvc`) engaged in direct message exchange. Both passed every existing gate. The two runs diverged because nothing in the schema or CLI enforces the profile's own defining constraint — it's currently pure convention, honored only if the authoring agent happens to follow it.

This task adds a structural check that fails a `system-sequence` design whose PlantUML source declares more than one non-actor participant, and updates authoring guidance so agents facing a spec with more than one actor-initiated flow split into multiple per-flow/per-actor designs (already supported by the reference-resolution loop in `specDesignReady`) rather than merging everything into one diagram with multiple system-side participants.

## Design Decisions

- **Detection is a source-level scan over explicit declarations only, not a full PlantUML grammar parser and not arrow/message parsing.** No PlantUML parser exists anywhere in the codebase today (`unsafePlantUML` is the only source-level regex, and it only guards against `!include`). Classify every explicitly *declared* identifier (`actor X` / `actor "Label" as X`, or `participant`/`boundary`/`control`/`entity`/`database`/`collections`/`queue` with or without an `as` alias) as actor or system-side; message/arrow lines are not scanned. Strip `note...end note`, `title...end title`, and `legend...endlegend` blocks first so their free text can't be misread as declarations.
- **Rule (superseded from the original "any actor count is fine" decision): exactly one actor and exactly one system-side identifier per design, and the system-side identifier must be exactly `System`.** More than one actor, more than one system-side declaration, zero of either, or a system participant not literally named/aliased `System` all fail. This closes a gap the original any-actor-count rule left open (no fixed convention to check the system box's *name* against) while also being simpler to implement: declaration-only scanning eliminates the arrow/message-parsing surface area entirely (decorated arrows, `create <type> <name>` shorthand, alias trailing-decoration handling) that a first implementation pass of this rule found buggy in practice. Reuse `DESIGN-VISUAL-INVALID` rather than mint a new failure code — this is the same bucket the existing "diagram_type not controlled for layer" profile check already uses, and `ARTIFACTS.md` documents the failure-code list as stable.
- **The error message states the remedy, not just the violation** — e.g. naming the extra actors/participants found and pointing at splitting into multiple per-flow/per-actor designs, or naming what the system participant was found identified as instead of `System`, so a failure is self-explanatory without reading this task.
- **Known, accepted limitations:** (1) this is a lint-style heuristic, not a semantic verifier — an author (human or agent) could mislabel an internal handler as `actor` to dodge the check; (2) declaration-only scanning does not see participants PlantUML auto-creates purely from an undeclared arrow reference (e.g. a bare `Handler -> Service: ...` where neither was ever declared) — the existing shipped template already always declares both the actor and the system participant explicitly, so this is not expected to be the honest/default-authoring case. Not closing either gap in this task — the goal is catching the honest/default-authoring case (exactly what both pilot runs actually produced, both of which used explicit declarations), not adversarial-proofing the format.
- **No framework capability changes needed for multi-diagram specs.** Confirmed by reading `specDesignReady`/`specReferencesDesign` (design.go:454-533): the reference-resolution loop already iterates every linked design in a spec's `## Design References` section and only requires "at least one" valid pair — a spec linking several `system-sequence` designs (one per actor, or one per user story/flow) already validates correctly today. What's missing is authoring *guidance* telling agents to do this, not a capability gap.
- **Guidance mirrors an existing pattern, not a new one.** `framework/LAYERS.md`'s Business row already carries this exact idea for `use-case`: *"None; split by actor goal when readability requires it."* Functional's row gets the equivalent note for `system-sequence`, and `agents/smaqit.functional.md`'s directives get an explicit instruction to split by actor/flow rather than merge.

## Implementation Steps

1. Add `validateSystemSequenceProfile(d *designArtifact) error` to `installer/design.go`, called from `validateDesignMetadata` immediately after the existing `diagram_type` profile-allowlist check (line ~317), gated on `f.DiagramType == "system-sequence"`.
2. Implement the declaration-only classification scan described in Design Decisions: strip `note`/`title`/`legend` blocks, extract explicit declarations (`actor`/`participant`/`boundary`/`control`/`entity`/`database`/`collections`/`queue`, with and without quoted labels / `as` aliases), classify each into actor vs. system-side. Fail with `DESIGN-VISUAL-INVALID` when actor count != 1 (naming any extras and suggesting the per-actor split), when system-side count != 1 (naming any extras and suggesting the per-flow split), or when the single system-side identifier is not exactly `System` (naming what it was found as instead). Arrow/message lines are not parsed.
3. Add regression tests to `installer/design_test.go` mirroring `TestParseDesignRejectsProseAndUnsafeIncludes`'s style:
   - Rejects a multi-participant system-sequence (use the second pilot run's actual `CreateHandler`/`ActivateHandler`/`ActSvc` diagram as a fixture).
   - Accepts a true black-box form (one actor, one system participant identified as `System`) — confirm the first pilot run's `Customer`/`IODIS CRM` diagram (aliased `as System`) still passes.
   - Rejects more than one actor; rejects a system participant declared but not identified as `System`; rejects a diagram where `System` is only ever referenced in messages and never explicitly declared.
   - Edge cases: quoted labels with `as` aliases plus trailing decoration (color/stereotype) still resolve to the correct alias, multi-line `title`/`legend` blocks are not scanned as content.
   - Confirm `functional.template.md`'s own placeholder example and whatever `TestGreenfieldReferencePlantUMLIsValid` exercises still pass unmodified.
4. Update `framework/LAYERS.md`'s Functional row (line 272) to add split-by-actor/flow guidance for `system-sequence`, matching the Business row's existing phrasing style.
5. Update `agents/smaqit.functional.md`'s directives (MUST/SHOULD list, near the existing "Use the `use-case` diagram profile..."-style line) with an explicit instruction: author one `system-sequence` design per actor or per user-story/flow when a spec's behavior spans more than one, rather than merging them into a single diagram with multiple system-side participants.
6. Run the full installer test suite (`go test ./installer/...` or the project's documented equivalent) and confirm no unrelated regressions.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] A `system-sequence` design with more than one declared actor, more than one declared system-side participant, or a system participant not identified as `System` fails `smaqit design validate` with `DESIGN-VISUAL-INVALID`, naming the problem and suggesting the remedy (per-actor or per-flow split, or renaming/aliasing the system participant to `System`)
- [x] A true black-box `system-sequence` design (exactly one actor, exactly one system participant identified as `System`) validates cleanly
- [x] `functional.template.md`'s placeholder example and existing greenfield-reference PlantUML fixtures still pass unmodified
- [x] `framework/LAYERS.md`'s Functional row documents the split-by-actor/flow guidance, mirroring the Business row's existing pattern
- [x] `agents/smaqit.functional.md`'s directives explicitly instruct authoring one `system-sequence` design per actor/flow rather than merging multiple flows into one diagram
- [x] `specDesignReady`'s existing multi-design-per-spec support is exercised by at least one test with two linked `system-sequence` designs on the same spec
- [x] Full installer test suite passes with no unrelated regressions

## Findings

**Implementation approach:**
- `validateSystemSequenceProfile` in `installer/design.go` scans only explicit `actor`/`participant`-family declaration lines (skipping `note`/`title`/`legend` blocks defensively), classifies each as actor or system-side, and requires exactly one actor, exactly one system-side declaration, and that declaration's identifier to be exactly `System`.
- No arrow/message-line parsing — the original heuristic's arrow-token, lifeline-op, and skip-prefix regexes were removed entirely rather than patched, closing every bug an initial code-review pass found (decorated arrows, `create` shorthand, alias trailing-decoration, unhandled multi-line title/legend bodies, declaration-order dependence).
- `framework/LAYERS.md` and `agents/functional.md` updated to require one actor plus one `System`-named participant, and to make the per-actor/per-flow design split mandatory rather than a readability suggestion.

**Decisions made:**
- Mid-task, superseded the original "any actor count is fine" rule with a hard one-actor requirement, and added a fixed identity requirement (`System`) for the system participant — both verified against the already-shipped `functional.template.md` convention (which already used exactly this shape) before implementing, so neither is a new convention being invented.
- Kept detection declaration-only rather than reintroducing arrow/message parsing to also catch PlantUML-auto-created implicit participants; documented as a known, accepted heuristic limitation consistent with the task's existing "lint check, not adversarial-proof" stance.
- `System` match is case-sensitive with no case-insensitive fallback.

**Blockers encountered:**
- The prior session's `task-start` for this task never committed its "In Progress" status/mode change on the primary checkout (only the worktree had it, uncommitted), which blocked `task-complete`'s lifecycle resolver. Repaired by committing the missing status/mode/`PLANNING.md` update on primary before retrying.

**Follow-up identified:**
- None beyond the two known, accepted heuristic limitations already documented in Design Decisions (actor mislabeling, undeclared arrow-implicit participants).

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/design.go` | Modify — add `validateSystemSequenceProfile` and call site |
| `installer/design_test.go` | Modify — regression tests per Implementation Step 3 |
| `framework/LAYERS.md` | Modify — Functional row split-guidance note |
| `agents/smaqit.functional.md` | Modify — explicit per-actor/flow authoring directive |
| `.smaqit/definitions/agents/*.frontmatter.yaml` / generated agent mirrors | Modify if `scripts/generate-agents.py` regenerates them from `agents/smaqit.functional.md` — confirm during implementation |

## Notes

Motivated directly by a downstream project's task 052 retry (2026-08-06/07), which produced two structurally different `system-sequence` diagrams for the same spec across two pilot runs, both passing every existing gate. Distinct from task 102 (a different, unrelated artifact-taxonomy gap — post-implementation realization diagrams) and task 103 (MCP client-registration, already completed/released as v2.2.0). This task is purely about validation-time enforcement of a profile the framework already documents but never checks.
