# Task 053: Fix Validation Frontmatter Updates

**Status:** new  
**Priority:** High (Release Blocker)  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 7

## Problem

After successful validation, spec frontmatter still shows `status: implemented` instead of `status: validated`. Validation agent did not update spec frontmatter to reflect validated state.

**Evidence:** All 7 Phase 1 specs retained `status: implemented` after successful validation run. Expected `status: validated` with `validated: [timestamp]` field added.

**Impact:**
- Violates stateful specification principle—specs should track progression through phases
- `smaqit status` command shows Phase 3 "✓ Complete" but spec files don't reflect validated state
- Loss of traceability—can't determine which specs have been validated by inspecting files
- Breaks incremental workflow—if validation is re-run, agent has no way to know specs were already validated
- Inconsistent with Development agent behavior (which correctly updates to `status: implemented`)

## Objective

Update Validation agent to modify spec frontmatter after successful validation, changing status to `validated` and adding `validated` timestamp field.

## Acceptance Criteria

- [ ] Updated `agents/smaqit.validation.agent.md` with explicit frontmatter update directive
- [ ] Agent directive specifies updating `status: validated`
- [ ] Agent directive specifies adding `validated: [ISO8601_TIMESTAMP]` field
- [ ] Added output requirement that all validated specs MUST have frontmatter updated
- [ ] Verified frontmatter update behavior with test execution (optional)

## Implementation Plan

1. **Add frontmatter update directive** (`agents/smaqit.validation.agent.md`):
   - Add to MUST directives: "Update all validated spec frontmatter to `status: validated` with `validated: [ISO8601_TIMESTAMP]` field"
   - Place in appropriate section (likely under "State Tracking" or "Completion Criteria")

2. **Add output requirement** (Output section or Completion Criteria):
   - Add: "All specs in validation scope MUST have frontmatter updated to reflect validated status"

3. **Verify consistency:**
   - Check Development agent for frontmatter update pattern (should already exist)
   - Ensure Validation agent follows same pattern

4. **Framework documentation** (optional):
   - Verify `framework/PHASES.md` Validate phase completion criteria mentions frontmatter updates
   - If missing, add to completion criteria checklist

## Files to Modify

- `agents/smaqit.validation.agent.md` (add directive + output requirement)
- `framework/PHASES.md` (optional, verify completion criteria mentions frontmatter)

## Testing

**Manual verification:**
1. Read updated directive
2. Confirm frontmatter update requirement is explicit
3. Confirm ISO8601 timestamp format specified

**Optional agent test:**
1. Run Validation agent in test project
2. After validation completes, inspect spec frontmatter
3. Verify `status: validated` present
4. Verify `validated: [timestamp]` field added
5. Verify timestamp is valid ISO8601 format

## Expected Frontmatter After Fix

**Before validation:**
```yaml
---
id: BUS-GREETING
status: implemented
created: 2026-01-04T10:00:00Z
implemented: 2026-01-04T16:03:00Z
prompt_version: main
---
```

**After validation:**
```yaml
---
id: BUS-GREETING
status: validated
created: 2026-01-04T10:00:00Z
implemented: 2026-01-04T16:03:00Z
validated: 2026-01-04T23:26:00Z
prompt_version: main
---
```

## Estimated Effort

30 minutes

## Dependencies

None (can be implemented independently)

## Blocks

- v0.5.0 release (this is a release blocker)

## Related Tasks

- Task 048: E2E Agent Workflow Testing (discovered this issue)
- Task 054: Add checkbox updates to Validation agent (related but separate concern)

## Notes

**Why this matters:** Stateful specifications (Task 014, Session 024) are a core smaqit principle. Specs must track lifecycle progression: `draft → implemented → deployed → validated`. Without frontmatter updates, validation history is lost and incremental workflows break.

**Pattern from Development agent:** Development agent correctly updates frontmatter to `status: implemented` with `implemented: [timestamp]`. Validation agent should follow identical pattern for consistency.

**CLI behavior:** `smaqit status` command scans spec frontmatter to determine phase status. Without frontmatter updates, CLI cannot accurately report validation state per-spec (though it can still aggregate based on Coverage spec status).

**Deployment agent consideration:** Deployment agent (Task 052) should also update frontmatter to `status: deployed` with `deployed: [timestamp]`. Check during Task 052 implementation.
