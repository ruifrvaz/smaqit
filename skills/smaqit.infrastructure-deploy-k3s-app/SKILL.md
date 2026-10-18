---
name: smaqit.infrastructure-deploy-k3s-app
description: Use when deploying an app into a Namespace-scoped Kubernetes cluster (k3s, or any cluster issuing a per-app kubeconfig), including provisioning_mode existing-k3s in greenfield Phase 4 and its deploy.yml. Checks mesh reachability for private-mesh targets, lints manifests against Pod Security restricted, applies the Ingress the spec's mode calls for, waits for rollout, and verifies.
metadata:
  version: "2.1.1"
  smaqit-slots: "platform-prereqs=mesh-check, app-deploy=deploy"
---

# Deploy Application to a Kubernetes Namespace

Everything that differs between targets comes from the Infrastructure spec's `## Environments`
row for the target environment, read through `smaqit spec environments --env=<environment>
--json` — never from guesses or binding `params`:

| Field | Values | Effect |
|-------|--------|--------|
| `reachability` | `public` / `private-mesh` | `private-mesh`: the API (and the app) are reachable only from mesh peers — this machine must be on the mesh (`mesh-check`), and CI joins it first |
| `ingressMode` | `own-tls` / `platform-default-tls` / `platform-owned` | which Ingress is applied, if any, and whether a `Certificate` is awaited |
| `host`, `ingressClass`, `clusterIssuer`, `healthPath` | from the platform team | stamped into the Ingress / probe at apply time |

