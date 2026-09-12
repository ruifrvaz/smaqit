---
status: PR Open
created: "2026-09-12"
mode: Assisted
started: "2026-09-12"
pr: 89
---

# Fix Registry-Write Semantics, Deployment Sizing, and Lifecycle Gaps in the k3s App-Deployment Skills

## Description

Follow-up review of tasks 116/117/118 (the k3s onboarding/deployment skill family), performed
from a downstream infrastructure repo's side, cross-checking the shipped artifacts against
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

### Finding 3 — `smaqit.task-start` / `smaqit.task-complete` push lifecycle metadata straight to `main` (dropped — see Out of Scope)

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

## Design Decisions

- **Finding 1 fix shape:** fetch the platform repo's current `registry_file_path` content first,
  append `entry_content` to that existing content, then commit — never a blind replace. The
  registry-file format stays genuinely unknown to smaqit (existing Gotchas note in
  `smaqit.infrastructure-request-k3s-onboarding/SKILL.md`); this is a positional append relative to
  fetched content, not schema-aware structural merging.
- **Finding 2 fix shape:** expose `replicas`, CPU request, and memory request as overridable
  template tokens (`__REPLICAS__`, `__CPU_REQUEST__`, `__MEM_REQUEST__`) defaulting to `1` /
  `50m` / `32Mi`, following the file's existing `__APP_SLUG__`/`__IMAGE__` token convention. The
  optional `manifest-lint.sh` quota-aware warning mentioned in the Finding is explicitly **not**
  built — the linter has no quota-ceiling input anywhere in its current interface, and adding one
  is a new mechanism beyond what any Acceptance Criterion requires.
