# Validate Stateful Specifications Infrastructure

**Status:** Phase 1 Complete, Phase 2 Ready  
**Created:** 2026-01-03  
**Phase 1 Completed:** 2026-01-03  
**Phase 2 Starting:** 2026-01-03 (v0.5.0-beta)  
**Related:** Task 014 (Stateful Specifications Implementation), Task 047 (Incremental Processing Implementation)

## Description

Comprehensive validation and testing of the stateful specifications system introduced in Task 014. This task is split into two phases:

**Phase 1: Infrastructure Validation** ✅ Complete - Validate state tracking mechanisms (documentation review)
**Phase 2: Incremental Workflow Testing** 🚀 Ready - Test agents with incremental processing (runtime execution)

This task completed Phase 1 with all infrastructure validation passing. Phase 2 unblocked after Task 047 completion (incremental processing fully implemented).

## Context

Task 014 implemented:
- Spec-level state tracking via YAML frontmatter (`status`, `created`, `prompt_version`, timestamps)
- Phase-level state aggregation in `.smaqit/state.json` (spec counts)
- Agent state tracking directives (update frontmatter and state.json)
- CLI state display (`smaqit status` shows counts)

Critical gap identified: Agents may not yet implement incremental processing (skipping already-implemented specs). Phase 1 validates infrastructure via documentation review, Phase 2 will test incremental behavior via runtime execution.

## Phase 1: Infrastructure Validation

### Validation Scenarios

#### 1. Installer Build
- [ ] Installer compiles without errors (`make build`)
- [ ] Binary created in `dist/smaqit`
- [ ] Binary is executable

#### 2. Template Validation
- [ ] All 5 spec templates contain YAML frontmatter
- [ ] Frontmatter includes required fields: `id`, `status`, `created`, `prompt_version`
- [ ] Frontmatter format is valid YAML
- [ ] Templates are copied to `.smaqit/templates/` on init

#### 3. Initialization Validation
- [ ] `smaqit init` creates `.smaqit/state.json`
- [ ] state.json has correct initial structure:
  ```json
  {
    "version": "1.0",
    "phases": {
      "develop": {
        "completed": false,
        "specs_processed": 0,
        "specs_succeeded": 0,
        "specs_failed": 0
      },
      "deploy": { ... },
      "validate": { ... }
    }
  }
  ```
- [ ] All three phases initialized with `completed: false`
- [ ] All spec counts initialized to `0`

#### 4. Status Display Validation
- [ ] `smaqit status` reads state.json correctly
- [ ] Status displays phase names (Develop, Deploy, Validate)
- [ ] Status shows "Not started" for incomplete phases
- [ ] Status displays spec counts when phase completed (format TBD)

#### 5. Agent Files Validation
- [ ] All 8 agents copied to `.github/agents/`
- [ ] Specification agents (5) have state tracking in Output section
- [ ] Implementation agents (3) have state tracking directives
- [ ] No syntax errors in agent files

### Test Environment

**Location:** `installer/test/stateful-test-{timestamp}/`

**Setup:**
```bash
cd installer
make build
mkdir -p test/stateful-test-$(date +%s)
cd test/stateful-test-$(date +%s)
```

**Cleanup:** Remove test directory after completion

### Expected Outcomes

**Success Criteria:**
- All infrastructure components present and correctly formatted
- State tracking mechanisms in place (even if not yet used by agents)
- CLI commands work without errors
- Documentation matches implementation

**Failure Scenarios:**
- Template frontmatter missing or malformed → Block Phase 2
- state.json initialization fails → Block Phase 2
- Status display errors → Fix before Phase 2
- Agent files have syntax issues → Fix before Phase 2

## Phase 2: Incremental Workflow Testing

**Status:** Ready to execute with v0.5.0-beta

**Unblocked by:** Task 047 completion (incremental processing fully implemented)

### Pre-Phase 2 Verification

✅ **All requirements satisfied:**
- [x] Agent files contain `smaqit plan --phase=[PHASE]` directives
- [x] Agents have MUST directives to process only draft/failed specs
- [x] CLI implements `filterSpecsByStatus()` function
- [x] Documentation describes incremental processing behavior (framework/PHASES.md)
- [x] `smaqit plan` command operational with --regen flag

### Test Environment Setup

**Version:** v0.5.0-beta (tagged and built)

