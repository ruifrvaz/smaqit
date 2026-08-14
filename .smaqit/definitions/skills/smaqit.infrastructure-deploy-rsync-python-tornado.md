# smaqit.infrastructure-deploy-rsync-python-tornado

## Description

Use when deploying a Python/Tornado application (single-process monolith, no Docker, no build
step, systemd-managed, SQLite persistence) to a remote VM via rsync. Synthesized for stacks that
match neither `smaqit.infrastructure-deploy-rsync` (Node.js + Vite/React, Docker Compose) nor
`smaqit.infrastructure-deploy-rsync-python-nextjs` (FastAPI + Next.js, Docker Compose) — this
variant has no container runtime, no frontend build, and no `backend/`/`frontend/` split. Used in
Phase 4 (Dev Environment Sweep) of `smaqit.new-greenfield-project`, invoked via its Phase 4 Step 6
no-match/synthesis procedure. Also usable as a manual fallback for direct VM deployment outside
the CI/CD pipeline.

## Provenance

- `synthesized: true`
- `synthesized-for-project: [a downstream project]`
- `synthesized-date: 2026-07-21`
- `synthesized-stack: Python 3.12/Tornado 6.4/SQLite (stdlib)/nginx 1.24/systemd 255, no Docker, no build step`
- Candidate for a future Task-086-style reconciliation into canonical `smaqit` once proven by real use.

## Required-inherited-context (per Task 087 Design Decisions — inherited verbatim, not reinvented)

1. **`__APP_DIR__` token convention** for the remote deploy path — same token
   `smaqit.infrastructure-deploy-rsync` and `smaqit.infrastructure-deploy-rsync-python-nextjs` use.
2. **nginx vhost writing delegates to the shared
   `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh`** — never a new
   `default_server`-vs-name-based implementation. Same invocation contract as the sibling skills.
3. **Deploy-stamp writing reuses `smaqit.infrastructure-hook-post-deploy-stamp`'s pattern** —
   `printf '%s'` (never `echo`, which appends a newline the health endpoint would then include),
   full (untruncated) SHA, ISO 8601 UTC timestamp.
4. **No Terraform-touching step in this skill** — `plan-guard.sh`/`ownership-guard.sh` reuse is
   not applicable here (this skill deploys application code only; provisioning is a separate skill,
   `smaqit.infrastructure-provision-cyso`). Documented for completeness per Task 087's
   required-inherited-context list, not omitted silently.

## Steps

### Pre-conditions

- VM bootstrapped (data volume mounted, app directory owned by `ubuntu`, nginx installed and
  enabled, `python3-tornado` installed via apt — see `smaqit.infrastructure-vm-bootstrap`; note
  that skill's packaged steps target a Docker+Node stack and were manually adapted for this
  project rather than run verbatim — see Notes)
- Local Vault running and unsealed (`smaqit.infrastructure-vault-loader` complete); SSH private
  key at `secret/<project-slug>/ssh`
- `deployment/nginx/<project-slug>.conf` and `deployment/systemd/<project-slug>.service` present
  locally

### Steps

1. **Fetch SSH key from Vault into a secure temp file:**
   ```bash
   export VAULT_ADDR=http://127.0.0.1:8200
   export PROJECT_SLUG=<project-slug>

   TMPKEY=$(mktemp)
   trap "rm -f $TMPKEY" EXIT
   vault kv get -field=private_key secret/${PROJECT_SLUG}/ssh > "$TMPKEY"
   chmod 600 "$TMPKEY"
   ```
   All subsequent `ssh`, `rsync`, and `scp` commands use `-i "$TMPKEY"`.

2. **No build step.** Python source runs directly — nothing to compile or bundle. Skip straight to
   transfer.

