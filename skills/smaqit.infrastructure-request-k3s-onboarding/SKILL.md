---
name: smaqit.infrastructure-request-k3s-onboarding
description: Use when an app project targeting `provisioning_mode: existing-k3s` needs to request onboarding of one environment onto a platform-owned, self-hosted k3s cluster. Given a declared platform repository, a registry-file path, and operator-supplied entry content, opens a PR against that external repository and gates on merge before reporting success — never requiring direct commit/`workflow_dispatch` access to a repo it doesn't own. Used in Phase 4 (test) and Phase 5 (prod) of `smaqit.new-greenfield-project`, before `smaqit.infrastructure-deploy-k3s-app`. Never performs the platform-side convergence itself (Namespace/RBAC/kubeconfig issuance) — that remains the platform team's own process, triggered by their own repo reacting to the merge.
metadata:
  version: "1.0.0"
---

# Request k3s App Onboarding

## Pre-conditions

- The Infrastructure spec declares, for the target environment: the platform repository
  (`owner/repo`, the `## Constraints` table's `Platform Repo` row), the registry-file path within
  that repository, and the target machine-slug.
- `secret/apps/<app-slug>/platform-repo` is populated (`smaqit.infrastructure-vault-loader`'s
  `existing-k3s` branch) — a fine-grained PAT scoped only to `contents:write` and
  `pull_requests:write` on the platform repository. Distinct from `secret/apps/<app-slug>/github`,
  which is scoped to this project's own repository.
- The operator supplies the exact registry-entry content (text or diff) to add — this skill never
  infers or validates a registry-file schema.
- The credential is a direct collaborator-level grant on the platform repository, not a fork-based
  flow — never use `gh pr create -H user:branch` (see Gotchas).

## Steps

1. **Resolve inputs** for the target environment from the Infrastructure spec: `platform_repo`
   (owner/repo), `registry_file_path`, `machine_slug`. The operator supplies `entry_content`
   directly.
2. **Check for an already-open PR before creating anything:**
   ```bash
   BRANCH="onboard-${APP_SLUG}-${MACHINE_SLUG}"
   gh pr list -R "$PLATFORM_REPO" --head "$BRANCH" --state open --json number,url
   ```
   If one exists, skip to Step 5 — never open a duplicate.
3. **Push the registry entry to a branch on the platform repository** (using the
   `platform-repo` credential): fetch the platform repo, create `$BRANCH` from its default branch
   if it doesn't already exist, write `entry_content` at `registry_file_path`, commit, push.
4. **Open the PR:**
   ```bash
   gh pr create -R "$PLATFORM_REPO" \
     --head "$BRANCH" \
     --title "Onboard ${APP_SLUG} (${MACHINE_SLUG})" \
     --body "Requests onboarding for app slug ${APP_SLUG} onto machine ${MACHINE_SLUG}. Opened by smaqit.infrastructure-request-k3s-onboarding."
   ```
   Report the PR URL. Never trigger `workflow_dispatch` on the platform repository.
5. **Gate on merge — a single state check, never a loop:**
   ```bash
   gh pr view -R "$PLATFORM_REPO" "$PR_NUMBER" --json state,mergedAt
   ```
   - `MERGED` → report success and stop. The platform team's own process now handles convergence
     and the kubeconfig hand-off (still out-of-band, unchanged).
   - `OPEN` → report "pending — awaiting platform-team review" with the PR URL and stop cleanly.
     Re-invoke later to re-check (Step 2 finds and reuses the same PR).
   - `CLOSED` (not merged) → report the closure; do not silently reopen or retry.

## Output

Either a newly opened PR on the platform repository (pending review), or confirmation that a
previously opened PR for this app/machine has since merged. Never a kubeconfig, a converged
Namespace, or any change to the platform repository outside the one PR.

## Scope

- Does NOT perform the platform-side onboarding mechanism itself — that is the platform
  repository's own process (documented for reference in `smaqit.infrastructure-onboard-k3s-app`),
  triggered by the platform team after this skill's PR merges.
- Does NOT trigger `workflow_dispatch` on the platform repository, read or write anything else in
  it, or hold any credential broader than PR-create rights on that one repository.
- Does NOT fetch, store, or validate the resulting kubeconfig — that remains
  `smaqit.infrastructure-vault-loader`'s existing out-of-band manual-paste step.
- Does NOT assume or validate a registry-file schema.
- Does NOT run inside the generated `deploy.yml` — it is a one-time setup step per machine-slug
  (Phase 4 for test, Phase 5 for prod), never a per-push CI job.

## Examples

**Input:** A project's Phase 4 dev sweep resolves `provisioning_mode: existing-k3s` and needs to
onboard its test environment onto machine-slug `test-cluster`.

**Output:** No existing open PR found for `onboard-<app-slug>-test-cluster`. A branch is pushed to
the platform repository with the registry entry; `gh pr create -R <platform_repo>` opens PR #42;
the skill reports the PR URL and "pending — awaiting platform-team review," then stops. Re-invoked
after the platform team merges PR #42, the skill reports success and Phase 4 proceeds to
`smaqit.infrastructure-deploy-k3s-app`.

## Gotchas

- **Never use the fork-based `gh pr create -H user:branch` flow.** `cli/cli#10093` (open) tracks a
  real limitation in that flow for cross-repo PRs. This skill avoids it entirely by using a
  directly-scoped collaborator credential to push a branch straight to the platform repository and
  opening a same-repo PR instead.
- **Gate is a single check, not a loop.** No busy-poll pattern exists anywhere else in this
  codebase for a PR-merge wait; `smaqit.feature-new`'s own deploy-PR gate creates, pauses, and
  re-checks only on the next invocation. Mirror that shape exactly.
- **Never trigger `workflow_dispatch` on the platform repository**, even if the credential happened
  to have that scope — this skill's job ends at "PR merged."
- **The registry-file format is genuinely unknown to smaqit.** `entry_content` is always exactly
  what the operator supplies — never inferred, templated, or validated.
- **One-time setup step, not per-push CI.** Do not add this to the generated `deploy.yml`.

## Completion

- [ ] Deterministic branch name computed; an existing open PR for this app/machine-slug is reused,
      never duplicated
- [ ] Branch pushed directly to the platform repository (never a fork; never `-H user:branch`)
- [ ] PR opened via `gh pr create -R <platform_repo>`, URL reported
- [ ] `workflow_dispatch` never triggered on the platform repository
- [ ] Merge state checked once per invocation, never polled in a loop
- [ ] On merge: success reported, no attempt made to fetch the resulting kubeconfig
- [ ] On pending or closed: reported plainly, no silent retry

## Failure Handling

| Situation | Action |
|-----------|--------|
| `secret/apps/<app-slug>/platform-repo` absent | Stop. Point at `smaqit.infrastructure-vault-loader`'s `existing-k3s` branch to populate it. Never fabricate a placeholder. |
| `entry_content` not supplied by the operator | Stop and request it. Never guess a registry-file schema. |
| PR already open for this app/machine-slug | Skip creation; proceed directly to the merge-state gate. Never open a duplicate. |
| PR state is `CLOSED` (not merged) | Report to the operator. Do not reopen or retry automatically. |
| PR state is `OPEN` | Report "pending" with the PR URL and stop cleanly. Re-invoke later — never busy-poll. |
| Credential lacks sufficient scope | Report the exact error. Do not fall back to `-H user:branch` (`cli/cli#10093`). |
| Platform repository or registry file path doesn't exist | Stop and report; do not speculatively create parent directories. |

## Allowed Tools

Bash(git:*), Bash(gh:*), Bash(vault:*)
