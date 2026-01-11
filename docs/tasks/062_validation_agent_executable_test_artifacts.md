# Validation Agent Should Generate Executable Test Artifacts

**Status:** Code Complete - Ready for E2E Testing  
**Created:** 2026-01-11  
**Updated:** 2026-01-11  
**Priority:** High (Release Blocker)  
**Related:** Issue 12 from Task 059 (E2E Regression Testing)

## Description

Validation agent performs manual verification and generates validation report but does NOT generate executable test artifacts (no `tests/*.py` files, no `pytest.ini`, no CI/CD workflow configurations). This breaks the CI/CD automation value proposition - validation becomes a one-time manual check instead of automated regression testing.

**Current Behavior:**
- Coverage agent generates `specs/coverage/*.md` with Gherkin-format test definitions
- Validation agent reads Coverage specs, performs manual verification
- Validation agent generates report in `.smaqit/reports/validation-phase-report-*.md`
- **NO executable test files generated** - validation is one-time manual activity

**Expected Behavior:**
- Coverage agent generates Coverage specs (existing - no change)
- Validation agent reads Coverage specs
- **Validation agent generates executable test framework:**
  - `tests/*.py` files implementing Gherkin scenarios
  - `pytest.ini` or equivalent test framework configuration
  - Test utilities/fixtures in `tests/conftest.py`
  - CI/CD workflow configuration (`.github/workflows/validation.yml`)
- Validation agent executes generated tests
- Validation agent generates report with pass/fail results
- **Tests are committable and re-runnable** - automated regression testing

## Acceptance Criteria

- [x] Update `agents/smaqit.validation.agent.md` Output section to include test artifacts
- [x] Update agent directives: "Generate executable test files from Coverage specs"
- [x] Directive: "Use test framework specified in Stack spec (pytest, unittest, etc.)"
- [x] Directive: "Generate tests in `tests/` directory with proper structure"
- [x] Directive: "Generate test framework configuration (`pytest.ini`, `unittest.cfg`, etc.)"
- [x] Directive: "Generate CI/CD workflow file in `.github/workflows/validation.yml`"
- [x] Update PHASES.md Validate phase completion criteria to include test artifact generation
- [x] Update ARTIFACTS.md to document test artifacts as implementation artifacts
- [ ] Validation: Re-run validation phase with Mario + Luigi test case
- [ ] Validation: Verify `tests/*.py` files exist and are executable
- [ ] Validation: Verify test framework configuration file exists
- [ ] Validation: Verify CI/CD workflow file exists
- [ ] Validation: Run `pytest tests/` independently (outside agent) - tests execute successfully
- [ ] Validation: Tests fail appropriately when requirements not met (negative test)

## Notes

**Severity:** High - Release blocker for v0.5.0-beta

**Rationale:** Without executable test artifacts, validation phase provides no CI/CD automation capability. Core value proposition of Coverage/Validation layers is continuous validation, not one-time manual checks.

**Impact:**
- Cannot commit tests to version control
- Cannot run tests in CI/CD pipeline
- Cannot perform automated regression testing
- Validation must be manually repeated after every code change
- Breaks "reproducible from input set" principle

**Test Framework Selection:**
- Should respect Stack spec technology choices
- Python: pytest (preferred), unittest
- JavaScript: jest, mocha
- Go: go test
- Java: JUnit, TestNG

**Test Structure:**
- Feature-based organization: `tests/test_[feature_name].py`
- Map Gherkin scenarios to test functions
- Given/When/Then structure preserved in test code
- Test data/fixtures in `tests/conftest.py` or `tests/fixtures/`

**CI/CD Workflow:**
- Trigger on push/pull request
- Install dependencies from Stack spec
- Run tests with coverage reporting
- Fail build on test failure
- Report results to PR/commit status

**Acceptance Criteria Note:**
Tests MUST be executable independently (outside agent context) to validate they are proper artifacts, not agent-specific code.

