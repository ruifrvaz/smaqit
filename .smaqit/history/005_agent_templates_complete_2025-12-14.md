# 005 - Agent Templates Complete

**Date**: 2025-12-14

## Summary

Completed both agent template tasks (002 and 003). Created specification agent template and rewrote implementation agent template. Resolved framework naming ambiguity between agent types and artifact types through restructuring.

## Actions Taken

### Task 002: Specification Agent Template
- Created `templates/agents/specification-agent.template.md`
- Established required sections: Role, Framework Reference, Input, Output, Directives (MUST/MUST NOT/SHOULD), Layer-Specific Rules, Completion Criteria, Failure Handling
- Defined placeholders: `[LAYER]`, `[LAYER_NAME]`, `[LAYER_PREFIX]`, `[UPSTREAM_SPEC_PATHS]`, `[LAYER_SPECIFIC_RULES]`

### Task 003: Implementation Agent Template
- Rewrote `templates/agents/implementation-agent.template.md` to align with specification template
- Added missing sections: Framework Reference, Directives, Failure Handling
- Added Cross-Layer Consolidation under MUST directive
- Defined placeholders: `[PHASE]`, `[PHASE_NAME]`, `[AGENT_NAME]`, `[UPSTREAM_SPEC_PATHS]`, `[PHASE_SPECIFIC_RULES]`

### Framework Restructuring
- Created `framework/TEMPLATES.md` — Level 1 template structure rules
- Created `framework/ARTIFACTS.md` — Level 2 artifact rules (merged SPECIFICATIONS.md + IMPLEMENTATIONS.md)
- Deleted `framework/SPECIFICATIONS.md` and `framework/IMPLEMENTATIONS.md`
- Reorganized templates folder: `templates/specs/` and `templates/agents/`

### Conventions Established
- Placeholder format: `[PLACEHOLDER]` (brackets, SCREAMING_CASE)
- Contributor taxonomy added to README.md (Level 0-3 hierarchy)
- Changed taxonomy terminology from "Layer" to "Level" to avoid collision with smaqit Layers

## Decisions Made

1. **Naming ambiguity resolution**: "Specification" and "implementation" refer to both agent types (who produces) and artifact types (what is produced). Resolved by separating TEMPLATES.md (templates) from ARTIFACTS.md (outputs).

2. **Agent alignment criterion removed**: The criterion "Existing agents align with template structure" was removed from tasks 002/003. This work is covered by tasks 004-011 (individual agent refactoring).

3. **No code fences in agent files**: Agent definitions use raw YAML frontmatter, not wrapped in code fences. Clarified in TEMPLATES.md.

4. **Taxonomy is human-facing**: The Level 0-3 taxonomy is for kit developers, not consumed by agents. Placed in README.md only.

## Files Modified

- `templates/agents/specification-agent.template.md` — Created
- `templates/agents/implementation-agent.template.md` — Rewritten
- `framework/TEMPLATES.md` — Created
- `framework/ARTIFACTS.md` — Created
- `framework/SPECIFICATIONS.md` — Deleted
- `framework/IMPLEMENTATIONS.md` — Deleted
- `framework/SMAQIT.md` — Updated cross-references
- `framework/AGENTS.md` — Added placeholder convention
- `README.md` — Added contributor taxonomy
- `.github/copilot-instructions.md` — Cleaned up, added placeholder convention
- `docs/tasks/002_*.md` — Marked completed
- `docs/tasks/003_*.md` — Marked completed
- `docs/tasks/PLANNING.md` — Updated statuses

## Next Steps

- Tasks 004-011: Refactor existing agents to align with new templates
- Task 001: Create smaq commands file (still new)
- Consider whether all 8 agent refactoring tasks should be combined or done individually
