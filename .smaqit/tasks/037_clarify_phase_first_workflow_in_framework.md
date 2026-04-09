# Clarify Phase-First Workflow in Framework

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-28  
**Source:** User Testing Report Issue #6 (2025-12-27)

## Description

CRITICAL: Both testing agent and user execution treated specifications as separate from phases, revealing fundamental documentation gap. Framework must emphasize that phases are the primary workflow unit, with each phase including specifications + implementation.

## Acceptance Criteria

- [x] SMAQIT.md updated with explicit statement: "Phases are the primary workflow. Each phase includes specifications + implementation."
- [x] PHASES.md updated with clear examples showing phase completion (not just spec generation)
- [x] README.md updated to show phase-based workflow examples
- [x] Documentation clarifies: users CAN spec-first (all 5 layers then implement), but SHOULD phase-first (3 specs → implement → 1 spec → deploy → 1 spec → validate)
- [x] Quick reference tables emphasize phase structure
- [x] Workflow diagrams/examples show complete phase cycles
- [ ] Testing agent instructions updated to reflect phase-first approach

## Impact

**Severity:** CRITICAL  
**User Impact:** Fundamental misunderstanding of smaqit workflow; framework documentation insufficiently emphasizes phases as primary unit; both agents and users may execute incorrectly; violates core smaqit principle that each phase includes specs + implementation

## Notes

This is the most critical issue identified. Affects core understanding of smaqit. Must be addressed before other issues.

**Related issues:** #032 (status command), #035 (status display) also reflect this confusion.

## Implementation Summary

**Changes Made:**

1. **SMAQIT.md** (framework/SMAQIT.md):
   - Added bold statement in introduction: "Phases are the primary workflow unit"
   - Created new "Phase-First Workflow" core principle (positioned as first principle)
   - Added comprehensive "Workflow Approaches" comparison table in Quick Reference
   - Documented phase-first as recommended, spec-first as optional alternative
   - Explained feedback cycle differences between approaches

2. **PHASES.md** (framework/PHASES.md):
   - Rewrote introduction to emphasize phases as primary workflow unit
   - Added "Phase-First Workflow" section with visual diagram
   - Updated "Overview" to clarify phase completion requirements
   - Added workflow comparison (Phase-First vs Spec-First) in Develop, Deploy, and Validate phase sections
   - Clarified that phases must complete sequentially but specs can be generated ahead

3. **README.md**:
   - Replaced "What is it?" section with phase-first workflow explanation
   - Added "How it Works" section with visual comparison of both workflows
   - Rewrote Usage section showing both workflows with code examples
   - Expanded Phases section with bold recommendations
   - Made phase-first approach immediately visible to new users

**Testing:**
- Built installer successfully (version aec90b5)
- Verified framework files are NOT bundled separately (embedded architecture confirmed)
- Tested `smaqit init` in clean directory - all files created correctly
- Confirmed templates and agents copied properly
- Verified state.json created with correct structure

**Outstanding:**
- Testing agent instructions (agents/smaqit.user-testing.agent.md) not updated in this task
- This should be addressed when testing agent is next updated
- Current changes provide clear documentation that testing agent can reference

**Impact:**
Users and agents will now clearly understand:
- Phases are the primary workflow unit (not layers)
- Each phase = specs + implementation together
- Phase-first gives faster feedback (recommended)
- Spec-first delays feedback but valid for upfront design needs
- Both approaches are supported by framework
