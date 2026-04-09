# Validation Agent Should Update Upstream Spec Frontmatter

**Status:** Complete  
**Created:** 2026-01-11  
**Completed:** 2026-01-11  
**Priority:** Medium  
**Related:** Issue 7 from Task 059 (E2E Regression Testing), same pattern as Issue 11/Task 061

## Description

Validation agent updates Coverage spec frontmatter to `status: validated` but does NOT update upstream specs (Business, Functional, Stack, Infrastructure) that it processes to reflect validation lifecycle progression.

**Current Behavior:**
- Validation agent references upstream specs (Business, Functional, Stack, Infrastructure, Coverage)
- Validation agent updates Coverage spec to `status: validated`
- Upstream specs remain at their previous status (`implemented`, `deployed`)
- Status lifecycle is incomplete - specs don't reflect that they've been validated

**Expected Behavior:**
- Validation agent references upstream specs for validation mapping
- Updates ALL referenced specs to `status: validated` (Business, Functional, Stack, Infrastructure, Coverage)
- Status lifecycle complete - all specs reflect actual validation state
- Per principle: Implementation agents update all upstream specs THAT THEY REFERENCE

## Acceptance Criteria

- [ ] Update `agents/smaqit.validation.agent.md` directive: "Update frontmatter of ALL referenced specs to `status: validated`"
- [ ] Directive specifies: Business, Functional, Stack, Infrastructure, Coverage specs all updated
- [ ] Validation: Re-run validation phase test
- [ ] Validation: Verify Business spec frontmatter shows `status: validated`, `validated: [ISO8601_TIMESTAMP]`
- [ ] Validation: Verify Functional spec frontmatter shows `status: validated`, `validated: [ISO8601_TIMESTAMP]`
- [ ] Validation: Verify Stack spec frontmatter shows `status: validated`, `validated: [ISO8601_TIMESTAMP]`
- [ ] Validation: Verify Infrastructure spec frontmatter shows `status: validated`, `validated: [ISO8601_TIMESTAMP]`
- [ ] Validation: Verify Coverage spec frontmatter shows `status: validated`, `validated: [ISO8601_TIMESTAMP]`
- [ ] Update PHASES.md Validate phase completion criteria to include "All referenced specs updated to `status: validated`"

## Notes

**Severity:** Medium - Status lifecycle incomplete but doesn't block core workflow. CLI aggregation still shows phase completion correctly through Coverage spec status.

**Principle:** Implementation agents update all upstream specs THAT THEY REFERENCE. Validation agent references Business, Functional, Stack, Infrastructure specs for validation mapping and Coverage specs from `smaqit plan`, therefore should update all of them.

**Same Pattern as Task 061 (Issue 11):** Deployment agent exhibits identical behavior - only updates Infrastructure spec, not upstream specs (Business, Functional, Stack) that it references.

**Design Decision:** Implementation agents cascade status updates to all specs they reference, not just the specs returned by `smaqit plan`.

**Rationale:**
- Validation agent reads and references requirements from all upstream specs
- If agent references a spec (reads it for validation mapping), that spec has been validated
- Status reflects reality: "This spec has been validated against deployed system"
- Enables accurate status reporting: `smaqit status` shows which specs reached which phase

**Clarification:**
- `smaqit plan --phase=validate` returns Coverage specs only (specs to directly implement)
- Validation agent ALSO reads upstream specs (Business, Functional, Stack, Infrastructure) for validation mapping
- All referenced specs (upstream + returned by plan) should be updated to `status: validated`

**Impact of Fix:**
- Complete status lifecycle tracking across all layers
- Clear audit trail of which specs reached validation phase
- Accurate CLI status reporting without intelligence/interpretation layer

**Affected Files:**
- `agents/smaqit.validation.agent.md` (primary)
- `framework/PHASES.md` (completion criteria update)

**Cross-Reference:** Task 061 requires identical fix for Deployment agent
