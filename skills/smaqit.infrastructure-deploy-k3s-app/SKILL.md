---
name: smaqit.infrastructure-deploy-k3s-app
description: Use when deploying an application into a Namespace-scoped Kubernetes cluster (k3s or any cluster that issues a scoped kubeconfig per application/environment). Given a kubeconfig restricted to the app's own Namespace, lints manifests against the platform's Pod Security `restricted` and quota guardrails, applies them, waits for rollout, stamps the Deployment with the commit SHA as pod environment, waits for the Ingress host's `Certificate` to become `Ready`, and verifies externally over HTTPS. Used in Phase 4 of `smaqit.new-greenfield-project` for `provisioning_mode: existing-k3s`, and again from the generated `deploy.yml` in Phase 5. Also usable as a manual fallback for direct deployment into an already-onboarded Namespace. Never touches anything cluster-scoped and never reads pod logs.
metadata:
  version: "1.0.0"
---

# Deploy Application to a Kubernetes Namespace

## Pre-conditions

- The target Namespace (`app-__APP_SLUG__`, both `test` and `prod`) has already been onboarded
  by the platform side (Namespace, RBAC, Pod Security label, NetworkPolicy,
  ResourceQuota/LimitRange, and the scoped kubeconfig already issued). This skill's job starts
  after onboarding — if the kubeconfig secret for the target environment is absent, fail loudly
  and point at the platform's own onboarding playbook rather than attempting anything
  cluster-scoped.
- Local Vault running and unsealed (`smaqit.infrastructure-vault-loader` complete); the target
  environment's registered machine-slug's kubeconfig at
  `secret/apps/<app-slug>/<machine-slug>/kubeconfig`.
- `kubectl` installed, at a version within the supported skew of the target cluster's minor
  version.
- `python3` with PyYAML available (`pip install pyyaml` if not already present) — required by
  `scripts/manifest-lint.sh`. On a fresh GitHub-hosted runner, install it explicitly rather than
  assuming the image ships it.
- The app's Deployment/Service manifests present locally under the project's deployment
  manifest directory.
- The Infrastructure spec declares, per environment: Namespace name (always
  `app-__APP_SLUG__`), machine-slug, ingress class, `ClusterIssuer` name, hostname, kubeconfig
  secret name, and the quota/Pod-Security limits the manifests must fit.

## Steps

1. **Resolve the target environment** (`test` for Phase 4's dev sweep, `prod` for Phase 5's
   CI/CD deploy) and read that environment's declared Namespace, ingress class, `ClusterIssuer`
   name, and hostname from the Infrastructure spec.

