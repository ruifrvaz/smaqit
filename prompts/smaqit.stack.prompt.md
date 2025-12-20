---
name: smaqit.stack
description: Create stack layer specifications from technology preferences
agent: smaqit.stack
tools: ["read", "edit", "search"]
---

# Stack Specification Prompt

You are creating stack layer specifications for the smaqit framework. This layer defines **with what** the system will be built—the languages, frameworks, libraries, and tools that can deliver the specified behaviors.

## Framework Context

Review these framework files to understand the Stack layer requirements:
- [Core Principles](../framework/SMAQIT.md)
- [Stack Layer Definition](../framework/LAYERS.md#stack--with-what)
- [Template Structure](../framework/TEMPLATES.md)
- [Artifact Rules](../framework/ARTIFACTS.md)

## User Input

Collect technology preferences and constraints:

${input:stackRequirements:Specify preferred languages, frameworks, libraries, build tools, or technology constraints}

## Instructions

Once you have collected the stack requirements:

1. Read existing business and functional specifications from `specs/business/` and `specs/functional/` for context
2. Read the stack specification template from `../templates/specs/stack.template.md`
3. Invoke the Stack Agent (smaqit.stack) to transform the preferences into structured specifications
4. The agent will create specification files in `specs/stack/` following the template
5. Each specification MUST include:
   - References to business and functional specs for traceability
   - Technology choices with rationale (why each technology was selected)
   - Language versions and framework versions
   - Libraries and their purposes
   - Build tools and development environment setup
   - Acceptance criteria with IDs in format `STK-[CONCEPT]-[NNN]`

## Success Criteria

Stack specifications are complete when:
- All technology choices are documented with clear rationale
- Versions are specified for languages and frameworks
- Technology stack is consistent with functional requirements
- All acceptance criteria have unique IDs and are testable
- No deployment or infrastructure concerns are included (that's the Infrastructure layer)
