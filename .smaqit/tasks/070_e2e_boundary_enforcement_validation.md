# Task 070: E2E Boundary Enforcement Validation

**Status:** New  
**Priority:** High  
**Created:** 2026-01-21  
**Updated:** 2026-02-08  
**Context:** Verify Task 068 (System Actor removal) boundary enforcement AND Task 078 (Assessment skill integration) work in practice

## Purpose

Validate two critical framework capabilities:

1. **Boundary enforcement** - Business Agent respects layer boundaries when given input containing scope violations, filtering or flagging technical concerns properly

2. **Assessment skill integration** - All agents automatically invoke assessment skill when detecting ambiguous requirements, conflicting inputs, insufficient detail, or complex planning scenarios

## Background

Task 068 removed the System Actor pattern and strengthened Business layer boundaries by:
- Removing System Actor from framework, templates, and agents
- Refining MUST NOT directives to focus on describing HOW (behaviors/mechanisms)
- Adding NFR stakeholder guidance (named stakeholders, not generic "System")
- Creating wiki documentation explaining proper layer separation

The test case `docs/test-cases/mario-hello.md` deliberately contains Business layer input with scope violations to test boundary enforcement:
- **Client Organizations actor** (infrastructure concern disguised as stakeholder)
- **"sees Mario greeting with character representation"** (describes visual rendering)
- **"sees iconic Mario catchphrase"** (describes what is displayed)
- **Technical details in flows** (HOW features work vs WHAT stakeholders need)

## Test Objectives

1. **Verify boundary filtering** - Business Agent should either:
   - Filter out technical concerns and produce clean Business spec, OR
   - Flag boundary violations and request clarification

2. **Verify System Actor absence** - Generated spec should NOT contain System Actor

3. **Verify technical verb absence** - Flows should NOT contain:
   - Behavioral verbs: display, render, output, execute, process, detect, handle
   - Technical artifacts: console, terminal, encoding, color support

4. **Verify proper actor usage** - Spec should use named stakeholders (Mario Fan, Accessibility Advocate) not generic "System"

5. **Verify other agents unaffected** - Confirm System Actor removal only impacted Business layer

6. **Verify assessment skill integration** - Agents automatically invoke assessment skill when detecting:
   - Ambiguous requirements (multiple interpretations possible)
   - Conflicting inputs (prompt vs upstream specs, or within prompt)
   - Insufficient detail (cannot proceed without assumptions)
   - Complex multi-part work (requires explicit planning)

## Implementation Checklist

### 1. Run Business Spec Generation

- [ ] Navigate to test environment: `mkdir -p /tmp/smaqit-test-070 && cd /tmp/smaqit-test-070`
- [ ] Initialize smaqit: `smaqit init` (or manually copy framework + agents)
- [ ] Copy Business layer input from `docs/test-cases/mario-hello.md` to `.github/prompts/smaqit.business.prompt.md`
- [ ] Invoke Business Agent: `/smaqit.business`
- [ ] Review generated spec in `specs/business/`

### 2. Analyze Generated Spec

**Check for System Actor:**
- [ ] Search spec for "System" actor in Actors table
- [ ] Verify NO System Actor present

**Check for technical verbs in flows:**
- [ ] Search Main Flow for: display, render, output, execute, process, detect, handle
- [ ] Search Alternative Flows for technical verbs
- [ ] Verify flows describe outcomes, not mechanisms

**Check for technical artifacts:**
- [ ] Search spec for: console, terminal, screen, encoding, color support
- [ ] Verify NO technical artifact references

**Check for proper actor usage:**
- [ ] Verify named stakeholders used (Mario Fan, Accessibility Advocate)
- [ ] Verify actors have clear goals/motivations
- [ ] Verify NFR stakeholders properly captured if applicable

**Check acceptance criteria:**
- [ ] Verify criteria are measurable, observable, unambiguous
- [ ] Verify NO technical implementation details in criteria

### 3. Review Agent Behavior

**If agent filtered violations:**
- [ ] Document what was filtered out
- [ ] Verify filtered content would belong in Functional/Stack layers
- [ ] Verify no information loss (stakeholder needs preserved)

**If agent flagged violations:**
- [ ] Document what violations were flagged
- [ ] Verify flags correctly identified boundary crossings
- [ ] Verify agent requested appropriate clarification

**If agent passed violations through:**
- [ ] Document what violations appeared in spec
- [ ] Identify gaps in boundary enforcement
- [ ] Flag for framework refinement

### 4. Review Other Agents

- [ ] Search all 8 agents for "System Actor" references: `grep -r "System Actor" agents/`
- [ ] Verify only Business agent was affected by removal
- [ ] Confirm no broken references in other layers

### 5. Test Assessment Skill Integration (Task 078 Validation)

