---
status: Completed
created: "2026-09-12"
parent: "117"
mode: Assisted
started: "2026-09-12"
completed: "2026-09-12"
---

# Request k3s App Onboarding via Platform-Repo PR

## Description

Task 116 documented skill #1 (the platform-owned infrastructure repo's own onboarding
automation — registry file + converge workflow, already live and hardened in a downstream
project's own infra repo) as an agent-executable playbook, `smaqit.infrastructure-onboard-k3s-app`.
Task 117 built skill #3 (the app-side deploy mechanism that consumes an already-issued kubeconfig,
`smaqit.infrastructure-deploy-k3s-app`). Neither covers skill #2: an app project's own means of
*requesting* onboarding from its own repo, without needing the broad direct-commit-and-dispatch
access task 116's playbook assumes.

This task builds skill #2 as a new, narrowly-scoped, standalone skill:
`smaqit.infrastructure-request-k3s-onboarding`. Given a declared platform repo, a registry file
path, and operator-supplied entry content, it opens a PR against that external repo adding the
app's registry entry for one target machine, then gates on that PR being approved and merged
before reporting success — mirroring the PR-as-human-gate pattern `smaqit.feature-new` already
uses, rather than requiring the app side to hold direct commit/`workflow_dispatch` rights on a
repo it doesn't own. The actual convergence (Namespace/RBAC/kubeconfig issuance) remains the
platform team's own process, triggered by their own repo reacting to the merge — outside this
skill's or this task's control.

