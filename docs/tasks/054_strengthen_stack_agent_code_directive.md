# Task 054: Strengthen Stack Agent Code Directive

**Status:** new  
**Priority:** Medium  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 2

## Problem

Stack agent generated specification (`specs/stack/cli-stack.md`) includes "Architecture Notes" section with actual Python code examples showing argparse implementation patterns (20+ lines of code).

**Impact:**
- Violates framework principle that specs define WHAT, not HOW
- Blurs boundary between specification and implementation
- May bias Development agent toward specific implementation patterns
- Undermines spec as technology-agnostic contract

**Framework Violation:** `framework/ARTIFACTS.md` states "Stack specs MUST NOT: Include application code or configurations"

**Root Cause:** Stack agent directive may not be explicit enough about code prohibition, or Stack template may have sections that invite code inclusion.

## Objective

Strengthen Stack agent directive to explicitly prohibit code examples, implementation patterns, and architecture code blocks in Stack specifications.

## Acceptance Criteria

- [ ] Added explicit directive to `agents/smaqit.stack.agent.md`: "MUST NOT include code examples, implementation patterns, or architecture code blocks"
- [ ] Reviewed `templates/specs/stack.template.md` for sections that invite code inclusion
- [ ] Removed or modified inviting sections (e.g., "Architecture Notes") from template
- [ ] Added guidance on what Stack specs SHOULD contain (technology choice + rationale only)
- [ ] Verified directive change with test execution (optional)

## Implementation Plan

1. **Update Stack agent directive** (`agents/smaqit.stack.agent.md`):
   - Add to MUST NOT section: "Include code examples, implementation patterns, or architecture code blocks"
   - Add to SHOULD section: "Specify technology choice with rationale, but never prescribe implementation"
   - Example guidance: "Specify 'argparse for CLI parsing' with rationale, not argparse configuration code"

2. **Review Stack template** (`templates/specs/stack.template.md`):
   - Locate any sections that invite code (e.g., "Architecture Notes", "Implementation Guidance", "Code Examples")
   - Remove or modify these sections
   - Ensure template focuses on: Technology choice, Version requirements, Rationale, Constraints

3. **Add positive examples** (optional, in AGENTS.MD or wiki):
   - **Good:** "Python 3.8+ with argparse for CLI argument parsing. Rationale: Built-in, no dependencies, sufficient for simple character selection."
   - **Bad:** "Python 3.8+ with argparse. Example: `parser = argparse.ArgumentParser()...`"

4. **Framework clarification**:
   - Verify `framework/LAYERS.md` Stack section prohibits code
   - If not explicit, add clarification

## Files to Modify

- `agents/smaqit.stack.agent.md` (strengthen directive)
- `templates/specs/stack.template.md` (remove inviting sections)
- `framework/LAYERS.md` (verify Stack layer prohibits code and update if not)

## Testing

**Manual verification:**
1. Read updated directive
2. Confirm code prohibition is explicit and unambiguous
3. Review template for inviting sections

**Optional agent test:**
1. Run Stack agent with same Mario/Luigi requirements
2. Verify generated spec contains NO code examples
3. Verify spec contains technology choice + rationale only

## Estimated Effort

1 hour

## Dependencies

None (can be implemented independently)

## Blocks

None directly, but improves framework principle enforcement for v0.5.0

## Related Tasks

- Task 048: E2E Agent Workflow Testing (discovered this issue)
- Task 055: Formalize Single Source of Truth Principle (related framework gap)

## Notes

**Why this matters:** Stack specs are contracts defining technology choices. Including implementation code crosses the boundary from specification to implementation, violating separation of concerns.

**Development agent resilience:** During E2E testing, Development agent did NOT blindly follow code patterns from Stack spec—agent implemented independently using appropriate patterns. This shows agents can resist code pollution, but specs should still maintain purity.

**User impact:** Code in Stack specs creates false expectation that implementation MUST follow specific patterns. This constrains Development agent unnecessarily and undermines spec as contract (WHAT) vs implementation (HOW).

**Alternative approach considered:** Could add validation tooling (`smaqit validate --specs`) to flag code blocks in Stack specs. Rejected as secondary—better to prevent at agent level first.
