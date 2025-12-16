# Task: Build Business Spec Template

**ID**: 017
**Status**: Completed

## Context

Build the specification template for the Business layer. This template defines the structure that the `smaqit.business` agent uses when producing business specifications.

**Source of truth**: Framework files only (LAYERS.md, TEMPLATES.md, ARTIFACTS.md)

## Acceptance Criteria

- [x] Template has required sections: Title, Scope, Acceptance Criteria
- [x] Business layer has no References section (it's the entry point)
- [x] Layer-specific content matches LAYERS.md Business MUST rules:
  - Actors and their goals
  - Measurable success metrics
  - Preconditions and postconditions
  - Main and alternative flows in business terms
- [x] Uses `[PLACEHOLDER]` convention for customizable values
- [x] Requirement ID format documented: `BUS-[CONCEPT]-[NNN]`
- [x] No empty sections — placeholder guidance for each
- [x] Testability requirements noted for acceptance criteria

## Layer Rules (from LAYERS.md)

**Business specs MUST:**
- Identify all actors and their goals
- Define measurable success metrics for each use case
- Include preconditions and postconditions
- Describe main and alternative flows in business terms

**Business specs MUST NOT:**
- Mention specific technologies, frameworks, or libraries
- Include implementation details or technical solutions
- Define data structures or API contracts
- Reference deployment or infrastructure concerns

## Notes

- Business is the entry point for user input — no upstream specs to reference
- Focus on "Why?" — value and intent behind the work
- One file per use case or business flow