**Affected Files:**
- `agents/smaqit.validation.agent.md` (primary)
- `framework/PHASES.md` (completion criteria update)
- `framework/ARTIFACTS.md` (document test artifacts)

## Implementation Log

### 2026-01-11: Framework and Agent Updates

**Approach:** Followed smaqit level hierarchy (Level 0 → Level 2) to cascade changes.

**Level 0: Framework Updates**

1. **PHASES.md** - Validate phase:
   - Updated workflow to include test artifact generation steps (a-c)
   - Expanded Output section to list test artifacts (test files, config, fixtures, CI/CD)
   - Added test artifact generation to completion criteria
   - Added "Tests are executable independently" validation requirement

2. **ARTIFACTS.md** - Implementation artifacts:
   - Renamed section from "Reports" to "Reports and Test Artifacts"
   - Listed test artifacts as executable, committable outputs
   - Added MUST requirement: "Test artifacts MUST be executable independently"

**Level 2: Agent Updates**

3. **agents/smaqit.validation.agent.md**:
   - Updated Output section with two artifact categories (test artifacts + validation report)
   - Expanded Format section with test implementation details
   - Added comprehensive test generation directives to MUST section:
     - Test file generation in `tests/` directory
     - Test framework selection from Stack spec with fallback defaults
     - Feature-based test organization
     - Gherkin scenario mapping to test functions
     - Test framework configuration generation
     - Test fixtures/utilities generation
     - CI/CD workflow generation
     - Independent executability requirement
   - Added new "Test Artifact Generation" section in Phase-Specific Rules with:
     - Test framework selection strategy (Stack spec → defaults)
     - Test file organization patterns
     - Test framework configuration requirements
     - Test fixtures and utilities guidance
     - CI/CD workflow requirements
     - Independent executability emphasis
   - Updated Completion Criteria to include test artifact validation

**Key Decisions:**

1. **Test framework selection**: Respect Stack spec choices, fallback to sensible defaults (pytest for Python, jest for JS, etc.)
2. **Test organization**: Feature-based (`tests/test_[feature].py`) for clarity and traceability
3. **CI/CD scope**: Basic workflow covering install, run, report - agent decides specifics based on Stack
4. **Independent executability**: Tests MUST run via framework CLI without agent context

**Build Validation:**
```bash
cd installer && make build
# Result: ✅ Build successful (version 90bf802-dirty)
```

**Remaining Work:**
- End-to-end validation with actual test case (Mario + Luigi or similar)
- Verify generated test artifacts are correct and executable
- Negative test to ensure tests fail appropriately

**Note:** The framework and agent instructions have been updated to specify test artifact generation. The actual behavior validation should occur when the Validation agent is invoked on a real project after these changes are merged. The testing agent (`smaqit.user-testing`) can be used for comprehensive end-to-end validation of the complete workflow.

## Consistency Verification

**Cross-Level Alignment:**
- ✅ Framework (PHASES.md) defines workflow with test artifact generation
- ✅ Framework (ARTIFACTS.md) documents test artifacts as implementation outputs
- ✅ Agent (smaqit.validation.agent.md) implements framework requirements
- ✅ Agent references Stack spec for technology choices (respects layer architecture)
- ✅ Agent maintains traceability to Coverage specs (COV-[CONCEPT]-NNN)

**Key Requirements Present:**
- ✅ Test framework selection from Stack spec with sensible defaults
- ✅ Feature-based test organization in `tests/` directory
- ✅ Test framework configuration generation
- ✅ Test fixtures and utilities generation
- ✅ CI/CD workflow generation
- ✅ Independent executability requirement (critical for CI/CD)
- ✅ Execution against deployed system
- ✅ Validation report generation

**Principle Alignment:**
- ✅ **Traceability**: Tests include `# Implements: COV-[CONCEPT]-NNN` comments
- ✅ **Template-Constrained Output**: Agent follows structured output format
- ✅ **Self-Validation**: Completion criteria include test artifact validation
- ✅ **Layer Independence**: Uses Stack spec for technology choices
- ✅ **Reproducible**: Test artifacts are committable and re-runnable