**Test Directory:** `installer/test/phase2-incremental-{timestamp}/`

**Setup Commands:**
```bash
cd installer
make build
mkdir -p test/phase2-incremental-$(date +%s)
cd test/phase2-incremental-$(date +%s)
../../dist/smaqit init
```

**Cleanup:** Remove test directory after completion or keep for inspection

### Test Scenarios

#### Scenario 1: Add New Feature (Incremental Spec Generation)

**Objective:** Verify specs are generated incrementally without regenerating existing specs

**Steps:**
1. Create initial minimal business prompt with 1 use case
2. Invoke business agent to generate spec(s)
3. Verify spec(s) created with `status: draft`, `prompt_version` set
4. Add second use case to business prompt
5. Invoke business agent again
6. Verify: NEW spec created for second use case, EXISTING spec unchanged (same prompt_version)
7. Check `smaqit plan --phase=develop` outputs both specs

**Expected Outcome:**
- Initial spec retained with original timestamps
- New spec added with current timestamp
- Both specs have `status: draft`
- Plan command returns both paths

**Success Criteria:**
- [ ] Existing spec unchanged (compare prompt_version, created timestamp)
- [ ] New spec created with current timestamp
- [ ] Both specs appear in plan output
- [ ] No duplicate specs created

#### Scenario 2: Incremental Implementation (Skip Completed Specs)

**Objective:** Verify development agent processes only draft specs, skips implemented specs

**Prerequisites:** Scenario 1 completed with 2 business specs

**Steps:**
1. Run `smaqit plan --phase=develop` before implementation
2. Manually set first spec to `status: implemented`, add `implemented: [timestamp]`
3. Run `smaqit plan --phase=develop` again
4. Verify plan now returns only the second spec (still draft)
5. (Simulated) Development agent would process only returned spec
6. Manually set second spec to `status: implemented`
7. Run `smaqit plan --phase=develop` again
8. Verify plan returns empty output (all specs implemented)

**Expected Outcome:**
- Plan command filters based on status field
- Only draft/failed specs returned
- Implemented specs skipped
- Empty output when all specs complete

**Success Criteria:**
- [ ] Plan returns only draft specs initially
- [ ] Plan excludes specs with `status: implemented`
- [ ] Plan returns empty when all specs implemented
- [ ] CLI suggests `--regen` flag when empty (check status command output)

#### Scenario 3: Failed Spec Reprocessing

**Objective:** Verify failed specs are reprocessed while successful specs remain untouched

**Prerequisites:** 2 specs exist, both implemented

**Steps:**
1. Manually change first spec to `status: failed`
2. Run `smaqit plan --phase=develop`
3. Verify only failed spec returned
4. Manually change failed spec back to `status: implemented`
5. Run plan again, verify empty output

**Expected Outcome:**
- Failed specs treated same as draft specs (reprocessed)
- Implemented specs skipped
- Agent can retry failed work without touching successful work

**Success Criteria:**
- [ ] Plan returns failed spec
- [ ] Plan excludes implemented spec
- [ ] Failed spec can be corrected and marked implemented
- [ ] Plan respects corrected status

#### Scenario 4: Full Regeneration with --regen Flag

**Objective:** Verify --regen flag processes all specs regardless of status

**Prerequisites:** Multiple specs with mixed statuses (draft, implemented, failed)

**Steps:**
1. Create test scenario with 3 specs:
   - Spec 1: `status: implemented`
   - Spec 2: `status: draft`
   - Spec 3: `status: failed`
2. Run `smaqit plan --phase=develop` (without --regen)
3. Verify returns only draft and failed specs (2 paths)
4. Run `smaqit plan --phase=develop --regen`
5. Verify returns ALL specs (3 paths)

**Expected Outcome:**
- Default mode: incremental (draft + failed only)
- Regen mode: all specs regardless of status
- Users can force full regeneration when needed

**Success Criteria:**
- [ ] Default plan excludes implemented specs
- [ ] --regen flag includes all specs
- [ ] Path output correct in both modes

#### Scenario 5: Phase Status Display with Spec Counts

**Objective:** Verify `smaqit status` shows accurate phase completion based on spec states

**Prerequisites:** Specs in various states across layers

