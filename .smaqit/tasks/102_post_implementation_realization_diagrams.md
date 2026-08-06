# Post-Implementation Realization Diagrams for the Development Phase

**Status:** Not Started
**Created:** 2026-08-06

## Description

`framework/ARTIFACTS.md`'s Design Artifacts are explicitly pre-implementation spec sidecars ("Designs are canonical, layer-scoped sidecars for specifications"), and Phase 1's Implementation Artifacts (`framework/PHASES.md` line 11) are just "Code, README, Development report" — no diagram at all. None of the five layers' controlled design profiles (`framework/LAYERS.md:271-275`) covers a design-level collaboration/sequence diagram showing how internal objects actually realize a use case — the standard next step after a `system-sequence` SSD in the analysis-design methodology SSDs come from (Larman, *Applying UML and Patterns*): use-case → system sequence diagram (external contract) → design/collaboration sequence diagram (internal realization).

Found via a downstream project's own task (2026-08-06): a pilot of the new PlantUML design-pair workflow on a brownfield project revealed that 5 of that project's 12 hand-authored Mermaid diagrams (its equivalents of `sequence-order-lifecycle.md`, `sequence-admin-login.md`, `sequence-customer-registration-login.md`, `sequence-reminder-escalation.md`, `sequence-error-recovery.md`) are exactly this artifact type — and that keeping them hand-authored has already required three separate "fold drift back in" tasks to stay accurate against real implementation. For a greenfield project there is no such hand-authored set to fall back on at all — this class of diagram simply doesn't get produced anywhere in the current model.

**Proposal:** `smaqit.development` (Phase 1's implementation agent) generates one realization diagram per implemented Functional spec, after code/tests pass and before phase completion, as a new "Realization Artifact" category — distinct from the existing Design Artifact sidecar model, since it is implementation-derived rather than spec-derived. Value: (1) a validation gate — check every operation the paired `system-sequence` SSD promises is actually represented in the realization diagram, closing a completeness check smaqit currently has no mechanism for; (2) an easier human-approval surface than a raw diff before Deploy begins; (3) living documentation that can't hand-drift, because it's regenerated rather than hand-maintained.

## Design Decisions

- **New artifact category** ("Realization Artifacts"), not a sixth spec layer and not a Design Artifact sidecar — owned by `smaqit.development`, not any specification agent. Specification agents do not touch implementation; this artifact only exists after implementation.
- **Cardinality:** one realization diagram per implemented Functional spec, pairing 1:1 with that spec's `system-sequence` SSD.
- **Grounding requirement:** must be derived from checkable references (file:line citations to the actual code just written), not free-recall narration — otherwise it inherits the same reliability risk as a spec that drifts, just shifted earlier (wrong from the start instead of wrong later). Exact mechanism TBD — options include the agent citing call sites as it writes the diagram, or a deterministic static-analysis pass extracting the real call graph.
- **Complement, not substitute, for code review.** Especially for security-sensitive or edge-case-heavy code, a diagram is a lossy abstraction. Documentation and agent instructions must say this explicitly rather than imply the diagram replaces review.
- **Storage location:** TBD — not `docs/designs/<layer>/`, since that tree is spec-sidecar territory by definition (`ARTIFACTS.md`'s Design Artifacts section). Likely a new top-level tree, e.g. `docs/designs/realization/`.
- **Sequencing:** build on task 101 (Reliable Design Toolchain)'s rendering/toolchain fixes once landed, rather than reintroducing the same per-worktree/MCP-reachability gaps found there.

## Implementation Steps

TBD — this is filed as a proposal with a concrete mechanism sketched in Design Decisions, not a fully planned implementation. Refine via `smaqit.task-plan 102` before starting; expect it to touch at minimum `framework/ARTIFACTS.md`, `framework/PHASES.md`, `agents/development.md`, and the Development-phase skill/workflow files, mirroring the shape of task 098's original Design Artifact rollout.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `framework/ARTIFACTS.md` documents the new Realization Artifacts category: ownership, storage location, content boundary
- [ ] `framework/PHASES.md`'s Phase 1/Develop table and completion criteria include the realization diagram as an output, generated after tests pass and before phase completion
- [ ] `smaqit.development` generates one realization diagram per implemented Functional spec, grounded in checkable references to the code it just wrote
- [ ] A defined check confirms every operation the paired `system-sequence` SSD promises is represented in the realization diagram
- [ ] Documentation and agent instructions frame the realization diagram as a complement to code review, never a substitute
- [ ] The feature builds on task 101 rather than reintroducing its toolchain gaps

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
| `framework/ARTIFACTS.md` | Modify — new Realization Artifacts category |
| `framework/PHASES.md` | Modify — Phase 1 output table and completion criteria |
| `agents/development.md` | Modify — generate/ground/attest realization diagrams before phase completion |
| Development-phase skill/workflow files | Modify — scope TBD, refine via `smaqit.task-plan 102` |

## Notes

Motivated by a downstream project's own pilot (2026-08-06) of the mandatory PlantUML design-pair workflow on one of its Functional specs. Related but distinct from task 077 (retroactive *specifications* for brownfield projects — specs, not diagrams; also a much older, thin stub referencing the since-superseded L0/L1/L2 architecture) and task 101 (toolchain reliability — orthogonal to this artifact-model gap, but a soft dependency for this task's own rendering pipeline).
