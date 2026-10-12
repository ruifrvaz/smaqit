---
status: In Progress
mode: Assisted
started: "2026-10-12"
created: "2026-10-11"
---

# Claude Code Headless E2E Test: Install Into a Scratch Folder and Build a Site Through smaqit

## Description

smaqit has no working Claude Code end-to-end test. The only agent-driven e2e path is the Copilot-SDK `make test-e2e` (with `.github/workflows/e2e-test.yml` and `.github/scripts/run_e2e_test.py`), and it is broken: it calls the `smaqit.ci-testing` agent, which was deleted in commit `fb5b29f`. Session 058 declared that path Copilot-only and called Claude Code e2e coverage "new scope". As a result, nothing exercises the globally installed Claude agents, commands and skills end to end, or the task-131 `NEEDS-INPUT`/`REVIEW-REQUESTED` ask-and-resume contract, whose live check was deferred.

Add `scripts/e2e-claude.sh`, behind `make -C installer e2e-claude`. It reuses `scripts/smoke-test-installer.sh`'s isolated-HOME pattern: it installs the dev build into a scratch `HOME`/`CLAUDE_CONFIG_DIR`, runs `smaqit init` in a scratch project, then drives headless `claude -p` through three scripted scenarios against a new static-site fixture. Results are checked deterministically with smaqit's own CLI (`validate`, `plan --phase=develop`, `status`) and the fixture's own test, and per-scenario evidence is kept.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, Node.js, Go
**Platforms/Environments:** Linux, WSL2, macOS
**Features/Integrations:** headless claude -p, CLAUDE_CONFIG_DIR isolation, CLAUDE_CODE_OAUTH_TOKEN, session resume, stream-json output, project MCP servers (.mcp.json), PlantUML MCP
**Versions/Constraints:** smaqit 3.10.0, Node 22+, Claude Code 2.1.268+ (verified locally on 2.1.296), every run spends real model usage

## Design Decisions

All confirmed by the user on 2026-10-11, during `task.plan`.

- **Runner:** a standalone script now (`scripts/e2e-claude.sh` plus a `make -C installer e2e-claude` target). Move it to a Bench manifest once smaqit-adk task 037 (Bench harness selector for Claude Code) lands.
- **Authentication** (revised by the user on 2026-10-12, so that no API-billed key is required): use `CLAUDE_CODE_OAUTH_TOKEN` or `ANTHROPIC_API_KEY` if set. Otherwise reuse the current claude.ai login: read only `claudeAiOauth.accessToken` from the real credentials file and pass it to the isolated run as `CLAUDE_CODE_OAUTH_TOKEN`.
  - Refuse to start when the token has less than 30 minutes left, and warn when it would expire before `E2E_TIMEOUT`.
  - Never use the refresh token, so the run cannot rotate the user's live login.
  - Never write a credential file into the scratch folder.
  - A probe confirmed that a regular claude.ai access token works through `CLAUDE_CODE_OAUTH_TOKEN` in an isolated config directory (`authMethod: oauth_token`, no credential file written).
- **Scenarios:** happy-path develop, the `NEEDS-INPUT` round trip, and Assisted `REVIEW-REQUESTED`.
- **Fourth scenario, `self-invocation`** (user, 2026-10-12): a plain-language request with no slash command must make Claude route through smaqit. This closes task 130's deferred live check.
- **Proof from the session log** (2026-10-12): delegation to smaqit and returned signals are checked by parsing the stream-json session log for tool calls and signal blocks, not only by looking at files on disk.
- **Fixture:** a new static site, a one-page "Hello, Mario!" pinned to plain HTML/CSS/vanilla JS, with a test using Node's built-in runner (`node --test`) and no package installs.
- **Assumption, local-only:** no CI workflow. Every run spends real model usage, so it's an explicit pre-release check, not a per-PR gate.
- **Assumption, dev build by default:** `installer/dist/smaqit-dev` is used unless `--binary <path>` selects a release binary.
- **Permissions:** `--permission-mode auto` with `--permission-prompts none` (user decision, 2026-10-12). Claude Code's classifier allows ordinary commands and denies risky ones, and `E2E_PERMISSION_MODE` overrides the mode.
  - This replaced the original per-command `Bash(...)` allowlist, which proved too brittle in the first live run: it denied a compound `ls … && smaqit status | head` command, so the `self-invocation` and `assisted` agents stopped.
  - A probe confirmed that headless auto mode allows compound commands with no denials.
  - Permissions are never bypassed entirely.
