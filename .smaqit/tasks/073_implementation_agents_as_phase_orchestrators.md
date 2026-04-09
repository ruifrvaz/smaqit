# Task 073: Implementation Agents as Phase Orchestrators

**Status:** completed  
**Priority:** High  
**Created:** 2026-01-23  
**Completed:** 2026-02-08

## Context

Currently, implementation agents expect specifications to already exist. Users must manually invoke specification agents (business, functional, stack) before running implementation agents. This creates friction and doesn't align with the "Phases as Workflow Units" principle.

The intended architecture: each implementation agent coordinates its entire phase, including spec generation when needed.

## Problem

**Current workflow (manual):**
```
User: /smaqit.business    → specs/business/*.md
User: /smaqit.functional  → specs/functional/*.md
User: /smaqit.stack       → specs/stack/*.md
User: /smaqit.development → code + tests
```

**Desired workflow (automated):**
```
User: /smaqit.development → specs (if missing) + code + tests
```

**Pain points:**
- Users must remember multi-step invocation sequence
- No automatic detection of missing/stale specs
- Breaks "Phases as Workflow Units" principle
- Implementation agents are passive consumers, not phase coordinators

## Goal

Enable implementation agents to invoke specification agents internally, making them true phase orchestrators.

## Scope

### Implementation Agent Responsibilities

**Development Agent (`smaqit.development`):**
- Check for specs: business, functional, stack
- If missing or `--regen` flag: invoke business → functional → stack agents
- Agents read from `.github/prompts/smaqit.[layer].prompt.md`
- After specs ready: proceed with code generation

**Deployment Agent (`smaqit.deployment`):**
- Check for infrastructure specs
- If missing or `--regen` flag: invoke infrastructure agent
- Agent reads from `.github/prompts/smaqit.infrastructure.prompt.md`
- After specs ready: proceed with deployment

**Validation Agent (`smaqit.validation`):**
- Check for coverage specs
- If missing or `--regen` flag: invoke coverage agent
- Agent reads from `.github/prompts/smaqit.coverage.prompt.md`
- After specs ready: proceed with validation

### Reference: Orchestrator Agent Section in AGENTS.md

The Orchestrator Agent section in `framework/AGENTS.md` (preserved after Task 072) contains valuable orchestration workflows and directives that should be adapted for this implementation:

- **Pre-run validation workflow** — Check prompt sufficiency before invoking spec agents
- **Agent invocation sequencing** — Order and dependency management
- **Error handling patterns** — Stop vs continue, retry strategies
- **Progress reporting** — User feedback during multi-agent workflows

These patterns should be incorporated into each implementation agent's orchestration logic with appropriate tuning for phase-specific contexts.

### Spec Detection Logic

Implementation agents use `smaqit plan --phase=[PHASE]` to determine work:
- If specs exist with `status: draft` → process existing specs
- If no specs exist → invoke spec agents first
- If `--regen` flag → invoke spec agents regardless of status

### Agent Invocation Pattern

Implementation agents gain orchestration capability:
- Read corresponding prompt files to verify sufficiency
- Invoke spec agents in correct sequence
- Wait for spec generation completion
- Validate specs before proceeding to implementation

### Files to Update

**Product agents:**
- `agents/smaqit.development.agent.md` — Add orchestration workflow
- `agents/smaqit.deployment.agent.md` — Add orchestration workflow
- `agents/smaqit.validation.agent.md` — Add orchestration workflow

**Agent templates:**
- `templates/agents/implementation-agent.template.md` — Add orchestration section

**Framework documentation:**
- `framework/PHASES.md` — Document implementation agent orchestration model
- `framework/AGENTS.md` — Update implementation agent section with orchestration responsibilities

**Installer:**
- Rebuild to include updated agents

## Implementation Steps

1. **Update implementation agent template** — Add orchestration workflow section
2. **Update Development agent** — Add spec agent invocation logic
3. **Update Deployment agent** — Add infrastructure agent invocation logic
4. **Update Validation agent** — Add coverage agent invocation logic
5. **Update framework docs** — PHASES.md and AGENTS.md reflect orchestration model
6. **Test E2E workflow** — Empty project → `/smaqit.development` → specs + code generated
7. **Rebuild installer** — Package updated agents

