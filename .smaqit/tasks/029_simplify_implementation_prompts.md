# Simplify implementation prompts to minimal orchestration inputs

**Status:** Completed  
**Created:** 2025-12-22

## Architecture Decision

**Decision:** Option 4 - Implementation prompts + orchestrator prompt

- 3 implementation prompts: `smaqit.development`, `smaqit.deployment`, `smaqit.validation`
- 1 orchestrator prompt: `smaqit.orchestrate` 
- Users choose granularity (single phase or full workflow)

### Current State

- 3 phase prompts (develop, deploy, validate)
- Each triggers a phase agent (Development, Deployment, Validation)
- Each phase agent orchestrates multiple spec agents + implementation

### Options Under Consideration

**Option 1: Phase prompts orchestrate entire phases (current)**
- `/smaqit.development` → Development agent → orchestrates Business → Functional → Stack → builds
- `/smaqit.deployment` → Deployment agent → orchestrates Infrastructure → deploys
- `/smaqit.validation` → Validation agent → orchestrates Coverage → tests
**(Note: Final names use full names for consistency with agent names)**

**Option 2: Implementation prompts trigger implementation agents only**
- `/smaqit.development` → Development agent (reads Business + Functional + Stack specs, builds)
- `/smaqit.deployment` → Deployment agent (reads Infrastructure spec, deploys)
- `/smaqit.validation` → Validation agent (reads Coverage spec, tests)
- Users manually invoke spec agents separately

**Option 3: Both phase prompts AND implementation prompts**
- Phase prompts for full workflows (convenience)
- Implementation prompts for targeted execution (flexibility)

**Option 4: Implementation prompts + 1 orchestrator prompt**
- `/smaqit.development`, `/smaqit.deployment`, `/smaqit.validation` → individual implementation agents
- `/smaqit.orchestrate` → Orchestrator agent → calls all 3 phases in sequence
- Users choose granularity

### Trade-offs

| Aspect | Option 1 | Option 2 | Option 3 | Option 4 |
|--------|----------|----------|----------|----------|
| Convenience | High (1 command) | Low (many commands) | High | High |
| Flexibility | Medium | High (surgical) | High | High |
| Complexity | Medium | Low | High (duplication) | Medium |
| Prompt count | 3 phases | 3 impl | 6 total | 4 (3 impl + orchestrator) |

### Decision Needed

Which architecture should we implement? This impacts:
- Number of prompt files
- Naming conventions (`develop` vs `development`, `deploy` vs `deployment`)
- Phase agent responsibilities (orchestrate specs or just implement?)
- User workflow patterns

**Recommendation pending user input.**

---

## Description

Implementation/phase prompts (develop, deploy, validate) should be minimal and contain only orchestration-specific inputs. Implementation agents collect their information from layer specs, not from prompts.

Additionally, `tools` field in prompt files overrides agent tool definitions and should be removed from all prompts.

## Problem

Current state:
- Phase prompts may contain unnecessary content
- Phase prompt template may suggest more than needed
- All prompt files have `tools` field which overrides agent definitions
- Unclear separation: what goes in prompts vs what agents read from specs

## Solution

**Phase prompts should only capture:**
- Orchestration parameters (e.g., "Run in watch mode", "Deploy to staging first")
- Phase-specific overrides (e.g., "Skip validation step X")
- Runtime configuration (e.g., "Use verbose logging")

**Phase prompts should NOT contain:**
- Requirements (those come from layer prompts)
- Implementation details (agents read specs)
- Duplicate information already in specs

**Remove tools field:**
- `tools` in prompt files overrides agent tool definitions
- Agents should control their own tools via `.agent.md` files
- Prompt files should only specify `name`, `description`, `agent`

## Scope

**Phase prompt files:**
- [x] `prompts/smaqit.develop.prompt.md` → renamed to `smaqit.development.prompt.md`
- [x] `prompts/smaqit.deploy.prompt.md` → renamed to `smaqit.deployment.prompt.md`
- [x] `prompts/smaqit.validate.prompt.md` → renamed to `smaqit.validation.prompt.md`

**Phase prompt template:**
- [x] `templates/prompts/phase-prompt.template.md` → renamed to `orchestrator-prompt.template.md`
- [x] Created `templates/prompts/implementation-prompt.template.md`

**All prompt files (tools removal):**
- [x] `prompts/smaqit.business.prompt.md`
- [x] `prompts/smaqit.functional.prompt.md`
- [x] `prompts/smaqit.stack.prompt.md`
- [x] `prompts/smaqit.infrastructure.prompt.md`
- [x] `prompts/smaqit.coverage.prompt.md`
- [x] `prompts/smaqit.development.prompt.md`
- [x] `prompts/smaqit.deployment.prompt.md`
- [x] `prompts/smaqit.validation.prompt.md`

**Template files (tools removal):**
- [x] `templates/prompts/specification-prompt.template.md`
- [x] `templates/prompts/orchestrator-prompt.template.md`
- [x] `templates/prompts/implementation-prompt.template.md`

## Acceptance Criteria

- [x] Phase prompt template simplified to minimal orchestration structure
- [x] All three phase prompts simplified (develop, deploy, validate)
- [x] `tools` field removed from all 8 prompt files
- [x] `tools` field removed from all 3 prompt templates
- [x] Framework documentation updated to reflect:
  - Phase prompts are for orchestration parameters only
  - Implementation agents read from specs, not phase prompts
  - Prompt files don't override agent tools
- [x] Orchestrator agent created with workflow logic
- [x] Orchestrator agent template created
- [x] Framework files updated (AGENTS.md, TEMPLATES.md, PROMPTS.md)

## What Phase Prompts Should Look Like

**Minimal structure:**
```markdown
---
name: smaqit.development
description: Orchestrate development phase
agent: smaqit.development
---

# Development Phase

## Orchestration Parameters

<!-- Example: "Run in watch mode for hot reload" -->
<!-- Example: "Skip build step, use existing artifacts" -->

[User's orchestration preferences here, if any]
```
**(Note: Final prompt names use full names: development, deployment, validation)**

**Not this:**
```markdown
## Requirements
[Detailed requirements - these come from layer prompts]

## Implementation Details
[How to build - agent decides from specs]
```

## Related

- Task 026: Rethink prompt architecture and integration
- PROMPTS.md: Prompt structure and agent interaction
- Phase agents: Development, Deployment, Validation

## Notes

This clarifies the role boundary:
- **Layer prompts**: Capture requirements for specs
- **Phase prompts**: Capture orchestration parameters for implementation
- **Agents**: Read specs, execute workflows, validate outputs
