# Session: Framework Consolidation and Template Structure

**Date:** 2025-12-13

## Summary

Major framework restructuring to resolve naming ambiguity between agent types (specification/implementation) and artifact types. Created clear hierarchical taxonomy for kit development. Established template structure conventions.

## Actions Taken

### 1. Task 002: Specification Agent Template

- Created `templates/agents/specification-agent.template.md` with full structure
- Adopted `[PLACEHOLDER]` convention (brackets, SCREAMING_CASE)
- Template includes: Role, Framework Reference, Input, Output, Directives, Completion Criteria, Failure Handling

### 2. Template Folder Restructure

Moved from flat structure to explicit hierarchy:
```
templates/
├── specs/                    # Specification templates (5)
│   ├── business.template.md
│   ├── functional.template.md
│   ├── stack.template.md
│   ├── infrastructure.template.md
│   └── coverage.template.md
└── agents/                   # Agent templates (2)
    ├── specification-agent.template.md
    └── implementation-agent.template.md
```

### 3. Framework File Consolidation

Identified naming collision: "specification" meant both agent type and artifact type.

**Created:**
- `framework/TEMPLATES.md` — Template structure rules, placeholder conventions
- `framework/ARTIFACTS.md` — Merged specification + implementation artifact rules

**Deleted:**
- `framework/SPECIFICATIONS.md` (content moved to ARTIFACTS.md + TEMPLATES.md)
- `framework/IMPLEMENTATIONS.md` (content moved to ARTIFACTS.md)

**New framework structure:**
```
framework/
├── SMAQIT.md      # Index + core principles
├── LAYERS.md      # Layer definitions
├── PHASES.md      # Phase workflows
├── TEMPLATES.md   # Template structure rules
├── AGENTS.md      # Agent behaviors (actors)
└── ARTIFACTS.md   # Artifact rules (outputs)
```

### 4. Hierarchical Taxonomy

Established levels for kit architecture (contributor-facing only):
- **Level 0:** Framework foundation (SMAQIT, LAYERS, PHASES)
- **Level 1:** Templates (structure rules)
- **Level 2:** Agents & Artifacts (instances)
- **Level 3:** Application (output)

Moved taxonomy to README.md — not part of agent-consumable framework.

### 5. Copilot Instructions Cleanup

- Made less opinionated, now points to framework files as source of truth
- Split "When Editing Templates" into "When Editing Specification Templates" and "When Editing Agent Templates"
- Updated session.recap file list

## Key Decisions

- **[PLACEHOLDER] convention** — Brackets + SCREAMING_CASE for all template placeholders
- **"Level" not "Layer"** — Taxonomy uses "Level" to avoid collision with smaqit Layers (Business, Functional, etc.)
- **Taxonomy is human-only** — LLMs don't need level awareness; they follow explicit references in each agent
- **ARTIFACTS.md merges both artifact types** — Single file covers specification artifacts and implementation artifacts

## Files Modified

- `framework/SMAQIT.md` — Updated framework files table
- `framework/AGENTS.md` — Removed placeholder convention (moved to TEMPLATES.md)
- `framework/LAYERS.md` — Updated See Also
- `framework/PHASES.md` — Fixed broken PRINCIPLES.md refs, updated See Also
- `framework/TEMPLATES.md` — Created
- `framework/ARTIFACTS.md` — Created
- `framework/SPECIFICATIONS.md` — Deleted
- `framework/IMPLEMENTATIONS.md` — Deleted
- `.github/copilot-instructions.md` — Multiple updates
- `templates/agents/specification-agent.template.md` — Created/updated
- `templates/specs/*.template.md` — Moved from `templates/`
- `agents/*.agent.md` — Updated template paths (5 files)
- `installer/main.go` — Updated comments
- `README.md` — Added Kit Architecture section
- `docs/tasks/PLANNING.md` — Task 002 in progress
- `docs/tasks/002_create_specification_agent_template.md` — Criteria updated

## Next Steps

1. **Complete task 002** — Verify all acceptance criteria met
2. **Task 003** — Create implementation agent template (use same structure)
3. **Tasks 004-011** — Refactor all agents using new templates
4. **Consider** — Whether existing spec templates need restructuring to align with TEMPLATES.md rules
