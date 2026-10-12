# Claude Code Headless E2E Harness

## Metadata

- **Date:** 2026-10-11 → 2026-10-12. This continues the conversation that session 090 documented through task 131's release.
- **Session focus:** A Claude Code end-to-end test that installs smaqit into an isolated scratch folder and drives headless `claude -p` to build a site through smaqit. It verifies live the deferred contracts of tasks 131 and 130.
- **Tasks completed:** 132 (released v3.12.0, PR #101)
- **Tasks created:** 132
- **Tasks referenced:** 131 (whose live check this closed), 130 (whose self-invocation live check this closed), 134 (a parallel session, released v3.11.x), 114 and 133 (released by other sessions), smaqit-adk 037

## Actions Taken

- **Planned task 132** (`task.plan`, Mode A). Discovery came from two threads:
  - **The existing e2e path:** the only one, Copilot-SDK `make test-e2e`, was broken because its `smaqit.ci-testing` agent had been deleted. Session 058 had called Claude Code e2e coverage "new scope".
  - **Claude Code's headless contract**, from the docs:
    - `CLAUDE_CONFIG_DIR` relocates config, credentials and transcripts
    - `-p` expands slash commands
    - project MCP servers load without approval
    - `--permission-prompts none` removes `AskUserQuestion`
    - `--resume <session_id>` continues a session

  The user chose a standalone script, credentials from an environment token, three scenarios and a static-site fixture. Task 132 was created.
- **Ran a session recap.** It showed that tasks 114 (v3.9.0) and 130 (v3.10.0) had shipped in parallel.
- **Re-planned task 132 for v3.10.0** (Mode B):
  - Checked what 130 shipped: front-matter-only command changes, so the headless fallback in the command bodies was still needed.
  - Verified the local Claude Code CLI flags and isolation, at no model cost.
  - The user added a fourth `self-invocation` scenario, which closes 130's deferred live check.
- **Started task 132:** research-map block, then issue triage, which came back Advisory. Key risk: claude-code #75043, where a subagent's own subagents run detached.
- **Implemented the harness:**
  - `scripts/e2e-claude.sh`
  - `scripts/e2e/log_check.py`
  - the `scripts/e2e/hello-site/` fixture
  - the `make e2e-claude` target
  - a headless fallback in the phase command bodies and the `AGENTS.md` return-signal rule
  - docs
- **Verified without spending model usage:** lint, synthetic log traps, preflight negatives, and a full pass with an invalid token.
- **Switched credentials to the current login.** The user rejected API-billed credentials, so the harness reuses the current claude.ai login's access token through `CLAUDE_CODE_OAUTH_TOKEN`. A probe confirmed a regular access token works and no credential file is written.
- **Ran the scenarios live, four rounds:**
  - `happy-path` passed first.
  - `needs-input`, `assisted` and `self-invocation` failed on harness defects. The user chose auto permission mode for one of them; the rest were fixed directly.
  - Final result: **all four scenarios pass**.
- **Retired the Copilot e2e path and refreshed `quickstart.md`**, at the user's request (follow-up 2).
- **Completed task 132:**
  - PR #101, `Prepare release v3.12.0`.
  - `origin/main` (v3.11.1) was merged into the branch first, to avoid a CHANGELOG conflict.
  - After the user merged it, Phase 2 cleanup ran and the v3.12.0 release, with 5 binaries, was confirmed.

## Problems Solved

- **Per-command Bash allowlist too brittle:** agents compose commands such as `ls … && smaqit status | head`, and the allowlist denied them. Resolved with `--permission-mode auto`, which the user chose: 0 denials.
- **Signal detection missed background subagents:** their results arrive as `system/task_notification` events and as extra `result` events, not as `tool_result` blocks. The checker now reads those too.
- **Checker crash:** the `message` of a `system/permission_denied` event is a string.
- **Misleading run summary:** a run emits several result events (the background-delegation notice, the relayed result, an empty closing event). The summary now combines them.
- **False "regenerated" alarm in assisted mode:** end-of-phase status updates, ticked checkboxes and the design-sequence link changed the specs legitimately. The digest now normalizes these, which was confirmed against a saved turn-4 diff.
- **Weak empty-project checks:** `node --test` passes when it finds no tests, so the check now requires at least one passing test.
- **Untrustworthy exit codes:** a 401 run reported `subtype: success`. Checks rely on files and the session log, never the exit code alone.
- **CHANGELOG conflict with v3.11.x:** resolved by merging `origin/main` into the task branch before writing the versioned section.

## Decisions Made

- **A standalone script now;** a Bench manifest after smaqit-adk task 037 (user).
- **Credentials** (user): an environment token, or else the current claude.ai login's access token only. The refresh token is never used, nothing is written to the scratch folder, and the run refuses to start with under 30 minutes left.
- **Permissions:** auto mode rather than a per-command allowlist or bypassing permissions (user).
- **Four scenarios:** `happy-path`, `needs-input`, `assisted`, `self-invocation` (user).
- **Static "Hello, Mario!" site fixture** (user).
- **Local-only:** a manual pre-release check, not CI, because every run spends model usage.
- **Scope extension** (user): retire the Copilot-SDK e2e path and fix the quickstart's out-of-date text. The `.test-venv` cleanup in `make uninstall` and `.gitignore` was kept for leftovers.
- **Version:** v3.12.0, a minor release.

## Files Modified

- **Task 132, shipped in PR #101 (v3.12.0):**
  - `scripts/e2e-claude.sh` and `scripts/e2e/log_check.py`: new
  - `scripts/e2e/hello-site/{requirements,thin-requirements,answers,approve,plain-request}.md`: new
  - `installer/Makefile`: `e2e-claude` added; `test-e2e` and `test-e2e-clean` removed
  - `commands/smaqit.{development,deployment,validation}.md` and `templates/AGENTS.md.template`: headless fallback
  - `docs/wiki/workflows/testing-smaqit.md`: Claude Code e2e section; stale user-testing agent reference removed
  - `docs/wiki/workflows/quickstart.md`: matches global installation and current invocation
  - `.smaqit/compendium.md`: e2e entry; the Bench entry links to it
  - `CHANGELOG.md`: v3.12.0
  - Deleted: `.github/workflows/e2e-test.yml`, `.github/scripts/{run_e2e_test.py,requirements.txt,SUMMARY.md}`, `docs/test-cases/mario-hello.automated.md`
- **Bookkeeping:**
  - `.smaqit/tasks/132_claude_code_headless_e2e_test.md` and `PLANNING.md`: lifecycle from create through Completed
  - `.smaqit/references/project-research.md`: task 132 block
  - `smaqit.code-workspace`
- **Memory:** `task_state_132.md`, an update to `task_state_131.md` recording the live check as passed, and `MEMORY.md`

## Next Steps

- **Product:** phase agents should invoke specification agents with `run_in_background: false`, to stop relying on polling as the workaround for claude-code #75043.
- Rerun `happy-path` under auto permission mode in the next full `make e2e-claude` run.
- Move the e2e test to a Bench manifest once smaqit-adk task 037 lands; 037 still needs `task.plan`.
- **Carried over:**
  - task 126 (install-path lock), which has a recommended design
  - task 113 (where its warning-only output goes)
  - the installer never prunes skills a release retires

## Session Metrics

- **Tasks:** 1 created and completed (132). Release: v3.12.0 (PR #101), 5 platform binaries.
- **Implementation:** 22 files in the feature commit (+631/−1093, including the retired Copilot path).
- **Live runs:** 4 rounds. Final result: 4 of 4 scenarios pass on Claude Code 2.1.296. Notional cost about $0.7–0.9 per scenario on the claude.ai subscription, with 0 permission denials.
- **Harness defects found and fixed through live runs:** 6
