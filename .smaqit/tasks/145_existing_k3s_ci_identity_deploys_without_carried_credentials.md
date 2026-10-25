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

- **The connect step is the platform's action** (Rui, 2026-10-25): new spec Constraints rows name a
  pinned `Cluster Connect Action` (`owner/repo/path@ref`) and its `Cluster Connect Machine Input` and
  `Cluster Connect Namespace Input`, mirroring the mesh-join rows; the generated step passes the
  machine slug and `app-<app-slug>`. The platform owns the API URL, the cluster CA (bundled per tag,
  so a CA change is a tag bump), the audience, and token refresh (exec plugin). smaqit never handles
  a token and adds no API/CA/audience fields.
- **Mesh columns decide the mesh join** (Rui, 2026-10-25): `Reachability` keeps meaning the app's
  reachability (HTTPS vs in-cluster verify). The mesh join runs whenever the row names a
  `Mesh Setup-Key Secret`: required for `private-mesh`, now allowed for `public` (the downstream
  platform's API is mesh-only even on machines whose apps are public).
- **Definitions are resynced first** (Rui, 2026-10-25): the five skills' definitions under
  `.smaqit/definitions/skills/` are brought back in line with today's compiled skills (vault-loader
  gets one; repo-config, cicd-generate, deploy-k3s-app and request-k3s-onboarding are stale), proven
  by recompiling through `smaqit.L2` with no behaviour change; the OIDC change is then made in the
  definitions and recompiled. Scripts, assets, Go and tests are edited directly (L2 compiles
  `SKILL.md` only).
- **Start now** (Rui, 2026-10-25): the platform's task 062 is completed (released v0.32.0/v0.32.1).
  Its `wait-for-port` (its task 057) is not, so `namespace-guard.sh` retries its API-reachability
  check for a bounded window before failing closed.
- **Deploy job permissions:** `id-token: write` and `contents: read` (a job-level `permissions:`
  block sets every unlisted scope to none, which would break checkout).
- **The connect action's kubeconfig must pin the context namespace** (`namespace-guard.sh` refuses a
  context without one); the contract field is the action's namespace input. The always-run cleanup
  step removes whatever `$KUBECONFIG` the action exported.
- **Nothing is synced into the Environment:** existing-k3s `target-credentials` and
  `environment-config` default to none; GitHub creates the machine-slug Environment on the first run
  that references it (proven live by the downstream platform); protection rules are the app repo's
  own choice. `sync-k3s-env.sh` is removed, as is repo-config's k3s branch.