- **Finding 4 resolved as removal, not fold:** `smaqit.infrastructure-onboard-k3s-app`'s content
  (defer to the infra repo's own onboarding process) is already fully covered by
  `smaqit.infrastructure-request-k3s-onboarding`'s own Pre-conditions ("if the target repo's
  onboarding process works some other way, this skill does not apply") and Gotchas ("Deliberately
  more opinionated than `smaqit.infrastructure-onboard-k3s-app`..."). Delete the directory outright
  rather than folding text that would duplicate what's already there.
- **Finding 3 dropped from this task's scope entirely**, per explicit user direction —
  `smaqit.task-start`/`smaqit.task-complete` are owned by the sibling `smaqit-extensions` repo, not
  this repo's own `skills/` tree (confirmed: no such directories exist here). Not rescoped to a
  companion task in this session.

## Implementation Steps

1. `smaqit.infrastructure-request-k3s-onboarding/SKILL.md` Step 3 — replace "write `entry_content`
   at `registry_file_path`" with a fetch-current-content-then-append flow; add an "existing
   registry content preserved" line to the Completion checklist.
2. `smaqit.infrastructure-deploy-k3s-app/assets/deployment.yaml.template` — tokenize
   `replicas`/`cpu` request/`memory` request as described in Design Decisions; update the file's
   top-of-file token-list comment. Update `smaqit.infrastructure-deploy-k3s-app/SKILL.md` wherever
   it documents template tokens to match.
3. Delete the `skills/smaqit.infrastructure-onboard-k3s-app/` directory.
4. Bump both hardcoded skill-count assertions in `installer/main_test.go`
   (`TestRemoveEmbeddedSkillDirsPreservesUnownedSharedContent`,
   `TestSharedSkillsServeCopilotAndCodex`) from 30 to 29. Update the onboarding-family Q&A entry in
   `.smaqit/compendium.md` to drop the removed skill as a distinct numbered concern.
5. Add a `CHANGELOG.md` `[Unreleased]` entry (Fixed: registry read-append, safer default sizing;
   Removed: orphaned onboarding-dispatcher skill). Run `go vet`/`go test` and
   `make -C installer prepare` to confirm the count assertions pass and staged output regenerates
   cleanly.

## Known Issues Triage
**Triaged:** 2026-09-12
**Tools searched:** GitHub CLI (`cli/cli`), Kubernetes (`kubernetes/kubernetes`)
**Result:** Historical

### Historical (Closed)
- [#11739 `gh gist edit` truncates large gist file and risks data loss](https://github.com/cli/cli/issues/11739) — `cli/cli` — closed 2025-10-31 — a different command/object (gist edit, not repo-contents write) but the same destructive-overwrite failure class this task's Finding 1 fix is designed to prevent; corroborating context, not a live regression against this task's actual code path.

### Omitted Tools
- GitHub REST API (repository contents) — not independently searched; `cli/cli` (the CLI wrapping it) was searched instead, since the task uses `gh`/`git`, not a direct raw API client.

### Search Warnings
- Research-map task block for 119 has no context fingerprint (legacy task-file format, no `## Issue Triage Context` section — same limitation documented in tasks 107/109/111/112); `task-map.sh select` could not be used, so tool resolution and search terms were derived directly from the legacy Description/Notes signal instead of a verified fingerprinted block.

No open bug/regression issues confirmed both platform and feature dimensions for either repository — `kubernetes/kubernetes` open-issue matches (Kubemark scalability, PV multi-tenancy, node-shutdown endpoints) and `cli/cli` open-issue matches (`pr create` branch-name mismatch, `gh run download` artifact scoping, input-preservation enhancement) are all unrelated to registry-file overwrite semantics or ResourceQuota/LimitRange sizing.

## Findings

**Implementation approach:**
- Finding 1: rewrote `smaqit.infrastructure-request-k3s-onboarding/SKILL.md` Step 3 to fetch the
  platform repo's current `registry_file_path` content on the request branch, then append
  `entry_content` to it rather than writing it as a full-file replace. Verified with a local git
  simulation — a bare repo seeded with one existing entry, then the fetch-append-write flow run
  against it — confirming the pushed branch's file contains both the pre-existing and new entries.
- Finding 2: tokenized `deployment.yaml.template`'s `replicas`/CPU request/memory request as
  `__REPLICAS__`/`__CPU_REQUEST__`/`__MEM_REQUEST__`, defaulting to `1`/`50m`/`32Mi`; extended the
  file's own token-list comment. Verified by rendering the template with sample substituted values
  and running it through the existing `manifest-lint.sh` — valid YAML, lint passes cleanly.
- Finding 4: confirmed zero callers of `smaqit.infrastructure-onboard-k3s-app` anywhere in the
  skill tree, then deleted the directory. Cleaned up two now-stale mentions of it in
  `smaqit.infrastructure-request-k3s-onboarding/SKILL.md`'s Scope/Gotchas sections that referenced
  it as if it still existed.
- Bumped both hardcoded skill-count assertions in `installer/main_test.go` 30 → 29; updated the
  onboarding-family Q&A entry in `.smaqit/compendium.md`; added `CHANGELOG.md` `[Unreleased]`
  entries. `go vet ./...` and `go test ./...` pass; `make -C installer prepare` regenerates staging
  cleanly at the new count.

**Decisions made:**
- Finding 4 resolved as removal, not fold, per the plan's Design Decisions — nothing in the
  orphaned stub was missing from `request-k3s-onboarding`'s own Pre-conditions/Gotchas.
- Implementation Step 2 also called for updating `smaqit.infrastructure-deploy-k3s-app/SKILL.md`'s
  token documentation — on inspection it never hardcoded the old `replicas: 2`/`100m`/`128Mi`
  values or enumerated template tokens itself (only the template file's own header comment does),
  so no change was needed there beyond what Finding 2's template edit already covers.
- `manifest-lint.sh`'s optional quota-aware warning stayed out of scope per the plan — no
  quota-ceiling input exists anywhere in its interface today.

**Blockers encountered:**
- `smaqit.task-start`'s research-map verification step (`task-context.sh --allow-legacy`) requires
  a `## Notes` section for a legacy-format task file; task 119 didn't have one. Added a minimal
  Notes section to unblock it — the same class of legacy-format gap already documented for other
  tasks (107/109/111/112).
- The task file's Description named a real downstream project by name ("Magnificah"), violating
  this repo's own `CONTRIBUTING.md` rule against naming consumer projects in task files. Corrected
  to generic phrasing before implementation began.
- No live platform repo or GitHub credential was available to test the registry-write fix against
  a real PR; verified instead via a local git simulation (see Implementation approach). Same class
  of environment limitation already recorded as Follow-up in tasks 117/118.

**Follow-up identified:**
- Live verification against a real platform repo (an actual PR open/merge cycle, a real registry
  file) remains outstanding — same environment limitation already recorded as Follow-up in tasks
  117/118, not newly introduced by this task.
- Finding 3 (branch-protection fallback for `smaqit.task-start`/`smaqit.task-complete`) was dropped
  entirely from this task's scope. A companion task in the sibling `smaqit-extensions` repo would
  need to be filed separately if that gap should still be fixed.

## Acceptance Criteria

- [x] `smaqit.infrastructure-request-k3s-onboarding` reads-merges-writes the registry file; a live or
  simulated test confirms an existing unrelated entry survives a new-app onboarding PR.
- [x] `deployment.yaml.template` defaults to `replicas: 1` and smaller request values, expressed as
  overridable tokens rather than hardcoded numbers.
- [x] `smaqit.infrastructure-onboard-k3s-app` is either merged into `request-k3s-onboarding` or
  removed; no orphaned skill directory remains.

## Out of Scope

- Anything specific to a consuming infrastructure repo's own playbook or Vault/credential
  provisioning steps for the k3s deployment flow — that stays in that repo's own task tracker.
- Re-litigating the k3s app-deployment architecture itself (tasks 116/117/118) — this task is
  fix-only against already-landed artifacts.
- Finding 3 (`smaqit.task-start`/`smaqit.task-complete` branch-protection fallback) — those skills
  are owned by the sibling `smaqit-extensions` repo, not this repo's own `skills/` tree; this task
  cannot act on them here. Dropped per explicit user direction, not rescoped to a companion task
  in this session.
- The optional `manifest-lint.sh` quota-aware warning suggested alongside Finding 2 — no
  quota-ceiling input exists anywhere in the linter's current interface; building one is a new
  mechanism, not a hard requirement of any Acceptance Criterion.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` | Modify — Step 3 fetch-then-append rewrite, Completion checklist |
| `skills/smaqit.infrastructure-deploy-k3s-app/assets/deployment.yaml.template` | Modify — tokenize replicas/cpu/memory |
| `skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md` | Modify — update token documentation |
| `skills/smaqit.infrastructure-onboard-k3s-app/` | Delete — orphaned, superseded |
| `installer/main_test.go` | Modify — skill count 30 → 29 (both assertions) |
| `.smaqit/compendium.md` | Modify — onboarding-family Q&A entry |
| `CHANGELOG.md` | Modify — `[Unreleased]` entry |

## Notes

Filed as a live-verification follow-up from a downstream infrastructure repo's own review of
tasks 116/117/118's shipped artifacts (see Description for the review's scope). Findings 1, 2, and
4 are self-contained fixes against artifacts this repo owns. Finding 3 was dropped from scope
during planning (`smaqit.task-plan`) — see Out of Scope — since it targets
`smaqit.task-start`/`smaqit.task-complete`, which live in the sibling `smaqit-extensions` repo, not
here.
