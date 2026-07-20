---
name: smaqit.infrastructure-deploy-rsync
description: Use when deploying a Node.js backend + React frontend application to a remote VM via rsync. Used in the Phase 5 dev environment sweep of `smaqit.new-greenfield-project` to validate the deployment approach locally before CI/CD. Also use as a manual fallback for direct VM deployment outside the CI/CD pipeline.
metadata:
  version: "1.2.0"
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
   export PROJECT_SLUG=<project-slug>   # from CLAUDE.md (or copilot-instructions.md)

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
   ssh -i "$TMPKEY" ubuntu@<host> "mkdir -p __APP_DIR__/backend/dist"
   rsync -avz --delete backend/dist/ ubuntu@<host>:__APP_DIR__/backend/dist/
   rsync -avz backend/package.json backend/package-lock.json ubuntu@<host>:__APP_DIR__/backend/
   ```
   CRITICAL: Always `mkdir -p __APP_DIR__/backend/dist` before rsyncing. The trailing slash on
   `backend/dist/` copies the directory's *contents* — if the target directory does not exist, rsync
   creates it one level too shallow and the container fails with
   `Cannot find module '/app/dist/index.js'`.

4. **Install `node_modules` on VM** (production only, inside a throwaway container):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "cd __APP_DIR__/backend && docker run --rm -v \$(pwd):/app -w /app node:22-alpine npm install --production"
   ```

5. **Transfer frontend build:**
   ```bash
   rsync -avz --delete frontend/dist/ ubuntu@<host>:__APP_DIR__/frontend/
   ```

6. **Transfer config files:**
   ```bash
   scp -i "$TMPKEY" deployment/docker-compose.yml ubuntu@<host>:__APP_DIR__/docker-compose.yml
   ```
   The nginx conf is written by `scripts/write-vhost.sh`, not `scp`'d directly — it decides
   `default_server` vs. name-based on its own by inspecting `/etc/nginx/sites-enabled/` on the
   target VM (this matters whenever the target is co-hosted, e.g.
   `provisioning_mode: existing-shared`, or any VM already serving another project), rather than
   that decision being made in prose here:
   ```bash
   scripts/write-vhost.sh "$TMPKEY" <host> deployment/nginx/him.conf <project-slug> [<server-name-if-co-hosted>]
   ```
   - **Nothing else enabled (first site on this VM):** the script sets `listen 80 default_server;`
     — this is the catch-all for requests with no matching `Host` header.
   - **Another site is already enabled (co-hosted, this is not the first):** the script requires
     an explicit `<server-name-if-co-hosted>` argument and writes a name-based vhost only — `listen
     80;` with that `server_name`, and no `default_server`. It refuses to proceed without that
     argument rather than guessing one. Two vhosts both claiming `default_server` on the same port
     fails `nginx -t` outright.
   - The script uploads to `/etc/nginx/sites-available/<project-slug>`, symlinks into
     `sites-enabled/`, and runs `nginx -t` itself — a non-zero exit means the previous config is
     still active and nothing was reloaded (see Step 9).

7. **Write deploy stamp files** (enables SHA verification in the health endpoint):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "printf '%s' '$(git rev-parse HEAD)' > __APP_DIR__/backend/DEPLOY_SHA && \
      printf '%s' '$(date -u +%Y-%m-%dT%H:%M:%SZ)' > __APP_DIR__/backend/DEPLOY_TIME"
   ```
   Write to `__APP_DIR__/backend/`, not `__APP_DIR__/` — the container mounts `__APP_DIR__/backend/` as `/app`;
   files one level up are invisible to the container.

8. **Restart containers:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "cd __APP_DIR__ && docker compose up -d --force-recreate"
   ```
   Use `docker compose` (v2, no hyphen). `--force-recreate` is required because the app is deployed
   as files, not as a new image.

9. **Reload nginx:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "sudo nginx -t && sudo systemctl reload nginx"
   ```

10. **Verify:** Invoke `smaqit.infrastructure-deploy-verify` with the VM URL.

## Output

Application artifacts deployed to `__APP_DIR__/` on the VM, container running, nginx serving.

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
  Always `mkdir -p __APP_DIR__/backend/dist` before rsyncing `backend/dist/`.
- **`VITE_DEMO_MODE` is build-time only** — changing the GitHub Actions variable and re-running the
  workflow triggers a rebuild. Changing it in the running environment has no effect.
- **Deploy stamp path** — write to `__APP_DIR__/backend/DEPLOY_SHA` and `__APP_DIR__/backend/DEPLOY_TIME`,
  not `__APP_DIR__/`. Files in `__APP_DIR__/` are invisible to the container.
- **`docker compose` vs `docker-compose`** — use `docker compose` (v2, no hyphen) on Ubuntu 24.04
  with Docker 24+.
- **`--force-recreate`** — required even when no image changed; ensures the container restarts with
  updated files.
- **`default_server` vs name-based vhost on a co-hosted VM** — only the *first* site deployed to a
  VM should claim `default_server` in its nginx conf. Every subsequent co-hosted site's vhost must
  be name-based only (`listen 80;` + explicit `server_name`), never `default_server` — a second
  `default_server` on the same port makes `nginx -t` fail. This applies whenever a VM serves more
  than one project, most commonly under `provisioning_mode: existing-shared`.
  `scripts/write-vhost.sh` (Step 6) makes this decision deterministically by inspecting
  `/etc/nginx/sites-enabled/` itself — it is not something to reconstruct in prose or by hand.

## Completion

- [ ] Backend TypeScript build succeeded
- [ ] Frontend Vite build succeeded
- [ ] Backend artifacts rsynced to `__APP_DIR__/backend/dist/`
- [ ] `node_modules` installed via Docker on VM
- [ ] Frontend dist rsynced to `__APP_DIR__/frontend/`
- [ ] `docker-compose.yml` transferred; nginx config written via `scripts/write-vhost.sh` (`default_server` vs name-based decided deterministically, not assumed) and `nginx -t` passed
- [ ] Deploy stamp files written to `__APP_DIR__/backend/`
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
| rsync permission denied | Check SSH key and `ubuntu` user write permissions on `__APP_DIR__/` |
| `Cannot find module` in container logs | Check rsync target depth; re-run step 3 with explicit `mkdir -p` |
| nginx config test fails | Show the nginx error; do not reload — current config remains active |
| `scripts/write-vhost.sh` exits 1 with "no server-name was supplied" | This VM already serves another site (co-hosted) — re-run with an explicit `<server-name-if-co-hosted>` argument |
| `scripts/write-vhost.sh`'s `nginx -t` step fails with a duplicate `default_server` error | Another site already claims `default_server` on this VM outside of what the script detected (e.g. a manually-added vhost) — investigate `/etc/nginx/sites-enabled/` directly before re-running |
| deploy-verify reports SHA mismatch | Check `docker ps`; old container may still be running; retry step 8 |
