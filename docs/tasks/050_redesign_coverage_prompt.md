# Task 050: Redesign Coverage Prompt

**Status:** new  
**Priority:** High (Release Blocker)  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 5

## Problem

Coverage prompt (`.github/prompts/smaqit.coverage.prompt.md`) asks users to specify requirements (performance benchmarks, security requirements, test scope) that should already exist in upstream specs. This creates logical contradiction with Coverage agent directive: "MUST NOT add requirements not present in upstream specs."

**Root Cause:** Coverage layer design ambiguity between:
- **Interpretation 1 (correct):** Coverage is pure traceability mapping—no user input needed beyond tooling preferences
- **Interpretation 2 (incorrect):** Coverage is hybrid—adds verification-specific requirements not in upstream specs

**Evidence:**
- `framework/LAYERS.md` states Coverage must "Reference every acceptance criterion from upstream specs by ID" and "MUST NOT add requirements not present in upstream specs"
- Coverage agent directives align with Interpretation 1 (pure mapping)
- Coverage prompt structure aligns with Interpretation 2 (adds requirements)
- E2E testing proved Coverage agent CAN work with minimal prompt (tooling/thresholds only), generating 652-line spec with 100% traceability from 7 upstream specs

**Impact:**
- Creates requirement duplication/conflict risk (prompt says one thing, upstream specs say another)
- Violates Coverage layer purpose as "meta-spec" that maps existing requirements to tests
- Forces users to duplicate information already specified in Business/Functional/Infrastructure layers
- Undermines traceability—unclear if tests verify prompt requirements or spec requirements

## Objective

Redesign Coverage prompt to focus ONLY on verification preferences (test environment, tooling, thresholds), removing sections that ask for requirements already present in upstream specs.

## Acceptance Criteria

- [ ] Removed "Performance Benchmarks" section (requirements should be in Business/Infrastructure specs)
- [ ] Removed "Security Requirements" section (requirements should be in Functional/Infrastructure specs)
- [ ] Removed "Integration Points" section (requirements should be in upstream specs)
- [ ] Kept "Test Environment" section (tooling/platform preferences)
- [ ] Kept "Acceptance Thresholds" section (coverage percentage goals)
- [ ] Updated prompt description to clarify Coverage derives verification strategy from upstream specs
- [ ] Added guidance that prompt is optional—agent can work with minimal or empty input
- [ ] Updated `agents/smaqit.coverage.agent.md` to emphasize deriving from upstream specs
- [ ] Updated `framework/LAYERS.md` Coverage section to clarify prompt provides preferences, not requirements

## Implementation Plan

1. **Update Coverage prompt** (`.github/prompts/smaqit.coverage.prompt.md`):
   - Remove sections: Performance Benchmarks, Security Requirements, Integration Points
   - Keep sections: Test Environment, Acceptance Thresholds
   - Add description: "This prompt captures verification preferences. Coverage agent derives test requirements from upstream specs automatically."
   - Add guidance: "Minimal input is acceptable—agent can generate comprehensive coverage with just tooling preferences."

2. **Update Coverage agent** (`agents/smaqit.coverage.agent.md`):
   - Add directive: "Use prompt ONLY for verification strategy preferences (tooling, environment, thresholds). All requirements come from upstream specs."
   - Strengthen: "Ignore prompt content that duplicates upstream specs. Trace every requirement from upstream acceptance criteria by ID."

3. **Update framework** (`framework/LAYERS.md` Coverage section):
   - Add: "Coverage prompt provides verification preferences (tooling, environment), NOT requirements. All requirements come from upstream specs."

4. **Update template** (`templates/prompts/specification-prompt.template.md` if applicable):
   - Ensure Coverage prompt template aligns with pure traceability mapping purpose

## Files to Modify

- `.github/prompts/smaqit.coverage.prompt.md` (remove requirement sections)
- `prompts/smaqit.coverage.prompt.md` (source file)
- `agents/smaqit.coverage.agent.md` (strengthen directive)
- `framework/LAYERS.md` (clarify Coverage prompt purpose)
- `templates/prompts/specification-prompt.template.md` (optional, if pattern exists)

## Testing

**Manual verification:**
1. Read updated prompt
2. Confirm only verification preferences are requested
3. Confirm guidance about minimal input is clear

**Optional agent test:**
1. Fill Coverage prompt with minimal input (tooling only)
2. Run Coverage agent
3. Verify agent generates comprehensive spec from upstream specs
4. Verify no requirement duplication or conflicts

## Estimated Effort

1 hour

## Dependencies

None (can be implemented independently)

## Blocks

- v0.5.0 release (this is a release blocker)

## Related Tasks

- Task 048: E2E Agent Workflow Testing (discovered this issue)

## Notes

**Validation from E2E testing:** Coverage agent successfully generated 652-line spec with 100% traceability (92/92 testable requirements mapped) from minimal prompt input (tooling and thresholds only). This proves Coverage CAN work as pure traceability mapping layer.

**Framework alignment:** This fix aligns Coverage layer with its stated purpose: "Enumerate every acceptance criterion and map it to a verification test." Coverage is a meta-layer that validates completeness, not a requirements-adding layer.

**User experience improvement:** Simplifying prompt reduces cognitive load—users don't need to re-specify requirements they already documented in Business/Functional layers.

**Correct workflow:**
```
# User fills minimal Coverage prompt (optional)
Test Environment: unittest, local execution
Acceptance Thresholds: 100% of testable criteria must have test cases

# Coverage agent executes
1. Scan ALL upstream specs (specs/business/, specs/functional/, specs/stack/, specs/infrastructure/)
2. Extract ALL acceptance criteria with IDs (BUS-*, FUN-*, STK-*, INF-*)
3. For each criterion, define test case: COV-[ID] → Test → Expected Outcome
4. Calculate coverage: (mapped criteria / total testable criteria) × 100%
5. Flag untestable criteria with justification
6. Output: Coverage spec is comprehensive test plan proving 100% traceability
```
