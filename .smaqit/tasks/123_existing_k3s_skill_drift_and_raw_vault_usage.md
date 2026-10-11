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

**Mode:** Skip
**Technologies:** None
**Platforms/Environments:** None
**Features/Integrations:** None
**Versions/Constraints:** skill text and scripts only, app- and platform-agnostic

## Design Decisions

- **Vault access in skills goes through a named script interface** the project provides, never
  inline `vault` commands; skills name the operation and the script, and stop if none exists.
- **The vault-loader patch is either upstreamed (generalized) or explicitly dropped** — decide after
  reading it; never left installed-only.

## Implementation Steps

1. Correct the stale facts.
2. Reconcile the drift items to one convention each (machine-slug Environments, health path and
   build arg, Ingress filename, Vault paths, onboarding completion semantics, mode gating).
3. Replace raw `vault` guidance and `Bash(vault:*)` grants.
4. Diff the installed vault-loader scripts against source; upstream or drop.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Every stale and drift item above is resolved to one convention
- [ ] No skill suggests a raw `vault` command or grants `Bash(vault:*)`
- [ ] Installed vault-loader matches source after a reinstall

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
| skills/smaqit.infrastructure-* (deploy-k3s-app, cicd-generate, repo-config, request-k3s-onboarding, vault-loader, provision-cyso, deploy-rsync*), skills/smaqit.input-deployment, skills/smaqit.new-greenfield-project, skills/smaqit.feature-new | Modify |

## Notes

- Source: a downstream platform's consistency audit, 2026-10-06.
- **Re-scoped by task 125 (2026-10-11):** "Vault access goes through a named script interface the project provides" becomes the `credentials` slot contract; skills reach Vault only through the filler bound to that slot.
- **Refreshed after task 125 implementation (2026-10-11):** the shipped credential slots are `access-credentials` (before onboarding; vault-loader `CREDENTIAL_SCOPE=access` for k3s) and `target-credentials` (kubeconfig / per-app key). task 125 also fixed `input-deployment`'s k3s step ordering only insofar as the agent now loads the kubeconfig after onboarding; the drift items listed above remain open.
