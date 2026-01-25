# Task 072: Remove Orchestrator Agent Pattern

**Status:** new  
**Priority:** High  
**Created:** 2026-01-23

## Context

The orchestrator agent was designed for full-workflow automation but never documented in user-facing materials (README, help output). Users invoke implementation agents directly (`/smaqit.development`, `/smaqit.deployment`, `/smaqit.validation`) which is the intended workflow pattern.

The orchestrator pattern conflicts with the "Phases as Workflow Units" principle where each implementation agent should coordinate its own phase (spec generation + implementation).

## Problem

- Orchestrator agent exists in codebase (`agents/`, `templates/`, `installer/`)
- Orchestrator prompt exists (`prompts/`, `templates/`, `installer/`)
- Not documented in README or CLI help output
- Vestigial reference in installer (`main.go:748`)
- Adds maintenance burden and conceptual complexity
- Implementation agents don't currently invoke spec agents (Task 073 will enable this)

## Goal

Remove orchestrator agent pattern completely from codebase, preparing for phase-orchestrator architecture (Task 073).

## Scope

### Files to Remove

**Product agents:**
- `agents/smaqit.orchestrator.agent.md`
- `prompts/smaqit.orchestrate.prompt.md`

**Templates:**
- `templates/agents/orchestrator-agent.template.md`
- `templates/prompts/orchestrator-prompt.template.md`

**Installer embedded files:**
- `installer/agents/smaqit.orchestrator.agent.md`
- `installer/prompts/smaqit.orchestrate.prompt.md`

### Files to Update

**Framework documentation:**
- `framework/AGENTS.md` — Remove "Orchestrator Agent" section, update agent count table, remove `runSubagent` tool (only used by orchestrator)
- `framework/PROMPTS.md` — Remove "Orchestrator Prompt" section
- `framework/TEMPLATES.md` — Remove orchestrator template references

**Project documentation:**
- `.github/copilot-instructions.md` — Update agent count (9 → 8 agents)

**Installer:**
- `installer/main.go:748` — Remove vestigial reference: `"Run '/smaqit.orchestrate' to iterate"`

### E2E Test Cleanup

Review `installer/test/` directories for orchestrator artifacts in test snapshots. Update if needed.

## Implementation Steps

1. Remove product agent and prompt files (agents/, prompts/)
2. Remove template files (templates/agents/, templates/prompts/)
3. Remove installer embedded files (installer/agents/, installer/prompts/)
4. Update framework documentation (AGENTS.md, PROMPTS.md, TEMPLATES.md)
5. Update copilot-instructions.md agent count
6. Remove vestigial installer reference
7. Verify no remaining references: `grep -r "orchestrat" --include="*.md" --include="*.go"`

## Acceptance Criteria

- [x] All orchestrator agent files removed (product, templates, installer)
- [x] All orchestrator prompt files removed (product, templates, installer)
- [x] Framework documentation updated (no orchestrator references)
- [x] Agent count reflects 8 agents (5 spec + 3 impl)
- [x] `runSubagent` tool removed from framework docs
- [x] No remaining "orchestrat" references in codebase (except "orchestration" in README tagline and history files)
- [x] Installer builds successfully
- [x] `smaqit init` works without orchestrator files

## Dependencies

- None (standalone cleanup)

## Follow-up Tasks

- Task 073: Implementation Agents as Phase Orchestrators (enables new architecture)
- Task 074: Update "Extensible Through Templates" Principle Context (optional documentation refinement)

## Notes

- Orchestrator was a well-intentioned design that added complexity without user-facing value
- Phase-first workflow (implementation agents as orchestrators) is simpler and more aligned with framework principles
- No backwards compatibility needed — orchestrator was never documented for user consumption
