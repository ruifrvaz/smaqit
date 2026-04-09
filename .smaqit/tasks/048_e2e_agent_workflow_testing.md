# End-to-End Agent Workflow Testing

**Status:** New  
**Created:** 2026-01-03  
**Agent:** smaqit.user-testing  
**Related:** Task 045 (CLI Infrastructure Validation), Task 047 (Incremental Processing), Task 014 (Stateful Specifications)

## Description

Comprehensive end-to-end testing of the complete smaqit workflow using real agent invocations to validate the full user experience from project initialization through validation. This tests all 8 agents, all 9 prompts, and all CLI commands in an incremental development pattern.

**Test Case:** Mario Hello World Console Application (`docs/test-cases/mario-hello.md`)

## Context

Task 045 Phase 2 validated CLI infrastructure (plan/status commands, frontmatter parsing, status filtering) through manual spec manipulation. Task 048 validates the complete user experience by invoking agents via prompts to generate specifications, implement features, deploy applications, and validate results.

**What This Tests:**
- All 8 agents (5 specification + 3 implementation)
- All 9 prompts (5 layer + 3 phase + 1 orchestrator)
- All CLI commands (init, status, plan, validate)
- Incremental workflow pattern (add features iteratively)
- Phase-by-phase execution with validation
- Real agent-generated specs and code
- Application functionality (not just file existence)

**Difference from Task 045:**

| Aspect | Task 045 Phase 2 | Task 048 E2E |
|--------|------------------|--------------|
| **Scope** | CLI infrastructure | Full agent workflows |
| **Spec Creation** | Manual (cat/sed) | Agent-generated via prompts |
| **Implementation** | Simulated (status changes) | Real (agents build code) |
| **Validation** | File existence | Application functionality |
| **Duration** | ~30 minutes | ~2 hours |
| **Test Type** | Infrastructure unit test | Integration/E2E test |

## Future Testing Agent Note

**Before starting:** Read this entire task file to understand the 7-phase workflow. This is a long-running test (1.5-2.25 hours). Plan accordingly and consider pausing between phases for user review.

**Key workflow:** You will coordinate testing by instructing the user to execute prompts (e.g., "Type `/smaqit.business` and paste these requirements"). You cannot invoke prompts programmatically—you guide the user through the workflow and validate outcomes.

## Test Phases

### Phase 1: Initial Feature (Core Greeting)

**Objective:** Establish baseline application with single feature

**Workflow:**
1. Build installer and initialize test project
2. Create Business prompt with minimal requirements:
   - Single use case: Greet user
   - Actor: Mario Fan
   - Goal: Display greeting
3. Instruct user: "Type `/smaqit.business` and paste requirements"
4. Wait for user confirmation: "Type 'done' when spec generated"
5. Validate: Business spec created with correct frontmatter
6. Create Functional prompt with core behavior:
   - Display text greeting
   - Single output format
7. Instruct user: "Type `/smaqit.functional` and paste requirements"
8. Validate: Functional spec created, references Business spec
9. Create Stack prompt with technology choice
10. Instruct user: "Type `/smaqit.stack` and paste requirements"
11. Validate: Stack spec created, references upstream specs
12. Check: `smaqit status` shows Phase 1 with 3 draft specs
13. Instruct user: "Type `/smaqit.development` to implement"
14. Validate: Application builds, specs show `status: implemented`
15. Test: Run application, verify greeting displays
16. Check: `smaqit status` shows Phase 1 complete

**Success Criteria:**
- [ ] Business spec generated with `BUS-` requirement IDs
- [ ] Functional spec generated with `FUN-` IDs, references Business
- [ ] Stack spec generated with `STK-` IDs, references Functional
- [ ] All specs have correct frontmatter (id, status, created, prompt_version)
- [ ] `smaqit plan --phase=develop` returned all 3 specs before implementation
- [ ] Development agent updated specs to `status: implemented`
- [ ] Application compiles without errors
- [ ] Application runs and displays greeting correctly
- [ ] `smaqit status` shows "✓ Complete" for Phase 1
- [ ] All 3 layers present (business, functional, stack)

**Expected Duration:** 20-30 minutes

---

### Phase 2: Add Feature (ASCII Art)

**Objective:** Validate incremental spec generation and implementation

**Workflow:**
1. Verify `smaqit status` shows Phase 1 complete
2. Add ASCII art requirement to Business prompt:
   - New use case or enhancement to existing
   - Mario character visual representation
3. Instruct user: "Type `/smaqit.business` and paste updated requirements"
4. Validate: New spec created OR existing spec unchanged (check `prompt_version`)
5. Check: `smaqit plan --phase=develop` shows new/changed specs only
6. Add ASCII rendering to Functional prompt:
   - ASCII art data structure
   - Rendering behavior
