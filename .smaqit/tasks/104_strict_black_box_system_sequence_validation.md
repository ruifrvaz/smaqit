# Strict Black-Box Validation for `system-sequence` Designs

**Status:** Not Started
**Created:** 2026-08-07

## Description

`system-sequence` is the Functional layer's default design profile (`framework/LAYERS.md:271-275`), modeled on the classic UML System Sequence Diagram: one actor (or several), one opaque system participant, no internal collaborators. Nothing in the current validation pipeline enforces that shape. `installer/design.go`'s `validateDesignMetadata` (line 299) checks that `diagram_type` is a *controlled string* for the layer (line 315-317) but never inspects the PlantUML body's actual participant structure.

Found via a downstream project's retry of its PlantUML design-pair pilot: the same spec produced two different diagrams across two runs, both tagged `diagram_type: system-sequence` — the first a true black box (`Customer` + one `IODIS CRM` system participant), the second naming three internal collaborators (`CreateHandler`, `ActivateHandler`, `ActSvc`) engaged in direct message exchange. Both passed every existing gate. The two runs diverged because nothing in the schema or CLI enforces the profile's own defining constraint — it's currently pure convention, honored only if the authoring agent happens to follow it.

This task adds a structural check that fails a `system-sequence` design whose PlantUML source declares more than one non-actor participant, and updates authoring guidance so agents facing a spec with more than one actor-initiated flow split into multiple per-flow/per-actor designs (already supported by the reference-resolution loop in `specDesignReady`) rather than merging everything into one diagram with multiple system-side participants.

## Design Decisions