**Steps:**
1. Create business spec (draft), functional spec (draft), stack spec (draft)
2. Run `smaqit status`
3. Verify Phase 1 (Develop) shows "⚙ In progress - 0 implemented, 3 draft"
4. Change all 3 specs to `status: implemented`
5. Run `smaqit status` again
6. Verify Phase 1 shows "✓ Complete - 3 implemented"

**Expected Outcome:**
- Status command scans spec frontmatter
- Aggregates counts by phase
- Shows completion when ALL required layers present + ALL specs at target status

**Success Criteria:**
- [ ] Status correctly identifies incomplete phases
- [ ] Status shows Complete only when all layers + correct status
- [ ] Spec counts accurate per phase
- [ ] Partial layer coverage shows "In progress" not "Complete"

#### Scenario 6: Cross-Phase State Progression

**Objective:** Verify specs progress through lifecycle states across all three phases

**Steps:**
1. Create 1 spec in each layer (business, functional, stack, infrastructure, coverage)
2. Verify all start with `status: draft`
3. Mark Phase 1 specs as `implemented`, verify status command shows Phase 1 complete
4. Mark infrastructure spec as `deployed`, verify Phase 2 complete
5. Mark coverage spec as `validated`, verify Phase 3 complete
6. Check `smaqit plan --phase=validate` returns empty (all validated)

**Expected Outcome:**
- Specs transition through expected lifecycle
- Phase completion detection works across all phases
- Timestamps added at each transition

**Success Criteria:**
- [ ] Each phase respects its target status (implemented/deployed/validated)
- [ ] Status command shows accurate phase completion
- [ ] Timestamps added to frontmatter at each phase
- [ ] Plan command respects phase-specific status filtering

### Test Artifacts

**Location:** `docs/user-testing/2026-01-03_phase2-incremental-workflow-testing.md`

**Contents:**
- Test execution log (commands run, outputs)
- Screenshots/output samples for each scenario
- Pass/fail results per scenario
- Issues identified (if any)
- Recommendations for improvements

### Validation Approach

**Manual testing** (not automated agents):
- Execute commands directly via terminal
- Manually edit spec frontmatter to simulate agent behavior
- Verify CLI output matches expectations
- Document results with evidence

**Rationale:** Phase 2 validates the INFRASTRUCTURE (CLI commands, frontmatter parsing, status aggregation). Actual agent behavior testing is separate (would be Task 046 or future work).

## Acceptance Criteria

### Phase 1: Infrastructure Validation (Documentation Review)
- [x] All infrastructure test scenarios executed
- [x] Test results documented with evidence (commands, outputs, screenshots)
- [x] Blockers identified and categorized (critical vs minor)
- [x] Validation report created: `docs/user-testing/2026-01-03_stateful-specs-infrastructure-validation.md`
- [x] Incremental processing gap documented (present vs absent)
- [x] Enhancement task created: Task 047 - Implement Incremental Processing

**Phase 1 Result:** ✅ **PASS** - All infrastructure components validated successfully

### Phase 2: Incremental Workflow Testing
- [ ] Pre-Phase 2 verification completed (Task 047 completion confirmed)
- [ ] v0.5.0-beta release created with incremental processing
- [ ] Test environment setup successful
- [ ] Scenario 1 executed: Add New Feature (incremental spec generation)
- [ ] Scenario 2 executed: Incremental Implementation (skip completed)
- [ ] Scenario 3 executed: Failed Spec Reprocessing
- [ ] Scenario 4 executed: Full Regeneration with --regen
- [ ] Scenario 5 executed: Phase Status Display with Spec Counts
- [ ] Scenario 6 executed: Cross-Phase State Progression
- [ ] Test report created: `docs/user-testing/2026-01-03_phase2-incremental-workflow-testing.md`
- [ ] All scenarios passed OR issues documented with recommendations
- [ ] Task 045 marked complete in PLANNING.md
- [ ] Task 014 marked as fully validated

## Notes

**Design Decision:** Split validation and testing into phases to avoid scope creep. Phase 1 validates foundational infrastructure through documentation review without runtime execution. Phase 2 will perform actual testing with built system. This approach:
- Provides early validation of state tracking mechanisms
- Identifies gaps before extensive testing
- Allows Task 014 release with infrastructure only (incremental as enhancement)
- Follows "validate what's documented, test what executes" principle

**Related Tasks:**
- Task 014: Implemented the stateful specifications system
- Task 046: Document implementation workflows (including incremental processing)
- Task 047: Implement incremental processing in agents
