---
name: smaqit.infrastructure-vm-bootstrap
description: Use when a freshly provisioned Ubuntu VM needs to be bootstrapped after cloud-init completes. Covers cloud-init verification, data volume mount and fstab registration, app directory ownership fix, `.env` file creation, and Docker group setup. Produces a VM ready for application deployment on Ubuntu 22.04/24.04.
metadata:
  version: "1.1.1"
---

# VM Bootstrap

## Steps

All steps run via SSH to the target VM (`ssh -i <key> ubuntu@<host>`).

1. **Verify cloud-init** — run `cloud-init status`. If not `done`, retry every 10 seconds for up to 120 seconds. If still not `done` at the end of the wait, report the current status and stop — do not proceed.

2. **Locate and mount data volume:**

   a. Run `lsblk` to identify the data volume device. On Cyso/OpenStack, Cinder volumes appear as `/dev/sdb` (virtio-scsi), not `/dev/vdb` (virtio-blk).

   b. Check for an existing filesystem: `sudo blkid <device>`. If no filesystem is reported, format: `sudo mkfs.ext4 <device>`. If a filesystem already exists, skip formatting.

   c. Create the mount point and mount:
      ```
      sudo mkdir -p /data
      sudo mount <device> /data
      ```

   d. Register in fstab using UUID (not device path):
      ```
      echo "UUID=$(sudo blkid -s UUID -o value <device>) /data ext4 defaults,nofail 0 2" | sudo tee -a /etc/fstab
      ```

   e. Verify: `mount | grep /data`

3. **Fix app directory ownership** — cloud-init creates `__APP_DIR__` as root; always correct this:
   ```
   sudo chown -R ubuntu:ubuntu __APP_DIR__
   ```
   If `__APP_DIR__` does not exist, create it first: `sudo mkdir -p __APP_DIR__ && sudo chown ubuntu:ubuntu __APP_DIR__`

4. **Create `.env` file** (only if the stack needs runtime environment variables — skip if the
   project has no such requirement):

   a. Write the minimum required values to `__APP_DIR__/.env`, e.g.:
      ```
      NODE_ENV=production
      PORT=<app-port>
      DB_PATH=/data/<project-slug>.db
      ```

   b. Set permissions: `chmod 600 __APP_DIR__/.env`

   c. **Do NOT write `ANTHROPIC_API_KEY` to this file now.** Populate it manually via SSH after deployment is complete — never write its value to any file tracked in version control or stored in Terraform state.

5. **Add ubuntu to docker group:**
   ```
   sudo usermod -aG docker ubuntu
   newgrp docker
   ```
   `usermod` does not take effect in the current SSH session. `newgrp docker` activates the group for the current session. CI/CD pipelines use a fresh SSH session and inherit the updated group automatically.

6. **Remove the stock default nginx site:**
   ```
   bash scripts/remove-default-nginx-site.sh
   ```
   The distro's `default` site and this project's own site both declaring `default_server` fails
   `nginx -t` with "duplicate default server", leaving the previous config silently active after a
   reload. Idempotent — safe to re-run on every bootstrap/deploy, including when a second,
   co-hosted site is added later.

7. **Smoke test** — run `docker ps`. Expected result: empty container list with no permission error.

## Output

VM ready for application deployment: `/data` mounted with UUID-based fstab entry (`nofail`), `__APP_DIR__` owned by `ubuntu:ubuntu`, `__APP_DIR__/.env` present at permissions `600` (if applicable), `ubuntu` in the `docker` group (if the stack uses Docker), and no stock `default` nginx site left enabled.

## Scope

- Does **not** install software (Docker, nginx, certbot) — handled by cloud-init user-data managed by Terraform (`smaqit.infrastructure-provision-cyso`).
- Does **not** deploy the application — use `smaqit.infrastructure-deploy-rsync`.
- Does **not** write the application's own nginx site config — that's pushed separately during deployment. This skill only removes the distro's stock `default` site so it can't conflict.

## Examples

**Input:** Terraform apply completed. VM IP is `<vm-fixed-ip>`, SSH key at `~/.ssh/<project-slug>_key`. Operator invokes `/vm.bootstrap`.

**Output:** VM bootstrapped — `/data` mounted (UUID in fstab with `nofail`), `__APP_DIR__/.env` at `600`, `ubuntu` in docker group. Ready for `smaqit.infrastructure-deploy-rsync`.

## Gotchas

- **`/dev/sdb` not `/dev/vdb`** — Cyso OpenStack attaches Cinder volumes as virtio-scsi (`sdb`). Documentation may reference `vdb` (virtio-blk); on Cyso this has consistently been wrong. Use whatever `lsblk` reports.
- **UUID in fstab, not device path** — device names can change across reboots. Always use `blkid` to get the UUID.
- **`nofail` in fstab** — mandatory. Without it, a missing or unattached volume causes an emergency shell on reboot.
- **`chown __APP_DIR__`** — cloud-init creates the directory as root even when `mkdir -p` runs in user-data under the `ubuntu` user. Always run the ownership fix.
- **`chmod 600` on `.env`** — file contains production secrets. Missing permissions is a security vulnerability.
- **`newgrp docker`** — `usermod` does not update the current session. Use `newgrp docker` for the current session or re-SSH.
- **Two vhosts both declaring `default_server`** — fails `nginx -t` with "duplicate default server", and the reload silently no-ops, leaving the *previous* config (or the stock welcome page) serving. Always run `remove-default-nginx-site.sh` before the application's own site is enabled, and re-run it on every deploy — it's idempotent. This is the scenario that matters most once a second, co-hosted site is added to the same VM.
- **`systemctl reload nginx` on a never-started service fails outright** — cloud-init's `systemctl enable nginx` only arranges for nginx to start on the *next* boot; on a VM that hasn't been rebooted since creation, `nginx.service` may still be inactive, and a plain `reload` errors with "nginx.service is not active, cannot reload." Downstream deploy skills should use `systemctl reload-or-restart nginx`, not `reload`.

## Completion

- [ ] `cloud-init status` reports `done`
- [ ] Data volume located, formatted (if needed), and mounted at `/data`
- [ ] UUID-based fstab entry added with `nofail`
- [ ] `__APP_DIR__` owned by `ubuntu:ubuntu`
- [ ] `__APP_DIR__/.env` present with permissions `600` (if applicable)
- [ ] `ubuntu` in docker group — verified by `docker ps` returning no permission error (if the stack uses Docker)
- [ ] Stock `default` nginx site removed; `nginx -t` passes with no duplicate `default_server`

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `cloud-init status` not `done` after 120s | Report current status and stop — VM is not ready for bootstrapping |
| No data volume found by `lsblk` | Report the full `lsblk` output; ask the operator to confirm the device or whether the volume was attached in the cloud console |
| `mkfs.ext4` refuses because device already has a filesystem | Do not reformat. Mount as-is and continue. |
| `nginx -t` fails with "duplicate default server" | Run `scripts/remove-default-nginx-site.sh`; re-run `nginx -t` before reloading |
| `systemctl reload nginx` fails with "nginx.service is not active" | Use `systemctl reload-or-restart nginx` instead — the service was never started, only enabled |
