---
name: smaqit.infrastructure-vm-bootstrap
description: Use when a freshly provisioned Ubuntu VM needs to be bootstrapped after cloud-init completes. Covers cloud-init verification, data volume mount and fstab registration, app directory ownership fix, `.env` file creation, and Docker group setup. Produces a VM ready for application deployment on Ubuntu 22.04/24.04.
metadata:
  version: "1.0.0"
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

3. **Fix app directory ownership** — cloud-init creates `/opt/him/` as root; always correct this:
   ```
   sudo chown -R ubuntu:ubuntu /opt/him
   ```
   If `/opt/him/` does not exist, create it first: `sudo mkdir -p /opt/him && sudo chown ubuntu:ubuntu /opt/him`

4. **Create `.env` file:**

   a. Write the following to `/opt/him/.env` (these are the minimum required values):
      ```
      NODE_ENV=production
      PORT=3001
      DB_PATH=/data/him.db
      ```

   b. Set permissions: `chmod 600 /opt/him/.env`

   c. **Do NOT write `ANTHROPIC_API_KEY` to this file now.** Populate it manually via SSH after deployment is complete — never write its value to any file tracked in version control or stored in Terraform state.

5. **Add ubuntu to docker group:**
   ```
   sudo usermod -aG docker ubuntu
   newgrp docker
   ```
   `usermod` does not take effect in the current SSH session. `newgrp docker` activates the group for the current session. CI/CD pipelines use a fresh SSH session and inherit the updated group automatically.

6. **Smoke test** — run `docker ps`. Expected result: empty container list with no permission error.

## Output

VM ready for application deployment: `/data` mounted with UUID-based fstab entry (`nofail`), `/opt/him/` owned by `ubuntu:ubuntu`, `/opt/him/.env` present at permissions `600`, and `ubuntu` in the `docker` group.

## Scope

- Does **not** install software (Docker, nginx, certbot) — handled by cloud-init user-data managed by Terraform (`smaqit.infrastructure-provision-cyso`).
- Does **not** deploy the application — use `smaqit.infrastructure-deploy-rsync`.
- Does **not** configure nginx — nginx config is pushed separately during deployment.

## Examples

**Input:** Terraform apply completed. VM IP is `81.24.10.203`, SSH key at `~/.ssh/him_key`. Operator invokes `/vm.bootstrap`.

**Output:** VM bootstrapped — `/data` mounted (UUID in fstab with `nofail`), `/opt/him/.env` at `600`, `ubuntu` in docker group. Ready for `smaqit.infrastructure-deploy-rsync`.

## Gotchas

- **`/dev/sdb` not `/dev/vdb`** — Cyso OpenStack attaches Cinder volumes as virtio-scsi (`sdb`). Documentation may reference `vdb` (virtio-blk); on Cyso this has consistently been wrong. Use whatever `lsblk` reports.
- **UUID in fstab, not device path** — device names can change across reboots. Always use `blkid` to get the UUID.
- **`nofail` in fstab** — mandatory. Without it, a missing or unattached volume causes an emergency shell on reboot.
- **`chown /opt/him/`** — cloud-init creates the directory as root even when `mkdir -p` runs in user-data under the `ubuntu` user. Always run the ownership fix.
- **`chmod 600` on `.env`** — file contains production secrets. Missing permissions is a security vulnerability.
- **`newgrp docker`** — `usermod` does not update the current session. Use `newgrp docker` for the current session or re-SSH.

## Completion

- [ ] `cloud-init status` reports `done`
- [ ] Data volume located, formatted (if needed), and mounted at `/data`
- [ ] UUID-based fstab entry added with `nofail`
- [ ] `/opt/him/` owned by `ubuntu:ubuntu`
- [ ] `/opt/him/.env` present with permissions `600`
- [ ] `ubuntu` in docker group — verified by `docker ps` returning no permission error

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
