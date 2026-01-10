# Testing smaqit

This document explains the rationale behind smaqit's testing approach for validating the framework itself.

## Testing Philosophy

### User Perspective Over Framework Internals

Test what users experience, not internal structure:
- Validate workflows end-to-end, not individual components
- Documentation quality matters (specs should be clear enough for implementation)
- Focus on actual pain points users encounter

**Rationale:** Users don't care about internal implementation details. They care whether the workflow is smooth, intuitive, and produces useful results.

### Comprehensive Reporting Over Binary Pass/Fail

Continue on failures to collect all painpoints:
- Categorize issues by severity and type (blockers, issues, UX friction, performance)
- Provide actionable recommendations
- Historical reports show progress and regressions over time
- Standardized format enables comparison across sessions

**Rationale:** Stopping at the first failure hides downstream problems. Collecting all failures in one run accelerates iteration and reveals systemic issues.

### Automation as Accelerator, Not Replacement

Agent handles tedious execution steps:
- Human reviews results and makes decisions
- Minimal validation reduces false positives
- Focus automation on high-value repeatability

**Rationale:** LLM-generated specs vary between runs (non-deterministic output). Exact content matching would create false negatives. File existence check is sufficient—detailed validation is delegated to `smaqit validate` command.

### Minimal Validation Strategy

Why we validate file existence only (not content):

- **LLM variance**: Generated specs vary in style between runs
- **Behavioral focus**: Tests verify workflows complete successfully
- **Reduced brittleness**: Exact matching would create false negatives
- **Proper delegation**: Content validation belongs in `smaqit validate` command

**Trade-off:** We might miss malformed specs, but we avoid maintenance burden of brittle content assertions. Trust that agents following templates will produce valid specs.

### Painpoint Categorization

Issues are categorized to prioritize fixes:

| Category | Definition | Priority |
|----------|------------|----------|
| **Blockers** | Critical issues preventing progress | P0 - Fix immediately |
| **Issues** | Problems that affect user experience | P1 - Fix before release |
| **UX Friction** | Workflow awkwardness or confusion | P2 - Improve iteratively |
| **Performance** | Timing or resource concerns | P3 - Optimize when data shows impact |

**Rationale:** Not all failures are equal. Categorization enables rational triage and prevents scope creep ("fix everything before shipping").

## Automated Testing

For automated end-to-end testing, see `.github/agents/smaqit.user-testing.agent.md`.

The testing agent orchestrates the complete workflow automatically and generates comprehensive reports.

### Pre-Release Testing Checklist

Before creating a release:

- [ ] Build for all platforms: `make build-all`
- [ ] Test init/status/validate/help/uninstall on current platform
- [ ] Verify version embedding: `dist/smaqit version` matches git tag
- [ ] Check embedded files count: init should create 14+ files in `.smaqit/`
- [ ] Validate clean uninstall: no residual files after uninstall

## Test Case Design

### Why Mario Hello World?

Mario Hello World was chosen as the standardized test case because:

- **Domain-agnostic**: Works for any project type (not web/API/CLI specific)
- **Simple**: Completes quickly (minutes, not hours)
- **Comprehensive**: Exercises all 5 specification layers meaningfully
- **Familiar**: Reduces cognitive load (everyone knows Mario)
- **Expandable**: Foundation for future test scenarios

**Trade-off:** Simple test case might not catch complex edge cases. But comprehensive coverage of happy path is more valuable than partial coverage of edge cases at this stage.

### Report Template Design

39-point checklist provides:

- **Regression tracking** across sessions (compare reports over time)
- **Issue categorization** by severity (enables rational triage)
- **Machine-parseable** structure for future automation
- **Consistent format** enables comparison (same categories every run)

**Rationale:** Strict format might feel rigid, but consistency is more valuable than flexibility for historical analysis. Reports are data, not prose.

## Future Enhancements

Potential improvements deferred until multi-contributor phase:

- **CI/CD Integration** (Task 025): GitHub Actions for automated testing on PRs
- **Parallel test execution**: Multiple scenarios simultaneously
- **Performance benchmarking**: Track workflow execution times
- **Coverage analysis**: Which framework paths are exercised

**Why defer:** Current manual testing is sufficient for single-maintainer workflow. Automation value increases with team size and contribution frequency.

## Prompt Usage Guidance

### test.start Prompt

**Purpose:** Start focused testing session with minimal context loading.

**Why This Approach:**
- **Focused context** - Only what's needed for test execution (~15-20K tokens vs 50K+ for full session)
- **Faster startup** - No history/planning/wiki reading
- **Clearer purpose** - Agent knows it's in test mode immediately
- **Self-contained** - Test task file has complete workflow
- **Repeatable** - Same pattern for any test task

**When to Use:**
- Running E2E tests (Task 048, 059, etc.)
- Running regression tests
- Running integration tests
- Any structured test with task file

**When NOT to Use:**
- Development work (use `/session.start`)
- Framework changes (use `/session.start`)
- Exploratory testing without task file
- General assistance (use normal chat)

**Trade-off:** Focused context means less background knowledge. Agent can't reference history or planning without additional context loading. But speed and clarity outweigh this limitation for structured testing.
