# Session: Copilot Prompts and User Testing Agent

**Date:** 2025-12-20  
**Session Type:** Implementation  
**Tasks Completed:** 001, 024, 025 (created)

## Overview

Implemented two major features: (1) Copilot prompt files enabling users to invoke smaqit workflows via chat commands, and (2) User testing agent that automates end-to-end validation with standardized reporting. Session included significant scope clarifications and architectural corrections.

## What Was Done

### Task 001: Copilot Prompt Files

**Problem:** Users needed a simple way to invoke smaqit workflows without remembering agent names or framework structure.

**Solution:** Created 8 `.prompt.md` files (5 layer + 3 phase) using GitHub Copilot prompt format:
- Layer prompts: Single agent invocation with user input collection
- Phase prompts: Multi-agent orchestration with sequential workflows

**Files created:**
- `prompts/smaqit.business.prompt.md` - Business specification prompt
- `prompts/smaqit.functional.prompt.md` - Functional specification prompt
- `prompts/smaqit.stack.prompt.md` - Stack specification prompt
- `prompts/smaqit.infrastructure.prompt.md` - Infrastructure specification prompt
- `prompts/smaqit.coverage.prompt.md` - Coverage specification prompt
- `prompts/smaqit.develop.prompt.md` - Development phase orchestration (renamed to .development in Task 029)
- `prompts/smaqit.deploy.prompt.md` - Deployment phase orchestration (renamed to .deployment in Task 029)
- `prompts/smaqit.validate.prompt.md` - Validation phase orchestration (renamed to .validation in Task 029)

**Integration:**
- Modified `installer/main.go` to embed and copy prompts to `.github/prompts/` in user projects
- Updated `installer/Makefile` prepare target to copy prompts before build
- Added `installer/prompts/` to `.gitignore`
- Documented prompt structure and editing guidelines in `.github/copilot-instructions.md`

### Task 024: User Testing Agent

**Problem:** Manual testing of complete smaqit workflows was time-consuming and inconsistent. Needed automated validation from user perspective to catch regressions and document painpoints.

**Solution:** Created testing agent that orchestrates build → init → 5 spec layers → report → cleanup workflow using standardized test case.

**Files created:**
- `.github/agents/smaqit.user-testing.agent.md` - Testing orchestration agent
- `docs/test-cases/mario-hello.md` - Standardized test feature (Mario hello world)
- `docs/user-testing/report-template.md` - Strict format for test reports

**Key decisions:**
- **Hybrid automation:** Agent executes commands but validates minimally (file existence only)
- **Continues on failure:** Collects comprehensive results rather than stopping at first error
- **Strict report template:** 39-point checklist, execution log, painpoints section, recommendations
- **Mario test case:** Simple, domain-agnostic, exercises all 5 layers
- **Development only:** Agent located in `.github/agents/` (NOT shipped to users)

### Task 025: CI/CD Integration (Created)

Created follow-up task for future automation of testing agent via Makefile and GitHub Actions. Deferred until multi-contributor phase.

## Scope Clarifications

### Testing Agent Location Misconception

**Initial understanding:** Testing agent would be part of smaqit executable, shipped to users.

**Correction:** Testing agent is for smaqit development only:
- Located in `.github/agents/` (development tooling)
- NOT in `agents/` (shipped to users via installer)
- Used by maintainers to validate smaqit itself
- Never seen or invoked by end users

**Actions taken:**
- Moved agent from `agents/` to `.github/agents/`
- Updated all references and documentation
- Added explicit note in agent file clarifying non-shipping status

### Test Case Location

**Initial placement:** `templates/testing-feature-mario-hello.md`

**Correction:** Test cases are documentation, not templates:
- Moved to `docs/test-cases/mario-hello.md`
- Created new `docs/test-cases/` directory for future test scenarios
- Updated all references in agent, task file, copilot instructions

## Files Modified

