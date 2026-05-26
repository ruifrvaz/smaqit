# Skill Definition: smaqit.infrastructure-deploy-rsync

## Identity
- **Name:** smaqit.infrastructure-deploy-rsync
- **Version:** 1.0.0
- **Description:** Deploy a Node.js + React application to a VM using rsync and Docker Compose. Covers: backend TypeScript build, frontend Vite build, rsync artifact transfer, `node_modules` install via Docker, and container restart. Use when deploying the HIM Corporate application (or similarly structured apps) directly from a local or CI environment to a remote VM.

## Pre-conditions
- VM bootstrapped (`smaqit.infrastructure-vm-bootstrap` complete)
- SSH key available (passphrase-free for CI; with passphrase for manual)
- Docker running on VM and `ubuntu` in docker group
- `docker-compose.yml` present in `deployment/` directory
- nginx config present in `deployment/nginx/`

## Steps
1. **Build backend:**
   ```
   cd backend && npm run build
   ```
   Produces `backend/dist/`.
2. **Build frontend:**
   ```
   cd frontend && npm run build
   ```
   Produces `frontend/dist/`. NOTE: if `VITE_DEMO_MODE` must be set, export it before running build: `VITE_DEMO_MODE=true npm run build`. Vite bakes this value in at build time — it cannot be changed without a rebuild.
3. **Transfer backend artifacts to VM:**
   ```
   ssh -i <key> ubuntu@<host> "mkdir -p /opt/him/backend/dist"
   rsync -avz --delete backend/dist/ ubuntu@<host>:/opt/him/backend/dist/
   rsync -avz backend/package.json backend/package-lock.json ubuntu@<host>:/opt/him/backend/
   ```
   CRITICAL: `rsync backend/dist/` (trailing slash) copies the CONTENTS of `dist/`, not `dist/` itself. Always create the target directory explicitly (`mkdir -p /opt/him/backend/dist`) before rsyncing. Otherwise artifacts land at the wrong depth and the Docker container fails with `Cannot find module '/app/dist/index.js'`.
4. **Install node_modules on VM** (production only, inside container):
   ```
   ssh -i <key> ubuntu@<host> "cd /opt/him/backend && docker run --rm -v \$(pwd):/app -w /app node:22-alpine npm install --production"
   ```
5. **Transfer frontend build:**
   ```
   rsync -avz --delete frontend/dist/ ubuntu@<host>:/opt/him/frontend/
   ```
6. **Transfer docker-compose.yml and nginx config:**
   ```
   scp -i <key> deployment/docker-compose.yml ubuntu@<host>:/opt/him/docker-compose.yml
   scp -i <key> deployment/nginx/him.conf ubuntu@<host>:/etc/nginx/sites-available/him
   ```
7. **Write deploy stamp files** (enables SHA verification in health endpoint):
   ```
   ssh -i <key> ubuntu@<host> "echo '$(git rev-parse HEAD)' > /opt/him/backend/DEPLOY_SHA && echo '$(date -u +%Y-%m-%dT%H:%M:%SZ)' > /opt/him/backend/DEPLOY_TIME"
   ```
8. **Restart containers:**
   ```
   ssh -i <key> ubuntu@<host> "cd /opt/him && docker compose up -d --force-recreate"
   ```
9. **Reload nginx:**
   ```
   ssh -i <key> ubuntu@<host> "sudo nginx -t && sudo systemctl reload nginx"
   ```
10. **Invoke `smaqit.infrastructure-deploy-verify`** with the VM URL.

## Output
- Application artifacts deployed on VM, container running, nginx serving

## Scope
- Does NOT provision the VM. Use `smaqit.infrastructure-provision-cyso` for that.
- Does NOT handle database migrations. Current project uses SQLite with append-only schema.
- Does NOT run in-place `docker compose pull` — uses local artifacts, not a registry.

## Gotchas
- `Cannot find module '/app/dist/index.js'` — almost always caused by rsync trailing slash putting files at the wrong level. The container mounts `/opt/him/backend/` as `/app`; the Node process expects `/app/dist/index.js`. Always `mkdir -p /opt/him/backend/dist` before rsyncing.
- `VITE_DEMO_MODE` is build-time, not runtime. Changing the GitHub Actions variable and re-running the workflow does a rebuild — no code change needed. Changing it in the running environment has no effect.
- Deploy stamps (`DEPLOY_SHA`, `DEPLOY_TIME`) must be written to `/opt/him/backend/`, not `/opt/him/`. The Docker container mounts `/opt/him/backend/` as `/app`, so files in `/opt/him/` are invisible to the container.
- `docker compose` (v2, no hyphen) vs `docker-compose` (v1) — use `docker compose` (v2 syntax) on Ubuntu 24.04 with Docker 24+.
- `--force-recreate` ensures the container restarts even if the image hasn't changed. Required because the app is deployed as files, not as a new image.

## Completion
- [ ] Backend TypeScript build succeeded
- [ ] Frontend Vite build succeeded
- [ ] Backend artifacts rsynced to `/opt/him/backend/dist/`
- [ ] `node_modules` installed via Docker on VM
- [ ] Frontend dist rsynced to `/opt/him/frontend/`
- [ ] `docker-compose.yml` and nginx config transferred
- [ ] Deploy stamp files written
- [ ] Container restarted with `--force-recreate`
- [ ] nginx reloaded
- [ ] `smaqit.infrastructure-deploy-verify` invoked and passed

## Failure Handling
| Situation | Action |
|-----------|--------|
| `npm run build` fails | Show the build error; stop — do not deploy broken artifacts |
| rsync permission denied | Check SSH key and `ubuntu` user write permissions on `/opt/him/` |
| `Cannot find module` in container logs | Check rsync target depth; re-run with explicit `mkdir -p` |
| nginx config test fails | Show the nginx error; do not reload — the current config remains active |
| deploy-verify reports SHA mismatch | The old container may still be running; check `docker ps` and retry restart |

## Examples
**Input:** Feature branch merged to main; CI/CD workflow runs or operator invokes `/app.deploy`.
**Output:** Backend and frontend artifacts on VM, container running, nginx serving, health endpoint returning correct SHA and deployedAt timestamp.
