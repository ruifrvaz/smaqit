# Skill Definition: smaqit.infrastructure-provision-cyso

## Identity
- **Name:** smaqit.infrastructure-provision-cyso
- **Version:** 1.0.0
- **Description:** Provision cloud infrastructure for the HIM Corporate application on Cyso Cloud (OpenStack) using Terraform. Covers: application credential setup, OpenRC sourcing, Object Storage state bucket, SSH keypair upload, `terraform init/plan/apply`, and fixed IP retrieval. Cyso-specific — for other cloud providers create a `provision-target-<provider>` skill using this as a template.

## Pre-conditions (one-time manual operator steps)
- Cyso Cloud account with access to region `ams2`
- Application credential created in Cyso Cloud Portal → Access → Credentials (name: `him-corporate-terraform`). Credential ID and secret noted — shown once only.
- OpenRC file downloaded from application credential page; saved outside repo at `~/him-openrc.sh`
- Object Storage container `him-corporate-tfstate` created in Cyso dashboard (private)
- S3 access key + secret key pair generated for the state bucket
- SSH keypair generated (passphrase-free deploy key): `ssh-keygen -t ed25519 -f ~/.ssh/him_deploy_key -N ""`
- SSH public key uploaded to Cyso → Compute → Key Pairs as `him-key`
- Terraform 1.14+ installed locally

## Steps
1. **Load OpenRC credentials:**
   ```
   source ~/him-openrc.sh
   ```
   Verify with `openstack token issue`. Do not proceed without a valid token.
2. **Confirm Ubuntu image ID and flavor:** These change occasionally in Cyso's catalog:
   ```
   openstack image list | grep "Ubuntu 24.04"
   openstack flavor list | grep s5.small
   ```
   Update `deployment/terraform/variables.tf` if the IDs differ from current values.
3. **Navigate to Terraform directory:**
   ```
   cd deployment/terraform
   ```
4. **Set backend environment variables** (required by `backend.tf`):
   ```
   export AWS_ACCESS_KEY_ID=<s3-access-key>
   export AWS_SECRET_ACCESS_KEY=<s3-secret-key>
   ```
5. **Set Terraform variables:**
   ```
   export TF_VAR_app_credential_id=<credential-id>
   export TF_VAR_app_credential_secret=<credential-secret>
   export TF_VAR_github_token=<fine-grained-pat>
   ```
   NOTE: `TF_VAR_github_token` — not `GITHUB_TOKEN` (reserved) and not `TF_VAR_GITHUB_TOKEN` (case-sensitive; Terraform maps `TF_VAR_github_token` → `var.github_token`).
6. **`terraform init`:**
   ```
   terraform init
   ```
   Confirms remote state backend connectivity. Lock file (`.terraform.lock.hcl`) is committed — run `terraform providers lock -platform=linux_amd64 -platform=darwin_arm64` if it needs regeneration.
7. **`terraform plan`:** Review the plan. Expected resources on first apply: 1 Nova VM, 1 boot volume (20 GB), 1 data volume (10 GB), 1 floating IP, 1 security group (ports 22/80/443), 1 keypair, 1 GitHub Actions variable.
8. **`terraform apply`:**
   ```
   terraform apply
   ```
   After apply, note the `fixed_ip` output — this is the public address (`81.24.10.203` for this project). The floating IP (`81.24.10.90`) does NOT route on Cyso's flat network — ignore it.
9. **SSH access test:** `ssh -i ~/.ssh/him_deploy_key ubuntu@<fixed_ip>` — should succeed within 60 seconds of apply.

## Output
- Cyso VM running, accessible via SSH at the fixed IP
- Cinder data volume attached (appears as `/dev/sdb` inside VM)
- Security group open on ports 22, 80, 443
- Remote Terraform state stored in Object Storage bucket

## Scope
- Does NOT bootstrap the VM post-provision. Use `smaqit.infrastructure-vm-bootstrap` for that.
- Does NOT deploy the application. Use `smaqit.infrastructure-deploy-rsync` for that.
- Does NOT configure nginx or Docker inside the VM (those are handled by cloud-init user-data in `main.tf`).
- Floating IP is provisioned by Terraform but should be treated as non-functional on Cyso flat network. Use fixed IP only.

## Gotchas
- **Floating IP does not route on Cyso flat network** — `81.24.10.90` is provisioned but not publicly accessible. Cyso assigns publicly-routable IPs directly to VM interfaces on the flat network; floating IP association via `openstack_networking_floatingip_associate_v2` does not route. Always use `outputs.fixed_ip`.
- **Data volume appears as `/dev/sdb` not `/dev/vdb`** — Cyso OpenStack presents Cinder volumes as SCSI devices, not virtio-blk. `lsblk` will show it as `sdb` (or `sdc` if multiple volumes). Do NOT assume `/dev/vdb`.
- **Keypair trailing newline drift** — the keypair resource will show as needing replacement if the public key string differs by a trailing newline between what Terraform has in state and what the variable provides. Always strip newlines when setting `TF_VAR_SSH_PUBLIC_KEY`: `export TF_VAR_SSH_PUBLIC_KEY=$(cat ~/.ssh/him_deploy_key.pub | tr -d '\n')`.
- **Provider lock file** — `.terraform.lock.hcl` is committed with hashes for `linux_amd64` and `darwin_arm64`. If running on a different platform, regenerate: `terraform providers lock -platform=<platform>`.
- **Application credential scope** — must be project-scoped in Cyso. A user-level credential without project scope will fail on resource creation.
- **`terraform destroy` not in scope** — do not run `terraform destroy` as part of deployment acceptance. It would tear down the live infrastructure.

## Completion
- [ ] OpenRC sourced; `openstack token issue` succeeds
- [ ] Ubuntu image ID and flavor confirmed
- [ ] Terraform backend environment variables set
- [ ] Terraform provider variables set (no `GITHUB_TOKEN` reserved name used)
- [ ] `terraform init` succeeded with remote state backend
- [ ] `terraform plan` reviewed (expected resource count confirmed)
- [ ] `terraform apply` succeeded; fixed IP noted from outputs
- [ ] SSH access to VM confirmed
- [ ] `.terraform.lock.hcl` committed (if regenerated)

## Failure Handling
| Situation | Action |
|-----------|--------|
| `openstack token issue` fails | Re-source OpenRC; check if application credential has expired |
| `terraform init` backend connectivity failure | Verify `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` are set; check Object Storage endpoint in `backend.tf` |
| Image ID not found | Run `openstack image list` and update `variables.tf` |
| Keypair replacement on plan (not first apply) | Strip trailing newline from public key; update `TF_VAR_SSH_PUBLIC_KEY` |
| SSH access fails after apply | Check security group allows port 22; verify key fingerprint matches uploaded keypair |

## Examples
**Input:** Cyso account set up, OpenRC file ready, state bucket created. Operator invokes `/provision.cyso`.
**Output:** VM at `81.24.10.203`, data volume attached, SSH accessible, Terraform state in `him-corporate-tfstate` bucket. Ready for `smaqit.infrastructure-vm-bootstrap`.
