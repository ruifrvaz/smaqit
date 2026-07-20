---
name: smaqit.infrastructure-deploy-rsync-python-nextjs
description: Use when deploying a Python/FastAPI backend + Next.js frontend application to a remote VM via rsync. Validated on Fashion App — AI Stylist deployed to Cyso Cloud (s5.small, Ubuntu 24.04). Covers Python source rsync, Next.js production build via Docker, database migration ordering, and Docker build gotchas. For Node.js + Vite/React deployments, use `smaqit.infrastructure-deploy-rsync`.
metadata:
  version: "1.3.0"
  validated: "2026-07-17"
  validated-stack: "Python 3.12, FastAPI 0.115, Next.js 15, pnpm 9, PostgreSQL 16, Docker Compose"
---

# Deploy Python/FastAPI + Next.js via rsync

Validated path for deploying a Python backend with a Next.js frontend to a remote VM via rsync.
Based on the Fashion App — AI Stylist deployment to Cyso Cloud.

## Pre-conditions

- VM bootstrapped (`smaqit.infrastructure-vm-bootstrap` complete)
- Local Vault running and unsealed; SSH private key at `secret/<project-slug>/ssh`
- Docker running on VM and `ubuntu` in docker group
- `deployment/docker-compose.yml` and `deployment/nginx/<project-slug>.conf` present locally
- Production docker-compose MUST include `127.0.0.1` port mappings for nginx (see Docker section)

## Steps

1. **Fetch SSH key from Vault:**
   ```bash
   export VAULT_ADDR=http://127.0.0.1:8200
   export PROJECT_SLUG=<project-slug>
   TMPKEY=$(mktemp) && trap "rm -f $TMPKEY" EXIT
   vault kv get -field=private_key secret/${PROJECT_SLUG}/ssh > "$TMPKEY"
   chmod 600 "$TMPKEY"
   ```
   If this fails with "error in libcrypto", see Failure Handling below — that's a Vault-stored-key
   trailing-newline issue (see also `smaqit.infrastructure-vault-loader` gotchas), not something to
   routinely work around here.

2. **Build frontend** (Next.js, production mode):
   ```bash
   cd frontend && NEXT_PUBLIC_API_URL= pnpm build
   ```
   Produces `frontend/.next/`. `NEXT_PUBLIC_API_URL=` (empty string) is REQUIRED — it is a
   client-side, build-time-baked variable; setting it empty makes the browser bundle call
   relative paths (`/api/v1/...`), which nginx then proxies to the backend. Leaving it unset
   bakes in the `??` fallback default instead (see Gotchas), which is wrong for any real
   deployment. If this fails with `EACCES` on `.next/`, see Gotchas / Failure Handling below —
   that's a leftover-permissions issue from a previous Docker build, not something to routinely
   work around here.

