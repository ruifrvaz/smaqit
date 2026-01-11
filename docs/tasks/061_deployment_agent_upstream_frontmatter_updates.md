# Deployment Agent Should Update Upstream Spec Frontmatter

**Status:** Not Started  
**Created:** 2026-01-11  
**Priority:** Medium  
**Related:** Issue 11 from Task 059 (E2E Regression Testing), follows same pattern as Issue 7

## Description

Deployment agent updates Infrastructure spec frontmatter to `status: deployed` but does NOT update upstream specs (Business, Functional, Stack) to reflect deployment lifecycle progression.

**Current Behavior:**
- Deployment agent references upstream specs (Business, Functional, Stack) for coherence validation
- Deployment agent updates only Infrastructure spec to `status: deployed`
- Upstream specs (Business, Functional, Stack) remain at `status: implemented`
- Status lifecycle is incomplete - specs don't reflect that they've been deployed

**Expected Behavior:**
- Deployment agent references upstream specs for coherence validation
- Updates ALL referenced specs to `status: deployed` (Business, Functional, Stack, Infrastructure)
- Status lifecycle complete - all specs reflect actual deployment state
- Per principle: Implementation agents update all upstream specs THAT THEY REFERENCE

## Acceptance Criteria

- [ ] Update `agents/smaqit.deployment.agent.md` directive: "Update frontmatter of ALL referenced specs to `status: deployed`"
- [ ] Directive specifies: Business, Functional, Stack, Infrastructure specs all updated
- [ ] Validation: Re-run deployment phase test
- [ ] Validation: Verify Business spec frontmatter shows `status: deployed`, `deployed: [ISO8601_TIMESTAMP]`
- [ ] Validation: Verify Functional spec frontmatter shows `status: deployed`, `deployed: [ISO8601_TIMESTAMP]`
- [ ] Validation: Verify Stack spec frontmatter shows `status: deployed`, `deployed: [ISO8601_TIMESTAMP]`
- [ ] Validation: Verify Infrastructure spec frontmatter shows `status: deployed`, `deployed: [ISO8601_TIMESTAMP]`
- [ ] Update PHASES.md Deploy phase completion criteria to include "All referenced specs updated to `status: deployed`"

## Notes
**Severity:** Medium - Status lifecycle incomplete but doesn't block core workflow. CLI aggregation still shows phase completion correctly through Infrastructure spec status.

**Principle:** Implementation agents update all upstream specs THAT THEY REFERENCE. Deployment agent references Business, Functional, Stack specs for coherence validation and Infrastructure specs from `smaqit plan`, therefore should update all of them.

**Same Pattern as Task 063 (Issue 7):** Validation agent exhibits identical behavior - only updates Coverage spec, not upstream specs (Business, Functional, Stack, Infrastructure) that it references.

**Design Decision:** Implementation agents cascade status updates to all specs they reference, not just the specs returned by `smaqit plan`.

**Rationale:**
- Deployment agent reads and references upstream specs for coherence validation
- If agent references a spec (reads it for deployment context), that spec has been deployed
- Status reflects reality: "This spec has been deployed to target environment"
- Enables accurate status reporting: `smaqit status` shows which specs reached which phase

**Note on Session 034:** Session 034 removed "Update existing specs" directives referring to CONTENT modification (adding requirements, changing acceptance criteria). Status frontmatter updates are METADATA tracking, not content modification - different concern.

**Clarification:**
- `smaqit plan --phase=deploy` returns Infrastructure specs only (specs to directly implement)
- Deployment agent ALSO reads upstream specs (Business, Functional, Stack) for coherence validation
- All referenced specs (upstream + returned by plan) should be updated to `status: deployed`

**Impact of Fix:**
- Complete status lifecycle tracking across all layers
- Clear audit trail of which specs reached deployment phase
- Accurate CLI status reporting without intelligence/interpretation layer

**Affected Files:**
- `agents/smaqit.deployment.agent.md` (primary)
- `framework/PHASES.md` (completion criteria update)

**Cross-Reference:** Task 063 requires identical fix for Validation agent
- `agents/smaqit.validation.agent.md` (Issue 7, same pattern)
