---
name: smaqit.validate
description: Run the validation phase - create coverage specs, then validate deployed system
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Validate Phase Prompt

This phase orchestrates Coverage specification agent, then invokes the Validation implementation agent to verify the deployed system.

## Pre-Run Validation

Before starting, verify `.github/prompts/smaqit.coverage.prompt.md` has content beyond template structure.

**If prompt is empty:** Halt and provide natural language guidance about test scenarios, validation criteria, and acceptance thresholds.

## Orchestration Workflow

### Step 1: Coverage Specifications

Invoke `@smaqit.coverage` agent to produce coverage layer specifications in `specs/coverage/`.

Agent performs:
1. Read ALL upstream specifications (business, functional, stack, infrastructure)
2. Enumerate all acceptance criteria by ID
3. Produce complete test definitions in Gherkin format
4. Map: Requirement ID → Test Case → Expected Outcome
5. Flag untestable criteria with justification
6. Calculate spec coverage percentage

**On failure:** Stop and report error. Coverage specs required before validation.

### Step 2: Execute Validation Tests

Invoke `@smaqit.validation` agent to execute tests against deployed system.

Agent performs:
1. Execute tests against deployed system in target environment
2. Collect pass/fail results for each test case
3. Calculate actual spec coverage percentage
4. Identify unverified requirements
5. Produce validation report with detailed results

**On failure:** Test failures do NOT trigger automatic retry. Validation agent reports failures, and human decides next action:
- Return to Develop phase (code or spec issue)
- Return to Deploy phase (environment issue)
- Investigate further
- Accept with known issues

## Completion Criteria

Phase complete when:
- [ ] Coverage specifications produced with all testable criteria mapped
- [ ] Tests executed against deployed system
- [ ] Validation report generated
- [ ] Spec coverage percentage calculated
- [ ] Untestable criteria documented with justification

## Success Criteria

The Validate phase is complete when:
- Coverage specifications are produced with all testable criteria mapped
- Tests are executed against the deployed system
- Validation report is generated with:
  - Spec coverage percentage
  - Pass/fail status per requirement
  - Unverified requirements with justification
  - Failure details for any failed tests
