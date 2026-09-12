---
name: smaqit.infrastructure-request-k3s-onboarding
description: Use when an app project targeting `provisioning_mode: existing-k3s` needs to request onboarding of one environment onto a platform-owned, self-hosted k3s cluster whose own onboarding process accepts requests via a PR-mergeable registry file. Given a declared platform repository, a registry-file path, and operator-supplied entry content, opens a PR against that external repository and gates on merge before reporting success — never requiring direct commit/`workflow_dispatch` access to a repo it doesn't own. Used in Phase 4 (test) and Phase 5 (prod) of `smaqit.new-greenfield-project`, before `smaqit.infrastructure-deploy-k3s-app`. A smaqit-declared, opinionated app-side convention, not dependent on how any given platform repo actually fulfills the request — never performs or assumes anything about the platform-side convergence mechanism itself (Namespace/RBAC/kubeconfig issuance, or anything else); that remains entirely the platform team's own process, triggered however their own repo reacts to the merge.
metadata:
  version: "1.0.0"
---

# Request k3s App Onboarding

## Pre-conditions

- **The target platform repository's own onboarding process accepts requests via a PR-mergeable
  registry file.** This is smaqit's own declared convention for the app side, not a guarantee
  about any given infra repo. If the target repo's onboarding process works some other way, this
  skill does not apply — that combination is out of scope, not a bug to work around.
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
3. **Fetch the current registry file, then append — never a blind overwrite.** Using the
   `platform-repo` credential: create `$BRANCH` from the platform repository's default branch if it
   doesn't already exist, then read the *current* content of `registry_file_path` on that branch
   (e.g. `gh api repos/"$PLATFORM_REPO"/contents/"$REGISTRY_FILE_PATH" --jq '.content' | base64 -d`,
   or an equivalent fetched-clone read). Append `entry_content` to that existing content — every
   other already-onboarded entry and any header/comment lines must survive unchanged — then write
   the merged result back to `registry_file_path`, commit, and push. If the fetch reports the path
   doesn't exist at all, treat it as the existing "registry file path doesn't exist" Failure
   Handling case below; do not invent initial file content.
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

- Does NOT perform or know anything about the platform-side onboarding mechanism itself — that is
  entirely that repository's own process, triggered however its own maintainers have set it up to
  react to this skill's PR merging. This skill assumes nothing about that process beyond "a merged
  PR triggers something."
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

- **Deliberately opinionated, unlike the platform-side onboarding process it targets.** The
  platform repo's own onboarding mechanism lives inside an infra repo smaqit doesn't control the
  internals of — every such repo can differ, and smaqit ships no skill for it (an earlier generic
  dispatcher stub, `smaqit.infrastructure-onboard-k3s-app`, was removed as an orphaned, unused
  duplicate of this Pre-conditions/Gotchas guidance — see task 119). This skill lives inside the
  smaqit-managed app repo instead, which smaqit fully owns the conventions for, so it declares one
  fixed request contract (PR-to-a-registry-file, gate-on-merge) rather than deferring to unknown
  mechanics on its own side. Do not water this down to match the infra side's genericness — the two
  concerns sit on opposite sides of a repo boundary for a reason.
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
- [ ] Registry file's current content fetched before writing; every pre-existing entry and
      header/comment line survives unchanged in the merged result — never a full-file replace
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
