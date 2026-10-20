# Toward CI-Only k3s Deploys

## Metadata

- **Date:** 2026-10-20
- **Session focus:**
  - stop CI-path existing-k3s deploys from needing the operator's machine on the mesh or a local
    kubeconfig, and verify them from the deploy run's own record;
  - plan a Claude Code benchmark for that change;
  - plan the larger change that every deploy goes through CI/CD workflows.
- **Tasks completed:** 142 (v3.19.0, PR #110).
- **Tasks created:** 143 (bench), 144 (all deploys through CI/CD). Task 142 had been created just
  before the session.
- **Tasks edited:** 137 (its `ci-evidence` work moved to 142).
- **Releases:** v3.19.0 (#110).

## Actions Taken

**Session start**
- Loaded the context. Task 142 was new and first in the planner; its creation commit was still local
  (correct by design: task bookkeeping stays on local `main`).

**Task 142: CI-driven deploys run no local target checks**
- **`task.plan 142`:** three discovery agents checked the plan against the code. Two findings changed
  it:
  - `metadata.smaqit-slots` already allowed two operations per slot, but `selectOperation` would then
    refuse a binding that names deploy-verify without an `operation`;
  - on `existing-cicd-pr`, `post-merge-deploy.yml` dispatches `deploy.yml` at `main`'s tip, so the
    run's `headSha` may be a later commit than the merge, and the agent was passing the run's own
    `headSha` as `--expected-sha`, making the SHA check compare the run with itself.
- **You decided:** `--expected-sha` is the intended commit (merge SHA or pushed SHA), proven by
  ancestry with one `gh api` compare call; the Environments row comes in as JSON, like `mesh-check.sh`.
- **Built (Assisted):**
  - **Resolver:** `slotDef` gained per-path, per-mode defaults; `platform-prereqs` defaults to none on
    both CI paths; existing-k3s `verify` defaults to the new `ci-evidence` operation there.
  - **deploy-verify 1.4.0:** its first script, `ci-evidence.sh`, decides from
    `gh run view --json headSha,status,conclusion,jobs`: a finished, successful run at the intended
    commit or a descendant, with "Verify revision" and the row's verify step successful.
  - **Text:** the Deployment agent, feature-new, project-new, deploy-k3s-app, the deployment-slots
    wiki and the compendium.
  - **Tests:** resolver cases on both CI paths, 22 stub-`gh` cases, a smoke-test assertion; four
    deliberate breaks each caught.
- **Released:** Phase 1 pushed the branch and opened PR #110 (`Prepare release v3.19.0`); you merged
  it; Phase 2 merged `origin/main` into local `main`, completed the task and cleaned up. The
  post-merge workflow published v3.19.0.

**Push credentials**
- The branch push returned 403: the default `gh` login is a downstream organisation's token (`organizations/<org>`),
  which can't write to `ruifrvaz/smaqit`. You chose the Vault route: after you unsealed Vault,
  `vault-gh-token.sh --login apps/smaqit` for the push and PR, then `--login organizations/<org>`
  to switch back. You asked whether the token is exposed to the agent: it is piped from Vault to
  `gh auth login --with-token` and never printed; `gh auth status` showed it masked.

**Task 143: Claude Code bench for `ci-evidence` (created)**
- `smaqit.bench-scaffold` gathered the design: target the deploy-verify skill, run on Claude Code
  (`adapter: claude-code`), compare against the v3.18.0 skill (1.3.0) as baseline, stub every tool,
  and land it as its own task after #110.

**Local deploys, then task 144 (created)**
- You asked why `mesh-check.sh` still says it is "for local runs", and when smaqit deploys locally at
  all. Three cases: project-new's Phase 4 dev sweep, a bare `/smaqit.deployment` (its path defaults
  to `standard`), and manual fallbacks; deploy-k3s-app's own gotchas already preferred CI even for a
  first deploy.
- **`task.plan` (new idea):** three discovery agents mapped project-new and the workflow sets, the
  resolver paths and agent defaults, and every skill that acts on a target locally. k3s is already
  CI-complete; VM modes have large workflow gaps (single-host `deploy.yml`, no vhost/compose/`.env`,
  no host bootstrap or TLS in CI, no Python templates, a local first apply).
- **You decided:** deploys run only in CI/CD (GitHub Actions by default, with a documented
  "ci-monitor" seam); local runs only bootstrap and configuration; enforced for existing-k3s now,
  VM modes transitional; no break-glass path; project-new Phase 4 becomes a CI deploy to the test
  environment through a new `ci-bootstrap` path.

## Problems Solved

- **A circular SHA check** (the run compared with itself on the dispatcher path): `ci-evidence`
  takes the intended commit and proves ancestry.
- **A binding without `operation` would break silently** once deploy-verify declared two operations:
  covered by a test and an action-required CHANGELOG note.
- **Push 403 on the task branch:** resolved with the `apps/smaqit` Vault token through
  `vault-gh-token.sh --login`, restored afterwards.
- **Smaller issues:**
  - a new binding test first bound `verify` in a project whose `provision` binding
    `existing-cicd-pr` refuses; it now targets the `test` environment;
  - one Edit failed because the file hadn't been read, so the first smoke run lacked the new
    assertion; it was re-applied and re-run;
  - the triage block was written without issue dates (filtered out of the search output, and the
    skill forbids repeating identical searches).

## Decisions Made

- **CI-path verification:** `ci-evidence` decides only from GitHub's run record plus one compare
  call; no kubeconfig, Vault, `kubectl`, mesh or logs. VM modes keep their HTTP verify.
- **Releases:** breaking changes still ship as MINOR with an action-required note (v3.19.0: bindings
  of deploy-verify must name `operation`).
- **Task 137:** its removed `ci-evidence` step became a narrower `standard`-path local-verify step;
  task 144 will drop it.
- **Bench design (143):** skill target, Claude Code, previous-release baseline, separate task.
- **Deploy direction (144):** CI/CD only for deploys, k3s enforced now, VM transitional, Phase 4
  through `ci-bootstrap`, GitHub Actions with a documented seam, no backwards compatibility.
- **Pushing this repository:** use `vault-gh-token.sh --login apps/smaqit`, then switch back to
  the previous `organizations/<org>` login; ask before switching.

## Files Modified

**smaqit (shipped in PR #110, v3.19.0):**
- `installer/slots.go`, `installer/slots_test.go`, `installer/skill_scripts_test.go`.
- `skills/smaqit.infrastructure-deploy-verify/SKILL.md` and new `scripts/ci-evidence.sh`.
- `skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md` and `scripts/mesh-check.sh` (comment).
- `skills/smaqit.feature-new/SKILL.md`, `skills/smaqit.project-new/SKILL.md`.
- `agents/deployment.md`, `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`.
- `scripts/smoke-test-installer.sh`, `CHANGELOG.md`.

**Task records (local `main`):** tasks 142 (plan, triage, findings, completion), 137 (rescope),
143 and 144 (new), `PLANNING.md`, `.smaqit/references/project-research.md` (task 142 block),
`smaqit.code-workspace`.

## Next Steps

- **Task 144** (High): all deploys through CI/CD; start with `task.start 144`.
- **Task 143** (Medium): the Claude Code bench for `ci-evidence`.
- **Task 137**: after 144, which drops its remaining local-verify step.
- **Optional:** `smaqit update` to install v3.19.0 on this machine; downstream projects pick it up on
  their next update.
- **Carried over:** task 126, task 113, smaqit-extensions task 044, the downstream platform
  repository's upgrade (PR #169 and v3.18.0+).

## Session Metrics

- **Duration:** one day (2026-10-20).
- **Tasks:** 1 completed (142), 2 created (143, 144), 1 rescoped (137).
- **Releases:** 1 (v3.19.0, PR #110).
- **Discovery agents:** 6 Explore agents (3 for `task.plan 142`, 3 for the CI-only plan).
- **Tests:** 22 `ci-evidence.sh` cases, new resolver cases on both CI paths, a smoke-test assertion;
  4 deliberate breaks, each caught.
- **Issue triage:** Advisory (no blocking issues).
