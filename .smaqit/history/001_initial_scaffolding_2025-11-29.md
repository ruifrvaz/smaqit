# Session 001: Initial Scaffolding

**Date**: 2025-11-29
**Previous Session**: None

## Objective

Bootstrap the smaqit kit with Phase 1 scaffolding: framework spec, layer templates, agent definitions, and installer pseudo-code.

## Work Done

1. **SMAQIT.md** — Created core framework spec (v0.1.0) with:
   - Layers table (Business → Functional → Stack → Infrastructure → Coverage)
   - Phases table (Develop → Deploy → Validate)
   - Specification Agents table (5 agents)
   - Implementation Agents table (3 agents)
   - Usage Rules section

2. **Layer Templates** (5 files in `templates/`):
   - `business.template.md` — Actors, Goals, Flows, Business Rules
   - `functional.template.md` — User Flows, Data Model, API Contracts
   - `stack.template.md` — Languages, Frameworks, Build Tools
   - `infrastructure.template.md` — Compute, Networking, Observability
   - `coverage.template.md` — Integration, E2E, Acceptance Tests

3. **Agent Definitions** (8 files in `agents/`):
   - 5 Specification Agents: business, functional, stack, infrastructure, coverage
   - 3 Implementation Agents: development, deployment, validation
   - All use GitHub Custom Agent format (YAML frontmatter + markdown)

4. **Installer** (`installer/main.go`):
   - Go CLI pseudo-code with commands: init, develop, deploy, validate, status, version
   - Version const synced with SMAQIT.md

5. **Documentation**:
   - `README.md` — User-friendly getting started guide
   - `.github/copilot-instructions.md` — Kit development guidance

6. **Session Workflow** (meta):
   - Added `history/` with session template
   - Added `tasks/` with task template and planner
   - Updated copilot-instructions with "recap" and "wrap up" commands

## Decisions Made

- **Name**: "smaqit" (pronounced "smack it") — avoids negative connotations of "smack-it"
- **No features concept**: Business layer contains use cases as individual files
- **Coverage = deployed app only**: No unit tests, only integration/E2E/acceptance against deployed system
- **Installer versioning**: Matches framework version
- **Session workflow**: Meta-only, not part of framework or installer

## Open Questions

- How should agents invoke each other in the CLI workflow?
- Should the installer validate existing specs before running phases?
- What's the error handling strategy when specs are incomplete?

## Next Session

- Implement `smaqit init` command in Go
- Define spec file naming conventions
- Consider adding a `smaqit new` command for creating individual specs
