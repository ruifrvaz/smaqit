---
name: smaqit.coverage
description: Create coverage layer specifications from verification requirements
agent: smaqit.coverage
tools: ["read", "edit", "search"]
---

# Coverage Specification Prompt

You are creating coverage layer specifications for the smaqit framework. This layer ensures **everything is verified**—all requirements from upstream layers are testable, traceable, and mapped to verification tests.

## Framework Context

Review these framework files to understand the Coverage layer requirements:
- [Core Principles](../framework/SMAQIT.md)
- [Coverage Layer Definition](../framework/LAYERS.md#coverage--whats-verified)
- [Template Structure](../framework/TEMPLATES.md)
- [Artifact Rules](../framework/ARTIFACTS.md)

## User Input

Collect verification and testing requirements:

${input:coverageRequirements:Specify test scope, performance benchmarks, security requirements, test environment details, and integration points}

## Instructions

Once you have collected the coverage requirements:

1. Read ALL upstream specifications (business, functional, stack, infrastructure) for complete traceability
2. Read the coverage specification template from `../templates/specs/coverage.template.md`
3. Invoke the Coverage Agent (smaqit.coverage) to transform requirements into structured verification specifications
4. The agent will create specification files in `specs/coverage/` following the template
5. Each specification MUST include:
   - References to all upstream layer specs
   - Coverage map tracing: Requirement ID → Test Case ID → Expected Outcome
   - Complete test definitions in Gherkin format (Integration, E2E, Performance, Security, Acceptance)
   - Untestable criteria explicitly flagged with verification decision
   - Coverage summary with spec coverage percentage calculation
   - Acceptance criteria with IDs in format `COV-[CONCEPT]-[NNN]`

## Success Criteria

Coverage specifications are complete when:
- Every testable requirement from upstream specs is mapped to a test case
- All test definitions are written in complete Gherkin format
- Untestable criteria are identified with justification
- Spec coverage percentage is calculated
- No new requirements are added (all tests trace to existing upstream requirements)