- **Scope extension** (user, 2026-10-12, after the live runs):
  - Retire the broken Copilot-SDK e2e path: `make test-e2e`/`test-e2e-clean`, `.github/workflows/e2e-test.yml`, `.github/scripts/`, and `docs/test-cases/mario-hello.automated.md`. It depended on the deleted `smaqit.ci-testing` agent.
  - Fix `docs/wiki/workflows/quickstart.md`'s drift (global installation, requirements passed in the command, clarifying questions, assisted mode, plain-language requests).
  - The `.test-venv` cleanup in `make uninstall` and `.gitignore` stays, so leftovers from earlier runs still get removed.
- **Out of scope:** moving to a Bench manifest (blocked on smaqit-adk task 037); the deploy and validate phases; CI.

## Implementation Steps

*Phase A: harness (`scripts/e2e-claude.sh`)*

1. **Isolation**, following `scripts/smoke-test-installer.sh` (lines 19–34):
   - a `mktemp -d` root, with `HOME`, `CLAUDE_CONFIG_DIR`, `COPILOT_HOME` and `CODEX_HOME` inside it
   - resolve the absolute paths of `claude` and `node` before switching HOME
   - put the dev binary on `PATH` as `smaqit`
   - clean up with a `trap`, or keep the folder with `KEEP_E2E_DIR=1`
2. **Preflight, failing fast:**
   - Node 22+ is installed
   - `claude --version` meets the minimum (`--permission-prompts none` needs v2.1.259+)
   - `CLAUDE_CODE_OAUTH_TOKEN` or `ANTHROPIC_API_KEY` is set
   - `claude auth status` reports a `configDirectory` inside the scratch folder, proving isolation
3. **Install:**
   - run `smaqit-dev --install-global`, or the binary given by `--binary`
   - check that agents, commands and skills landed under the scratch `CLAUDE_CONFIG_DIR`
   - `git init` the scratch project, then `smaqit init`
4. **A `run_turn` helper.** Each turn calls `claude -p` with:
   - `--output-format stream-json --verbose`
   - `--permission-prompts none` and `--disallowedTools AskUserQuestion`
   - scoped `--allowedTools`
   - `--max-turns` and `--max-budget-usd`, wrapped in the shell `timeout`

   It captures `session_id` and the exit code. Follow-up turns use `--resume <session_id>`.
5. **Evidence per scenario:** the stream-json logs, session IDs, a copy of `$CLAUDE_CONFIG_DIR/projects/**`, the final project tree listing, and one pass/fail line per check, under a report folder. Print a summary and exit nonzero on any failure.

*Phase B: fixture (can run in parallel with A)*

6. `scripts/e2e/hello-site/` holds:
   - `requirements.md`: the full requirements, with the stack pinned
   - `thin-requirements.md`: deliberately missing actors and success metrics
   - `answers.md`: answers keyed by the business gate's question ids (`actors`, `success-metrics`, …)
   - `approve.md`: the approval text for review turns
   - `plain-request.md`: the same requirements worded as a natural request, with no slash command and no mention of smaqit

*Phase C: scenarios (depends on A and B)*

7. **`happy-path`:** one turn of `/smaqit.development <requirements>`. Checks:
   - exit code 0
   - at least one spec in each of `specs/{business,functional,stack}/`
   - `smaqit validate` passes
   - `smaqit plan --phase=develop` returns nothing
   - `smaqit status` contains `Phase 1 (Develop): ✓ Complete`
   - `index.html` exists and `node --test` passes
   - `.smaqit/reports/development-phase-report-*.md` exists
