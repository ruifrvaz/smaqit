---
name: smaqit.infrastructure-deploy-verify
description: Use after any deployment, or when re-verifying after a fix, to confirm the application is healthy before closing the task. Checks the health endpoint, the deployed commit SHA (local HEAD or --expected-sha) and the SPA root; for an existing-k3s target the deployer can't reach, a probe Job in the cluster; after a CI deploy of existing-k3s, the deploy run's own verify steps. PASS/FAIL per check.
metadata:
  version: "1.4.0"
  smaqit-slots: "verify=verify, verify=ci-evidence"
---

# Deploy Verify

## Choose the operation

The deployment slot plan names the operation; run the one it names.

- **`verify`** — checks the target from this machine, in HTTP or in-cluster mode (below). The
  default for every VM mode, and for `existing-k3s` on the `standard` path (a local deploy).
- **`ci-evidence`** — decides from the deploy run's own recorded steps (CI Evidence Steps below). The
  default for `existing-k3s` on the `existing-cicd-pr` and `ci-push` paths: the generated `deploy.yml`
  already verified the target from inside its mesh and cluster access, so this machine needs no mesh,
  kubeconfig, or Vault access.

## Choose the mode (operation `verify`)

- **HTTP mode** (Steps 1–8) — every VM target, and an `existing-k3s` environment whose
  Environments row is `public` with an Ingress the app owns (`own-tls` or
  `platform-default-tls`).
- **In-cluster mode** (Kubernetes Steps K1–K5) — an `existing-k3s` environment whose row is
  `private-mesh`, or whose Ingress Mode is `platform-owned` (for example behind the platform's SSO
  gate). The deployer cannot reach such an app from outside the cluster, so HTTP checks from this
  machine would test the wrong thing.

For `existing-k3s`, the caller passes `--environment <environment>`; read its row with
`smaqit spec environments --env=<environment> --json` and choose from `reachability` and
`ingressMode`. Never fall back from in-cluster mode to HTTP checks the deployer cannot reach.

## CI Evidence Steps (operation `ci-evidence`)

Inputs from the Deployment agent: the deploy run's ID (`--run-id`), the commit the deploy was meant to
ship (`--expected-sha`: the PR merge SHA on `existing-cicd-pr`, the pushed SHA on `ci-push`, never the
run's own `headSha`), and the environment the run deployed (`--environment`).

C1. **Run the check** with the environment's row:
    ```bash
    smaqit spec environments --env=<environment> --json \
      | bash scripts/ci-evidence.sh - <environment> --run-id <run-id> --expected-sha <sha>
    ```
    It reads `gh run view <run-id> --json headSha,status,conclusion,jobs` and checks, in order:
    - the run finished (`status` is `completed`) and concluded `success`;
    - the run's `headSha` is the expected commit, or contains it: when they differ, one
      `gh api repos/{owner}/{repo}/compare/<expected>...<headSha>` call must report `identical` or
      `ahead` (the dispatcher deploys the branch tip at dispatch time, which may include later merges);
    - the `deploy` job's "Verify revision" step succeeded (the Deployment runs the run's commit image);
    - the row's verify step succeeded: "Verify in cluster" for a `private-mesh` or `platform-owned`
      row, "Verify over HTTPS" otherwise (the same rule the workflow's `VERIFY` uses).

C2. **Report** the script's single PASS or FAIL line. Exit 0 is PASS, 1 is FAIL, 2 is a usage error.

## Kubernetes Steps (existing-k3s, operation `verify`)

Run with the target's scoped kubeconfig already exported (see
`smaqit.infrastructure-deploy-k3s-app`), against Namespace `app-<app-slug>`.

K1. **Revision check.** The rolled-out Deployment must run this commit's image; its tag is the full
    commit SHA:
    ```bash
    kubectl get -n "app-${APP_SLUG}" "deployment/${APP_SLUG}" -o jsonpath='{.spec.template.spec.containers[0].image}'
    ```
    Assert the tag equals the expected SHA (`--expected-sha`, or `git rev-parse HEAD`).

K2. **HTTP mode for a public, app-owned Ingress:** run Steps 3 and 6 against
    `https://<host><healthPath>` and `https://<host>/` from the row, retrying every 5 seconds for up
    to 60 seconds. Skip Steps 4, 5, and 7 — the revision is already proven by K1, and there is no
    nginx or stamp file.

K3. **In-cluster mode:** stamp the committed probe Job with this run's image and the row's health
    path into a temporary copy, lint it, and run it:
    ```bash
    PROBE="$(mktemp -d)/probe-job.yaml"
    sed -e "s|__IMAGE__|<image>|g" -e "s|__HEALTH_PATH__|<healthPath>|g" deployment/k3s/probe-job.yaml > "$PROBE"
    bash deployment/k3s/scripts/manifest-lint.sh "$PROBE"
    kubectl delete -n "app-${APP_SLUG}" "job/${APP_SLUG}-probe" --ignore-not-found --wait=true
    kubectl apply -n "app-${APP_SLUG}" -f "$PROBE"
    ```
    The Job runs the app's own image and requests `http://<app-slug>:80<healthPath>` over cluster
    DNS, retrying for up to 60 seconds itself.

