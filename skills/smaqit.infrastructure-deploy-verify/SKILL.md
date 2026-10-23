---
name: smaqit.infrastructure-deploy-verify
description: Use after any deployment, or when re-verifying after a fix, to confirm the application is healthy before closing the task. For a VM target, checks the health endpoint, the deployed commit SHA (local HEAD or --expected-sha) and the SPA root; for existing-k3s, which deploys only from CI, decides from the deploy run's own verify steps. PASS/FAIL per check.
metadata:
  version: "2.0.0"
  smaqit-slots: "verify=verify, verify=ci-evidence"
---

# Deploy Verify

## Choose the operation

The deployment slot plan names the operation; run the one it names.

- **`verify`** — HTTP checks against the target from this machine (Steps 1–8). The default for every
  VM mode.
- **`ci-evidence`** — decides from the deploy run's own recorded steps (CI Evidence Steps below). The
  only operation for `existing-k3s`, on every path it plans (`ci-bootstrap`, `ci-push`,
  `existing-cicd-pr`): existing-k3s deploys only from CI, and the generated `deploy.yml` already
  verified the target from inside its mesh and cluster access, so this machine needs no mesh,
  kubeconfig, or Vault access.

## CI Evidence Steps (operation `ci-evidence`)

Inputs from the Deployment agent: the deploy run's ID (`--run-id`), the commit the deploy was meant to
ship (`--expected-sha`: the PR merge SHA on `existing-cicd-pr` and `ci-bootstrap`, the pushed SHA on
`ci-push`, never the run's own `headSha`), and the environment the run deployed (`--environment`).

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

## Steps (operation `verify`)

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
- Does NOT verify TLS certificate validity — use `curl -vI` manually or a future `smaqit.infrastructure-domain-tls` skill. For `existing-k3s` `own-tls`, the generated `deploy.yml` waits for the `Certificate` instead.
- Does NOT reach an `existing-k3s` target from this machine: no `kubectl`, probe Job, or HTTP check
  against it. Its deploy run verified the target in-cluster or over HTTPS; `ci-evidence` reads only
  GitHub's run record and one commit comparison — no kubeconfig, Vault, mesh access, or run logs.

## Examples

**Input:** After deploy workflow completes, orchestrator invokes verify with URL `http://<vm-fixed-ip>`.

**Input (with explicit SHA):** After a CI/CD deploy where local HEAD may not match the deployed revision, the orchestrator passes `--expected-sha <deploy-run-headSha>`.

**Input (existing-k3s):** After the Deployment agent watched `deploy.yml` run `9876543210`
deploy the merge of PR #42 to `prod`, it runs `ci-evidence` with `--run-id 9876543210
--expected-sha <merge-sha>` and `--environment prod`.

**Output (ci-evidence):**
```
ci-evidence: PASS — run 9876543210 deployed 4f1c… (contains 4f1c…); 'Verify revision' and 'Verify in cluster' succeeded
```

**Output (verify):**
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
- [ ] `verify`: target URL resolved
- [ ] `verify`: expected SHA determined (via `--expected-sha` or `git rev-parse HEAD` fallback)
- [ ] `verify`: health check returned HTTP 200
- [ ] `verify`: SHA match confirmed (or absence/unknown state noted with reason)
- [ ] `verify`: `deployedAt` present and recent, or stale/absent flagged
- [ ] `verify`: SPA root returned HTTP 200 with `Content-Type: text/html`
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
| The plan names `verify` for an `existing-k3s` target | Stop. The plan is stale or hand-edited; existing-k3s verifies only with `ci-evidence`. Re-resolve it with `smaqit slots plan` on a CI path |
| `existing-k3s` row can't be read (`SLOTS-SPEC-*`) | Stop and surface the lines; the fix belongs in the Infrastructure spec |
| `ci-evidence`: run has not finished | Wait for the run to complete, then run the check again; never pass an unfinished run |
| `ci-evidence`: run concluded other than `success` | Stop and report the run; the deploy itself failed |
| `ci-evidence`: wrong SHA (`behind` or `diverged`) | Stop. The run did not deploy the intended commit; find the run that did, or redeploy |
| `ci-evidence`: "Verify revision" or the applicable verify step failed or is missing | Stop and report the step. A missing step means `deploy.yml` predates the current template; regenerate it with `smaqit.infrastructure-cicd-generate` |
| `ci-evidence`: applicable verify step was skipped | Stop. The workflow and the Environments row disagree on how to verify (or the run deployed another machine); regenerate `deploy.yml` or check the dispatched environment |
| `ci-evidence`: `gh run view` or the compare call fails | Report FAIL with the reason; check the run ID, `gh`'s repository (`GH_REPO`) and token. A failed read is never a pass |
