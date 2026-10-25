# smaqit.infrastructure-request-k3s-onboarding

## Frontmatter

```yaml
name: smaqit.infrastructure-request-k3s-onboarding
description: Use when an existing-k3s app needs one environment onboarded onto a platform-owned k3s cluster whose onboarding accepts PR-mergeable registry-file requests (project-new Phases 4 and 5, before deploy-k3s-app). Opens a PR against the platform repository and gates on its merge; never needs direct commit or workflow_dispatch access, and never assumes how the platform converges.
metadata:
  version: "3.0.0"
  smaqit-slots: "platform-onboarding=request"
```

## Provenance

- `synthesized: false`
- `authored-for-project: smaqit itself`
- `authored-date: 2026-09-12`
- Authored directly into canonical `smaqit` as a child of task 117 (`smaqit.infrastructure-deploy-k3s-app`), closing a gap identified during that task's post-implementation review: neither an infra repo's own onboarding process (task 116, `smaqit.infrastructure-onboard-k3s-app`) nor the app-side deploy skill (task 117) gave an app project a declared way to *request* onboarding without also requiring the broad direct-commit-and-`workflow_dispatch` access an infra repo's own maintainers would have.
- **Deliberately more opinionated than task 116.** Task 116 is a thin dispatcher precisely because it lives inside an infra-owning repo it doesn't control the internals of — every such repo can differ, so smaqit cannot prescribe its mechanics. This skill instead lives inside the smaqit-managed **app** repo, which smaqit fully owns the conventions for, so it can and does declare one fixed, opinionated request contract (PR-to-a-registry-file, gate-on-merge) rather than deferring to unknown mechanics on its own side. Using `existing-k3s` with this skill implicitly requires the target infra repo's onboarding process to accept requests this way; an infra repo that doesn't work this way is simply not compatible with this skill as-is — a scoping boundary, not something this skill adapts around.
- No family precedent exists for this exact shape (a cross-repository PR request). Its gate mechanics instead mirror `smaqit.feature-new`'s already-proven deploy-PR pattern: create the PR, pause, re-check merge state on the next invocation — never a busy-poll loop.
- Resynced with the shipped SKILL.md in task 145 (2026-10-25); the body below is applied to SKILL.md by hand.
- Task 137 (2026-10-25): the platform-repo PAT is retired; the PR opens with the operator's own `gh` login after a push-access preflight (3.0.0).

## Skill Body

Applied to `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` by hand, verbatim (task 145: `smaqit.L2`'s skill
rules conflict with this repo's description, metadata and section conventions).

<!-- skill-body:begin -->
# Request k3s App Onboarding

## Pre-conditions

- **The target platform repository's own onboarding process accepts requests via a PR-mergeable
  registry file.** This is smaqit's own declared convention for the app side, not a guarantee
  about any given infra repo. If the target repo's onboarding process works some other way, this
  skill does not apply — that combination is out of scope, not a bug to work around.
- The Infrastructure spec declares the platform repository (`owner/repo`, the `Platform Repo`
  Constraints row) and the registry-file path within it (`Registry File Path` row), and the target
  environment's `## Environments` row names its `Machine Slug` — all sourced from the platform
  team's own onboarding documentation, never invented locally.
- The operator's own `gh` login has push access to the platform repository (a collaborator write
  grant; Step 1 checks it). The app holds no credential for onboarding, and this skill never loads
  one from Vault.
- The operator supplies the exact registry-entry content (text or diff) to add — this skill never
  infers or validates a registry-file schema. A platform that binds the app's CI identity to its
  repository usually asks for that repository in the entry (Step 1 prints it).
- That access is a direct collaborator-level grant on the platform repository, not a fork-based
  flow — never use `gh pr create -H user:branch` (see Gotchas).

## Steps

1. **Resolve inputs** for the target environment from the Infrastructure spec: `platform_repo`
   (owner/repo), `registry_file_path`, `machine_slug`. Every `gh`/`git` call in this skill runs
   with the operator's own `gh` login. Print this repository's identity; `gh repo view --json id`
   would print a GraphQL node ID, not the numeric ID a platform binds:
   ```bash
   gh api "repos/$(gh repo view --json nameWithOwner --jq .nameWithOwner)" --jq '"\(.full_name) \(.id)"'
   ```
   Then confirm the login can push to the platform repository before writing anything:
   ```bash
   [ "$(gh api "repos/${PLATFORM_REPO}" --jq .permissions.push)" = "true" ]
   ```
   Anything but `true` stops the skill (see Failure Handling). The operator supplies
   `entry_content` directly; when the platform's registry asks for the requesting repository, the
   entry carries the values printed above.
2. **Check for an already-open PR before creating anything:**
   ```bash
   BRANCH="onboard-${APP_SLUG}-${MACHINE_SLUG}"
   gh pr list -R "$PLATFORM_REPO" --head "$BRANCH" --state open --json number,url
   ```
   If one exists, skip to Step 5 — never open a duplicate.
3. **Fetch the current registry file, then append — never a blind overwrite.** Create `$BRANCH` from the platform repository's default branch if it
   doesn't already exist, then read the *current* content of `registry_file_path` on that branch
   (e.g. `gh api repos/"$PLATFORM_REPO"/contents/"$REGISTRY_FILE_PATH" --jq '.content' | base64 -d`,
   or an equivalent fetched-clone read). If `entry_content` names a repository, confirm it is exactly
   the one Step 1 printed (name and numeric ID): a wrong ID binds another repository's deploy
   identity to this app's Namespace. Append `entry_content` to that existing content — every
   other already-onboarded entry and any header/comment lines must survive unchanged — then write
   the merged result back to `registry_file_path`, commit, and push. If the fetch reports the path
   doesn't exist at all, treat it as the existing "registry file path doesn't exist" Failure
   Handling case below; do not invent initial file content.
