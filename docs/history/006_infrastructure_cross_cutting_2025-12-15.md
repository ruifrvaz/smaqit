# 006 - Infrastructure Cross-Cutting Input

**Date**: 2025-12-15

## Summary

Major framework amendment: Infrastructure agent now reads all Phase 1 specs (Business, Functional, Stack) instead of only Stack specs. Added phase numbering convention, "Fail-Fast on Inconsistency" principle, and clarified Phase 1 deliverables.

## Key Decisions

### Infrastructure Dependencies

**Problem identified:** Infrastructure decisions require more than Stack specs:
- Business specs → compliance requirements, availability SLAs
- Functional specs → API constraints, rate limits, data retention
- Stack specs → runtime requirements, technology choices

**Options considered:**
1. Development agent produces build output manifest → rejected (creates meta-artifact, shifts source of truth)
2. Infrastructure agent compiles manifest from all sources → rejected (duplicates work, complex)
3. Infrastructure reads all Phase 1 specs + expanded user input → **adopted**

**Rationale:** Keeps specs as source of truth, no new artifact types, user provides topology/resource info as Phase 2 input.

### Phase Numbering

Introduced explicit phase numbering for clarity:
- Phase 1 = Develop (produces Business, Functional, Stack specs)
- Phase 2 = Deploy (produces Infrastructure specs)
- Phase 3 = Validate (produces Coverage specs)

"Phase 1 specs" now means Business + Functional + Stack.

### Fail-Fast on Inconsistency

New principle added to AGENTS.md: Agents must verify coherence across all inputs before producing output. Stops and reports when inputs contradict.

### Technology Neutrality

Added Content Guidelines to copilot-instructions.md: Framework documentation describes what kind of information is needed, not specific technologies. Avoids biasing agents toward particular vendors/architectures.

### Phase 1 Deliverables

Clarified that Phase 1 produces:
- **Code** — Source, tests, configurations, build files
- **README** — Build, test, and run instructions
- **Development report** — Build/test/run results

Running application during Phase 1 is verification, not a deliverable. Code is the artifact; it's proven runnable.

## Files Modified

- `framework/PHASES.md` — Phase numbering, expanded user input table, Phase 1 artifacts
- `framework/LAYERS.md` — Infrastructure upstream changed, dependency graph updated
- `framework/AGENTS.md` — Added Fail-Fast on Inconsistency principle, updated agent mapping
- `framework/ARTIFACTS.md` — Updated Develop Phase artifacts
- `agents/smaqit.infrastructure.agent.md` — Input section expanded to all Phase 1 specs
- `.github/copilot-instructions.md` — Added Technology Neutrality guideline
- `docs/tasks/016_infrastructure_cross_cutting_input.md` — Created and completed
- `docs/tasks/PLANNING.md` — Added task 016, fixed task 007 status

## Open Considerations

- Phase 2 user input categories are abstract (purpose-based, not example-based)
- Infrastructure agent has pre-condition: verify coherence before producing output
- README and development report now explicit Phase 1 requirements

## Next Steps

- Continue agent refactoring (tasks 005, 008-011 remain)
- Task 001: Create smaq commands file
- Task 014: Define iterative development using smaqit
