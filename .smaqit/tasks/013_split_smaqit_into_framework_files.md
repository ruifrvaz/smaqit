# Task: Split SMAQIT.md into Framework Files

**ID**: 013
**Status**: completed
**Completed**: 2025-12-13

## Context

The current `framework/SMAQIT.md` contains all framework documentation in a single file. Split it into focused files, each describing a core feature of smaqit:

1. **SMAQIT.md** — Index + core principles
2. **PHASES.md** — The development phases (Develop, Deploy, Validate)
3. **LAYERS.md** — The specification layers (Business, Functional, Stack, Infrastructure, Coverage)
4. **SPECIFICATIONS.md** — The specification artifact rules
5. **IMPLEMENTATIONS.md** — The implementation artifact rules and principles
6. **AGENTS.md** — The agent definitions, roles, and how they map to layers/phases

This separation improves maintainability and allows agents to load only the context they need.

## Acceptance Criteria

- [x] Create `framework/PHASES.md` — Phase definitions, ordering, and transitions
- [x] Create `framework/LAYERS.md` — Layer definitions, dependencies, and agent mappings
- [x] Create `framework/SPECIFICATIONS.md` — Specification format and validation rules
- [x] Create `framework/IMPLEMENTATIONS.md` — Implementation artifact rules and principles
- [x] Create `framework/AGENTS.md` — Agent roles, responsibilities, and layer/phase mappings
- [x] Update `framework/SMAQIT.md` to be an index/overview that references the framework files
- [x] Update installer TODOs to copy all framework files (not just SMAQIT.md)
- [x] Update copilot-instructions to reference new framework structure

## Notes

- SMAQIT.md serves as entry point + contains core principles (merged from PRINCIPLES.md)
- All framework files have See Also cross-references
- Installer copies entire `framework/` directory
