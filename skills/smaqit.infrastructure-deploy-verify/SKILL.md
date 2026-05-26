---
name: smaqit.infrastructure-deploy-verify
description: Use after any deployment to confirm the application is healthy before closing the task or proceeding. Checks the health endpoint, verifies the deployed commit SHA against the current local commit, and validates the SPA root HTTP response. Produces a PASS/FAIL report per check with a final summary. Also use when a deployment task asks to confirm success or when re-verifying after a fix.
metadata:
  version: "1.0.0"
---

# Deploy Verify

## Steps

1. **Resolve the target URL.** Check in this order:
   - Explicit `--url` argument provided by the caller
   - `specs/infrastructure/deployment-topology.md` — `Fixed IP` or `Domain` value
   - Ask the user if not found in either location

2. **Determine the expected commit SHA.** Run:
   ```
   git rev-parse HEAD
   ```
   Use the first 8 characters for comparison.

3. **Health check.** Run:
   ```
   curl -sf <url>/api/health
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

7. **API proxy confirmation.** The `curl` in step 3 already exercises the nginx reverse proxy path (`/api/health` via `<url>`). Explicitly note this distinguishes nginx-routed access from direct backend access at `<host>:3001/api/health`. No additional request is needed unless nginx routing is in doubt.

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
- Does NOT verify TLS certificate validity — use `curl -vI` manually or a future `smaqit.infrastructure-domain-tls` skill.

## Examples

**Input:** After deploy workflow completes, orchestrator invokes verify with URL `http://81.24.10.203`.

**Output:**
```
PASS: health 200
PASS: SHA d040a5f matches
PASS: deployedAt 2026-04-27T14:32:00Z (3 min ago)
PASS: SPA root 200 text/html
Deployment verified — SHA d040a5f, deployed at 2026-04-27T14:32:00Z.
```

## Gotchas

- The health route is `/api/health`, **not** `/health`. Direct backend access (bypassing nginx) is at `<host>:3001/api/health`.
- The `sha` field is populated from `/opt/him/backend/DEPLOY_SHA`, written by the CI deploy workflow. If the workflow writes the stamp to the host path but the container only mounts a subdirectory, the field returns `"unknown"` — this is a deployment bug, not a health endpoint bug.
- `deployedAt` is the CI timestamp when the stamp was written, not the container start time. On first deploy the field may be absent if stamp steps were omitted from the workflow.
- Re-deploying without a new commit (re-running the same workflow on the same SHA) will produce a SHA match — this is correct and expected.

## Completion

- [ ] Target URL resolved
- [ ] Expected SHA determined via `git rev-parse HEAD`
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
