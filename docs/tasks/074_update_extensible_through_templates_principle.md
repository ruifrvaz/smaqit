# Task 074: Update "Extensible Through Templates" Principle Context

**Status:** new  
**Priority:** Low  
**Created:** 2026-01-23

## Context

The "Extensible Through Templates" principle in `framework/SMAQIT.md` (lines 61-77) discusses agent extensibility through templates. The current text references "Core agents" vs "Auxiliary agents" in a context that assumed orchestrator-driven workflows:

> Core agents participate in phase workflows: they consume prompts, generate specifications, execute implementations, track state. Auxiliary agents operate orthogonally: they retrieve knowledge, answer questions, perform supporting tasks without altering project state.

With orchestrator removed (Task 072) and implementation agents becoming phase orchestrators (Task 073), the agent taxonomy needs slight adjustment.

## Problem

The principle's language implies orchestrator-centric workflow where "core agents participate in phase workflows" suggests they're coordinated externally. With the phase-orchestrator model, implementation agents ARE the coordinators.

## Goal

Update "Extensible Through Templates" section to reflect phase-orchestrator architecture while preserving the core philosophical principle about template-based extension.

## Scope

### Current Text to Revise

```markdown
Core agents participate in phase workflows: they consume prompts, generate specifications, execute implementations, track state. Auxiliary agents operate orthogonally: they retrieve knowledge, answer questions, perform supporting tasks without altering project state.

Both inherit from shared foundations. Both respect boundary principles. The distinction lies not in quality or importance but in relationship to the workflow. Core agents drive phases forward. Auxiliary agents support developers throughout.
```

### Proposed Revision

Clarify agent taxonomy in phase-orchestrator model:

**Phase Agents** (3):
- Development, Deployment, Validation
- Orchestrate phase execution (invoke spec agents + implement)
- Track state, update frontmatter
- Primary workflow drivers

**Specification Agents** (5):
- Business, Functional, Stack, Infrastructure, Coverage
- Generate specifications from prompt files
- Invoked by phase agents or directly by users
- State-altering but phase-coordinated

**Auxiliary Agents** (variable):
- Q&A, documentation, analysis
- Support tasks without state changes
- Orthogonal to phase workflow
- Template-extensible like core agents

### Files to Update

- `framework/SMAQIT.md` — "Extensible Through Templates" section (lines 61-77)

## Implementation Steps

1. Read current "Extensible Through Templates" section fully
2. Identify orchestrator-centric language
3. Revise to reflect phase-orchestrator taxonomy
4. Preserve core principle: templates enable extension through shared foundations
5. Ensure clarity: phase agents coordinate, spec agents generate, auxiliary agents support

## Acceptance Criteria

- [ ] Section reflects phase-orchestrator architecture (not orchestrator-driven)
- [ ] Agent taxonomy clear: phase agents (3), spec agents (5), auxiliary agents (extensible)
- [ ] Core principle preserved: templates constrain and enable extension simultaneously
- [ ] No references to orchestrator pattern
- [ ] Language is philosophically consistent with framework principles

## Dependencies

- Task 072: Remove Orchestrator Agent Pattern (provides context)
- Task 073: Implementation Agents as Phase Orchestrators (defines new architecture)

## Follow-up Tasks

- None (standalone documentation refinement)

## Notes

- This is a documentation polish task, not an architectural change
- Low priority because the principle itself remains philosophically valid
- Main change: update agent taxonomy to match phase-orchestrator reality
- Can be deferred if higher-priority implementation work takes precedence
