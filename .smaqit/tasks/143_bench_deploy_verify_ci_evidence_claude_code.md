---
status: Not Started
created: "2026-10-20"
---

# Bench: deploy-verify `ci-evidence` on Claude Code, Against the Previous Release

## Description

Task 142 (PR #110, v3.19.0) made existing-k3s deploys on the `existing-cicd-pr` and `ci-push` paths
verify from the deploy run's own record: the `verify` slot defaults to
`smaqit.infrastructure-deploy-verify`'s new `ci-evidence` operation. Its script is covered by 22 stub
`gh` unit cases, but nothing yet shows that a real agent, following the skill text at the `verify`
slot, does the right thing: passes the intended commit (the PR merge SHA), not the deploy run's own
`headSha`; runs `ci-evidence.sh` with the run ID and the environment's row; and never falls back to
`kubectl`, a local kubeconfig, Vault, or a mesh check.

This task adds a Bench manifest for the deploy-verify skill that answers that on Claude Code, and
compares the current skill (1.4.0) with the previous release's (1.3.0, from v3.18.0), which has no
`ci-evidence` and is expected to reach for in-cluster verification instead. It is this repository's
first Claude Code manifest.

Start only after PR #110 merges: the manifest stages the skill from `skills/`, which contains
`ci-evidence` only once task 142 is on `main`.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, smaqit-adk bench, GitHub CLI, jq, Bash
**Platforms/Environments:** None
**Features/Integrations:** claude-code bench adapter, skill treatment install, stub gh run view, deploy-verify ci-evidence, previous-release baseline
**Versions/Constraints:** smaqit-adk v3.1.0, smaqit v3.19.0 baseline, deploy-verify 1.4.0 vs 1.3.0, sequential live runs on one OAuth token

## Design Decisions

- **Target: the deploy-verify skill** (Rui, 2026-10-20): a new manifest at
  `.smaqit/bench/skills/smaqit.infrastructure-deploy-verify/`. The prompt places the agent at the
  `verify` slot after a CI deploy, with the plan entry (`deploy-verify:ci-evidence`), the PR merge
  SHA, the deploy run ID, and the environment. The Deployment agent's CI-path walk is not part of
  this task.
- **Harness: Claude Code** (Rui, 2026-10-20): `adapter: claude-code`, following smaqit-adk's
  "Reusable Claude Code variant block": `permissionMode: dontAsk` with a minimal `allowedTools`
  (`Skill`, `Read`, `Write`, `Bash`), a pinned model, `maxTurns`, `credential:
  CLAUDE_CODE_OAUTH_TOKEN`, `install: [{treatment: skill, as: skill}]`, an explicit
  `execution.timeoutSeconds`, and a header comment showing the inline credential loader. Bench
  isolates `CLAUDE_CONFIG_DIR`, so the globally installed deploy-verify never shadows the treatment.
- **Baseline: the previous release's skill** (Rui, 2026-10-20): the baseline variant stages
  deploy-verify 1.3.0, frozen from `git show v3.18.0:skills/smaqit.infrastructure-deploy-verify/SKILL.md`
  into the manifest's own directory, installed under the same skill name. Both variants get the
  same prompt and fixture; only the skill differs.
- **Everything is stubbed:** the fixture provides `bin/gh` (serves a recorded deploy run for
  `run view` and a fixed status for the `compare` call, and logs every argv), `bin/kubectl` and
  `bin/vault` (log every call and fail), and `bin/smaqit` (serves the Environments JSON for
  `spec environments --env=<env> --json`). No network, cluster, Vault or GitHub.
- **The agent writes its verdict** to `project/verify-outcome.txt`: `PASS` or `FAIL: <reason>`.
- **Separate from task 142** (Rui, 2026-10-20): its own task, branch and PR after #110 merges.

## Implementation Steps

1. Freeze the baseline: write `git show v3.18.0:skills/smaqit.infrastructure-deploy-verify/SKILL.md`
   to `.smaqit/bench/skills/smaqit.infrastructure-deploy-verify/baseline/smaqit.infrastructure-deploy-verify/SKILL.md`
   (a directory, so `as: skill` can install it).
2. Fixture `fixtures/ci-deploy/`: the `bin/` stubs above; a recorded run JSON per Case (two-job k3s
   `deploy.yml`: `build`, and `deploy` with "Verify revision", "Verify in cluster", "Verify over
   HTTPS", step names from `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`);
   an Environments JSON with a `private-mesh` row; a minimal `AGENTS.md`.
3. Resolve how the stub `gh` reaches `ci-evidence.sh`, which calls `gh` from `PATH`: check whether
   the claude-code adapter accepts `PATH` in `claudeCode.environment.set` (and with which
   interpolation); otherwise have the prompt run the script as `PATH="$PWD/project/bin:$PATH" bash …`.
   Record which in the manifest's header comment.
4. Cases, each self-contained and single-shot (no clarifying questions; nobody will answer):
   - `dispatcher-later-head`: run `headSha` ≠ merge SHA, compare `ahead` → expect `PASS`, a compare
     call whose base is the merge SHA, no `kubectl`/`vault` calls.
   - `wrong-run`: compare `behind` → expect `FAIL` naming the wrong SHA; no `kubectl`/`vault` calls.
   - `skipped-verify-step`: "Verify in cluster" skipped for the private-mesh row → expect `FAIL`;
     no `kubectl`/`vault` calls (no fallback to in-cluster verification).
   Expectations are `command`-type checks on `project/verify-outcome.txt` and the stubs' argv logs.
5. Variants `current-skill` (treatment `../../../../skills/smaqit.infrastructure-deploy-verify`) and
   `previous-release` (treatment `baseline/smaqit.infrastructure-deploy-verify`), both installed
   `as: skill`; prompt names the skill and tells the agent to invoke it with the Skill tool.
6. `smaqit-adk bench validate` the manifest; fix diagnostics.
7. Live run via `smaqit.bench-run` (credential loaded inline, runs kept sequential); record per-variant
   pass rates. Expected: `current-skill` passes all three Cases; `previous-release` fails the
   no-`kubectl` expectation at least once. If it doesn't, record why rather than tuning the prompt
   toward the expected result.
8. Update `.smaqit/bench/README.md` with a pointer to the repo's first Claude Code manifest.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `.smaqit/bench/skills/smaqit.infrastructure-deploy-verify/bench.yaml` exists with the three Cases and the `current-skill` / `previous-release` claude-code variants, and `smaqit-adk bench validate` passes
- [ ] The fixture is fully stubbed: no network, cluster, Vault or GitHub access; `kubectl` and `vault` stubs log and fail every call
- [ ] Expectations check the outcome file, the compare call's base SHA, and the absence of `kubectl`/`vault` calls with `command`-type checks
- [ ] A live run completed sequentially with an inline-loaded credential, and its per-variant results (and the reason for any unexpected result) are recorded in Findings
- [ ] `.smaqit/bench/README.md` points to the new Claude Code manifest

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
| .smaqit/bench/skills/smaqit.infrastructure-deploy-verify/bench.yaml | Create |
| .smaqit/bench/skills/smaqit.infrastructure-deploy-verify/prompts/*.md | Create |
| .smaqit/bench/skills/smaqit.infrastructure-deploy-verify/fixtures/ci-deploy/** | Create |
| .smaqit/bench/skills/smaqit.infrastructure-deploy-verify/baseline/smaqit.infrastructure-deploy-verify/SKILL.md | Create |
| .smaqit/bench/README.md | Modify |

## Notes

- Depends on task 142 (PR #110) merging.
- Bench runs spend real model usage on the claude.ai login; keep them sequential (concurrent
  headless calls on one login token have been reported to invalidate the session).
- `.smaqit/bench/` is repo-local data, not shipped by the installer, so this task changes no release
  content; it can ship as a PATCH-level or chore release per `release-analysis`.