K4. **Read the verdict from the Job's status**, polling every 5 seconds for up to 120 seconds:
    `kubectl get -n "app-${APP_SLUG}" "job/${APP_SLUG}-probe" -o jsonpath='{.status.succeeded}/{.status.failed}'`
    — `1/` is PASS, `/1` is FAIL. This needs no `pods/log` grant. Only on FAIL or timeout, try
    `kubectl logs job/<app-slug>-probe` for diagnosis, and report "logs not readable" if the
    platform denies it.

K5. **Clean up and report:** delete the Job (`--ignore-not-found`), then print PASS/FAIL for the
    revision and the health check, and the one-line summary.

## Steps (HTTP mode)

1. **Resolve the target URL.** Check in this order:
   - Explicit `--url` argument provided by the caller
   - The infrastructure spec (e.g. `specs/infrastructure/deployment.md`) — `Fixed IP` or `Domain` value
   - Ask the user if not found in either location

2. **Determine the expected commit SHA.**
   - If `--expected-sha <sha>` is provided by the caller: use it directly. Accept full (40-char) or short (8-char) SHA; normalize to first 8 characters for comparison.
   - Otherwise (backward-compatible default): run `git rev-parse HEAD` and use the first 8 characters.

3. **Health check.** Run:
   ```
   curl -sf <url>/<health-check-path>
   ```
   Assert HTTP 200. Capture the full JSON response body for steps 4 and 5.

4. **SHA verification.** From the JSON response, read the `sha` field.
   - If present: assert it starts with the first 8 characters of the expected SHA.
   - If absent: note "SHA field absent — health endpoint may not include deploy stamp."
   - If `"unknown"`: note "Stamp file missing on container filesystem — see Gotchas."

5. **`deployedAt` check.** Confirm `deployedAt` is present in the JSON response and within the last 30 minutes. Flag if stale or absent.

6. **SPA root check.** Run:
   ```
   curl -sI <url>/
   ```
   Assert HTTP 200 and `Content-Type: text/html`.

7. **API proxy confirmation.** The `curl` in step 3 already exercises the nginx reverse proxy path (`/<health-check-path>` via `<url>`). Explicitly note this distinguishes nginx-routed access from direct backend access at `<host>:<backend-port>/<health-check-path>`. No additional request is needed unless nginx routing is in doubt.

8. **Report.** Print PASS or FAIL for each check. If all pass:
   ```
   Deployment verified — SHA <sha>, deployed at <deployedAt>.
   ```
   If any fail: list each failing check with actual vs expected values.

## Output

Console report: PASS/FAIL for each check (health, SHA, deployedAt, SPA root) plus a one-line summary.
`ci-evidence` prints one `ci-evidence: PASS — …` or `ci-evidence: FAIL — …` line naming the run, its
head commit, and the step that decided.

## Scope

- Does NOT run functional or integration tests — use `smaqit.validation` for test suites.
- Does NOT restart or redeploy if checks fail — reports only.
- Does NOT verify TLS certificate validity — use `curl -vI` manually or a future `smaqit.infrastructure-domain-tls` skill. For `existing-k3s` `own-tls`, `smaqit.infrastructure-deploy-k3s-app` waits for the `Certificate` instead.
- Does NOT read pod logs to decide a result, and never uses `pods/exec` — in-cluster mode's verdict is the probe Job's status.
- `ci-evidence` reads only GitHub's run record and one commit comparison — no kubeconfig, Vault,
  `kubectl`, mesh access, or run logs.

## Examples

**Input:** After deploy workflow completes, orchestrator invokes verify with URL `http://<vm-fixed-ip>`.

**Input (with explicit SHA):** After a CI/CD deploy where local HEAD may not match the deployed revision, the orchestrator passes `--expected-sha <deploy-run-headSha>`.

**Input (existing-k3s, in-cluster):** After `smaqit.infrastructure-deploy-k3s-app` rolls out to a `private-mesh` test machine, it invokes verify with `--environment test --expected-sha <sha>`.

