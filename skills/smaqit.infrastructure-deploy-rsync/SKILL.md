---
name: smaqit.infrastructure-deploy-rsync
description: Use when deploying a Node.js backend + React frontend application to a remote VM via rsync. Used in the Phase 5 dev environment sweep of `smaqit.new-greenfield-project` to validate the deployment approach locally before CI/CD. Also use as a manual fallback for direct VM deployment outside the CI/CD pipeline.
metadata:
  version: "1.0.0"
---

# Deploy Application via rsync

## Pre-conditions

- VM bootstrapped (`smaqit.infrastructure-vm-bootstrap` complete)
- Local Vault running and unsealed (`smaqit.infrastructure-vault-loader` complete); SSH private key at `secret/<project-slug>/ssh`
- Docker running on VM and `ubuntu` in docker group
- `deployment/docker-compose.yml` and `deployment/nginx/him.conf` present locally

## Steps

1. **Fetch SSH key from Vault into a secure temp file:**
   ```bash
   export VAULT_ADDR=http://127.0.0.1:8200
   export PROJECT_SLUG=<project-slug>   # from copilot-instructions.md

   TMPKEY=$(mktemp)
   trap "rm -f $TMPKEY" EXIT
   vault kv get -field=private_key secret/${PROJECT_SLUG}/ssh > "$TMPKEY"
   chmod 600 "$TMPKEY"
   ```
   All subsequent `ssh`, `rsync`, and `scp` commands use `-i "$TMPKEY"`. The file is wiped
   automatically when the shell exits or the script completes.

2. **Build backend:**
   ```bash
   cd backend && npm run build
   ```
   Produces `backend/dist/`.

2. **Build frontend:**
   ```bash
   cd frontend && npm run build
   ```
   Produces `frontend/dist/`. If `VITE_DEMO_MODE` must be set, export it before building — Vite bakes
   this value in at build time and it cannot be changed without a rebuild:
   ```bash
   VITE_DEMO_MODE=true npm run build
   ```

3. **Transfer backend artifacts to VM:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "mkdir -p /opt/him/backend/dist"
   rsync -avz --delete backend/dist/ ubuntu@<host>:/opt/him/backend/dist/
   rsync -avz backend/package.json backend/package-lock.json ubuntu@<host>:/opt/him/backend/
   ```
   CRITICAL: Always `mkdir -p /opt/him/backend/dist` before rsyncing. The trailing slash on
   `backend/dist/` copies the directory's *contents* — if the target directory does not exist, rsync
   creates it one level too shallow and the container fails with
   `Cannot find module '/app/dist/index.js'`.

4. **Install `node_modules` on VM** (production only, inside a throwaway container):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "cd /opt/him/backend && docker run --rm -v \$(pwd):/app -w /app node:22-alpine npm install --production"
   ```

5. **Transfer frontend build:**
   ```bash
   rsync -avz --delete frontend/dist/ ubuntu@<host>:/opt/him/frontend/
   ```

6. **Transfer config files:**
   ```bash
   scp -i "$TMPKEY" deployment/docker-compose.yml ubuntu@<host>:/opt/him/docker-compose.yml
   scp -i "$TMPKEY" deployment/nginx/him.conf ubuntu@<host>:/etc/nginx/sites-available/him
   ```

7. **Write deploy stamp files** (enables SHA verification in the health endpoint):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "printf '%s' '$(git rev-parse HEAD)' > /opt/him/backend/DEPLOY_SHA && \
      printf '%s' '$(date -u +%Y-%m-%dT%H:%M:%SZ)' > /opt/him/backend/DEPLOY_TIME"
   ```
   Write to `/opt/him/backend/`, not `/opt/him/` — the container mounts `/opt/him/backend/` as `/app`;
   files one level up are invisible to the container.

8. **Restart containers:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "cd /opt/him && docker compose up -d --force-recreate"
   ```
   Use `docker compose` (v2, no hyphen). `--force-recreate` is required because the app is deployed
   as files, not as a new image.

9. **Reload nginx:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "sudo nginx -t && sudo systemctl reload nginx"
   ```

10. **Verify:** Invoke `smaqit.infrastructure-deploy-verify` with the VM URL.

## Output

Application artifacts deployed to `/opt/him/` on the VM, container running, nginx serving.

## Scope

- Does NOT provision the VM — use `smaqit.infrastructure-provision-cyso` for that.
- Does NOT handle database migrations — current project uses SQLite with append-only schema.
- Does NOT pull from a container registry — deploys local build artifacts directly.

## Examples

**Input:** Feature branch merged to main; CI/CD workflow runs or operator invokes `/app.deploy`.  
**Output:** Backend and frontend artifacts on VM, container running, nginx serving, health endpoint
returning correct SHA and `deployedAt` timestamp.

## Gotchas

- **Hardcoded source paths** — this skill assumes `backend/` and `frontend/` as local source directories. If the project uses different paths, update steps 1–5 accordingly and ensure the stack spec declares the same paths.
- **`Cannot find module '/app/dist/index.js'`** — rsync trailing slash puts files at the wrong depth.
  Always `mkdir -p /opt/him/backend/dist` before rsyncing `backend/dist/`.
- **`VITE_DEMO_MODE` is build-time only** — changing the GitHub Actions variable and re-running the
  workflow triggers a rebuild. Changing it in the running environment has no effect.
- **Deploy stamp path** — write to `/opt/him/backend/DEPLOY_SHA` and `/opt/him/backend/DEPLOY_TIME`,
  not `/opt/him/`. Files in `/opt/him/` are invisible to the container.
- **`docker compose` vs `docker-compose`** — use `docker compose` (v2, no hyphen) on Ubuntu 24.04
  with Docker 24+.
- **`--force-recreate`** — required even when no image changed; ensures the container restarts with
  updated files.

## Completion

- [ ] Backend TypeScript build succeeded
- [ ] Frontend Vite build succeeded
- [ ] Backend artifacts rsynced to `/opt/him/backend/dist/`
- [ ] `node_modules` installed via Docker on VM
- [ ] Frontend dist rsynced to `/opt/him/frontend/`
- [ ] `docker-compose.yml` and nginx config transferred
- [ ] Deploy stamp files written to `/opt/him/backend/`
- [ ] Container restarted with `--force-recreate`
- [ ] nginx reloaded without errors
- [ ] `smaqit.infrastructure-deploy-verify` invoked and passed

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `npm run build` fails | Show the build error; stop — do not deploy broken artifacts |
| rsync permission denied | Check SSH key and `ubuntu` user write permissions on `/opt/him/` |
| `Cannot find module` in container logs | Check rsync target depth; re-run step 3 with explicit `mkdir -p` |
| nginx config test fails | Show the nginx error; do not reload — current config remains active |
| deploy-verify reports SHA mismatch | Check `docker ps`; old container may still be running; retry step 8 |
