# Phase Design-Readiness Gate Scans All Active Specs, Not Just the Touched Feature

**Status:** In Progress
**Created:** 2026-08-15
**Started:** 2026-08-14
**Mode:** Assisted

## Description

`smaqit plan --phase=develop|deploy|validate` runs a design-readiness gate (`validatePhaseDesignReadiness` in `installer/spec.go:187`) before returning any work-plan output. The gate is fed by `getPhaseDesignGateSpecs` (`installer/spec.go:223`), which — for each phase — collects **every non-deprecated spec in the phase's layers**, with no filtering for whether that spec was touched, drafted, or is even relevant to the feature currently being planned:

```go
switch phase {
case "develop":
    for _, layer := range []string{"business", "functional", "stack"} { appendActive(layer) }
case "deploy":
    for _, layer := range []string{"stack", "infrastructure"} { appendActive(layer) }
case "validate":
    for _, layer := range []string{"business", "functional", "stack", "infrastructure", "coverage"} { appendActive(layer) }
}
```

`validatePhaseDesignReadiness` then fails fast on the **first** spec in that list whose `## Design References` section is missing or unresolvable (`specDesignReady`, `installer/design.go:746`), via a bare `return fmt.Errorf(...)` on first miss (`installer/spec.go:190`) — not an aggregate report.

This makes the gate structurally incompatible with any project that is incrementally adopting the PlantUML design-pair convention (i.e., has pre-existing specs from before the convention existed, and is intentionally not batch-migrating all of them at once — an explicitly supported, documented workflow). In such a project, the gate can fail on a spec with zero relationship to the feature actually being planned, and there is no way to proceed short of giving every legacy spec in the phase's layers a design pair.

**Real-world impact, observed same-day in `iodis-crm-poc` (2026-08-13/14), independently by two concurrent sessions working unrelated features:**

- `--phase=develop`: failed on `specs/business/admin-authentication.md` — a spec neither session's feature touched, last modified weeks earlier.
- `--phase=deploy`: failed identically on `specs/stack/platform-stack.md` / `specs/infrastructure/deployment.md` — again, unrelated to either feature.
- `--phase=validate`: **far broader** — failed on 21 pre-existing specs across Business (7), Functional (7), and Coverage (7), spanning bounded contexts (Identity & Access, Currency, Email, CRM Navigation) neither feature owned. One of the 21 was the *feature's own* Coverage spec — its project had an established, deliberate convention that Coverage-layer specs never carry a design pair (traceability lives in the Coverage Map/Gherkin scenarios instead), so even a "fix" attempt on that one spec would have required violating that project's own convention.

One team member's assessment, verified directly against this file: the design-readiness gate is not scoped by phase output relevance at all — it is a blanket sweep of "everything active in these layers," decoupled from `getPhaseSpecs` (`installer/spec.go:196`), the sibling function that already exists for phase-scoped spec resolution and presumably has (or could have) the filtering this gate lacks.

**What worked as a real, adopted fix (not a bypass) for the `deploy`-phase case:** the project team authored genuine, minimal PlantUML design pairs for the two blocking specs, rather than accepting an agent-proposed bypass — this is the correct outcome the gate is meant to force, and it worked exactly as intended *when the blocking spec count is small and within the feature's reach*. The problem is specifically the `validate`-phase case (and, in a large enough legacy project, potentially `develop`/`deploy` too): when the blocking set is large and spans unrelated bounded contexts, "author real design pairs for everything" stops being a reasonable ask for a single feature task, and becomes the batch migration this framework's own incremental-adoption model is meant to avoid forcing.

## Design Decisions

TBD — open questions for whoever picks this up:

- Should the gate scope to specs with `status: draft` (i.e., specs a spec agent actually touched this cycle), mirroring how `smaqit plan`'s non-gate path already reports "N pending" specs? This seems like the most direct fix: reuse whatever draft-detection already drives the plan's own pending-count output, rather than the full active-layer sweep.
- Should `validatePhaseDesignReadiness` report *all* failing specs (aggregate) rather than fail-fast on the first, even after scoping is fixed? An aggregate report is strictly more useful once the set is meant to be small (post-fix), and removes the "which one blocked me, and are there more after I fix it" guessing game users currently hit.
- Should a project be able to declare "layer X's specs are exempt from this gate" (e.g., a documented incremental-adoption flag), for projects that have decided a whole layer (like Coverage, in the observed case) never carries design pairs by convention? Or is scoping-to-draft sufficient to make this moot?

## Implementation Steps

TBD — sketch, not committed:

1. Determine what already distinguishes "this cycle's specs" from "all active specs in the layer" elsewhere in the codebase (likely status-based: `draft`/`failed` vs. `implemented`/`deployed`/`validated`) — `installer/spec.go`'s existing status-transition logic (referenced by `spec-lifecycle-reference.md` in `smaqit.feature-new`) is the probable source of truth.
2. Change `getPhaseDesignGateSpecs` to filter to that same touched/draft set, for all three phases, instead of every non-deprecated spec in the layer.
3. Decide and implement fail-fast vs. aggregate reporting in `validatePhaseDesignReadiness` (see Design Decisions).
4. Add regression coverage: a project fixture with an old, unrelated, un-paired spec plus a new draft spec in the same layer — confirm the gate no longer blocks on the old one.
5. Update any docs/help text describing this gate's scope.

## Known Issues Triage

**Triaged:** 2026-08-15
**Tools searched:** none
**Result:** Clear — this is an internal bug in `installer/spec.go`/`design.go`, not a third-party dependency issue.

## Acceptance Criteria

- [ ] `smaqit plan --phase=develop|deploy|validate` no longer fails on a spec that the current feature/cycle did not touch and is not `status: draft`/`failed`
- [ ] The gate's failure output (if it still fails) reflects only specs actually relevant to the current cycle's scope
- [ ] Regression test fixture: an old, unrelated, un-paired spec in the same layer as a feature's new draft spec does not block that feature's phase plan
- [ ] Decision recorded on fail-fast vs. aggregate reporting, and implemented accordingly

## Findings

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
| `installer/spec.go` | Modify — `getPhaseDesignGateSpecs`, `validatePhaseDesignReadiness` |
| `installer/spec_test.go` | Modify — regression coverage |
| `installer/design.go` | Reference only — `specDesignReady` itself is not the bug; the *input set* fed to it is |

## Notes

Found and root-caused live in `iodis-crm-poc` during a `smaqit.feature-new` cycle (their tasks 055/061-065 and, concurrently and independently, tasks 054/056-060), same calendar day, by two separate sessions hitting the identical bug from different features. Also referenced there as a documented project-level workaround pattern (`smaqit-framework-gap-plan-phase-develop-scope` / `smaqit-framework-gap-plan-phase-validate-scope` in that project's session memory) until this upstream fix lands. One session in that project also independently proposed an agent-side bypass for the `deploy`-phase case; the project's real user explicitly rejected it in favor of authoring genuine design pairs — worth noting as a signal that users want this gate *enforced correctly*, not routed around, which is exactly why the scoping (not the enforcement) is the right thing to fix here.