The same JSON carries three spec-wide values: `probeClient` (`curl`, `wget`, or `python3`),
`registry` (`{host, visibility}`, or `null` for a public image), and `deployerExec` (`denied`
unless the spec's `Deployer Exec` row says `allowed`).

Verification runs **inside the cluster** (a probe Job) when the deployer cannot reach the app
from outside — `private-mesh`, or `platform-owned` (for example behind the platform's SSO gate)
— and **over HTTPS** otherwise.

## Pre-conditions

- The target Namespace (`app-__APP_SLUG__`, on each target environment's machine) has already
  been onboarded by the platform side (Namespace, RBAC, Pod Security label, NetworkPolicy,
  ResourceQuota/LimitRange, and the scoped kubeconfig already issued). This skill's job starts
  after onboarding — if the kubeconfig secret for the target environment is absent, fail loudly
  and point at the platform's own onboarding playbook rather than attempting anything
  cluster-scoped.
- Local Vault running and unsealed (`smaqit.infrastructure-vault-loader` complete); the target
  environment's registered machine-slug's kubeconfig at
  `secret/apps/<app-slug>/<machine-slug>/kubeconfig`. A merged onboarding request alone doesn't
  put it there — the platform's own hand-off delivers it, and the `target-credentials` load
  stores it.
- `smaqit spec environments --env=<environment> --json` succeeds: the Infrastructure spec's
  `## Environments` row for the target is complete. `smaqit slots plan --mode=existing-k3s`
  refuses to plan otherwise.
- `kubectl` installed, at a version within the supported skew of the target cluster's minor
  version; `jq` installed.
- `python3` with PyYAML available (`pip install pyyaml` if not already present) — required by
  `scripts/manifest-lint.sh`. On a fresh GitHub-hosted runner, install it explicitly rather than
  assuming the image ships it.
- The app's manifests under `deployment/k3s/` — `deployment.yaml`, `service.yaml`, `pvc.yaml` when
  the app keeps state, and, once rendered (Step 5), the Ingress file for each Ingress mode in use
  and `probe-job.yaml` when any environment verifies in-cluster — the same paths the generated CI
  workflow uses.

## Operations

### `mesh-check` (slot `platform-prereqs`)

Confirms this machine can reach a `private-mesh` target's API before credentials, onboarding, or
any `kubectl` call. A `public` target passes without probing anything.

```bash
ENVIRONMENTS="$(mktemp)"
smaqit spec environments --json > "$ENVIRONMENTS"
bash scripts/mesh-check.sh "$ENVIRONMENTS" <environment>     # --port <api-port>, default 6443
```

It resolves the row's `meshWaitForHost` (the API's mesh name) and opens a TCP connection to the
API port. On failure it stops with instructions: join the platform's mesh on this machine (its
mesh client, with the operator's own login or a platform-issued setup key), then run the slot
again. Never work around it by reaching the API some other way. In CI the generated `deploy.yml`
joins the mesh itself, so this operation runs only for local deployments.

### `deploy` (slot `app-deploy`)

1. **Resolve the target environment** (the test environment for Phase 4's dev sweep, prod for
   Phase 5's CI/CD deploy) and read its row:
   ```bash
   smaqit spec environments --env=<environment> --json
   ```
   The row's machine slug also names the app's GitHub Environment for that target. For a
   `private-mesh` row, `mesh-check` must already have passed.

2. **Load the kubeconfig into a private temp file** (local invocation) through the vault-loader
   script, which prints only the file's path:
   ```bash
   export APP_SLUG=<app-slug>
   export MACHINE_SLUG=<machine-slug-for-target-environment>   # from the Environments row

   KUBECONFIG="$(bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/credential-file.sh \
     "apps/${APP_SLUG}/${MACHINE_SLUG}/kubeconfig")"
   trap 'rm -rf "$(dirname "$KUBECONFIG")"' EXIT
   export KUBECONFIG
   ```
   If the script reports the kubeconfig missing, stop: the platform's onboarding hand-off hasn't
   completed (see Failure Handling). In CI/CD, `KUBECONFIG` is instead written from the
   `KUBECONFIG` secret of the GitHub Environment named after the target machine slug (see
   `smaqit.infrastructure-cicd-generate`'s `k3s` mode) — every step from here on is identical
   either way.

3. **Run the namespace guard before touching anything else:**
   ```bash
   scripts/namespace-guard.sh "$KUBECONFIG" "app-${APP_SLUG}" --deployer-exec <deployerExec>
   ```
   It confirms the API answers and accepts the credential, that the kubeconfig's current context
   targets exactly `app-__APP_SLUG__`, that no cluster-scoped verb is permitted, and that the
   deployer cannot exec into pods (`kubectl auth can-i create|get pods --subresource=exec` must
   both answer `no`). It fails closed: only an explicit "no" counts as denied, so an unreachable
   API is a FAIL, not a PASS. Under `--deployer-exec allowed` an exec grant is a warning instead
   of a failure; an unclear answer still fails. A non-zero exit means stop — do not proceed with
   a broader credential than the onboarding contract issues, and do not retry with different flags.

4. **Create an `imagePullSecret`, only if the spec's `registry` is private.** The credential
   reaches `kubectl` through a file descriptor, never as a command-line argument, where the
   process list would show it:
   ```bash
   REGISTRY_HOST=<registry host> \
   kubectl create secret generic app-registry \
     --namespace "app-${APP_SLUG}" \
     --type=kubernetes.io/dockerconfigjson \
     --from-file=.dockerconfigjson=<(jq -n '{auths: {(env.REGISTRY_HOST): {username: env.REGISTRY_USERNAME, password: env.REGISTRY_TOKEN, auth: ("\(env.REGISTRY_USERNAME):\(env.REGISTRY_TOKEN)" | @base64)}}}') \
     --dry-run=client -o yaml | kubectl apply -n "app-${APP_SLUG}" -f -
   ```
   `REGISTRY_USERNAME`/`REGISTRY_TOKEN` come from the app's own GitHub Environment secret (in CI,
   through the step's `env:`), never anything platform-owned. Skip this step entirely for a
   public image; the generated `deploy.yml` renders it only for a private registry.

5. **Render the manifests the spec's environments need from the shared templates** — never
   hand-write them — once, and commit them. Substitute only `__APP_SLUG__` (and, for the probe,
   `__RUN_AS_USER__` and `__PROBE_FETCH__`); the per-environment tokens stay in the committed files
   and are stamped at apply time:

   | Needed when some environment is… | Template | Committed as |
   |---|---|---|
   | `own-tls` | `assets/ingress.yaml.template` | `deployment/k3s/ingress.yaml` |
   | `platform-default-tls` | `assets/ingress-default-tls.yaml.template` | `deployment/k3s/ingress-default-tls.yaml` |
   | `private-mesh` or `platform-owned` | `assets/probe-job.yaml.template` | `deployment/k3s/probe-job.yaml` |

   ```bash
   sed -e "s/__APP_SLUG__/${APP_SLUG}/g" assets/ingress.yaml.template > deployment/k3s/ingress.yaml
   sed -e "s/__APP_SLUG__/${APP_SLUG}/g" -e "s/__RUN_AS_USER__/<deployment's runAsUser>/" \
     -e "s#__PROBE_FETCH__#<fetch command>#" assets/probe-job.yaml.template > deployment/k3s/probe-job.yaml
   ```
   `<fetch command>` follows the spec's `Probe Client`:

   | Probe Client | Fetch command |
   |---|---|
   | `curl` | `curl -fsS -o /dev/null` |
   | `wget` | `wget -q -O /dev/null` |
   | `python3` | `python3 -c 'import sys, urllib.request; urllib.request.urlopen(sys.argv[1], timeout=5)'` |

   Pick the client the app's own image already contains (a `python:*-slim` image has `python3` but
   neither `curl` nor `wget`); never add an OS package to the production image only for the probe.
   `platform-owned` needs no Ingress file at all. For this run, stamp a
   temporary copy, never the committed file:
   ```bash
   STAMPED="$(mktemp -d)"
   INGRESS="${STAMPED}/ingress.yaml"      # own-tls; ingress-default-tls.yaml for platform-default-tls
   sed -e "s/__HOST__/<host>/g" -e "s/__INGRESS_CLASS__/<ingress-class>/g" \
     -e "s/__CLUSTER_ISSUER__/<cluster-issuer>/g" deployment/k3s/ingress.yaml > "$INGRESS"
   ```

6. **Lint every manifest before any apply** — `pvc.yaml` when present, the Ingress for this mode
   (none for `platform-owned`), and the stamped probe Job when this run verifies in-cluster:
   ```bash
   scripts/manifest-lint.sh deployment/k3s/deployment.yaml deployment/k3s/service.yaml "$INGRESS"
   ```
   Checks the Pod Security `restricted` fields (`runAsNonRoot: true`, an explicit numeric
   `runAsUser` wherever `runAsNonRoot` is true, `allowPrivilegeEscalation: false`,
   `capabilities.drop: [ALL]`, `seccompProfile.type: RuntimeDefault`, no `privileged: true`, no
   `hostPath` volumes, no `hostNetwork`/`hostPID`/`hostIPC`), explicit CPU/memory
   `requests`/`limits` on every container, and the absence of any cluster-scoped `kind` in the
   manifest set. A rejected manifest exits non-zero and names the offending field — fix the
   manifest, not the check. The admission rejection this prevents is terse, and a pod that never
   starts leaves no logs to diagnose it with; a base image whose own `USER`
   directive names a user rather than a numeric UID is the most common way `runAsNonRoot` alone
   still fails admission with `CreateContainerConfigError`.

7. **Apply into the app's own Namespace only**, `pvc.yaml` first when present, then the
   Deployment, the Service, and this mode's Ingress (none for `platform-owned`):
   ```bash
   kubectl apply -n "app-${APP_SLUG}" -f deployment/k3s/deployment.yaml -f deployment/k3s/service.yaml -f "$INGRESS"
   ```
   `deployment.yaml` must set `DEPLOY_SHA`/`DEPLOY_TIME` as pod environment (see
   `assets/deployment.yaml.template`), from the current commit SHA and an ISO 8601 UTC
   timestamp, at apply time — never a file written to a host path; this credential has no host
   filesystem to write to.

8. **Wait for rollout, then prove exec is refused:**
   ```bash
   kubectl rollout status -n "app-${APP_SLUG}" "deployment/${APP_SLUG}" --timeout=180s
   scripts/exec-refused.sh "app-${APP_SLUG}" "${APP_SLUG}" --deployer-exec <deployerExec>
   ```
   `exec-refused.sh` attempts a real `kubectl exec` into the Deployment and passes only when the
   API refuses it as Forbidden. An exec that goes through (even one whose command then fails
   inside the container) fails under `denied` and warns under `allowed`; NotFound, a timeout, or
   any other answer fails closed.

9. **`own-tls` only — wait for the `Certificate` to reach `Ready`.** The RBAC allowlist grants
   read access to this object specifically, and it covers the TLS-validity gap
   `smaqit.infrastructure-deploy-verify` explicitly does not check. `platform-default-tls` and
   `platform-owned` request no certificate, so there is nothing to wait for:
   ```bash
   kubectl wait -n "app-${APP_SLUG}" "certificate/${APP_SLUG}-tls" --for=condition=Ready --timeout=180s
   ```

10. **Verify through `smaqit.infrastructure-deploy-verify`**, invoked unmodified — exactly as any
    other deploy skill in the product calls it. It reads the same Environments row and chooses
    in-cluster verification (the probe Job) for `private-mesh` or `platform-owned`, and HTTPS
    otherwise:
    ```bash
    smaqit.infrastructure-deploy-verify --environment <environment> --expected-sha "$(git rev-parse HEAD)"
    ```
    Right after a rollout the new pods may not be serving yet: both modes retry every 5 seconds
    for up to 60 seconds (the same window the generated CI workflow uses) before reporting FAIL.
    A failure that persists past the window is real.

## Output

`mesh-check`: PASS when the target is `public` or its mesh name answers on the API port.

`deploy`: application manifests applied into `app-__APP_SLUG__` (Deployment, Service, a PVC when
committed, and the Ingress this mode calls for), rollout complete, `Certificate` `Ready` for
`own-tls`, deploy stamp present as pod environment, and verification PASS — the Deployment runs
this commit's image, and the health path answers in-cluster or over HTTPS.

## Scope

- Does NOT onboard the Namespace, RBAC, Pod Security label, NetworkPolicy, or quota — that is
  the platform side's own skill; this skill's precondition, not its job.
- Does NOT provision the cluster itself — a separate, unrelated concern.
- Does NOT join a mesh for the operator — `mesh-check` only confirms access and stops with
  instructions; CI's mesh join is a workflow step the spec configures.
- Does NOT rely on pod logs for verification — the verdict is the probe Job's status or the HTTPS
  checks. Whether the scoped credential may read `pods/log` is the platform's decision; when it
  can, a failed rollout's or probe's logs may be read for diagnosis. Never `pods/exec`:
  `namespace-guard.sh` and `exec-refused.sh` fail the deploy when the deployer could exec, unless
  the spec's `Deployer Exec` row records a deliberate grant.
- Does NOT touch anything cluster-scoped — `namespace-guard.sh` refuses to proceed if the
  kubeconfig's own permissions would even allow it.
- Does NOT request onboarding on the app's behalf — if the kubeconfig secret is absent, it
  fails loudly and points at the platform's own onboarding playbook.
- Does NOT ship an Ingress for `platform-owned`: the platform owns it, and an app-owned one would
  conflict with it.

## Examples

**Input:** Generated `deploy.yml` (k3s mode) dispatched by `post-merge-deploy.yml` after a marked
merge — the default, recommended path for every deploy, including the very first one to a new
environment (see Gotchas: a manual local first deploy exercises a different code path than
production and has hidden real defects, e.g. an image-tag casing bug, in past live use).

**Output:** The prod machine's row resolved from `K3S_TARGETS`; `KUBECONFIG` written from the
secret of the GitHub Environment named after that machine slug; the `build` job's pushed image
reference and this same guard → lint → apply → rollout → `Certificate` → verify sequence run
against `app-<app-slug>`; verification passes over `https://<prod-host>` for a public own-tls row.

**Input:** `provisioning_mode: existing-k3s` project's Phase 4 dev sweep against a test machine
whose row is `private-mesh` + `platform-owned` (mesh-only, behind the platform's SSO gate).

**Output:** `mesh-check` confirms the API's mesh name answers on 6443; kubeconfig loaded from
`secret/apps/<app-slug>/<test-machine-slug>/kubeconfig`; `namespace-guard.sh` passes;
`manifest-lint.sh` passes; Deployment and Service applied, no Ingress; rollout completes; no
`Certificate` wait; `deploy-verify` runs the probe Job, which succeeds on the health path, and
confirms the Deployment runs this commit's image.

## Gotchas

- **Never widen scope to work around a guard refusal.** `namespace-guard.sh` failing signals the
  onboarding contract or the supplied kubeconfig is wrong, or the API is unreachable — not an
  obstacle to route around.
- **`kubectl auth can-i <verb> pods/exec` asks the wrong question.** kubectl parses its argument
  as `TYPE`, `TYPE/NAME`, or a non-resource URL, so `pods/exec` means "a pod named exec", and any
  deployer allowed to create pods gets `yes`. Name a subresource with `--subresource`:
  `kubectl auth can-i create pods --subresource=exec`. `can-i` does not check that the
  subresource exists, so a misspelled one answers silently — copy the exact flag.
- **`bash -e` ends a check on the answer it was waiting for.** GitHub runs every `run:` step with
  `bash -e`, so `ANSWER="$(kubectl auth can-i …)"` (exit 1 means "no") or a `kubectl exec` that
  must be refused ends the step before the answer is evaluated. Bracket the command with
  `set +e` … `RC=$?` … `set -e` and decide from the captured output and exit code, as
  `namespace-guard.sh` and `exec-refused.sh` do.
- **An unreachable API used to pass the guard.** `kubectl auth can-i` exits non-zero both for
  "no" and for a connection error; the guard now accepts only an explicit "no", and checks that
  the API answers first. A private-mesh machine reached before the mesh is joined fails here.
- **Verification is never log-based.** The probe Job's status, or the HTTPS checks, decide
  success. The platform decides whether the scoped credential may read `pods/log` (some grant it
  read-only, some don't); use it only to diagnose a failure, and never ask for `pods/exec`. A Role
  change on the platform side may not reach an app that was onboarded before it — ask the
  platform team rather than assuming a permission is there.
- **The probe Job needs two platform grants.** `create`/`delete` on `jobs` in the Namespace, and a
  NetworkPolicy that lets a pod in the Namespace reach the app's Service. A platform whose
  default-deny policy admits only its ingress controller will time the probe out; ask the
  platform team to allow same-Namespace traffic rather than relabeling the probe to look like the
  app (the Service would then route to it).
- **Changing an environment from `own-tls` to another mode leaves cert-manager's `Certificate`
  behind** — removing the issuer annotation does not delete it (cert-manager issue #8733). Delete
  `certificate/<app-slug>-tls` and its secret once the new Ingress serves.
- **The Ingress templates are shared, not per-app.** Always render from
  `assets/ingress.yaml.template` or `assets/ingress-default-tls.yaml.template`; a hand-written
  Ingress risks drifting from the platform's own ingress-class/`ClusterIssuer` conventions.
- **Deploy stamp is pod environment, not a file.** Do not adapt the rsync family's
  `printf '%s' > file` pattern here — there is no host filesystem this credential can write to.
  Set `DEPLOY_SHA`/`DEPLOY_TIME` as container env vars in the Deployment manifest at apply time.
- **`kubectl` version skew.** Confirm the local/CI `kubectl` minor version is within the
  supported skew of the target cluster before applying.
- **Kubeconfig rotation is the platform's.** Only the platform issues a kubeconfig, and how it
  rotates one is its decision — some reissue in place and revoke the old token, leaving
  workloads untouched; some only by re-onboarding. Never regenerate one locally: replace it in
  Vault with `smaqit.infrastructure-vault-loader`'s `rotate-credential.sh
  apps/<app-slug>/<machine-slug>/kubeconfig`, re-sync the GitHub Environment
  (`smaqit.infrastructure-repo-config`), and redeploy with a **fresh** workflow dispatch — a
  re-run of an earlier run keeps the secret values from when that run was created.
- **Prefer the generated CI workflow over a manual local invocation, even for the first deploy to
  a brand-new environment.** A manual local run exercises a different code path (this skill's own
  shell commands) than what actually runs in production (the generated `deploy.yml`) — in past
  live use, that difference hid a real defect (an image-tag casing bug that a local shell would
  never have hit) until the workflow path was used instead. Local invocation remains a valid
  fallback when no CI is configured yet, not something to reach for by default when it is.

## Completion

- [ ] `mesh-check`: PASS (public, or the mesh name answers on the API port)
- [ ] Target environment's Environments row read with `smaqit spec environments --env=<environment> --json`
- [ ] Kubeconfig loaded through `credential-file.sh` (local) or the machine-slug Environment's `KUBECONFIG` secret (CI/CD)
- [ ] `namespace-guard.sh` passed
- [ ] `imagePullSecret` created only when a private registry is declared
- [ ] The Ingress file for each Ingress mode in use, and `probe-job.yaml` when any environment verifies in-cluster, rendered from the shared templates and committed with their per-environment tokens left in; this run's values stamped into temporary copies
- [ ] `manifest-lint.sh` passed before any `kubectl apply`
- [ ] `kubectl apply -n app-__APP_SLUG__` succeeded (`pvc.yaml` first when present; no Ingress for `platform-owned`)
- [ ] `kubectl rollout status` succeeded
- [ ] `exec-refused.sh` passed (exec refused as Forbidden, or permitted with a warning under `Deployer Exec: allowed`)
- [ ] `Certificate` reached `Ready` (`own-tls` only)
- [ ] `smaqit.infrastructure-deploy-verify` invoked for the environment with the expected SHA, and passed

## Failure Handling

| Situation | Action |
|-----------|--------|
| `smaqit spec environments` reports `SLOTS-SPEC-*` lines | Stop and surface them verbatim. The fix belongs in the Infrastructure spec, with values from the platform team — never a guessed one. |
| `mesh-check` fails (mesh name doesn't resolve, or the API port doesn't answer) | Stop. Join the platform's mesh on this machine and run the slot again. Do not reach the API another way. |
| Kubeconfig secret absent for the target environment | Stop. A merged onboarding request is not enough: the platform's own hand-off must deliver the kubeconfig, then the `target-credentials` load stores it (`PROVISIONING_MODE=existing-k3s MACHINE_SLUG=<machine-slug> load-credentials.sh`). Point at the platform's onboarding playbook. Never attempt a cluster-scoped fix. |
| `namespace-guard.sh` refuses (API unreachable, check undetermined, cluster-scoped verb permitted, exec into pods permitted, or Namespace mismatch) | Stop immediately. For an unreachable API on a `private-mesh` target, join the mesh first. For an exec grant, ask the platform team to remove `pods/exec` from the deployer's Role; set `Deployer Exec: allowed` only when the operator confirms the grant is deliberate. Do not proceed with a broader credential than the onboarding contract issues. |
| `exec-refused.sh` fails | Stop. A real exec went through (fix the Role as above), or the attempt gave no clear answer (NotFound, timeout): check the Deployment name and the API connection, then run again. |
| A `kubectl auth can-i` check the project added answers `yes` for something the platform denies (typically exec into pods) | The check almost certainly used the `TYPE/NAME` form (`can-i create pods/exec`), which kubectl reads as "a pod named exec". Rewrite it with `--subresource=<name>` and run it again. Trust `namespace-guard.sh`'s answer, which already uses that form. Never widen scope or skip a check on the strength of the misleading `yes`. |
| A `--subresource` check the project added answers `no` for everything, including grants the platform documents | `can-i` does not validate subresource names, so a misspelled one (`--subresource=exe`) silently answers `no`. Check the spelling against the Kubernetes API reference before trusting the result. |
| A check step the project added ends at a command expected to exit non-zero (a `can-i` "no", a refused exec) before it reports a verdict | `bash -e` stopped the step on the answer it was waiting for. Bracket the command with `set +e` … `RC=$?` … `set -e` and decide from the captured output and exit code, or call `exec-refused.sh`. Never append `\|\| true`: it also hides a real error, such as an unreachable API, and turns it into a PASS. |
| `manifest-lint.sh` rejects a manifest | Stop before any `kubectl apply`. Report the exact offending field. Fix the manifest, not the check. |
| `kubectl apply` fails (admission rejection) | Report the admission error verbatim. Do not retry with `--force` or widened scope. |
| `kubectl rollout status` times out | Report the timeout and the last known state from `kubectl get deployment`/`kubectl get events`. If the platform grants `pods/log`, include the failing container's `kubectl logs`; never `pods/exec`. |
| `Certificate` never reaches `Ready` (`own-tls`) | Check the `ClusterIssuer` name and the issuer's challenge reachability against the Environments row; do not bypass the wait. |
| Probe Job cannot be created (`jobs` forbidden) | Stop. Ask the platform team to grant `create`/`delete` on `jobs` in the Namespace. Do not fall back to external checks the deployer cannot reach. |
| Probe Job fails or times out | Report the Job status; include its logs only if `pods/log` is granted. Check the health path against the spec and whether a NetworkPolicy blocks same-Namespace traffic. |
| `smaqit.infrastructure-deploy-verify` still reports FAIL after the 60-second retry window | Stop. Report the failing check. Do not mark the deployment complete. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |
| Output artifact already exists | Confirm with user before overwriting. |

## Allowed Tools

Bash(kubectl:*), Bash(bash:*), Bash(smaqit:*), Bash(jq:*)