3. **Transfer backend to VM** (rsync entire source tree):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "sudo find __APP_DIR__/backend \( -name '__pycache__' -o -name '*.egg-info' -o -name '.eggs' -o -name 'build' \) -exec rm -rf {} + 2>/dev/null || true"
   ssh -i "$TMPKEY" ubuntu@<host> "mkdir -p __APP_DIR__/backend"
   rsync -avz --delete backend/ ubuntu@<host>:__APP_DIR__/backend/
   ```
   CRITICAL: Rsync `backend/` recursively — this must include top-level files
   (`app/__init__.py`, `app/main.py`, `app/db.py`, `pyproject.toml`, `Dockerfile`).
   Do NOT rsync only subdirectories.

   The pre-rsync cleanup MUST use `find` recursively, not a flat glob like
   `backend/*.egg-info backend/__pycache__` — Python creates one `__pycache__` per package
   directory, and any left root-owned from a past deploy (e.g. `alembic/__pycache__`,
   `app/models/__pycache__`) will make `rsync --delete` fail with `Permission denied` on the
   nested ones a flat glob never reaches.

4. **Transfer frontend build** (production artifacts only — no full source needed):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "sudo rm -rf __APP_DIR__/frontend && mkdir -p __APP_DIR__/frontend"
   rsync -avz frontend/.next/ ubuntu@<host>:__APP_DIR__/frontend/.next/
   rsync -avz frontend/package.json frontend/pnpm-lock.yaml frontend/next.config.js frontend/Dockerfile ubuntu@<host>:__APP_DIR__/frontend/
   ```
   CRITICAL: `Dockerfile` MUST be included in this transfer. `docker-compose.yml`'s `frontend`
   service builds from this context on the VM — without the Dockerfile present, `docker compose
   build` fails outright. `next start` (production mode) needs only `.next/`, `package.json`,
   `pnpm-lock.yaml`, and `next.config.js` at runtime — no `src/`, `tsconfig.json`, or other
   source/config files. If the project has a `public/` directory, rsync it too
   (`rsync -avz frontend/public/ ubuntu@<host>:__APP_DIR__/frontend/public/`).

   CRITICAL: Wipe and recreate `__APP_DIR__/frontend` first, don't rely on per-subpath
   `--delete`. `rsync --delete` on just the `.next/` subfolder leaves anything *outside* it
   untouched — a stray `node_modules/` left over from an earlier pipeline iteration (e.g. one
   that used to install deps directly on the VM) gets picked up by the Dockerfile's `COPY . .`
   and shadows the fresh `RUN pnpm install` layer, corrupting pnpm's internal state. Symptom:
   the container logs `ERROR packages field missing or empty` and the frontend never starts,
   even though the image builds successfully — nginx then serves 502 for every page.

5. **Transfer config files:**
   ```bash
   scp -i "$TMPKEY" deployment/docker-compose.yml ubuntu@<host>:__APP_DIR__/docker-compose.yml
   ssh -i "$TMPKEY" ubuntu@<host> "bash -s" < [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vm-bootstrap/scripts/remove-default-nginx-site.sh
   [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh "$TMPKEY" <host> deployment/nginx/<project-slug>.conf <project-slug> [<server-name-if-co-hosted>]
   ```
   The stock distro `default` site must be removed before the first reload — see
   `smaqit.infrastructure-vm-bootstrap`'s nginx Gotcha. That script is idempotent; safe to run on
   every deploy.

   The nginx vhost itself is written by the same `write-vhost.sh` that
   `smaqit.infrastructure-deploy-rsync` (the Node.js sibling) uses — vhost-writing is not
   stack-specific, so there is exactly one implementation of the `default_server`-vs-name-based
   decision shared by both deploy skills, not two. It inspects `/etc/nginx/sites-enabled/` on the
   target VM itself: `default_server` if this is the first site on the VM, name-based only (using
   the supplied `<server-name-if-co-hosted>`) otherwise — refusing to proceed without that
   argument on a co-hosted VM rather than guessing. It uploads to
   `/etc/nginx/sites-available/<project-slug>`, symlinks into `sites-enabled/`, and runs
   `nginx -t` itself; a non-zero exit means the previous config is still active (see Step 8).

6. **Build and restart containers:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "cd __APP_DIR__ && docker compose build && docker compose up -d --force-recreate"
   ```
   CRITICAL: `docker compose build` is REQUIRED before `up`. Compose does not rebuild images on
   `up` by default — without an explicit `build`, a redeploy silently keeps running the previous
   image, and rsynced code changes never take effect. Use `up -d --force-recreate`, NOT `restart`
   — `restart` reuses stale container config.

7. **Run database migrations** (only after containers are up and `db` is healthy):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "bash -s" -- __APP_DIR__ db api -- alembic upgrade head \
     < [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-deploy-rsync-python-nextjs/scripts/run-migrations.sh
   ```
   REQUIRED, and REQUIRED to run after step 6, not before. The script waits for `db`'s
   healthcheck, then runs the migration inside the already-running `api` container via
   `docker compose exec` — never a standalone throwaway container, which cannot reach the
   compose network's `db` hostname. Skipping this causes 500 errors: `relation "users" does
   not exist`.

8. **Reload nginx:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "sudo nginx -t && sudo systemctl reload-or-restart nginx"
   ```
   Use `reload-or-restart`, not a bare `reload` — cloud-init's `systemctl enable nginx` only
   arranges for nginx to start on the *next* boot, it does not start it immediately. On a VM
   that's never been rebooted since creation, nginx.service may still be inactive, and a plain
   `reload` fails with "nginx.service is not active, cannot reload." `reload-or-restart` starts
   it if inactive and reloads it if already running, so it works in both cases.

9. **Write deploy stamps:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "printf '%s' '$(git rev-parse HEAD)' > __APP_DIR__/backend/DEPLOY_SHA && \
      printf '%s' '$(date -u +%Y-%m-%dT%H:%M:%SZ)' > __APP_DIR__/backend/DEPLOY_TIME"
   ```

10. **Verify:** Invoke `smaqit.infrastructure-deploy-verify` with the VM URL.

## Docker-Related Deployments

### Production docker-compose.yml requirements

1. **Port mappings for nginx:** Add `127.0.0.1` bindings so nginx on the host can reach containers:
   ```yaml
   api:
     ports: ["127.0.0.1:8000:8000"]
   frontend:
     ports: ["127.0.0.1:3000:3000"]
   ```
   Without these, nginx returns 502 because it cannot reach the Docker network.

2. Remove dev bind-mount volumes (`./backend:/app`). Files are baked into images at build time.

3. Remove `--reload` flags from uvicorn — production uses the base command.

### Python/FastAPI Dockerfile — validated pattern

```dockerfile
FROM python:3.12-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends curl libpq-dev
RUN pip install --no-cache-dir uv
COPY . .
RUN uv pip install --system ".[dev]"
ENV PYTHONPATH=/app
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

CRITICAL: `COPY . .` must come BEFORE `RUN uv pip install`. Do NOT use `-e` (editable) in Docker —
it requires the source tree at link time, which the layered COPY doesn't provide.
`ENV PYTHONPATH=/app` is required when the setuptools config doesn't register top-level imports.

### Next.js frontend — production mode via Docker build

`next start` (production mode) works via the standard Docker build path — it does not require
`public/`, `src/`, or any TypeScript config at runtime, only `.next/`, `node_modules/` (installed
by the Dockerfile's own `pnpm install`), `package.json`, `pnpm-lock.yaml`, and `next.config.js`.
The `docker-compose.yml` `frontend` service overrides the Dockerfile's default `CMD` with
`command: pnpm start`; `docker compose build` runs the Dockerfile's `pnpm install` fresh and then
`COPY . .` layers in the rsynced `.next/` build output.

Set `NEXT_PUBLIC_API_URL=` (empty string) at **build time** (step 2), not as a runtime container
env var — it is a `NEXT_PUBLIC_*` variable, baked into the client JS bundle at `pnpm build`, and
has no effect if only set at container runtime. The frontend code must use `??` (nullish
coalescing), not `||`, for the fallback:
```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
```
`||` treats an empty string as falsy and would fall through to the `localhost` default even
after the fix.

### Container lifecycle

- `docker restart` does NOT recreate bind mounts — use `docker compose up -d --force-recreate`
- After `docker compose build`, the image must be re-created with `up`, not `restart`
- `sed` on YAML files is fragile — prefer rsyncing a fresh file from local

## Output

Application artifacts deployed to `__APP_DIR__/` on the VM, containers running, nginx serving.

## Scope

- Does NOT provision the VM — use `smaqit.infrastructure-provision-cyso` for that.
- Does NOT pull from a container registry — deploys local build artifacts directly via
  `docker compose build` on the VM.
- For Node.js + Vite/React deployments, use `smaqit.infrastructure-deploy-rsync`.

## Gotchas

### Python backend
- **`ModuleNotFoundError: No module named 'app.main'`** — Dockerfile has `RUN uv pip install` before
  `COPY . .`. Fix: move `COPY . .` before the install step, add `ENV PYTHONPATH=/app`.
- **Top-level source files missing** — rsyncing `backend/app/*/` only copies subdirectories,
  misses `app/__init__.py`, `app/main.py`, `app/db.py`. Always rsync `backend/` recursively.
- **Do not add a standalone "install deps on VM" step** — `docker compose build` already runs the
  Dockerfile's install step. A separate `docker run ... pip install` against the bind-mounted
  source directory is redundant and, run as root, leaves root-owned `*.egg-info`/`build/`
  artifacts that then break a subsequent `rsync --delete` with `Permission denied`.
- **Pre-rsync cleanup must be recursive** — a flat `rm -rf backend/__pycache__` only clears the
  top-level directory. Python creates one `__pycache__` per package, so root-owned ones from a past
  bad deploy can still be sitting at `alembic/__pycache__`, `app/models/__pycache__`, etc. Use
  `find backend \( -name '__pycache__' -o -name '*.egg-info' ... \) -exec rm -rf {} +` instead.

### Next.js frontend
- **`.next/` Docker-owned permissions** — EACCES on a plain `rm -rf .next` after a Docker build
  wrote to it as root. Fix: `docker run --rm -v "$(pwd)":/app -w /app alpine rm -rf .next`.
- **`NEXT_PUBLIC_API_URL=""` truthiness** — `||` treats empty string as falsy, falling back to
  `localhost:8080`. Use `??` (nullish coalescing) instead.
- **`NEXT_PUBLIC_API_URL` must be set at build time, not runtime** — it's baked into the client
  bundle by `pnpm build`; setting it as a container env var afterward has no effect.
- **Dockerfile must be rsynced** — `docker compose build` needs it present in the build context on
  the VM; it is easy to forget since it isn't a build *output*.
- **Do not add a standalone "install deps on VM" step for the frontend either** — same reasoning
  as the backend; `docker compose build` already runs `pnpm install` from the Dockerfile.
- **`ERROR packages field missing or empty` at container start, image builds fine** — a stale
  `node_modules/` (or other leftover) outside `.next/` in `__APP_DIR__/frontend` from an earlier
  pipeline iteration got `COPY`'d over the Dockerfile's own fresh `pnpm install`. Fix: wipe the
  whole `frontend` directory before rsyncing, don't rely on `--delete` scoped to just `.next/`.

### Docker
- **Port exposure** — production compose needs `127.0.0.1` port mappings for nginx on host.
  Without them: 502 Bad Gateway.
- **`docker compose build` before `up`** — Compose does not rebuild on `up` by default; skipping
  the build step silently redeploys the previous image even after rsyncing new code.
- **`docker restart` vs `--force-recreate`** — `restart` keeps stale bind mounts.
  Always use `up -d --force-recreate`.

### Database
- **Migration ordering** — migrations must run after `docker compose up`, once `db` is healthy, via
  `docker compose exec` inside the already-running app container — never a standalone container
  reaching for a `db` hostname it can't resolve outside the compose network. Use
  `scripts/run-migrations.sh`, which waits for the healthcheck before running.

## Failure Handling

| Situation | Action |
|-----------|--------|
| SSH key "error in libcrypto" | Store key as base64 in Vault; decode on fetch |
| `pnpm build` EACCES on `.next/` | Remove via a root container (`docker run --rm -v "$(pwd)":/app -w /app alpine rm -rf .next`), then rebuild |
| `ModuleNotFoundError: app.main` | Dockerfile: COPY before RUN; add ENV PYTHONPATH=/app |
| nginx 502 Bad Gateway | Check `ss -tlnp` for ports; add `127.0.0.1` mappings to compose |
| `write-vhost.sh` exits 1 with "no server-name was supplied" | This VM already serves another site (co-hosted) — re-run with an explicit `<server-name-if-co-hosted>` argument |
| `docker compose build` fails, "Dockerfile not found" | Confirm the Dockerfile was included in the frontend/backend rsync step |
| "relation does not exist" in API logs | `db` migration hasn't run; invoke `scripts/run-migrations.sh` after `docker compose up` |
| Redeployed code not reflected in running container | Missing `docker compose build` before `up` — image was reused unchanged |
| Frontend container logs `packages field missing or empty`, nginx 502s on `/` | Stale files outside `.next/` in `__APP_DIR__/frontend` (e.g. old `node_modules/`) shadowed the fresh install — wipe and recreate the whole directory before rsyncing |
| compose validation error | YAML corrupted; rsync clean file from local |
| SHA mismatch in verify | Use `up -d --force-recreate`, not `restart`; confirm `docker compose build` ran first |