**Created:**
- `prompts/*.prompt.md` (8 files)
- `.github/agents/smaqit.testing.agent.md`
- `docs/test-cases/mario-hello.md`
- `docs/user-testing/report-template.md`
- `docs/tasks/025_testing_agent_ci_integration.md`

**Modified:**
- `installer/main.go` - Added prompts embedding and copying
- `installer/Makefile` - Added prompts to prepare target
- `.gitignore` - Added installer/prompts/ ignore rule
- `.github/copilot-instructions.md` - Added prompt editing and testing agent sections
- `docs/tasks/PLANNING.md` - Moved Task 024 to Completed, added Task 025
- `docs/tasks/024_debate_smaqit_testing_agent.md` - Updated file paths and clarifications

## Why These Decisions

### Prompt Files Format

Used GitHub Copilot `.prompt.md` format (YAML frontmatter + markdown) because:
- Native integration with Copilot chat
- Input variable collection via `${input:}` syntax
- Agent orchestration via natural language "Run subagent X"
- No official VS Code docs found, so Task 001 spec served as authoritative reference

### Testing Agent Automation Level

Chose hybrid (automated execution + minimal validation) over full automation because:
- LLM-generated specs vary between runs (non-deterministic output)
- File existence check sufficient - detailed validation delegated to `smaqit validate` command
- Continuing on failure enables comprehensive painpoint collection
- Reduces false negatives from formatting variations

### Mario Hello World Test Case

Selected for standardization because:
- Domain-agnostic (works for any project type)
- Simple enough to complete quickly
- Rich enough to exercise all 5 layers (business use case, functional flows, stack choices, infrastructure deployment, coverage verification)
- Culturally familiar (reduces cognitive load)
- Expandable foundation for future test scenarios

### Strict Report Template

Enforced structure with 39-point checklist because:
- Enables regression tracking across sessions
- Separates painpoints into categories (blockers/issues/UX friction/performance)
- Standardizes recommendations format
- Machine-parseable structure for future automation

## Next Steps

**Immediate:**
1. Build installer with `make clean && make build` to verify prompt embedding
2. Test installer: `smaqit init test-project` to verify prompt copying
3. Invoke `@smaqit.user-testing` to validate end-to-end workflow
4. Review first test report for painpoints and iterate

**Future (Task 025):**
- Integrate testing agent with Makefile (`make test-e2e`)
- Add GitHub Actions workflow for automated testing on PR
- Consider parallel test execution for multiple scenarios
- Defer until multi-contributor phase (more valuable with team)

**Open Tasks:**
- Task 014: Define iterative development using smaqit
- Task 015: Investigate framework bundling at installation
- Task 022: Create GitHub Action for automated releases

## Lessons Learned

**1. Scope clarification is critical early:**
- Testing agent location misconception took multiple corrections
- Could have been resolved with single clarifying question upfront
- Ask about deployment/usage context before implementing architectural components

**2. File organization reflects purpose:**
- Templates are for agent consumption
- Docs are for human reference (including test cases)
- Development tooling goes in `.github/` (not shipped)
- This organizational pattern should be consistent

**3. Minimal validation reduces brittleness:**
- LLM output variability makes exact matching fragile
- Behavioral validation (files exist, commands succeed) more robust
- Delegate detailed validation to purpose-built tools (`smaqit validate`)

**4. Standardization enables regression tracking:**
- Strict report format allows comparison across sessions
- Painpoints categorization helps prioritize improvements
- Execution log captures workflow timing and bottlenecks

## Testing Philosophy Established

This session established smaqit's approach to self-validation:

**User perspective over framework internals:**
- Test what users experience, not internal structure
- Validate workflows end-to-end, not individual components
- Documentation quality matters (specs should be clear enough for implementation)

**Comprehensive reporting over binary pass/fail:**
- Continue on failures to collect all painpoints
- Categorize issues by severity and type
- Provide actionable recommendations
- Historical reports show progress/regressions

**Automation as accelerator, not replacement:**
- Agent handles tedious execution steps
- Human reviews results and makes decisions
- Minimal validation reduces false positives
- Focus automation on high-value repeatability
