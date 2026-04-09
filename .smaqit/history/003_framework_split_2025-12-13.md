# Session: Framework Split

**Date**: 2025-12-13

## Summary

Split monolithic `framework/SMAQIT.md` into 6 focused framework files. This major restructuring improves maintainability and allows agents to load only the context they need.

## Actions Taken

1. **Created framework files**:
   - `AGENTS.md` — Agent behaviors, unified principles, spec/impl agent definitions, validation
   - `LAYERS.md` — 5 layers with dependencies, MUST/MUST NOT directives
   - `SPECIFICATIONS.md` — Requirement IDs, acceptance criteria, traceability, Gherkin translation
   - `IMPLEMENTATIONS.md` — Anchoring Principle, Isolation Principle, three dimensions
   - `PHASES.md` — 3 phases, workflows, trusted execution layer, failure handling

2. **Refactored SMAQIT.md** as index + core principles (merged from PRINCIPLES.md)

3. **Deleted PRINCIPLES.md** after merging into SMAQIT.md

4. **Updated supporting files**:
   - `copilot-instructions.md` — References new framework structure
   - `installer/main.go` — TODOs updated for copying `framework/` directory

5. **Completed task 013** in PLANNING.md

## Key Decisions

- **Spec vs Implementation distinction**: Specs are declarative (what must be true), implementations are imperative (how to make it true)
- **Anchoring Principle**: Implementations must be "structurally recognizable, behaviorally equivalent"
- **Isolation Principle**: Agents use references (`${secrets.X}`), never values — trusted execution layer resolves them
- **Framework index**: SMAQIT.md serves as entry point with core principles, not a standalone monolith

## Files Modified

- `framework/SMAQIT.md` (refactored)
- `framework/AGENTS.md` (created)
- `framework/LAYERS.md` (created)
- `framework/PHASES.md` (created)
- `framework/SPECIFICATIONS.md` (created)
- `framework/IMPLEMENTATIONS.md` (created)
- `framework/PRINCIPLES.md` (deleted)
- `.github/copilot-instructions.md` (updated)
- `installer/main.go` (TODOs updated)
- `docs/tasks/013_split_smaqit_into_framework_files.md` (completed)
- `docs/tasks/PLANNING.md` (status updated)

## Next Steps

- Task 002: Create specification agent template (uses SPECIFICATIONS.md)
- Task 003: Create implementation agent template (uses IMPLEMENTATIONS.md)
- Consider task 014: Define iterative development using smaqit
