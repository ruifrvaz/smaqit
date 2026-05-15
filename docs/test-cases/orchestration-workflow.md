# Test Case 002: Phase Orchestration Workflow

**Test ID:** 002  
**Feature:** Implementation agent orchestration of specification generation  
**Purpose:** Validate Task 073 implementation - agents automatically generate missing specs  
**Status:** Active  
**Created:** 2026-02-09

---

## Test Feature Description

Validates that implementation agents (development, deployment, validation) automatically detect missing specifications and invoke specification agents to generate them, simplifying the user workflow from multi-step to single-command per phase.

**Test validates:**
- Pre-orchestration validation (input validation, dependency checks, execution readiness)
- Specification generation coordination (detection, agent invocation, dependency ordering)
- Multi-agent coordination (runSubagent tool usage)
- Progress tracking and error handling
- Orchestration completion validation

**Success criteria:**
- User runs ONE command per phase (e.g., `/smaqit.development`)
- Agent automatically invokes specification agents when specs missing
- Specs generated in correct dependency order
- Implementation proceeds after specs ready
- Workflow completes successfully

---

## Test Scenario 1: Development Phase Orchestration

**Objective:** Validate development agent orchestrates business → functional → stack spec generation

**Initial State:**
- Fresh smaqit init (no specs exist)
- Prompt files filled:
  - Session context with requirements provided to agent

**Test Steps:**
1. Run `/smaqit.development` (single command)
2. Observe agent behavior

**Expected Behavior:**

**Pre-Orchestration Validation:**
- [ ] Agent reads requirements from session context
- [ ] Agent verifies no missing dependencies (none required for development)
- [ ] Agent verifies execution tools available
- [ ] Validation passes → proceed to orchestration

**Specification Generation Coordination:**
- [ ] Agent executes `smaqit plan --phase=develop`
- [ ] Agent detects missing business specs
- [ ] Agent invokes `/smaqit.business` using runSubagent tool
- [ ] Business specs generated in `specs/business/`
- [ ] Agent detects missing functional specs
- [ ] Agent invokes `/smaqit.functional` using runSubagent tool
- [ ] Functional specs generated in `specs/functional/`
- [ ] Agent detects missing stack specs
- [ ] Agent invokes `/smaqit.stack` using runSubagent tool
- [ ] Stack specs generated in `specs/stack/`

**Implementation Generation:**
- [ ] Agent consolidates specifications
- [ ] Agent generates implementation artifacts (code, tests)
- [ ] Agent executes build
- [ ] Agent runs tests

**Orchestration Completion Validation:**
- [ ] Agent verifies all specs generated
- [ ] Agent verifies implementation artifacts created
- [ ] Agent verifies tests passed
- [ ] Agent reports phase success

**Expected Outcome:**
- Business specs: `specs/business/*.md` created
- Functional specs: `specs/functional/*.md` created
- Stack specs: `specs/stack/*.md` created
- Implementation: Code + tests generated
- Build: Successful
- Tests: Passing
- Development report: `.smaqit/reports/development-phase-report-YYYY-MM-DD.md`

**User Workflow Comparison:**

*Before orchestration (manual):*
```
User: /smaqit.business
User: /smaqit.functional
User: /smaqit.stack
User: /smaqit.development
```

*After orchestration (automated):*
```
User: /smaqit.development
```

---

## Test Scenario 2: Deployment Phase Orchestration

**Objective:** Validate deployment agent orchestrates infrastructure spec generation

**Initial State:**
- Development phase complete (business, functional, stack specs + code exist)
- Infrastructure requirements provided in session context

**Test Steps:**
1. Run `/smaqit.deployment` (single command)
2. Observe agent behavior

**Expected Behavior:**

**Pre-Orchestration Validation:**
- [ ] Agent reads infrastructure requirements from session context
- [ ] Agent verifies development artifacts present (code, tests)
- [ ] Agent verifies execution tools available
- [ ] Validation passes → proceed to orchestration

**Specification Generation Coordination:**
- [ ] Agent executes `smaqit plan --phase=deploy`
- [ ] Agent detects missing infrastructure specs
- [ ] Agent invokes `/smaqit.infrastructure` using runSubagent tool
- [ ] Infrastructure specs generated in `specs/infrastructure/`

