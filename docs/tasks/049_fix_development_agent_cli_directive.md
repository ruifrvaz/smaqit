# Task 049: Fix Development Agent CLI Directive

**Status:** new  
**Priority:** High (Release Blocker)  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 4

## Problem

Development agent processed incremental implementation without executing `smaqit plan --phase=develop` to determine which specs require implementation. Agent relied on implicit understanding rather than explicit CLI state query.

**Evidence:** Development phase report (`development-phase-report-2026-01-04-luigi.md`) contains no `smaqit plan` or `smaqit status` commands in execution log.

**Root Cause:** Current directive is instructional ("Determine which specs to process using `smaqit plan --phase=develop`") rather than imperative ("Execute `smaqit plan --phase=develop` as first action").

**Impact:**
- Violates established phase workflow directive
- Undermines CLI as single source of truth for spec state
- Risk of processing wrong specs or missing specs in complex scenarios
- Could lead to re-implementing already-implemented specs or skipping failed specs

## Objective

Update Development agent directive to mandate explicit CLI command execution as first step, ensuring CLI is authoritative source for determining which specs require processing.

## Acceptance Criteria

- [ ] Updated `agents/smaqit.development.agent.md` directive from instructional to imperative phrasing
- [ ] Added output requirement that development report MUST document CLI command execution
- [ ] Agent directive explicitly states command must be "first action"
- [ ] Agent directive specifies "process ONLY the specs returned" by CLI
- [ ] Verified directive change with test execution (optional)

## Implementation Plan

1. **Update agent directive** (`agents/smaqit.development.agent.md` line ~49):
   - **From:** "Determine which specs to process using `smaqit plan --phase=develop`"
   - **To:** "Execute `smaqit plan --phase=develop` as the first action and process ONLY the specs returned"

2. **Add output requirement** (Output section):
   - Add: "Development report MUST document the output of `smaqit plan --phase=develop` command"

3. **Verify consistency:**
   - Check template (`templates/agents/implementation-agent.template.md`) for similar patterns
   - If template uses same weak phrasing, update it as well

## Files to Modify

- `agents/smaqit.development.agent.md` (directive + output requirements)
- `templates/agents/implementation-agent.template.md` (optional, if pattern exists)

## Testing

**Manual verification:**
1. Read updated directive
2. Confirm imperative phrasing is unambiguous
3. Confirm output requirement is clear

**Optional agent test:**
1. Run Development agent in test project
2. Verify agent executes `smaqit plan --phase=develop`
3. Verify report includes command output

## Estimated Effort

1 hour

## Dependencies

None (can be implemented independently)

## Blocks

- v0.5.0 release (this is a release blocker)

## Related Tasks

- Task 051: Fix Validation Agent CLI Directive (same issue, different agent)
- Task 052: Fix Deployment Agent CLI Directive (preventive)
- Task 048: E2E Agent Workflow Testing (discovered this issue)

## Notes

**Why this matters:** In complex projects with many specs across multiple layers, implicit understanding will fail. CLI must be programmatically queried to ensure correct spec filtering (draft/failed vs implemented/deployed/validated).

**Alternative considered:** Adding workflow section with numbered steps. Rejected because directive-based approach (MUST/MUST NOT) is established smaqit pattern.

**Success criteria:** Agent directive is so clear that no interpretation ambiguity remains. "Execute X as first action" leaves no room for "I satisfied the spirit without executing the command."
