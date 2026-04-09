# Session: Spec Templates Complete

**Date:** 2025-12-18

## Summary

Completed the specification template layer. Finalized terminology refinements (consistency → coherence), built remaining spec templates (Stack, Infrastructure, Coverage), and added Specification Coverage as a core principle.

## Actions Taken

### Terminology Refinement: Coherence

Replaced "consistency" with "coherence" throughout framework:
- **Rationale:** "Consistency" implies sameness; "coherence" better describes compatibility for consolidation
- **Pattern established:** "Upstream Specifications (for traceability and coherence)"
- Updated all framework files and spec agents

### New Core Principle: Specification Coverage

Added to SMAQIT.md:
> Every requirement MUST be verified through traceable test coverage.

Coverage layer traces requirements through all upstream specs to ensure nothing is missed.

### Specification Templates Built

**Task 019: Stack Template** — Complete
- References with `Enables` pattern (technology enables behaviors)
- Technology Stack tables: Languages, Frameworks, Libraries, Build Tools
- Inline rationale column for technology choices
- Constraints and Acceptance Criteria (`STK-[CONCEPT]-[NNN]`)

**Task 020: Infrastructure Template** — Complete  
- References to all Phase 1 specs (Business, Functional, Stack)
- Compute Resources with Service Topology subsection
- Networking, Scaling, Observability, Secrets Management sections
- Acceptance Criteria (`INF-[CONCEPT]-[NNN]`)

**Task 021: Coverage Template** — Complete
- References ALL upstream layers (unique to Coverage)
- Verification Requirements table for user input (SLAs, security, test env)
- Coverage Map: Requirement ID → Test Case ID → Expected Outcome
- Test Definitions with full Gherkin organized by type (Integration, E2E, Performance, Security, Acceptance)
- Untestable Criteria: reference-only pattern with verification decision
- Coverage Summary: pre-execution metrics table with formula

### Design Decisions

1. **Full Gherkin** — Coverage specs contain complete Gherkin scenarios, not summaries
2. **Pre-execution metrics** — Coverage Summary calculated at spec time, not post-execution
3. **Untestable reference pattern** — Reference upstream criteria by ID with link, don't re-list full text

## Files Modified

- `templates/specs/stack.template.md` — Created
- `templates/specs/infrastructure.template.md` — Created
- `templates/specs/coverage.template.md` — Created
- `framework/SMAQIT.md` — Added Specification Coverage principle, coherence terminology
- `framework/LAYERS.md` — Expanded Layer Independence, coherence terminology
- `framework/AGENTS.md` — Coherence terminology
- `framework/PHASES.md` — Coherence terminology
- `framework/ARTIFACTS.md` — Coherence terminology
- `agents/smaqit.*.agent.md` — All 5 spec agents updated with coherence terminology
- `docs/tasks/PLANNING.md` — Tasks 019-021 marked complete

## Template Layer Status

All 5 specification templates now complete:

| Layer | Template | Key Structure |
|-------|----------|---------------|
| Business | ✓ | Actors, Use Cases, Success Metrics |
| Functional | ✓ | User Flow, Data Model, API Contract, States |
| Stack | ✓ | Technology tables with inline rationale |
| Infrastructure | ✓ | Compute, Networking, Scaling, Observability |
| Coverage | ✓ | Coverage Map, Gherkin by test type, Metrics |

## Next Steps

- Task 001: Create smaq commands file
- Task 012: Create cross-platform Go installer build system
- Task 014: Define iterative development using smaqit
- Task 015: Refine installation approach
