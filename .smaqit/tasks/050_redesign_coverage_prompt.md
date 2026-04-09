# Task 050: Redesign Coverage Prompt

**Status:** Completed (2026-01-06)  
**Priority:** High (Release Blocker)  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 5

## Problem

Coverage prompt design was ambiguous about whether Coverage layer adds requirements or purely maps upstream requirements to tests. First attempt (Session 027, 2026-01-05) made Coverage "preferences only" which was too restrictive.

**Corrected Understanding:** Coverage has TWO input sources:
1. **Test requirements from prompt:** HOW to test (test scope, environment, integration points, thresholds)
2. **Acceptance criteria from upstream specs:** WHAT to test (all criteria from Business, Functional, Stack, Infrastructure)

**Root Cause:** Confusion between "test requirements" (legitimate Coverage concerns) and "functional requirements" (what system does—comes from upstream).

## Objective

Clarify Coverage layer's dual input model: test requirements from prompt + upstream acceptance criteria to verify.

## Acceptance Criteria

- [x] Clarified Coverage Input in framework/LAYERS.md: "test requirements" from prompt, "upstream acceptance criteria to verify" from context
- [x] Updated Coverage MUST directives to use precise terminology ("upstream acceptance criteria" not "functional requirements")
- [x] Kept 4 test requirement sections in Coverage prompt: Test Scope, Test Environment, Integration Points, Acceptance Thresholds
- [x] Removed functional requirement sections that belong in upstream layers: Performance Benchmarks (→ Business/Infrastructure), Security Requirements (→ Functional/Infrastructure)
- [x] Updated agents/smaqit.coverage.agent.md with dual-input model: prompt for test requirements, upstream specs for acceptance criteria
- [x] Updated agent MUST directive: "Scan ALL upstream specs and map every upstream acceptance criterion by ID to a test case"
- [x] Added Prompt File section to agent Input (handles HTML comments, validation, prompt reading)
- [x] Disambiguated Validation prompt: Renamed "Test Scope" → "Execution Scope" to distinguish from Coverage "Test Scope"
- [x] Updated templates/prompts/ to reflect Coverage and Validation changes
- [x] Updated framework/PROMPTS.md and framework/ARTIFACTS.md with clarified terminology

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

**First Attempt (Session 027, 2026-01-05):** Made Coverage "preferences only" which was too restrictive. Coverage DOES have requirements—test requirements (how to test), not functional requirements (what to test).

**Corrected Approach (Session 031, 2026-01-06):** Coverage has dual inputs: test requirements from prompt + upstream acceptance criteria to verify.

**Validation from E2E testing:** Coverage agent successfully generated 652-line spec with 100% traceability (92/92 testable requirements mapped), proving dual-input model works.

**Framework alignment:** Coverage layer has legitimate test requirements (test scope, environment, integration points, thresholds) separate from functional requirements (which come from upstream specs).

**Correct workflow:**
```
# User fills Coverage prompt with test requirements
Test Scope: Integration testing, E2E testing
Test Environment: pytest on GitHub Actions
Integration Points: None (standalone app)
Acceptance Thresholds: 100% of testable criteria must have test cases

# Coverage agent executes
1. Read test requirements from prompt (HOW to test)
2. Scan ALL upstream specs for acceptance criteria (WHAT to test)
3. Map each upstream criterion to test case using prompt's test requirements
4. Format: Upstream Requirement ID → Test Case → Expected Outcome
5. Calculate coverage: (mapped criteria / total testable criteria) × 100%
6. Flag untestable criteria with justification
7. Output: Coverage spec proving 100% traceability
```

## Completion Summary

**Completed:** 2026-01-06 (Corrected approach after Session 027's overly restrictive attempt)

**Key Insight:** Coverage layer has test requirements (how to test) but not functional requirements (what to test).

**Files Modified:** 7 (3 framework + 2 templates + 2 agents/prompts)

**Testing:** ✅ Build successful, terminology consistent, all levels aligned
