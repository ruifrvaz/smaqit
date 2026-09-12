---
name: smaqit.infrastructure-onboard-k3s-app
description: Use when onboarding an application as a tenant of an existing, self-hosted, single-server k3s cluster — per-app Namespace, least-privilege RBAC, Pod Security Admission enforcement, network isolation, resource limits, and scoped kubeconfig issuance, driven by a git-committed registry file plus a converge GitHub Actions workflow. This is app onboarding (granting a tenant slot on a cluster that already exists), not cluster provisioning and not the onboarded app's own deployment. Invoke directly when a project's Infrastructure spec targets a pre-existing k3s cluster — not yet wired into `smaqit.new-greenfield-project`/`smaqit.feature-new`'s automated phase flow.
metadata:
  version: "1.0.0"
  validated: "2026-09-11"
  validated-stack: "k3s (self-hosted, single server per machine), GitHub Actions (workflow_dispatch, Environment protection rules), kubectl"
---

# Onboard an App onto a k3s Cluster

Hardened across four rounds of real production use (registry-based reconciliation,
PSA/NetworkPolicy/quota hardening, then a second-machine extension). Grants an application its own
isolated tenant slot — Namespace, least-privilege RBAC, network isolation, resource limits, and a
scoped kubeconfig — on a k3s cluster that already exists.

## Pre-conditions

- A k3s cluster is already provisioned and reachable — this skill does not provision it.
- The cluster-owning infrastructure repo has a git-committed per-machine app registry file (opaque
  app slugs only — no cross-repo application names or logic).
- A converge GitHub Actions workflow already exists **on the default branch**. If this is the very
  first converge workflow being added, land it via a normal merge before ever attempting
  `workflow_dispatch` — GitHub refuses to dispatch a workflow that isn't already on the default
  branch, so a brand-new workflow file needs an interim landing PR first.
- An external secrets store is available for the long-lived kubeconfig handoff — it is never
  committed to any repository.

## Steps

1. **Register the app.** Add an entry for the app's opaque slug to the per-machine registry file
   and commit it to the cluster-owning infrastructure repo.
2. **Trigger convergence** via `workflow_dispatch`. For every registered slug, the workflow:
   - Creates a Namespace named after the app slug (idempotent — skips if it already exists).
   - Grants least-privilege RBAC scoped to that Namespace only: a ServiceAccount, a Role (never a
     ClusterRole), and a RoleBinding. No cluster-wide access is ever granted.
   - Applies Pod Security Admission `restricted` enforcement on the Namespace.
   - Applies a default-deny NetworkPolicy (both ingress and egress) — cross-tenant traffic requires
     explicit opt-in, never implicit reachability.
   - Applies a ResourceQuota and LimitRange sized for the tenant.
   - **Waits, fail-closed, for the ServiceAccount's token and the cluster CA data to actually be
     populated** before assembling the kubeconfig — never assembles one from an empty or
     not-yet-issued token.
   - Assembles a scoped, long-lived kubeconfig bound to the tenant's ServiceAccount.
3. **Hand off the credential.** Publish the assembled kubeconfig as a workflow artifact and push it
   into the external secrets store. Never commit it to a repository.

## Output

A Namespace scoped to the app slug, containing least-privilege RBAC (ServiceAccount/Role/
RoleBinding), PSA `restricted` enforcement, a default-deny NetworkPolicy, a ResourceQuota/
LimitRange, and a scoped kubeconfig delivered via workflow artifact plus an external secrets store.

## Scope

- Does NOT provision the k3s cluster — use a dedicated cluster-provisioning mechanism for that.
- Does NOT deploy the onboarded app's own workloads. Building and applying the app's containers
  using the issued kubeconfig is a distinct, app-repo-side concern with no corresponding smaqit
  skill yet.
- Does NOT (yet) grant read access to `pods/log`, and does NOT (yet) provide non-destructive
  in-place credential rotation — see Known Gaps below.

## Known Gaps (not yet incorporated)

- **Read-only RBAC access to `pods/log`** — debugging a tenant's container log output currently
  requires access outside this skill's granted RBAC. In-flight in the source project, not yet
  folded into the proven mechanism.
- **Non-destructive in-place credential rotation** — invalidating every previously-issued token for
  an onboarded ServiceAccount (by recreating the object, which changes its UID) without touching
  the Namespace or the app's own workloads. In-flight in the source project, not yet folded in.

## Examples

**Input:** A downstream project's app-agnostic infrastructure repo already runs a self-hosted k3s
cluster with the converge workflow already landed on its default branch. A new application is
ready to move from local development to the shared cluster.