This task ships as a child of task 117: it joins task 117's existing branch/worktree and inherits
its Assisted mode, since it only matters for the `existing-k3s` path task 117 introduced and both
ship together in one PR/release.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions/API (cross-repo `gh pr create`), Git, smaqit skill-definition conventions
**Platforms/Environments:** A separate, platform-owned GitHub repository hosting the k3s onboarding registry — never a real downstream project, repository, or machine name
**Features/Integrations:** `smaqit.new-greenfield-project` Phase 4/5, `smaqit.infrastructure-vault-loader`, `smaqit.infrastructure-repo-config`, `templates/specs/infrastructure.template.md`, task 117 (parent, `smaqit.infrastructure-deploy-k3s-app`), task 116 (`smaqit.infrastructure-onboard-k3s-app`, referenced as the platform-side playbook this skill's PR ultimately triggers)
**Versions/Constraints:** Credential scoped only to the platform repo (`contents:write` + `pull_requests:write`, never broader — no `workflow_dispatch` needed on the app side); no registry-file schema assumed, since task 116's Provenance is anonymized and the actual format is unknown to smaqit

## Design Decisions

- **Child of task 117**, not a standalone task — joins its branch/worktree, ships in one PR, inherits Assisted mode. Confirmed explicitly by the user.
- **Two PRs, one per environment** (test at Phase 4, prod at Phase 5), not one PR covering both upfront — matches how onboarding is naturally needed at different times and keeps each PR's diff/review scope minimal. Confirmed explicitly by the user.
- **Format-agnostic, operator-supplied registry content.** The skill takes `platform_repo` (owner/repo), `registry_file_path`, and `entry_content` (exact text/diff to add) as declared inputs. It never assumes or validates a registry-file schema — its job is opening the PR and gating on merge, not authoring registry syntax it can't verify. Confirmed explicitly by the user.
- **`machine_slug`, not `environment`, is the skill's primary targeting input.** The registry file is inherently per-machine (task 116: "a git-committed *per-machine* app registry file"), and Vault's kubeconfig storage (task 117, revised) now keys by machine-slug (`secret/apps/<app-slug>/<machine-slug>/kubeconfig`), not environment name. The calling flow (`smaqit.new-greenfield-project`) resolves which machine-slug applies to which environment from the Infrastructure spec — this skill itself only ever operates on one named machine per invocation, consistent with `smaqit.infrastructure-deploy-k3s-app`'s own `MACHINE_SLUG` resolution.
- **A new, narrowly-scoped credential**, `secret/apps/<app-slug>/platform-repo` — a PAT scoped only to `contents:write` + `pull_requests:write` on the platform repo. Distinct from the existing `secret/apps/<app-slug>/github` field, which is scoped to the app's own repo (`variables:write`, used for `TF_VAR_github_token`). No `workflow_dispatch` permission is needed on the app side at all, since triggering convergence is the platform repo's own concern once the PR merges.
- **Gate follows `smaqit.feature-new`'s existing human-gate philosophy** (create the PR, pause, re-check on next invocation) rather than a busy-poll loop — no such loop exists anywhere else in this codebase, and a PR-review gate blocks on human review time, which a tight polling interval wouldn't shorten.
- **Idempotent by deterministic naming.** Branch/PR name derives predictably from `<app-slug>-<machine-slug>` so a re-invocation detects and reuses an existing open PR (`gh pr list -R <platform_repo> --head <branch>`) rather than opening a duplicate.
- **Compiled directly, not definitions-only first.** The design is narrow, precedented (mirrors `smaqit.feature-new`'s already-proven PR-gate shape), and low-risk enough not to need a synthesis/proof period the way the original tornado skill did.
- **Not embedded in the generated `deploy.yml`.** A PR-review gate on every push would stall production deploys on reviewer availability; this only runs as a one-time setup step per machine-slug (Phase 4 for test, Phase 5 for prod), not per-push CI.
- **Kubeconfig retrieval stays out of scope**, unchanged from task 117's existing precondition — this skill's job ends at "PR merged"; the platform team's own process still hands off the resulting kubeconfig out-of-band, and populating Vault with it remains `smaqit.infrastructure-vault-loader`'s existing manual-paste step.
- **`templates/specs/infrastructure.template.md` gets one new `Constraints` table row**, `Platform Repo` (`existing-k3s`-only), not a new section — mirrors the existing compact external-fact rows (Target Environment/Geographic/Budget).
- **No credential-scope preflight guard** (e.g. a `namespace-guard.sh`-style refusal if the platform-repo PAT has broader-than-declared rights) — out of scope for this task; noted as a possible future hardening, not built here.

## Implementation Steps

1. Read `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` (task 116's playbook
   this skill's PR ultimately triggers), `skills/smaqit.feature-new/SKILL.md` (the existing
   PR-as-human-gate pattern to mirror — deploy-PR creation, pause, re-check on next invocation),
   and `skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md` (task 117, parent — for the
   `MACHINE_SLUG` resolution convention to stay consistent with).
2. Author `.smaqit/definitions/skills/smaqit.infrastructure-request-k3s-onboarding.md` and compile
   `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` directly (no separate
   definitions-only period):
   - **Inputs:** `platform_repo` (owner/repo), `registry_file_path`, `entry_content`
     (operator-supplied, format-agnostic), `machine_slug`.
   - **Steps:** authenticate with the `secret/apps/<app-slug>/platform-repo` credential →
     create/reuse a deterministically-named branch on `platform_repo` (`onboard-<app-slug>-<machine_slug>`)
     → commit `entry_content` at `registry_file_path` → `gh pr create -R <platform_repo>` (or reuse
     an existing open PR for that branch) → report the PR URL → gate: check merge state via
     `gh pr view -R <platform_repo> <pr-number> --json state,mergedAt` (single check per
     invocation, not a loop) → on merge, report success; on not-yet-merged, report "pending" and
     stop cleanly (re-entrant on next invocation, same as `smaqit.feature-new`'s own PR gate).
   - **Never**: trigger `workflow_dispatch` on the platform repo, read/write anything else in that
     repo, or attempt to fetch/store the resulting kubeconfig (out of scope — stays manual).
3. Add a `Platform Repo` row to `templates/specs/infrastructure.template.md`'s `## Constraints`
   table: `| Platform Repo | owner/repo hosting the onboarding registry (existing-k3s only) | PR
   must target that repo, never this project's own |`.
4. `skills/smaqit.infrastructure-vault-loader/`: add `secret/apps/<app-slug>/platform-repo` (field
   `token`) to the namespace-convention table and `existing-k3s` scheme-detection section in
   `SKILL.md`; extend `scripts/load-credentials.sh`'s `existing-k3s` branch to also
   check/populate it (standard delete-and-repopulate shape on `rotate-credential.sh`, unlike
   `kubeconfig`'s re-prompt-only shape, since this is a smaqit-managed PAT the operator can freely
   regenerate — not a platform-issued artifact).
5. `skills/smaqit.new-greenfield-project/SKILL.md`: add new `→ existing-k3s:` callouts —
   - Phase 4, before the deploy-skill invocation (Step 6): invoke
     `smaqit.infrastructure-request-k3s-onboarding` for the test environment's registered
     machine-slug; gate on merge before proceeding.
   - Phase 5, before the CI/CD trigger step: same for the prod environment's registered
     machine-slug.
   - Update Gotchas/Scope to mention the new skill; bump `metadata.version`.
6. Build the installer (`installer/Makefile`) and confirm the new skill lands in both
   `installer/skills-shared/` and `installer/skills-claude/`; bump the two hardcoded skill-count
   assertions in `installer/main_test.go`; run `go vet ./...` and `go test ./... -count=1`.
7. Add a `CHANGELOG.md` entry under `[Unreleased]/Added`.
8. Grep every new/modified file for real project, repository, or machine names to confirm
   anonymization is clean.
9. Manual dry-run: against a real scratch/test GitHub repo the operator controls (never a real
   downstream project), confirm the skill actually opens a PR with the supplied `entry_content`,
   correctly detects the PR merging, and correctly reuses an already-open PR on a second
   invocation rather than duplicating it.

## Known Issues Triage
**Triaged:** 2026-09-12
**Tools searched:** GitHub CLI (`cli/cli`), Git (`git/git`)
**Result:** Advisory

### Blocking Issues
_None._

### Advisory Issues
- [#10093 Enhance `gh pr create` to support cross repo pull requests within the same organization](https://github.com/cli/cli/issues/10093) — `cli/cli` — opened 2024-12-16 — `enhancement`, `more-info-needed`, `needs-triage` — fetched detail to confirm scope: this is specifically about `gh pr create -H user:branch`'s fork-based `headRefName` flow (an upstream-repo-to-user-fork PR). It does **not** apply to this task's actual design, which pushes a branch directly to the platform repo using a collaborator-scoped `contents:write` token and opens a same-repo PR via `gh pr create -R <platform_repo>` — no fork, no `-H user:branch`. Confirms this task's Design Decisions should keep stating that explicitly, to avoid ever reaching for the limited fork-based flow this issue describes.

### Historical (Closed)
_None relevant._ (`cli/cli`'s closed search returned 10 issues; none confirmed a match to this task's platform/feature dimensions — general `gh pr`/`gh status` bugs unrelated to cross-repo branch pushes or PR creation.)

### Unresolvable Tools
_None — both named tools resolved to a repository._

### Omitted Tools
_None — 2 resolved repositories, within the 5-repository limit. "GitHub REST API" was not searched separately (same underlying platform as GitHub CLI; `gh pr create`/`gh pr view` are the actual interface this task's skill uses)._

### Search Warnings
_None._

### Notes
- `git/git` searches (open and closed) returned zero results for "remote branch"/"push" — Clear for that repository specifically.

## Acceptance Criteria

- [x] `.smaqit/definitions/skills/smaqit.infrastructure-request-k3s-onboarding.md` and compiled
      `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` exist; no real project,
      repository, or machine name anywhere
- [x] Skill takes `platform_repo`, `registry_file_path`, `entry_content`, and `machine_slug` as
      declared inputs; never assumes or validates a registry-file schema
- [x] PR-opening is idempotent — a second invocation for the same `app-slug`/`machine_slug` reuses
      the existing open PR rather than opening a duplicate
- [x] The skill gates on PR merge (single state check per invocation, re-entrant — not a busy-poll
      loop) before reporting success; never triggers `workflow_dispatch` on the platform repo and
      never attempts to fetch or store the resulting kubeconfig
- [x] `templates/specs/infrastructure.template.md` has a new `Platform Repo` row in `##
      Constraints`, documented as `existing-k3s`-only
- [x] `smaqit.infrastructure-vault-loader` supports `secret/apps/<app-slug>/platform-repo`
      (distinct from `github`), with `load-credentials.sh`/`rotate-credential.sh` support
- [x] `smaqit.new-greenfield-project` Phase 4 (test) and Phase 5 (prod) each gate on this skill
      before invoking `smaqit.infrastructure-deploy-k3s-app`; `metadata.version` bumped
- [x] Installer rebuilds cleanly (new skill present in both `installer/skills-shared/` and
      `installer/skills-claude/`); `go vet`/`go test` pass with skill-count assertions updated
- [x] `CHANGELOG.md` has a new entry under `[Unreleased]/Added`
- [ ] A manual dry-run against a real scratch repo confirms the skill opens a PR, detects its
      merge, and correctly reuses an existing open PR on a repeat invocation — **not done**: no
      scratch repo or `platform-repo`-scoped credential was available in this environment; see
      Follow-up identified below

## Findings

**Implementation approach:**
- Authored `.smaqit/definitions/skills/smaqit.infrastructure-request-k3s-onboarding.md` mirroring task 117's own Provenance shape (authored directly, no downstream contribution), then compiled `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` directly (no separate definitions-only period), with inputs `platform_repo`/`registry_file_path`/`entry_content`/`machine_slug`, an idempotent open-PR check, and a single-state-check merge gate mirroring `smaqit.feature-new`'s existing pause-and-recheck pattern.
- Added a `Platform Repo` row to `templates/specs/infrastructure.template.md`'s `## Constraints` table.
- Extended `smaqit.infrastructure-vault-loader` (SKILL.md + both scripts) with a new `secret/apps/<app-slug>/platform-repo` credential, distinct from `github`, using the standard delete-and-repopulate rotation shape.
- Wired `smaqit.new-greenfield-project` Phase 4 and Phase 5 to invoke the new skill and gate on merge before their respective kubeconfig-loading/deploy steps; bumped `metadata.version` to 1.7.0; added a Gotchas entry.
- Rebuilt the installer (skill count 28→29 in `installer/main_test.go`), ran `go vet`/`go test -count=1` (pass), grepped all new/modified files for real project/repo/machine names (clean).

**Decisions made:**
- Confirmed as a child of task 117 (shares its branch/worktree/Assisted mode) rather than a standalone task, per explicit user direction during planning.
- Two PRs per app (one per environment/machine-slug), not one PR covering both — each a one-time Phase 4/Phase 5 setup step, never embedded in the generated `deploy.yml`.
- `machine_slug`, not `environment`, is the skill's primary targeting input — kept consistent with task 117's same-day Vault revision to `secret/apps/<app-slug>/<machine-slug>/kubeconfig`.
- Deliberately avoids `gh pr create -H user:branch` (fork-based cross-repo PRs) — issue triage found `cli/cli#10093`, an open upstream limitation in that exact flow; the skill instead uses a directly-scoped collaborator credential to push a branch straight to the platform repo.
- No credential-scope preflight guard (e.g., a `namespace-guard.sh`-style refusal if the platform-repo PAT has broader-than-declared rights) — explicitly deferred, not built in this task.

**Blockers encountered:**
- None during implementation. The one real gap is the final acceptance criterion (see Follow-up below), which was a known limitation of this environment rather than an implementation blocker.

**Follow-up identified:**
- The manual dry-run acceptance criterion (open a real PR against a scratch repo, confirm merge detection and idempotent reuse) was not performed — no scratch GitHub repo or `platform-repo`-scoped PAT was available in this session. Left unchecked rather than falsely marked done; run it manually against a real repo before relying on this skill in production. Confirmed with the user (2026-09-12) to complete the task now rather than block on setting this up.
- No credential-scope preflight guard exists for the `platform-repo` PAT (unlike `namespace-guard.sh` for the k3s kubeconfig) — noted as a possible future hardening, not required for this task.

## Files to Create / Modify

| File | Action |
|------|--------|
| `.smaqit/definitions/skills/smaqit.infrastructure-request-k3s-onboarding.md` | Create |
| `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` | Create |
| `templates/specs/infrastructure.template.md` | Modify (new `Platform Repo` Constraints row) |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify (new `platform-repo` credential) |
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify |
| `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh` | Modify |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify (Phase 4/5 callouts, version bump) |
| `installer/main_test.go` | Modify (skill-count bump) |
| `CHANGELOG.md` | Modify |

## Notes

Origin: identified 2026-09-12 during a 3-way decomposition of k3s onboarding, prompted by a direct
question during task 117's post-implementation review ("did you add both app onboard and deploy to
the feature-new and greenfield skill flows?"). The three skills are: (1) the platform repo's own
onboarding automation — task 116's `smaqit.infrastructure-onboard-k3s-app`, an agent-executable
playbook, not reusable scaffolding, written for an actor with full direct access; (2) THIS task,
the app-side request-and-gate step, previously unbuilt; (3) `smaqit.infrastructure-deploy-k3s-app`
— task 117. Task 116's own Findings assumed task 117 already covered the deferred
routing/wiring work; on inspection this was inaccurate — task 117 explicitly decided to never
request onboarding, so this gap was real and unaddressed until this task.

This task's design is downstream of a same-day revision to task 117's own Vault work: the
kubeconfig storage convention was corrected from `secret/apps/<app-slug>/kubeconfig` (test/prod
fields) to `secret/apps/<app-slug>/<machine-slug>/kubeconfig` (one value per machine-slug, no
`secret/machines/*` counterpart) specifically because a flat, environment-keyed structure couldn't
represent test and prod living on genuinely different machines. This task's own `machine_slug`
input follows that same convention rather than an `environment` input, to stay consistent.
