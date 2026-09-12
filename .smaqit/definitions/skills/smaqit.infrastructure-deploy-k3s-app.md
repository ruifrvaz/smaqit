# smaqit.infrastructure-deploy-k3s-app

## Description

Use when deploying an application into a Namespace-scoped Kubernetes cluster (k3s or any
cluster that issues a scoped kubeconfig per application/environment) — given a kubeconfig
whose credential is restricted to a single Namespace, lint the app's manifests against the
platform's Pod Security `restricted` and quota guardrails, apply them into the app's own
Namespace, wait for rollout, stamp the Deployment with the commit SHA as pod environment, wait
for the Ingress host's `Certificate` to become `Ready`, and verify externally over HTTPS. Used
in Phase 4 (Dev Environment Sweep) of `smaqit.new-greenfield-project` for `provisioning_mode:
existing-k3s`, invoked via Phase 4 Step 6's family-aware match/synthesis procedure, and again
from the generated `deploy.yml` (`smaqit.infrastructure-cicd-generate`'s `k3s` mode) for Phase
5's CI/CD production deploy. Also usable as a manual fallback for direct deployment into an
already-onboarded Namespace outside the CI/CD pipeline. Named for the platform vocabulary this
capability pairs with (an onboarding skill that provisions the Namespace/RBAC/quota and issues
the kubeconfig), even though this skill's own content is plain Kubernetes and carries no
k3s-only assumption — it works against any cluster that issues a scoped kubeconfig the same
way. It never requests onboarding, never touches anything cluster-scoped, and never reads pod
logs — the scoped credential this skill consumes cannot grant either.

## Provenance

- `synthesized: false`
- `contributed-for-project: [a downstream platform]`
- `contributed-date: 2026-09-12`
- `contributed-stack: Kubernetes (Deployment/Service/Ingress, Pod Security Admission restricted, ResourceQuota/LimitRange, NetworkPolicy, namespace-scoped RBAC), k3s, cert-manager (Certificate/ClusterIssuer), Traefik ingress class, kubectl`
- Authored directly into canonical `smaqit` (task-106 "reconcile into product" shape) rather
  than staged as a downstream-only definition first — the framework owner authored this
  capability directly, so no separate contribution/reconciliation handoff was needed.

## Required-inherited-context (inherited verbatim by any future synthesized member of this family, not reinvented)

1. **`__APP_SLUG__` token convention.** The Namespace a scoped kubeconfig is restricted to is
   always `app-__APP_SLUG__`, never a free-form name. Every manifest, guard, and lint check in
   this family addresses the Namespace via this token — the same discipline the rsync family
   applies to `__APP_DIR__`.
2. **A shared Ingress template (bundled asset) sets `ingressClassName` and the cert-manager
   `cluster-issuer` annotation from declared spec values** — never a hand-written Ingress per
   app. This is the k3s family's analog of the rsync family's shared `write-vhost.sh` — a
   platform-facing concern solved once, centrally, not reinvented per application.
3. **The deploy stamp is delivered as pod environment (`DEPLOY_SHA`, `DEPLOY_TIME`), set at
   apply time** — never written to a file on a host filesystem, because there is no host
   filesystem this credential can reach. This preserves the existing health-endpoint contract
   (`{sha, deployedAt}`) so `smaqit.infrastructure-deploy-verify` is reused **unchanged**,
   called with `--url https://<host>`. This is the k3s family's analog of the rsync family's
   `smaqit.infrastructure-hook-post-deploy-stamp` reuse.
4. **A preflight guard (bundled `namespace-guard.sh`)** reads the kubeconfig's current context,
   runs `kubectl auth can-i --list` in the target Namespace, and refuses to proceed if any
   cluster-scoped verb is permitted or the Namespace doesn't match `app-__APP_SLUG__`. This is
   the k3s family's analog of the rsync/Terraform family's `plan-guard.sh`/`ownership-guard.sh`
   — guarding against ever running with a broader credential than the onboarding contract
   issues.

## Steps

### Pre-conditions

- The target Namespace (`app-__APP_SLUG__`, both `test` and `prod` environments) has already
  been onboarded by the platform side (Namespace, RBAC, Pod Security label, NetworkPolicy,
  ResourceQuota/LimitRange, and the scoped kubeconfig already issued) — this skill's job starts
  after onboarding, never before. If the kubeconfig secret for the target environment is
  absent, fail loudly with a pointer to the platform's own onboarding playbook rather than
  attempting anything cluster-scoped to fix it.
- Local Vault running and unsealed (`smaqit.infrastructure-vault-loader` complete); the target
  environment's registered machine-slug's kubeconfig at
  `secret/apps/<app-slug>/<machine-slug>/kubeconfig`.
- `kubectl` installed locally (or available on the CI runner) at a version within the
  supported skew of the target cluster's minor version.
- `python3` with PyYAML available (required by `manifest-lint.sh`) — install it explicitly on a
  fresh CI runner rather than assuming the image ships it.
- The app's own Deployment/Service manifests (and, when the image is private, a declared
  registry credential) present locally under the project's deployment manifest directory.
