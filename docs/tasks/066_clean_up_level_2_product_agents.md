# Task 066: Clean Up Level 2 Product Agents

**Status:** new  
**Priority:** High  
**Created:** 2026-01-13

## Context

Task B001.3 created Agent-L2 (Agent Compiler) to compile Level 1 directives into Level 2 product agents. Task 065 will extract implementation details from L1 templates, creating a compilation workload for Agent-L2.

Current L2 product agents in `agents/`:
- `smaqit.business.agent.md` — Business specification agent
- `smaqit.functional.agent.md` — Functional specification agent
- `smaqit.stack.agent.md` — Stack specification agent
- `smaqit.infrastructure.agent.md` — Infrastructure specification agent
- `smaqit.coverage.agent.md` — Coverage specification agent
- `smaqit.development.agent.md` — Development implementation agent
- `smaqit.deployment.agent.md` — Deployment implementation agent
- `smaqit.validation.agent.md` — Validation implementation agent
- `smaqit.orchestrator.agent.md` — Orchestrator agent

These agents may contain:
- Implementation details compiled from L1 (correctly placed)
- L1 directives not yet extracted to templates (creating duplication)
- L0 principles that belong in framework files
- Inconsistent compilation across agent types

## Problem

Without systematic L2 cleanup:
- Implementation details extracted from L1 may duplicate existing L2 content
- Some L2 agents may contain L1-level directives or L0 principles
- Compilation from L1→L2 may be inconsistent across agents
- Changes to L1 directives won't propagate correctly to L2

## Goal

Clean up Level 2 product agents to ensure they:
1. Contain all implementation details extracted from L1 (no missing compilation)
2. Contain only executable behaviors, not L1 directives or L0 principles
3. Compile L1 directives consistently across agent types
4. Eliminate duplication with L1 templates

## Scope

### In Scope
- Compile extracted L1 implementation details into appropriate L2 agents
- Remove L1 directives from L2 agents (move to L1 templates)
- Remove L0 principles from L2 agents (move to L0 framework files)
- Verify compilation consistency across specification agents
- Verify compilation consistency across implementation agents
- Document compilation rationale
- Validate agent completeness (all L1 directives compiled)

### Out of Scope
- L0 principle cleanup (that's Task 064)
- L1 template compilation (that's Task 065)
- Creating new L1 directives (only compiling existing)
- Agent behavior changes (focus on compilation correctness)

## Input

From Task 065:
- Extracted implementation details from all L1 templates
- Target agent mappings for each implementation detail
- Duplication identification between L1 extractions and existing L2

## Acceptance Criteria

- [ ] All implementation details extracted from L1 (Task 065) compiled into appropriate L2 agents
- [ ] Zero L1 directives in L2 agents (all moved to L1 templates)
- [ ] Zero L0 principles in L2 agents (all moved to L0 framework files)
- [ ] Consistent compilation pattern across specification agents
- [ ] Consistent compilation pattern across implementation agents
- [ ] No implementation duplication within single agent
- [ ] No implementation duplication across agents (unless intentional shared behavior)
- [ ] Compilation rationale documented
- [ ] Agent completeness verified (no missing L1 directive compilation)
- [ ] All agents follow Level 2 structure (executable behaviors only)
- [ ] Completion documented in session history

## Dependencies

- Task 065 (Clean Up Level 1 Templates) — BLOCKS this task (need extraction list)
- Task B001.3 (Agent-L2 created) — ✅ Completed

## Blocks

None (this completes the Level Up cleanup sequence)

## Compilation Pattern

**L1 Directive (in template):**
```markdown
### Foundation Reference

**Agents MUST:**
- Reference existing specs rather than duplicate information
- Use Foundation Reference section for same-layer shared requirements
```

**L2 Implementation (in agent):**
```markdown
## Duplication Prevention Workflow

Before creating new specification:

1. **Search existing specs**
   ```bash
   grep -r "[CONCEPT]" specs/[layer]/
   ```

2. **If concept found:**
   - Read existing spec to understand scope
   - Determine if new requirements extend existing concept
   - If yes: Update existing spec (add to Scope, Requirements, Acceptance Criteria)
   - If no: Create new spec with Foundation Reference to related spec

3. **If concept not found:**
   - Create new spec
   - Check for shared requirements in other specs
   - Add Foundation Reference section if dependencies exist

4. **Validate references:**
   - All Foundation Reference paths resolve
   - Referenced concepts exist in target specs
   - No circular references
```

## Notes

### Expected L2 Contamination Sources

1. **L1 directives in agents:**
   - MUST/SHOULD/MUST NOT statements without implementation
   - Abstract rules without concrete steps
   - Template structure definitions

2. **L0 principles in agents:**
   - Philosophical statements about "why"
   - Conceptual explanations without procedures
   - Design rationale

3. **Duplicate implementations:**
   - Same workflow in multiple agents
   - Similar procedures with slight variations
   - Shared behaviors not extracted to common section

### Cleanup Strategy

Use Agent-L2 systematically:
1. Audit existing L2 agents for contamination (L1 directives, L0 principles)
2. Extract L1 directives back to templates (with documentation for Task 065)
3. Extract L0 principles back to framework (with documentation for Task 064)
4. Compile L1 implementation details from Task 065
5. Consolidate duplicate implementations
6. Verify completeness against L1
7. Document compilation rationale

### Agent Categories

**Specification Agents (5):**
- Business, Functional, Stack, Infrastructure, Coverage
- Should share common compilation patterns (duplication check, reference creation, validation)

**Implementation Agents (3):**
- Development, Deployment, Validation
- Should share common compilation patterns (spec consolidation, frontmatter updates, reporting)

**Orchestrator Agent (1):**
- Unique role coordinating other agents
- May have distinct compilation pattern
