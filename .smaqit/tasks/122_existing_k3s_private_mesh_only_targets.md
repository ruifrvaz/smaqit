---
status: Not Started
created: "2026-10-06"
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
**Technologies:** Kubernetes (k3s), Traefik, cert-manager, NetBird, GitHub Actions, HashiCorp Vault
**Platforms/Environments:** None
**Features/Integrations:** existing-k3s provisioning mode, k3s CI/CD template, Ingress template, namespace guard, infrastructure spec template, greenfield orchestration, repo-config, platform-prereqs slot, deploy-verify
**Versions/Constraints:** breaking change: every existing-k3s environment must declare its reachability and Ingress mode; stays app- and platform-agnostic (no platform names in canonical text)

## Design Decisions

- **Breaking, no opt-in.** Every existing-k3s environment in the Infrastructure spec must declare
  `reachability` (`public` | `private-mesh`) and `ingress_mode` (`own-tls` |
  `platform-default-tls` | `platform-owned`). Planning fails when either is missing: no default, no
  fallback, no legacy rendering path.
- **`private-mesh` also declares** the mesh-join action reference (a pinned `uses:` string), the
  setup-key secret name, and the host to wait for (the API's mesh name). Fillers render these
  verbatim, so no platform name appears in canonical text.
- **The default fillers read the spec, not binding `params`.** `cicd-generate`, `deploy-k3s-app`
  and `deploy-verify` branch on the declared class and mode. Binding `params` remain only for
  project-bound fillers. Supersedes the 2026-10-11 params note below.
- **Ingress modes:** `own-tls` renders an Ingress with `tls`, the issuer annotation and a
  Certificate wait (today's behaviour); `platform-default-tls` renders an Ingress with no `tls`, no
  issuer and no wait; `platform-owned` renders, lints and applies no Ingress at all.
- **Verification inside the cluster** for `private-mesh` or `platform-owned`: rollout status, then
  a short-lived probe Job (restricted securityContext, `imagePullSecrets`, `backoffLimit: 0`,
  `ttlSecondsAfterFinished`) requesting the Service's health path over cluster DNS. Its result is
  read through `kubectl logs`, then the Job is deleted. The revision is proven from the
  Deployment's image tag (the full commit SHA), not from a `/health.json` field; the health check
  only has to succeed on the health path the spec declares. External HTTPS checks run only for
  public, ungated hosts.
- **Dispatch-only deploy trigger:** the generated `deploy.yml` has `workflow_dispatch` only;
  `post-merge-deploy.yml`'s marker dispatch is the sole automatic trigger, for every k3s target.
- **Apply set follows the manifests:** `pvc.yaml` is applied before the Deployment when present;
  `ingress.yaml` only when the mode isn't `platform-owned`.
- **`platform-prereqs` default for `private-mesh`:** checks that the API's mesh name resolves and the
  API port is reachable from the local machine, and stops with instructions if not. In CI the mesh
  join is a workflow step.
- **`namespace-guard.sh` fails closed:** any connection or authentication error is a FAIL, with a
  message distinct from "denied".

## Implementation Steps

1. Spec contract: `templates/specs/infrastructure.template.md`, `agents/smaqit.infrastructure.md`
   and `smaqit.input-deployment` gain the required `reachability`/`ingress_mode` fields, the
   `private-mesh` fields and the health path; the ClusterIssuer becomes required only for `own-tls`;
   the elicitation question lists existing-k3s.
2. A deterministic rejection of an existing-k3s spec missing `reachability` or `ingress_mode`.
3. `deploy.yml.k3s.template`: dispatch-only trigger; mesh-join step for `private-mesh`; apply set
   from the manifests present; Certificate wait only for `own-tls`; verification per class (in-cluster
   probe Job, image-tag revision check) with retry.
4. `deploy-k3s-app`: Ingress variants (including none), Steps 9-10 per class, failure table, PVC
   handling.
5. `deploy-verify`: an in-cluster verification mode.
6. `installer/slots.go`: the `platform-prereqs` default filler for `private-mesh`; update the
   deployment-slots wiki table.
7. `namespace-guard.sh`: distinguish "denied" from "unreachable", failing closed; add a test.
8. `new-greenfield-project`, `repo-config`, `cicd-generate`: per-class steps (no public DNS or
   domain-tls for private targets; setup-key secret sync into the app's Environment).
9. Regenerate golden files deliberately for every reachability × Ingress-mode combination; the
   CHANGELOG records the breaking change and the migration (declare the two fields).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] A `private-mesh` + `platform-owned` spec renders a dispatch-only deploy workflow with a mesh join, no Ingress or Certificate, and in-cluster verification
- [ ] An SSO-gated slug on a public target deploys without an app-owned Ingress and verifies in-cluster
- [ ] An existing-k3s spec missing `reachability` or `ingress_mode` is rejected
- [ ] A marked merge triggers exactly one deploy
- [ ] `pvc.yaml` is applied when present
- [ ] namespace-guard fails when the API is unreachable
- [ ] Golden files cover every reachability × Ingress-mode combination and were regenerated deliberately
- [ ] No platform-specific names in canonical skill text

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
| skills/smaqit.infrastructure-cicd-generate/** | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/** | Modify |
| skills/smaqit.infrastructure-deploy-verify/** | Modify |
| skills/smaqit.input-deployment/SKILL.md, agents/smaqit.infrastructure.md, templates/specs/infrastructure.template.md | Modify |
| skills/smaqit.new-greenfield-project/SKILL.md, skills/smaqit.infrastructure-repo-config/SKILL.md | Modify |
| installer/slots.go, installer/slots_test.go | Modify |
| docs/wiki/workflows/deployment-slots.md | Modify |
| CHANGELOG.md | Modify |

## Notes

- Source: a downstream platform's consistency audit, 2026-10-06 (its `.smaqit/reports/consistency-audit-2026-10-06.md`, B8-B12).
- **Re-scoped by task 125 (2026-10-11):** the per-environment reachability class and Ingress modes become slot parameters (`verify`, `tls`, `ci-generate`, `app-deploy`) instead of new branches on the closed `provisioning_mode` enum. Rebase this task on 125.
- **Refreshed after task 125 implementation (2026-10-11):** slot names above are as shipped. A private-mesh target adds a `platform-prereqs` filler (mesh join) and parameterizes `verify`/`ci-generate`/`app-deploy` through binding `params`; `tls` is already `none` for `existing-k3s`. Resolution and refusals come from `smaqit slots plan`.
- **Refreshed 2026-10-16** from a downstream app's planning: the first app onto a mesh-only, SSO-gated machine, which is blocked on this task. User direction: a breaking change where the default fillers read the spec, and no new slot. This supersedes the binding-`params` approach in the previous note (the class and mode come from the Infrastructure spec; `provisioning_mode` stays closed). Downstream facts: the CI deployer peer reaches TCP 6443 only; the mesh-join action takes a setup key and a wait-for host and pins its client version; the platform's wildcard certificate is the default TLS store and also covers the bare mesh domain.
