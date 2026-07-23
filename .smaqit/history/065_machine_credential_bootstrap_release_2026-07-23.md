# Machine Credential Bootstrap and Release

**Date:** 2026-07-23
**Session Focus:** Close out task 089, plan and implement task 090 (machine-scoped Vault credential namespace + app bootstrap script), cut and push release v1.8.0, and close out four completed tasks (089, 084, 087, 086, 085, 090) with disciplined acceptance-criteria enforcement.
**Tasks Referenced:** 089, 090, 084, 087, 086, 085
**Tasks Completed:** 089, 084, 087, 086, 085, 090

---

## Actions Taken

### Session Setup and Task 089 Close-out
- Loaded project context, planning state, glossary/compendium, and prior history via `session.start`.
- Completed task 089 (`smaqit.feature-new` skill) — all 8 ACs were already checked from the prior session; wrote Findings and moved it to PLANNING.md's Completed table.

### Task 090 Planning
- Ran `task.plan 090` against a task file with fully-drafted Design Decisions but two real gaps: an unresolved dependency (task 084's status label said "In Progress" despite being functionally complete) and a namespace design that hadn't been checked against real infrastructure. Three parallel Explore subagents confirmed 084 was done in-codebase (stale label only), surveyed `smaqit.infrastructure-vault-loader`'s existing script conventions, and located the exact SKILL.md text needing edits in `new-greenfield-project` and `provision-cyso`.
- User then supplied a real correction: apps live under `secret/apps/<app-slug>/*`, not the originally-drafted flat `secret/<project-slug>/*`, and `cyso`/`tfstate` (not just `base-ssh`/`metadata`) belong at the machine level. Verified this directly against the real, already-unsealed local Vault instance — `secret/machines/<machine-slug>/{base-ssh,cyso,tfstate,metadata}` and `secret/apps/<app-slug>/ssh` already existed as hand-built prototypes, while legacy projects still used the old flat scheme.
- Updated task 090's Design Decisions, Implementation Steps, Acceptance Criteria, and Files table in place to match the real namespace shape, adding two implementation steps the original draft missed (`load-credentials.sh` and `smaqit.infrastructure-repo-config` both needed real changes, not just confirmation).

### Task 084 Close-out
- Completed task 084 (Deploy Target Resolution) — Findings were already fully written and 8/9 ACs checked from a prior session; only the status label was stale. Status → Completed.

### Task 090 Implementation
- Ran `task.start 090` (Assisted mode); issue triage against `hashicorp/vault` and `openssh/openssh-portable` returned Advisory only (one UI-only Vault issue, not applicable to this task's CLI-only multi-segment path usage) — non-blocking.
- Rewrote `smaqit.infrastructure-vault-loader/SKILL.md` for the two-namespace scheme (`apps/`+`machines/`), documenting the legacy flat scheme as still-supported-but-unmigrated.
- Wrote `scripts/bootstrap-app-to-machine.sh` new — idempotent, matches existing credential-handling idioms (`@file` private keys, `mktemp -d`+`rm -rf`, piped stdin for remote `authorized_keys` writes, `ssh -o BatchMode=yes` verification).
- Restructured `scripts/rotate-credential.sh` into dual roots (`apps/*` for ssh/github, `machines/*` for base-ssh/cyso/tfstate), with `base-ssh` given its own generate/install-with-old-key/retire sequence rather than delegating to `load-credentials.sh`.
- Restructured `scripts/load-credentials.sh` with new-scheme detection (via an explicit `MACHINE_SLUG` or an existing `secret/apps/<slug>/machine` pointer) that populates only `github` for new-scheme apps, leaving `ssh` exclusively to the bootstrap script and `cyso`/`tfstate` to machine registration.
- Updated `smaqit.infrastructure-provision-cyso/SKILL.md`: `base-ssh` now generated pre-`terraform apply` (so Terraform's keypair resource installs it) rather than post-apply as the original draft implied; added a post-apply machine-registration + self-bootstrap step.
- Updated `smaqit.new-greenfield-project/SKILL.md`'s `existing-shared` branch to call the bootstrap script instead of the old "reuse owner's key across Vault namespaces" guidance.
- Updated `smaqit.infrastructure-repo-config/SKILL.md` + `scripts/sync-secrets.sh` to be scheme-aware (ssh/github from the app root, tfstate/cyso from the machine root, collapsing to the same root on the legacy scheme).
- Verified all touched/new scripts with `bash -n` and `shellcheck` (clean). Live-tested `bootstrap-app-to-machine.sh` against real Vault/infra — confirmed the idempotent no-op path works correctly against a genuinely live, already-authenticating pairing.
- Handed back in Assisted mode with 2/6 ACs checked and 3 explicitly left unverified (no fresh VM available to test machine registration or `base-ssh` rotation in-sandbox), plus a flagged discrepancy: the real test app's pairing already authenticated, contradicting the task's original "disconnected keypair" framing.

### Release v1.8.0
- Invoked the `smaqit-release-local` agent. It judged the change a backward-compatible **minor** feature (not a patch layered onto the prior unreleased CI fixes), landing on v1.8.0: updated `CHANGELOG.md` and `installer/main.go`, verified `go build`/`go test`, committed `e72a360` ("Release v1.8.0"), tagged `v1.8.0`, and reported push was not possible from the sandbox (same known SSH/askpass limitation as prior sessions).
- Operator had separately updated the `smaqit.release-git-local` skill with a new "Desktop Linux SSH Agent Recovery" procedure. Followed it: discovered an already-unlocked GNOME Keyring SSH agent socket (`/run/user/1000/gcr/ssh` — unlocked at desktop login, not by this session) via the documented discovery steps, then retried `git push origin main` scoped to that socket for a single command. Push succeeded (`123fa7a..e72a360`).
- The subsequent `git push origin v1.8.0` (tag) and a read-only `ls-remote` check were both denied by the Claude Code auto-mode permission classifier (generic denial, no further reasoning surfaced). Operator pushed the tag from their own shell; confirmed via `gh run list`/`gh release view` that `post-merge-release.yml` fired successfully and the GitHub Release published with all 5 platform binaries.

### Disciplined Task Close-out (089, 084, 087, 086, 085, 090)
- Session recap surfaced 4 tasks marked "In Progress" in PLANNING.md (090, 087, 086, 085) at session-finish time. Rather than blindly completing all on the operator's initial "complete all tasks" instruction, read each task file in full first:
  - **087** and **086**: all ACs checked, Findings fully written from prior sessions — completed immediately.
  - **085**: flagged as blocked — its own Design Decisions declare a "Final review gate... cannot be skipped even in Autonomous mode," and the one unchecked AC (real target project validation) was explicitly still open per its own Findings addendum.
  - **090**: flagged as blocked — operator had just said they'd validate the remaining 3 ACs later, directly conflicting with immediate completion.
- Operator confirmed 085 was independently validated against a downstream project for real (closing the final review gate) and directed completing 090 anyway, citing an imminent validation against a new project (same precedent as task 084's prior closure with a documented, accepted gap). Operator explicitly affirmed the pushback was correct and asked for it to continue in future sessions.
- Completed 085 (AC checked, Findings addendum added citing the operator-confirmed downstream project validation) and 090 (Findings written fresh — this was 090's first completion pass — 3 ACs left honestly unchecked, matching what was and wasn't actually verified this session).

## Problems Solved

- **Task 090's original namespace draft was wrong** — verified against real Vault state before any implementation began, avoiding building automation around a documented-but-unimplemented shape.
- **`rotate-credential.sh`'s `base-ssh` rotation couldn't reuse the existing delete-and-repopulate pattern** — recognized during planning that `load-credentials.sh` has no machine-scoped keygen path, so `base-ssh` needed its own generate/install-with-old-key/retire sequence.
- **Terraform key-installation timing tension** — reconciled a real ordering conflict in the original task draft (base-ssh "generated after apply" vs. "Terraform installs it at provision time") by moving base-ssh generation to pre-apply in `provision-cyso`.
- **A stale skill-count assumption from the prior session's release was not repeated** — this release's build/test verification caught nothing broken, confirming the two independent fixes from the previous session held.
- **Blind "complete all tasks" instruction was not blindly executed** — read all four task files first, found two with genuine unmet, self-declared-mandatory gates, and surfaced the conflict instead of fabricating false completion status.

## Decisions Made

- `secret/apps/<app-slug>/*` and `secret/machines/<machine-slug>/*` replace the originally-drafted flat scheme for new projects; legacy flat-scheme projects remain unmigrated and unmodified by this task, by design.
- `cyso`/`tfstate` are machine-scoped, not app-scoped — provisioning is a property of the machine, not any one app on it (correction from real Vault evidence).
- `provision-cyso`'s own `cyso`/`tfstate` credential sourcing was deliberately left un-migrated to the machine root pending a live `terraform apply` test, rather than changed speculatively.
- Task completion requires genuinely met acceptance criteria or explicit, informed operator override — never a default "complete everything" sweep. Operator confirmed this is the desired standing behavior.
- Release versioning judgment (minor vs. patch) is made fresh each release by evaluating the actual diff, not assumed from the prior session's open recommendation.

## Files Modified

### Task 089 close-out
- `.smaqit/tasks/089_feature_new_post_mvp_workflow_skill.md`, `.smaqit/tasks/PLANNING.md`

### Task 090 (implementation)
- `skills/smaqit.infrastructure-vault-loader/SKILL.md`, `scripts/bootstrap-app-to-machine.sh` (new), `scripts/rotate-credential.sh`, `scripts/load-credentials.sh`
- `skills/smaqit.infrastructure-provision-cyso/SKILL.md`
- `skills/smaqit.new-greenfield-project/SKILL.md`
- `skills/smaqit.infrastructure-repo-config/SKILL.md`, `scripts/sync-secrets.sh`
- `.smaqit/tasks/090_machine_credential_namespace_and_app_bootstrap.md`, `.smaqit/tasks/PLANNING.md`

### Release v1.8.0
- `CHANGELOG.md`, `installer/main.go`

### Task close-outs (084, 087, 086, 085, 090)
- `.smaqit/tasks/084_deploy_target_resolution_provisioning_branch.md`
- `.smaqit/tasks/087_dynamic_stack_detection_and_skill_synthesis.md`
- `.smaqit/tasks/086_reconcile_downstream_python_nextjs_deploy_skill.md`
- `.smaqit/tasks/085_deterministic_cicd_workflow_templates_and_guard_vendoring.md`
- `.smaqit/tasks/PLANNING.md`

## Next Steps

- Operator to validate task 090's remaining 3 items (fresh machine registration, `base-ssh` rotation, mutating bootstrap path) against a new project.
- Optional: backfill `secret/apps/<app-slug>/machine` pointer in real Vault (one-line, additive, left undone as shared infra state).
- Legacy flat-scheme projects remain unmigrated to the new namespace, by design — a future task if ever needed.
- PLANNING.md Active is now down to 4 not-yet-started items (077, 074, 071, 070) — no in-progress work carried into the next session.

## Session Metrics

- **Tasks completed:** 6 (089, 084, 087, 086, 085, 090)
- **New scripts shipped:** 1 (`bootstrap-app-to-machine.sh`)
- **Scripts restructured:** 3 (`rotate-credential.sh`, `load-credentials.sh`, `sync-secrets.sh`)
- **Skills modified:** 5 (`vault-loader`, `provision-cyso`, `new-greenfield-project`, `repo-config`, plus doc-only touches)
- **Release:** v1.8.0 (published, all platform binaries, `post-merge-release` verified green)
- **Real-infra live tests:** 1 (idempotent bootstrap path against real Vault/VM)
- **Premature-completion pushbacks:** 2 (tasks 085 and 090, both resolved by operator confirmation before completing)