7. Instruct user: "Type `/smaqit.functional` and paste requirements"
8. Validate: Functional spec generated for ASCII feature
9. Update Stack prompt if needed (ASCII libraries)
10. Instruct user: "Type `/smaqit.stack` if stack changes needed"
11. Check: `smaqit plan --phase=develop` shows only draft specs
12. Instruct user: "Type `/smaqit.development` to implement"
13. Validate: Only draft/failed specs processed (incremental behavior)
14. Test: Run application, verify ASCII art displays
15. Regression test: Verify original greeting still works
16. Check: `smaqit status` updated correctly

**Success Criteria:**
- [ ] Incremental spec generation works (existing specs unchanged)
- [ ] `smaqit plan` returns only new/draft specs
- [ ] Development agent processes only required specs
- [ ] New feature implemented correctly
- [ ] Existing functionality preserved (no regression)
- [ ] Status shows correct spec counts

**Expected Duration:** 10-15 minutes

---

### Phase 3: Add Feature (Color Output)

**Objective:** Validate multiple incremental additions

**Workflow:**
1. Add color requirement to Business prompt:
   - Mario colors (red hat, blue overalls)
   - Visual appeal goal
2. Generate/update specs through all 3 layers (business → functional → stack)
3. Verify `smaqit plan` continues filtering correctly
4. Implement color feature via Development agent
5. Test: Run application, verify colored output
6. Integration test: All features work together (greeting + ASCII + color)

**Success Criteria:**
- [ ] Multiple incremental additions work correctly
- [ ] All features integrate properly
- [ ] No regressions in previous features
- [ ] Plan command filtering remains accurate

**Expected Duration:** 10-15 minutes

---

### Phase 4: Deploy Phase

**Objective:** Validate infrastructure specification and deployment

**Workflow:**
1. Create Infrastructure prompt:
   - Deployment target (e.g., local Docker, AWS Lambda, simple script)
   - Runtime requirements
   - Health check criteria
2. Instruct user: "Type `/smaqit.infrastructure` and paste requirements"
3. Validate: Infrastructure spec created, references Phase 1 specs
4. Check: `smaqit status` shows Phase 2 in progress
5. Instruct user: "Type `/smaqit.deployment` to deploy"
6. Validate: Deployment succeeds, specs updated to `status: deployed`
7. Test: Application accessible in target environment
8. Test: Health checks pass
9. Check: `smaqit status` shows Phase 2 complete

**Success Criteria:**
- [ ] Infrastructure spec references all Phase 1 specs
- [ ] Infrastructure spec has `INF-` requirement IDs
- [ ] Deployment executes successfully
- [ ] Application runs in deployed environment
- [ ] Specs show `status: deployed` with timestamp
- [ ] Phase 2 shows "✓ Complete"

**Expected Duration:** 15-20 minutes

**Note:** Deployment complexity depends on target. Consider simple local deployment (Docker, shell script) to keep test manageable.

---

### Phase 5: Validate Phase

**Objective:** Validate coverage specification and testing

**Workflow:**
1. Create Coverage prompt:
   - Test scenarios for all features
   - Acceptance criteria to verify
   - Test types (integration, E2E)
2. Instruct user: "Type `/smaqit.coverage` and paste requirements"
3. Validate: Coverage spec created, maps all upstream requirements
4. Check: Coverage spec references all `BUS-`, `FUN-`, `STK-`, `INF-` IDs
5. Instruct user: "Type `/smaqit.validation` to run tests"
6. Validate: Tests execute against deployed system
7. Check: Validation report generated in `.smaqit/reports/`
8. Check: Specs updated to `status: validated`
9. Check: All acceptance criteria tested or flagged as untestable
10. Check: `smaqit status` shows Phase 3 complete

**Success Criteria:**
- [ ] Coverage spec maps all testable requirements by ID
- [ ] Coverage spec has `COV-` IDs
- [ ] Tests execute successfully
- [ ] Validation report contains results per requirement
- [ ] Spec coverage percentage calculated
- [ ] Specs show `status: validated` with timestamp
- [ ] Phase 3 shows "✓ Complete"
- [ ] All phases now complete

**Expected Duration:** 15-20 minutes

---

### Phase 6: Failed Spec Recovery

**Objective:** Validate failure handling and recovery workflow

**Workflow:**
1. Manually mark one spec as `status: failed` (e.g., `specs/functional/greeting.md`)
2. Check: `smaqit plan --phase=develop` includes failed spec
3. Check: `smaqit status` shows Phase 1 "In progress" with failed count
4. Simulate fix: Change spec back to `status: implemented`
5. Optional: Re-run Development agent to demonstrate recovery
6. Validate: Recovery workflow completes successfully

**Success Criteria:**
- [ ] Failed specs correctly identified by `smaqit plan`
- [ ] Status command shows accurate failed spec counts
- [ ] Agents can reprocess failed specs
- [ ] Recovery restores phase completion status

**Expected Duration:** 5-10 minutes

---

### Phase 7: Force Regeneration

**Objective:** Validate --regen flag with agents

