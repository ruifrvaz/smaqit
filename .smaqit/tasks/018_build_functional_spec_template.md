# Task: Build Functional Spec Template

**ID**: 018
**Status**: Completed

## Context

Build the specification template for the Functional layer. This template defines the structure that the `smaqit.functional` agent uses when producing functional specifications.

**Source of truth**: Framework files only (LAYERS.md, TEMPLATES.md, ARTIFACTS.md)

## Acceptance Criteria

- [x] Template has required sections: Title, References, Scope, Acceptance Criteria
- [x] References section links to upstream Business specs
- [x] Layer-specific content matches LAYERS.md Functional MUST rules:
  - User flows that implement business use cases
  - Data models with attributes and relationships
  - API contracts (inputs, outputs, error conditions)
  - State transitions where applicable
  - References to originating business specs
- [x] Uses `[PLACEHOLDER]` convention for customizable values
- [x] Requirement ID format documented: `FUN-[CONCEPT]-[NNN]`
- [x] No empty sections — placeholder guidance for each
- [x] Testability requirements noted for acceptance criteria

## Layer Rules (from LAYERS.md)

**Functional specs MUST:**
- Define user flows that implement business use cases
- Specify data models with attributes and relationships
- Define API contracts (inputs, outputs, error conditions)
- Include state transitions where applicable
- Reference originating business specs

**Functional specs MUST NOT:**
- Specify technology choices (languages, frameworks, databases)
- Include deployment or infrastructure concerns
- Define performance benchmarks (those belong in Infrastructure)
- Prescribe implementation patterns

## Notes

- Functional layer answers "What?" — behaviors and contracts
- Upstream: Business specs
- Downstream: Stack layer, Coverage layer
- One file per user flow, API contract, or data model