8. **`needs-input`:**
   - Turn 1 sends the thin requirements. It must end with no `specs/**/*.md` and with the stream containing `NEEDS-INPUT`.
   - Turn 2 is `--resume` with the answers. Afterwards a business spec must exist and the happy-path checks must pass.
9. **`assisted`:**
   - Turn 1 asks for assisted mode. It must end with `REVIEW-REQUESTED`, a business spec, and no functional spec.
   - The script records the approved specs' hashes, then resumes with the approval until the run finishes, capped at 6 turns.
   - Approved specs' hashes must be unchanged, and the happy-path checks must pass.
10. **`self-invocation`:** the prompt is `plain-request.md`, with no slash command. It passes when:
    - the session log shows a delegation to smaqit: a `Task`/`Agent` call with `subagent_type` `smaqit-development`, or a `Skill`/command call to `smaqit.development`
    - the happy-path checks pass
11. **Log-check helper:** shared by all scenarios. It parses the stream-json session log (one JSON object per line) for tool calls and for `NEEDS-INPUT`/`REVIEW-REQUESTED` blocks, and doesn't rely on plain text matches in the final result.
12. **Knobs:** `SCENARIOS=happy-path,needs-input,assisted,self-invocation` (default all), `E2E_BUDGET_USD`, `E2E_TIMEOUT`. Add the `e2e-claude` target to `installer/Makefile`, depending on `build`.

*Phase D: making headless runs possible, plus docs*

13. In the `commands/smaqit.{development,deployment,validation}.md` bodies, add a step 3 fallback: "if no question tool is available, print the signal's questions (or review files) and end your turn; the next user message carries the answers. Then resume the subagent as before". Without this the `needs-input` and `assisted` scenarios cannot work headless.
14. `docs/wiki/workflows/testing-smaqit.md`: add a "Claude Code e2e" section (authentication via `claude setup-token`, cost and budget knobs, scenarios, evidence), and remove the stale reference to the deleted `.github/agents/smaqit.user-testing.agent.md`. Add a compendium entry.

*Verification*

15. Run `bash -n` (and `shellcheck` if available), the preflight negative cases, then each scenario live with `KEEP_E2E_DIR=1`. Confirm no `.credentials.json` exists under the scratch `CLAUDE_CONFIG_DIR`, and that the real `~/.claude` was untouched (mtimes).

## Known Issues Triage
**Triaged:** 2026-10-12
**Tools searched:** Claude Code (`anthropics/claude-code`)
**Result:** Advisory

