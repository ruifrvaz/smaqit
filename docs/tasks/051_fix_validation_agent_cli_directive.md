# Task 051: Fix Validation Agent CLI Directive

**Status:** new  
**Priority:** High (Release Blocker)  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 6

## Problem

Validation agent processed validation without executing `smaqit plan --phase=validate` to determine which specs require validation. Agent relied on implicit understanding rather than explicit CLI state query.

**Evidence:** Validation report (`validation-phase-report-2026-01-04.md`) contains no `smaqit plan` or `smaqit status` commands in execution log.

**Root Cause:** Same as Task 049 (Development agent)—directive is instructional rather than imperative.

**Impact:**
- Same violations as Issue 4 (Development agent)
- Validation agent risks validating wrong specs or missing failed/draft specs
- Undermines CLI as single source of truth for spec state

## Objective

Update Validation agent directive to mandate explicit CLI command execution as first step, ensuring CLI is authoritative source for determining which specs require validation.

## Acceptance Criteria

- [ ] Updated `agents/smaqit.validation.agent.md` directive from instructional to imperative phrasing
- [ ] Added output requirement that validation report MUST document CLI command execution
- [ ] Agent directive explicitly states command must be "first action"
- [ ] Agent directive specifies "process ONLY the specs returned" by CLI
- [ ] Verified directive change with test execution (optional)

## Implementation Plan

1. **Update agent directive** (`agents/smaqit.validation.agent.md`):
   - **From:** "Determine which specs to process using `smaqit plan --phase=validate`"
   - **To:** "Execute `smaqit plan --phase=validate` as the first action and process ONLY the specs returned"

2. **Add output requirement** (Output section):
   - Add: "Validation report MUST document the output of `smaqit plan --phase=validate` command"

3. **Verify consistency:**
   - Check template (`templates/agents/implementation-agent.template.md`) for similar patterns
   - Apply same fix as Task 049 if not already done

## Files to Modify

- `agents/smaqit.validation.agent.md` (directive + output requirements)
- `templates/agents/implementation-agent.template.md` (if not updated in Task 049)

## Testing

**Manual verification:**
1. Read updated directive
2. Confirm imperative phrasing is unambiguous
3. Confirm output requirement is clear

**Optional agent test:**
1. Run Validation agent in test project
2. Verify agent executes `smaqit plan --phase=validate`
3. Verify report includes command output

## Estimated Effort

1 hour (same as Task 049)

## Dependencies

None (can be implemented independently, but logically follows same pattern as Task 049)

## Blocks

- v0.5.0 release (this is a release blocker)

## Related Tasks

- Task 049: Fix Development Agent CLI Directive (same issue, different agent)
- Task 052: Fix Deployment Agent CLI Directive (preventive)
- Task 048: E2E Agent Workflow Testing (discovered this issue)

## Notes

**Same root cause as Task 049:** Both Development and Validation agents have weak directive phrasing that invites interpretation ambiguity. Applying same fix pattern.

**Pattern consistency:** This fix establishes pattern that ALL implementation agents MUST execute `smaqit plan --phase=[PHASE]` as first action. Task 052 applies this preventively to Deployment agent.
