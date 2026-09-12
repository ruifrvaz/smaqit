# smaqit.infrastructure-request-k3s-onboarding

## Description

Use when an app project targeting `provisioning_mode: existing-k3s` needs to request onboarding
of one of its environments onto a platform-owned, self-hosted k3s cluster whose own onboarding
process accepts requests via a PR-mergeable registry file. Given a declared platform repository, a
registry-file path, and operator-supplied entry content, opens a PR against that external
repository adding the app's registry entry for one named machine, then gates on that PR being
approved and merged before reporting success. Used in Phase 4 (test) and Phase 5 (prod) of
`smaqit.new-greenfield-project`, before invoking `smaqit.infrastructure-deploy-k3s-app`. This is
the app-side *request* half of onboarding: a smaqit-declared, opinionated convention for how an
app repo requests k3s onboarding, owned entirely by this skill — never inherited from or dependent
on how any particular platform repo actually fulfills the request. Whatever happens after the PR
merges (Namespace/RBAC/kubeconfig issuance, or anything else) is entirely the platform team's own
process, outside this skill's knowledge or control. This skill never touches that fulfillment
mechanism directly, and never requests broader access than opening a PR.

## Provenance

- `synthesized: false`
- `authored-for-project: smaqit itself`
- `authored-date: 2026-09-12`
- Authored directly into canonical `smaqit` as a child of task 117 (`smaqit.infrastructure-deploy-k3s-app`), closing a gap identified during that task's post-implementation review: neither an infra repo's own onboarding process (task 116, `smaqit.infrastructure-onboard-k3s-app`) nor the app-side deploy skill (task 117) gave an app project a declared way to *request* onboarding without also requiring the broad direct-commit-and-`workflow_dispatch` access an infra repo's own maintainers would have.
- **Deliberately more opinionated than task 116.** Task 116 is a thin dispatcher precisely because it lives inside an infra-owning repo it doesn't control the internals of — every such repo can differ, so smaqit cannot prescribe its mechanics. This skill instead lives inside the smaqit-managed **app** repo, which smaqit fully owns the conventions for, so it can and does declare one fixed, opinionated request contract (PR-to-a-registry-file, gate-on-merge) rather than deferring to unknown mechanics on its own side. Using `existing-k3s` with this skill implicitly requires the target infra repo's onboarding process to accept requests this way; an infra repo that doesn't work this way is simply not compatible with this skill as-is — a scoping boundary, not something this skill adapts around.
- No family precedent exists for this exact shape (a cross-repository PR request). Its gate mechanics instead mirror `smaqit.feature-new`'s already-proven deploy-PR pattern: create the PR, pause, re-check merge state on the next invocation — never a busy-poll loop.

## Steps

### Pre-conditions

- **The target platform repository's own onboarding process accepts requests via a PR-mergeable
  registry file.** This is smaqit's own declared convention for the app side, not a guarantee
  about any given infra repo. If the target repo's onboarding process works some other way (a
  ticket, a chat request, a direct API call, anything not triggered by merging a file change),
  this skill does not apply — that combination is out of scope, not a bug to work around.
