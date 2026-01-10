# Task 059: E2E Regression Test - Verify Issue Fixes

**Status:** new  
**Priority:** High  
**Created:** 2026-01-09  
**Agent:** smaqit.user-testing  
**Related:** Task 048 (Original E2E test), Tasks 049-056 & 058 (Issue fixes)

## Description

Regression test of the smaqit E2E workflow focusing on validation of the 8 critical issues discovered in Task 048 (2026-01-04 E2E test). This test verifies that fixes implemented in Tasks 049-056 and 058 successfully resolve the identified problems while maintaining overall workflow functionality.

**Test Case:** Mario Hello World Console Application (`docs/test-cases/mario-hello.md`)

## Context

Task 048 performed comprehensive E2E testing that discovered 8 critical issues:
1. **Issue 1:** Agent context pollution between layers → Fixed by Task 056
2. **Issue 2:** Stack agent includes implementation code → Fixed by Task 054
3. **Issue 3:** Stack spec duplication → Fixed by Task 055
4. **Issue 4:** Development agent ignores CLI → Fixed by Task 049
5. **Issue 5:** Validation agent ignores CLI → Fixed by Task 051
6. **Issue 6:** Deployment agent ignores CLI → Fixed by Task 052
7. **Issue 7:** Validation agent doesn't update frontmatter → Fixed by Task 053
8. **Issue 8:** Agents don't update checkboxes → Fixed by Task 058

Additionally, recent session (034 - Foundation Reference Pattern Refinement) identified:
9. **Issue 9:** Foundation vs Feature spec distinction → Addressed in Session 034

All fix tasks are marked complete. This regression test validates the fixes work in practice.
### Primary Objectives (Issue Validation)
1. Verify agents respect CLI authority (`smaqit plan --phase=X`)
2. Verify agents update spec frontmatter through phase transitions
3. Verify agents update acceptance criteria checkboxes
4. Verify Stack agent excludes implementation code from specs
5. Verify specification agents avoid duplication (single source of truth)
6. Verify context pollution mitigation guidance works
7. Verify agents distinguish foundation specs from feature specss
4. Verify Stack agent excludes implementation code from specs
5. Verify specification agents avoid duplication (single source of truth)
6. Verify context pollution mitigation guidance works

### Secondary Objectives (Workflow Validation)
7. Validate core workflow still functions (Phase 1 → Phase 3, skip Phase 2)
8. Validate incremental spec generation
9. Validate application functionality
10. Validate CLI commands accuracy

## Test Phases

### Phase 0: Setup and Baseline

**Objective:** Establish test environment and verify installer

**Workflow:**
1. Build installer: `cd installer && make build`
2. Verify version: `./dist/smaqit --version`
3. Create test project: `mkdir -p test/e2e-regression-$(date +%Y%m%d-%H%M%S) && cd test/e2e-regression-*`
4. Initialize: `../../dist/smaqit init`
5. Verify initialization: Check `.smaqit/`, `.github/agents/`, `.github/prompts/`, `specs/` structure
6. Check initial status: `../../dist/smaqit status`

**Success Criteria:**
- [ ] Installer builds without errors
- [ ] Test project initialized successfully
- [ ] Status shows "Not started" for all phases

**Expected Duration:** 5 minutes

---

### Phase 1A: Business Layer (Issue 1, 5 Focus)

**Objective:** Validate single specification agent execution and context handling

**Issue 1 Validation:** Test context pollution mitigation (Task 056 fix)
- Check if agent documentation includes context awareness guidance
- Monitor agent behavior for inappropriate cross-layer context references