**Implementation Generation:**
- [ ] Agent consolidates infrastructure specifications
- [ ] Agent generates deployment artifacts (IaC, configs)
- [ ] Agent deploys to target environment

**Orchestration Completion Validation:**
- [ ] Agent verifies infrastructure specs generated
- [ ] Agent verifies deployment successful
- [ ] Agent reports phase success

**Expected Outcome:**
- Infrastructure specs: `specs/infrastructure/*.md` created
- Deployment artifacts: IaC code generated
- Deployment: Successful
- Deployment report: `.smaqit/reports/deployment-phase-report-YYYY-MM-DD.md`

**User Workflow Comparison:**

*Before orchestration (manual):*
```
User: /smaqit.infrastructure
User: /smaqit.deployment
```

*After orchestration (automated):*
```
User: /smaqit.deployment
```

---

## Test Scenario 3: Validation Phase Orchestration

**Objective:** Validate validation agent orchestrates coverage spec generation

**Initial State:**
- Development and deployment phases complete
- Coverage requirements provided in session context

**Test Steps:**
1. Run `/smaqit.validation` (single command)
2. Observe agent behavior

**Expected Behavior:**

**Pre-Orchestration Validation:**
- [ ] Agent reads coverage requirements from session context
- [ ] Agent verifies development and deployment artifacts present
- [ ] Agent verifies execution tools available
- [ ] Validation passes → proceed to orchestration

**Specification Generation Coordination:**
- [ ] Agent executes `smaqit plan --phase=validate`
- [ ] Agent detects missing coverage specs
- [ ] Agent invokes `/smaqit.coverage` using runSubagent tool
- [ ] Coverage specs generated in `specs/coverage/`

**Implementation Generation:**
- [ ] Agent consolidates coverage specifications
- [ ] Agent generates test artifacts
- [ ] Agent executes tests
- [ ] Agent generates validation report

**Orchestration Completion Validation:**
- [ ] Agent verifies coverage specs generated
- [ ] Agent verifies tests executed
- [ ] Agent reports phase success

**Expected Outcome:**
- Coverage specs: `specs/coverage/*.md` created
- Test execution: All tests run
- Validation report: `.smaqit/reports/validation-phase-report-YYYY-MM-DD.md`

**User Workflow Comparison:**

*Before orchestration (manual):*
```
User: /smaqit.coverage
User: /smaqit.validation
```

*After orchestration (automated):*
```
User: /smaqit.validation
```

---

## Test Scenario 4: Spec Regeneration Flag

**Objective:** Validate `--regen` flag triggers spec regeneration regardless of status

**Initial State:**
- All specs exist with `status: implemented`
- Prompt files have been updated with new requirements

**Test Steps:**
1. Re-invoke agent with updated requirements in session context
2. Run `/smaqit.development --regen`
3. Observe agent behavior

**Expected Behavior:**
- [ ] Agent detects `--regen` flag
- [ ] Agent invokes specification agents despite specs existing
- [ ] Existing specs regenerated with updated requirements
- [ ] Spec status reset to `status: draft`
- [ ] Implementation regenerated based on updated specs

**Expected Outcome:**
- Specs regenerated with new requirements
- Implementation reflects updated specs
- All specs have fresh timestamps

---

## Test Scenario 5: Partial Spec Scenario

**Objective:** Validate agent processes only missing specs

**Initial State:**
- Business and functional specs exist
- Stack specs missing

**Test Steps:**
1. Run `/smaqit.development`
2. Observe agent behavior

**Expected Behavior:**
- [ ] Agent detects business and functional specs present
- [ ] Agent detects stack specs missing
- [ ] Agent invokes ONLY `/smaqit.stack` (skips business, functional)
- [ ] Stack specs generated
- [ ] Implementation proceeds with all specs

**Expected Outcome:**
- Business and functional specs unchanged
- Stack specs newly created
- Implementation successful

---

## Test Scenario 6: Empty Prompt File Handling

**Objective:** Validate agent halts with guidance when requirements are insufficient

**Initial State:**
- Fresh smaqit init
- Session context provides no requirements or insufficient requirements

**Test Steps:**
1. Run `/smaqit.development`
2. Observe agent behavior

**Expected Behavior:**

**Pre-Orchestration Validation:**
- [ ] Agent detects session context lacks sufficient requirements
- [ ] Pre-orchestration validation fails
- [ ] Agent halts with diagnostic report
- [ ] Report identifies what requirements are needed
- [ ] Report provides remediation guidance