4. **Open the PR:**
   ```bash
   gh pr create -R "$PLATFORM_REPO" \
     --head "$BRANCH" \
     --title "chore: onboard ${APP_SLUG} (${MACHINE_SLUG})" \
     --body "Requests onboarding for app slug ${APP_SLUG} onto machine ${MACHINE_SLUG}. Opened by smaqit.infrastructure-request-k3s-onboarding."
   ```
   Report the PR URL. Never trigger `workflow_dispatch` on the platform repository.
5. **Gate on merge — a single state check, never a loop:**
   ```bash
   gh pr view -R "$PLATFORM_REPO" "$PR_NUMBER" --json state,mergedAt
   ```
   - `MERGED` → report the request merged. The app becomes deployable once the platform has
     reconciled the merge — automatically on many platforms (their onboarding documentation says
     so), otherwise after the platform's own step, which this skill never performs. There is no
     credential to collect: the deploy connects with its own OIDC identity, and its namespace
     guard is the convergence check — a Forbidden answer means the binding has not reconciled yet.
   - `OPEN` → report "pending — awaiting platform-team review" with the PR URL and stop cleanly.
     Re-invoke later to re-check (Step 2 finds and reuses the same PR).
   - `CLOSED` (not merged) → report the closure; do not silently reopen or retry.

## Output

Either a newly opened PR on the platform repository (pending review), or confirmation that a
previously opened PR for this app/machine has since merged, with the reminder that the platform
must reconcile it before the first deploy succeeds. Never a credential, a converged Namespace, or
any change to the platform repository outside the one PR.

## Scope

- Does NOT perform or know anything about the platform-side onboarding mechanism itself — that is
  entirely that repository's own process, triggered however its own maintainers have set it up to
  react to this skill's PR merging. This skill assumes nothing about that process beyond "a merged
  PR triggers something."
- Does NOT trigger `workflow_dispatch` on the platform repository, or read or write anything else in
  it. It holds no credential of its own: the operator's `gh` login opens this one PR and is used for
  nothing else.
- Does NOT handle any deploy credential — the platform binds the app's CI identity (its repository
  and machine-slug GitHub Environment) to the Namespace, so there is nothing to collect.
- Does NOT assume or validate a registry-file schema.
- Does NOT run inside the generated `deploy.yml` — it is a one-time setup step per machine-slug
  (Phase 4 for test, Phase 5 for prod), never a per-push CI job.

## Examples

**Input:** A project's Phase 4 dev sweep resolves `provisioning_mode: existing-k3s` and needs to
onboard its test environment onto machine-slug `test-cluster`.

**Output:** No existing open PR found for `onboard-<app-slug>-test-cluster`. A branch is pushed to
the platform repository with the registry entry; `gh pr create -R <platform_repo>` opens PR #42;
the skill reports the PR URL and "pending — awaiting platform-team review," then stops. Re-invoked
after the platform team merges PR #42, the skill reports the request merged; once the platform has
reconciled it, Phase 4's dispatched `deploy.yml` connects with the job's OIDC identity and deploys.

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
  real limitation in that flow for cross-repo PRs. This skill avoids it entirely: the operator's
  collaborator access pushes a branch straight to the platform repository and opens a same-repo PR
  instead.
- **Gate is a single check, not a loop.** No busy-poll pattern exists anywhere else in this
  codebase for a PR-merge wait; `smaqit.feature-new`'s own deploy-PR gate creates, pauses, and
  re-checks only on the next invocation. Mirror that shape exactly.
- **Never trigger `workflow_dispatch` on the platform repository**, even when the operator's login
  has that permission — this skill's job ends at "PR merged." Merged does not mean reconciled: report
  it plainly, and let the first deploy's namespace guard confirm the binding (Forbidden until then).
- **PR conventions:** branch `onboard-<app-slug>-<machine-slug>` and title
  `chore: onboard <app-slug> (<machine-slug>)` — a deterministic name the platform side can
  recognise, and the reason Step 2 can find and reuse an earlier request.
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
- [ ] Repository identity printed and push access to the platform repository confirmed before any write; any repository named in `entry_content` matches it
- [ ] On merge: request reported merged, deployable once the platform reconciles it; no credential collected
- [ ] On pending or closed: reported plainly, no silent retry

## Failure Handling

| Situation | Action |
|-----------|--------|
| Infrastructure spec doesn't declare `Platform Repo`, `Registry File Path`, or the environment's `Machine Slug` | Stop. Name exactly which field(s) are missing and point at the platform repository's own onboarding documentation as the only valid source. Never fabricate a value or guess a path. |
| The operator's `gh` login lacks push access (`permissions.push` is not `true`, or the repository is not visible) | Stop. Ask the platform team for collaborator write access on the platform repository. Never fall back to a fork, `-H user:branch`, or a stored token. |
| `entry_content` not supplied by the operator | Stop and request it. Never guess a registry-file schema. |
| PR already open for this app/machine-slug | Skip creation; proceed directly to the merge-state gate. Never open a duplicate. |
| PR state is `CLOSED` (not merged) | Report to the operator. Do not reopen or retry automatically. |
| PR state is `OPEN` | Report "pending" with the PR URL and stop cleanly. Re-invoke later — never busy-poll. |
| A push or `gh pr create` is refused despite the preflight | Report the exact error. Do not fall back to `-H user:branch` (`cli/cli#10093`). |
| `entry_content` names a different repository or ID than Step 1 printed | Stop before writing; ask the operator to correct the entry. |
| Platform repository or registry file path doesn't exist | Stop and report; do not speculatively create parent directories. |

## Allowed Tools

Bash(git:*), Bash(gh:*), Bash(bash:*)
<!-- skill-body:end -->
