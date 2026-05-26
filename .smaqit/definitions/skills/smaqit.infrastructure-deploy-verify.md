# Skill Definition: smaqit.infrastructure-deploy-verify

## Identity
- **Name:** smaqit.infrastructure-deploy-verify
- **Version:** 1.0.0
- **Description:** Verify a deployment is healthy by checking the health endpoint, confirming the deployed commit SHA matches the expected commit, and validating the HTTP response from the SPA root. Use after any deployment to confirm success before closing the task or proceeding.

## Steps
1. Resolve the target URL. Look for it in the following order:
   a. Explicit `--url` argument provided by the caller
   b. `specs/infrastructure/deployment-topology.md` — `Fixed IP` or `Domain` value
   c. Ask the user if not found
2. Determine the expected commit SHA:
   a. Run `git rev-parse HEAD` locally to get the current commit SHA (first 8 chars for comparison)
3. **Health check:** `curl -sf <url>/api/health` — assert HTTP 200.
4. **SHA verification:** Parse the JSON response; check that `sha` starts with the first 8 characters of the expected SHA. If the field is missing, note "SHA field absent — health endpoint may not include deploy stamp."
5. **deployedAt check:** Confirm `deployedAt` is present and is a recent timestamp (within the last 30 minutes). Flag if stale.
6. **SPA root check:** `curl -sI <url>/` — assert HTTP 200 and `Content-Type: text/html`.
7. **API proxy check:** `curl -sf <url>/api/health` via the nginx path (same as step 3 but now explicitly confirming the reverse proxy is routing, not direct backend access).
8. Report: `PASS` or `FAIL` for each check. If all pass: "Deployment verified — SHA `<sha>`, deployed at `<deployedAt>`." If any fail: list failing checks with the actual vs expected values.

## Output
- Console report of all check results (PASS/FAIL per check + summary)

## Scope
- Does NOT run functional or integration tests. Use `smaqit.validation` for test suites.
- Does NOT restart or redeploy if checks fail. It only reports.
- Does NOT verify TLS certificate validity (use `curl -vI` manually or a future `smaqit.infrastructure-domain-tls` verify step).

## Gotchas
- The health route is `/api/health`, NOT `/health`. This has caused confusion in prior sessions. Direct backend access (bypassing nginx) is at `<host>:3001/api/health`.
- `sha` field in health response is populated from a stamp file written by the deploy workflow (`/opt/him/backend/DEPLOY_SHA`). If the stamp file path is wrong (e.g. written to host `/opt/him/` but container only mounts `/opt/him/backend/`), the field returns `"unknown"`. This is a deployment bug, not a health endpoint bug.
- `deployedAt` is the timestamp when the deploy stamp was written by CI, not when the container started. On first deploy the field may be absent if `DEPLOY_SHA`/`DEPLOY_TIME` stamp steps were not included in the workflow.
- If SHA comparison fails after a re-deploy without a new commit (e.g., re-running the same workflow on the same commit), the SHA will match — this is correct and expected.

## Completion
- [ ] Target URL resolved
- [ ] Expected SHA determined
- [ ] Health check: HTTP 200 received
- [ ] SHA match confirmed (or absence noted)
- [ ] `deployedAt` present and recent
- [ ] SPA root: HTTP 200 with text/html
- [ ] Summary report delivered

## Failure Handling
| Situation | Action |
|-----------|--------|
| Target URL not resolvable | Ask the user for the URL |
| Health check returns non-200 | Report the HTTP status code + response body snippet; suggest checking Docker container status |
| SHA mismatch | Report expected vs actual SHA; suggest re-running the deploy workflow |
| Connection refused | Report; suggest checking nginx status (`sudo systemctl status nginx`) and Docker container (`docker ps`) on the VM |

## Examples
**Input:** After deploy workflow completes, orchestrator invokes verify with URL `http://81.24.10.203`.
**Output:** PASS: health 200, SHA `d040a5f` matches, deployedAt 2026-04-27T14:32:00Z (3 min ago), SPA root 200 text/html. "Deployment verified."
