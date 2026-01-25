# Task 065: Clean Up Level 1 Templates

**Status:** completed  
**Priority:** High  
**Created:** 2026-01-13  
**Completed:** 2026-01-25

## Completion Summary

Completed inadvertently through systematic template extraction work (Todos 1-6 from session 025). Instead of cleaning existing templates in place, we extracted all directive content to compilation files and refactored templates to pure placeholder structure.

**Approach taken:**
1. Created `specification.rules.md` with all specification-extension directives (Role, Input, Output content structures + MUST/MUST NOT/SHOULD + Compilation Guidance)
2. Created `implementation.rules.md` with all implementation-extension directives (10 content structures + MUST/MUST NOT/SHOULD + Compilation Guidance)
3. Refined `base.rules.md` to foundation directives only (removed workflow-specific to spec/impl.rules)
4. Refactored `specification-agent.template.md` to pure placeholders (78 lines, zero directive content)
5. Refactored `implementation-agent.template.md` to pure placeholders (74 lines, zero directive content)
6. Updated Agent-L2 with 4-way merge pattern documentation + SDK extensibility (base agents, spec agents, impl agents)

**Results achieved:**
- ✅ L1 templates are pure structure (no L2 implementation details)
- ✅ All directives consolidated in compilation files (zero duplication)
- ✅ Directive distribution systematic and documented (Compilation Guidance sections)
- ✅ Template completeness verified (information gap analysis performed)
- ✅ SDK extensibility preserved (base template + base.rules for Q&A/helper agents)
- ✅ 4-way merge hierarchy established (foundation → workflow extension → role-specific)

**Files modified:**
- `templates/agents/specification-agent.template.md` (pure placeholders)
- `templates/agents/implementation-agent.template.md` (pure placeholders)
- `templates/agents/compiled/base.rules.md` (refined to foundation only)
- `templates/agents/compiled/specification.rules.md` (created, 276 lines)
- `templates/agents/compiled/implementation.rules.md` (created, 432 lines)
- `.github/agents/smaqit.L2.agent.md` (updated with 3 compilation patterns)

## Context

Task B001.2 created Agent-L1 (Template Compiler) to compile Level 0 principles into Level 1 template directives. Task 064 will extract directives from contaminated L0 framework files, creating a compilation workload for Agent-L1.

Current L1 templates exist in:
- `templates/specs/` — 5 specification templates (business, functional, stack, infrastructure, coverage)
- `templates/agents/` — 2 agent templates (specification.template.agent.md, implementation.template.agent.md)
- `templates/prompts/` — 2 prompt templates (layer.template.prompt.md, phase.template.prompt.md)

These templates may contain:
- Directives already compiled from L0 (correctly placed)
- Directives not yet extracted from L0 (creating duplication)
- Implementation details that belong in L2 agents
- Inconsistent directive distribution across templates

## Problem

Without systematic L1 cleanup:
- Extracted L0 directives may duplicate existing L1 directives
- Some L1 templates may contain L2-level implementation details
- Directive distribution across templates may be inconsistent
- Changes to L0 principles won't propagate correctly to L1

## Goal

Clean up Level 1 templates to ensure they:
1. Contain all directives extracted from L0 (no missing compilation)
2. Contain only directives, not L2 implementation details
3. Distribute directives consistently across template types
4. Eliminate duplication between templates

## Scope

### In Scope
- Compile extracted L0 directives into appropriate L1 templates
- Remove L2 implementation details from L1 templates (extract for Task 066)
- Verify directive consistency across specification templates
- Verify directive consistency across agent templates
- Document directive distribution rationale
- Validate template completeness (all L0 principles compiled)

### Out of Scope
- L0 principle cleanup (that's Task 064)
- L2 agent compilation (that's Task 066)
- Creating new L0 principles (only compiling existing)
- Template structure changes (focus on directive content)

## Input

From Task 064:
- Extracted directives from all L0 framework files
- Target template mappings for each directive
- Duplication identification between L0 extractions and existing L1

## Acceptance Criteria

- [ ] All directives extracted from L0 (Task 064) compiled into appropriate L1 templates
- [ ] Zero L2 implementation details in L1 templates (all extracted and documented for Task 066)
- [ ] Consistent MUST/SHOULD/MUST NOT usage across all templates
- [ ] No directive duplication within single template
- [ ] No directive duplication across templates (unless intentional cross-cutting)
- [ ] All specification templates have consistent directive structure
- [ ] All agent templates have consistent directive structure
- [ ] Directive distribution rationale documented
- [ ] Template completeness verified (no missing L0 principle compilation)
- [ ] Completion documented in session history

## Dependencies

- Task 064 (Complete Level 0 Principle Cleanup) — BLOCKS this task (need extraction list)
- Task B001.2 (Agent-L1 created) — ✅ Completed

## Blocks

- Task 066 (Clean Up Level 2 Product Agents) — Cannot compile L1→L2 until L1 is clean

## Template Compilation Pattern

**L0 Principle:**
```markdown
### Single Source of Truth

Each piece of information exists in exactly one place. When information is needed 
in multiple contexts, reference the source rather than duplicate.
```

**L1 Directive (in specification template):**
```markdown
### Foundation Reference

**Agents MUST:**
- Reference existing specs rather than duplicate information
- Use Foundation Reference section for same-layer shared requirements
- Use Implements/Enables for upstream layer references
```

**L2 Implementation (in product agent):**
```markdown
## Duplication Check

Before creating new spec:
1. Search existing specs in `specs/[layer]/` for related concepts
2. If concept exists, update existing spec (add requirements, extend scope)
3. If concept is distinct, create new spec with references to related specs
```

## Notes

### Expected L1 Contamination Sources

1. **Implementation details in templates:**
   - File path specifications beyond structure
   - Command-line invocations
   - Specific tool usage (beyond "use X for Y")

2. **L2-level procedural steps:**
   - "Before X, do Y" workflows
   - Conditional logic ("If X, then Y")
   - Iteration strategies

3. **Duplicate directives:**
   - Same directive in multiple templates (may be intentional cross-cutting)
   - Reworded versions of same requirement

### Cleanup Strategy

Use Agent-L1 systematically:
1. Audit existing L1 templates for contamination
2. Extract L2 implementation details
3. Compile L0 directives from Task 064
4. Consolidate duplicate directives
5. Verify completeness against L0
6. Document distribution rationale
