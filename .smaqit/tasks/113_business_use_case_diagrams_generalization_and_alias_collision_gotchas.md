---
status: Not Started
created: "2026-08-20"
---

# Business Use-Case Diagrams: No Guardrail Against Generalization-Inheritance Misuse or PlantUML-Alias Collision With the Project's Own UC-ID Convention

## Description

Found live in a downstream project (`iodis-crm-poc`, task 082 — Phase 1/Spec Revalidation of a `smaqit.feature-new` cycle for an RBAC feature). The `smaqit-business` agent authored a PlantUML use-case diagram (`docs/designs/business/dsg-bus-rbac-use-case.md`) modeling 6 specialized actors (Administrator, Sales Manager, Sales Representative, Finance Manager, Operations, Read-only) each generalizing a shared base actor (`Admin`). The diagram had two distinct, real defects, both caught by the downstream project's human operator during review — not by any smaqit tooling:

### Defect 1 — a use-case association meant to be exclusive was drawn from the shared base actor, silently granting it to every specialization

The generated source had:
```plantuml
Admin --> UC1   ' Access Feature-Area Page or Action — legitimately universal
Admin --> UC3   ' Manage Admin Accounts & Roles — meant to be Administrator-only
```
with all 6 role actors generalizing `Admin` via `X --|> Admin`. In real UML use-case semantics, generalization is inherited exactly like class inheritance: every association on the base actor is automatically inherited by every specialization. So `Admin --> UC3` meant the diagram was asserting that all 6 roles — including Read-only — could manage admin accounts and roles, the exact opposite of the feature's own design (only a role holding a specific permission can do this; by default, only Administrator). The fix (verified correct by the same human reviewer after a second pass) was moving that one association from the shared base actor to the specific specialization actor that legitimately has it (`Administrator --> UC3`), leaving only the genuinely universal association (`Admin --> UC1`) on the base.

Nothing in `agents/business.md`, `framework/ARTIFACTS.md`, `templates/designs/business.template.md`, or `skills/smaqit.design-validate/SKILL.md`'s visual-review checklist mentions UML generalization/inheritance semantics at all — confirmed via a full grep sweep for `generaliz`/`inherit` across all of `framework/*.md`, `agents/*.md`, and the design-validate skill. The design-validate checklist's closest item ("coherent system... boundaries") is about visual clipping/boundary rendering, not inheritance-of-access semantics. `templates/designs/business.template.md`'s entire PlantUML skeleton is a single-actor, single-use-case toy example — it models nothing about multiple actors, generalization, or exclusive-vs-shared use-case access, so the agent had zero worked example to follow or deviate from correctly.

### Defect 2 — the diagram's own PlantUML aliases collided with the project's own established use-case ID convention

The generated source named its six use-case bubbles `UC1` through `UC6` as local PlantUML aliases (`usecase "..." as UC1`, etc. — pure internal shorthand, invisible in the rendered image, which only ever shows the full descriptive labels). But this exact project already has real business-spec files named `uc3-email.md` (`id: BUS-EMAIL`) and, ironically, the RBAC spec this very diagram belongs to is itself `uc8-rbac.md` (`id: BUS-RBAC`) — both following the `UC[N]-[CONCEPT]` convention that `agents/business.md`'s own "Use Case ID Format" section (L127-148) establishes. When the human reviewer read the diagram's PlantUML source directly (visible in the same file as the spec it documents), seeing `Admin --> UC3` was genuinely ambiguous with "this connects to `uc3-email.md`'s business use case" — a real, avoidable confusion, not a misunderstanding on the reviewer's part. The fix was renaming every alias to a descriptive name derived from its own use-case label (`AccessFeatureArea`, `DenyAccess`, `ManageAdminAccounts`, `CreateAdminAccount`, `AdministerRoles`, `RejectSelfLockout`) — a purely mechanical rename with zero visual/rendered impact (confirmed: `image_sha256` was byte-identical before and after).

The section of `agents/business.md` that *establishes* the `UC[N]` convention the diagram's aliases collided with is in the same file as the (nonexistent) instruction that should have warned the agent away from reusing that exact pattern as a diagram-internal alias — the agent had no cross-reference connecting the two.

## Design Decisions

Both fixes below are adopted, following this framework's own established preference (see task 111's Design Decisions) for real code-level enforcement over prose-only discipline, with prose/template updates as a necessary companion where full semantic detection isn't feasible.