- **vault-loader for existing-k3s loads only `platform-repo`** (the onboarding PR's token): no
  kubeconfig, no mesh key, no `organizations/` registry credential, and no `github` PAT, whose only
  k3s consumer was the Environment sync. `MACHINE_SLUG` and `CREDENTIAL_SCOPE` lose their k3s
  meaning. VM-mode and platform-repository paths (`target`, `register`, `sync`, `sync-machine`)
  are untouched.
- **Onboarding stays declarative:** request-k3s-onboarding still never templates `entry_content`,
  but prints this repository's `owner/repo` and numeric ID (`gh repo view --json nameWithOwner,id`)
  for the operator to include when the platform's registry asks for them, and checks the entry names
  this repository. After the merge it reports the request merged; the first deploy's namespace guard
  is the convergence check (a not-yet-reconciled binding answers Forbidden). No kubeconfig hand-off
  or `target-credentials` gate remains.
- **`Container Registry` row removed** from the parser, template and docs: with platform-owned pulls
  it drives nothing.

## Implementation Steps

*Phase A: resync definitions (no behaviour change)*
1. Rewrite `.smaqit/definitions/skills/` for request-k3s-onboarding, repo-config, cicd-generate and
   deploy-k3s-app to match today's compiled `SKILL.md`; create one for vault-loader.
2. Recompile each through `smaqit.L2` and diff against the current `SKILL.md`: only wording noise may
   differ. Fix the definition, not the output, until it does.

*Phase B: OIDC in the definitions, recompiled*
3. cicd-generate: the connect rows, deploy-job permissions, mesh join keyed on the row's secret,
   no kubeconfig or pull-secret steps, the cleanup of `$KUBECONFIG`.
4. deploy-k3s-app: the deploy sequence starts with mesh join (when named) and the connect action;
   no `KUBECONFIG` secret, no `imagePullSecret`; failure rows for a missing `id-token` permission,
   an unreconciled binding (Forbidden), and an unshared organisation mesh secret (empty value).
5. vault-loader, repo-config, request-k3s-onboarding as decided above. Recompile all five.

*Phase C: Go, scripts, templates*
6. `installer/infra_spec.go`: the three `Cluster Connect` Constraints rows (required for every
   existing-k3s spec, `actionRefPattern`/`actionInputPattern`), mesh columns allowed on `public`
   rows, `Container Registry` removed; `--json` and text output.
7. `installer/slots.go`: existing-k3s `target-credentials` and `environment-config` default to none.
8. `deploy.yml.k3s.template` and `render-k3s-deploy.sh`: `permissions:` on `deploy`, the
   `smaqit:cluster-connect` step after kubectl install, mesh join `if:` on the secret name, pull
   secret and kubeconfig write removed, cleanup of `$KUBECONFIG`; regenerate the golden files.
9. `namespace-guard.sh`: bounded retry of the API-reachability check.
10. Scripts: remove `sync-k3s-env.sh`; remove the k3s kubeconfig, mesh-key and `organizations/`
    branches from `load-credentials.sh`, `rotate-credential.sh` and `credential-file.sh`.

*Phase D: tests and docs*
11. Tests: infra_spec, slots, render contract and refusals, guard retry, credential scripts (keep
    the kv-v2-once assertion), golden files; remove the pull-secret, mesh-key and sync tests.
    Deliberate breaks for the new rules.
12. Docs: spec template and `agents/infrastructure.md`, deployment-slots wiki, compendium,
    project-new, input-deployment, feature-new, README if needed.
13. CHANGELOG with an action-required note (add the `Cluster Connect` rows; delete the `KUBECONFIG`
    secret, the Vault kubeconfig/mesh-key paths and the `organizations/` registry credential; remove
    the `Container Registry` row); `make -C installer test` and `smoke-test`; MINOR release.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The five skill definitions match the compiled skills before the change (L2 recompile reproduces them), and every skill change in this task is made in its definition and recompiled
- [ ] The spec carries `Cluster Connect Action`/`Machine Input`/`Namespace Input` rows, required for existing-k3s; a `public` row may name a mesh secret; the `Container Registry` row is gone
- [ ] The generated `deploy.yml` grants `id-token: write` and `contents: read` to `deploy`, joins the mesh first whenever the row names a mesh secret, connects through the named action with the machine slug and `app-<slug>`, removes `$KUBECONFIG` always, and holds no `KUBECONFIG`, registry or per-app mesh secret
- [ ] No shipped skill, script, template or resolver default loads, rotates, syncs or uses an issued kubeconfig, a per-app mesh setup key or an app-held registry pull credential; existing-k3s `target-credentials` and `environment-config` default to none; `sync-k3s-env.sh` is gone
- [ ] request-k3s-onboarding prints and checks this repository's identity for the entry, and reports a merged request without any kubeconfig hand-off
- [ ] `namespace-guard.sh` retries the API check for a bounded window and still fails closed
- [ ] Docs describe only the OIDC model; `make -C installer test` and `smoke-test` pass; CHANGELOG has an action-required note; MINOR release via the task PR

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
| `.smaqit/definitions/skills/smaqit.infrastructure-{request-k3s-onboarding,repo-config,cicd-generate,deploy-k3s-app}.md` | Modify |
| `.smaqit/definitions/skills/smaqit.infrastructure-vault-loader.md` | Create |
| `skills/smaqit.infrastructure-{vault-loader,repo-config,request-k3s-onboarding,cicd-generate,deploy-k3s-app}/SKILL.md` | Recompile |
| `skills/smaqit.infrastructure-vault-loader/scripts/{load-credentials,rotate-credential,credential-file}.sh` | Modify |
| `skills/smaqit.infrastructure-repo-config/scripts/sync-k3s-env.sh` | Delete |
| `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`, `scripts/render-k3s-deploy.sh` | Modify |
| `skills/smaqit.infrastructure-deploy-k3s-app/scripts/namespace-guard.sh`, `assets/deployment.yaml.template` | Modify |
| `installer/infra_spec.go`, `installer/slots.go` and their tests | Modify |
| `installer/k3s_render_test.go`, `installer/k3s_scripts_test.go`, `installer/skill_scripts_test.go`, `installer/testdata/k3s-deploy/*` | Modify |
| `templates/specs/infrastructure.template.md`, `agents/infrastructure.md` | Modify |
| `skills/smaqit.{project-new,input-deployment,feature-new}/SKILL.md` | Modify |
| `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`, `CHANGELOG.md` | Modify |

## Notes

- Parent: the downstream platform repository's task 061. Platform side: that repo's task 062.
  Origin: its issues #175, #177 and #178, assessed 2026-10-22.
- Depends on task 144 (CI-only deploys, released v3.20.0). Task 137 follows this task; with
  `environment-config` none for k3s, 137's runtime-secret sync reintroduces a k3s Environment sync.
- The platform's task 062 completed 2026-10-25 (v0.32.0/v0.32.1); its `wait-for-port` (task 057) is
  not done yet.
- Planned with `task.plan` on 2026-10-23/25: three discovery agents mapped the platform contract, the
  credential plumbing, and the workflow generator.
- vault-loader has no definition today; Phase A creates one.