## Acceptance Criteria

- [x] Development agent invokes business → functional → stack when specs missing
- [x] Deployment agent invokes infrastructure when specs missing
- [x] Validation agent invokes coverage when specs missing
- [x] Agents check prompt files for sufficiency before invoking spec agents (Pre-Orchestration Validation)
- [x] Agents respect `--regen` flag for forced spec regeneration
- [x] Framework documentation reflects orchestration model (AGENTS.md updated with orchestration concepts)
- [ ] E2E test passes: empty project → `/smaqit.development` → working app (testing pending)
- [x] User workflow simplified: one command per phase

## Dependencies

- Task 072: Remove Orchestrator Agent Pattern (must complete first)

## Follow-up Tasks

- Task 074: Update "Extensible Through Templates" Principle Context (optional)

## Testing Strategy

**E2E Test Scenario:**
1. `smaqit init` in empty directory
2. Fill development phase prompts (business, functional, stack)
3. Run `/smaqit.development`
4. Verify: specs generated, code generated, tests pass
5. No manual spec agent invocation required

**Edge Cases:**
- Empty prompt files → agent halts with guidance
- Partial specs exist → agent processes only missing specs
- `--regen` flag → regenerate all specs regardless of status

## Notes

- This completes the "Phases as Workflow Units" architecture
- Aligns with framework principle: phases are primary workflow units
- Simplifies user mental model: one command per phase
- Maintains spec-first principle: specs generated before implementation
- No backwards compatibility needed — this is enhancement, not breaking change

## Completion Summary

**Completed 2026-02-08:**

**L0 Framework Updates:**
- Added Phase Orchestration concept to AGENTS.md (specification generation coordination, multi-agent coordination, progress tracking, error context preservation)
- Added Pre-Orchestration Validation concept (input validation, dependency verification, execution readiness)
- Added Orchestration Completion Validation concept (activity completion, outcome validation, completion status)
- Updated PHASES.md to reference pre-orchestration validation in all three phases

**L1 Template and Compilation Updates:**
- Added multi-format compilation support to Agent-L1 (7 format types: directive, checklist, workflow, table, role, structure, frontmatter)
- Added format inference process to identify compilation format from L0 content patterns
- Added implementation-agent.template.md orchestration section headers (Pre-Orchestration Validation, Phase Orchestration, Orchestration Completion Validation)
- Added `runSubagent` tool to implementation agent template
- Compiled orchestration sections in implementation.rules.md using proper format types:
  - Pre-Orchestration Validation → checklist format (12 validation checks + Pass/Fail outcomes)
  - Phase Orchestration → workflow format (7-step sequential workflow)
  - Orchestration Completion Validation → checklist format (11 completion checks + Success/Partial/Failed status)

**L2 Product Agent Updates:**
- Compiled orchestration sections into smaqit.development.agent.md with `phase=develop` values
- Compiled orchestration sections into smaqit.deployment.agent.md with `phase=deploy` values
- Compiled orchestration sections into smaqit.validation.agent.md with `phase=validate` values
- All agents now include `runSubagent` tool for agent invocation
- All agents now include 7-step phase workflow for specification generation and implementation coordination

**Format Correction Impact:**
- Identified foundational issue: Agent-L1 lacked multi-format compilation capability
- Corrected orchestration sections from directive format (MUST/MUST NOT/SHOULD) to operational formats (checklists/workflows)
- Reduced orchestration content by 59% (222 lines → 92 lines) through format-appropriate compilation

**Commits:**
- `f511f0c` Agent: Update L2 agent tools to namespaced format
- `cc0a610` L1: Add multi-format compilation support
- `1f14183` L1: Rewrite orchestration sections with proper format types
- `b0f48eb` Docs: Add L2 orchestration compilation log

**Pending:**
- E2E testing with actual phase execution
- Installer rebuild with updated agents (deferred to release workflow)

