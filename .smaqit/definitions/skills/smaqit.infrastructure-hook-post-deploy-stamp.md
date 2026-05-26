# Skill Definition: smaqit.infrastructure-hook-post-deploy-stamp

## Identity
- **Name:** smaqit.infrastructure-hook-post-deploy-stamp
- **Version:** 1.0.0
- **Description:** Write deploy stamp files (DEPLOY_SHA, DEPLOY_TIME) to the backend application directory on the VM after a successful deployment, and optionally add a deploy notification comment to the merged PR. Invoked by `smaqit.infrastructure-deploy-rsync` at the end of a deploy, or called directly to re-stamp without a full redeploy.

## Steps
1. **Resolve target:** Determine VM host (from secrets, Infrastructure spec, or explicit argument) and app directory (default: `/opt/him/backend/`).
2. **Determine SHA:** Run `git rev-parse HEAD` locally to get the current commit SHA.
3. **Write stamp files on VM:**
   ```
   ssh -i <key> ubuntu@<host> "printf '%s' '$(git rev-parse HEAD)' > /opt/him/backend/DEPLOY_SHA && printf '%s' '$(date -u +%Y-%m-%dT%H:%M:%SZ)' > /opt/him/backend/DEPLOY_TIME"
   ```
   CRITICAL: use `printf '%s'` not `echo` — `echo` appends a newline which `readFileSync` includes in the SHA string, causing `"sha": "d040a5f...\n"` in the health endpoint response.
   CRITICAL: stamp files must be in `/opt/him/backend/`, NOT `/opt/him/`. The Docker container mounts `/opt/him/backend/` as `/app`, so only files inside `backend/` are accessible to the Node.js process.
4. **Verify stamp (optional but recommended):** `curl -sf http://<host>/api/health | jq '{sha, deployedAt}'` — confirm `sha` matches first 8 chars of HEAD and `deployedAt` is within the last 60 seconds.
5. **PR notification (conditional):** If a PR number is available (from CI context `${{ github.event.pull_request.number }}` or explicit argument):
   a. Post a comment: "✅ Deployed — SHA `<sha>` at `<deployedAt>`. Health: `http://<host>/api/health`"
   b. Use `gh pr comment <pr-number> --body "..."` or the GitHub API.

## Output
- `/opt/him/backend/DEPLOY_SHA` written on VM (contains commit SHA, no trailing newline)
- `/opt/him/backend/DEPLOY_TIME` written on VM (contains UTC ISO 8601 datetime)
- Health endpoint returns populated `sha` and `deployedAt` fields
- PR comment posted (if PR number available)

## Scope
- Does NOT restart the container. Stamp files are read by the running process — a running container will serve the new stamps on the next health request without restart.
- Does NOT verify the full deployment (use `smaqit.infrastructure-deploy-verify` for that).
- Does NOT manage release tags. Use `smaqit.release-git-local` for that.

## Gotchas
- **Stamp file path = `/opt/him/backend/`** — writing to `/opt/him/DEPLOY_SHA` (without `backend/`) causes the health endpoint to return `"sha": "unknown"`. The Docker container bind-mount is `./backend:/app`, not `./:/app`.
- **No trailing newline** — `echo` appends `\n`; `readFileSync` in the Node process returns it as part of the string. Use `printf '%s'` to avoid `"sha": "d040a5f...\n"` in the response.
- **SHA is the full git SHA** — the health endpoint typically shows only the first 8 characters for display, but the full SHA is stored. Do not truncate before writing.
- **PR comment is best-effort** — if no PR number is available (e.g., direct push to main), skip the comment silently.

## Completion
- [ ] SHA determined from `git rev-parse HEAD`
- [ ] `DEPLOY_SHA` written to `/opt/him/backend/` (no trailing newline)
- [ ] `DEPLOY_TIME` written to `/opt/him/backend/` (no trailing newline)
- [ ] Health endpoint confirms `sha` and `deployedAt` populated
- [ ] PR comment posted (if PR number available) or skipped (if not)

## Failure Handling
| Situation | Action |
|-----------|--------|
| SSH connection fails | Report; do not silently continue — stampless deploys cause `"sha": "unknown"` |
| Health endpoint returns `"sha": "unknown"` after stamp | Check file path — likely written to `/opt/him/` instead of `/opt/him/backend/`; re-run with correct path |
| `jq` not available for verification | Use `grep sha` as fallback verification |
| PR comment fails (permission) | Report but do not fail the overall operation — stamping is the primary concern |

## Examples
**Input:** Deployment complete. CI workflow calls stamp step with VM host `81.24.10.203` and PR #52.
**Output:** DEPLOY_SHA and DEPLOY_TIME files on VM. `curl /api/health` returns `"sha": "d040a5f...", "deployedAt": "2026-04-27T14:32:00Z"`. PR #52 comment: "✅ Deployed — SHA `d040a5f` at 2026-04-27T14:32:00Z."