3. **Transfer application source to VM** (rsync entire source tree, excluding the live database,
   Python cache artifacts, and any local development virtualenv):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "mkdir -p __APP_DIR__/app __APP_DIR__/data"
   rsync -avz --delete \
     --exclude='*.db' \
     --exclude='__pycache__' \
     --exclude='*.pyc' \
     --exclude='.git' \
     --exclude='.venv' \
     src/ ubuntu@<host>:__APP_DIR__/app/
   ```
   `--exclude='*.db'` is CRITICAL — the SQLite database lives on the persistent data volume, not
   in the source tree; an accidental sync would overwrite live production data with an empty or
   stale local copy. `--exclude='.venv'` is also required — a local dev virtualenv's symlinks
   point at the developer machine's own Python path and don't resolve on the VM, which relies on
   the system-wide `python3-tornado` apt package instead.

4. **Transfer systemd unit and nginx config:**
   ```bash
   scp -i "$TMPKEY" deployment/systemd/<project-slug>.service ubuntu@<host>:/tmp/<project-slug>.service
   ssh -i "$TMPKEY" ubuntu@<host> "sudo mv /tmp/<project-slug>.service /etc/systemd/system/<project-slug>.service"
   ```
   The nginx conf is written by `scripts/write-vhost.sh` (reused verbatim from
   `smaqit.infrastructure-deploy-rsync`), not `scp`'d directly — it decides `default_server` vs.
   name-based on its own by inspecting `/etc/nginx/sites-enabled/` on the target VM:
   ```bash
   [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh \
     "$TMPKEY" <host> deployment/nginx/<project-slug>.conf <project-slug> [<server-name-if-co-hosted>]
   ```

5. **Write deploy stamp files** (enables SHA verification in the health endpoint):
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "printf '%s' '$(git rev-parse HEAD)' > __APP_DIR__/app/DEPLOY_SHA && \
      printf '%s' '$(date -u +%Y-%m-%dT%H:%M:%SZ)' > __APP_DIR__/app/DEPLOY_TIME"
   ```
   Write to `__APP_DIR__/app/`, matching the systemd unit's `WorkingDirectory` — the health
   endpoint handler reads these files relative to its own working directory. Use `printf '%s'`,
   never `echo` (see required-inherited-context item 3).

6. **Reload systemd and restart the application:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> \
     "sudo systemctl daemon-reload && \
      sudo systemctl enable <project-slug> && \
      sudo systemctl restart <project-slug>"
   ```
   `daemon-reload` is required whenever the unit file changed; `enable` is idempotent and safe to
   run on every deploy; `restart` (not `start`) picks up new code on redeploys.

7. **Reload nginx:**
   ```bash
   ssh -i "$TMPKEY" ubuntu@<host> "sudo nginx -t && sudo systemctl reload nginx"
   ```

8. **Verify:** Invoke `smaqit.infrastructure-deploy-verify` with the VM URL.

## Output

Application source deployed to `__APP_DIR__/app/` on the VM, systemd service running and enabled,
nginx serving via the shared `write-vhost.sh`-written vhost, deploy stamps present.

## Scope

- Does NOT provision the VM — use `smaqit.infrastructure-provision-cyso` for that.
- Does NOT handle database migrations — SQLite with append-only schema changes.
- Does NOT build or bundle anything — Python source runs directly, no build step.
- Does NOT manage Docker in any form — this stack has no container runtime.

## Completion

- [ ] Application source rsynced to `__APP_DIR__/app/`, excluding `*.db`/`__pycache__`/`.git`
- [ ] systemd unit transferred and installed at `/etc/systemd/system/<project-slug>.service`
- [ ] nginx vhost written via shared `write-vhost.sh` (not a new implementation)
- [ ] Deploy stamp files written to `__APP_DIR__/app/` using `printf '%s'` (no trailing newline)
- [ ] `systemctl daemon-reload && enable && restart` completed without error
- [ ] `nginx -t` passed; nginx reloaded
- [ ] `smaqit.infrastructure-deploy-verify` invoked and passed

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| rsync accidentally includes `*.db` | Stop immediately; verify the live data volume's `.db` file was not overwritten; restore from data volume snapshot if available |
| `systemctl restart` fails | Run `journalctl -u <project-slug> --no-pager -n 50` to inspect the failure; do not proceed to nginx reload |
| `nginx -t` fails | Report the error; do not reload — previous config remains active |
| `write-vhost.sh` exits 1 with "no server-name was supplied" | This VM already serves another site (co-hosted) — re-run with an explicit `<server-name-if-co-hosted>` argument |
| Health endpoint SHA mismatch after deploy | Confirm `systemctl restart` actually ran (not just `daemon-reload`); re-check stamp file path matches the unit's `WorkingDirectory` |

## Gotchas

- **`--exclude='*.db'` is non-negotiable** — the single most destructive mistake this skill could
  make is syncing the local (empty/stale) SQLite file over the VM's live database.
- **Stamp path must match systemd `WorkingDirectory`** — if the unit's `WorkingDirectory` differs
  from `__APP_DIR__/app/`, the health handler won't find the stamp files; keep both in sync.
- **`restart`, not `start`** — `start` is a no-op if the service is already running, which would
  silently deploy nothing on redeploy.
- **No Docker anywhere** — resist the temptation to add a Docker step by analogy with sibling
  deploy skills; this stack deliberately has none (see stack spec Excluded section).

## Allowed Tools

Bash(ssh:*), Bash(rsync:*), Bash(scp:*), Bash(vault:*), Bash(git:*)

## Examples

**Input:** a downstream project (Python 3.12/Tornado 6.4/SQLite/nginx/systemd) Phase 4 dev sweep invokes
this skill after Phase 4 Step 6 determined no existing deploy skill matches the declared stack.

**Output:** `src/` rsynced to `81.24.7.14:__APP_DIR__/app/` (excluding `*.db`), systemd unit
installed and the application service restarted, nginx vhost written via `write-vhost.sh` as
`default_server` (first site on this dedicated VM), deploy stamps written, health endpoint returns
correct SHA.

