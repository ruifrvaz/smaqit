# Platform Repositories and Script-Only Vault

## Metadata

- **Date:** 2026-10-11 → 2026-10-13
- **Session focus:** Finish the slot-era follow-ups for repositories that own machines (platform repositories) and for `existing-k3s` apps: machine registration and Environments, guarded instance replacement, a platform-repository guide, and moving every Vault operation in skill text behind scripts
- **Tasks completed:** 114 (v3.9.0, PR #96), 115 (v3.11.0, PR #99), 134 (v3.11.1, PR #100), 123 (v3.13.0, PR #102)
- **Tasks created:** 134 (smaqit); 042 (smaqit-extensions, left uncommitted at the user's request)
- **Tasks referenced:** 130, 131, 132, 133 (completed in parallel sessions), 122 (next), 126, 113

## Actions Taken

- Ran `session.start`; corrected stale memory for tasks 130/131 (131's plan was already approved; 130's YAML step already done by task 127). Pushed local `main`; found task 131 already started by a parallel session.
- Checked whether task 131 blocked the slot follow-ups: 114, 115, 123 could proceed; 122 waits for 131's `input-deployment` contract.
- **Task 114 (machine-monorepo learnings):** planned with three discovery agents. Finding 6 (release preflight) belonged to smaqit-extensions and moved to its new task 042, with the isolated-worktree caveat. Implemented: corrected Cyso facts (virtio-scsi `/dev/sdb`, `tier-2` default, retype, one image ID, org-level PAT grants); `register-machine.sh` (`target-credentials=register`); repo-config `sync-machine` and `check-pat-permissions.sh`; cicd-generate gotchas; deployment-agent provider routing and worktree caveat; a stub vault/gh Go test harness. Released v3.9.0.
- **Task 115 (rescoped):** discovery against the downstream platform repository showed every machine had moved to k3s over a mesh network and its VM tenancy had been replaced by k3s Namespace onboarding; its mechanisms are its own project skills, bound into slots by its own task. Dropped the four canonical skills. Shipped instead: `plan-guard.sh --allow-replace/--out`, the Terraform 1.6+ Cyso backend block and Nova `user_data` limit, a Platform Repositories wiki guide, and app-side cross-references. Released v3.11.0.
- Verified the user-run Vault metadata backfill from task 114 was unnecessary: `vault-add-credential --check` showed all three downstream machines already complete.
- **Task 134 (created this session):** the generated `full`-mode `deploy.yml` now applies the guarded plan file instead of re-planning with `-auto-approve`. Released v3.11.1.
- **Task 123 (k3s drift and raw Vault):** re-audited 13 items on main (12 still present) and confirmed platform facts downstream. Made `pods/log` and rotation platform-neutral; moved k3s apps to machine-slug GitHub Environments; added `--build-arg DEPLOY_SHA`, `/health.json` with a 60 s retry, one Ingress path with a per-environment `__HOST__` token; onboarding gate on the kubeconfig; new `credential-file.sh`, `terraform-env.sh`, `sync-k3s-env.sh`; `REGISTRY_ORG` and kv-v2 enable in `load-credentials.sh`; removed every raw `vault` command and `Bash(vault:*)` grant from skill text, enforced by a Go test. Released v3.13.0.
- Pushed task bookkeeping and ran `session.finish`.

## Problems Solved

- **`sync-machine` ordering:** `environment-config` runs before `provision`, so the machine has no IP on the first run; `FIXED_IP` became optional and is set on a single-slot re-run.
- **Changelog collisions across parallel releases:** merged `origin/main` (never rebasing) into task branches before writing their changelog section when another release had landed after branching.
- **One committed Ingress, several hosts:** kept `__HOST__` as a token in `deployment/k3s/ingress.yaml`, stamped per environment at apply time.
- **kv-v2 enable before unseal:** impossible in `setup-vault.sh`; moved to `load-credentials.sh` after authentication.
- **Shipped skills can't link to the wiki:** `docs/wiki/` isn't installed, so cross-references in skills are self-contained sentences.
- **Guard testing:** every new script and guard was proven by a deliberate break that a test caught.

## Decisions Made

- smaqit defines hand-off points, not one contributor's mechanisms (task 116 rule): no canonical machine-provision, verify, host-baseline, or tenancy skills.
- Skill text reaches Vault only through purpose-named scripts; scripts may call `vault`; secrets go to `vault` and `gh` on stdin.
- The interfaces the operator's own Vault tooling calls (`load-credentials.sh` with no arguments, `bootstrap-app-to-machine.sh <app> <machine>`) and all Vault path and field conventions stay stable.
- k3s app GitHub Environments are named by machine slug; the slot plan's logical environments are unchanged.
- Onboarding stays thin: a merged request is not deployable; `target-credentials` gates on the platform's kubeconfig hand-off.
- `terraform-env.sh` keeps provision-cyso's project-path-first credential sources (never verified with a live apply).
- v3.13.0 shipped as MINOR with an action-required changelog note for the Environment rename.
- Every live Vault step runs through the user's own scripts; the agent ran only `vault-add-credential --check`.

## Files Modified

- `skills/smaqit.infrastructure-vault-loader/` — `SKILL.md`; new `register-machine.sh`, `credential-file.sh`, `terraform-env.sh`; `load-credentials.sh`, `rotate-credential.sh`, `setup-vault.sh`
- `skills/smaqit.infrastructure-repo-config/` — `SKILL.md`; new `sync-machine-env.sh`, `check-pat-permissions.sh`, `sync-k3s-env.sh`
- `skills/smaqit.infrastructure-provision-cyso/` — `SKILL.md`, `scripts/plan-guard.sh` (`--allow-replace`, `--out`)
- `skills/smaqit.infrastructure-provider-cyso/` — `SKILL.md`, `references/{cyso-reference,cyso-deployment-setup,cyso-pricing,openstack-forcenew-attributes}.md`
- `skills/smaqit.infrastructure-cicd-generate/` — `SKILL.md`, `assets/deploy.yml.full.template`, `assets/deploy.yml.k3s.template`
- `skills/smaqit.infrastructure-deploy-k3s-app/` — `SKILL.md`, `assets/ingress.yaml.template`
- `skills/smaqit.infrastructure-request-k3s-onboarding`, `-image-build-push-static`, `-deploy-rsync*`, `smaqit.input-deployment`, `smaqit.new-greenfield-project`, `smaqit.feature-new` — `SKILL.md`
- `agents/deployment.md` — provider routing, worktree caveat, verify via the bound filler
- `.smaqit/definitions/skills/` — three definitions without `Bash(vault:*)`
- `installer/skill_scripts_test.go`, `installer/cicd_templates_test.go`, `installer/skill_text_test.go`, `installer/slots_test.go`
- `docs/wiki/workflows/platform-repositories.md` (new), `docs/wiki/workflows/deployment-slots.md`, `README.md`
- `.smaqit/compendium.md`, `CHANGELOG.md`, `.smaqit/references/project-research.md`
- `.smaqit/tasks/114, 115, 123, 134`, `PLANNING.md`; smaqit-extensions `.smaqit/tasks/042_*.md`, `PLANNING.md` (uncommitted)

## Next Steps

- Task 122 (private, mesh-only k3s targets): both blockers (131, 123) are merged.
- smaqit-extensions task 042 is written but uncommitted, at the user's request.
- Existing k3s apps: run `sync-k3s-env.sh` per target machine before regenerating workflows, then delete the old `test`/`prod` Environments.
- Downstream platform repository: its own slot-binding task can use `plan-guard.sh --allow-replace`.
- Follow-ups recorded in task 123's Findings: the unseal key and legacy-scheme writes still pass values as arguments; the k3s template's `certificate/<app>-tls` wait may not match a platform that owns the Ingress.
- Unfiled from earlier sessions: installer skill-retirement pruning gap; tasks 126 and 113.

## Session Metrics

- Duration: three days (2026-10-11 → 2026-10-13)
- Tasks completed: 4 (114, 115, 134, 123); tasks created: 1 here (134) plus 1 in smaqit-extensions (042)
- Releases: v3.9.0, v3.11.0, v3.11.1, v3.13.0 (PRs #96, #99, #100, #102)
- New scripts: 7 (`register-machine.sh`, `sync-machine-env.sh`, `check-pat-permissions.sh`, `credential-file.sh`, `terraform-env.sh`, `sync-k3s-env.sh`, plus `plan-guard.sh` rewritten); new Go test files: 3
- Discovery agents: 10 Explore agents across four plans
- Deliberate-break checks: 7, each caught by a test
