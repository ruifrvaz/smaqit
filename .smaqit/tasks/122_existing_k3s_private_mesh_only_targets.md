---
status: PR Open
pr: 103
created: "2026-10-06"
mode: Assisted
started: "2026-10-16"
---

# existing-k3s: Support Private (Mesh-Only) Cluster Targets Alongside Public Ones

## Description

A downstream platform repository now runs two classes of k3s machine. Public machines expose
80/443 (Traefik via ServiceLB, HTTP-01 ClusterIssuer) and, for now, the k3s API on a public 6443.
Private machines have no public ingress at all: the k3s API and the apps' web traffic are reachable
only from peers of a WireGuard mesh (NetBird), app hostnames live in a private DNS zone served by the
mesh, and TLS is one DNS-01 wildcard certificate served as Traefik's default, so an app's Ingress
carries no `tls` section and requests no certificate. CI reaches a private cluster only after
joining the mesh with a per-app setup key (a reusable composite action), and the issued kubeconfig's
`server:` is the machine's mesh name. Even on the public class, apps behind the platform's SSO gate
must not ship their own Ingress (the platform owns it).

The existing-k3s skill family assumes every target is public. Found in a read-only audit of the
installed skills against that platform (2026-10-06), re-verified against `main` (`c014886`) on
2026-10-16 with current line references:

- `smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`: no mesh-join step before
  `kubectl`; waits for `certificate/__APP_SLUG__-tls` (:147); verifies by
  `curl https://${HOST}/health.json` from the GitHub-hosted runner (:162-171). All three fail
  against a private target; the last two also fail for SSO-gated apps on public targets.
- The same template triggers on `push: branches: [main]` as well as `workflow_dispatch` (:4-6).
  Combined with `post-merge-deploy.yml`'s marker dispatch, one marked merge deploys twice (observed
  downstream).
- The same template lints and applies only `deployment.yaml`, `service.yaml` and `ingress.yaml`
  (:130-137): a stateful app's `pvc.yaml` is never applied, and `ingress.yaml` is required even when
  the platform owns the Ingress.
- `smaqit.infrastructure-deploy-k3s-app/assets/ingress.yaml.template:21-27` always emits the
  cluster-issuer annotation and a `tls` block; no variant without `tls`, none for a gated slug.
  SKILL.md Steps 9-10 and its failure table assume HTTP-01 and public HTTPS verification.
- `deploy-k3s-app/scripts/namespace-guard.sh:76-84` reports PASS when the API is unreachable
  (any non-zero `kubectl auth can-i` counts as "not allowed").
- Verification "over the mesh" from CI is not possible on such a platform: the CI deployer peer's
  mesh group reaches only the API port, with no route to the ingress and no private DNS for app
  hostnames. Only in-cluster verification works from CI.
- `platform-prereqs` has no `existing-k3s` default filler, although the slot catalogue defines it
  as "network or mesh access".
- Spec contract: `smaqit.input-deployment`, `agents/smaqit.infrastructure.md` and
  `templates/specs/infrastructure.template.md` have no reachability class and make a
  ClusterIssuer mandatory; the elicitation question doesn't list existing-k3s.
- `smaqit.new-greenfield-project` assumes public DNS A records, the Certbot `domain-tls` skill and
  `https://<host>` checks from the agent's machine.
- `smaqit.infrastructure-repo-config` and `cicd-generate` treat `APP_HOST` as a public host the
  runner can curl; nothing syncs a per-app mesh setup key to the app's Environment.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Kubernetes (k3s), Traefik, cert-manager, NetBird, GitHub Actions, HashiCorp Vault, jq, Go
**Platforms/Environments:** None
**Features/Integrations:** existing-k3s provisioning mode, Infrastructure spec Environments table, smaqit slots plan, smaqit spec environments, k3s CI/CD template and renderer, Ingress templates, probe Job, namespace guard, greenfield orchestration, repo-config, vault-loader, platform-prereqs slot, deploy-verify
**Versions/Constraints:** breaking change: every existing-k3s environment must declare its reachability and Ingress mode; stays app- and platform-agnostic (no platform names in canonical text)