**Input (existing-k3s, CI path):** After the Deployment agent watched `deploy.yml` run `9876543210`
deploy the merge of PR #42 to `prod`, it runs `ci-evidence` with `--run-id 9876543210
--expected-sha <merge-sha>` and `--environment prod`.

**Output (ci-evidence):**
```
ci-evidence: PASS — run 9876543210 deployed 4f1c… (contains 4f1c…); 'Verify revision' and 'Verify in cluster' succeeded
```

**Output (in-cluster):**
```
PASS: Deployment runs ghcr.io/<org>/<repo>:<sha>
PASS: probe Job succeeded — http://<app-slug>:80/health.json answered in cluster
Deployment verified in cluster — SHA <sha>.
```

**Output (HTTP mode):**
```
PASS: health 200
PASS: SHA d040a5f matches
PASS: deployedAt 2026-04-27T14:32:00Z (3 min ago)
PASS: SPA root 200 text/html
Deployment verified — SHA d040a5f, deployed at 2026-04-27T14:32:00Z.
```

## Gotchas

- The health route path is project-specific — confirm the exact path (e.g. `/health`, `/api/health`)
  from the infrastructure spec rather than assuming. Direct backend access (bypassing nginx) is at
  `<host>:<backend-port>/<health-check-path>`.
- The `sha` field is populated from `__APP_DIR__/backend/DEPLOY_SHA`, written by the CI deploy workflow. If the workflow writes the stamp to the host path but the container only mounts a subdirectory, the field returns `"unknown"` — this is a deployment bug, not a health endpoint bug.
- `deployedAt` is the CI timestamp when the stamp was written, not the container start time. On first deploy the field may be absent if stamp steps were omitted from the workflow.
- Re-deploying without a new commit (re-running the same workflow on the same SHA) will produce a SHA match — this is correct and expected.
- **`ci-evidence` takes the intended commit, not the run's `headSha`.** Passing the run's own
  `headSha` would compare the run with itself. On `existing-cicd-pr` the dispatcher deploys `main`'s
  tip at dispatch time, so a later merge may already be included; the compare call accepts that and
  rejects a run that doesn't contain the merge.
- **`--expected-sha` overrides local HEAD.** When the orchestrator knows the exact revision a CI run deployed (e.g. from `gh run view --json headSha`), passing `--expected-sha` prevents false negatives when the local checkout is on a different branch or behind the deployed revision. When absent, `git rev-parse HEAD` is the unchanged default — existing callers are unaffected.

## Completion

- [ ] Operation taken from the slot plan (`verify` or `ci-evidence`)
- [ ] `ci-evidence`: `ci-evidence.sh` printed PASS for the deploy run and the intended commit
- [ ] Mode chosen (HTTP, or in-cluster for an `existing-k3s` `private-mesh` / `platform-owned` row)
- [ ] In-cluster mode: Deployment image tag matches the expected SHA, probe Job succeeded, Job deleted
- [ ] Target URL resolved (HTTP mode)
- [ ] Expected SHA determined (via `--expected-sha` or `git rev-parse HEAD` fallback)
- [ ] Health check: HTTP 200 received
- [ ] SHA match confirmed (or absence/unknown state noted with reason)
- [ ] `deployedAt` present and recent, or stale/absent flagged
- [ ] SPA root: HTTP 200 with `Content-Type: text/html`
- [ ] Summary report delivered

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| Target URL not resolvable | Ask the user for the URL directly |
| Health check returns non-200 | Report the HTTP status code and response body snippet; suggest checking Docker container status (`docker ps`) |
| SHA mismatch | Report expected vs actual SHA; suggest re-running the deploy workflow |
| Connection refused | Report; suggest checking nginx (`sudo systemctl status nginx`) and Docker containers (`docker ps`) on the VM |
| `existing-k3s` row can't be read (`SLOTS-SPEC-*`) | Stop and surface the lines; the fix belongs in the Infrastructure spec |
| Probe Job forbidden (`jobs` not granted) | Stop. Ask the platform team for `create`/`delete` on `jobs` in the Namespace; never switch to HTTP checks the deployer cannot reach |
| Probe Job fails or times out | Report FAIL with the Job status; add logs only if `pods/log` is granted. Check the health path and whether a NetworkPolicy blocks same-Namespace traffic |
| Deployment image tag differs from the expected SHA | Report both; the rollout did not apply this commit's image |
| `ci-evidence`: run has not finished | Wait for the run to complete, then run the check again; never pass an unfinished run |
| `ci-evidence`: run concluded other than `success` | Stop and report the run; the deploy itself failed |
| `ci-evidence`: wrong SHA (`behind` or `diverged`) | Stop. The run did not deploy the intended commit; find the run that did, or redeploy |
| `ci-evidence`: "Verify revision" or the applicable verify step failed or is missing | Stop and report the step. A missing step means `deploy.yml` predates the current template; regenerate it with `smaqit.infrastructure-cicd-generate` |
| `ci-evidence`: applicable verify step was skipped | Stop. The workflow and the Environments row disagree on how to verify (or the run deployed another machine); regenerate `deploy.yml` or check the dispatched environment |
| `ci-evidence`: `gh run view` or the compare call fails | Report FAIL with the reason; check the run ID, `gh`'s repository (`GH_REPO`) and token. A failed read is never a pass |
