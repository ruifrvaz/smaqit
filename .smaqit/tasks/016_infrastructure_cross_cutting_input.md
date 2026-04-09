# Task: Infrastructure Cross-Cutting Input

**ID**: 016
**Status**: completed

## Context

During review of the infrastructure agent refactoring, identified that Infrastructure layer should read all Phase 1 specs (Business, Functional, Stack) rather than only Stack specs. Infrastructure decisions depend on:

- Business specs: compliance requirements, availability SLAs
- Functional specs: API constraints, rate limits, data retention policies
- Stack specs: runtime requirements, technology choices
- User input: environment, provider, topology, resources, budget

This is a framework amendment based on real design analysis (per Amendment Protocol in SMAQIT.md).

## Acceptance Criteria

- [x] Add phase numbering convention to PHASES.md (Phase 1/2/3)
- [x] Update Infrastructure layer upstream in LAYERS.md to "Phase 1 specs"
- [x] Update dependency graph in LAYERS.md
- [x] Expand Phase 2 user input documentation in PHASES.md
- [x] Add "Fail-Fast on Inconsistency" principle to AGENTS.md
- [x] Update Infrastructure agent mapping in AGENTS.md
- [x] Update Infrastructure agent input in smaqit.infrastructure.agent.md

## Notes

- Infrastructure becomes cross-cutting like Coverage, but only for specs (no new artifact types)
- User provides topology/resource information as Phase 2 input
- Adds coherence validation as pre-condition (Fail-Fast on Inconsistency)