## Design Decisions

Resolved in `task.plan` on 2026-10-16 against `main` (`d0af642`).

- **Breaking, no opt-in.** Every existing-k3s environment must declare `reachability` (`public` |
  `private-mesh`) and `ingress_mode` (`own-tls` | `platform-default-tls` | `platform-owned`). No
  default, no fallback, no legacy rendering path. Released as MINOR with an action-required
  CHANGELOG note (v3.13.0 precedent).
- **Spec shape: a parsed `## Environments` table** in the Infrastructure spec, one row per
  environment: Environment, Machine Slug, Host, Reachability, Ingress Mode, Ingress Class,
  ClusterIssuer, Health Path, Mesh Setup-Key Secret, Mesh Wait-For Host. The `Machine Slug`,
  `Ingress Class` and `ClusterIssuer` Constraints rows move into it. New Constraints rows:
  `Mesh Join Action` (a pinned `uses:` ref), `Mesh Setup-Key Input` and `Mesh Wait-For Input` (the
  action's own input names), required only when any environment is `private-mesh`; `Probe Client`
  (`curl` | `wget`). Platform Repo, Registry File Path, Namespace Quota and Container Registry stay in
  Constraints. Values still come from the platform team, never invented.
- **Per-mode requirements:** ClusterIssuer only for `own-tls`; Ingress Class unless
  `platform-owned`; mesh fields only for `private-mesh`; one mesh-join action per app.
- **Gate: `smaqit slots plan`.** With `--mode=existing-k3s`, `resolveSlotPlan` parses the spec and
  refuses with `SLOTS-SPEC-*` diagnostics for the `--env` row, reported together with binding
  errors. `smaqit validate` checks the shape of any Environments table.
- **One parser, one interface:** `installer/infra_spec.go` parses the table;
  `smaqit spec environments --json` emits the rows plus the mesh and probe constraints. Fillers and
  the renderer consume that JSON and never parse Markdown.
- **The default fillers read the spec, not binding `params`.** Binding `params` remain only for
  project-bound fillers. Supersedes the 2026-10-11 params note below.
- **One `deploy.yml`, branched at run time.** A deterministic renderer
  (`cicd-generate/scripts/render-k3s-deploy.sh`, jq + awk) writes a `K3S_TARGETS` JSON map keyed by
  machine slug into the workflow; a "Resolve target" step exports the dispatched slug's fields, and
  steps branch with `if:`. The mesh-join step's `uses:` is rendered in (it can't be an expression)
  and the step is deleted when no environment is `private-mesh`. This is the one sanctioned renderer,
  not a general templating engine.