**Workflow:**
1. Check: `smaqit plan --phase=develop` returns empty (all implemented)
2. Test: `smaqit plan --phase=develop --regen` returns all specs
3. Optional: Have agent regenerate one spec to verify forced regeneration works

**Success Criteria:**
- [ ] --regen flag returns all specs regardless of status
- [ ] Agents can process specs in regen mode if needed

**Expected Duration:** 5 minutes

---

## Success Criteria Summary

**Must Pass:**
- [ ] All 7 phases execute successfully
- [ ] All agents produce valid output
- [ ] All CLI commands work as expected
- [ ] Incremental processing validated end-to-end
- [ ] Phase transitions work correctly (draft → implemented → deployed → validated)
- [ ] Application works as specified at each stage
- [ ] All spec frontmatter updated correctly throughout lifecycle
- [ ] Status display accurate throughout all phases
- [ ] Cross-layer references valid (Functional→Business, Stack→Functional, etc.)
- [ ] Traceability maintained (all requirements have IDs, coverage maps them)

**Deliverables:**
- [ ] E2E test report: `docs/user-testing/YYYY-MM-DD_e2e-agent-workflow-testing.md`
- [ ] Working Mario Hello application in test project
- [ ] Complete set of specs across all 5 layers
- [ ] Validation report from validation phase
- [ ] Screenshots or output samples showing functionality
- [ ] Task 048 marked complete in PLANNING.md

## Test Environment

**Setup:**
```bash
cd installer
make build
mkdir -p test/e2e-mario-$(date +%Y%m%d-%H%M%S)
cd test/e2e-mario-*
../../dist/smaqit init
```

**Cleanup:** Keep test project for inspection after report generation, or remove if space constrained

**Version:** Should test with latest build (v0.5.0-beta or later)

## Painpoints to Watch For

**Document these if encountered:**

1. **Agent context limits** - Mario test case designed small, but watch for token overflow
2. **Prompt clarity** - Are prompt templates clear enough for users?
3. **Agent errors** - Do agents handle edge cases gracefully?
4. **Status display lag** - Does status update immediately or require refresh?
5. **Deployment complexity** - Is deployment too complex for testing?
6. **Test determinism** - Do agents generate consistent structures (not identical code)?
7. **Incremental behavior** - Do agents correctly skip implemented specs?
8. **Cross-layer coherence** - Do agents validate upstream references?
9. **Error messages** - Are failures actionable for users?
10. **Time investment** - Is 2-hour test acceptable for validation?

## Report Format

Use template structure from Task 045 Phase 2 report:
- Test Information (date, version, OS, duration)
- Executive Summary (PASS/FAIL with rationale)
- Phase-by-phase results (detailed workflow, evidence, checkpoints)
- Painpoints Identified (blockers, issues, UX friction, performance)
- Recommendations (for release, for enhancements, for testing)
- Conclusion (release readiness assessment)

**Evidence Requirements:**
- Command outputs for all CLI commands
- Screenshots/samples of application output
- Spec file samples showing frontmatter evolution
- Validation report contents
- Error messages if failures occur

## Related Tasks

**Depends on:**
- Task 045: CLI infrastructure validated
- Task 047: Incremental processing implemented
- Task 014: Stateful specifications infrastructure

**Validates:**
- Complete smaqit workflow from user perspective
- All agents working correctly
- All prompts providing adequate guidance
- Incremental development pattern functional
- Phase transitions smooth and intuitive

**Enables:**
- Confidence for v0.5.0 stable release
- User onboarding documentation
- Tutorial content based on test execution
- Bug identification before user adoption

## Follow-Up Actions

**If tests pass:**
- Tag v0.5.0 stable release
- Create user tutorial based on test workflow
- Document successful patterns for users
- Update README with incremental workflow examples
- Consider Task 025 (CI/CD integration) next

**If tests fail:**
- Document failure point (which phase, which agent)
- Classify issue (infrastructure bug, agent bug, UX issue, test issue)
- File bug report or enhancement task
- Determine if blocker for v0.5.0 or enhancement for v0.6.0
- Iterate until resolution, then re-test

## Notes

**Test Philosophy:**
- This tests the **user experience**, not just infrastructure
- Agents may generate different code each run—validate **behavior**, not **implementation**
- Focus on **workflow smoothness**: Are prompts clear? Is guidance helpful? Are errors actionable?
- Document **painpoints objectively**: What happened, not why (leave root cause analysis for later)

**Deployment Consideration:**
For Phase 4 (Deploy), choose simplest viable deployment to keep test manageable:
- **Option 1:** Local shell script (run application locally as "deployed")
- **Option 2:** Docker container (docker run)
- **Option 3:** Cloud function (AWS Lambda, Google Cloud Function) - more realistic but slower

Recommendation: Start with Option 1 or 2 for speed, use Option 3 in future iterations for realism.

**Pause Points:**
Consider pausing for user review after:
- Phase 1 complete (baseline established)
- Phase 3 complete (all features implemented)
- Phase 5 complete (full workflow validated)

This allows user to inspect progress and decide whether to continue.
