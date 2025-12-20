# Task: Create smaqit User Testing Agent

**ID**: 024
**Status**: Completed
**Created:** 2025-12-19
**Completed:** 2025-12-20

## Description

Create a specialized testing agent that validates smaqit from a user perspective by orchestrating end-to-end workflows. The agent automates installer execution, project initialization, and guides through implementing a standardized test feature (Mario hello world console application) across all specification layers and phases.

## Context

User perspective testing validates the complete smaqit workflow:
1. Build installer from source
2. Initialize new project with `smaqit init`
3. Use layer prompts to create specifications
4. Verify spec outputs follow templates
5. Document user experience and painpoints

The testing agent provides **hybrid automation** (executes commands, creates files) with **minimal validation** (file existence only). Continues execution on failures to generate comprehensive reports with standardized checklists.

## Investigation Areas

### 1. Framework Validation
Test Case: Mario Hello World Console Application

A standardized, domain-agnostic test feature used to validate all smaqit layers and phases.

**Test Feature:** Console application that greets users in Nintendo Mario style
- **Business:** Greet Mario fans with iconic character catchphrases
- **Functional:** ASCII art output with colorful console text
- **Stack:** Programming language supporting console color output
- **Infrastructure:** Local execution or containerized deployment
- **Coverage:** Verify Mario-themed elements appear in output

**Why this test case:**
- Simple enough for quick validation
- Complex enough to exercise all layers
- Domain-agnostic (works across project types)
- Expandable (foundation for comprehensive test suite)
This is an investigation task. Completion criteria:

- [ ] Evaluate each investigation area for feasibility and value
- [ ] Determine if agent-based testing is appropriate for smaqit self-validation
- [ ] Decide on one of the following outcomes:
  - **Create specialized testing agent** — If automated validation provides clear value
  - **Enhance manual testing procedures** — If agent overhead exceeds benefit
  - **Hybrid approach** — Some automated checks, some manual workflows
  - **Defer** — Current testing approach is sufficient for now
- [ ] Document decision with rationale

## Implementation Notes

**Hybrid Automation:**
- Agent executes terminal commands (build, init, create files)
- User confirms outputs at minimal validation checkpoints
- Continues on failure to collect comprehensive results

**Report Structure:**
- Strict template ensures consistency across test runs
- Standardized checklist enables regression tracking
- Painpoints section captures UX friction for iteration

**Future Expansion:**
- Current: Single test case (Mario hello world)
- Future: Comprehensive test suite with multiple use cases (Task 025: CI/CD integration)
- Expandable architecture supports additional test scenarios

**Agent Location:**
- Located in `.github/agents/` (smaqit development only)
- NOT in `agents/` (which gets shipped to users via installer)
- This is a development/testing tool for smaqit maintainers
- Users never see or interact with this agent

**Validation Philosophy:**
- Minimal validation (file existence only)
- Detailed validation delegated to `smaqit validate` command
- Focus on user workflow, not framework internals

**Alternative: GitHub Actions CI**
- Could achieve similar validation with traditional CI pipelines
- Consider hybrid: CI for structural checks, agent for semantic validation

## Open Questions

- Would a testing agent need special tools beyond read/search?
- Should validation reports be machine-readable (JSON) or human-readable (markdown)?
- Does testing agent fit smaqit's "agents produce specs first" model?
- Would this testing agent itself need specs? (Meta-layer concern)
- [x] Create testing agent (`.github/agents/smaqit.user-testing.agent.md`)
- [x] Define standardized test case (`docs/test-cases/mario-hello.md`)
- [x] Create report template with strict format (`docs/user-testing/report-template.md`)
- [x] Document testing workflow in copilot instructions
- [x] Agent generates dated reports in `docs/user-testing/`
- [x] Agent cleans up test projects after execution
- [x] Report includes standardized checklist (pass/fail per validation point)
- [x] Report includes painpoints identification section