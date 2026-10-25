---
name: smaqit.infrastructure-deploy-k3s-app
description: Use when deploying an app into a Namespace-scoped Kubernetes cluster (k3s, or any cluster that trusts the CI job's OIDC identity), provisioning_mode existing-k3s. Deploys run only in the generated deploy.yml, which joins a mesh, connects through the platform's action, guards the credential, lints manifests against Pod Security restricted, applies the Ingress, waits for rollout, and verifies.
metadata:
  version: "4.2.0"
  smaqit-slots: "app-deploy=deploy"
---

# Deploy Application to a Kubernetes Namespace

An existing-k3s deploy runs only in the CI/CD pipeline: the generated `deploy.yml`
(`smaqit.infrastructure-cicd-generate`'s `k3s` mode) performs every step below, and nothing acts on
the target from the operator's machine. `smaqit slots plan` refuses a local `app-deploy` for
existing-k3s (`SLOTS-LOCAL-DEPLOY-REFUSED`); this skill's `deploy` operation is what the Deployment
agent does instead: it triggers and monitors that run. The run authenticates with its own GitHub Actions OIDC identity through the platform's connect
action: the platform binds the app's Namespace to this repository and the machine-slug GitHub
Environment the job runs in, holds the mesh key, and pulls the images, so the app stores no deploy
credential at all. This skill is smaqit's default `app-deploy` filler: its operation (trigger and
monitor), the deploy run it watches (what the workflow requires and the order it runs in), and what
to do when a run fails. A platform can bind its own `app-deploy` filler in `.smaqit/slots.yml`; the
contract it must meet is `smaqit slots contract --slot=app-deploy --mode=existing-k3s`.

Everything that differs between targets comes from the Infrastructure spec's `## Environments`
row for the target environment, read through `smaqit spec environments --env=<environment>
--json` — never from guesses or binding `params`:

| Field | Values | Effect |
|-------|--------|--------|
| `reachability` | `public` / `private-mesh` | the app's reachability: `private-mesh` apps are reachable only from mesh peers, so verification runs in-cluster |
| `meshSetupKeySecret`, `meshWaitForHost` | the platform's mesh secret and the API's mesh name | set when the API is reachable only from mesh peers (always for `private-mesh`; for `public` when the platform keeps its API mesh-only): the workflow joins the mesh before any `kubectl` call |
| `ingressMode` | `own-tls` / `platform-default-tls` / `platform-owned` | which Ingress is applied, if any, and whether a `Certificate` is awaited |
| `host`, `ingressClass`, `clusterIssuer`, `healthPath` | from the platform team | stamped into the Ingress / probe at apply time |

The same JSON carries the spec-wide values: `connect` (the platform's `Cluster Connect` action and
its machine and namespace input names), `mesh` (the mesh-join action and its input names),
`probeClient` (`curl`, `wget`, or `python3`), `deployerExec` (`denied` unless the spec's
`Deployer Exec` row says `allowed`), `vaultAccess`, and `runtimeSecrets` (the spec's
`## Runtime Secrets` table; each row also lists its own `runtimeSecrets` names).

Verification runs **inside the cluster** (a probe Job) when the runner cannot reach the app
from outside — `private-mesh`, or `platform-owned` (for example behind the platform's SSO gate)
— and **over HTTPS** otherwise.

## Pre-conditions

- The target Namespace (`app-__APP_SLUG__`, on each target environment's machine) has already
  been onboarded by the platform side (Namespace, RBAC, Pod Security label, NetworkPolicy,
  ResourceQuota/LimitRange) and its deployer Role bound to this repository and the machine-slug
  GitHub Environment. This skill's job starts after onboarding — nothing here is cluster-scoped.
- The Infrastructure spec declares the platform's `Cluster Connect` rows. GitHub creates the
  machine-slug Environment on the first run that references it; it holds no deploy secret, only the
  app's runtime secrets when the spec declares any (the `environment-config` slot syncs them). For a
  row that joins a mesh, the platform shares the mesh secret the row names with this repository
  (for example an organisation secret).
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

### `deploy` (slot `app-deploy`): trigger and monitor the deploy run

The Deployment agent follows this at the `app-deploy` slot, after the path's own commit step (the
merged PR on `ci-bootstrap` and `existing-cicd-pr`, the pushed default branch on `ci-push`). Inputs:
the deployment path, the environment's machine slug, and the intended commit (the merge SHA, or the
pushed SHA). The deploy itself runs in the workflow (The Deploy Run, below); nothing here touches
the target.

1. **Trigger the run.**
   - `ci-bootstrap` and `ci-push`: note the time, then dispatch for this environment only:
     ```bash
     gh workflow run deploy.yml --ref <default-branch> -f environment=<machine-slug>
     ```
   - `existing-cicd-pr`: the merged PR's marker makes `post-merge-deploy.yml` dispatch it. Select
     that dispatcher's run for the merge commit
     (`gh run list --workflow post-merge-deploy.yml --commit <merge-sha> --json databaseId,status`),
     watch it with `gh run watch <id> --exit-status`, and note when it finished.
2. **Select the deploy run.** List the dispatched runs on the default branch:
   ```bash
   gh run list --workflow deploy.yml --branch <default-branch> --event workflow_dispatch \
     --json databaseId,headSha,status,createdAt
   ```
   Keep the runs created after the trigger. Exactly one is the deploy run. None yet: wait 10
   seconds and list again, for up to two minutes. More than one: stop and report their IDs — never
   pick one.
3. **Watch it** with `gh run watch <run-id> --exit-status`. Never a bare `gh run watch`, which
   follows whatever run is newest.
4. **Hand off** the run's ID and conclusion. A failed run stops the walk: surface the run ID and
   its failed step. A successful one goes to the `verify` slot (`smaqit.infrastructure-deploy-verify`
   `ci-evidence`, with this run ID and the intended commit), which proves the run's head is that
   commit or a later one containing it.

## The Deploy Run

The `deploy` job of the generated `deploy.yml` runs this sequence, one machine at a time, after
its `build` job pushed the commit's image.

1. **Resolve the target.** The dispatched machine slug selects its row from the workflow's
   `K3S_TARGETS` map, rendered from the Environments table; the job runs in the GitHub
   Environment of that name.

2. **Join the mesh (when the row names a mesh secret)** with the pinned action the spec's
   Constraints rows name and the platform's shared key, then wait for the API's mesh name.

3. **Connect to the cluster** through the spec's pinned `Cluster Connect Action`, passing the
   machine slug and `app-__APP_SLUG__`. The job holds `id-token: write` (and `contents: read`);
   the action requests the job's OIDC token, writes a kubeconfig whose context pins
   `app-__APP_SLUG__`, refreshes the short-lived token through an exec credential plugin, and
   exports `KUBECONFIG`. An `always()` step removes that file at the end.

4. **Run the namespace guard before touching anything else:**
   ```bash
   deployment/k3s/scripts/namespace-guard.sh "$KUBECONFIG" "app-${APP_SLUG}" --deployer-exec <deployerExec>
   ```
   It confirms the API answers and accepts the identity, that the kubeconfig's current context
   targets exactly `app-__APP_SLUG__`, that no cluster-scoped verb is permitted, and that the
   deployer cannot exec into pods (`kubectl auth can-i create|get pods --subresource=exec` must
   both answer `no`). It fails closed: only an explicit "no" counts as denied, so an unreachable
   API is a FAIL, not a PASS — after a bounded retry, because a freshly joined mesh can resolve
   the API's name before its port answers. An Unauthorized or Forbidden answer is never retried.
   Under `--deployer-exec allowed` an exec grant is a warning instead of a failure; an unclear
   answer still fails.

5. **Stamp this run's values** (host, Ingress class, ClusterIssuer) into a temporary copy of the
   committed Ingress for this mode — never into the committed file.

6. **Lint every manifest before any apply** — `pvc.yaml` when present, the Ingress for this mode
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
   When the spec declares runtime secrets, it also checks that the Deployment references every
   runtime secret this machine receives with its own `secretKeyRef` (the env var has the
   secret's name), that a reference to another environment's secret sets `optional: true`, and
   that the Secret is never pulled in with `envFrom`.

   **Runtime secrets only — create the `__APP_SLUG__-runtime` Secret** from this machine's
   runtime secrets: each value reaches the step only through `env:` from the machine-slug
   Environment, `jq` builds the Secret from the environment, and `kubectl apply --server-side`
   applies it, so no value is on a command line or in a last-applied annotation. A value missing
   from the Environment fails the step before `kubectl` runs.

7. **Apply into the app's own Namespace only**, `pvc.yaml` first when present, then the
   Deployment, the Service, and this mode's Ingress (none for `platform-owned`). `DEPLOY_SHA` and
   `DEPLOY_TIME` are set as pod environment (see `assets/deployment.yaml.template`) from the run's
   commit SHA and an ISO 8601 UTC timestamp — never a file on a host path. No `imagePullSecrets`:
   the platform pulls the image with its own credential.

8. **Wait for rollout, then prove exec is refused:**
   ```bash
   kubectl rollout status -n "app-${APP_SLUG}" "deployment/${APP_SLUG}" --timeout=180s
   deployment/k3s/scripts/exec-refused.sh "app-${APP_SLUG}" "${APP_SLUG}" --deployer-exec <deployerExec>
   ```
   `exec-refused.sh` attempts a real `kubectl exec` into the Deployment and passes only when the
   API refuses it as Forbidden. An exec that goes through fails under `denied` and warns under
   `allowed`; NotFound, a timeout, or any other answer fails closed.

9. **`own-tls` only — wait for the `Certificate` to reach `Ready`.** `platform-default-tls` and
    `platform-owned` request no certificate.

10. **Verify the revision and health.** "Verify revision" checks that the Deployment's image tag
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
- Does NOT request onboarding on the app's behalf — an identity the platform hasn't bound yet
  answers Forbidden, and the deploy stops and points at the onboarding request.
- Does NOT hold any deploy credential — no kubeconfig, mesh key, or registry credential is stored
  or synced; the platform's connect action and shared secrets supply everything.
- Does NOT ship an Ingress for `platform-owned`: the platform owns it, and an app-owned one would
  conflict with it.

## Examples

**Input:** `provisioning_mode: existing-k3s` project's first deploy (`ci-bootstrap`) to a test
machine whose row is `private-mesh` + `platform-owned` (mesh-only, behind the platform's SSO gate).

**Output:** The workflows and manifests land through a PR without the deploy marker; after the
merge the agent dispatches `deploy.yml` with `-f environment=<test-machine-slug>`. The run joins
the mesh, connects through the platform's action with the job's OIDC identity, passes `namespace-guard.sh` and
`manifest-lint.sh`, applies the Deployment and Service (no Ingress), completes the rollout, skips
the `Certificate` wait, and its probe Job succeeds on the health path; "Verify revision" confirms
this commit's image. `ci-evidence` then passes against the merge SHA.

**Input:** A later production deploy (`ci-push` or `existing-cicd-pr`) to a `public` + `own-tls` row.

**Output:** The prod machine's row resolved from `K3S_TARGETS`; the same guard → lint → apply →
rollout → `Certificate` → verify sequence runs against `app-<app-slug>`; verification passes over
`https://<prod-host>`.

## Gotchas

- **Never widen scope to work around a guard refusal.** `namespace-guard.sh` failing signals the
  onboarding binding or the connect action's kubeconfig is wrong, or the API is unreachable — not
  an obstacle to route around.
- **Unauthorized and Forbidden mean different things.** Unauthorized (401): the platform did not
  accept the job's identity at all — the token's audience, the repository, or the GitHub
  Environment it ran in is not one it trusts (a dispatch to a non-machine-slug Environment answers
  this way). Forbidden (403): the identity was accepted but is not bound to the Namespace yet — the
  onboarding request has not merged or reconciled.
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
  API answers first. A mesh-only API reached before the mesh join completes fails here, after the
  guard's bounded retry.
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
- **A cluster CA change is a new connect-action tag.** The platform bundles the CA in its action,
  so after a machine replacement bump the spec's `Cluster Connect Action` pin and render
  `deploy.yml` again; nothing is rotated on the app side.
- **Why there is no local deploy.** A local run exercises different commands than the workflow
  that deploys production; in past live use that difference hid a real defect (an image-tag
  casing bug a local shell never hit) until the workflow ran.

## Completion

- [ ] Target environment's Environments row complete (`smaqit spec environments --env=<environment> --json`)
- [ ] `deployment.yaml`, `service.yaml`, and the Ingress and probe manifests the environments need committed; the generated `deploy.yml` on the default branch
- [ ] The spec declares the `Cluster Connect` rows; the `deploy` job holds `id-token: write` and `contents: read`
- [ ] The deploy run triggered for this environment only, selected by trigger time and commit (exactly one), and watched by its ID
- [ ] That run completed successfully: mesh join (when the row names a mesh secret), connect, namespace guard, lint, apply, rollout, exec refused, `Certificate` (own-tls), "Verify revision", and the row's verify step
- [ ] Its run ID handed to the `verify` slot, whose `ci-evidence` passed against it

## Failure Handling

| Situation | Action |
|-----------|--------|
| `smaqit spec environments` reports `SLOTS-SPEC-*` lines | Stop and surface them verbatim. The fix belongs in the Infrastructure spec, with values from the platform team — never a guessed one. |
| More than one deploy run started after the trigger | Stop and report the run IDs. Never pick one: two dispatches would deploy twice, and `verify` must read the run that deployed the intended commit. |
| No deploy run appears within two minutes of the trigger | Check the dispatch's own output, and that `deploy.yml` is on the default branch (a workflow only there can be dispatched). On `existing-cicd-pr`, check the dispatcher run and the PR's deploy marker. |
| The `post-merge-deploy.yml` dispatcher run fails (`existing-cicd-pr`) | Stop and report it; nothing deployed. |
| `smaqit slots plan` reports `SLOTS-LOCAL-DEPLOY-REFUSED` | Stop. Plan the deploy on a CI path (`--path=ci-bootstrap` for the first deploy, then `ci-push` or `existing-cicd-pr`); never deploy from this machine. |
| The mesh join step fails, or the API's mesh name never resolves | Check the row's mesh secret name and the mesh action's inputs against the spec's Constraints rows. An empty key means this repository cannot read the platform's secret: ask the platform team to share it, then dispatch again. |
| The connect step fails asking for `id-token: write` | The `deploy` job lost its `permissions:` block; render `deploy.yml` again with `smaqit.infrastructure-cicd-generate`. |
| `namespace-guard.sh` reports Unauthorized | The platform did not accept the job's identity: check that the run used the machine-slug Environment and the spec's `Cluster Connect Action` pin is current. Never fall back to a stored kubeconfig. |
| `namespace-guard.sh` reports Forbidden (identity not bound) | The onboarding request has not merged or reconciled yet; check it with `smaqit.infrastructure-request-k3s-onboarding` and dispatch again once the platform reconciles. |
| `namespace-guard.sh` refuses (API unreachable after its retries, check undetermined, cluster-scoped verb permitted, exec into pods permitted, or Namespace mismatch) | Stop. For an unreachable API on a mesh-only target, check the mesh join step. For an exec grant, ask the platform team to remove `pods/exec` from the deployer's Role; set `Deployer Exec: allowed` only when the operator confirms the grant is deliberate. Do not proceed with a broader credential than the onboarding contract issues. |
| `exec-refused.sh` fails | Stop. A real exec went through (fix the Role as above), or the attempt gave no clear answer (NotFound, timeout): check the Deployment name and the API connection, then dispatch again. |
| A `kubectl auth can-i` check the project added answers `yes` for something the platform denies (typically exec into pods) | The check almost certainly used the `TYPE/NAME` form (`can-i create pods/exec`), which kubectl reads as "a pod named exec". Rewrite it with `--subresource=<name>`. Trust `namespace-guard.sh`'s answer, which already uses that form. Never widen scope or skip a check on the strength of the misleading `yes`. |
| A `--subresource` check the project added answers `no` for everything, including grants the platform documents | `can-i` does not validate subresource names, so a misspelled one (`--subresource=exe`) silently answers `no`. Check the spelling against the Kubernetes API reference before trusting the result. |
| A check step the project added ends at a command expected to exit non-zero (a `can-i` "no", a refused exec) before it reports a verdict | `bash -e` stopped the step on the answer it was waiting for. Bracket the command with `set +e` … `RC=$?` … `set -e` and decide from the captured output and exit code, or call `exec-refused.sh`. Never append `\|\| true`: it also hides a real error, such as an unreachable API, and turns it into a PASS. |
| `manifest-lint.sh` rejects a manifest | Report the exact offending field. Fix the manifest, not the check, then dispatch again. |
| `manifest-lint.sh` fails with `FileNotFoundError` on an Ingress or `probe-job.yaml` | The manifest was never rendered. Regenerate through `smaqit.infrastructure-cicd-generate`'s `k3s` mode and commit the result; never fabricate a placeholder host. |
| "Create runtime Secret" reports a runtime secret not set in the machine's Environment | Run the `environment-config` slot (`smaqit.infrastructure-repo-config` `runtime`) for that environment, completing any operator handoff it prints, then dispatch again. Never paste a value into the workflow. |
| "Create runtime Secret" is Forbidden | The platform's deployer Role does not grant `secrets` in the Namespace. Ask the platform team for `create`/`patch` on `secrets` there; never move the value into the Deployment as a literal. |
| `kubectl apply` fails (admission rejection) | Report the admission error verbatim. Do not retry with `--force` or widened scope. |
| `kubectl rollout status` times out | Report the timeout and the run's last known Deployment state. If the platform grants `pods/log`, include the failing container's logs; never `pods/exec`. |
| `Certificate` never reaches `Ready` (`own-tls`) | Check the `ClusterIssuer` name and the issuer's challenge reachability against the Environments row; do not bypass the wait. |
| Probe Job cannot be created (`jobs` forbidden) | Stop. Ask the platform team to grant `create`/`delete` on `jobs` in the Namespace. Do not fall back to checks the runner cannot reach. |
| Probe Job fails or times out | Report the Job status; include its logs only if `pods/log` is granted. Check the health path against the spec and whether a NetworkPolicy blocks same-Namespace traffic. |
| The run's verify step still fails after the 60-second retry window | Stop. Report the failing check. Do not mark the deployment complete. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |

## Allowed Tools

Bash(gh:*), Bash(smaqit:*), Bash(jq:*)
