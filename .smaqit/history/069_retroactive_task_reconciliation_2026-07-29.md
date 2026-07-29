# Retroactive Task Reconciliation

**Date:** 2026-07-29
**Session Focus:** Discover and reconcile task 096, whose implementation had been written in a prior session but never committed or merged; review that implementation for correctness, fix a real gap found during review, and complete the task properly.
**Tasks Referenced:** 096
**Tasks Completed:** 096

---

## Actions Taken

### Session Initialization
- Ran `session.start`: loaded README, `.github/copilot-instructions.md`, PLANNING.md, compendium, and history 068.
- Found task 096 already marked "Completed" in `PLANNING.md`/its own task file (Findings fully populated, 7/8 ACs checked) but with all 15 implementation files sitting as **uncommitted working-tree changes on `main`** — no branch, no worktree, nothing in git history. Flagged this as the top priority before any new task work.

### Retroactive `task.start 096`
- Created `task/096-existing-unmanaged-provisioning-mode` branch + worktree via the standard `task-start` → `smaqit.utils.worktree` flow.
- Moved the uncommitted implementation off `main` into the new worktree via a scoped `git stash push -- <13 files>` / `stash pop` in the worktree, keeping `main` clean throughout.
- Copied the untracked task file into the worktree; ran `smaqit.utils.triage-issues 096` (Terraform, HashiCorp Vault — Result: Clear, no blocking upstream issues).

### Implementation Review
- Reviewed the 13-file, 175-line diff against all 8 Acceptance Criteria, independently re-running the grep AC6 claimed was clean rather than trusting the self-reported Findings.
- ACs 1–5 and 7 confirmed correct. **AC6's "no orphaned `existing-shared` mention" claim was false**: `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` (the legacy flat-scheme credential loader, documented as "run every session") was never touched or mentioned anywhere in the task's Design Decisions/Implementation Steps/Findings, despite being the other place `PROVISIONING_MODE`-aware branching lives.
- User pushed back on the initial framing ("does it fail to add credentials?") — clarified the real defect: `ssh`/`github` still get created either way; the actual bug was (a) wrongly prompting for unneeded `cyso`/`tfstate` credentials, and (b) more seriously, the SSH-keypair step falling into the `provision`/`existing-owned` branch, which generates a keypair silently because Terraform is expected to push it via cloud-init — `existing-unmanaged` has no such Terraform run, so the operator would get a keypair with no instruction to manually install it, a silent precondition for the next deploy's SSH step.

### Fix and Completion
- Fixed `load-credentials.sh`: added a `RESTRICTED_MODE` variable covering both `existing-shared`/`existing-unmanaged` for the cyso/tfstate skip and path accounting, and gave `existing-unmanaged` its own SSH branch (always generate + print manual-install instructions, no "copy from owning project" option since none exists in this mode). Updated `smaqit.infrastructure-vault-loader/SKILL.md`'s scheme-detection section and usage examples to match. Also fixed a minor prose gap in `smaqit.feature-new/SKILL.md`'s Scope section.
- Verified with `bash -n`, `shellcheck` (no new findings beyond pre-existing style-level info), and a standalone simulation of the mode-decision logic across all four `provisioning_mode` values.
- Updated the task file's Findings with a "Review Findings" section documenting what was found and fixed, corrected the AC6 checkbox annotation, and updated the Files to Create/Modify table.
- Committed on the task branch (implementation + fix, then PLANNING.md status), merged to `main` with `--no-ff`, removed the worktree, deleted the branch. Restored the two Active-table rows (095, 097) that had been separately pending in `main`'s uncommitted `PLANNING.md` before this session — left untouched and unrelated to 096.

## Problems Solved

- **Uncommitted "Completed" work:** a task marked Completed in planning files but never actually committed — now fully reconciled through the standard branch/worktree/merge lifecycle.
- **Silent AC6 falsification:** the task's own verification claim (clean full-tree grep) didn't hold under independent re-verification; caught before merge rather than shipping a real gap under a false "verified clean" label.
- **Legacy flat-scheme `existing-unmanaged` gap:** a project on the pre-`apps/`+`machines/` credential scheme choosing `existing-unmanaged` would previously get a Vault SSH keypair with no install instructions — an unexplained SSH failure at the next deploy. Now handled with its own branch, no false parity with `existing-shared`'s "copy from owning project" option.

## Decisions Made

- Treat `load-credentials.sh`'s gap as in-scope for task 096 itself (not a new follow-up task) since it directly falsifies an AC the task claims to have already verified — fixing forward rather than shipping a known-false claim.
- Reused the stash-based worktree-migration technique (scoped `git stash push -- <paths>` / `pop` in the new worktree) to move already-written, uncommitted implementation into a freshly created task branch without touching unrelated uncommitted content on `main`.
- Left the pre-existing, unrelated 095/097 `PLANNING.md` Active-table rows and their untracked task files exactly as found — out of scope for this session, no `task.start` run against them.

## Files Modified

- `skills/smaqit.input-deployment/SKILL.md`, `skills/smaqit.new-greenfield-project/SKILL.md`, `skills/smaqit.feature-new/SKILL.md`, `skills/smaqit.infrastructure-provision-cyso/SKILL.md` (+ `ownership-guard.sh`), `skills/smaqit.infrastructure-repo-config/SKILL.md` (+ `sync-secrets.sh`), `skills/smaqit.infrastructure-cicd-generate/SKILL.md` (+ both deploy templates), `skills/smaqit.infrastructure-vault-loader/SKILL.md` (+ `bootstrap-app-to-machine.sh`, `load-credentials.sh`), `skills/smaqit.infrastructure-deploy-rsync/SKILL.md` — `existing-unmanaged` provisioning mode threaded through, plus the review-pass fix to `load-credentials.sh`
- `.smaqit/tasks/096_existing_unmanaged_provisioning_mode.md` — created (committed), Findings extended with Review Findings, AC6 annotated, Files table updated
- `.smaqit/tasks/PLANNING.md` — 096 moved to Completed

## Next Steps

- Tasks 095 and 097 remain Active/untouched — 095 needs `smaqit.task-plan 095` (cross-repo design question re: `smaqit.feature-new`'s per-phase branch/worktree behavior); 097 needs a repo-wide downstream-project-name audit before redaction.
- Task 096's one deliberately open AC (live walkthrough via the originating downstream project's real second-VPS deployment) remains open, tracked in that project's own planning, unaffected by this session.
- 094, 077, 074, 071, 070 remain in the Active backlog, unchanged.

## Session Metrics

- **Tasks completed:** 1 (096, reconciled from a prior session's uncommitted state)
- **Files modified:** 16 (15 implementation + 1 task file), plus `PLANNING.md`
- **Real defect found and fixed during review:** 1 (legacy-scheme `load-credentials.sh` `existing-unmanaged` gap)
- **Merge:** `cec4864` (`--no-ff`) on `main`; task branch and worktree cleaned up
