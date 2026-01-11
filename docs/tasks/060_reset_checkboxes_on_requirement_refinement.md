# Reset Checkboxes on Requirement Refinement

**Status:** Not Started  
**Created:** 2026-01-11  
**Priority:** Medium  
**Related:** Issue 10 from Task 059 (E2E Regression Testing)

## Description

When specification agents modify existing acceptance criteria to expand scope during incremental additions, they should reset the checkbox state from `[x]` to `[ ]` to indicate revalidation is needed for the expanded scope.

**Current Behavior:**
- Functional agent updated FUN-OUTPUT-006 and FUN-OUTPUT-013 to include Luigi character scope
- Checkboxes remained `[x]` (checked) even though requirements expanded
- This misleads developers - criteria appear satisfied but actually need revalidation

**Expected Behavior:**
- When agent modifies acceptance criterion text, checkbox resets to `[ ]`
- Development agent later updates checkbox to `[x]` when expanded criterion is satisfied
- Provides accurate per-requirement status during incremental development

## Acceptance Criteria

- [ ] Business agent directive: Reset checkbox to `[ ]` when modifying existing acceptance criteria text
- [ ] Functional agent directive: Reset checkbox to `[ ]` when modifying existing acceptance criteria text
- [ ] Stack agent directive: Reset checkbox to `[ ]` when modifying existing acceptance criteria text
- [ ] Validation: Re-run Luigi incremental addition test case
- [ ] Validation: Verify modified requirements show `[ ]` after spec update
- [ ] Validation: Verify Development agent later updates to `[x]` after implementation

## Notes

**Severity:** Medium (not High) because frontmatter `status: draft` still indicates global revalidation needed, reducing impact of misleading checkbox state.

**Impact:** Per-requirement accuracy is lost during intermediate state. Developers checking individual criteria status could be misled about what actually needs revalidation.

**Affected Agents:**
- `agents/smaqit.business.agent.md`
- `agents/smaqit.functional.agent.md`
- `agents/smaqit.stack.agent.md`

**Not Affected:**
- Infrastructure agent (Phase 2 - no prior implementation to modify)
- Coverage agent (reads only, doesn't modify upstream specs)