- **Detection is a source-level heuristic scan, not a full PlantUML grammar parser.** No PlantUML parser exists anywhere in the codebase today (`unsafePlantUML` is the only source-level regex, and it only guards against `!include`). Classify every declared/referenced identifier as *actor* (`actor X` / `actor "Label" as X`) or *system-side* (explicit `participant`/`boundary`/`control`/`entity`/`database`/`collections`/`queue` declarations, plus any identifier that appears only implicitly on one side of a message arrow — PlantUML auto-creates undeclared participants exactly as the second pilot run's diagram would have without an explicit `System` declaration). Strip `note...end note`, `title`, `skinparam`, and `==divider==` lines first so their free text can't be misread as declarations.
- **Rule: exactly one system-side identifier per design.** Any actor count is fine; more than one non-actor identifier fails. Reuse `DESIGN-VISUAL-INVALID` rather than mint a new failure code — this is the same bucket the existing "diagram_type not controlled for layer" profile check already uses, and `ARTIFACTS.md` documents the failure-code list as stable.
- **The error message states the remedy, not just the violation** — e.g. naming the extra participants found and pointing at splitting into multiple per-flow/per-actor designs, so a failure is self-explanatory without reading this task.
- **Known, accepted limitation:** this is a lint-style heuristic, not a semantic verifier. An author (human or agent) could mislabel an internal handler as `actor` to dodge the check. Not closing that gap in this task — the goal is catching the honest/default-authoring case (exactly what both pilot runs actually produced), not adversarial-proofing the format.
- **No framework capability changes needed for multi-diagram specs.** Confirmed by reading `specDesignReady`/`specReferencesDesign` (design.go:454-533): the reference-resolution loop already iterates every linked design in a spec's `## Design References` section and only requires "at least one" valid pair — a spec linking several `system-sequence` designs (one per actor, or one per user story/flow) already validates correctly today. What's missing is authoring *guidance* telling agents to do this, not a capability gap.
- **Guidance mirrors an existing pattern, not a new one.** `framework/LAYERS.md`'s Business row already carries this exact idea for `use-case`: *"None; split by actor goal when readability requires it."* Functional's row gets the equivalent note for `system-sequence`, and `agents/smaqit.functional.md`'s directives get an explicit instruction to split by actor/flow rather than merge.

## Implementation Steps

1. Add `validateSystemSequenceProfile(d *designArtifact) error` to `installer/design.go`, called from `validateDesignMetadata` immediately after the existing `diagram_type` profile-allowlist check (line ~317), gated on `f.DiagramType == "system-sequence"`.
2. Implement the identifier-classification scan described in Design Decisions: strip non-structural lines, extract explicit declarations (`actor`/`participant`/`boundary`/`control`/`entity`/`database`/`collections`/`queue`, with and without quoted labels / `as` aliases), extract arrow-referenced identifiers (`->`, `-->`, `->>`, `-->>`, `<-`, `<--`, etc., including `activate`/`deactivate`/`create`/`destroy` lifeline statements), classify each into actor vs. system-side, fail with `DESIGN-VISUAL-INVALID` naming every system-side identifier found beyond the first, and suggesting the per-flow/per-actor split.
3. Add regression tests to `installer/design_test.go` mirroring `TestParseDesignRejectsProseAndUnsafeIncludes`'s style:
   - Rejects a multi-participant system-sequence (use the second pilot run's actual `CreateHandler`/`ActivateHandler`/`ActSvc` diagram as a fixture).
   - Accepts a true black-box form (actor(s) + one system participant) — confirm the first pilot run's `Customer`/`IODIS CRM` diagram still passes.
   - Edge cases: participant referenced only implicitly via an arrow (never explicitly declared), `activate`/`deactivate` lines, quoted labels with `as` aliases, multiple actors with one system participant (must pass).
   - Confirm `functional.template.md`'s own placeholder example and whatever `TestGreenfieldReferencePlantUMLIsValid` exercises still pass unmodified.
4. Update `framework/LAYERS.md`'s Functional row (line 272) to add split-by-actor/flow guidance for `system-sequence`, matching the Business row's existing phrasing style.
5. Update `agents/smaqit.functional.md`'s directives (MUST/SHOULD list, near the existing "Use the `use-case` diagram profile..."-style line) with an explicit instruction: author one `system-sequence` design per actor or per user-story/flow when a spec's behavior spans more than one, rather than merging them into a single diagram with multiple system-side participants.
6. Run the full installer test suite (`go test ./installer/...` or the project's documented equivalent) and confirm no unrelated regressions.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] A `system-sequence` design with more than one non-actor participant (explicit or arrow-implicit) fails `smaqit design validate` with `DESIGN-VISUAL-INVALID`, naming the extra participants and suggesting a per-flow/per-actor split
- [ ] A true black-box `system-sequence` design (any number of actors, exactly one system participant) validates cleanly
- [ ] `functional.template.md`'s placeholder example and existing greenfield-reference PlantUML fixtures still pass unmodified
- [ ] `framework/LAYERS.md`'s Functional row documents the split-by-actor/flow guidance, mirroring the Business row's existing pattern
- [ ] `agents/smaqit.functional.md`'s directives explicitly instruct authoring one `system-sequence` design per actor/flow rather than merging multiple flows into one diagram
- [ ] `specDesignReady`'s existing multi-design-per-spec support is exercised by at least one test with two linked `system-sequence` designs on the same spec
- [ ] Full installer test suite passes with no unrelated regressions

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
| `installer/design.go` | Modify — add `validateSystemSequenceProfile` and call site |
| `installer/design_test.go` | Modify — regression tests per Implementation Step 3 |
| `framework/LAYERS.md` | Modify — Functional row split-guidance note |
| `agents/smaqit.functional.md` | Modify — explicit per-actor/flow authoring directive |
| `.smaqit/definitions/agents/*.frontmatter.yaml` / generated agent mirrors | Modify if `scripts/generate-agents.py` regenerates them from `agents/smaqit.functional.md` — confirm during implementation |

## Notes

Motivated directly by a downstream project's task 052 retry (2026-08-06/07), which produced two structurally different `system-sequence` diagrams for the same spec across two pilot runs, both passing every existing gate. Distinct from task 102 (a different, unrelated artifact-taxonomy gap — post-implementation realization diagrams) and task 103 (MCP client-registration, already completed/released as v2.2.0). This task is purely about validation-time enforcement of a profile the framework already documents but never checks.
