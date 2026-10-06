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
installed skills against that platform (2026-10-06):

- `smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`: no mesh-join step before
  `kubectl` (:50-70); verifies by `curl https://${HOST}/health` from the GitHub-hosted runner
  (:141-156); waits for `certificate/__APP_SLUG__-tls` (:133-135). All three fail against a
  private target; the last two also fail for SSO-gated apps on public targets.
- `smaqit.infrastructure-deploy-k3s-app/assets/ingress.yaml.template:18-25` always emits the
  cluster-issuer annotation and a `tls` block; no variant without `tls`, none for a gated slug.
  SKILL.md Steps 9-10 and its failure table assume HTTP-01 and public HTTPS verification.
- `deploy-k3s-app/scripts/namespace-guard.sh:75-92` reports PASS when the API is unreachable
  (any non-zero `kubectl auth can-i` counts as "not allowed").
- Spec contract: `smaqit.input-deployment` (:39, :46), `agents/smaqit.infrastructure.md:152-156` and
  `templates/specs/infrastructure.template.md:131` have no reachability class and make a
  ClusterIssuer mandatory; the elicitation question doesn't list existing-k3s.
- `smaqit.new-greenfield-project` (:44-45, :177-183, :211-225) assumes public DNS A records, the
  Certbot `domain-tls` skill and `https://<host>` checks from the agent's machine.
- `smaqit.infrastructure-repo-config` and `cicd-generate:58` treat `APP_HOST` as a public host the
  runner can curl; nothing syncs a per-app mesh setup key to the app's Environment.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Kubernetes (k3s), Traefik, cert-manager, NetBird, GitHub Actions, HashiCorp Vault
**Platforms/Environments:** None
**Features/Integrations:** existing-k3s provisioning mode, k3s CI/CD template, Ingress template, namespace guard, infrastructure spec template, greenfield orchestration, repo-config
**Versions/Constraints:** stays app- and platform-agnostic (no platform names in canonical text), public targets keep working unchanged

## Design Decisions

- **A reachability class in the Infrastructure spec per environment:** `public` or `private-mesh`
  (with the mesh-join action reference, the setup-key secret name, and the private domain). It drives
  the template variant; the ClusterIssuer becomes optional (absent for a wildcard-default platform).
- **Ingress modes:** `own-tls` (today), `platform-default-tls` (no `tls`, no issuer annotation),
  `platform-owned` (no Ingress at all, e.g. an SSO-gated slug).
- **Verification from inside the cluster or over the mesh**, never assumed from the runner's public
  network: rollout status, endpoints, and an HTTP probe through the mesh or `kubectl`.
- **namespace-guard fails closed** on any connection error.

## Implementation Steps

1. Spec template and `input-deployment`: reachability class, optional ClusterIssuer, Ingress mode.
2. `deploy.yml.k3s.template`: mesh-join step for `private-mesh`; Certificate wait only for `own-tls`;
   verification per class; retry.
3. `deploy-k3s-app`: Ingress variants, Steps 9-10 per class, failure table.
4. `namespace-guard.sh`: distinguish "denied" from "unreachable".
5. `new-greenfield-project`, `repo-config`, `cicd-generate`: per-class steps (no public DNS or
   domain-tls for private targets; setup-key secret sync).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] An Infrastructure spec can declare a private-mesh existing-k3s environment, and the skills render a working deploy workflow for it (mesh join, no per-app Certificate, in-mesh verification)
- [ ] An SSO-gated slug on a public target deploys without an app-owned Ingress
- [ ] namespace-guard fails when the API is unreachable
- [ ] Public targets render exactly as before (golden-file comparison)
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
| skills/smaqit.input-deployment/SKILL.md, agents/smaqit.infrastructure.md, templates/specs/infrastructure.template.md | Modify |
| skills/smaqit.new-greenfield-project/SKILL.md, skills/smaqit.infrastructure-repo-config/SKILL.md | Modify |

## Notes

- Source: a downstream platform's consistency audit, 2026-10-06 (its `.smaqit/reports/consistency-audit-2026-10-06.md`, B8-B12).
