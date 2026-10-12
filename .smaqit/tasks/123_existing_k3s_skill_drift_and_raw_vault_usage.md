---
status: Completed
mode: Assisted
started: "2026-10-12"
created: "2026-10-06"
completed: "2026-10-13"
---

# existing-k3s Skill Family: Stale Text, Cross-Skill Drift, Raw Vault Usage, and an Un-Upstreamed vault-loader Patch

## Description

The same 2026-10-06 audit found the existing-k3s skills out of date with the platform they were
modelled on, inconsistent with each other, and telling users to run raw `vault` commands:

- **Stale platform facts:** `deploy-k3s-app/SKILL.md:137-138, 169-171` says the scoped credential
  cannot read `pods/log` (it can); `deploy-k3s-app:180` and `vault-loader:261-262, 321-326` call
  kubeconfig rotation destructive (the platform rotates in place by recreating the ServiceAccount).
- **Cross-skill drift:** Environment naming `test`/`prod` vs the machine-slug Environments the live
  app uses; `/health` (template) vs `/health.json` (image-build skill) and no `DEPLOY_SHA` build arg;
  `ingress.rendered.yaml` vs `deployment/k3s/ingress.yaml`; `input-deployment:46` drops the
  machine segment of `secret/apps/<app>/<machine>/kubeconfig`; `new-greenfield-project:37` PAT path;
  `vault-loader:72-82` claims a k3s machine slug has no `secret/machines/` counterpart;
  `request-k3s-onboarding:61-78` treats "PR merged" as done (the platform also needs a dispatch and a
  Vault handoff) and its branch/title differ from the platform runbook; `repo-config` runs the
  VM-only `sync-secrets.sh` in k3s mode; `feature-new:121-122` resolves a deployed k3s project to
  the VM mode; `cicd-generate:91-92` says "single job" for a two-job template; the live app's verify
  retry was never upstreamed.
- **Raw `vault` usage** suggested to users (a user rule routes all Vault work through dedicated
  scripts): `deploy-k3s-app:40-47`, `repo-config:70-91` and `sync-secrets.sh`, `vault-loader:110-162`,
  `provision-cyso:74-198`, the rsync skills; `Bash(vault:*)` in allowed-tools.
- **Installed-only patch:** the installed `smaqit.infrastructure-vault-loader` scripts carry a local
  edit (a `netbird-machine` provisioning mode and a `machines/<slug>/netbird-setup-key` rotation path,
  `load-credentials.sh:45-52, 241-275`, `rotate-credential.sh:16-20, 91, 239-294`) that exists on no
  branch; a reinstall drops it.

## Issue Triage Context

**Mode:** Auto
**Technologies:** HashiCorp Vault, Bash, GitHub CLI, GitHub Actions, Kubernetes, Docker, Terraform
**Platforms/Environments:** Linux, WSL2, GitHub-hosted runners
**Features/Integrations:** Vault KV v2, GitHub Environment secrets/variables, kubeconfig, docker build args, Terraform environment variables, smaqit deployment slots
**Versions/Constraints:** smaqit 3.11.1, skill text and scripts stay app- and platform-agnostic, vault-loader script interfaces called by operator scripts stay stable

## Design Decisions

Confirmed by the user on 2026-10-12, during `task.plan`:

- **Vault access goes through purpose-named shipped scripts.** Skill text never tells the agent
  or user to run a `vault` command; scripts may call `vault` internally. New scripts:
  vault-loader `credential-file.sh` (one allowed credential to a 0600 file, path printed, value
  never) and `terraform-env.sh` (sourced provider and backend credential exports), and
  repo-config `sync-k3s-env.sh` (an app's machine-slug GitHub Environment).
- **Stable interfaces:** `load-credentials.sh` with no arguments (legacy flat scheme),
  `bootstrap-app-to-machine.sh <app> <machine>`, the scripts' location and names, and every Vault
  path and field convention stay unchanged, because operator scripts call them directly.
- **k3s app GitHub Environments are named by machine slug**, matching
  `secret/apps/<app>/<machine>/kubeconfig`. The slot plan's logical environments are unchanged.
- **Onboarding stays thin (task 116 rule):** `request-k3s-onboarding` ends at the merge and says
  the platform's own handoff must deliver the kubeconfig; `target-credentials` stops with an
  operator message while the kubeconfig is absent. No platform dispatch or handoff mechanics are
  encoded. PR title: `chore: onboard <app> (<machine>)`.
- **Stale facts become platform-neutral:** `pods/log` is read only when the platform grants it
  (never `pods/exec`); kubeconfig rotation is whatever the platform defines — reload the reissued
  kubeconfig at the same path and redeploy.
- **The installed-only vault-loader patch is closed:** a reinstall already dropped it, it exists
  nowhere as code, and the operator's own Vault tooling covers it. Nothing to upstream.
- **Delivery:** one task, one PR, from `origin/main`.

## Implementation Steps

*Phase A — stale facts (platform-neutral)*

1. `deploy-k3s-app` (~96, 138, 170–171): `pods/log` is the platform's decision; read a failed
   rollout's logs only when granted, never `pods/exec`. Check the description against task 130's.
2. Rotation: `deploy-k3s-app` ~181, `vault-loader` ~326/~386, `rotate-credential.sh` ~25/~99 —
   rotation is platform-defined; reload the reissued kubeconfig at the same path and redeploy.

*Phase B — drift, one convention each*

3. Machine-slug GitHub Environments: `deploy.yml.k3s.template` (dispatch choice input,
   `__TEST_MACHINE_SLUG__`/`__PROD_MACHINE_SLUG__` tokens, prod default) and cicd-generate's token
   table; repo-config Step 5 via `sync-k3s-env.sh`; `deploy-k3s-app` ~13/35/152, the Ingress
   template comment, greenfield ~142–144.
4. `deploy.yml.k3s.template`: verify `/health.json`; `docker build --build-arg DEPLOY_SHA`.
   Align `image-build-push-static` and cicd-generate.
5. Verify retry: the k3s template retries 12 times, 5 s apart, then fails; `deploy-k3s-app`
   Step 10 re-runs deploy-verify within the same window.
6. One Ingress path: `deploy-k3s-app` renders into `deployment/k3s/ingress.yaml` and lints/applies
   `deployment/k3s/*.yaml`; no `ingress.rendered.yaml`.
7. Vault paths: `input-deployment` ~58 kubeconfig machine segment; greenfield ~26–30 PAT paths
   (`secret/apps/<app>/github`) and a `platform-repo` pre-condition for `existing-k3s`;
   `vault-loader` ~74–84 and `load-credentials.sh` ~162 drop the "no machines counterpart" claim.
8. Onboarding gate: `request-k3s-onboarding` final report and PR title; `target-credentials`
   stops with an operator message while the kubeconfig is absent.
9. Mode gating: repo-config Steps 2–4 VM modes only; `feature-new` (~124–126, ~256, ~269)
   reuses the deployed spec's `provisioning_mode` (confirm where the spec records it); remove
   cicd-generate ~82 "single job".

*Phase C — Vault through scripts*

10. `vault-loader/scripts/credential-file.sh <path> <field>`: allowlisted paths only, 0600 file in
    a private temp dir, prints the path.
11. `vault-loader/scripts/terraform-env.sh` (sourced, `--unset`): machine-scoped paths when
    `MACHINE_SLUG` is set, legacy flat paths otherwise.
12. `repo-config/scripts/sync-k3s-env.sh <owner>/<repo> <app> <machine> [--host] [--registry-org]`:
    create/verify the Environment, PAT preflight, stdin secrets, `APP_HOST`, verification.
13. `load-credentials.sh` optional `REGISTRY_ORG=<org>` (existing-k3s) for
    `secret/organizations/<org>/github-package-read`; `rotate-credential.sh` accepts that path.
14. `setup-vault.sh` enables kv-v2 idempotently, as its header claims.
15. Replace every inline `vault` usage: deploy-k3s-app, repo-config Step 5, provision-cyso Step 2
    (`register-machine.sh` + `terraform-env.sh`) and ~170, the three rsync skills, vault-loader
    ~112/116/148, request-k3s-onboarding's token; reword illustrative gotchas.
16. Remove `Bash(vault:*)` from deploy-k3s-app, deploy-rsync-python-tornado,
    request-k3s-onboarding and their `.smaqit/definitions/skills/` files.

*Phase D — guards, tests, docs*

17. Go test over the embedded skills: fail on a run-able `vault` command or `Bash(vault` in any
    SKILL.md or references file.
18. Stub tests for the new scripts, `REGISTRY_ORG`, and setup-vault's kv-v2 step; extend the k3s
    template test (`/health.json`, build arg, retry, machine-slug environment).
19. Version bumps, compendium entries, `make -C installer prepare test smoke-test`; CHANGELOG via
    `task.complete` (notes the Environment rename for projects regenerating k3s workflows).

## Known Issues Triage
**Triaged:** 2026-10-12
**Tools searched:** GitHub CLI (`cli/cli`), HashiCorp Vault (`hashicorp/vault`)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#8817 gh variable/secret commands should url-encode environment name](https://github.com/cli/cli/issues/8817) — `cli/cli` — bug, priority-3. Machine-slug Environment names need no encoding; `sync-k3s-env.sh` validates the slug's characters.
- [#5149 create/edit/list/delete repo environment](https://github.com/cli/cli/issues/5149) — `cli/cli` — enhancement. No `gh` Environment command; create Environments with `gh api … -X PUT`.
- [#13293 Unable to get just the secret values (data) if the secret itself contains a "data" key](https://github.com/hashicorp/vault/issues/13293) — `hashicorp/vault` — secret/kv. Not triggered by smaqit's field names; scripts read single fields with `-field=`.

### Historical (Closed)
- [#13522 gh run rerun uses stale secret snapshot from workflow-run-create time, not rerun-execute time](https://github.com/cli/cli/issues/13522) — `cli/cli` — closed. After a kubeconfig rotation, redeploy with a fresh dispatch, not a re-run.
- [#3579 `gh secret load_file envfile`](https://github.com/cli/cli/issues/3579) — `cli/cli` — closed

### Unresolvable Tools
- None

### Omitted Tools
- Bash, GitHub Actions, Kubernetes, Docker, Terraform — excluded after relevance prioritization; Terraform and Docker were searched earlier today for tasks 114, 115, and 134 with no blocking results

### Search Warnings
- None. Research-map task block had no GitHub URLs; all repositories came from the helper fallback.

## Acceptance Criteria

- [x] The `pods/log` and rotation statements are platform-neutral in deploy-k3s-app, vault-loader, and `rotate-credential.sh`.
- [x] k3s app skills and the k3s template use machine-slug GitHub Environments; the template verifies `/health.json` with a bounded retry and passes `--build-arg DEPLOY_SHA`; one Ingress path (`deployment/k3s/ingress.yaml`) everywhere.
- [x] Vault paths are complete and consistent (kubeconfig machine segment, app-scheme PATs, no "no machines counterpart" claim); repo-config's VM steps are gated off for `existing-k3s`; feature-new no longer defaults a deployed k3s project to `existing-owned`; cicd-generate's "single job" leftover is gone.
- [x] request-k3s-onboarding reports that merged is not yet deployable and uses `chore: onboard <app> (<machine>)`; the `target-credentials` operation stops with an operator message while the kubeconfig is absent.
- [x] `credential-file.sh`, `terraform-env.sh`, `sync-k3s-env.sh`, `REGISTRY_ORG`, and the setup-vault kv-v2 step exist and are covered by stub tests; existing script interfaces are unchanged.
- [x] No SKILL.md or references file contains a run-able `vault` command or `Bash(vault`, enforced by a Go test that fails when one is reintroduced.
- [x] The installed-only netbird patch is recorded as already dropped (nothing to upstream).
- [x] `make -C installer test` and `make -C installer smoke-test` pass; compendium and `CHANGELOG.md` are updated.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Re-checked every audit item on main (v3.11.1) during `task.plan`: 12 of 13 still present, one partly fixed; two new items found (no `DEPLOY_SHA` build arg in the k3s template; vault-loader's "no machines counterpart" contradicting task 114). Confirmed the platform-side facts against the live downstream repositories.
- Stale facts made platform-neutral in deploy-k3s-app, vault-loader, and `rotate-credential.sh`: `pods/log` read only when granted (never `pods/exec`); rotation is the platform's decision; redeploy with a fresh dispatch after rotating.
- k3s template: machine-slug Environment via a `workflow_dispatch` choice (`__PROD_MACHINE_SLUG__`/`__TEST_MACHINE_SLUG__`, prod on push), `--build-arg DEPLOY_SHA`, `__HOST__` stamped into the committed Ingress before lint, `/health.json` verified with a 12×5 s retry. deploy-k3s-app renders `deployment/k3s/ingress.yaml` (host left as a token) and re-runs verification within the same window.
- New scripts: vault-loader `credential-file.sh` (allowlisted credential to a 0600 file, path printed) and `terraform-env.sh` (sourced exports, `--unset`); repo-config `sync-k3s-env.sh` (machine-slug Environment, stdin secrets, PAT preflight). `load-credentials.sh` gained `REGISTRY_ORG` and an idempotent kv-v2 enable and now writes k3s secrets on stdin; `rotate-credential.sh` accepts `organizations/<org>/github-package-read`; `setup-vault.sh` no longer claims to enable kv-v2.
- Every run-able `vault` command and `Bash(vault:*)` grant removed from shipped skill text and the three definition files; drift fixed in input-deployment, greenfield, feature-new, repo-config, cicd-generate, image-build-push-static, request-k3s-onboarding, provision-cyso, and the rsync skills; compendium updated with a new script-only Vault entry.
- Tests: four stub-script tests, a k3s template contract test, and `skill_text_test.go`, which fails on a run-able `vault` command in any shipped skill document. Deliberate breaks (a leaked token, `/health`, an inserted `vault kv get`) each made a test fail. `make -C installer test` and `make -C installer smoke-test` pass.

**Decisions made:**
- The committed `deployment/k3s/ingress.yaml` keeps `__HOST__` as a token because each environment has its own hostname; CI substitutes `APP_HOST` before lint and a local run writes a temp copy — a single rendered file could hold only one host.
- kv-v2 is enabled by `load-credentials.sh` after unseal and login: `setup-vault.sh` runs while Vault is still sealed, so it can't enable an engine.
- `terraform-env.sh` reads the project's own paths first and falls back to machine and app paths, preserving provision-cyso's existing credential sources (never verified with a live apply) rather than migrating them.
- `feature-new` detects a deployed k3s project from the Infrastructure spec's `Platform Repo`/`Machine Slug` Constraints rows, since the spec doesn't record `provisioning_mode` itself.
- Script interfaces the operator's own tooling calls (`load-credentials.sh` with no arguments, `bootstrap-app-to-machine.sh <app> <machine>`) stayed byte-identical in behaviour; the legacy and new-scheme sections of `load-credentials.sh` are unchanged.
- The installed-only netbird vault-loader patch was already dropped by a reinstall and exists nowhere as code; the operator's own Vault tooling covers it, so nothing was upstreamed.

**Blockers encountered:**
- None

**Follow-up identified:**
- Existing k3s app projects adopt machine-slug Environments when they regenerate workflows: sync into the new Environments with `sync-k3s-env.sh`, then delete the old `test`/`prod` ones.
- `load-credentials.sh` still passes the unseal key to `vault operator unseal` as an argument, and the legacy flat-scheme writes still pass values as arguments; moving them to stdin was out of scope (the legacy section had to stay unchanged).
- The k3s template still waits on `certificate/<app>-tls`, which a platform that owns the Ingress itself (for example behind an SSO gate) may name differently; a project-level edit for now.

## Files to Create / Modify

| File | Action |
|------|--------|
| skills/smaqit.infrastructure-* (deploy-k3s-app, cicd-generate, repo-config, request-k3s-onboarding, vault-loader, provision-cyso, deploy-rsync*, image-build-push-static), skills/smaqit.input-deployment, skills/smaqit.new-greenfield-project, skills/smaqit.feature-new | Modify |
| skills/smaqit.infrastructure-vault-loader/scripts/{credential-file,terraform-env}.sh, skills/smaqit.infrastructure-repo-config/scripts/sync-k3s-env.sh | Create |
| .smaqit/definitions/skills/{deploy-k3s-app,deploy-rsync-python-tornado,request-k3s-onboarding} definitions | Modify (remove `Bash(vault:*)`) |
| installer/skill_scripts_test.go, installer/cicd_templates_test.go, new no-raw-vault test | Modify / Create |
| .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Source: a downstream platform's consistency audit, 2026-10-06.
- **Planned 2026-10-12 (`task.plan`):** discovery re-checked every item on main (v3.11.1): 12 of
  13 still present, cicd-generate's "single job" partly fixed. New items found: the k3s template
  never passes the `DEPLOY_SHA` build arg the image-build skill expects; vault-loader's "no
  machines counterpart" contradicts task 114. The platform side confirmed `pods/log` is granted,
  rotation is in place, and onboarding completes only after a dispatch and a Vault handoff.
- **Re-scoped by task 125 (2026-10-11):** "Vault access goes through a named script interface the project provides" becomes the `credentials` slot contract; skills reach Vault only through the filler bound to that slot.
- **Refreshed after task 125 implementation (2026-10-11):** the shipped credential slots are `access-credentials` (before onboarding; vault-loader `CREDENTIAL_SCOPE=access` for k3s) and `target-credentials` (kubeconfig / per-app key). task 125 also fixed `input-deployment`'s k3s step ordering only insofar as the agent now loads the kubeconfig after onboarding; the drift items listed above remain open.