- **Fix 1 (Defect 2 — alias collision): a genuine, fully deterministic Go-level check.** This is mechanically detectable from source text alone with no semantic ambiguity — a PlantUML alias literally matching the project's own reserved `UC\d+` pattern (the same pattern `agents/business.md`'s Use Case ID Format section defines) inside a `layer: business`, `diagram_type: use-case` design is unconditionally wrong, since that exact token space is reserved for the project's own spec IDs/filenames. Add a hard validation failure to `installer/design.go` (alongside `validateSystemSequenceProfile`'s existing precedent of a dedicated per-diagram-type semantic profile — a new `validateUseCaseProfile` is the natural home) that regex-matches every `as <alias>` declaration in a `use-case` diagram's PlantUML source against `^UC\d+$` (case-insensitive) and fails with a message naming the offending alias and recommending a descriptive replacement.

- **Fix 2 (Defect 1 — generalization/inheritance misuse): a deterministic *advisory* diagnostic, not a hard failure, plus mandatory reviewer guidance.** True semantic detection ("was this association meant to be inherited by every specialization, or was that an authoring mistake") is not fully automatable from source alone — sometimes an association on a generalized-from base actor is exactly correct (as `Admin --> AccessFeatureArea` legitimately is in the fixed diagram). What *is* deterministically detectable is the coexistence pattern itself: one or more actors declared via `X --|> BaseActor`, combined with one or more direct `BaseActor --> SomeUseCase` associations. Add a new advisory diagnostic (not a validation failure — `smaqit design validate` should still pass) emitted whenever this pattern is present, e.g. `DESIGN-REVIEW-ADVISORY: actor "<Base>" is generalized by <N> other actor(s) and has <M> direct use-case association(s) — confirm each is intentionally meant to be inherited by every specialization, not exclusive to one`. Pair this with:
  - A new mandatory checklist item in `skills/smaqit.design-validate/SKILL.md`'s shared visual-review checklist (used by all five spec agents, not just business): explicitly verify, for any use-case diagram containing actor generalization, that no use-case association on the generalized-from base actor is meant to be exclusive to a subset of its specializations — and if it is, that association must be redrawn from the specific specialization actor(s), not the base.
  - A worked example added to `templates/designs/business.template.md` (or a new example block in `agents/business.md`'s Layer-Specific Rules) showing the correct pattern side-by-side: a base actor with a genuinely universal association, plus a specialization actor with its own additional, exclusive association — matching the corrected shape of the real downstream diagram this task is based on.

- **No grandfathering, no opt-out, no "only diagrams authored after this ships" carve-out** — consistent with this repository's own current "No Grandfathering" principle (see task 112, being formalized concurrently with this filing). Fix 1 will immediately fail `smaqit design validate` for any pre-existing downstream use-case diagram that happens to use a `UC\d+`-shaped alias (a narrow, mechanically-checkable pattern — low expected blast radius, but real regardless of count). Fix 2's advisory is non-blocking by design (see above), so it carries no grandfathering question at all — it always evaluates against current source, for every diagram, every time.

- **Scope boundary:** both fixes are scoped to `layer: business`, `diagram_type: use-case` diagrams only, matching where both defects were actually observed and where the `UC[N]` convention Fix 1 protects is actually defined. Do not extend either fix speculatively to `system-sequence`/`design-sequence`/other diagram types in this task — if an analogous ID-collision or generalization-misuse pattern is later found in another layer's diagram type (e.g. a `FUN[N]`/`COV[N]`-shaped alias collision), file that as its own task once actually observed, rather than guessing at a shared abstraction now.

## Implementation Steps

TBD — sketch, not committed:

1. Fix 1: add `validateUseCaseProfile` (or equivalent) to `installer/design.go`, gated on `diagram_type == "use-case"` (mirroring how `validateSystemSequenceProfile` is gated on `diagram_type == "system-sequence"`). Regex-scan the PlantUML source for `\bas\s+(\w+)` declarations (both `actor ... as X` and `usecase ... as X` forms) and fail if any captured alias matches `^UC\d+$` case-insensitively. Add a regression test: a use-case diagram with an alias like `UC3` fails with a clear message; the same diagram with a descriptive alias (e.g. `ManageAdminAccounts`) passes; an alias that merely *contains* "uc" as a substring of a real word (e.g. `Truck`) is not a false-positive match.
2. Fix 2: add generalization/base-association detection to the same or an adjacent function — parse `X --|> Y` generalization edges and `Y --> Z` direct associations from the same PlantUML source; for every actor `Y` that appears as a generalization target with 1+ specializations AND has 1+ direct use-case associations, emit the advisory diagnostic (non-fatal — collected separately from hard validation errors, surfaced in `smaqit design validate`'s output but not counted toward its pass/fail exit code). Add a regression test confirming the advisory fires for the exact shape of the original (broken) downstream diagram and does not fire for a use-case diagram with no generalization at all.
3. Update `skills/smaqit.design-validate/SKILL.md`'s visual-review checklist with the new mandatory generalization-inheritance verification item.
4. Update `templates/designs/business.template.md` and/or `agents/business.md` with a worked correct-vs-incorrect example for the generalization/exclusive-association pattern.
5. Update any doc/help text describing `smaqit design validate`'s output shape to mention the new advisory-diagnostic category (distinct from hard `DESIGN-*-INVALID`/`DESIGN-ARTIFACT-*` failures), if this is the first diagnostic of that kind — check whether an advisory/non-fatal diagnostic channel already exists anywhere in `design.go` before inventing a new one.

## Known Issues Triage

**Triaged:** 2026-08-20
**Tools searched:** none
**Result:** Clear — internal gap in `installer/design.go`/`agents/business.md`/`skills/smaqit.design-validate/SKILL.md`, not a third-party dependency issue.

## Acceptance Criteria

- [ ] A `layer: business`, `diagram_type: use-case` design whose PlantUML source declares any alias matching `^UC\d+$` (case-insensitive) fails `smaqit design validate` with a clear message naming the offending alias
- [ ] A use-case diagram with descriptive aliases (no `UC\d+` collisions) is unaffected by Fix 1
- [ ] A use-case diagram containing actor generalization plus a direct use-case association on the generalized-from base actor produces a non-fatal advisory diagnostic, visible in `smaqit design validate`'s output, without failing validation
- [ ] A use-case diagram with no actor generalization produces no such advisory
- [ ] `skills/smaqit.design-validate/SKILL.md`'s visual-review checklist includes the new mandatory generalization-inheritance verification item
- [ ] `templates/designs/business.template.md` and/or `agents/business.md` includes a worked correct-vs-incorrect example for the generalization/exclusive-association pattern
- [ ] Regression tests cover both fixes independently, including the false-positive guard for Fix 1 (an alias merely containing "uc" as a substring must not match)
- [ ] No grandfathering/legacy-opt-out mechanism exists for either fix

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
| `installer/design.go` | Modify — `validateUseCaseProfile` (Fix 1, hard failure), generalization/base-association advisory detection (Fix 2, non-fatal) |
| `installer/design_test.go` | Modify — regression coverage for both fixes |
| `skills/smaqit.design-validate/SKILL.md` | Modify — new mandatory checklist item |
| `templates/designs/business.template.md` | Modify — worked generalization example |
| `agents/business.md` | Modify — cross-reference between Use Case ID Format and diagram-alias guidance; worked example if not added to the template instead |
| `CHANGELOG.md` | Modify — version entry |

## Notes

Found live in `iodis-crm-poc`, task 082 (Phase 1/Spec Revalidation of task 080's `smaqit.feature-new` RBAC cycle, 2026-08-20). Both defects were caught by the downstream project's human operator during Assisted-mode review of the `smaqit-business` agent's output — asking, first, "does this mean all these 6 types are admins?" (Defect 1) and then, after that fix, "what is UC3? from a user story perspective uc3 is the uc3-email business story" (Defect 2). Neither defect was caught by any smaqit tooling (`smaqit design validate`, the MCP `check_syntax`/`render_diagram` tools, or the visual-attestation step) at any point — both fixes were applied by hand in the downstream project, re-rendered, and re-attested, but the underlying authoring gap in `smaqit-business` itself remains open until this task resolves it, so the same defects will recur on the next business-layer use-case diagram this framework generates for any project with multiple specialized actors.

This is the same general category as task 111 (a design-authoring gap found live in a downstream project, not hypothesized, with concrete file citations and a verified fix already applied downstream by hand) — recommend similar priority tier.