- **Ingress modes:** `own-tls` keeps `deployment/k3s/ingress.yaml` (tls, issuer annotation,
  Certificate wait — today's behaviour, no rename); `platform-default-tls` uses
  `deployment/k3s/ingress-default-tls.yaml` (no tls, no issuer, no wait); `platform-owned` renders,
  lints and applies no Ingress. Host and ClusterIssuer are stamped from the map at apply time.
- **Dispatch-only deploy trigger:** `workflow_dispatch` only; `post-merge-deploy.yml`'s marker
  dispatch is the sole automatic trigger.
- **Apply set follows the manifests:** `pvc.yaml` is linted and applied before the Deployment when
  present.
- **Verification:** the revision is always proven from the Deployment's image tag (`github.sha`),
  for every class. For `private-mesh` or `platform-owned`, an in-cluster probe Job
  (`deployment/k3s/probe-job.yaml`: restricted securityContext, `imagePullSecrets`,
  `backoffLimit: 0`, `ttlSecondsAfterFinished`, the app's own image running the declared Probe Client
  against the Service's cluster-DNS health path). Pass/fail comes from
  `kubectl wait --for=condition=complete` vs `condition=failed`; `kubectl logs` is diagnostics only,
  when `pods/log` is granted. External HTTPS checks (declared health path and root) run only for
  `public` targets that aren't `platform-owned`. Both keep the 12 × 5 s retry.
- **`platform-prereqs` default for existing-k3s:** the base `deploy-k3s-app` skill's `mesh-check`
  operation (`scripts/mesh-check.sh`): resolves the wait-for host and checks the API port over TCP,
  stops with instructions on failure, and is a no-op for `public` rows. In CI the mesh join is a
  workflow step. No new slot; `provisioning_mode` stays closed.
- **Mesh setup key:** Vault path `secret/apps/<app>/<machine>/mesh-setup-key` (field `value`),
  loaded by `load-credentials.sh` when `MESH_SETUP_KEY=1`, synced by
  `sync-k3s-env.sh --mesh-secret <name>`. Not added to `credential-file.sh` (CI is its only
  consumer).
- **`APP_HOST`:** generated k3s workflows read the host from the spec map, not `vars.APP_HOST`;
  `sync-k3s-env.sh --host` stays and is documented as unused by generated workflows.
- **`namespace-guard.sh` fails closed:** an API reachability check first; any connection or
  authentication error is a FAIL with a message distinct from "denied".
- **Out of scope:** a CronJob probe, verifying from CI over the mesh (not possible on such
  platforms), and a digest-pinned probe image (possible later as a third `Probe Client` value).

## Implementation Steps

*Phase A — spec contract and gate (everything else depends on this)*

1. `templates/specs/infrastructure.template.md`: the conditional `## Environments` table and the
   new Constraints rows; move `Machine Slug`, `Ingress Class`, `ClusterIssuer` out of Constraints.
2. `installer/infra_spec.go` (new): parse the Environments table from `specs/infrastructure/`
   bodies (reuse the section-slicing of `designReferenceLinks` in `installer/design.go`); enum and
   per-mode validation, duplicate environments, single mesh action. Tests in a new
   `installer/infra_spec_test.go`.
3. `installer/slots.go` `resolveSlotPlan`: `SLOTS-SPEC-*` diagnostics for `--mode=existing-k3s`;
   `installer/main.go` `cmdValidate`: Environments table shape checks.
4. `smaqit spec environments --json` subcommand and help text (depends on 2).
5. `agents/infrastructure.md` fills the Environments table; `smaqit.input-deployment` lists
   existing-k3s in its provisioning-mode question and points its existing-k3s entry at the gate.

*Phase B — CI/CD rendering (depends on A)*

6. `deploy.yml.k3s.template`: dispatch-only; `K3S_TARGETS` map and "Resolve target" step; rendered,
   conditional mesh-join step; host and issuer stamped from the map; PVC and per-mode Ingress
   lint/apply; Certificate wait only for `own-tls`; image-tag revision check; in-cluster probe or
   external HTTPS verify with retry.
7. `cicd-generate/scripts/render-k3s-deploy.sh` (new, parallel with 6) and the cicd-generate
   SKILL steps, token table, gotchas and scope sentence.
8. Golden files under `installer/testdata/k3s-deploy/` for all six reachability × mode pairs plus
   one mixed public/private pair, compared by a test with an `-update` flag; rewrite
   `TestK3sDeployTemplateContract` in `installer/cicd_templates_test.go`.

*Phase C — k3s deploy, verify, prereqs (depends on A; parallel with B)*

9. `deploy-k3s-app` assets: `ingress.yaml.template` stays `own-tls`; new
   `ingress-default-tls.yaml.template` and `probe-job.yaml.template`.
10. `deploy-k3s-app` SKILL Steps 1 and 5–10 per row (Ingress file or none, PVC, Certificate wait
    only for `own-tls`, verify through the verify slot), failure rows for a failed probe and an
    unreachable mesh; new `platform-prereqs=mesh-check` operation and `scripts/mesh-check.sh`.
11. `deploy-verify`: in-cluster mode chosen from the row (probe Job verdict by Job status, image-tag
    revision check, logs only if granted); HTTP mode unchanged for public targets.
12. `installer/slots.go`: existing-k3s `platform-prereqs` default `skillFiller(skillDeployK3s,
    "mesh-check")`; update `TestSlotPlanDefaultsMatchCurrentBehaviour` and the existing-k3s column
    of `docs/wiki/workflows/deployment-slots.md`.
13. `namespace-guard.sh`: reachability preflight, per-call stderr/exit capture, fail closed with a
    distinct message; a stub `kubectl` in `newScriptHarness` with tests for the guard and
    `mesh-check.sh`.

*Phase D — credentials, orchestration, docs (depends on A–C)*

14. `load-credentials.sh` (`MESH_SETUP_KEY=1`, `secret/apps/<app>/<machine>/mesh-setup-key`) and
    `sync-k3s-env.sh --mesh-secret <name>`; extend `TestLoadCredentials*` and `TestSyncK3sEnv*`.
15. `new-greenfield-project`, `feature-new`, `repo-config` SKILL text: per-class steps, no public
    DNS or `domain-tls` for private targets, verify target per class, setup-key sync, `APP_HOST` not
    needed by generated k3s workflows.
16. CHANGELOG (breaking change and migration: add the Environments table, move the three rows,
    regenerate `deploy.yml` with the renderer, commit `ingress-default-tls.yaml`/`probe-job.yaml`
    where needed, load the mesh setup key per private machine); README CLI table; compendium
    existing-k3s entries; `docs/wiki/workflows/platform-repositories.md` if affected.

Verification: `make -C installer test` and `make -C installer smoke-test`; a deliberate break per
new guard (gate accepting a missing field, guard passing on an unreachable API, renderer keeping
the mesh step in an all-public render), each caught by a test; a grep of `skills/`, `agents/` and
`templates/` for platform-specific product names returns nothing.

## Known Issues Triage
**Triaged:** 2026-10-16
**Tools searched:** k3s, Traefik, cert-manager, NetBird, GitHub Actions
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#13584 Goroutine leak: one go-cache janitor per dynamic-config apply, from TLS store rebuild in tlsmanager.UpdateConfigs](https://github.com/traefik/traefik/issues/13584) — `traefik/traefik` — opened 2026-07-29 — kind/bug/confirmed, area/tls, contributor/wanted (platform-side runtime; this task only renders manifests)
- [#9296 ingress-shim: certNeedsUpdate ignores ipAddresses, so IP SAN changes never reach existing Certificates](https://github.com/cert-manager/cert-manager/issues/9296) — `cert-manager/cert-manager` — opened 2026-09-07 — kind/bug (own-tls Ingress uses no IP SANs)
- [#9295 ingress-shim: alt-names and ip-sans annotations re-introduce duplicate SANs after #8978](https://github.com/cert-manager/cert-manager/issues/9295) — `cert-manager/cert-manager` — opened 2026-09-07 — kind/bug (template sets neither annotation)
- [#8733 Updates or removal of solver ingress class annotations on ingresses are not propagated to existing certificates](https://github.com/cert-manager/cert-manager/issues/8733) — `cert-manager/cert-manager` — opened 2026-04-18 — kind/bug (relevant to migrating an Ingress between own-tls and platform-default-tls: removing the issuer annotation may leave the old Certificate behind)

### Historical (Closed)
- [#2828 Netbird doesn't work with GitHub Actions and setup-key](https://github.com/netbirdio/netbird/issues/2828) — `netbirdio/netbird` — closed 2024-11-05
- [#14396 update.k3s.io intermittently serves TRAEFIK DEFAULT CERT, causing TLS verification failures](https://github.com/k3s-io/k3s/issues/14396) — `k3s-io/k3s` — closed 2026-07-13
- [#9145 ingress-shim accepts annotation values that the API server rejects, so the Certificate is never created](https://github.com/cert-manager/cert-manager/issues/9145) — `cert-manager/cert-manager` — closed (date not in compact view)

### Unresolvable Tools
- None

### Omitted Tools
- Kubernetes, HashiCorp Vault, jq, Go — excluded after relevance prioritization and the five-repository limit

### Search Warnings
- `actions/runner` open search — HTTP 403 (anonymous search rate limit)
- `actions/runner` closed search — HTTP 403 (anonymous search rate limit)

## Acceptance Criteria

- [x] A `private-mesh` + `platform-owned` spec renders a dispatch-only deploy workflow with a mesh join, no Ingress or Certificate, and in-cluster verification
- [x] An SSO-gated slug on a public target deploys without an app-owned Ingress and verifies in-cluster
- [x] An existing-k3s spec missing `reachability` or `ingress_mode` is rejected
- [x] A marked merge triggers exactly one deploy
- [x] `pvc.yaml` is applied when present
- [x] namespace-guard fails when the API is unreachable
- [x] Golden files cover every reachability × Ingress-mode combination and were regenerated deliberately
- [x] No platform-specific names in canonical skill text
- [x] `smaqit spec environments --json` emits the parsed Environments rows, and every default k3s filler and the renderer consume it
- [x] A `private-mesh` environment's mesh setup key loads from Vault and syncs into its machine-slug Environment without appearing in argv or output
- [x] The probe Job's verdict comes from Job status, so in-cluster verification works without a `pods/log` grant

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Added a parsed `## Environments` table to the Infrastructure spec template (`installer/infra_spec.go`), validated per reachability class and Ingress mode, and exposed it through `smaqit spec environments [--env] [--json]`
- Gated `smaqit slots plan --mode=existing-k3s` on the targeted row (`SLOTS-SPEC-*`, reported together with binding errors) and added the row to the plan JSON as `target`; `smaqit validate` checks the whole table
- Rewrote `deploy.yml.k3s.template` as one dispatch-only workflow branching at run time on a rendered per-machine `K3S_TARGETS` map, rendered by the new `render-k3s-deploy.sh` (jq + awk), with the mesh-join step kept only when some environment is private-mesh
- Added the `platform-default-tls` Ingress and probe Job templates, the `mesh-check` operation (`platform-prereqs` default for existing-k3s), deploy-verify's in-cluster mode, and the mesh setup-key load/rotate/sync path
- Made `namespace-guard.sh` fail closed: reachability preflight, and only an explicit "no" counts as denied
- Tests: parser and gate tables, seven golden renders with `-update`, a test executing the rendered steps for every combination (with and without `pvc.yaml`), a stub `kubectl` for the guard, mesh-check against a live local listener, probe Job and default-TLS Ingress through manifest-lint, mesh-key load and sync; smoke test covers the gate and the new command through the compiled binary

**Decisions made:**
- Spec shape, gate location, run-time variance in one workflow, and a Job-status probe verdict on the app's own image were chosen in `task.plan` Q&A (2026-10-16)
- The Certificate-wait step stays in the single workflow and is skipped by `if:` for non-own-tls machines; AC1 is met at run time, proven by the step-execution test rather than by the rendered file alone
- Revision is proven from the Deployment's image tag for every class; the external check only requests the declared health path and the root
- `own-tls` keeps `deployment/k3s/ingress.yaml`; Host, Ingress Class and ClusterIssuer became apply-time tokens so environments can differ
- repo-config and vault-loader take the row's values (machine slug, mesh secret name) as arguments rather than reading the JSON themselves; the fillers that branch on the row (cicd-generate, deploy-k3s-app, deploy-verify, mesh-check) consume it directly
- The deployment agent's `ci-push` path dispatches a workflow that has no push trigger; feature-new already resolves dispatch-only workflows through its dispatcher mode
- Released as MINOR with an action-required CHANGELOG note, matching the v3.13.0 precedent

**Blockers encountered:**
- None; issue triage was Advisory (cert-manager ingress-shim bugs, one Traefik TLS-store leak, `actions/runner` searches rate-limited)
- Caught during implementation: `getent` exit 2 under `pipefail` aborted mesh-check, an unescaped apostrophe broke a single-quoted skill description in `make prepare`, and actionlint flagged an unused loop variable in the rendered workflow

**Follow-up identified:**
- `.smaqit/definitions/skills/` entries for deploy-k3s-app, cicd-generate and deploy-verify still describe the pre-122 behaviour
- No live run yet against a private-mesh or SSO-gated machine; the probe Job needs `jobs` create/delete and same-Namespace NetworkPolicy from the platform
- Moving an environment off `own-tls` leaves cert-manager's old Certificate behind (cert-manager#8733); documented as a manual cleanup

## Files to Create / Modify

| File | Action |
|------|--------|
| templates/specs/infrastructure.template.md | Modify |
| installer/infra_spec.go, installer/infra_spec_test.go | Create |
| installer/slots.go, installer/slots_test.go, installer/main.go, installer/spec_test.go | Modify |
| installer/cicd_templates_test.go, installer/skill_scripts_test.go | Modify |
| installer/testdata/k3s-deploy/** | Create |
| skills/smaqit.infrastructure-cicd-generate/** (incl. new scripts/render-k3s-deploy.sh) | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/** (new ingress-default-tls/probe-job templates, scripts/mesh-check.sh) | Modify |
| skills/smaqit.infrastructure-deploy-verify/SKILL.md | Modify |
| skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh, skills/smaqit.infrastructure-repo-config/** | Modify |
| skills/smaqit.input-deployment/SKILL.md, agents/infrastructure.md | Modify |
| skills/smaqit.new-greenfield-project/SKILL.md, skills/smaqit.feature-new/SKILL.md | Modify |
| docs/wiki/workflows/deployment-slots.md, docs/wiki/workflows/platform-repositories.md | Modify |
| README.md, .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Source: a downstream platform's consistency audit, 2026-10-06 (its `.smaqit/reports/consistency-audit-2026-10-06.md`, B8-B12).
- **Re-scoped by task 125 (2026-10-11):** the per-environment reachability class and Ingress modes become slot parameters (`verify`, `tls`, `ci-generate`, `app-deploy`) instead of new branches on the closed `provisioning_mode` enum. Rebase this task on 125.
- **Refreshed after task 125 implementation (2026-10-11):** slot names above are as shipped. A private-mesh target adds a `platform-prereqs` filler (mesh join) and parameterizes `verify`/`ci-generate`/`app-deploy` through binding `params`; `tls` is already `none` for `existing-k3s`. Resolution and refusals come from `smaqit slots plan`.
- **Refreshed 2026-10-16** from a downstream app's planning: the first app onto a mesh-only, SSO-gated machine, which is blocked on this task. User direction: a breaking change where the default fillers read the spec, and no new slot. This supersedes the binding-`params` approach in the previous note (the class and mode come from the Infrastructure spec; `provisioning_mode` stays closed). Downstream facts: the CI deployer peer reaches TCP 6443 only; the mesh-join action takes a setup key and a wait-for host and pins its client version; the platform's wildcard certificate is the default TLS store and also covers the bare mesh domain.
- **Planned 2026-10-16 (`task.plan`):** discovery found no per-environment spec structure, no Go spec-body parsing for infrastructure fields, no conditional rendering or golden files, and one `deploy.yml` serving several environments. Resolved: parsed Environments table, `smaqit slots plan` gate, run-time branch on a rendered per-machine map, probe verdict from Job status on the app's own image. `agents/smaqit.infrastructure.md` above means `agents/infrastructure.md`.
