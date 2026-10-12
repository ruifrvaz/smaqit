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

Agent-driven user testing (the `smaqit.user-testing` agent and the `smaqit.test-*` skills) ships with the sibling `smaqit-extensions` project, not with this repository.

### Claude Code End-to-End Test

`make -C installer e2e-claude` checks that a real Claude Code session can build an application through smaqit. It runs `scripts/e2e-claude.sh`, which works in five stages:

1. **Builds** the dev binary.
2. **Isolates** Claude Code. The script creates a scratch folder and points `HOME` and `CLAUDE_CONFIG_DIR` into it, so your own `~/.claude` is never read or written. It confirms the isolation with `claude auth status` before any model call.
3. **Installs** smaqit globally into that scratch home, then runs `smaqit init` in a fresh scratch project for each scenario.
4. **Drives** headless `claude -p` through the scenarios below. Live questions are turned off, so every question has to come back through smaqit's return signals.
5. **Checks** the results with smaqit's own CLI and the fixture's tests, and keeps the evidence.

Every scenario builds the same fixture, `scripts/e2e/hello-site/`: a one-page "Hello, Mario!" site in plain HTML, CSS, and JavaScript, tested with `node --test`.

| Scenario | What it proves |
|---|---|
| `happy-path` | `/smaqit.development` with complete requirements finishes the Develop phase: specifications for every layer, `smaqit validate` passing, nothing pending in `smaqit plan --phase=develop`, the site built, its tests passing, and a development report |
| `needs-input` | With thin requirements, the first turn writes no specification and returns `NEEDS-INPUT`. A `--resume` turn with the answers then completes the phase |
| `assisted` | Assisted mode pauses with `REVIEW-REQUESTED` after each specification layer. Approvals resume it, and approved specifications are never regenerated |
| `self-invocation` | A plain request with no slash command still runs through smaqit: the session log shows a delegation to `smaqit-development` |

**Credentials.** The script uses the first of these it finds:
1. `CLAUDE_CODE_OAUTH_TOKEN` (a claude.ai subscription token from `claude setup-token`) or `ANTHROPIC_API_KEY` (pay-per-use) in your environment.
2. Otherwise, your current claude.ai login. The script reads only that login's short-lived access token from your credentials file and passes it to the isolated run as `CLAUDE_CODE_OAUTH_TOKEN`. It refuses to start when the token has less than 30 minutes left, and warns when the token would expire before `E2E_TIMEOUT`. The refresh token is never used, so the run cannot rotate your real login.

Either way, nothing is copied into the scratch folder, and the run fails if a credential file ever appears there.

**Cost.** Every run spends real model usage, so it is a manual pre-release check, not part of CI. These variables control it:
- `SCENARIOS`: which scenarios run, for example `SCENARIOS=happy-path`. The default is all four.
- `E2E_BUDGET_USD`: spend cap per turn (default 10)
- `E2E_TIMEOUT`: seconds per turn (default 3600)
- `E2E_MAX_TURNS`: turn cap per run (default 200)

**Evidence.** Each run writes to `installer/test/e2e-claude/<timestamp>/`, which is git-ignored. Each scenario gets the stream-json session log for every turn, a run summary (Claude Code version, MCP server status, cost, permission denials), copied transcripts, the final project tree, and one PASS/FAIL line per check. Set `KEEP_E2E_DIR=1` to keep the scratch folder as well.

**Reading a failure:**
- Never trust a `claude -p` exit code on its own; a headless run can exit 0 without doing the work. The checks look at the files written and the session log instead.
- If `happy-path` stalls after the Development agent hands work to a specification agent, check upstream Claude Code issue #75043, which makes a subagent's own subagents run detached.
- **Permissions.** Turns run in Claude Code's `auto` permission mode (`E2E_PERMISSION_MODE`). Its classifier allows ordinary shell commands, including the compound commands agents compose, and denies risky ones; with live prompts off, anything blocked becomes a denial. Each turn's `permission_denials` lists them. A per-command allowlist proved too brittle: it denied a compound `ls … && smaqit status | head` command and stopped the run. Don't bypass permissions entirely, because the scratch folder isolates `HOME` but does not sandbox the shell.
- **Background subagents.** The main conversation often starts `smaqit-development` in the background, and headless mode waits for it. The development agent's own result then arrives as a `system/task_notification` event and a second `result` event, not as a tool result. The log checks read all of these.

### Pre-Release Testing Checklist

Before creating a release:

- [ ] Build for all platforms: `make build-all`
- [ ] Run generator/unit checks: `make test`
- [ ] Run the installer lifecycle smoke test: `make smoke-test`
- [ ] For releases that change agents, commands, or input skills, run the Claude Code end-to-end test: `make e2e-claude` (spends model usage)
- [ ] Test init/status/validate/help/uninstall on current platform
- [ ] Test Node 22+ fail-fast initialization, embedded runtime integrity, offline PlantUML rendering, PNG/hash staleness, visual attestation, and strict existing-spec migration failures
- [ ] Verify version embedding: `dist/smaqit version` matches git tag
- [ ] Check generated platform artifacts: 9 agents and 28 skills for Copilot, Claude Code, and Codex
- [ ] Validate selective MCP uninstall: smaqit-owned registrations are removed while custom `.vscode/mcp.json`, `.mcp.json`, `.codex/agents/`, `.agents/skills/`, and `.codex/config.toml` content survives

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
