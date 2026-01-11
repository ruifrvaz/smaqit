# Validation Agent Should Generate Executable Test Artifacts

**Status:** Not Started  
**Created:** 2026-01-11  
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

- [ ] Update `agents/smaqit.validation.agent.md` Output section to include test artifacts
- [ ] Update agent directives: "Generate executable test files from Coverage specs"
- [ ] Directive: "Use test framework specified in Stack spec (pytest, unittest, etc.)"
- [ ] Directive: "Generate tests in `tests/` directory with proper structure"
- [ ] Directive: "Generate test framework configuration (`pytest.ini`, `unittest.cfg`, etc.)"
- [ ] Directive: "Generate CI/CD workflow file in `.github/workflows/validation.yml`"
- [ ] Update PHASES.md Validate phase completion criteria to include test artifact generation
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
- Possibly `framework/ARTIFACTS.md` (document test artifacts)
