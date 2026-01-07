# Task 052: Fix Deployment Agent CLI Directive (Preventive)

**Status:** Completed (2026-01-05)  
**Priority:** High  
**Created:** 2026-01-05  
**Completed with:** Task 051  
**Related:** Task 048 (E2E Testing), Tasks 049, 051

## Problem

Deployment agent likely has same weak directive phrasing as Development (Task 049) and Validation (Task 051) agents. Applying preventive fix before issue manifests in testing.

**Context:** E2E testing did not cover Phase 2 (Deploy/Infrastructure), so this issue was not observed directly. However, pattern consistency suggests Deployment agent has same directive weakness.

**Root Cause:** Same as Tasks 049 and 051—directive phrasing is likely instructional rather than imperative.

**Impact (if not fixed):**
- Deployment agent risks deploying wrong specs or missing failed/draft specs
- Undermines CLI as single source of truth for spec state
- Same violations as Issues 4 and 6

## Objective

Preventively update Deployment agent directive to mandate explicit CLI command execution as first step, ensuring pattern consistency across all implementation agents.

## Acceptance Criteria

- [x] Verified current `agents/smaqit.deployment.agent.md` directive phrasing
- [x] Updated directive from instructional to imperative (confirmed needed)
- [x] Added output requirement that deployment report MUST document CLI command execution
- [x] Agent directive explicitly states command must be "first action"
- [x] Agent directive specifies "process ONLY the specs returned" by CLI
- [x] Verified directive change with manual review

## Implementation Plan

1. **Check current directive** (`agents/smaqit.deployment.agent.md`):
   - Locate spec filtering directive
   - Confirm if weak phrasing exists (likely similar to Development/Validation)

2. **Update agent directive** (if needed):
   - **From:** "Determine which specs to process using `smaqit plan --phase=deploy`"
   - **To:** "Execute `smaqit plan --phase=deploy` as the first action and process ONLY the specs returned"

3. **Add output requirement** (Output section):
   - Add: "Deployment report MUST document the output of `smaqit plan --phase=deploy` command"

4. **Verify consistency:**
   - Ensure template (`templates/agents/implementation-agent.template.md`) reflects this pattern
   - All three implementation agents should have identical CLI query pattern

## Files to Modify

- `agents/smaqit.deployment.agent.md` (directive + output requirements, if needed)
- `templates/agents/implementation-agent.template.md` (if not updated in Tasks 049/051)

## Testing

**Manual verification:**
1. Read current directive
2. Confirm if change is needed
3. If changed, verify imperative phrasing is unambiguous
4. Confirm output requirement is clear

**Optional agent test:**
1. Run Deployment agent in test project with Infrastructure specs
2. Verify agent executes `smaqit plan --phase=deploy`
3. Verify report includes command output

## Estimated Effort

30 minutes (verification + update if needed)

## Dependencies

None (can be implemented independently, though logically follows Tasks 049 and 051)

## Blocks

None directly, but strengthens v0.5.0 release quality

## Related Tasks

- Task 049: Fix Development Agent CLI Directive
- Task 051: Fix Validation Agent CLI Directive
- Task 048: E2E Agent Workflow Testing (discovered pattern in Development/Validation)

## Notes

**Preventive fix rationale:** Rather than waiting for Phase 2 testing to discover same issue in Deployment agent, applying fix now ensures pattern consistency across all implementation agents.

**Pattern established:** This task completes the pattern: ALL three implementation agents (Development, Deployment, Validation) MUST execute `smaqit plan --phase=[PHASE]` as first action.

**Testing gap:** Phase 2 was not covered in E2E testing (Task 048), so Deployment agent behavior not directly validated. Future testing should include full Phase 2 workflow.

## Completion Summary

**Completed:** 2026-01-05 as part of Task 051 work

**Confirmed:** Deployment agent had same weak directive phrasing as Development and Validation agents.

See Task 051 completion summary for full details of changes across all three implementation agents.
