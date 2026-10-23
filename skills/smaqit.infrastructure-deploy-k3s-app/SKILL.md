---
name: smaqit.infrastructure-deploy-k3s-app
description: Use when deploying an app into a Namespace-scoped Kubernetes cluster (k3s, or any cluster issuing a per-app kubeconfig), provisioning_mode existing-k3s. Deploys run only in the generated deploy.yml, which joins a private mesh, guards the credential, lints manifests against Pod Security restricted, applies the Ingress the spec's mode calls for, waits for rollout, and verifies.
metadata:
  version: "3.0.0"
  smaqit-slots: "app-deploy=deploy"
---

# Deploy Application to a Kubernetes Namespace

An existing-k3s deploy runs only in the CI/CD pipeline: the generated `deploy.yml`
(`smaqit.infrastructure-cicd-generate`'s `k3s` mode) performs every step below, and nothing acts on
the target from the operator's machine. `smaqit slots plan` refuses a local `app-deploy` for
existing-k3s (`SLOTS-LOCAL-DEPLOY-REFUSED`), so the Deployment agent only triggers and monitors the
run. This skill is the contract that workflow implements: what it requires, the order it runs in,
and what to do when a run fails.

Everything that differs between targets comes from the Infrastructure spec's `## Environments`
row for the target environment, read through `smaqit spec environments --env=<environment>
--json` — never from guesses or binding `params`:

| Field | Values | Effect |
|-------|--------|--------|
| `reachability` | `public` / `private-mesh` | `private-mesh`: the API (and the app) are reachable only from mesh peers; the workflow joins the mesh before any `kubectl` call |
| `ingressMode` | `own-tls` / `platform-default-tls` / `platform-owned` | which Ingress is applied, if any, and whether a `Certificate` is awaited |
| `host`, `ingressClass`, `clusterIssuer`, `healthPath` | from the platform team | stamped into the Ingress / probe at apply time |

The same JSON carries three spec-wide values: `probeClient` (`curl`, `wget`, or `python3`),
`registry` (`{host, visibility}`, or `null` for a public image), and `deployerExec` (`denied`
unless the spec's `Deployer Exec` row says `allowed`).

Verification runs **inside the cluster** (a probe Job) when the runner cannot reach the app
from outside — `private-mesh`, or `platform-owned` (for example behind the platform's SSO gate)
— and **over HTTPS** otherwise.

## Pre-conditions

- The target Namespace (`app-__APP_SLUG__`, on each target environment's machine) has already
  been onboarded by the platform side (Namespace, RBAC, Pod Security label, NetworkPolicy,
  ResourceQuota/LimitRange, and the scoped kubeconfig already issued). This skill's job starts
  after onboarding — nothing here is cluster-scoped.
- The GitHub Environment named after the target machine slug holds the `KUBECONFIG` secret (and,
  for a `private-mesh` row, the mesh setup-key secret the row names), synced by the
  `environment-config` slot (`smaqit.infrastructure-repo-config`). A merged onboarding request
  alone doesn't provide the kubeconfig — the platform's own hand-off delivers it, the
  `target-credentials` load stores it, and the sync copies it into the Environment.
- `smaqit spec environments --env=<environment> --json` succeeds: the Infrastructure spec's
  `## Environments` row for the target is complete. `smaqit slots plan --mode=existing-k3s`
  refuses to plan otherwise.
- The generated `deploy.yml` is on the default branch: GitHub dispatches only workflows that
  exist there.
- The app's manifests are committed under `deployment/k3s/`: `deployment.yaml` and `service.yaml`
  (authored by the `artifact-build` slot), `pvc.yaml` when the app keeps state, and the Ingress
  file for each Ingress mode in use plus `probe-job.yaml` when any environment verifies in-cluster
  (rendered from this skill's shared templates by `smaqit.infrastructure-cicd-generate`'s `k3s`
  mode; see Shared Templates).

## Operations

### `deploy` (slot `app-deploy`, performed by CI)

The `deploy` job of the generated `deploy.yml` runs this sequence, one machine at a time, after
its `build` job pushed the commit's image. The agent dispatches it for the environment's machine
slug (`gh workflow run deploy.yml --ref <default-branch> -f environment=<machine-slug>`), selects
the run by commit, watches it, and then runs the `verify` slot, which reads this run's record
(`smaqit.infrastructure-deploy-verify` `ci-evidence`).

1. **Resolve the target.** The dispatched machine slug selects its row from the workflow's
   `K3S_TARGETS` map, rendered from the Environments table; the job runs in the GitHub
   Environment of that name.

2. **Join the mesh (`private-mesh` only)** with the pinned action and setup-key secret the spec's
   Constraints rows name, then wait for the API's mesh name to resolve.

3. **Write the kubeconfig** from the Environment's `KUBECONFIG` secret, owner-only, to
   `$RUNNER_TEMP`; an `always()` step removes it at the end.

4. **Run the namespace guard before touching anything else:**
   ```bash
   deployment/k3s/scripts/namespace-guard.sh "$KUBECONFIG" "app-${APP_SLUG}" --deployer-exec <deployerExec>
   ```
   It confirms the API answers and accepts the credential, that the kubeconfig's current context
   targets exactly `app-__APP_SLUG__`, that no cluster-scoped verb is permitted, and that the
   deployer cannot exec into pods (`kubectl auth can-i create|get pods --subresource=exec` must
   both answer `no`). It fails closed: only an explicit "no" counts as denied, so an unreachable
   API is a FAIL, not a PASS. Under `--deployer-exec allowed` an exec grant is a warning instead
   of a failure; an unclear answer still fails.

5. **Create an `imagePullSecret`, only if the spec's `registry` is private.** The credential
   comes from the app's own Environment secrets through the step's `env:` and reaches `kubectl`
   through a file descriptor, never as a command-line argument.

6. **Stamp this run's values** (host, Ingress class, ClusterIssuer) into a temporary copy of the
   committed Ingress for this mode — never into the committed file.

7. **Lint every manifest before any apply** — `pvc.yaml` when present, the Ingress for this mode
   (none for `platform-owned`), and the stamped probe Job when this run verifies in-cluster:
   ```bash
   deployment/k3s/scripts/manifest-lint.sh deployment/k3s/deployment.yaml deployment/k3s/service.yaml "$INGRESS"
   ```
   Checks the Pod Security `restricted` fields (`runAsNonRoot: true`, an explicit numeric
   `runAsUser` wherever `runAsNonRoot` is true, `allowPrivilegeEscalation: false`,
   `capabilities.drop: [ALL]`, `seccompProfile.type: RuntimeDefault`, no `privileged: true`, no
   `hostPath` volumes, no `hostNetwork`/`hostPID`/`hostIPC`), explicit CPU/memory
   `requests`/`limits` on every container, and the absence of any cluster-scoped `kind` in the
   manifest set. A rejected manifest names the offending field — fix the manifest, not the check.
   A base image whose own `USER` directive names a user rather than a numeric UID is the most
   common way `runAsNonRoot` alone still fails admission with `CreateContainerConfigError`.

8. **Apply into the app's own Namespace only**, `pvc.yaml` first when present, then the
   Deployment, the Service, and this mode's Ingress (none for `platform-owned`). `DEPLOY_SHA` and
   `DEPLOY_TIME` are set as pod environment (see `assets/deployment.yaml.template`) from the run's
   commit SHA and an ISO 8601 UTC timestamp — never a file on a host path.

9. **Wait for rollout, then prove exec is refused:**
   ```bash
   kubectl rollout status -n "app-${APP_SLUG}" "deployment/${APP_SLUG}" --timeout=180s
   deployment/k3s/scripts/exec-refused.sh "app-${APP_SLUG}" "${APP_SLUG}" --deployer-exec <deployerExec>
   ```
   `exec-refused.sh` attempts a real `kubectl exec` into the Deployment and passes only when the
   API refuses it as Forbidden. An exec that goes through fails under `denied` and warns under
   `allowed`; NotFound, a timeout, or any other answer fails closed.

10. **`own-tls` only — wait for the `Certificate` to reach `Ready`.** `platform-default-tls` and
    `platform-owned` request no certificate.

11. **Verify the revision and health.** "Verify revision" checks that the Deployment's image tag
    is the run's commit SHA. Then the row's verify step runs the probe Job in-cluster
    (`private-mesh` or `platform-owned`) or checks the health path over HTTPS, retrying every 5
    seconds for up to 60 seconds. The `verify` slot later reads both steps' conclusions from the
    run's record.

## Shared Templates

Manifests are rendered from these templates — never hand-written — once, and committed.
`smaqit.infrastructure-cicd-generate`'s `k3s` mode renders the files the spec's environments need
when it generates `deploy.yml`, substituting only `__APP_SLUG__` (and, for the probe,
`__RUN_AS_USER__` and `__PROBE_FETCH__`); the per-environment tokens stay in the committed files and
the workflow stamps them at apply time:

| Needed when some environment is… | Template | Committed as |
|---|---|---|
| `own-tls` | `assets/ingress.yaml.template` | `deployment/k3s/ingress.yaml` |
| `platform-default-tls` | `assets/ingress-default-tls.yaml.template` | `deployment/k3s/ingress-default-tls.yaml` |
| `private-mesh` or `platform-owned` | `assets/probe-job.yaml.template` | `deployment/k3s/probe-job.yaml` |

`__RUN_AS_USER__` is the numeric `runAsUser` of the committed `deployment.yaml`. `__PROBE_FETCH__`
follows the spec's `Probe Client`:

| Probe Client | Fetch command |
|---|---|
| `curl` | `curl -fsS -o /dev/null` |
| `wget` | `wget -q -O /dev/null` |
| `python3` | `python3 -c 'import sys, urllib.request; urllib.request.urlopen(sys.argv[1], timeout=5)'` |

Pick the client the app's own image already contains (a `python:*-slim` image has `python3` but
neither `curl` nor `wget`); never add an OS package to the production image only for the probe.
`platform-owned` needs no Ingress file at all.

`scripts/namespace-guard.sh`, `scripts/manifest-lint.sh`, and `scripts/exec-refused.sh` are vendored
into `deployment/k3s/scripts/` by the same generation.

## Output

`deploy` (in the CI run): application manifests applied into `app-__APP_SLUG__` (Deployment,
Service, a PVC when committed, and the Ingress this mode calls for), rollout complete,
`Certificate` `Ready` for `own-tls`, deploy stamp present as pod environment, and the run's
"Verify revision" and verify steps successful — the Deployment runs this commit's image, and the
health path answers in-cluster or over HTTPS.

## Scope

- Does NOT deploy from the operator's machine: there is no local invocation, and no break-glass
  path. A deploy that cannot run in CI is a workflow or configuration problem to fix.
- Does NOT onboard the Namespace, RBAC, Pod Security label, NetworkPolicy, or quota — that is
  the platform side's own skill; this skill's precondition, not its job.
- Does NOT provision the cluster itself — a separate, unrelated concern.
- Does NOT rely on pod logs for verification — the verdict is the probe Job's status or the HTTPS
  checks. Whether the scoped credential may read `pods/log` is the platform's decision; when it
  can, a failed rollout's or probe's logs may be read for diagnosis. Never `pods/exec`:
  `namespace-guard.sh` and `exec-refused.sh` fail the deploy when the deployer could exec, unless
  the spec's `Deployer Exec` row records a deliberate grant.
- Does NOT touch anything cluster-scoped — `namespace-guard.sh` refuses to proceed if the
  kubeconfig's own permissions would even allow it.
- Does NOT request onboarding on the app's behalf — if the kubeconfig secret is absent, the
  deploy stops and points at the platform's own onboarding playbook.
- Does NOT ship an Ingress for `platform-owned`: the platform owns it, and an app-owned one would
  conflict with it.

## Examples

**Input:** `provisioning_mode: existing-k3s` project's first deploy (`ci-bootstrap`) to a test
machine whose row is `private-mesh` + `platform-owned` (mesh-only, behind the platform's SSO gate).

**Output:** The workflows and manifests land through a PR without the deploy marker; after the
merge the agent dispatches `deploy.yml` with `-f environment=<test-machine-slug>`. The run joins
the mesh, writes `KUBECONFIG` from that machine's Environment, passes `namespace-guard.sh` and
`manifest-lint.sh`, applies the Deployment and Service (no Ingress), completes the rollout, skips
the `Certificate` wait, and its probe Job succeeds on the health path; "Verify revision" confirms
this commit's image. `ci-evidence` then passes against the merge SHA.

**Input:** A later production deploy (`ci-push` or `existing-cicd-pr`) to a `public` + `own-tls` row.

**Output:** The prod machine's row resolved from `K3S_TARGETS`; the same guard → lint → apply →
rollout → `Certificate` → verify sequence runs against `app-<app-slug>`; verification passes over
`https://<prod-host>`.

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
  "no" and for a connection error; the guard accepts only an explicit "no", and checks that the
  API answers first. A private-mesh machine reached before the mesh join completes fails here.
- **Verification is never log-based.** The probe Job's status, or the HTTPS checks, decide
  success. The platform decides whether the scoped credential may read `pods/log`; use it only to
  diagnose a failure, and never ask for `pods/exec`. A Role change on the platform side may not
  reach an app that was onboarded before it — ask the platform team rather than assuming a
  permission is there.
- **The probe Job needs two platform grants.** `create`/`delete` on `jobs` in the Namespace, and a
  NetworkPolicy that lets a pod in the Namespace reach the app's Service. A platform whose
  default-deny policy admits only its ingress controller will time the probe out; ask the
  platform team to allow same-Namespace traffic rather than relabeling the probe to look like the
  app (the Service would then route to it).
- **Changing an environment from `own-tls` to another mode leaves cert-manager's `Certificate`
  behind** — removing the issuer annotation does not delete it (cert-manager issue #8733). Delete
  `certificate/<app-slug>-tls` and its secret once the new Ingress serves.
- **The Ingress templates are shared, not per-app.** A hand-written Ingress risks drifting from
  the platform's own ingress-class/`ClusterIssuer` conventions; regenerate through
  `smaqit.infrastructure-cicd-generate` instead.
- **Deploy stamp is pod environment, not a file.** Do not adapt the rsync family's
  `printf '%s' > file` pattern here — there is no host filesystem this credential can write to.
- **`kubectl` version skew.** The runner's `kubectl` minor version must be within the supported
  skew of the target cluster.
- **Kubeconfig rotation is the platform's.** Only the platform issues a kubeconfig, and how it
  rotates one is its decision — some reissue in place and revoke the old token, leaving
  workloads untouched; some only by re-onboarding. Never regenerate one locally: replace it in
  Vault with `smaqit.infrastructure-vault-loader`'s `rotate-credential.sh
  apps/<app-slug>/<machine-slug>/kubeconfig`, re-sync the GitHub Environment
  (`smaqit.infrastructure-repo-config`), and redeploy with a **fresh** workflow dispatch — a
  re-run of an earlier run keeps the secret values from when that run was created.
- **Why there is no local deploy.** A local run exercises different commands than the workflow
  that deploys production; in past live use that difference hid a real defect (an image-tag
  casing bug a local shell never hit) until the workflow ran.

## Completion

- [ ] Target environment's Environments row complete (`smaqit spec environments --env=<environment> --json`)
- [ ] `deployment.yaml`, `service.yaml`, and the Ingress and probe manifests the environments need committed; the generated `deploy.yml` on the default branch
- [ ] The machine-slug Environment holds `KUBECONFIG` (and the mesh setup key for a `private-mesh` row)
- [ ] The dispatched `deploy.yml` run, selected by commit, completed successfully: mesh join (private-mesh), namespace guard, lint, apply, rollout, exec refused, `Certificate` (own-tls), "Verify revision", and the row's verify step
- [ ] The `verify` slot's `ci-evidence` passed against that run

## Failure Handling

| Situation | Action |
|-----------|--------|
| `smaqit spec environments` reports `SLOTS-SPEC-*` lines | Stop and surface them verbatim. The fix belongs in the Infrastructure spec, with values from the platform team — never a guessed one. |
| `smaqit slots plan` reports `SLOTS-LOCAL-DEPLOY-REFUSED` | Stop. Plan the deploy on a CI path (`--path=ci-bootstrap` for the first deploy, then `ci-push` or `existing-cicd-pr`); never deploy from this machine. |
| The mesh join step fails, or the API's mesh name never resolves | Check the setup-key secret name and the mesh action's inputs against the spec's Constraints rows, and that the platform issued a valid key for this machine; re-sync the Environment and dispatch again. |
| `KUBECONFIG` secret absent from the machine-slug Environment | Stop. A merged onboarding request is not enough: the platform's own hand-off must deliver the kubeconfig, the `target-credentials` load stores it (`PROVISIONING_MODE=existing-k3s MACHINE_SLUG=<machine-slug> load-credentials.sh`), and `environment-config` syncs it. Point at the platform's onboarding playbook. Never attempt a cluster-scoped fix. |
| `namespace-guard.sh` refuses (API unreachable, check undetermined, cluster-scoped verb permitted, exec into pods permitted, or Namespace mismatch) | Stop. For an unreachable API on a `private-mesh` target, check the mesh join step. For an exec grant, ask the platform team to remove `pods/exec` from the deployer's Role; set `Deployer Exec: allowed` only when the operator confirms the grant is deliberate. Do not proceed with a broader credential than the onboarding contract issues. |
| `exec-refused.sh` fails | Stop. A real exec went through (fix the Role as above), or the attempt gave no clear answer (NotFound, timeout): check the Deployment name and the API connection, then dispatch again. |
| A `kubectl auth can-i` check the project added answers `yes` for something the platform denies (typically exec into pods) | The check almost certainly used the `TYPE/NAME` form (`can-i create pods/exec`), which kubectl reads as "a pod named exec". Rewrite it with `--subresource=<name>`. Trust `namespace-guard.sh`'s answer, which already uses that form. Never widen scope or skip a check on the strength of the misleading `yes`. |
| A `--subresource` check the project added answers `no` for everything, including grants the platform documents | `can-i` does not validate subresource names, so a misspelled one (`--subresource=exe`) silently answers `no`. Check the spelling against the Kubernetes API reference before trusting the result. |
| A check step the project added ends at a command expected to exit non-zero (a `can-i` "no", a refused exec) before it reports a verdict | `bash -e` stopped the step on the answer it was waiting for. Bracket the command with `set +e` … `RC=$?` … `set -e` and decide from the captured output and exit code, or call `exec-refused.sh`. Never append `\|\| true`: it also hides a real error, such as an unreachable API, and turns it into a PASS. |
| `manifest-lint.sh` rejects a manifest | Report the exact offending field. Fix the manifest, not the check, then dispatch again. |
| `manifest-lint.sh` fails with `FileNotFoundError` on an Ingress or `probe-job.yaml` | The manifest was never rendered. Regenerate through `smaqit.infrastructure-cicd-generate`'s `k3s` mode and commit the result; never fabricate a placeholder host. |
| `kubectl apply` fails (admission rejection) | Report the admission error verbatim. Do not retry with `--force` or widened scope. |
| `kubectl rollout status` times out | Report the timeout and the run's last known Deployment state. If the platform grants `pods/log`, include the failing container's logs; never `pods/exec`. |
| `Certificate` never reaches `Ready` (`own-tls`) | Check the `ClusterIssuer` name and the issuer's challenge reachability against the Environments row; do not bypass the wait. |
| Probe Job cannot be created (`jobs` forbidden) | Stop. Ask the platform team to grant `create`/`delete` on `jobs` in the Namespace. Do not fall back to checks the runner cannot reach. |
| Probe Job fails or times out | Report the Job status; include its logs only if `pods/log` is granted. Check the health path against the spec and whether a NetworkPolicy blocks same-Namespace traffic. |
| The run's verify step still fails after the 60-second retry window | Stop. Report the failing check. Do not mark the deployment complete. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |

## Allowed Tools

Bash(gh:*), Bash(smaqit:*), Bash(jq:*)
