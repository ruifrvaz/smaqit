# Define Iterative Development Using smaqit

**Status:** Completed (2026-01-03)  
**Created:** 2025-12-13  
**Completed:** 2026-01-03

## Description

smaqit should embrace iterative development. New features should be backwards compatible, meaning specs need to be either modular or refactored upon feature introduction.

This task evolved into implementing **stateful specifications** to enable incremental development patterns.

## Implementation Summary

**Phases Completed (Session 2026-01-03):**

1. **Framework Updates** - Added "Stateful Specifications" principle and state tracking to 5 framework files
2. **Template Updates** - Added YAML frontmatter to all 5 specification templates
3. **Agent Updates** - Updated 8 agents with state tracking directives
4. **CLI Updates** - Extended `state.json` schema and status display in installer
5. **Wiki Documentation** - Created 2 new articles, updated 2 existing articles

**Key Deliverables:**
- Spec-level state tracking via YAML frontmatter (`status`, `created`, `prompt_version`, timestamps)
- Phase-level state aggregation in `.smaqit/state.json` with spec counts
- State lifecycle: draft → implemented → deployed → validated (or failed)
- CLI displays spec counts per phase

**Testing Status:** Task 045 created for comprehensive testing (Phase 1: Infrastructure, Phase 2: Incremental workflows)

## Acceptance Criteria

- [x] New principle documented - "Stateful Specifications" in `framework/SMAQIT.md`
- [x] Process defined for backwards compatibility - State tracking enables incremental development
- [x] Guidelines for modular vs. refactored specs - Documented in wiki articles
- [x] Decision tree for extend vs. refactor - State machine and workflow docs created
- [x] Framework documentation updated - 5 framework files, 6 templates, 8 agents, 1 CLI file

## Related Tasks

- **Task 045** - Test stateful specifications infrastructure (Phase 1: Infrastructure, Phase 2: Incremental)
- **Task 046** - Document implementation workflows (defines incremental processing patterns)

## Notes

**Original Questions Addressed:**

- **When should a spec be extended vs. refactored?** 
  - Specs carry state; new features generate new specs (status: draft)
  - Refactoring existing specs maintains their ID but regenerates content
  
- **How do we version specs?**
  - Git versioning + `prompt_version` field tracks prompt evolution
  - Enables stale detection (spec predates prompt changes)
  
- **What triggers a spec refactor vs. new spec?**
  - User decision based on prompt content (new requirement = new spec, modified requirement = refactor)
  
- **How do implementation agents handle spec changes?**
  - Agents should check `status` field and skip already-implemented specs
  - **Gap identified:** Incremental processing may need additional implementation (Task 045 Phase 2 validates)

**Evolution of Task:**

Original scope was conceptual (define principles). Implementation delivered concrete mechanisms (state tracking) enabling iterative patterns. Testing and workflow documentation split into separate tasks (045, 046) to maintain focus.

**Session History:** See `docs/history/018_task_014_stateful_specifications_2026-01-03.md` for complete implementation details.

## Files Modified (24 total)

**Framework (5):** SMAQIT.md, PHASES.md, ARTIFACTS.md, TEMPLATES.md, AGENTS.md  
**Templates (6):** 5 spec templates + 1 implementation agent template  
**Agents (8):** All 5 specification agents + 3 implementation agents  
**CLI (1):** installer/main.go  
**Wiki (4):** 2 new articles + 2 updated articles
