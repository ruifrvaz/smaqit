# Clarify Phase-First Workflow in Framework

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #6 (2025-12-27)

## Description

CRITICAL: Both testing agent and user execution treated specifications as separate from phases, revealing fundamental documentation gap. Framework must emphasize that phases are the primary workflow unit, with each phase including specifications + implementation.

## Acceptance Criteria

- [ ] SMAQIT.md updated with explicit statement: "Phases are the primary workflow. Each phase includes specifications + implementation."
- [ ] PHASES.md updated with clear examples showing phase completion (not just spec generation)
- [ ] README.md updated to show phase-based workflow examples
- [ ] Documentation clarifies: users CAN spec-first (all 5 layers then implement), but SHOULD phase-first (3 specs → implement → 1 spec → deploy → 1 spec → validate)
- [ ] Quick reference tables emphasize phase structure
- [ ] Workflow diagrams/examples show complete phase cycles
- [ ] Testing agent instructions updated to reflect phase-first approach

## Impact

**Severity:** CRITICAL  
**User Impact:** Fundamental misunderstanding of smaqit workflow; framework documentation insufficiently emphasizes phases as primary unit; both agents and users may execute incorrectly; violates core smaqit principle that each phase includes specs + implementation

## Notes

This is the most critical issue identified. Affects core understanding of smaqit. Must be addressed before other issues.

**Related issues:** #032 (status command), #035 (status display) also reflect this confusion.
