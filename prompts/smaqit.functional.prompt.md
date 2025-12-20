---
name: smaqit.functional
description: Create functional layer specifications from user experience requirements
agent: smaqit.functional
tools: ["read", "edit", "search"]
---

# Functional Specification Prompt

You are creating functional layer specifications for the smaqit framework. This layer defines **what** the system must do—the behaviors, contracts, and data models required to fulfill business goals.

## Framework Context

Review these framework files to understand the Functional layer requirements:
- [Core Principles](../framework/SMAQIT.md)
- [Functional Layer Definition](../framework/LAYERS.md#functional--what)
- [Template Structure](../framework/TEMPLATES.md)
- [Artifact Rules](../framework/ARTIFACTS.md)

## User Input

Collect user experience requirements:

${input:functionalRequirements:Describe the user experience, behaviors, interactions, and data models needed}

## Instructions

Once you have collected the functional requirements:

1. Read existing business specifications from `specs/business/` for context and traceability
2. Read the functional specification template from `../templates/specs/functional.template.md`
3. Invoke the Functional Agent (smaqit.functional) to transform the requirements into structured specifications
4. The agent will create specification files in `specs/functional/` following the template
5. Each specification MUST include:
   - References to business specs (Implements for features, Enables for foundations)
   - User flows that implement business use cases
   - Data models with attributes and relationships
   - API contracts with inputs, outputs, and error conditions
   - State transitions where applicable
   - Acceptance criteria with IDs in format `FUN-[CONCEPT]-[NNN]`

## Success Criteria

Functional specifications are complete when:
- All business use cases have corresponding user flows
- Data models and API contracts are precisely defined
- References to business specs are valid and traceable
- All acceptance criteria have unique IDs and are testable
- No technology choices are specified (that's the Stack layer)