- The Infrastructure spec declares, for the target environment: the platform repository
  (`owner/repo`, via the `## Constraints` table's `Platform Repo` row), the registry-file path
  within that repository, and the target machine-slug.
- `secret/apps/<app-slug>/platform-repo` is populated in Vault — a fine-grained PAT scoped only to
  `contents:write` and `pull_requests:write` on the platform repository, distinct from
  `secret/apps/<app-slug>/github` (which is scoped to this project's own repository). Populated via
  `smaqit.infrastructure-vault-loader`'s `existing-k3s` branch.
- The operator supplies the exact registry-entry content (text or diff) to add — this skill never
  infers or validates a registry-file schema, since the platform repository's actual format is not
  standardized or known to smaqit.
- The credential is a direct collaborator-level grant on the platform repository (`contents:write`)
  — never a fork-based flow. `gh pr create -H user:branch`'s fork-based cross-repo path has a known
  upstream limitation (`cli/cli#10093`, open) that this skill deliberately avoids entirely by
  pushing its branch straight to the platform repository and opening a same-repo PR via
  `gh pr create -R <platform-repo>`.

### Steps

1. **Resolve inputs** for the target environment (test or prod) from the Infrastructure spec:
   `platform_repo` (owner/repo), `registry_file_path`, `machine_slug`. The operator supplies
   `entry_content` directly (the exact text/diff to add at that path).
2. **Compute a deterministic branch name**: `onboard-<app-slug>-<machine_slug>`. Check for an
   already-open PR on that branch before creating anything new:
   ```bash
   gh pr list -R <platform_repo> --head "onboard-<app-slug>-<machine_slug>" --state open --json number,url
   ```
   If one exists, skip straight to Step 5 (gate) — never open a duplicate.
3. **Create or update the branch on the platform repository** (using the
   `secret/apps/<app-slug>/platform-repo` credential): clone or fetch the platform repo, create
   the branch from its default branch if it doesn't already exist, write `entry_content` at
   `registry_file_path`, commit, and push.
4. **Open the PR**:
   ```bash
   gh pr create -R <platform_repo> \
     --head "onboard-<app-slug>-<machine_slug>" \
     --title "Onboard <app-slug> (<machine_slug>)" \
     --body "Requests onboarding for app slug <app-slug> onto machine <machine_slug>. Opened by smaqit.infrastructure-request-k3s-onboarding."
   ```
   Report the PR URL. Never trigger `workflow_dispatch` on the platform repository — convergence is
   entirely the platform team's own concern once the PR merges.
5. **Gate on merge** — a single state check per invocation, never a busy-poll loop:
   ```bash
   gh pr view -R <platform_repo> <pr-number> --json state,mergedAt
   ```
   - `state: "MERGED"` → report success and stop. Onboarding has been requested and accepted; the
     platform team's own process now handles convergence and the eventual kubeconfig hand-off
     (still out-of-band, unchanged from `smaqit.infrastructure-deploy-k3s-app`'s existing
     precondition).
   - `state: "OPEN"` → report "pending — awaiting platform-team review" with the PR URL, and stop
     cleanly. Re-entrant: invoking this skill again later re-checks the same PR rather than
     re-creating it (Step 2's idempotency check).
   - `state: "CLOSED"` (not merged) → report the closure to the operator; do not silently reopen or
     retry. This requires human judgment about what changed.

## Output

Either: a newly opened PR on the platform repository (pending review), or confirmation that a
previously opened PR for this app/machine has since merged. Never a kubeconfig, a converged
Namespace, or any change to the platform repository outside the one PR.

## Scope

- Does NOT perform or know anything about the platform-side onboarding mechanism itself
  (Namespace/RBAC/PSA/NetworkPolicy/Quota/kubeconfig issuance, or whatever else a given platform
  repo does) — that is entirely that repository's own process, triggered however its own
  maintainers have set it up to react to this skill's PR merging. `smaqit.infrastructure-onboard-k3s-app`
  is one possible shape such a process could take on the infra side, but this skill assumes
  nothing about it beyond "a merged PR triggers something."
- Does NOT trigger `workflow_dispatch` on the platform repository, read or write anything else in
  it, or hold any credential broader than PR-create rights on that one repository.
- Does NOT fetch, store, or validate the resulting kubeconfig — that remains
  `smaqit.infrastructure-vault-loader`'s existing out-of-band manual-paste step.
- Does NOT assume or validate a registry-file schema — content is exactly what the operator
  supplies.
- Does NOT run as part of the generated `deploy.yml` — it is a one-time setup step per
  machine-slug (Phase 4 for test, Phase 5 for prod), not a per-push CI job.

## Completion

- [ ] Deterministic branch name computed; an existing open PR for this app/machine-slug is reused,
      never duplicated
- [ ] Branch pushed directly to the platform repository (never a fork; never `-H user:branch`)
- [ ] PR opened via `gh pr create -R <platform_repo>`, URL reported
- [ ] `workflow_dispatch` never triggered on the platform repository
- [ ] Merge state checked once per invocation (`gh pr view --json state,mergedAt`), never polled in
      a loop
- [ ] On merge: success reported, no attempt made to fetch the resulting kubeconfig
- [ ] On pending or closed: reported plainly to the operator, no silent retry

## Failure Handling

| Situation | Action |
|-----------|--------|
| `secret/apps/<app-slug>/platform-repo` absent | Stop. Point at `smaqit.infrastructure-vault-loader`'s `existing-k3s` branch to populate it. Never fabricate a placeholder. |
| `entry_content` not supplied by the operator | Stop and request it. Never guess a registry-file schema. |
| PR already open for this app/machine-slug | Skip creation; proceed directly to the merge-state gate. Never open a duplicate. |
| PR state is `CLOSED` (not merged) | Report to the operator. Do not reopen or retry automatically — requires human judgment. |
| PR state is `OPEN` | Report "pending" with the PR URL and stop cleanly. Re-invoke later to re-check — never busy-poll. |
| Credential lacks sufficient scope (`gh` reports a permission error on push or PR create) | Report the exact error. Do not attempt a fork-based fallback (`-H user:branch`) — that path has a known upstream limitation (`cli/cli#10093`) and was deliberately designed around. |
| Platform repository or registry file path doesn't exist | Stop and report; do not create the file's parent directories speculatively. |

## Gotchas

- **Never use the fork-based `gh pr create -H user:branch` flow.** `cli/cli#10093` (open, as of
  this skill's authoring) tracks a real limitation in that flow for cross-repo PRs. This skill
  avoids it entirely by using a directly-scoped collaborator credential to push a branch straight
  to the platform repository and opening a same-repo PR instead.
- **Gate is a single check, not a loop.** No busy-poll pattern exists anywhere else in this
  codebase for a PR-merge wait; `smaqit.feature-new`'s own deploy-PR gate creates, pauses, and
  re-checks only on the next invocation. Mirror that shape exactly — do not add a `sleep`/retry
  loop here.
- **Never trigger `workflow_dispatch` on the platform repository.** Even if the credential happened
  to have that scope, this skill's job ends at "PR merged" — triggering convergence directly would
  bypass the platform team's own review-then-converge process this skill exists to respect.
- **The registry-file format is genuinely unknown to smaqit.** Do not infer, template, or validate
  its shape — `entry_content` is always exactly what the operator supplies.
- **This is a one-time setup step, not a per-push CI job.** Do not add it to the generated
  `deploy.yml` — a PR-review gate on every push would stall production deploys on reviewer
  availability.

## Allowed Tools

Bash(git:*), Bash(gh:*), Bash(vault:*)

## Examples

**Input:** A project's Phase 4 dev sweep resolves `provisioning_mode: existing-k3s` and needs to
onboard its test environment onto machine-slug `test-cluster` before invoking the k3s deploy skill.

**Output:** No existing open PR found for `onboard-<app-slug>-test-cluster`. A branch is pushed to
the platform repository with the app's registry entry; `gh pr create -R <platform_repo>` opens PR
#42; the skill reports the PR URL and "pending — awaiting platform-team review," then stops.
Re-invoked the next day after the platform team merges PR #42, the skill reports success and Phase
4 proceeds to `smaqit.infrastructure-deploy-k3s-app`.
