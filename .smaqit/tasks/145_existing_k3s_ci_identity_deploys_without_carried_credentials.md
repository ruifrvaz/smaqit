---
status: In Progress
created: "2026-10-22"
started: "2026-10-25"
mode: Assisted
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
  by an exact body comparison; the OIDC change is then made in the definitions and applied to the
  skills. Scripts, assets, Go and tests are edited directly.
- **Definitions are applied by hand, not compiled by `smaqit.L2`** (Rui, 2026-10-25): L2's skill
  rules conflict with this repo's conventions (they forbid the "Use when…" trigger-form description,
  carry no `smaqit-slots`, emit a fixed section set, and split bodies over 400 lines). Each
  definition holds the skill's frontmatter, its provenance, and its body; the `SKILL.md` body is
  kept identical to it. No L2 run, no `smaqit-adk` registry entry. A smaqit-adk task should let L2
  honour a product's description, metadata and section conventions so later tasks can recompile.
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

- **Reframed: abstract base, concrete platform fillers** (Rui, 2026-10-25, after the downstream
  platform's session proposed it): smaqit's existing-k3s slots are contracts with generic defaults; a
  platform publishes its own fillers (the downstream platform's hub task 004) that app repositories
  bind in `.smaqit/slots.yml`, replacing a default with nothing else to change.
- **Contracts live in the resolver** (Rui, 2026-10-25): each slot in `installer/slots.go` carries its
  inputs, handoff, completion signal and per-mode notes; `smaqit slots contract [--slot=<slot>]
  [--mode=<mode>]` prints them; the Deployment agent and filler authors read that, and the
  deployment-slots wiki's catalogue table is the same text (a test keeps them equal).
- **app-deploy owns trigger and monitor under a CI-only mode** (Rui, 2026-10-25): the deploy itself
  runs in the workflow `ci-generate` produced, but the agent reads the `app-deploy` filler to trigger
  the run (dispatch inputs, or the merged PR's dispatcher), select it by commit, wait, and hand its
  run ID to `verify`. smaqit's default (deploy-k3s-app `deploy`) is today's `gh` procedure; a
  platform filler may replace it. VM modes keep the agent's built-in monitoring.
- **Exposure is declared by the platform** (Rui, 2026-10-25): an `Allowed Ingress Modes` Constraints
  row (comma list, default all) refuses Environments rows using a disallowed mode, and
  `manifest-lint.sh` refuses any Service whose type isn't ClusterIP, so a forbidden manifest fails
  before apply rather than at the platform's admission.
- **The namespace guard stays the reconcile check** (Rui, 2026-10-25): no platform reconcile-workflow
  row; a merged onboarding request is reported, and the first deploy's guard answers 403 until the
  platform reconciles.
- **`tls`, `replace` and `teardown` stay refused for existing-k3s bindings** (task 144): no path runs
  them in CI, and the platform declares them none.

## Implementation Steps

*Phase A: resync definitions (no behaviour change)*
1. Rewrite `.smaqit/definitions/skills/` for request-k3s-onboarding, repo-config, cicd-generate and
   deploy-k3s-app to match today's compiled `SKILL.md`; create one for vault-loader.
2. Each definition holds `## Frontmatter`, `## Provenance` and the skill body verbatim; an exact
   comparison of the definition body with `SKILL.md` proves the resync.

*Phase B: OIDC in the definitions, applied to the skills*
3. cicd-generate: the connect rows, deploy-job permissions, mesh join keyed on the row's secret,
   no kubeconfig or pull-secret steps, the cleanup of `$KUBECONFIG`.
4. deploy-k3s-app: the deploy sequence starts with mesh join (when named) and the connect action;
   no `KUBECONFIG` secret, no `imagePullSecret`; failure rows for a missing `id-token` permission,
   an unreconciled binding (Forbidden), and an unshared organisation mesh secret (empty value).
5. vault-loader, repo-config, request-k3s-onboarding as decided above. Apply all five definitions to
   their `SKILL.md` and re-run the body comparison.

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

*Phase E: reframe — contracts, app-deploy monitor, exposure (added 2026-10-25)*
14. `installer/slots.go`: per-slot contract fields and per-mode notes; `smaqit slots contract`
    (text and `--json`) with help text; a test that the wiki catalogue table equals the contracts.
15. `installer/infra_spec.go`: the `Allowed Ingress Modes` row (`allowedIngressModes` in `--json`),
    validation of each row's mode against it; spec template and `agents/infrastructure.md`.
16. `manifest-lint.sh`: refuse a Service whose `spec.type` is set to anything but ClusterIP.
17. deploy-k3s-app `deploy`: the trigger-and-monitor operation (dispatch with `--ref`, select by
    event, commit and creation time, stop on ambiguity, watch, output the run ID); the Deployment
    agent delegates that part to the plan's `app-deploy` filler under a CI-only mode.
18. Definitions resync, wiki, compendium, CHANGELOG; tests and deliberate breaks for each rule.

## Known Issues Triage
**Triaged:** 2026-10-25
**Tools searched:** k3s (`k3s-io/k3s`), kubectl (`kubernetes/kubectl`), NetBird (`netbirdio/netbird`)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#5634 Add a setting to bypass "JWT Authentication" without `--disable-ssh-auth`](https://github.com/netbirdio/netbird/issues/5634) — `netbirdio/netbird` — opened 2026-03-19 — feature-request. Not this task's dimension (SSH auth); listed because the search matched. The other two open NetBird results (#3921, #496) are feature requests for custom devices, likewise unrelated.

### Historical (Closed)
- [#2828 Netbird doesn't work with GitHub Actions and setup-key](https://github.com/netbirdio/netbird/issues/2828) — `netbirdio/netbird` — closed 2024-11-05. The mesh join is the platform's pinned action (NetBird 0.79.0); relevant only if a platform pins an old client.
- [#1109 client-go credential plugins ignores 'expirationTimestamp' field](https://github.com/kubernetes/kubectl/issues/1109) — `kubernetes/kubectl` — closed 2021-09-16. The connect action's exec plugin relies on `expirationTimestamp` to refresh the ~5 min token; a current kubectl honours it (live-proven at 685 s by the platform).
- [#1492 Process leak when unable to find exec-credential binary](https://github.com/kubernetes/kubectl/issues/1492) — `kubernetes/kubectl` — closed 2023-12-22. The plugin command is `bash` with a script inside the action; the namespace guard's many `can-i` calls each run it.
- [#5278 Getting Unauthorized error when connecting to k3s cluster using OIDC](https://github.com/k3s-io/k3s/issues/5278) — `k3s-io/k3s` — closed 2022-04-15. A wrong audience or Environment claim answers 401 (Unauthorized), unlike an unbound identity (403 Forbidden); the skill's failure rows must tell them apart.

### Unresolvable Tools
- GitHub Actions OIDC — helper resolved `unfunco/terraform-aws-oidc-github` (an unrelated Terraform module)
- Kubernetes structured authentication — helper resolved an unrelated proxy repository
- GitHub organisation secrets — helper resolved an unrelated repository
- GHCR — helper resolved an unrelated crawler repository

### Omitted Tools
- Go — only the standard library and `go test` are used; no feature dimension to search

### Search Warnings
- Categorization limitation: four technologies are hosted GitHub or Kubernetes features without a usable issue tracker from the resolver; no search ran for them.

## Acceptance Criteria

- [ ] The five skill definitions match their skills before the change (vault-loader's created), every skill change in this task is made in its definition and applied to `SKILL.md`, and each definition body equals its `SKILL.md` body at the end
- [ ] The spec carries `Cluster Connect Action`/`Machine Input`/`Namespace Input` rows, required for existing-k3s; a `public` row may name a mesh secret; the `Container Registry` row is gone
- [ ] The generated `deploy.yml` grants `id-token: write` and `contents: read` to `deploy`, joins the mesh first whenever the row names a mesh secret, connects through the named action with the machine slug and `app-<slug>`, removes `$KUBECONFIG` always, and holds no `KUBECONFIG`, registry or per-app mesh secret
- [ ] No shipped skill, script, template or resolver default loads, rotates, syncs or uses an issued kubeconfig, a per-app mesh setup key or an app-held registry pull credential; existing-k3s `target-credentials` and `environment-config` default to none; `sync-k3s-env.sh` is gone
- [ ] request-k3s-onboarding prints and checks this repository's identity for the entry, and reports a merged request without any kubeconfig hand-off
- [ ] `namespace-guard.sh` retries the API check for a bounded window and still fails closed
- [ ] Docs describe only the OIDC model; `make -C installer test` and `smoke-test` pass; CHANGELOG has an action-required note; MINOR release via the task PR
- [ ] `smaqit slots contract` prints every slot's inputs, handoff, completion and existing-k3s notes from `installer/slots.go`, and the wiki catalogue matches it (tested)
- [ ] Under a CI-only mode the Deployment agent runs the plan's `app-deploy` filler to trigger and monitor the deploy run; deploy-k3s-app's default implements the `gh` procedure and stops on an ambiguous run selection
- [ ] An `Allowed Ingress Modes` row refuses Environments rows using a mode it doesn't list; absent, every mode is allowed
- [ ] `manifest-lint.sh` refuses a non-ClusterIP Service
- [ ] A platform can replace `platform-onboarding`, `artifact-build`, `ci-generate` and `app-deploy` by binding its own fillers, with no other change, as the contracts state

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
| `skills/smaqit.infrastructure-{vault-loader,repo-config,request-k3s-onboarding,cicd-generate,deploy-k3s-app}/SKILL.md` | Modify (applied from definitions) |
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
