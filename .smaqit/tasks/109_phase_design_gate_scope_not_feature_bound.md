# Phase Design-Readiness Gate Scans All Active Specs, Not Just the Touched Feature

**Status:** Completed
**Created:** 2026-08-15
**Started:** 2026-08-14
**Completed:** 2026-08-14
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

**Real-world impact, observed same-day in a downstream project (2026-08-13/14), independently by two concurrent sessions working unrelated features:**

- `--phase=develop`: failed on `specs/business/admin-authentication.md` — a spec neither session's feature touched, last modified weeks earlier.
- `--phase=deploy`: failed identically on `specs/stack/platform-stack.md` / `specs/infrastructure/deployment.md` — again, unrelated to either feature.
- `--phase=validate`: **far broader** — failed on 21 pre-existing specs across Business (7), Functional (7), and Coverage (7), spanning bounded contexts (Identity & Access, Currency, Email, CRM Navigation) neither feature owned. One of the 21 was the *feature's own* Coverage spec — its project had an established, deliberate convention that Coverage-layer specs never carry a design pair (traceability lives in the Coverage Map/Gherkin scenarios instead), so even a "fix" attempt on that one spec would have required violating that project's own convention.

One team member's assessment, verified directly against this file: the design-readiness gate is not scoped by phase output relevance at all — it is a blanket sweep of "everything active in these layers," decoupled from `getPhaseSpecs` (`installer/spec.go:196`), the sibling function that already exists for phase-scoped spec resolution and presumably has (or could have) the filtering this gate lacks.

**What worked as a real, adopted fix (not a bypass) for the `deploy`-phase case:** the project team authored genuine, minimal PlantUML design pairs for the two blocking specs, rather than accepting an agent-proposed bypass — this is the correct outcome the gate is meant to force, and it worked exactly as intended *when the blocking spec count is small and within the feature's reach*. The problem is specifically the `validate`-phase case (and, in a large enough legacy project, potentially `develop`/`deploy` too): when the blocking set is large and spans unrelated bounded contexts, "author real design pairs for everything" stops being a reasonable ask for a single feature task, and becomes the batch migration this framework's own incremental-adoption model is meant to avoid forcing.

## Design Decisions

- **Scope to `draft`/`failed` status — adopted.** `getPhaseDesignGateSpecs` now calls a shared `isCycleRelevant(status)` predicate (`status == "draft" || status == "failed"`) instead of only excluding `deprecated`. This is the exact predicate `filterSpecsByStatus` already used for the "N pending" accounting, so the gate and the pending-count logic now agree on what "this cycle" means — no new concept introduced.
- **Aggregate reporting — adopted.** `validatePhaseDesignReadiness` now collects every blocking spec in the scoped input and returns one newline-joined error instead of returning on the first miss. Once the input is scoped to the current cycle the blocking set is expected to be small, so surfacing all of it in one pass removes the "fix one, rerun, find the next" loop. The existing call site (`main.go:251-254`) needed no change — it already just prints `%v` and exits.
- **Per-layer exemption flag — not implemented, left as a follow-up.** The `validate`-phase real-world impact included the *feature's own* Coverage spec failing because that project's own convention is "Coverage specs never carry a design pair." Scoping to draft/failed does **not** fix that specific case — a draft Coverage spec is still in-cycle and still lacks a design pair by that project's deliberate choice, so it still blocks. That is a distinct question from this task's bug (structural over-scoping to unrelated legacy specs) — it's about whether smaqit's canonical design convention (which does include a Coverage design template, per `templates/designs/`) can be intentionally opted out of per layer, which is a product decision requiring its own design pass rather than something to improvise inside this bug fix. Recorded here so it isn't lost; a fresh task should propose the exemption mechanism (or an explicit "no, author the pair" answer) on its own merits.

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

- [x] `smaqit plan --phase=develop|deploy|validate` no longer fails on a spec that the current feature/cycle did not touch and is not `status: draft`/`failed`
- [x] The gate's failure output (if it still fails) reflects only specs actually relevant to the current cycle's scope
- [x] Regression test fixture: an old, unrelated, un-paired spec in the same layer as a feature's new draft spec does not block that feature's phase plan
- [x] Decision recorded on fail-fast vs. aggregate reporting, and implemented accordingly

## Findings

**Implementation approach:**
- Added `isCycleRelevant(status string) bool` in `installer/spec.go`, shared between `getPhaseDesignGateSpecs`'s `appendActive` closure and (conceptually) `filterSpecsByStatus`'s existing inline checks — same `draft`/`failed` predicate, now named and reused rather than duplicated as a bare comparison.
- `validatePhaseDesignReadiness` now builds a `[]string` of `"path: reason"` per blocking spec and joins with `\n` into a single `fmt.Errorf`; returns `nil` unchanged when nothing blocks. No caller changes needed — `main.go`'s `fmt.Fprintf(os.Stderr, "Phase design readiness failed: %v\n", err)` already prints the full (now multi-line) message.
- Left `getPhaseSpecs` (main.go's separate phase-output-layer resolver) and `specDesignReady`/`design.go` untouched — confirmed via `grep` that neither needed changes; the bug was entirely in `getPhaseDesignGateSpecs`'s input filter.

**Decisions made:**
- See Design Decisions above: scoping and aggregate reporting both adopted; per-layer exemption explicitly deferred as a separate follow-up rather than bundled into this fix.
- No help-text or `framework/PHASES.md` changes were needed — the existing doc wording ("A phase is incomplete when any required design is missing...") doesn't claim a specific scope and remains accurate after the fix.

**Blockers encountered:**
- None.

**Follow-up identified:**
- A new task should decide whether smaqit needs a per-layer/per-project design-gate exemption mechanism (e.g., a project declaring "Coverage specs never carry a design pair" as smaqit's own convention allows, and having the gate honor that) — see Design Decisions for the concrete case this task's fix does not resolve.

**Verification:**
- `go build ./...`, `go vet ./...`, and `go test ./...` all pass in `installer/` (full suite, 20.4s) after `make prepare` regenerated the gitignored embedded-asset staging required to build.
- New/updated tests: `TestGetPhaseDesignGateSpecsIncludesConsumedUpstreamLayers` (updated to use draft-status fixtures so the existing layer-scope assertion still holds under the new filter), `TestGetPhaseDesignGateSpecsScopesToCurrentCycle` (new — the task's own regression scenario: legacy implemented/deployed/deprecated specs excluded, failed/draft specs included, scoped set passes the gate cleanly), `TestValidatePhaseDesignReadinessAggregatesAllFailures` (new — two blocking specs both appear in one error, a ready spec does not).

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/spec.go` | Modify — `getPhaseDesignGateSpecs`, `validatePhaseDesignReadiness` |
| `installer/spec_test.go` | Modify — regression coverage |
| `installer/design.go` | Reference only — `specDesignReady` itself is not the bug; the *input set* fed to it is |

## Notes

Found and root-caused live in a downstream project during a `smaqit.feature-new` cycle (their tasks 055/061-065 and, concurrently and independently, tasks 054/056-060), same calendar day, by two separate sessions hitting the identical bug from different features. Also referenced there as a documented project-level workaround pattern (`smaqit-framework-gap-plan-phase-develop-scope` / `smaqit-framework-gap-plan-phase-validate-scope` in that project's session memory) until this upstream fix lands. One session in that project also independently proposed an agent-side bypass for the `deploy`-phase case; the project's real user explicitly rejected it in favor of authoring genuine design pairs — worth noting as a signal that users want this gate *enforced correctly*, not routed around, which is exactly why the scoping (not the enforcement) is the right thing to fix here.
