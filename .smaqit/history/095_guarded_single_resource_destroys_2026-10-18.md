# Guarded Single-Resource Destroys

## Metadata

- **Date:** 2026-10-18
- **Session focus:** Let a guarded repository apply a reviewed removal of one resource through `plan-guard.sh`, then close the gap that let an agent turn a blocked side-effect delete into an allowed one
- **Tasks completed:** 135 (v3.15.0, PR #104), 140 (v3.17.1, PR #107; created this session)
- **Tasks referenced:** 136, 137, 138, 139 (handled by a concurrent session), 126, 115 (introduced `--allow-replace`)

## Actions Taken

- **Session start:** loaded README, history 094, PLANNING, the compendium, and tasks 135 and 136. Found three problems in task 135's plan:
  - a step to hand-copy the script into the gitignored `installer/skills-*` build output;
  - a no-change plan with `--allow-delete` was not handled;
  - one address passed to both flags was not handled.
- **Task 135, start:**
  - created the branch and worktree from `origin/main`;
  - re-verified the 7-day-old project research map and added a task block;
  - issue triage came back Advisory (two sensitivity-rendering bugs in Terraform; `forget` actions noted as out of scope);
  - corrected the task's plan in the start commit.
- **Task 135, implementation:**
  - `--allow-delete <address>` accepts only a pure destroy (actions exactly `["delete"]`) of the named address, and adds nothing to `terraform plan`;
  - it is repeatable and combines with `--allow-replace` and `--out`;
  - a BLOCKED line now shows what the plan does with a listed address instead (`planned: …`);
  - added two Go tests and two bad-usage cases; broke the script three ways on purpose and each break failed a test;
  - updated provision-cyso (1.10.0), cicd-generate (3.1.0), the platform-repositories guide, the compendium and CHANGELOG.
- **Task 135, completion:**
  - Phase 1: release analysis gave v3.15.0 (MINOR); opened PR #104 and wrote the changelog on the branch;
  - Phase 2: after the user merged, verified the merge, marked the task Completed and removed the worktree.
- **Bench assessment, skipped by the user:** designed a 7-Case suite for agents choosing the wrong delete address. A live offline Terraform probe with `terraform_data` confirmed the count-index trap: removing element 0 of a `count` list plans a delete of `rule[2]` plus updates.
- **Design assessments:**
  - **The BLOCKED-output gap:** the guard's own message said "apply outside of this guarded path", and nothing said where an allow-flag address may come from.
  - **Break-glass-only deletes:** assessed, not adopted.
  - **"Does a delete always need a PR?"** No; see Decisions.
  - **Blocking local runs against GitHub Environment targets:** assessed; deferred by the user.
- **Task 140, created, started and completed:**
  - wording-only fix to the guard's BLOCKED footer, provision-cyso 1.10.1, cicd-generate 3.2.1, the Deployment agent's guard rule, the guide and the compendium;
  - added a Go test for the message;
  - merged `origin/main` (v3.17.0, task 138) into the branch before opening the PR, to avoid a CHANGELOG conflict;
  - released as v3.17.1 (PR #107) and cleaned up after the user merged.
- Checked that v3.15.0 published with all five binaries, and that it is an ancestor of the concurrent session's v3.16.0.

## Problems Solved

- **Hand-copying into build output:** task 135's plan said to copy the script into `installer/skills-shared/` and `skills-claude/`. Those are gitignored `make prepare` output, so the step was replaced with `make prepare`.
- **Turning a block into an allow:** the guard checks that the plan matches the flags, not that the flags match anyone's intent. Its BLOCKED message, the provision-cyso failure table and the Deployment agent's rules all failed to say "stop; the address must come from a human". Task 140 put the rule in each place, most importantly in the message the agent reads when it decides what to do next.
- **A stray edit on the primary checkout:** a compendium edit landed on the primary checkout instead of the task worktree. It was confirmed to be the only change there, reverted, and redone in the worktree without assuming a release version.
- **Concurrent sessions sharing the primary checkout:**
  - another session left an uncommitted task 138 file and PLANNING row; task 135's completion staged only its own PLANNING change, through `git update-index --cacheinfo`;
  - task 138's `[3.17.0]` CHANGELOG section would have clashed with task 140's; merging `origin/main` into the task branch before the PR avoided it.

## Decisions Made

- **`--allow-delete` semantics:**
  - exact address and pure destroy only;
  - fail-closed for a listed address the plan replaces, updates or doesn't touch;
  - a no-change plan with `--allow-delete` is BLOCKED;
  - one address in both flags is a usage error;
  - `forget` (a `removed` block with `destroy = false`) is out of scope.
- **Bench suite skipped.** The design is recorded above if it's revisited. The proposed comparison was the 1.10.0 skill text against 1.9.0, not against having no skill. Bench supports only Codex for now.
- **Break-glass-only deletes not adopted.** Downstream projects guard their own apply paths. smaqit ships the guard and wording, not a human-confirmation gate.
- **A delete does not always need a PR:**
  - local runs need none, and neither do variable-driven deletes;
  - only repository settings enforce one: branch protection, Environment branch policies, CI-only applies;
  - the generated `deploy.yml` runs the guard without allow flags, so it never applies a delete.
- **CI-only applies for GitHub Environment targets deferred.** Credential placement is the only enforcement that holds: scripts would refuse to hand Environment-target credentials to a local shell, and test runs would use separate state. GitHub OIDC is the ideal if Cyso supports it (unverified). A `GITHUB_ACTIONS` check only gives clear errors.
- **Task 140 is wording only, as a patch.** The verdict logic, flags and exit codes are unchanged, and the rule covers `--allow-replace` too.

## Files Modified

- `skills/smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh`: `--allow-delete` (135); BLOCKED footer (140)
- `skills/smaqit.infrastructure-provision-cyso/SKILL.md`: 1.10.0 (135), 1.10.1 (140); guard step and failure table
- `skills/smaqit.infrastructure-cicd-generate/SKILL.md`: 3.1.0 (135), 3.2.1 (140); gated replace/remove inputs
- `agents/deployment.md`: guard rule extended to blocked plans (140)
- `installer/skill_scripts_test.go`:
  - `TestPlanGuardAllowsOnlyTheNamedDestroy` and `TestPlanGuardCombinesAllowedDestroyReplaceAndOut` (135);
  - `TestPlanGuardBlockedOutputSaysStopAndReport` (140);
  - bad-usage cases.
- `docs/wiki/workflows/platform-repositories.md`, `.smaqit/compendium.md`, `CHANGELOG.md`: both tasks
- `.smaqit/tasks/135_plan_guard_allow_delete.md`, `.smaqit/tasks/140_plan_guard_blocked_guidance_stop_and_report.md` (new), `.smaqit/tasks/PLANNING.md`
- `.smaqit/references/project-research.md`: project table re-verified; task 135 and 140 blocks
- `smaqit.code-workspace`: worktree entries added and removed

## Next Steps

- **Deferred, no task filed:** CI-only applies for GitHub Environment targets (credential placement, separate test state, OIDC if Cyso supports it).
- **Bench suite for allow-flag behaviour:** designed and skipped. It is still the only way to measure whether agents follow the new BLOCKED rule.
- **Carried over:**
  - task 137 (credential slots under a scripts-only Vault policy);
  - task 126 (design-tools install lock);
  - the gap where installed skills a release drops are never removed;
  - sync `.smaqit/definitions/skills/` for the k3s skills;
  - smaqit-extensions task 042.
- A concurrent session is preparing v3.17.2 (`release/v3.17.2-k3s-trap-failure-rows`).

## Session Metrics

- Duration: one day (2026-10-18)
- Tasks completed: 2 (135, 140); 1 task created (140)
- Releases: v3.15.0 (PR #104, MINOR) and v3.17.1 (PR #107, PATCH)
- New tests: 3 Go test functions and 3 bad-usage cases; 4 deliberate-break checks, each caught
- Design assessments: 4 (Bench suite, break-glass deletes, PR requirement for deletes, local-run prevention); none adopted beyond the wording fix
- Live probe: 1 offline Terraform run confirming the count-index trap
