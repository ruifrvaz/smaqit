---
name: smaqit.infrastructure-deploy-verify
description: Use after any deployment, or when re-verifying after a fix, to confirm the application is healthy before closing the task or proceeding. Checks the health endpoint, compares the deployed commit SHA with the local commit (or --expected-sha), and validates the SPA root; for an existing-k3s target the deployer can't reach, verifies inside the cluster with a probe Job. PASS/FAIL per check.
metadata:
  version: "1.3.0"
  smaqit-slots: "verify=verify"
---

# Deploy Verify

## Choose the mode

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

## Kubernetes Steps (existing-k3s)

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

## Scope

- Does NOT run functional or integration tests — use `smaqit.validation` for test suites.
- Does NOT restart or redeploy if checks fail — reports only.
- Does NOT verify TLS certificate validity — use `curl -vI` manually or a future `smaqit.infrastructure-domain-tls` skill. For `existing-k3s` `own-tls`, `smaqit.infrastructure-deploy-k3s-app` waits for the `Certificate` instead.
- Does NOT read pod logs to decide a result, and never uses `pods/exec` — in-cluster mode's verdict is the probe Job's status.

## Examples

**Input:** After deploy workflow completes, orchestrator invokes verify with URL `http://<vm-fixed-ip>`.

**Input (with explicit SHA):** After a CI/CD deploy where local HEAD may not match the deployed revision, the orchestrator passes `--expected-sha <deploy-run-headSha>`.

**Input (existing-k3s, in-cluster):** After `smaqit.infrastructure-deploy-k3s-app` rolls out to a `private-mesh` test machine, it invokes verify with `--environment test --expected-sha <sha>`.

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
- **`--expected-sha` overrides local HEAD.** When the orchestrator knows the exact revision a CI run deployed (e.g. from `gh run view --json headSha`), passing `--expected-sha` prevents false negatives when the local checkout is on a different branch or behind the deployed revision. When absent, `git rev-parse HEAD` is the unchanged default — existing callers are unaffected.

## Completion

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
