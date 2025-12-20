---
name: smaqit.business
description: Create business layer specifications from stakeholder requirements
agent: smaqit.business
tools: ["read", "edit", "search"]
---

# Business Specification Prompt

You are creating business layer specifications for the smaqit framework. This is the first layer that defines **why** the system exists and **what success looks like** from a business perspective.

## Framework Context

Review these framework files to understand the Business layer requirements:
- [Core Principles](../framework/SMAQIT.md)
- [Business Layer Definition](../framework/LAYERS.md#business--why)
- [Template Structure](../framework/TEMPLATES.md)
- [Artifact Rules](../framework/ARTIFACTS.md)

## User Input

Collect stakeholder requirements:

${input:businessRequirements:Describe the use cases, actors, business goals, and success metrics}

## Instructions

Once you have collected the business requirements:

1. Read the business specification template from `../templates/specs/business.template.md`
2. Invoke the Business Agent (smaqit.business) to transform the requirements into structured specifications
3. The agent will create specification files in `specs/business/` following the template
4. Each specification MUST include:
   - Identified actors and their goals
   - Use cases with preconditions, postconditions, main flow, and alternative flows
   - Measurable success metrics
   - Acceptance criteria with IDs in format `BUS-[CONCEPT]-[NNN]`

## Success Criteria

Business specifications are complete when:
- All use cases are documented with clear flows
- All actors and their goals are identified
- Success metrics are measurable and observable
- All acceptance criteria have unique IDs and are testable
- No technology or implementation details are mentioned (keep business-level only)
