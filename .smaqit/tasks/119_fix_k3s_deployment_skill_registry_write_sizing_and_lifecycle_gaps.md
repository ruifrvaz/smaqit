---
status: Not Started
created: "2026-09-12"
---

# Fix Registry-Write Semantics, Deployment Sizing, and Lifecycle Gaps in the k3s App-Deployment Skills

## Description

Follow-up review of tasks 116/117/118 (the k3s onboarding/deployment skill family), performed
from the Magnificah `infrastructure` repo's side, cross-checking the shipped artifacts against
that repo's actual live RBAC allowlist, kubeconfig-context shape, ResourceQuota/LimitRange values,
and NetworkPolicy contract (`specs/infrastructure/k3s-app-onboarding.md` there, all 24 criteria
live-verified). Four concrete defects found in already-landed artifacts; none are design-level.

### Finding 1 — `smaqit.infrastructure-request-k3s-onboarding` registry write is a full overwrite, not an append/merge

Step 3 (per the skill's own wording) writes `entry_content` at `registry_file_path`. Read
literally, this replaces the platform repo's entire registry file, dropping every other
already-onboarded slug and any header comment. On a live registry this would silently deregister
every previously onboarded app in the same PR.

The platform-repo-side `confirm_offboard` rail and the codeowner's PR review make this
recoverable in practice (the diff would show mass deletions), but the skill should not depend on
manual review to catch a class of bug the PR-time dry-run validation doesn't check for either.

**Fix**: the skill must read the current `registry_file_path` content first (via the platform
repo's contents API or a fetched clone), append/merge the new entry into the existing structure,
and only then write and PR the merged result. Never write `entry_content` as a full-file replace.

### Finding 2 — `deployment.yaml.template` default sizing does not fit a standard onboarded quota

The template ships `replicas: 2` with per-container requests of `cpu: 100m` / `memory: 128Mi`.
Two replicas at that request size alone consume 256Mi of memory requests — exactly a full per-app
`ResourceQuota` ceiling live-verified in a downstream infrastructure repo's onboarding contract.
Any transient pod sharing the namespace's quota (a `RollingUpdate` surge pod, or cert-manager's
HTTP-01 solver pod, which schedules inside the app's own namespace and picks up the namespace's
`LimitRange` default request) pushes the namespace over quota. Observed effect: the Certificate
never reaches `Ready` (solver pod stuck `Pending`, quota-rejected) and/or a rolling update stalls
because the surge pod can't schedule.

**Fix**: default the template to `replicas: 1` and smaller per-container requests (a
conservative, proven-live default such as `cpu: 50m` / `memory: 32Mi`), expressed as overridable
template tokens rather than hardcoded numbers — this is a default-safety fix, not a
platform-specific constant. Optionally extend `manifest-lint.sh` to warn when
`replicas × (requests + a configurable surge/solver headroom)` would exceed a declared quota
ceiling, if quota values are available to the linter at lint time.

### Finding 3 — `smaqit.task-start` / `smaqit.task-complete` push lifecycle metadata straight to `main`

Neither skill detects branch protection on the target repo (a required-PR-review rule). Against
such a repo every lifecycle step (`chore: create task N`, `chore: start task N`,
`chore: complete task N`) fails to push directly and has to be manually turned into its own chore
PR. In a downstream infrastructure repo this manual pattern directly caused two missed
changelog-promotion steps and a near-lost commit (recovered from reflog).

**Fix**: both skills should detect a protected default branch (a failed direct push, or a
pre-check via the repo's branch-protection/ruleset API) and fall back to opening a small chore PR
automatically, rather than surfacing a raw git failure for the operator to notice and route by
hand.

### Finding 4 — `smaqit.infrastructure-onboard-k3s-app` is an orphaned stub

Nothing invokes it — no caller found anywhere in the skill tree. It exists only as prose
referenced by `smaqit.infrastructure-request-k3s-onboarding`'s own documentation. Either fold its
content into that skill's Gotchas/Pre-conditions section, or remove it, so there is no dangling
skill directory that looks load-bearing but isn't.

## Acceptance Criteria

- `smaqit.infrastructure-request-k3s-onboarding` reads-merges-writes the registry file; a live or
  simulated test confirms an existing unrelated entry survives a new-app onboarding PR.
- `deployment.yaml.template` defaults to `replicas: 1` and smaller request values, expressed as
  overridable tokens rather than hardcoded numbers.
- `smaqit.task-start` and `smaqit.task-complete` fall back to a chore PR when a direct push to the
  default branch is rejected by branch protection, verified against a protected-branch repo.
- `smaqit.infrastructure-onboard-k3s-app` is either merged into `request-k3s-onboarding` or
  removed; no orphaned skill directory remains.

## Out of Scope

- Anything specific to a consuming infrastructure repo's own playbook or Vault/credential
  provisioning steps for the k3s deployment flow — that stays in that repo's own task tracker.
- Re-litigating the k3s app-deployment architecture itself (tasks 116/117/118) — this task is
  fix-only against already-landed artifacts.