**Output:** The app's slug is added to the per-machine registry file and committed;
`workflow_dispatch` converges the cluster, creating a Namespace with least-privilege RBAC, PSA
`restricted` enforcement, a default-deny NetworkPolicy, and a sized ResourceQuota/LimitRange; a
scoped kubeconfig is issued only after its token and CA data are confirmed populated, then handed
off via a workflow artifact and the external secrets store — ready for the app's own (separate)
deploy mechanism to consume.

## Gotchas

- **`while read` + `ssh` silently truncates multi-slug runs to one entry.** The classic
  `while read slug; do ssh ...; done < registry-file` pattern lets the inner `ssh` consume the
  loop's own stdin, so only the first registered slug ever gets processed — with no error, just
  silent under-processing. Always give the remote command its own stdin (`< /dev/null` or `-n`).
- **`ca.crt` needs jsonpath escaping, not just quoting the whole expression.** The literal dot in
  the Secret data key `ca.crt` reads as a nested-field separator to `kubectl`'s jsonpath unless
  escaped (`ca\.crt`) or the key is bracket-quoted.
- **Kubeconfig assembly must be fail-closed on token/CA readiness.** ServiceAccount token
  population is not synchronous with ServiceAccount creation; assembling the kubeconfig
  immediately can silently produce one with an empty credential. Wait and fail loudly rather than
  handing off a broken kubeconfig.
- **A self-minted-token gap remains open.** A workload with a broad `secrets: create` grant in its
  own Namespace can mint an additional token for its own ServiceAccount, bypassing the platform's
  issuance/rotation flow. Tightening this (e.g. denying `create` on ServiceAccount-token-typed
  Secrets specifically) is not yet part of the proven mechanism.
- **A brand-new `workflow_dispatch` workflow must land on the default branch before its first
  dispatch.** This is a GitHub platform constraint, not a bug in the mechanism itself — plan an
  interim landing PR when introducing the converge workflow for the first time.
- **GitHub Environment deployment-branch policies are keyed by environment name only.** They apply
  to every job and trigger that references that environment name, regardless of intent — a
  PR-triggered, host-untouched dry-run job silently inherits the same branch restriction as a
  separately-gated `workflow_dispatch` apply job if they happen to share an environment name.

## Completion

- [ ] App slug added to the per-machine registry file and committed
- [ ] Converge workflow dispatched and completed without error
- [ ] Namespace exists, scoped to the app slug
- [ ] RBAC (ServiceAccount/Role/RoleBinding) scoped to the Namespace only — no cluster-wide grants
- [ ] PSA `restricted` enforcement active on the Namespace
- [ ] Default-deny NetworkPolicy present (ingress and egress)
- [ ] ResourceQuota and LimitRange applied
- [ ] Kubeconfig issued only after token/CA data confirmed populated (fail-closed wait observed)
- [ ] Kubeconfig delivered via workflow artifact and external secrets store — never committed

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Multi-slug converge run only processes the first registered slug | A `while read`-loop-plus-SSH bug: the inner `ssh` call inherits the loop's stdin and consumes the rest of the registry file on its first iteration. Fix: redirect the `ssh` call's stdin from `/dev/null` (or pass `-n`) so it never reads from the loop's input stream. |
| Assembled kubeconfig has an empty or missing CA field | A `ca.crt` jsonpath-quoting bug: `{.data.ca.crt}` misparses the literal dot in the key name as a nested-field separator. Fix: quote the key explicitly, e.g. `{.data['ca\.crt']}`. |
| Assembled kubeconfig has an empty or truncated token | The ServiceAccount token/CA data was read before the platform finished populating it. Fix: wait, fail-closed, until both are confirmed non-empty; never assemble a kubeconfig from a not-yet-issued token. |
| `workflow_dispatch` for a brand-new converge workflow returns "workflow does not exist" | GitHub only allows dispatching a workflow already present on the default branch. Land the new workflow file via a normal merge first, then dispatch it. |
| A PR-triggered dry-run/validation job unexpectedly fails an environment-gate check meant only for the separately-gated apply job | A GitHub Environment's deployment-branch policy applies per environment **name**, regardless of which job or trigger references it. Give the dry-run/read-only job either no environment or a distinctly named one — never share an environment name with the gated-apply job. |
| A workload in the onboarded Namespace has minted its own additional ServiceAccount token Secret | A self-minted-token gap: a broad `secrets: create` grant lets a workload mint extra tokens outside the platform's issuance flow. Not yet closed — treat any unexpected token Secret in a tenant Namespace as a signal to audit that Namespace's RBAC grants. |

## Allowed Tools

Bash(git:*), Bash(gh:*), Bash(kubectl:*)
