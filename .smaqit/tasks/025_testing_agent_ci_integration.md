# Task: Integrate Testing Agent with CI/CD

**ID**: 025
**Status**: new
**Created:** 2025-12-20

## Description

Automate end-to-end smaqit testing via CI/CD integration. Create Makefile targets that invoke the testing agent programmatically, and optionally create GitHub Actions workflow that runs tests on pull requests and releases.

## Context

Task 024 created `@smaqit.testing` agent for manual invocation via Copilot chat. This works well for interactive testing but doesn't integrate into automated workflows.

To enable:
- **Local automation:** `make test-e2e` runs testing agent via command line
- **CI/CD automation:** GitHub Actions executes testing agent on PR/release
- **Regression prevention:** Automated tests catch issues before merge

Testing agent already generates machine-parseable reports with standardized checklists, making CI integration straightforward.

## Acceptance Criteria

### Makefile Integration

- [ ] Add `test-e2e` target to `installer/Makefile`
- [ ] Target builds installer, invokes testing agent, reports results
- [ ] Target exits with non-zero code on FAIL (for CI)
- [ ] Target works on all platforms (Linux, macOS, Windows)

### GitHub Actions Workflow

- [ ] Create `.github/workflows/test-e2e.yml`
- [ ] Workflow runs on pull requests to `main` branch
- [ ] Workflow runs on release tags
- [ ] Workflow reports results as PR check
- [ ] Workflow uploads test report as artifact

### Report Parsing

- [ ] Script to parse test report and extract PASS/FAIL status
- [ ] Script outputs results in CI-friendly format (exit codes, summary)
- [ ] Script handles malformed or missing reports gracefully

### Documentation

- [ ] Update `.github/copilot-instructions.md` with CI/CD usage
- [ ] Document how to run `make test-e2e` locally
- [ ] Document GitHub Actions workflow configuration

## Implementation Notes

**Makefile Target Structure:**
```makefile
.PHONY: test-e2e
test-e2e: build
	@echo "Running end-to-end test with testing agent..."
	# Invoke testing agent programmatically
	# Parse report for PASS/FAIL
	# Exit with appropriate code
```

**GitHub Actions Workflow:**
```yaml
name: End-to-End Test
on:
  pull_request:
    branches: [main]
  release:
    types: [published]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make test-e2e
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-report
          path: docs/user-testing/*.md
```

**Agent Invocation Challenge:**
Testing agent is designed for Copilot chat interaction. For programmatic invocation, need to:
1. Call agent directly via Copilot API (if available)
2. Extract agent logic into standalone script
3. Use Copilot CLI (if exists)
4. Defer until GitHub provides agent invocation API

**Recommended Approach:**
Start with option 2 (standalone script) to unblock CI/CD integration. Migrate to agent invocation when API becomes available.

## Dependencies

- Task 024 completed (testing agent exists)
- Go 1.25+ installed (for builds)
- GitHub Actions enabled on repository

## Open Questions

- Can GitHub Copilot agents be invoked programmatically from CI/CD?
- Should CI run testing agent on every commit or only on PR/release?
- Should failed tests block PR merge or just report results?
- Do we need multiple test cases in CI or is Mario hello world sufficient?

## Future Enhancements

- Parameterize test case (not just Mario, run multiple)
- Matrix testing (multiple OS, Go versions)
- Performance benchmarking over time
- Test coverage trending

## Notes

This task enables "shift-left" testing by catching issues earlier in development cycle. However, full automation may not be necessary initially—manual testing with `@smaqit.testing` may be sufficient until smaqit reaches multi-contributor phase.

Consider deferring GitHub Actions until:
- Multiple contributors actively committing
- Release cadence increases (> 1 per week)
- Manual testing becomes bottleneck