2. **Load the kubeconfig into a secure temp file** (local invocation):
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
   Environment secret (see `smaqit.infrastructure-cicd-generate`'s `k3s` mode) — every step from
   here on is identical either way.

3. **Run the namespace guard before touching anything else:**
   ```bash
   scripts/namespace-guard.sh "$KUBECONFIG" "app-${APP_SLUG}"
   ```
   It reads the kubeconfig's current context, confirms the target Namespace is exactly
   `app-__APP_SLUG__`, and runs `kubectl auth can-i --list -n app-__APP_SLUG__` to confirm no
   cluster-scoped verb is permitted. A non-zero exit means stop — do not proceed with a broader
   credential than the onboarding contract issues, and do not retry with different flags.

4. **Create an `imagePullSecret`, only if the spec names a private registry:**
   ```bash
   kubectl create secret docker-registry app-registry \
     --namespace "app-${APP_SLUG}" \
     --docker-server=<registry-host> \
     --docker-username="$REGISTRY_USERNAME" \
     --docker-password="$REGISTRY_TOKEN" \
     --dry-run=client -o yaml | kubectl apply -n "app-${APP_SLUG}" -f -
   ```
   `REGISTRY_USERNAME`/`REGISTRY_TOKEN` come from the app's own GitHub Environment secret, never
   anything platform-owned. Skip this step entirely for a public image.

5. **Render the Ingress from the shared template** — never hand-write a per-app Ingress:
   ```bash
   sed \
     -e "s/__APP_SLUG__/${APP_SLUG}/g" \
     -e "s/__INGRESS_CLASS__/<ingress-class>/g" \
     -e "s/__CLUSTER_ISSUER__/<cluster-issuer-name>/g" \
     -e "s/__HOST__/<host>/g" \
     assets/ingress.yaml.template > ingress.rendered.yaml
   ```

6. **Lint every manifest before any apply:**
   ```bash
   scripts/manifest-lint.sh deployment.yaml service.yaml ingress.rendered.yaml
   ```
   Checks the Pod Security `restricted` fields (`runAsNonRoot: true`,
   `allowPrivilegeEscalation: false`, `capabilities.drop: [ALL]`, `seccompProfile.type:
   RuntimeDefault`, no `privileged: true`, no `hostPath` volumes, no `hostNetwork`/`hostPID`/
   `hostIPC`), explicit CPU/memory `requests`/`limits` on every container, and the absence of any
   cluster-scoped `kind` in the manifest set. A rejected manifest exits non-zero and names the
   offending field — fix the manifest, not the check. The admission rejection this prevents is
   terse and this skill's credential cannot read logs to diagnose it after the fact.

7. **Apply into the app's own Namespace only:**
   ```bash
   kubectl apply -n "app-${APP_SLUG}" -f deployment.yaml -f service.yaml -f ingress.rendered.yaml
   ```
   `deployment.yaml` must set `DEPLOY_SHA`/`DEPLOY_TIME` as pod environment (see
   `assets/deployment.yaml.template`), from the current commit SHA and an ISO 8601 UTC
   timestamp, at apply time — never a file written to a host path; this credential has no host
   filesystem to write to.

8. **Wait for rollout:**
   ```bash
   kubectl rollout status -n "app-${APP_SLUG}" "deployment/${APP_SLUG}" --timeout=180s
   ```

9. **Wait for the `Certificate` to reach `Ready`** — the RBAC allowlist grants read access to
   this object specifically, and it covers the TLS-validity gap
   `smaqit.infrastructure-deploy-verify` explicitly does not check:
   ```bash
   kubectl wait -n "app-${APP_SLUG}" "certificate/${APP_SLUG}-tls" --for=condition=Ready --timeout=180s
   ```

10. **Verify externally:**
    ```bash
    smaqit.infrastructure-deploy-verify --url "https://<host>" --expected-sha "$(git rev-parse HEAD)"
    ```
    Invoked unmodified — exactly as any other deploy skill in the product calls it.

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
  fails loudly and points at the platform's own onboarding playbook.

## Examples

**Input:** `provisioning_mode: existing-k3s` project's Phase 4 dev sweep invokes this skill.

**Output:** Kubeconfig loaded from `secret/apps/<app-slug>/<test-machine-slug>/kubeconfig`
(`value` field); `namespace-guard.sh` confirms the Namespace is `app-<app-slug>` with no
cluster-scoped verb permitted; `manifest-lint.sh` passes; Deployment/Service/rendered Ingress
applied into `app-<app-slug>`; rollout completes; `Certificate` reaches `Ready`; `deploy-verify`
reports PASS on health, SHA, and the SPA root over `https://<test-host>`.

**Input:** Generated `deploy.yml` (k3s mode) runs from GitHub Actions on push to `main`.

**Output:** `KUBECONFIG` written from the `prod` Environment secret; same guard → lint → apply →
rollout → `Certificate` → verify sequence runs against `app-<app-slug>` in the `prod` Namespace;
`deploy-verify` reports PASS over `https://<prod-host>`.

## Gotchas

- **Never widen scope to work around a guard refusal.** `namespace-guard.sh` failing signals the
  onboarding contract or the supplied kubeconfig is wrong, not an obstacle to route around.
- **No pod logs, ever.** The scoped credential's RBAC allowlist deliberately excludes
  `pods/log` and `pods/exec` — this is why verification is external (`Certificate` `Ready` +
  `smaqit.infrastructure-deploy-verify` over HTTPS) rather than log inspection.
- **The Ingress template is shared, not per-app.** Always render from
  `assets/ingress.yaml.template`; a hand-written Ingress risks drifting from the platform's own
  ingress-class/`ClusterIssuer` conventions.
- **Deploy stamp is pod environment, not a file.** Do not adapt the rsync family's
  `printf '%s' > file` pattern here — there is no host filesystem this credential can write to.
  Set `DEPLOY_SHA`/`DEPLOY_TIME` as container env vars in the Deployment manifest at apply time.
- **`kubectl` version skew.** Confirm the local/CI `kubectl` minor version is within the
  supported skew of the target cluster before applying.
- **Kubeconfig rotation is destructive and platform-side.** Never regenerate one locally — see
  `smaqit.infrastructure-vault-loader`'s `rotate-credential.sh` support for
  `apps/<app-slug>/<machine-slug>/kubeconfig`, which re-prompts for a freshly reissued value
  instead.

## Completion

- [ ] Target environment resolved; Namespace/ingress class/`ClusterIssuer`/hostname read from
      the Infrastructure spec
- [ ] Kubeconfig loaded from Vault (local) or the `KUBECONFIG` Environment secret (CI/CD)
- [ ] `namespace-guard.sh` passed
- [ ] `imagePullSecret` created only when a private registry is declared
- [ ] Ingress rendered from the shared template
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
| `manifest-lint.sh` rejects a manifest | Stop before any `kubectl apply`. Report the exact offending field. Fix the manifest, not the check. |
| `kubectl apply` fails (admission rejection) | Report the admission error verbatim. Do not retry with `--force` or widened scope. |
| `kubectl rollout status` times out | Do not read pod logs — the credential cannot. Report the timeout and the last known state from `kubectl get deployment`/`kubectl get events` (both within the RBAC allowlist). |
| `Certificate` never reaches `Ready` | Check the `ClusterIssuer` name and ACME HTTP-01 reachability from the Infrastructure spec's declared values; do not bypass the wait. |
| `smaqit.infrastructure-deploy-verify` reports FAIL | Stop. Report the failing check. Do not mark the deployment complete. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |
| Output artifact already exists | Confirm with user before overwriting. |

## Allowed Tools

Bash(kubectl:*), Bash(vault:*)
