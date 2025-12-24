# Task 029: Implementation Prompts Architecture

**Date:** 2025-12-24  
**Session Focus:** Complete task 029 - Simplify implementation prompts to minimal orchestration inputs

## Actions Taken

### Architecture Decision: Option 4

Decided on Option 4 architecture for implementation workflow:
- 3 implementation prompts (`development`, `deployment`, `validation`) for individual phase execution
- 1 orchestrator prompt (`orchestrate`) for full workflow coordination
- Users choose granularity: single phase or complete workflow

### Prompt Files Converted

Transformed 3 phase prompts from orchestration workflows to minimal parameter collectors:

1. **Development prompt** (`smaqit.develop` → `smaqit.development`):
   - Renamed for consistency with agent name
   - Reduced to: Build Options, Output Preferences, Environment
   - Removed orchestration logic (moved to agent)

2. **Deployment prompt** (`smaqit.deploy` → `smaqit.deployment`):
   - Renamed for consistency
   - Reduced to: Deployment Target, Verification, Output Preferences
   - Removed pre-run validation logic

3. **Validation prompt** (`smaqit.validate` → `smaqit.validation`):
   - Renamed for consistency
   - Reduced to: Test Scope, Failure Handling, Output Preferences
   - Removed completion criteria logic

### New Orchestrator Components Created

**Orchestrator prompt** (`smaqit.orchestrate.prompt.md`):
- Collects workflow execution parameters:
  - Phases to Execute (which phases to run)
  - Pre-Validation (validate prompts before execution)
  - Error Handling (stop vs continue on errors)
  - Output Preferences (verbosity level)

**Orchestrator agent** (`smaqit.orchestrator.agent.md`):
- Coordinates full workflow execution
- Pre-run validation: checks all 5 layer prompts + implementation prompts
- Sequential invocation: 5 spec agents → 3 implementation agents
- Error handling: stop-on-error or continue-through modes
- Completion verification: validates all phases executed successfully

**Orchestrator agent template** (`orchestrator-agent.template.md`):
- Created with proper `[PLACEHOLDER]` format
- Sections: Role, Input, Output, Directives, Orchestration Workflow, Failure Handling

### Tools Field Removal

Removed `tools` field from all prompt files (8 total):
- Layer prompts: business, functional, stack, infrastructure, coverage
- Implementation prompts: development, deployment, validation
- Orchestrator prompt: orchestrate

**Rationale:** Prompts should not override agent tool definitions. Agents control their own tools via `.agent.md` frontmatter.

### Template Updates

**Created:**
- `implementation-prompt.template.md` - Structure for development/deployment/validation prompts

**Modified:**
- `specification-prompt.template.md` - Removed tools field
- `phase-prompt.template.md` → `orchestrator-prompt.template.md` - Renamed and updated to parameter-based structure

### Framework Documentation Updates

**AGENTS.md:**
- Added orchestrator agent section
- Updated naming convention table to include orchestrator
- Documented orchestrator directives (MUST/MUST NOT/SHOULD)
- Added orchestrator tooling requirements (`agent` tool for invoking other agents)
- Removed excessive meta-rationale per instructions vs rationale rule

**TEMPLATES.md:**
- Added `orchestrator-agent.template.md` to agent templates listing
- Already had `orchestrator-prompt.template.md` documented

**PROMPTS.md:**
- Already documented orchestrator prompt architecture
- Confirmed implementation prompts + orchestrator prompt structure

## Problems Solved

### Process Issues Identified

**Catastrophic Forgetting Pattern:**
- Root cause: session.recap only read first 100 lines of framework files
- Result: Repeatedly violated same principles (meta-rationale in templates, tools field preservation)
- Solution: Updated copilot-instructions.md to explicitly require reading complete files without line truncation

**Template vs Instance Confusion:**
- Initially created orchestrator agent template as direct copy of instance
- Should have used placeholders (`[PLACEHOLDER]`) for customizable sections
- Fixed template to proper structure with placeholders

### Separation of Concerns Clarified

