---
name: smaqit.validate
description: Run the validation phase - create coverage specs, then validate deployed system
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Validation Phase Prompt

You are orchestrating the **Validate phase** of the smaqit framework. This phase verifies that the deployed system satisfies all specification requirements by producing coverage specifications that map all requirements to tests, then executing those tests against the running system.

## Framework Context

Review these framework files to understand the Validate phase workflow:
- [Core Principles](../framework/SMAQIT.md)
- [Validate Phase Definition](../framework/PHASES.md#validate--verify-spec-compliance)
- [Coverage Layer Definition](../framework/LAYERS.md#coverage--whats-verified)

## User Input

Collect verification and testing requirements:

${input:coverageRequirements:Specify test scope, performance benchmarks, security requirements, test environment details, and integration points}

${input:targetEnvironment:Specify the target environment where the system is deployed (must match Deploy phase)}

## Orchestration Workflow

Once all inputs are collected, execute the following sequence:

### Step 1: Coverage Specifications

Run subagent **smaqit.coverage** with the coverage requirements to produce coverage layer specifications in `specs/coverage/`.

The agent will:
1. Read ALL upstream specifications (business, functional, stack, infrastructure)
2. Enumerate all acceptance criteria by ID
3. Produce complete test definitions in Gherkin format
4. Map: Requirement ID → Test Case → Expected Outcome
5. Flag untestable criteria with justification
6. Calculate spec coverage percentage

**On failure:** Stop and report the error. Coverage specs are required before validation.

### Step 2: Execute Validation Tests

Run subagent **smaqit.validation** to execute tests against the deployed system.

The validation agent will:
1. Execute tests against the deployed system in the target environment
2. Collect pass/fail results for each test case
3. Calculate actual spec coverage percentage
4. Identify unverified requirements
5. Produce validation report with detailed results

**On failure:** Test failures do NOT trigger automatic retry. The validation agent reports failures, and human decides next action:
- Return to Develop phase (code or spec issue)
- Return to Deploy phase (environment issue)
- Investigate further
- Accept with known issues

## Success Criteria

The Validate phase is complete when:
- Coverage specifications are produced with all testable criteria mapped
- Tests are executed against the deployed system
- Validation report is generated with:
  - Spec coverage percentage
  - Pass/fail status per requirement
  - Unverified requirements with justification
  - Failure details for any failed tests
