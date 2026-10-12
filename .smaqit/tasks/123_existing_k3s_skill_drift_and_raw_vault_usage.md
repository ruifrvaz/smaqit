---
status: Not Started
created: "2026-10-06"
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

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The `pods/log` and rotation statements are platform-neutral in deploy-k3s-app, vault-loader, and `rotate-credential.sh`.
- [ ] k3s app skills and the k3s template use machine-slug GitHub Environments; the template verifies `/health.json` with a bounded retry and passes `--build-arg DEPLOY_SHA`; one Ingress path (`deployment/k3s/ingress.yaml`) everywhere.
- [ ] Vault paths are complete and consistent (kubeconfig machine segment, app-scheme PATs, no "no machines counterpart" claim); repo-config's VM steps are gated off for `existing-k3s`; feature-new no longer defaults a deployed k3s project to `existing-owned`; cicd-generate's "single job" leftover is gone.
- [ ] request-k3s-onboarding reports that merged is not yet deployable and uses `chore: onboard <app> (<machine>)`; the `target-credentials` operation stops with an operator message while the kubeconfig is absent.
- [ ] `credential-file.sh`, `terraform-env.sh`, `sync-k3s-env.sh`, `REGISTRY_ORG`, and the setup-vault kv-v2 step exist and are covered by stub tests; existing script interfaces are unchanged.
- [ ] No SKILL.md or references file contains a run-able `vault` command or `Bash(vault`, enforced by a Go test that fails when one is reintroduced.
- [ ] The installed-only netbird patch is recorded as already dropped (nothing to upstream).
- [ ] `make -C installer test` and `make -C installer smoke-test` pass; compendium and `CHANGELOG.md` are updated.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

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
