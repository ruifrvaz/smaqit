# Skill Definition: smaqit.infrastructure-vm-bootstrap

## Identity
- **Name:** smaqit.infrastructure-vm-bootstrap
- **Version:** 1.0.0
- **Description:** Bootstrap a freshly provisioned Ubuntu VM after cloud-init completes. Covers: cloud-init verification, data volume mount, app directory ownership, `.env` file creation, Docker group setup. Target-agnostic for any Ubuntu 22.04/24.04 VM.

## Steps
All steps are executed via SSH to the target VM (`ssh -i <key> ubuntu@<host>`).

1. **Verify cloud-init** — run `cloud-init status`. If status is not `done`, wait up to 120 seconds and retry. If it times out, report the error and stop.
2. **Locate and mount data volume:**
   a. Run `lsblk` to identify the data volume device. NOTE: Cyso/OpenStack VMs present Cinder volumes as `/dev/sdb` (virtio-scsi), NOT `/dev/vdb` (virtio-blk) as the documentation may suggest.
   b. Format if unformatted: `sudo mkfs.ext4 <device>` (skip if already formatted — check with `sudo blkid <device>`).
   c. Create mount point: `sudo mkdir -p /data`.
   d. Mount: `sudo mount <device> /data`.
   e. Add to `/etc/fstab` (use UUID, not device path): `echo "UUID=$(sudo blkid -s UUID -o value <device>) /data ext4 defaults,nofail 0 2" | sudo tee -a /etc/fstab`.
   f. Verify: `mount | grep /data`.
3. **Fix app directory ownership** — cloud-init may create `/opt/him/` as root: `sudo chown -R ubuntu:ubuntu /opt/him`.
4. **Create `.env` file:**
   a. Write to `/opt/him/.env` with at minimum: `NODE_ENV=production`, `PORT=3001`, `DB_PATH=/data/him.db`.
   b. Set permissions: `chmod 600 /opt/him/.env`.
   c. NEVER write `ANTHROPIC_API_KEY` value to any file tracked in version control or stored in Terraform state. Populate it manually via SSH post-deploy.
5. **Add ubuntu to docker group:** `sudo usermod -aG docker ubuntu`. Note: group membership takes effect on next SSH session (`newgrp docker` for the current session).
6. **Smoke test:** run `docker ps` (after `newgrp docker`). Should return empty container list without permission error.

## Output
- VM ready for application deployment: data volume mounted, app directory writable, `.env` in place, Docker group active.

## Scope
- Does NOT install software (Docker, nginx, certbot). Those are handled by cloud-init user-data scripts managed by Terraform (`smaqit.infrastructure-provision-cyso`).
- Does NOT deploy the application. Use `smaqit.infrastructure-deploy-rsync` for that.
- Does NOT configure nginx. Nginx config is pushed separately during deployment.

## Gotchas
- `/dev/sdb` not `/dev/vdb` — Cyso OpenStack mounts Cinder volumes as SCSI devices. If `lsblk` shows `/dev/vdb`, use that, but on Cyso it has consistently been `/dev/sdb`.
- `fstab` with `nofail` — always use `nofail` so a missing volume on reboot doesn't cause an emergency shell. Use UUID, not device name (device names can change on reboot).
- `chown` on `/opt/him/` — cloud-init creates the directory as root even when `mkdir -p` is in user-data run as ubuntu. Always run the ownership fix.
- `.env` chmod 600 — mandatory. File contains production secrets. Failure to restrict permissions is a security vulnerability.
- `newgrp docker` — the `usermod -aG docker ubuntu` command does not take effect in the current SSH session. Use `newgrp docker` for the current session or re-SSH. CI/CD pipelines use a fresh SSH session so they get the updated group automatically.

## Completion
- [ ] cloud-init status: done confirmed
- [ ] Data volume located, formatted (if needed), and mounted at `/data`
- [ ] UUID-based fstab entry added
- [ ] `/opt/him/` owned by `ubuntu:ubuntu`
- [ ] `/opt/him/.env` present with correct permissions (600)
- [ ] ubuntu in docker group (verified with `groups` or `docker ps`)

## Failure Handling
| Situation | Action |
|-----------|--------|
| cloud-init not done after 120s | Report status and stop — VM is not ready |
| No data volume found by `lsblk` | Report available block devices; ask operator to confirm the device or whether volume was attached |
| `/opt/him/` does not exist | Create it: `mkdir -p /opt/him && sudo chown ubuntu:ubuntu /opt/him` |
| `mkfs.ext4` refuses (device has filesystem) | Do not reformat. Mount as-is. |

## Examples
**Input:** Terraform apply completed; VM is at `81.24.10.203` with SSH key at `~/.ssh/him_key`. Operator invokes `/vm.bootstrap`.
**Output:** VM bootstrapped — `/data` mounted (UUID in fstab), `/opt/him/.env` at 600, ubuntu in docker group. Ready for deployment.
