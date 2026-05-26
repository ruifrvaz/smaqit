---
name: smaqit.infrastructure-hook-post-deploy-stamp
description: Use when writing deploy stamp files (DEPLOY_SHA, DEPLOY_TIME) to the VM after a successful deployment, or when re-stamping without a full redeploy. Also use when the health endpoint returns "sha":"unknown" after deployment, when a deployed PR needs a notification comment, or when smaqit.infrastructure-deploy-rsync invokes the post-deploy stamp step. Produces DEPLOY_SHA and DEPLOY_TIME files in /opt/him/backend/ on the VM and optionally posts a deploy comment on the merged PR.
metadata:
  version: "1.0.0"
---

# Post-Deploy Stamp

## Steps

1. **Resolve target:** Determine the VM host from the `VM_HOST` secret, the Infrastructure spec, or an explicit argument. Confirm the app directory is `/opt/him/backend/`.

2. **Determine SHA:** Run `git rev-parse HEAD` locally to obtain the current commit SHA.

3. **Write stamp files on VM:**

   ```bash
   ssh -i <key> ubuntu@<host> "printf '%s' '$(git rev-parse HEAD)' > /opt/him/backend/DEPLOY_SHA && printf '%s' '$(date -u +%Y-%m-%dT%H:%M:%SZ)' > /opt/him/backend/DEPLOY_TIME"
   ```

   **CRITICAL — use `printf '%s'`, not `echo`:** `echo` appends a newline; `readFileSync` in the Node process includes it in the returned string, producing `"sha": "d040a5f...\n"` in the health response.

   **CRITICAL — path must be `/opt/him/backend/`:** The Docker bind-mount is `./backend:/app`. Files written to `/opt/him/` (without `backend/`) are not visible to the Node process and cause `"sha": "unknown"`.

   Write the full git SHA — do not truncate before writing.

4. **Verify stamp (recommended):**

   ```bash
   curl -sf http://<host>/api/health | jq '{sha, deployedAt}'
   ```

   Confirm `sha` matches the first 8 characters of HEAD and `deployedAt` is within the last 60 seconds. If `jq` is unavailable, use `grep sha` as a fallback.

5. **PR notification (conditional):** If a PR number is available (from CI context `${{ github.event.pull_request.number }}` or an explicit argument):

   ```bash
   gh pr comment <pr-number> --body "✅ Deployed — SHA \`<sha>\` at \`<deployedAt>\`. Health: \`http://<host>/api/health\`"
   ```

   If no PR number is available (e.g., direct push to main), skip silently.

## Output

- `/opt/him/backend/DEPLOY_SHA` written on VM — full commit SHA, no trailing newline
- `/opt/him/backend/DEPLOY_TIME` written on VM — UTC ISO 8601 datetime, no trailing newline
- Health endpoint at `/api/health` returns populated `sha` and `deployedAt` fields
- PR comment posted if a PR number is available

## Scope

- Does NOT restart the container. Stamp files are read by the running process — the health endpoint reflects new values on the next request without restart.
- Does NOT verify the full deployment. Use `smaqit.infrastructure-deploy-verify` for that.
- Does NOT manage release tags. Use `smaqit.release-git-local` for that.

## Examples

**Input:** Deployment complete. CI workflow calls stamp step with VM host `81.24.10.203` and PR #52.

**Output:** `DEPLOY_SHA` and `DEPLOY_TIME` written to `/opt/him/backend/` on the VM. `curl /api/health` returns `"sha": "d040a5f...", "deployedAt": "2026-04-27T14:32:00Z"`. PR #52 receives comment: "✅ Deployed — SHA `d040a5f` at 2026-04-27T14:32:00Z."

## Gotchas

- **Stamp file path is `/opt/him/backend/`** — writing to `/opt/him/DEPLOY_SHA` (without `backend/`) causes `"sha": "unknown"` in the health response. The Docker bind-mount is `./backend:/app`, not `./:/app`.
- **No trailing newline** — use `printf '%s'`, not `echo`. `echo` appends `\n`; `readFileSync` returns it as part of the string.
- **Store the full SHA** — the health endpoint displays only the first 8 characters for UI purposes, but the stamp file must contain the full SHA.
- **PR comment is best-effort** — if no PR number is available, skip silently; do not fail the overall operation.

## Completion

- [ ] SHA obtained from `git rev-parse HEAD`
- [ ] `DEPLOY_SHA` written to `/opt/him/backend/` (no trailing newline)
- [ ] `DEPLOY_TIME` written to `/opt/him/backend/` (no trailing newline)
- [ ] Health endpoint confirms `sha` and `deployedAt` are populated
- [ ] PR comment posted (if PR number available) or skipped (if not)

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| SSH connection fails | Report immediately — do not continue; a stampless deploy causes `"sha": "unknown"` in the health response |
| Health endpoint returns `"sha": "unknown"` after stamp | Check file path — stamp was likely written to `/opt/him/` instead of `/opt/him/backend/`; re-run with the correct path |
| PR comment fails (permission error) | Report but do not fail the overall operation — stamping is the primary concern |