**Expected Outcome:**
- Workflow halted (no specs generated)
- Clear error message identifying missing requirements
- Guidance on what content is required

---

## Test Scenario 7: Progress Tracking

**Objective:** Validate agent provides progress visibility during orchestration

**Test Steps:**
1. Run `/smaqit.development` with missing specs
2. Observe console output

**Expected Progress Reports:**
- [ ] "Pre-orchestration validation: [status]"
- [ ] "Detecting missing specifications..."
- [ ] "Missing specs: business, functional, stack"
- [ ] "Invoking business agent..."
- [ ] "Business specs generated: specs/business/[files]"
- [ ] "Invoking functional agent..."
- [ ] "Functional specs generated: specs/functional/[files]"
- [ ] "Invoking stack agent..."
- [ ] "Stack specs generated: specs/stack/[files]"
- [ ] "Consolidating specifications..."
- [ ] "Generating implementation artifacts..."
- [ ] "Executing build..."
- [ ] "Running tests..."
- [ ] "Orchestration completion validation: [status]"
- [ ] "Development phase: SUCCESS"

**Expected Outcome:**
- User can track progress through each orchestration step
- Milestones visible during execution
- Clear success/failure indication

---

## Test Scenario 8: Error Handling

**Objective:** Validate agent preserves error context when specification generation fails

**Initial State:**
- Business prompt has conflicting requirements

**Test Steps:**
1. Fill business prompt with contradictory requirements
2. Run `/smaqit.development`
3. Observe error handling

**Expected Behavior:**
- [ ] Agent invokes business agent
- [ ] Business agent detects conflict and fails
- [ ] Development agent captures failure context
- [ ] Development agent reports:
  - Which agent failed (business)
  - What input was provided (session context requirements)
  - What error occurred (conflict description)
  - Remediation guidance (how to fix prompt)
- [ ] Workflow halts with partial completion tracking
- [ ] Report indicates "Phase: PARTIAL" (pre-orchestration validation passed, spec generation failed)

**Expected Outcome:**
- Clear error message with context
- User knows what requirements to provide
- User knows what the conflict is
- Workflow state preserved for resume after fix

---

## Layer Requirements for Test Case

### Business Layer

Use existing Mario Hello World test case for content:
- See `docs/test-cases/mario-hello.md` Business Layer Input

### Functional Layer

Use existing Mario Hello World test case for content:
- See `docs/test-cases/mario-hello.md` Functional Layer Input

### Stack Layer

Use existing Mario Hello World test case for content:
- See `docs/test-cases/mario-hello.md` Stack Layer Input

### Infrastructure Layer

Use existing Mario Hello World test case for content:
- See `docs/test-cases/mario-hello.md` Infrastructure Layer Input

### Coverage Layer

Use existing Mario Hello World test case for content:
- See `docs/test-cases/mario-hello.md` Coverage Layer Input

---

## Test Validation Criteria

**Orchestration capability validated when:**
- [ ] Development agent automatically generates business, functional, stack specs
- [ ] Deployment agent automatically generates infrastructure specs
- [ ] Validation agent automatically generates coverage specs
- [ ] User workflow simplified from multi-step to single command per phase
- [ ] Specs generated in correct dependency order (business → functional → stack)
- [ ] Agent invocations tracked with input/output context
- [ ] Progress visible to user during orchestration
- [ ] Error context preserved when agent invocations fail
- [ ] `--regen` flag triggers spec regeneration
- [ ] Insufficient session context detected and reported
- [ ] Partial spec scenarios handled correctly

**Primary Success Metric:**
User can run `/smaqit.development` from empty project with filled prompts and get working application without manually invoking spec agents.

---

## Test Execution Notes

**Time Estimate:** 30-45 minutes (8 scenarios)

**Prerequisites:**
- smaqit v0.7.0-beta or later (Task 073 implementation)
- Go toolchain installed
- Test project isolation (use `installer/test/` directory)

**Failure Documentation:**
- Log complete console output
- Capture agent invocation sequence
- Document session context requirements used
- Note exact error messages
- Record workflow state at failure point

**Success Documentation:**
- Confirm workflow simplification achieved
- Verify spec generation automatic
- Validate progress tracking visible
- Confirm error handling comprehensive