**Test automatic invocation with ambiguous input:**
- [ ] Create prompt with ambiguous requirements (multiple interpretations)
- [ ] Invoke any specification agent (Business, Functional, Stack)
- [ ] Verify agent invokes `.github/skills/assessment/` automatically
- [ ] Verify skill returns structured output with 6 required components
- [ ] Verify agent incorporates assessment results before proceeding

**Test automatic invocation with conflicting input:**
- [ 7. Update Task Status

- [ ] Mark Task 068 as completed in `docs/tasks/PLANNING.md`
- [ ] Add Task 070 to Completed table
- [ ] Document outcome in Task 068 file
- [ ] Document assessment skill validation results in Task 07fication rather than proceeding with conflicting spec

**Test automatic invocation with insufficient detail:**
- [ ] Create prompt with insufficient detail (missing key information for spec generation)
- [ ] Invoke specification agent
- [ ] Verify agent detects insufficiency and invokes assessment skill
- [ ] Verify skill identifies gaps in "Current State Findings" section
- [ ] Verify agent requests additional information before generating spec

**Test automatic invocation with complex planning:**
- [ ] Create prompt with multi-tier, multi-layer work requiring explicit planning
- [ ] Invoke implementation agent (Development, Deployment, Validation)
- [ ] Verify agent invokes assessment skill for complex scenarios
- [ ] Verify skill generates execution plan in "Execution Plan" section
- [ ] Verify agent follows structured plan rather than ad-hoc execution

**Test skill output consumption:**
- [ ] Verify agents correctly parse skill's structured output
- [ ] Verify "Approval Status" component controls agent continuation
- [ ] Verify "Flagged Problems" halt execution when critical
- [ ] Verify "Execution Plan" guides agent implementation steps

**Document assessment skill behavior:**
- [ ] Record which trigger conditions successfully invoke skill
- [ ] Document any false positives (skill invoked unnecessarily)
- [ ] Document any false negatives (skill NOT invoked when it should be)
- [ ] Identify gaps in automatic detection logic

### 6. Document Results

- [ ] Create test report: `docs/user-testing/2026-01-21_task-068-boundary-validation.md`
- [ ] Include:
  - Generated spec (or relevant excerpts)
  - Analysis of boundary enforcement
  - Agent behavior assessment (filtered/flagged/passed-through)
  - Comparison with v0.6.0-beta results (Painpoint #2 status)
- [ ] Update `docs/user-testing/2026-01-17_v0.6.0-beta-validation.md` noting Painpoint #2 addressed

### 6. Update Task Status

- [ ] Mark Task 068 as completed in `docs/tasks/PLANNING.md`
- [ ] Add Task 070 to Completed table
- [ ] Document outcome in Task 068 file

## Success Criteria

**Boundary enforcement working:**
- [ ] Generated Business spec has NO System Actor
- [ ] Generated Business spec has NO technical verbs in flows
- [ ] Generated Business spec has NO technical artifacts referenced
- [ ] Named stakeholders properly used for all actors including NFRs
- [ ] Acceptance criteria are business-level (outcomes, not mechanisms)

**Agent behavior acceptable:**
- [ ] Agent either filtered violations cleanly OR flagged them clearly
- [ ] No loss of stakeholder needs (even if technical details filtered)
- [ ] Clear guidance provided if violations flagged

**Framework integrity:**
- [ ] Only Business agent affected by System Actor removal
- [ ] No broken references in other layers

**Assessment skill integration working (Task 078 validation):**
- [ ] Agents automatically invoke skill for ambiguous inputs
- [ ] Agents automatically invoke skill for conflicting requirements
- [ ] Agents automatically invoke skill for insufficient detail
- [ ] Agents automatically invoke skill for complex planning scenarios
- [ ] Skill returns properly structured output (6 components present)
- [ ] Agents consume skill output and adjust behavior accordingly
- [ ] No false positives (skill invoked when NOT needed)
- [ ] No false negatives (skill NOT invoked when needed)
- [ ] Boundary enforcement consistent with directives

## Expected Outcomes
- **Task 078:** Iterative Assessment Before Execution (assessment skill complete, functional validation needed)

**Best case:** Business Agent filters out Client Organizations actor and technical flow details, generating clean Business spec focused on stakeholder outcomes (Mario Fan experience, accessibility requirements).

**Acceptable:** Business Agent flags boundary violations and requests clearer business requirements, refusing to proceed with polluted input.

**Failure:** Business Agent generates spec with System Actor, technical verbs (display, render, detect), and Stack artifacts (console, encoding, color support), indicating boundary enforcement not working.

## Related Tasks

- **Task 068:** Remove System Actor from Business Layer (framework changes complete)
- **Task 067:** v0.6.0-beta Validation Testing (where Painpoint #2 was discovered)

## Notes

This is a validation-only task. If boundary enforcement fails, new tasks will be created to address gaps. The test case deliberately contains violations to stress-test the boundaries—passing violations through would indicate framework needs strengthening.

**Test environment:** Can use temporary directory (`/tmp/smaqit-test-070`) or installer test directory. Does not affect main repository state.
