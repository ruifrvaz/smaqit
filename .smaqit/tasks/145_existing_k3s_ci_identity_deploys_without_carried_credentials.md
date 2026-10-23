---
status: Not Started
created: "2026-10-22"
---

# existing-k3s: CI-identity deploys without carried credentials (OIDC token, shared mesh key, platform-owned pulls)

## Description

The existing-k3s app-side skills assume the platform hands each app a long-lived scoped
kubeconfig, and on a mesh-only target a per-app NetBird setup key. The app then stores both as
`KUBECONFIG` and `NETBIRD_SETUP_KEY` on its machine-slug GitHub Environments:
- `smaqit.infrastructure-request-k3s-onboarding` opens the registry PR;
- `smaqit.infrastructure-vault-loader` and `smaqit.infrastructure-repo-config` read the kubeconfig
  and setup key from Vault and sync them to the Environment;
- `smaqit.infrastructure-cicd-generate` and `smaqit.infrastructure-deploy-k3s-app` generate and run
  the deploy workflow with those secrets.

In live use on a downstream platform, onboarding one app took about ten manual steps across three
people, and every credential passed through one operator's local Vault. That platform is
switching to identity-based deploys, in its task 061 (cross-repo parent) and task 062 (platform
side), both in its platform repository:
- the k3s API trusts GitHub Actions OIDC tokens and binds the namespace-scoped `deployer` Role to
  the app's repository and Environment, so no kubeconfig is issued;
- mesh reach uses one shared deploy key per environment (machine), held as an organisation secret;
- image pulls are platform-owned;
- a merged registry change reconciles automatically.

This task makes the app-side skills consume that model generically, through the Platform Repo
contract. It must not hard-code that platform. OIDC replaces the kubeconfig mode outright (no
backwards compatibility): the issued-kubeconfig path is removed, not kept beside OIDC.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions OIDC, Kubernetes structured authentication, k3s, kubectl, NetBird, GitHub organisation secrets, GHCR, Go
**Platforms/Environments:** GitHub-hosted runners, existing-k3s targets (public and private-mesh)
**Features/Integrations:** request-k3s-onboarding, repo-config, vault-loader, cicd-generate (k3s deploy.yml), deploy-k3s-app, Infrastructure spec template Environments table, Platform Repo contract, slot resolver
**Versions/Constraints:** builds on task 144 (deploys only in CI; OIDC tokens exist only inside a workflow); workflows need `permissions: id-token: write`; OIDC tokens are short-lived (minutes); platform-agnostic wording

## Design Decisions

- **OIDC replaces kubeconfig mode** (Rui, 2026-10-23): no backwards compatibility and no auth-mode
  switch. Removed, not deprecated:
  - the `secret/apps/<app-slug>/<machine-slug>/kubeconfig` Vault path and its load, rotate and
    `credential-file.sh` handling in vault-loader;
  - the `KUBECONFIG` Environment secret and its sync in repo-config's `sync-k3s-env.sh`;
  - the kubeconfig steps in the generated k3s `deploy.yml`, replaced by the OIDC token step;
  - existing-k3s `target-credentials` defaults to none in `installer/slots.go`: with the kubeconfig
    and the per-app mesh key gone (below), the target issues the app nothing to load.

  A platform that only issues kubeconfigs is out of scope, as a platform whose onboarding isn't
  PR-based already is. The CHANGELOG's action-required note tells app repositories to delete the
  `KUBECONFIG` secret and Vault path after their first OIDC deploy.
- **A per-environment shared mesh key replaces the per-app key** (Rui, 2026-10-23): one key per
  environment (machine slug), shared by every app deployed there; never one key for the whole
  platform. The platform issues and holds it; the app's Environments row names its secret
  (`Mesh Setup-Key Secret`), and the app neither stores nor syncs it. Removed: the
  `secret/apps/<app-slug>/<machine-slug>/mesh-setup-key` Vault path, `MESH_SETUP_KEY` loading and
  rotation in vault-loader, and `sync-k3s-env.sh --mesh-secret`.
- **Platform-owned pulls replace the app's pull credential** (Rui, 2026-10-23): the platform pulls
  images for the apps it hosts. Removed:
  - `REGISTRY_ORG` and the `secret/organizations/<org>/github-package-read` load in vault-loader's
    `access` scope;
  - `sync-k3s-env.sh --registry-org` and its `REGISTRY_USERNAME`/`REGISTRY_TOKEN` Environment secrets;
  - the generated `Create imagePullSecret` step (`smaqit:pull-secret` block) and its
    `render-k3s-deploy.sh` condition, and the `imagePullSecrets` hint in `deployment.yaml.template`.

  A private `Container Registry` Constraints row no longer drives any app-side step.

Still to be confirmed during assessment, once the downstream platform's task 062 publishes its
Platform Repo contract. Starting points:
- **Contract fields:**
  - the OIDC audience per environment;
  - where the cluster CA comes from;
  - the subject form the platform binds (repo + Environment);
  - the per-environment mesh key's secret name, and how the platform grants the app's repository
    access to it (for example, an organisation secret limited to selected repositories);
  - whether image pulls are platform-owned;
  - whether registry merges reconcile automatically.