**Issue 5 Validation:** N/A (Business layer doesn't involve Issue 5)

**Workflow:**
1. Create Business prompt with Mario greeting requirements (minimal, single use case)
2. Instruct user: "Type `/smaqit.business` and paste requirements"
3. Wait for user confirmation: "Type 'done' when spec generated"
4. Validate spec creation: `specs/business/*.md`
5. Check spec content: No code examples, proper frontmatter, acceptance criteria with `BUS-` IDs
6. Verify status: `../../dist/smaqit status`

**Success Criteria:**
- [ ] Business spec generated with correct structure
- [ ] Spec contains no implementation details
- [ ] Frontmatter includes: id, status: draft, created, prompt_version
- [ ] Status shows Phase 1 "⚙ In progress (1 pending)"

**Expected Duration:** 10 minutes

---

### Phase 1B: Functional Layer (Issue 1, 3 Focus)

**Objective:** Validate cross-layer references and single source of truth

**Issue 1 Validation:** Monitor for Business context carryover
**Issue 3 Validation:** Not yet testable (requires incremental addition)

**Workflow:**
1. Create Functional prompt with console behavior requirements
2. Instruct user: "Type `/smaqit.functional` and paste requirements"
3. Validate spec references Business spec correctly
4. Check for proper separation (no Business details duplicated)

**Success Criteria:**
- [ ] Functional spec created with `FUN-` IDs
- [ ] Spec references Business spec using Implements/Enables
- [ ] No duplication of Business requirements
- [ ] Status shows Phase 1 "⚙ In progress (2 pending)"

**Expected Duration:** 10 minutes

---
### Phase 1C: Stack Layer (Issue 2, 3, 9 Focus)

**Objective:** Validate Stack agent excludes implementation code and recognizes foundation spec pattern

**Issue 2 Validation (CRITICAL):** Stack spec MUST NOT contain code examples
- Check for absence of code blocks with implementation patterns
- Verify "Architecture Notes" or similar sections don't include code
- Confirm spec describes WHAT technologies, not HOW to use them

**Issue 3 Validation:** Not yet testable (requires incremental addition)

**Issue 9 Validation:** Foundation spec pattern recognition
- Initial Stack spec should be foundation spec (serves greeting feature)
- Check for appropriate References section structure
- Verify spec uses Enables pattern for multi-requirement foundation
**Issue 3 Validation:** Not yet testable (requires incremental addition)

**Workflow:**
1. Create Stack prompt with Python console technology choice
**Success Criteria:**
- [ ] Stack spec created with `STK-` IDs
- [ ] **[Issue 2] Spec contains NO code examples or implementation patterns**
- [ ] Spec describes technology choices with rationale only
- [ ] No "Architecture Notes" with code blocks
- [ ] **[Issue 9] Spec uses Enables pattern (foundation serving multiple features)**
- [ ] **[Issue 9] References section shows foundation relationship to upstream specs**
- [ ] Status shows Phase 1 "⚙ In progress (3 pending)"

**Expected Duration:** 10 minutesO code examples or implementation patterns**
- [ ] Spec describes technology choices with rationale only
- [ ] No "Architecture Notes" with code blocks
- [ ] Status shows Phase 1 "⚙ In progress (3 pending)"

**Expected Duration:** 10 minutes

---

### Phase 1D: Development Phase (Issue 4, 7, 8 Focus)

**Objective:** Validate Development agent respects CLI and updates specs correctly

**Issue 4 Validation (CRITICAL):** Development agent MUST use `smaqit plan --phase=develop`
- Monitor agent execution for CLI command invocation
- Verify agent processes only specs returned by plan command
- Confirm agent doesn't manually filter specs

**Issue 7 Validation:** N/A (Development doesn't set `status: validated`)

**Issue 8 Validation (CRITICAL):** Development agent MUST update checkboxes
- After implementation, check Business, Functional, Stack specs
- Verify acceptance criteria checkboxes changed: `[ ]` → `[x]` or `[!]`
- Confirm checkbox updates match actual implementation status

**Workflow:**
1. Check pre-implementation state: `../../dist/smaqit plan --phase=develop`
2. Instruct user: "Type `/smaqit.development` to implement"
3. **CRITICAL CHECK:** Watch for agent invoking `smaqit plan --phase=develop`
4. Validate implementation completes successfully
5. **CRITICAL CHECK:** Open all 3 Phase 1 specs and verify:
   - Frontmatter updated: `status: implemented`, `implemented: [timestamp]`
   - Acceptance criteria checkboxes updated: `[ ]` → `[x]` or `[!]`
6. Test application: Run and verify greeting displays
7. Check status: `../../dist/smaqit status` shows Phase 1 "✓ Complete"

**Success Criteria:**
- [ ] **[Issue 4] Agent invoked `smaqit plan --phase=develop` before processing**
- [ ] Agent processed only specs returned by plan command
- [ ] All 3 specs updated to `status: implemented` with timestamp
- [ ] **[Issue 8] All acceptance criteria checkboxes updated in Business spec**
- [ ] **[Issue 8] All acceptance criteria checkboxes updated in Functional spec**
- [ ] **[Issue 8] All acceptance criteria checkboxes updated in Stack spec**
- [ ] Application compiles and runs correctly
- [ ] Status shows Phase 1 "✓ Complete"

**Expected Duration:** 20 minutes

### Phase 2: Incremental Addition (Issue 3, 9 Focus)

**Objective:** Validate single source of truth and foundation vs feature spec distinction

**Issue 3 Validation (CRITICAL):** Agents MUST NOT duplicate existing spec information
- Add Luigi feature (requires new character, extends existing console app)
- Monitor whether Stack agent updates existing spec vs creates duplicate
- Verify agent uses Foundation Reference pattern if creating new spec
- Confirm no duplication of base Python requirements

**Issue 9 Validation (CRITICAL):** Agents distinguish foundation from feature specs
- Luigi addition may create feature spec that references foundation
- Verify agent recognizes when to create feature spec vs update foundation
- Check for Foundation Reference section if feature spec created
- Confirm feature spec uses Implements pattern (1:1 with upstream requirement)vs creates duplicate
- Verify agent uses Foundation Reference pattern if creating new spec
- Confirm no duplication of base Python requirements

**Workflow:**
9. **CRITICAL CHECK:** If new Stack spec created:
   - Verify it references base Stack spec for shared requirements
   - Confirm no duplication of Python version, build tools, etc.
   - Check for "Foundation Reference" section with proper references
   - **[Issue 9] Verify feature spec uses Implements (1:1) not Enables (1:many)**
   - **[Issue 9] Confirm foundation spec remains separate with Enables pattern**
10. Check plan filtering: `../../dist/smaqit plan --phase=develop`s"
6. Validate: New/updated Functional spec references existing specs appropriately
7. Add any Stack changes to Stack prompt (e.g., character data file handling)
8. Instruct user: "Type `/smaqit.stack` if needed"
9. **CRITICAL CHECK:** If new Stack spec created:
**Success Criteria:**
- [ ] **[Issue 3] No duplication of base requirements across Stack specs**
- [ ] **[Issue 3] New Stack spec uses Foundation Reference if created**
- [ ] **[Issue 9] Foundation spec uses Enables pattern (1:many with upstream)**
- [ ] **[Issue 9] Feature spec uses Implements pattern (1:1 with upstream) if created**
- [ ] **[Issue 9] Foundation Reference section present in feature spec if created**
- [ ] Incremental spec generation works (existing implemented specs unchanged)
- [ ] Plan command returns only draft/failed specs
- [ ] **[Issue 8] Checkboxes updated in new/changed specs**
- [ ] Luigi feature works correctly
- [ ] Original greeting still works (no regression)

**Expected Duration:** 20 minutesof base requirements across Stack specs**
- [ ] **[Issue 3] New Stack spec uses Foundation Reference if created**
- [ ] Incremental spec generation works (existing implemented specs unchanged)
- [ ] Plan command returns only draft/failed specs
- [ ] **[Issue 8] Checkboxes updated in new/changed specs**
- [ ] Luigi feature works correctly
- [ ] Original greeting still works (no regression)

**Expected Duration:** 20 minutes

---

### Phase 3: Infrastructure and Deployment (Issue 6 Focus)

**Objective:** Validate Deployment agent respects CLI authority

**Issue 6 Validation (CRITICAL):** Deployment agent MUST use `smaqit plan --phase=deploy`
- Monitor agent execution for CLI command invocation
- Verify agent processes only specs returned by plan command

**Workflow:**
1. Create Infrastructure prompt with local container deployment requirements:
   - Target: Local Docker container
   - Runtime: Python environment
   - Health check: Container runs, application accessible
   - Simple deployment (keep test manageable)
2. Instruct user: "Type `/smaqit.infrastructure` and paste requirements"
3. Validate Infrastructure spec created with `INF-` IDs
4. Check: `../../dist/smaqit plan --phase=deploy`
5. Instruct user: "Type `/smaqit.deployment` to deploy"
6. **CRITICAL CHECK:** Watch for agent invoking `smaqit plan --phase=deploy`
7. Validate deployment completes (container running)
8. Test: Access deployed application, verify functionality
9. **CRITICAL CHECK:** Open Infrastructure spec and verify:
   - Frontmatter updated: `status: deployed`, `deployed: [timestamp]`
   - Acceptance criteria checkboxes updated
10. Verify status: `../../dist/smaqit status` shows Phase 2 "✓ Complete"

**Success Criteria:**
- [ ] Infrastructure spec created with deployment requirements
- [ ] **[Issue 6] Agent invoked `smaqit plan --phase=deploy` before processing**
- [ ] Deployment executed successfully (container running)
- [ ] Application accessible in container
- [ ] Infrastructure spec updated to `status: deployed` with timestamp
- [ ] Acceptance criteria checkboxes updated in Infrastructure spec
- [ ] Status shows Phase 2 "✓ Complete"

**Expected Duration:** 20 minutes

---

### Phase 4: Coverage and Validation (Issue 5, 7, 8 Focus)

**Objective:** Validate Coverage/Validation agents respect CLI and update specs

**Issue 5 Validation (CRITICAL):** Validation agent MUST use `smaqit plan --phase=validate`
- Monitor agent execution for CLI command invocation
- Verify agent processes only specs returned by plan command

### Phase 5: Verification Summary

**Objective:** Collect evidence and document all issue validationsmestamp]`

**Issue 8 Validation (CRITICAL):** Validation agent MUST update checkboxes
- Check acceptance criteria in validated specs
- Verify checkboxes reflect test results: `[x]` for pass, `[!]` for fail

**Workflow:**
1. Create Coverage prompt with test requirements
2. Instruct user: "Type `/smaqit.coverage` and paste requirements"
3. Validate Coverage spec maps all upstream requirements
4. Check: `../../dist/smaqit plan --phase=validate`
5. Instruct user: "Type `/smaqit.validation` to run tests"
6. **CRITICAL CHECK:** Watch for agent invoking `smaqit plan --phase=validate`
7. Validate tests execute successfully
8. **CRITICAL CHECK:** Open ALL specs (Business, Functional, Stack, Coverage) and verify:
   - Frontmatter updated: `status: validated`, `validated: [timestamp]`
   - Acceptance criteria checkboxes reflect test results
9. Check validation report: `.smaqit/reports/validation-phase-report-*.md`
10. Verify status: `../../dist/smaqit status` shows Phase 3 "✓ Complete"

**Success Criteria:**
- [ ] Coverage spec created with comprehensive requirement mapping
- [ ] **[Issue 5] Agent invoked `smaqit plan --phase=validate` before processing**
- [ ] Tests executed successfully
- [ ] **[Issue 7] ALL validated specs updated to `status: validated` with timestamp**
- [ ] **[Issue 7] Business spec has validated status and timestamp**
- [ ] **[Issue 7] Functional spec has validated status and timestamp**
- [ ] **[Issue 7] Stack spec has validated status and timestamp**
- [ ] **[Issue 7] Coverage spec has validated status and timestamp**
- [ ] **[Issue 8] Checkboxes updated in all validated specs to reflect test results**
- [ ] Validation report generated with results
- [ ] Status shows Phase 3 "✓ Complete"
3. Document evidence of each issue fix:
   - Issue 1: Context pollution guidance present, no observed cross-contamination
   - Issue 2: Stack specs contain no code examples
   - Issue 3: No duplication across specs, Foundation Reference used correctly
   - Issue 4: Development agent used CLI (`smaqit plan --phase=develop`)
   - Issue 5: Validation agent used CLI (`smaqit plan --phase=validate`)
   - Issue 6: Deployment agent used CLI (`smaqit plan --phase=deploy`)
   - Issue 7: All validated specs have updated frontmatter
   - Issue 8: All checkboxes updated throughout workflow
   - Issue 9: Foundation vs feature specs distinguished correctly (Enables vs Implements)
**Workflow:**
1. Review all generated specs for issue compliance
3. Document evidence of each issue fix:
   - Issue 1: Context pollution guidance present, no observed cross-contamination
   - Issue 2: Stack specs contain no code examples
   - Issue 3: No duplication across specs, Foundation Reference used correctly
   - Issue 4: Development agent used CLI (`smaqit plan --phase=develop`)
   - Issue 5: Validation agent used CLI (`smaqit plan --phase=validate`)
**Success Criteria:**
- [ ] Evidence collected for all 9 issues
- [ ] All critical checks passed
- [ ] Application functions correctly
- [ ] CLI reports accurate state

**Expected Duration:** 10 minutes
- [ ] Evidence collected for all 8 issues
- [ ] All critical checks passed
- [ ] Application functions correctly
- [ ] CLI reports accurate state

**Expected Duration:** 10 minutes

### Critical Issue Validations (MUST PASS)
- [ ] **Issue 1:** No context pollution observed, guidance present
- [ ] **Issue 2:** Stack specs contain ZERO code examples
- [ ] **Issue 3:** No duplication of base requirements across specs
- [ ] **Issue 4:** Development agent invoked `smaqit plan --phase=develop`
- [ ] **Issue 5:** Validation agent invoked `smaqit plan --phase=validate`
- [ ] **Issue 6:** Deployment agent invoked `smaqit plan --phase=deploy`
- [ ] **Issue 7:** ALL validated specs updated with `status: validated` and timestamp
- [ ] **Issue 8:** Acceptance criteria checkboxes updated throughout all phases
- [ ] **Issue 9:** Foundation specs use Enables, feature specs use Implements, Foundation Reference present
- [ ] **Issue 5:** Validation agent invoked `smaqit plan --phase=validate`
- [ ] **Issue 6:** Deployment agent invoked `smaqit plan --phase=deploy`
- [ ] **Issue 7:** ALL validated specs updated with `status: validated` and timestamp
- [ ] **Issue 8:** Acceptance criteria checkboxes updated throughout all phases

### Workflow Validations (SHOULD PASS)
- [ ] All agents produce valid output
- [ ] Incremental processing works correctly
- [ ] Application builds and runs successfully
- [ ] Cross-layer references maintain coherence
- [ ] CLI commands report accurate state

## Test Environment

**Setup:**
```bash
cd installer
make build
mkdir -p test/e2e-regression-$(date +%Y%m%d-%H%M%S)
cd test/e2e-regression-*
../../dist/smaqit init
```

**Cleanup:** Keep test project for inspection after report generation

**Version:** v0.5.0-beta (post-fixes from Tasks 049-056, 058)

- [ ] Regression test report: `docs/user-testing/YYYY-MM-DD_e2e-regression-test.md`
- [ ] Working Mario + Luigi application
- [ ] Complete spec set with proper frontmatter state tracking
- [ ] Validation report showing test execution
- [ ] Evidence screenshots/samples for each of 9 issues
- [ ] Task 059 marked complete in PLANNING.mdn
- [ ] Evidence screenshots/samples for each of 8 issues
- [ ] Task 059 marked complete in PLANNING.md

## Issue Validation Checklist

For report, document each issue with PASS/FAIL and evidence:

| Issue | Fix Task | Validation Focus | Evidence Required |
|-------|----------|------------------|-------------------|
| 1 | 056 | Context pollution | Agent behavior observation, documentation check |
| 2 | 054 | Stack code exclusion | Spec content inspection for code blocks |
| 3 | 055 | No duplication | Cross-spec comparison, Foundation Reference usage |
| 6 | 052 | Deployment CLI | Agent execution log showing CLI invocation |
| 7 | 053 | Frontmatter updates | Spec frontmatter inspection for all layers |
| 8 | 058 | Checkbox updates | Acceptance criteria checkbox state changes |
| 9 | 034 | Foundation vs Feature | References section inspection (Enables vs Implements) |
| 7 | 053 | Frontmatter updates | Spec frontmatter inspection for all layers |
| 8 | 058 | Checkbox updates | Acceptance criteria checkbox state changes |

## Report Format

Use existing report template structure:
- Test Information (date, version, OS, duration)
- Standardized Checklist (setup, phases, issue validations)
- Execution Log (timestamped steps with critical checks highlighted)
- Issue Validation Results (PASS/FAIL per issue with evidence)
- Painpoints Identified (if any new issues discovered)
- Recommendations (release readiness, remaining work)
- Overall Result (PASS/FAIL with rationale)

**Critical:** For each issue, include:
- Issue number and description
- Fix task reference
- Validation method
- Evidence (screenshot, spec excerpt, CLI output)
- Result (PASS/FAIL/N/A)

## Regression Test Philosophy

**Focus:** Validate fixes work, not comprehensive workflow testing (Task 048 already did that)

**Efficiency:** 
- Skip deployment phase (Issue 6 validated separately if needed)
- Minimal feature set (Mario + Luigi, no additional complexity)
- Critical checks at key decision points (CLI usage, frontmatter updates, checkbox updates)

**Evidence-Driven:**
- Every issue requires concrete evidence (not assumption)
- Document exact locations (file paths, line numbers)
**Time Budget:** 2-2.5 hours (includes local container deployment)

## Pass/Fail Criteria

**PASS:** All 9 issues validate successfully

**PASS:** All 8 issues validate successfully

**FAIL:** Any issue validation fails, indicating fix didn't work or introduced regression

**PARTIAL:** Most issues pass but non-critical issues fail (document for follow-up)

## Related Tasks

**Validated fixes:**
- Task 049: Development Agent CLI Directive
- Task 050: Coverage Prompt (not directly tested here)
- Task 051: Validation Agent CLI Directive
- Task 052: Deployment Agent CLI Directive
- Task 055: Single Source of Truth Principle
- Task 056: Context Pollution Workaround
- Task 058: Checkbox Updates
- Session 034: Foundation Reference Pattern Refinement Truth Principle
- Task 056: Context Pollution Workaround
- Task 058: Checkbox Updates

**Depends on:**
- Task 048: Original E2E test that discovered issues

**Enables:**
- v0.5.0 release confidence
- Issue closure confirmation
- User adoption without known critical bugs

## Notes

**Why Regression Test Instead of Full E2E:**
- Full E2E already validated workflow (Task 048)
- This test validates specific fixes to discovered issues
- More efficient: focused validation vs comprehensive testing
- Provides evidence that fixes work without re-testing entire framework

**Deployment Approach:**
- Local Docker container deployment keeps test manageable
- Simple Python environment container (not complex cloud infrastructure)
- Validates Issue 6 (Deployment CLI directive) without excessive overhead
**Success Definition:**
- If all 9 issues PASS → Fixes verified, v0.5.0 ready
- If any issue FAILS → Fix didn't work, revisit corresponding task or session
- If new issues found → Document as new tasks, don't conflate with original 9

**Issue 9 Context:**
Session 034 unified Foundation Reference pattern across the framework:
- **Foundation specs** serve multiple upstream requirements (use Enables reference pattern)
- **Feature specs** implement single upstream requirement (use Implements reference pattern)
- Feature specs reference foundation specs via "Foundation Reference" section (not "Base Requirements")
- This distinction ensures proper traceability and prevents spec fragmentation
- Pattern extended to Stack and Infrastructure layers in Session 034
- If any issue FAILS → Fix didn't work, revisit corresponding task
- If new issues found → Document as new tasks, don't conflate with original 8

