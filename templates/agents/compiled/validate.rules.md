# Validation Phase Compilation Rules

**L1 Transformation Rules:** Compile L0 principles → L1 directives for Validation phase agent

**Target Agent:** `agents/smaqit.validation.agent.md`

---

## Source L0 Principles

### Primary Source: ARTIFACTS.md § The Test Independence Principle

> "Test artifacts exist independently of agent execution. Tests can run in any environment with the appropriate runtime, enabling continuous integration, local developer workflows, and automated verification outside the validation phase."

### Secondary Source: PHASES.md § Validate Phase Activities

> "The Coverage agent translates acceptance criteria from all upstream specs into executable test definitions, mapping each requirement to expected outcomes and flagging criteria that cannot be automatically verified.
>
> The Validation agent generates test artifacts that can run independently of agent execution, executes those tests against the deployed system, and produces a validation report documenting coverage and results."

### Tertiary Source: ARTIFACTS.md § Implementation Artifacts by Phase

> **Validate Phase:**
> - Test artifacts (executable, committable): Test files, framework configuration, fixtures/utilities, CI/CD workflow configuration
> - Validation report with spec coverage, pass/fail status, unverified requirements, failure details

---

## L1 Directive Compilation

### Output Artifacts

**L0 Source:** "Test artifacts that exist independently...These include test implementations, framework configuration, test utilities, and CI/CD integration."

**Compile to [OUTPUT_ARTIFACTS]:**
```markdown
- **Test artifacts (executable, committable):**
  - Test files in `tests/` directory implementing Coverage spec test cases
  - Test framework configuration (e.g., `pytest.ini`, `unittest.cfg`, `jest.config.js`)
  - Test fixtures and utilities (e.g., `tests/conftest.py`, `tests/fixtures/`)
  - CI/CD workflow configuration (e.g., `.github/workflows/validation.yml`)
- **Validation report** in `.smaqit/reports/validation-phase-report-YYYY-MM-DD.md` containing:
  - Spec coverage percentage
  - Pass/fail status per requirement
  - Unverified requirements with justification
  - Failure details for failed tests
```

### Phase-Specific Rules: Test Artifact Generation

**L0 Source:** "Test artifacts exist independently of agent execution"

**Compile to MUST directives:**
- Generate executable test artifacts from Coverage specifications
- Create test files in `tests/` directory implementing Coverage spec test cases
- Use test framework specified in Stack spec
- Organize tests by feature with clear mapping to Coverage spec scenarios
- Preserve Given/When/Then structure from Gherkin scenarios in test code
- Generate test framework configuration file
- Generate test fixtures and utilities as needed
- Generate CI/CD workflow configuration
- Ensure test artifacts are executable independently (outside agent context)
- Ensure tests can run in any environment with appropriate runtime
- Execute generated tests against deployed system

**L0 Source:** "Tests can run in any environment with the appropriate runtime"

**Compile to MUST NOT directives:**
- Embed test logic within agent execution flow
- Hardcode environment-specific values in test code
- Generate tests that depend on agent context to execute
- Skip test artifact generation (tests must be committable files)

**Format Requirements:**
- Test files follow naming convention: `tests/test_[feature_name].[ext]`
- CI/CD workflow triggers on push/pull request, runs tests, reports results
- Test framework configuration includes all necessary settings for independent execution
- Tests use environment variables or configuration files for environment-specific values

### Completion Criteria

**L0 Source:** "The Validate phase completes when...test artifacts have been generated and executed against the deployed system"

**Compile to [ADDITIONAL_COMPLETION_CRITERIA]:**
- [ ] Test artifacts generated (executable test files, framework configuration, CI/CD workflow)
- [ ] Tests executable independently (verified outside agent context)
- [ ] Test framework configuration includes all necessary settings
- [ ] CI/CD workflow configuration created and functional

---

## Compilation Guidance for Agent-L2

1. **Replace [PHASE_SPECIFIC_RULES]** with "Phase-Specific Rules: Test Artifact Generation" section above
2. **Replace [OUTPUT_ARTIFACTS]** with "Output Artifacts" section above
3. **Replace [ADDITIONAL_COMPLETION_CRITERIA]** with "Completion Criteria" section above
4. **Replace generic placeholders:**
   - `[PHASE]` → `validate`
   - `[PHASE_NAME]` → `Validation`
   - `[phase]` → `validation`
   - Other placeholders per implementation-agent.template.md

5. **Preserve traceability:** Keep L0 source citations as HTML comments for debugging

---

## Version

- **Created:** 2026-01-14
- **L0 Sources:** ARTIFACTS.md (Test Independence Principle), PHASES.md (Validate Phase)
- **Compilation Target:** agents/smaqit.validation.agent.md