- **No deploy-credential sync:** `repo-config` and `vault-loader` sync no deploy credential for an
  existing-k3s target. Only non-secret Environment variables remain, if any. (App runtime secrets
  are task 137's, which follows this task.)
- **No Environment creation either:** GitHub creates the machine-slug Environment on the first
  workflow run that references it (`environment: <machine>`). The downstream platform proved this
  live on 2026-10-22, in its task 062 probe repos. `repo-config`'s `gh api PUT environments/<machine>`
  is unnecessary for an existing-k3s target. Protection rules, such as a main-only branch policy,
  are the app repo's own optional choice.
- **Platform facts, from the downstream platform (live 2026-10-22/23):**
  - the contract fields the platform publishes are in its platform repository's
    `docs/onboarding-an-app.md` (auth mode `oidc`);
  - the actions are `.github/actions/netbird-connect` and `.github/actions/k3s-connect`;
  - the shared mesh secret is the organisation secret `NETBIRD_DEPLOY_KEY_<MACHINE>`;
  - a job must declare `permissions: id-token: write`;
  - OIDC tokens last about 5 minutes, and `k3s-connect`'s exec plugin refreshes them;
  - when a `run:` script holds any `${{ }}`, GitHub turns the whole script into one expression,
    capped at 21,000 characters. Generated workflows should pass values through `env:`.
- **Generated workflow:** `id-token: write`; mesh join first on a private-mesh target (with
  `wait-for-host`, plus `wait-for-port` once the platform ships it); then a token step (the
  platform's composite action when the contract names one, otherwise a generic `ACTIONS_ID_TOKEN`
  request); then kubectl.
- **request-k3s-onboarding:** the registry entry carries the app's repository when the contract
  asks for it. The merge gate is unchanged; the converge wait becomes "reconciled automatically"
  when the contract says so.

## Implementation Steps

1. Wait for the downstream platform's task 062 contract (or co-design it). Read task 144's
   outcome, which this builds on.
2. Update the Skill Definitions (`.smaqit/definitions/skills/`) for request-k3s-onboarding,
   repo-config, vault-loader, cicd-generate and deploy-k3s-app, then recompile through smaqit.L2.
3. Update the Infrastructure spec template's existing-k3s Environments table and the wiki
   (`docs/wiki/workflows/platform-repositories.md`, `deployment-slots.md`).
4. Bench or fixture coverage for oidc mode: generated workflow content, no secret sync, and the
   private-mesh order of steps.
5. Release with an action-required note for app repos.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The Platform Repo contract documents the OIDC fields (audience, CA source, bound subject, mesh key secret, pull ownership, auto-reconcile); it has no auth-mode switch
- [ ] No shipped skill, script, template or resolver default loads, rotates, syncs or uses an issued kubeconfig, a per-app mesh setup key or an app-held registry pull credential; existing-k3s `target-credentials` defaults to none; the generated `deploy.yml` creates no imagePullSecret
- [ ] The generated existing-k3s `deploy.yml` requests `id-token: write`, joins the mesh first on a private-mesh target with that environment's shared mesh key, authenticates with a short-lived OIDC token, and holds no other deploy secret
- [ ] request-k3s-onboarding writes the repository field when the contract requires it, and reports auto-reconcile instead of waiting on a manual dispatch
- [ ] The Infrastructure spec template, wiki and compendium describe only the OIDC model; a release carries an action-required note (delete the `KUBECONFIG` secret and Vault path)

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
| `.smaqit/definitions/skills/smaqit.infrastructure-request-k3s-onboarding.md` | Modify |
| `.smaqit/definitions/skills/smaqit.infrastructure-repo-config.md` | Modify |
| `.smaqit/definitions/skills/smaqit.infrastructure-cicd-generate.md` | Modify |
| `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s-app.md` | Modify |
| `skills/smaqit.infrastructure-*/` (compiled) | Modify |
| `skills/smaqit.infrastructure-vault-loader/` (kubeconfig load, rotate, `credential-file.sh` allowlist) | Modify |
| `skills/smaqit.infrastructure-repo-config/scripts/sync-k3s-env.sh` | Modify |
| `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`, `scripts/render-k3s-deploy.sh`, `installer/testdata/k3s-deploy/*` | Modify |
| `skills/smaqit.infrastructure-deploy-k3s-app/assets/deployment.yaml.template` (`imagePullSecrets` hint) | Modify |
| `installer/slots.go` (existing-k3s `target-credentials` default), installer tests | Modify |
| `templates/specs/infrastructure.template.md` | Modify |
| `docs/wiki/workflows/platform-repositories.md`, `deployment-slots.md`, `.smaqit/compendium.md` | Modify |
| `CHANGELOG.md` | Modify |

## Notes

- Parent: the downstream platform repository's task 061. Platform side: that repo's task 062.
  Origin: its issues #175, #177 and #178, assessed 2026-10-22.
- Depends on task 144 (CI-only deploys). Task 137 follows this task.
- The vault-loader skill definition may live only as a compiled skill. Check before step 2.