- The Infrastructure spec declares, per environment: Namespace name (always
  `app-__APP_SLUG__`), machine-slug, ingress class, `ClusterIssuer` name, hostname, kubeconfig
  secret name, and the quota/Pod-Security limits the manifests must fit.

### Steps

1. **Resolve the target environment** (`test` for Phase 4's dev sweep, `prod` for Phase 5's
   CI/CD deploy) and read that environment's declared Namespace, ingress class, `ClusterIssuer`
   name, and hostname from the Infrastructure spec.

2. **Load the environment's kubeconfig** from Vault into a secure temp file, never printed or
   logged:
   ```bash
   export VAULT_ADDR=http://127.0.0.1:8200
   export APP_SLUG=<app-slug>
   export MACHINE_SLUG=<machine-slug-for-target-environment>   # from the Infrastructure spec

   TMPKUBECONFIG=$(mktemp)
   trap "rm -f $TMPKUBECONFIG" EXIT
   vault kv get -field=value "secret/apps/${APP_SLUG}/${MACHINE_SLUG}/kubeconfig" > "$TMPKUBECONFIG"
   export KUBECONFIG="$TMPKUBECONFIG"
   ```
   In CI/CD, `KUBECONFIG` is instead written from the environment's `KUBECONFIG` GitHub
   Environment secret (`smaqit.infrastructure-cicd-generate`'s `k3s` mode) — same downstream
   steps either way.

3. **Run the namespace guard** (`scripts/namespace-guard.sh`) before touching anything else. It
   reads the kubeconfig's current context, confirms the Namespace is exactly
   `app-__APP_SLUG__`, and runs `kubectl auth can-i --list -n app-__APP_SLUG__` to confirm no
   cluster-scoped verb is permitted. Refuses (non-zero exit) otherwise — do not proceed past a
   refusal by widening scope or retrying with different flags.

4. **Create an `imagePullSecret`, only if the spec names a private registry.** Declared registry
   credential (`REGISTRY_USERNAME`/`REGISTRY_TOKEN` or equivalent, sourced from the app's own
   GitHub Environment secret, never anything platform-owned):
   ```bash
   kubectl create secret docker-registry app-registry \
     --namespace app-__APP_SLUG__ \
     --docker-server=<registry-host> \
     --docker-username="$REGISTRY_USERNAME" \
     --docker-password="$REGISTRY_TOKEN" \
     --dry-run=client -o yaml | kubectl apply -n app-__APP_SLUG__ -f -
   ```
   Skip entirely for a public image — this step must not run unconditionally.

5. **Render the Ingress from the shared template** (`assets/ingress.yaml.template`),
   substituting `__APP_SLUG__`, the declared ingress class, `ClusterIssuer` name, and hostname.
   Never hand-write a per-app Ingress.

6. **Lint every manifest before any apply** (`scripts/manifest-lint.sh`) against: the Pod
   Security `restricted` fields (`runAsNonRoot: true`, `allowPrivilegeEscalation: false`,
   `capabilities.drop: [ALL]`, `seccompProfile.type: RuntimeDefault`, no `privileged: true`, no
   `hostPath` volumes, no host namespaces `hostNetwork`/`hostPID`/`hostIPC`); explicit CPU/memory
   `requests`/`limits` on every container; and the absence of any cluster-scoped `kind`
   (`Namespace`, `ClusterRole`, `ClusterRoleBinding`, `PersistentVolume`, etc.) anywhere in the
   manifest set. A rejected manifest fails fast, naming the offending field — the platform's
   admission rejection after the fact is terse and this skill's credential cannot read logs to
   diagnose it after an apply.

7. **Apply into the app's own Namespace only:**
   ```bash
   kubectl apply -n app-__APP_SLUG__ -f deployment.yaml -f service.yaml -f ingress.yaml
   ```
   The Deployment manifest carries `DEPLOY_SHA`/`DEPLOY_TIME` as pod environment, set from the
   current commit SHA and an ISO 8601 UTC timestamp at apply time (see required-inherited-context
   item 3) — never a file written to a host path, since this credential has no host filesystem
   access.

8. **Wait for rollout:**
   ```bash
   kubectl rollout status -n app-__APP_SLUG__ deployment/<app-slug> --timeout=180s
   ```

9. **Wait for the `Certificate` to reach `Ready`** (the RBAC allowlist grants read access to
   this object specifically, so this check substitutes for the TLS-validity gap
   `smaqit.infrastructure-deploy-verify` explicitly does not cover):
   ```bash
   kubectl wait -n app-__APP_SLUG__ certificate/<app-slug>-tls --for=condition=Ready --timeout=180s
   ```

10. **Verify externally:** invoke `smaqit.infrastructure-deploy-verify --url https://<host>
    --expected-sha <sha>` — unmodified, exactly as any other deploy skill in the product calls
    it.

## Output

Application manifests applied into `app-__APP_SLUG__` (Deployment, Service, Ingress), rollout
complete, `Certificate` `Ready`, deploy stamp present as pod environment, external verification
PASS on health, SHA, and SPA root over HTTPS.

## Scope

- Does NOT onboard the Namespace, RBAC, Pod Security label, NetworkPolicy, or quota — that is
  the platform side's own skill; this skill's precondition, not its job.
- Does NOT provision the cluster itself — a separate, unrelated concern.
- Does NOT read pod logs, at any step — the scoped credential cannot grant `pods/log`, and
  verification is deliberately external for exactly this reason.
- Does NOT touch anything cluster-scoped — `namespace-guard.sh` refuses to proceed if the
  kubeconfig's own permissions would even allow it.
- Does NOT request onboarding on the app's behalf — if the kubeconfig secret is absent, it
  fails loudly and points at the platform's own onboarding playbook. The direction stays apps →
  infrastructure, never the reverse.

## Completion

- [ ] Target environment resolved; Namespace/ingress class/`ClusterIssuer`/hostname read from
      the Infrastructure spec
- [ ] Kubeconfig loaded from Vault (local) or the `KUBECONFIG` Environment secret (CI/CD) —
      never typed, never a cluster call to obtain or validate it
- [ ] `namespace-guard.sh` passed — Namespace matches `app-__APP_SLUG__`, no cluster-scoped verb
      permitted
- [ ] `imagePullSecret` created only when a private registry is declared
- [ ] Ingress rendered from the shared template, not hand-written
- [ ] `manifest-lint.sh` passed before any `kubectl apply`
- [ ] `kubectl apply -n app-__APP_SLUG__` succeeded
- [ ] `kubectl rollout status` succeeded
- [ ] `Certificate` reached `Ready`
- [ ] `smaqit.infrastructure-deploy-verify --url https://<host> --expected-sha <sha>` invoked and
      passed

## Failure Handling

| Situation | Action |
|-----------|--------|
| Kubeconfig secret absent for the target environment | Fail loudly; point at the platform's onboarding playbook. Never attempt a cluster-scoped fix. |
| `namespace-guard.sh` refuses (cluster-scoped verb permitted, or Namespace mismatch) | Stop immediately. Do not proceed with a broader credential than the onboarding contract issues. |
| `manifest-lint.sh` rejects a manifest | Stop before any `kubectl apply`. Report the exact offending field (e.g. missing `runAsNonRoot`, a `privileged: true` container, an over-quota resource request). Fix the manifest, not the lint check. |
| `kubectl apply` fails (admission rejection) | Report the admission error verbatim. Do not retry with `--force` or widened scope. |
| `kubectl rollout status` times out | Do not attempt to read pod logs — the credential cannot. Report the timeout and the last known rollout state from `kubectl get deployment`/`kubectl get events` (both within the RBAC allowlist). |
| `Certificate` never reaches `Ready` | Check the `ClusterIssuer` name and ACME HTTP-01 reachability from the Infrastructure spec's declared values; do not bypass the wait. |
| `smaqit.infrastructure-deploy-verify` reports FAIL | Stop. Report the failing check. Do not mark the deployment complete. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |
| Output artifact already exists | Confirm with user before overwriting. |

## Gotchas

- **Never widen scope to work around a guard refusal.** `namespace-guard.sh` failing is a
  signal the onboarding contract or the supplied kubeconfig is wrong, not an obstacle to route
  around.
- **No pod logs, ever.** The scoped credential's RBAC allowlist deliberately excludes
  `pods/log` and `pods/exec` — this is why verification is external (`Certificate` `Ready` +
  `smaqit.infrastructure-deploy-verify` over HTTPS) rather than log inspection. Do not add a log
  read anywhere in this skill even for debugging.
- **The Ingress template is shared, not per-app.** A hand-written Ingress risks drifting from
  the platform's own ingress-class/`ClusterIssuer` conventions; always render from
  `assets/ingress.yaml.template`.
- **Deploy stamp is pod environment, not a file.** Do not adapt the rsync family's
  `printf '%s' > file` pattern here — there is no host filesystem this credential can write to.
  Set `DEPLOY_SHA`/`DEPLOY_TIME` as container env vars in the Deployment manifest at apply time
  instead.
- **`kubectl` version skew.** Confirm the local/CI `kubectl` minor version is within the
  supported skew of the target cluster before applying — a version mismatch can silently drop
  or misinterpret manifest fields on `apply`.
- **Kubeconfig rotation is destructive and platform-side.** If the kubeconfig ever needs
  rotating, that happens on the platform's side, not by generating a new one locally — see
  `smaqit.infrastructure-vault-loader`'s `rotate-credential.sh` support for
  `apps/<app-slug>/<machine-slug>/kubeconfig`, which re-prompts for a freshly reissued value
  rather than regenerating one.

## Allowed Tools

Bash(kubectl:*), Bash(vault:*)

## Examples

**Input:** `provisioning_mode: existing-k3s` project's Phase 4 dev sweep invokes this skill
after Phase 4 Step 6 matched the declared stack against this skill's family (or, on first use of
a new stack shape within the family, synthesized a sibling from this skill as exemplar).

**Output:** Kubeconfig loaded from `secret/apps/<app-slug>/<test-machine-slug>/kubeconfig`
(`value` field); `namespace-guard.sh` confirms the Namespace is `app-<app-slug>` with no
cluster-scoped verb permitted; `manifest-lint.sh` passes; Deployment/Service/rendered Ingress
applied into `app-<app-slug>`; rollout completes; `Certificate` reaches `Ready`; `deploy-verify`
reports PASS on health, SHA, and the SPA root over `https://<test-host>`.