**Prompts vs Agents:**
- **Prompts:** Collect user input parameters
- **Agents:** Execute workflow logic (pre-validation, orchestration, error handling)
- This separation enables flexible customization without modifying agent code

**Orchestration Logic Location:**
- Originally put workflow steps in prompts (wrong)
- Moved to orchestrator agent: pre-run validation, agent invocation sequence, error handling strategies (correct)

## Decisions Made

1. **Option 4 Architecture:** Implementation prompts + orchestrator prompt over other options
2. **Naming Consistency:** Phase names match agent names (development not develop)
3. **Parameter-Based Prompts:** Prompts collect parameters, not workflows
4. **Tools Field Removal:** Agents control tools, prompts don't override
5. **Template Structure:** All templates use `[PLACEHOLDER]` format, no full content copies

## Files Modified

**Prompt files (8):**
- `prompts/smaqit.business.prompt.md` - Removed tools field
- `prompts/smaqit.functional.prompt.md` - Removed tools field
- `prompts/smaqit.stack.prompt.md` - Removed tools field
- `prompts/smaqit.infrastructure.prompt.md` - Removed tools field
- `prompts/smaqit.coverage.prompt.md` - Removed tools field
- `prompts/smaqit.develop.prompt.md` → `prompts/smaqit.development.prompt.md` - Renamed, converted to parameters
- `prompts/smaqit.deploy.prompt.md` → `prompts/smaqit.deployment.prompt.md` - Renamed, converted to parameters
- `prompts/smaqit.validate.prompt.md` → `prompts/smaqit.validation.prompt.md` - Renamed, converted to parameters

**New prompt:**
- `prompts/smaqit.orchestrate.prompt.md` - Created orchestrator prompt

**Template files (3):**
- `templates/prompts/specification-prompt.template.md` - Removed tools field, removed meta-rationale
- `templates/prompts/phase-prompt.template.md` → `templates/prompts/orchestrator-prompt.template.md` - Renamed, converted to parameters
- `templates/prompts/implementation-prompt.template.md` - Created new template

**Agent files:**
- `agents/smaqit.orchestrator.agent.md` - Created orchestrator agent
- `templates/agents/orchestrator-agent.template.md` - Created template with placeholders
- `templates/agents/implementation-agent.template.md` - Added `agent` tool (for orchestrator use case)

**Framework files:**
- `framework/AGENTS.md` - Added orchestrator section, removed meta-rationale
- `framework/TEMPLATES.md` - Added orchestrator-agent.template.md listing
- `framework/PROMPTS.md` - Already documented orchestrator prompt

**Process files:**
- `.github/copilot-instructions.md` - Updated session.recap to read complete files
- `docs/tasks/029_simplify_implementation_prompts.md` - Updated status to Completed
- `docs/tasks/PLANNING.md` - Moved task 029 from Active to Completed

## Key Principles Reinforced

1. **Framework Levels:** Level 0 (framework) → Level 1 (templates) → Level 2 (instances) must be respected
2. **No Meta-Rationale:** Framework files contain instructions, not explanations of why
3. **Templates Have Placeholders:** Templates use `[PLACEHOLDER]` format, not full content
4. **Complete Context Required:** Must read entire framework files, not just first 100 lines
5. **Separation of Input/Logic:** Prompts collect parameters, agents execute workflows

## Next Steps

Task 029 complete. Active tasks remaining:
- Task 014: Define iterative development using smaqit
- Task 015: Investigate framework bundling at installation
- Task 022: Create GitHub Action for automated releases
- Task 025: Integrate testing agent with CI/CD
- **Task 030: Move session and task commands into prompts** (created this session, pick up next)

## Notes

This task established the final prompt architecture for smaqit:
- **6 layer prompts:** Capture requirements for specifications
- **3 implementation prompts:** Trigger individual phases with optional parameters
- **1 orchestrator prompt:** Coordinate full workflow with execution preferences

The architecture enables both surgical execution (single phase) and full workflow orchestration (all phases), giving users flexibility in how they invoke the framework.