### Advisory Issues
- [#75043 Nested subagents: children spawned by a subagent are always async (regardless of run_in_background), completion notifications never reach the subagent parent, and TaskStop fails with ownership errors after resume](https://github.com/anthropics/claude-code/issues/75043) — `anthropics/claude-code` — opened 2026-07-07 — bug, has repro, platform:macos, area:agents, reproduced. **The most important finding.** smaqit's phase orchestration on Claude Code is exactly this shape: `smaqit-development` (depth 1) invokes the spec agents (depth 2). If this reproduces on the current version, the orchestrator stalls waiting for a child's completion. The happy-path scenario is what would expose it, so it doesn't block building the test, but a failure there may be this upstream bug rather than a harness defect. Record the Claude Code version and the stall signature in the evidence.
- [#90256 Nested subagent (depth-2) completion notifications route to root session and are discarded on 2.1.250 — regression of #75043 item 2](https://github.com/anthropics/claude-code/issues/90256) — `anthropics/claude-code` — opened 2026-09 — duplicate, has repro, area:agents, regression, platform:wsl. The same class of failure, on WSL.
- [#85066 Headless session exits "success" seconds after dispatching subagents — work orphaned, no error surfaced](https://github.com/anthropics/claude-code/issues/85066) — `anthropics/claude-code` — opened 2026-08-08 — area:core. A `-p` run can exit 0 with the work never done. Checks must never trust the exit code alone (artifacts and the session log decide), and backgrounded subagents must be avoided or waited for (`CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS`).
- [#85129 -p (headless) mode silently kills run_in_background Bash/agents at turn end — no notification loop exists](https://github.com/anthropics/claude-code/issues/85129) — `anthropics/claude-code` — area:core. The same risk as #85066.
- [#85895 CC spawns project .mcp.json stdio server twice per launch](https://github.com/anthropics/claude-code/issues/85895) — `anthropics/claude-code` — bug, has repro, platform:linux, area:mcp, reproduced. `smaqit-plantuml` is a project `.mcp.json` stdio server with a runtime lock (tasks 124/126). A double spawn can contend for that lock, so record the `system/init` `mcp_servers` status in the evidence.
- [#78988 CLAUDE_CONFIG_DIR containing ~ is treated as a cwd-relative path](https://github.com/anthropics/claude-code/issues/78988) — `anthropics/claude-code` — bug, has repro, platform:macos, area:core. Always export an absolute `CLAUDE_CONFIG_DIR`.
- [#97350 apiKeyHelper output that is a claude setup-token OAuth token is rejected with 401; the same token works via CLAUDE_CODE_OAUTH_TOKEN](https://github.com/anthropics/claude-code/issues/97350) — `anthropics/claude-code` — bug, has repro, platform:macos, area:auth. Pass the token through `CLAUDE_CODE_OAUTH_TOKEN`, never through `apiKeyHelper`.
- [#91020 2.1.251 regression: CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1 strips CLAUDE_CONFIG_DIR from hooks and Bash subprocesses](https://github.com/anthropics/claude-code/issues/91020) — `anthropics/claude-code` — bug, has repro, platform:windows, regression. Don't set `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` in the harness.

### Historical (Closed)
- [#59642 Silent failure (exit 0, empty output) with CLAUDE_CONFIG_DIR isolation on 2.1.123+](https://github.com/anthropics/claude-code/issues/59642) — `anthropics/claude-code` — closed (stale). Another reason never to trust exit code 0 alone.
- [#74926 -p/--print: stdio MCP servers stay "pending" — tool snapshot races server startup and never refreshes](https://github.com/anthropics/claude-code/issues/74926) — `anthropics/claude-code` — closed. If it recurs, spec agents would stop with `DESIGN-TOOLCHAIN-UNAVAILABLE`; check the MCP status in `system/init`.
- [#73067 Auth errors in claude -p are written to stdout instead of stderr](https://github.com/anthropics/claude-code/issues/73067) — `anthropics/claude-code` — closed. The preflight auth check should read stdout as well as stderr.
- [#83599 A subagent resumed with SendMessage reports its completion to the main conversation instead of the subagent that resumed it](https://github.com/anthropics/claude-code/issues/83599) — `anthropics/claude-code` — closed.

### Omitted Tools
- Node.js (`nodejs/node`) — used only for the fixture's `node --test`; excluded after relevance prioritization
- Go — used only to build the installer; excluded after relevance prioritization

## Acceptance Criteria

- [ ] `make -C installer e2e-claude` installs into an isolated scratch `HOME`/`CLAUDE_CONFIG_DIR` and verifies the isolation with `claude auth status`. It never writes the user's real `~/.claude`, and reads from it only the claude.ai access token, when no environment token is set
- [ ] Preflight fails fast with a clear message on: no usable credentials (no environment token and no claude.ai login, or a login token with less than 30 minutes left), a Claude Code version below the required minimum, or Node older than 22
- [ ] The happy-path scenario passes:
  - a spec in each of business, functional and stack
  - `smaqit validate` passes
  - `smaqit plan --phase=develop` returns nothing
  - `smaqit status` shows Develop complete
  - `index.html` exists, `node --test` passes, and a development report exists
- [ ] The needs-input scenario passes: turn 1 writes no spec and returns `NEEDS-INPUT`, and a `--resume` turn with the scripted answers produces the specs and reaches the happy-path checks
- [ ] The assisted scenario passes: a `REVIEW-REQUESTED` after each layer, approved specs' hashes unchanged after approval, and the happy-path checks pass at the end
- [ ] The self-invocation scenario passes: a plain-language request with no slash command delegates to `smaqit-development`, shown in the session log, and the happy-path checks pass
- [ ] The phase command bodies define a fallback for when no question tool is available (print the questions and end the turn)
- [ ] Every run is capped by budget, turns and wall-clock time; evidence is saved per scenario; the script exits nonzero on any failure; no credential file is ever written
- [ ] `docs/wiki/workflows/testing-smaqit.md` documents the run, the authentication (`claude setup-token`) and the cost, and its stale user-testing agent reference is removed

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `scripts/e2e-claude.sh` | Create — harness and scenarios |
| `scripts/e2e/hello-site/{requirements,thin-requirements,answers,approve}.md` | Create — fixture |
| `installer/Makefile` | Modify — `e2e-claude` target |
| `commands/smaqit.{development,deployment,validation}.md` | Modify — headless fallback in step 3 |
| `docs/wiki/workflows/testing-smaqit.md` | Modify — Claude Code e2e section, stale reference fix |
| `.smaqit/compendium.md` | Modify — e2e entry |
| `CHANGELOG.md` | Modify |

## Notes

- Planned 2026-10-11 in session 090 (`task.plan`, Mode A). Claude Code headless facts were checked against the official docs (headless, CLI reference, authentication, MCP, sessions):
  - `CLAUDE_CONFIG_DIR` relocates agents, skills, commands, settings, `.credentials.json` and transcripts.
  - `-p` expands `/command` prompts.
  - Project `.mcp.json` servers load without an approval prompt in `-p`.
  - `--permission-prompts none` removes `AskUserQuestion`.
  - `--resume <session_id>` continues a `-p` session.
  - Behavior of an unanswered `AskUserQuestion` in default `-p` mode is undocumented, hence the explicit flags.
- **Overlap with task 130** (In Progress in another session): it edits the same phase command files' `description`/`when_to_use`. This task edits only the bodies, consistent with the 130/131 ownership split. Rebase onto whichever lands first.
- **Undocumented case:** whether a resumed `-p` session can `SendMessage` a subagent from the previous turn. If it can't, the command's fallback (a fresh invocation with `answers` plus `approved_spec_paths`) applies; both pass the checks, and the evidence records which path happened.
- **Follow-ups (not in scope):**
  - retire the broken Copilot `make test-e2e`/`e2e-test.yml`/`run_e2e_test.py`
  - fix `docs/wiki/workflows/quickstart.md`, which still says `smaqit init` creates `.claude/`/`.github/agents/`
  - move to a Bench manifest after smaqit-adk task 037
- Related: task 131 (whose deferred live check this automates), smaqit-adk task 037.
- **Re-planned 2026-10-12** after tasks 114 (v3.9.0) and 130 (v3.10.0) shipped:
  - 130 changed only the command `description`/`when_to_use` lines, so step 13's body fallback is still needed and does not conflict.
  - 130 added a "When to use smaqit" block to `AGENTS.md`, which the self-invocation scenario exercises.
  - 130's deferred live check (plain requests use smaqit, and a directly delegated `NEEDS-INPUT` reaches the user) is closed by the `self-invocation` scenario.
  - Open claude-code issue #85027 (descriptions dropped in the VS Code extension's agent host) doesn't affect this CLI-based test, so a CLI pass says nothing about the VS Code extension.
  - Checked locally at no model cost: all the `claude -p` flags used exist in 2.1.296, and with a scratch `CLAUDE_CONFIG_DIR` and no token, `claude auth status` reports `configDirectory` inside the scratch folder and `loggedIn: false`.

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
